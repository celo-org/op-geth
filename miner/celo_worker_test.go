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
	"fmt"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
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

		miner       = createCeloMiner(t)
		signer      = types.LatestSigner(miner.chainConfig)
		parentBlock = miner.chain.CurrentBlock()
	)

	// BaseFee: 0.875 GWei
	txs := []*types.Transaction{
		// Effective: 100 GWei
		types.MustSignNewTx(key1, signer, &types.LegacyTx{
			Nonce:    0,
			To:       &common.ZeroAddress,
			Value:    big.NewInt(1),
			Gas:      71000,
			GasPrice: big.NewInt(100 * params.GWei),
		}),
		// Effective: 99 GWei
		types.MustSignNewTx(key2, signer, &types.LegacyTx{
			Nonce:    0,
			To:       &common.ZeroAddress,
			Value:    big.NewInt(2),
			Gas:      71000,
			GasPrice: big.NewInt(99 * params.GWei),
		}),
		// Effective: 98.875 GWei
		types.MustSignNewTx(key2, signer, &types.DynamicFeeTx{
			ChainID:   miner.chainConfig.ChainID,
			Nonce:     1,
			To:        &common.ZeroAddress,
			Value:     big.NewInt(3),
			Gas:       71000,
			GasFeeCap: big.NewInt(10000 * params.GWei),
			GasTipCap: big.NewInt(98 * params.GWei),
		}),
		// Effective: 97.875 GWei
		types.MustSignNewTx(key1, signer, &types.CeloDynamicFeeTxV2{
			ChainID:     miner.chainConfig.ChainID,
			Nonce:       1,
			To:          &common.ZeroAddress,
			Value:       big.NewInt(4),
			Gas:         71000,
			GasFeeCap:   big.NewInt(10000 * params.GWei),
			GasTipCap:   big.NewInt(194 * params.GWei),
			FeeCurrency: &core.DevFeeCurrencyAddr,
		}),
		// Effective: 90 GWei
		types.MustSignNewTx(key1, signer, &types.AccessListTx{
			Nonce:    2,
			To:       &common.ZeroAddress,
			Value:    big.NewInt(5),
			Gas:      71000,
			GasPrice: big.NewInt(95 * params.GWei),
		}),
		// Effective: 50 GWei
		types.MustSignNewTx(key2, signer, &types.LegacyTx{
			Nonce:    2,
			To:       &common.ZeroAddress,
			Value:    big.NewInt(6),
			Gas:      71000,
			GasPrice: big.NewInt(50 * params.GWei),
		}),
		// Effective: 200 GWei
		types.MustSignNewTx(key2, signer, &types.LegacyTx{
			Nonce:    3,
			To:       &common.ZeroAddress,
			Value:    big.NewInt(7),
			Gas:      71000,
			GasPrice: big.NewInt(200 * params.GWei),
		}),
	}

	assertNoErrors := func(t *testing.T, errs []error) {
		t.Helper()
		for _, err := range errs {
			require.NoError(t, err)
		}
	}

	// Verify that transaction ordering depends only on nonce and gas price when all transactions in the TxPool are local
	t.Run("all local transactions", func(t *testing.T) {
		miner := createCeloMiner(t)
		txs := shuffle(txs)

		errs := miner.txpool.Add(txs, true, false)
		assertNoErrors(t, errs)

		res := miner.generateWork(&generateParams{
			parentHash: parentBlock.Hash(),
			timestamp:  parentBlock.Time + 1,
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
		txs := shuffle(txs)

		miner.txpool.Clear()
		errs := miner.txpool.Add(txs, false, false)
		assertNoErrors(t, errs)

		res := miner.generateWork(&generateParams{
			parentHash: parentBlock.Hash(),
			timestamp:  parentBlock.Time + 1,
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
		txs := shuffle(txs)

		var acc1TxNum, acc2TxNum uint64

		miner.txpool.Clear()
		// Add all transactions from account1 as local, and those from account2 as remote
		for _, tx := range txs {
			sender, err := types.Sender(signer, tx)
			require.NoError(t, err)

			isAccount1 := sender == address1
			errs := miner.txpool.Add([]*types.Transaction{tx}, isAccount1, false)
			assertNoErrors(t, errs)

			if isAccount1 {
				acc1TxNum++
			} else {
				acc2TxNum++
			}
		}

		res := miner.generateWork(&generateParams{
			parentHash: parentBlock.Hash(),
			timestamp:  parentBlock.Time + 1,
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
				fmt.Printf("address1 %s\n", sender)
				assert.Equal(t, acc1TxCount, tx.Nonce())
				assert.Zero(t, acc2TxCount, "transactions from Account2 should not be ordered before transactions from Account1")
				acc1TxCount++
			} else {
				fmt.Printf("address2 %s\n", sender)
				assert.Equal(t, acc2TxCount, tx.Nonce())
				assert.Equal(t, acc1TxNum, acc1TxCount, "transactions from Account1 should not be ordered after transactions from Account2")
				acc2TxCount++
			}
		}
	})
}

func shuffle[T any](original []T) []T {
	shuffled := make([]T, len(original))
	copy(shuffled, original)

	gen := rand.New(rand.NewSource(time.Now().UnixNano()))
	n := len(shuffled)
	for i := n - 1; i > 0; i-- {
		j := gen.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled
}
