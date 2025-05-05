// Copyright 2025 The celo Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

package miner

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/misc/eip1559"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMinerFillTransactionsOrdering verifies that the miner orders transactions
// based solely on their locality (local or remote), nonce, and gas price.
func TestMinerFillTransactionsOrdering(t *testing.T) {
	t.Parallel()

	var (
		key1     = core.DevPrivateKey
		address1 = core.DevAddr
		key2     = core.DevPrivateKey2

		miner        = createCeloMiner(t)
		signer       = types.LatestSigner(miner.chainConfig)
		parentHeader = miner.chain.CurrentBlock()
	)

	txAndExpectedEffectiveTips := []struct {
		tx                   *types.Transaction
		expectedEffectiveTip *big.Int // expected effective gas price on Celo after base-fee reduction
	}{
		{
			expectedEffectiveTip: big.NewInt(99.125 * params.GWei),
			tx: types.MustSignNewTx(key1, signer, &types.LegacyTx{
				Nonce:    0,
				To:       &common.ZeroAddress,
				Value:    big.NewInt(1),
				Gas:      71000,
				GasPrice: big.NewInt(100 * params.GWei),
			}),
		},
		{
			expectedEffectiveTip: big.NewInt(98.125 * params.GWei),
			tx: types.MustSignNewTx(key2, signer, &types.LegacyTx{
				Nonce:    0,
				To:       &common.ZeroAddress,
				Value:    big.NewInt(2),
				Gas:      71000,
				GasPrice: big.NewInt(99 * params.GWei),
			}),
		},
		{
			expectedEffectiveTip: big.NewInt(98 * params.GWei),
			tx: types.MustSignNewTx(key2, signer, &types.DynamicFeeTx{
				ChainID:   miner.chainConfig.ChainID,
				Nonce:     1,
				To:        &common.ZeroAddress,
				Value:     big.NewInt(3),
				Gas:       71000,
				GasFeeCap: big.NewInt(10000 * params.GWei),
				GasTipCap: big.NewInt(98 * params.GWei),
			}),
		},
		{
			expectedEffectiveTip: big.NewInt(97 * params.GWei),
			tx: types.MustSignNewTx(key1, signer, &types.CeloDynamicFeeTxV2{
				ChainID:     miner.chainConfig.ChainID,
				Nonce:       1,
				To:          &common.ZeroAddress,
				Value:       big.NewInt(4),
				Gas:         71000,
				GasFeeCap:   big.NewInt(10000 * params.GWei),
				GasTipCap:   big.NewInt(194 * params.GWei),
				FeeCurrency: &core.DevFeeCurrencyAddr,
			}),
		},
		{
			expectedEffectiveTip: big.NewInt(94.125 * params.GWei),
			tx: types.MustSignNewTx(key1, signer, &types.AccessListTx{
				Nonce:    2,
				To:       &common.ZeroAddress,
				Value:    big.NewInt(5),
				Gas:      71000,
				GasPrice: big.NewInt(95 * params.GWei),
			}),
		},
		{
			expectedEffectiveTip: big.NewInt(49.125 * params.GWei),
			tx: types.MustSignNewTx(key2, signer, &types.LegacyTx{
				Nonce:    2,
				To:       &common.ZeroAddress,
				Value:    big.NewInt(6),
				Gas:      71000,
				GasPrice: big.NewInt(50 * params.GWei),
			}),
		},
		{
			expectedEffectiveTip: big.NewInt(199.125 * params.GWei),
			tx: types.MustSignNewTx(key2, signer, &types.LegacyTx{
				Nonce:    3,
				To:       &common.ZeroAddress,
				Value:    big.NewInt(7),
				Gas:      71000,
				GasPrice: big.NewInt(200 * params.GWei),
			}),
		},
	}

	testCases := []struct {
		name      string
		txIndices []int // list of indices used to determine order of transactions
	}{
		{
			name:      "original order",
			txIndices: []int{0, 1, 2, 3, 4, 5, 6},
		},
		{
			name:      "reverse order",
			txIndices: []int{6, 5, 4, 3, 2, 1, 0},
		},
		{
			name:      "Account 1’s transactions first, followed by Account 2’s transactions",
			txIndices: []int{0, 3, 4, 1, 2, 5, 6},
		},
		{
			name:      "Account 2’s transactions first, followed by Account 1’s transactions",
			txIndices: []int{1, 2, 5, 6, 0, 3, 4},
		},
		{
			name:      "random order",
			txIndices: []int{3, 0, 5, 1, 6, 2, 4},
		},
	}

	// Get BaseFee and Rates
	baseFee := eip1559.CalcBaseFee(miner.chainConfig, parentHeader, parentHeader.Time+1)

	stateDb, err := miner.backend.BlockChain().StateAt(parentHeader.Root)
	require.NoError(t, err)
	rates := core.GetExchangeRates(parentHeader, miner.chainConfig, stateDb)

	// Creates a slice of transactions while validating effective gas prices
	txs := make([]*types.Transaction, len(txAndExpectedEffectiveTips))
	for idx := range txAndExpectedEffectiveTips {
		txs[idx] = txAndExpectedEffectiveTips[idx].tx

		effectiveGasPrice, err := txs[idx].EffectiveGasTipInCelo(baseFee, rates)
		require.NoError(t, err)

		require.Equal(t, txAndExpectedEffectiveTips[idx].expectedEffectiveTip, effectiveGasPrice)
	}

	requireNoErrors := func(t *testing.T, errs []error) {
		t.Helper()
		for _, err := range errs {
			require.NoError(t, err)
		}
	}

	reorderTxs := func(original []*types.Transaction, indicies []int) []*types.Transaction {
		ordered := make([]*types.Transaction, len(original))
		for idx := range indicies {
			ordered[idx] = original[indicies[idx]]
		}
		return ordered
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			txs := reorderTxs(txs, test.txIndices)

			// Verify that transaction ordering depends only on nonce and gas price when all transactions in the TxPool are local
			t.Run("all local transactions", func(t *testing.T) {
				miner := createCeloMiner(t)

				errs := miner.txpool.Add(txs, true, false)
				requireNoErrors(t, errs)

				res := miner.generateWork(&generateParams{
					parentHash: parentHeader.Hash(),
					timestamp:  parentHeader.Time + 1,
					random:     common.HexToHash("0xcafebabe"),
					noTxs:      false,
					forceTime:  true,
				}, false)
				require.NoError(t, res.err)

				require.Len(t, res.block.Transactions(), len(txs))
				for index, tx := range res.block.Transactions() {
					assert.Equal(t, tx.Value(), big.NewInt(int64(index+1)))
				}
			})

			// Verify that transaction ordering depends only on nonce and gas price when all transactions in the TxPool are remote
			t.Run("all remote transactions", func(t *testing.T) {
				miner := createCeloMiner(t)

				miner.txpool.Clear()
				errs := miner.txpool.Add(txs, false, false)
				requireNoErrors(t, errs)

				res := miner.generateWork(&generateParams{
					parentHash: parentHeader.Hash(),
					timestamp:  parentHeader.Time + 1,
					random:     common.HexToHash("0xcafebabe"),
					noTxs:      false,
					forceTime:  true,
				}, false)
				require.NoError(t, res.err)

				require.Len(t, res.block.Transactions(), len(txs))
				for index, tx := range res.block.Transactions() {
					assert.Equal(t, tx.Value(), big.NewInt(int64(index+1)))
				}
			})

			// verify that all transactions from Account1 are prioritized by adding them to the TxPool as local transactions,
			// while transactions from Account2 are added as remote transactions
			t.Run("mixed local & remote transactions", func(t *testing.T) {
				miner := createCeloMiner(t)

				var acc1TxNum, acc2TxNum uint64

				miner.txpool.Clear()
				// Add all transactions from account1 as local, and those from account2 as remote
				for _, tx := range txs {
					sender, err := types.Sender(signer, tx)
					require.NoError(t, err)

					isAccount1 := sender == address1
					errs := miner.txpool.Add([]*types.Transaction{tx}, isAccount1, false)
					requireNoErrors(t, errs)

					if isAccount1 {
						acc1TxNum++
					} else {
						acc2TxNum++
					}
				}

				res := miner.generateWork(&generateParams{
					parentHash: parentHeader.Hash(),
					timestamp:  parentHeader.Time + 1,
					random:     common.HexToHash("0xcafebabe"),
					noTxs:      false,
					forceTime:  true,
				}, false)
				require.NoError(t, res.err)

				var acc1TxCount, acc2TxCount uint64

				require.Len(t, res.block.Transactions(), len(txs))
				for _, tx := range res.block.Transactions() {
					sender, _ := types.Sender(signer, tx)

					if sender == address1 {
						assert.Equal(t, acc1TxCount, tx.Nonce())
						assert.Zero(t, acc2TxCount, "transactions from Account2 should not be ordered before transactions from Account1")
						acc1TxCount++
					} else {
						assert.Equal(t, acc2TxCount, tx.Nonce())
						assert.Equal(t, acc1TxNum, acc1TxCount, "transactions from Account1 should not be ordered after transactions from Account2")
						acc2TxCount++
					}
				}
			})
		})
	}
}
