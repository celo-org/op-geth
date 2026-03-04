// Copyright 2024 The celo Authors
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

package gasprice

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/exchange"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/beacon"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/contracts"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type celoTestBackend struct {
	testBackend
	rates common.ExchangeRates
}

func (b *celoTestBackend) GetExchangeRates(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (common.ExchangeRates, error) {
	return b.rates, nil
}

func newCeloTestBackend(t *testing.T, pending bool, genBlock func(i int, b *core.BlockGen, gspec *core.Genesis, signer types.Signer, key *ecdsa.PrivateKey, address common.Address)) *celoTestBackend {
	var (
		key, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
		addr   = crypto.PubkeyToAddress(key.PublicKey)

		config = *params.TestChainConfig
		gspec  = &core.Genesis{
			Config: &config,
			Alloc:  core.CeloGenesisAccounts(addr),
		}
		signer = types.LatestSigner(gspec.Config)
	)
	// Enable all forks from genesis
	config.LondonBlock = big.NewInt(0)
	config.GingerbreadBlock = big.NewInt(0)
	config.ArrowGlacierBlock = big.NewInt(0)
	config.GrayGlacierBlock = big.NewInt(0)
	config.ShanghaiTime = &gspec.Timestamp

	var engine consensus.Engine = beacon.New(ethash.NewFaker())
	td := params.GenesisDifficulty.Uint64()

	// Generate testing blocks
	db, blocks, _ := core.GenerateChainWithGenesis(gspec, engine, testHead+1, func(i int, b *core.BlockGen) {
		b.SetPoS()
		genBlock(i, b, gspec, signer, key, addr)
		td += b.Difficulty().Uint64()
	})

	gspec.Config.TerminalTotalDifficulty = new(big.Int).SetUint64(td)
	chain, err := core.NewBlockChain(db, gspec, engine, nil)
	if err != nil {
		t.Fatalf("failed to create local chain, %v", err)
	}
	if i, err := chain.InsertChain(blocks); err != nil {
		t.Fatalf("error inserting block %d: %v", i, err)
	}

	chain.SetFinalized(chain.GetBlockByNumber(25).Header())
	chain.SetSafe(chain.GetBlockByNumber(25).Header())

	state, _ := chain.State()
	head := chain.CurrentBlock()

	backend := contracts.CeloBackend{
		ChainConfig: &config,
		State:       state,
		BlockNumber: head.Number,
		Time:        head.Time,
	}
	exchangeRates, err := contracts.GetExchangeRates(&backend)
	if err != nil {
		t.Fatal("could not get exchange rates")
	}

	return &celoTestBackend{
		testBackend: testBackend{chain: chain, pending: pending},
		rates:       exchangeRates,
	}
}

// TestCeloFeeHistory verifies that the fee history calculation correctly accounts for currency conversion
// in Cel2. This test is similar to TestFeeHistory but focuses specifically on verifying the
// accuracy of fee currency conversions.
func TestCeloFeeHistory(t *testing.T) {
	t.Parallel()

	feeCurrencies := []*common.Address{nil, &core.DevFeeCurrencyAddr, &core.DevFeeCurrencyAddr2}
	feeCaps := []*big.Int{
		big.NewInt(10 * params.GWei),
		big.NewInt(30 * params.GWei), // 15 GWei in Celo
		big.NewInt(50 * params.GWei), // 100 GWei in Celo
	}

	backend := newCeloTestBackend(t, false, func(i int, b *core.BlockGen, gspec *core.Genesis, signer types.Signer, key *ecdsa.PrivateKey, address common.Address) {
		addTx := func(gasFeeCap, gasTipCap *big.Int, feeCurrency *common.Address) {
			if feeCurrency == nil {
				b.AddTx(types.MustSignNewTx(key, signer, &types.DynamicFeeTx{
					ChainID:   gspec.Config.ChainID,
					Nonce:     b.TxNonce(address),
					To:        &common.Address{},
					Gas:       80000,
					GasFeeCap: gasFeeCap,
					GasTipCap: gasTipCap,
					Data:      []byte{},
				}))
			} else {
				b.AddTx(types.MustSignNewTx(key, signer, &types.CeloDynamicFeeTxV2{
					ChainID:     gspec.Config.ChainID,
					Nonce:       b.TxNonce(address),
					To:          &common.Address{},
					Gas:         80000,
					FeeCurrency: feeCurrency,
					GasFeeCap:   gasFeeCap,
					GasTipCap:   gasTipCap,
					Data:        []byte{},
				}))
			}
		}

		if i == 19 {
			// Bulding Block #20
			// Set sufficiently high gas fee cap so that the gas tip cap is used
			for idx := range feeCurrencies {
				addTx(big.NewInt(10000*params.GWei), feeCaps[idx], feeCurrencies[idx])
			}
		} else if i == 20 {
			// Bulding Block #21
			// Set fee cap and tip cap to the same amount so that (fee cap - base fee) is used
			for idx := range feeCurrencies {
				addTx(feeCaps[idx], feeCaps[idx], feeCurrencies[idx])
			}
		}
	})

	oracle := NewOracle(backend, Config{
		MaxHeaderHistory: 1000,
		MaxBlockHistory:  1000,
	}, nil)
	first, reward, baseFee, _, _, _, err := oracle.FeeHistory(context.Background(), 2, 21, []float64{0, 50, 100})
	backend.teardown()

	require.NoError(t, err)
	require.Equal(t, big.NewInt(20), first)

	rates, err := backend.GetExchangeRates(context.Background(), rpc.BlockNumberOrHashWithNumber(rpc.BlockNumber(21)))
	require.NotNil(t, rates)
	require.NoError(t, err)

	// create expected values
	expectedReward20 := make([]*big.Int, 3)
	expectedReward21 := make([]*big.Int, 3)
	for idx := range expectedReward20 {
		expectedReward20[idx], err = exchange.ConvertCurrencyToCelo(rates, feeCurrencies[idx], feeCaps[idx])
		require.NoError(t, err)

		baseFeeInCurrency, err := exchange.ConvertCeloToCurrency(rates, feeCurrencies[idx], baseFee[1])
		require.NoError(t, err)

		expectedReward21[idx], err = exchange.ConvertCurrencyToCelo(
			rates,
			feeCurrencies[idx],
			new(big.Int).Sub(feeCaps[idx], baseFeeInCurrency),
		)
		require.NoError(t, err)
	}

	assert.Equal(t, expectedReward20, reward[0])
	assert.Equal(t, expectedReward21, reward[1])
}
