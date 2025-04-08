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

package miner

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/exchange"
	"github.com/ethereum/go-ethereum/consensus/beacon"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/txpool"
	"github.com/ethereum/go-ethereum/core/txpool/legacypool"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	key, _  = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	address = crypto.PubkeyToAddress(key.PublicKey)
)

func createCeloMiner(t *testing.T) *Miner {
	var (
		config  = *params.TestChainConfig
		genesis = &core.Genesis{
			Config:     &config,
			GasLimit:   11500000,
			Difficulty: big.NewInt(1048576),
			Alloc:      core.CeloGenesisAccounts(address),
			Timestamp:  uint64(time.Now().Unix()),
		}
		minerConfig = Config{
			PendingFeeRecipient:                   common.HexToAddress("123456789"),
			RollupTransactionConditionalRateLimit: params.TransactionConditionalMaxCost,
			FeeCurrencyLimits: map[common.Address]float64{
				core.DevFeeCurrencyAddr: 1e9,
			},
		}
	)
	// Enable all forks from genesis
	config.LondonBlock = big.NewInt(0)
	config.GingerbreadBlock = big.NewInt(0)
	config.ArrowGlacierBlock = big.NewInt(0)
	config.GrayGlacierBlock = big.NewInt(0)
	config.ShanghaiTime = &genesis.Timestamp
	config.TerminalTotalDifficulty = common.Big0

	engine := beacon.New(ethash.NewFaker())
	chainDB := rawdb.NewMemoryDatabase()
	triedb := triedb.NewDatabase(chainDB, nil)

	chainConfig, _, _, err := core.SetupGenesisBlock(chainDB, triedb, genesis)
	if err != nil {
		t.Fatalf("can't create new chain config: %v", err)
	}

	// Create Ethereum backend
	bc, err := core.NewBlockChain(chainDB, genesis, engine, nil)
	if err != nil {
		t.Fatalf("can't create new chain %v", err)
	}
	statedb, _ := state.New(bc.Genesis().Root(), bc.StateCache())
	blockchain := &testBlockChain{bc.Genesis().Root(), chainConfig, statedb, 10000000, new(event.Feed)}

	pool := legacypool.New(testTxPoolConfig, blockchain)
	txpool, _ := txpool.New(testTxPoolConfig.PriceLimit, blockchain, []txpool.SubPool{pool}, nil)

	// Create Miner
	backend := NewMockBackend(bc, txpool, false, nil)

	return New(backend, minerConfig, engine)
}

// TestMinerFeeCalculationWithCurrencyConversion verifies whether the transaction fee calculation
// for miners correctly considers the currency of the transaction fee
func TestMinerFeeCalculationWithCurrencyConversion(t *testing.T) {
	miner := createCeloMiner(t)
	rates := common.ExchangeRates{
		core.DevFeeCurrencyAddr: big.NewRat(2, 1),
	}

	signer := types.LatestSigner(miner.chainConfig)
	tx1 := types.MustSignNewTx(key, signer, &types.CeloDynamicFeeTxV2{
		Nonce:       0,
		To:          &testUserAddress,
		Value:       big.NewInt(1),
		Gas:         71000,
		FeeCurrency: &core.DevFeeCurrencyAddr,
		GasFeeCap:   big.NewInt(10 * params.GWei),
		GasTipCap:   big.NewInt(10 * params.GWei),
	})
	tx2 := types.MustSignNewTx(key, signer, &types.CeloDynamicFeeTxV2{
		Nonce:       1,
		To:          &testUserAddress,
		Value:       big.NewInt(1),
		Gas:         71000,
		FeeCurrency: &core.DevFeeCurrencyAddr,
		GasFeeCap:   big.NewInt(100 * params.GWei),
		GasTipCap:   big.NewInt(10 * params.GWei),
	})
	txs := types.Transactions{tx1, tx2}

	parentBlock := miner.chain.CurrentBlock()
	// Add transactions & request block
	miner.txpool.Add(txs, true)
	r := miner.generateWork(&generateParams{
		parentHash: parentBlock.Hash(),
		timestamp:  parentBlock.Time + 1,
		random:     common.HexToHash("0xcafebabe"),
		noTxs:      false,
		forceTime:  true,
	}, false)

	// Make sure the transactions are finalized
	require.Equal(t, len(txs), len(r.block.Transactions()), "block should have 2 transactions")
	require.False(t, tx1.Rejected(), "tx1 should not be rejected")
	require.False(t, tx2.Rejected(), "tx2 should not be rejected")

	// Calculate expected values
	baseFee := r.block.BaseFee()
	baseFeeInCurrency, err := exchange.ConvertCeloToCurrency(rates, &core.DevFeeCurrencyAddr, baseFee)
	require.NoError(t, err)

	fee1, err := exchange.ConvertCurrencyToCelo(rates, &core.DevFeeCurrencyAddr,
		new(big.Int).Sub(tx1.GasFeeCap(), baseFeeInCurrency),
	)
	require.NoError(t, err)
	fee2, err := exchange.ConvertCurrencyToCelo(rates, &core.DevFeeCurrencyAddr,
		tx2.GasTipCap(),
	)
	require.NoError(t, err)

	expectedMinerFee := new(big.Int).Add(
		new(big.Int).Mul(fee1, big.NewInt(int64(r.receipts[0].GasUsed))),
		new(big.Int).Mul(fee2, big.NewInt(int64(r.receipts[1].GasUsed))),
	)
	assert.Equal(t, expectedMinerFee, r.fees)
}

func TestBlocklistOnlyForSequencing(t *testing.T) {
	miner := createCeloMiner(t)

	signer := types.LatestSigner(miner.chainConfig)
	parentBlock := miner.chain.CurrentBlock()
	miner.feeCurrencyBlocklist.Add(core.DevFeeCurrencyAddr, *parentBlock)

	var nonce uint64 = 0
	makeTx := func(addToPool bool) *types.Transaction {
		tx := types.MustSignNewTx(key, signer, &types.CeloDynamicFeeTxV2{
			Nonce:       nonce,
			To:          &testUserAddress,
			Value:       big.NewInt(1),
			Gas:         71000,
			FeeCurrency: &core.DevFeeCurrencyAddr,
			GasFeeCap:   big.NewInt(100 * params.GWei),
			GasTipCap:   big.NewInt(10 * params.GWei),
		})
		assert.NotNil(t, parentBlock)
		if addToPool {
			miner.txpool.Add(types.Transactions{tx}, true)
		}
		return tx
	}

	tx := makeTx(true)
	require.True(t, miner.feeCurrencyBlocklist.IsBlocked(*tx.FeeCurrency(), parentBlock))

	// pending block building, blocklist disabled
	r := miner.generateWork(&generateParams{
		parentHash: parentBlock.Hash(),
		timestamp:  parentBlock.Time + 1,
		random:     common.HexToHash("0xcafebabe"),
		noTxs:      false,
		forceTime:  true,
		isPending:  true,
	}, false)

	// Make sure the transactions are finalized
	require.Equal(t, 1, len(r.block.Transactions()), "block should have 1 transactions")
	require.False(t, tx.Rejected(), "tx should not be rejected")

	// make a transaction that's not in the pool
	tx2 := makeTx(false)

	// add to pool once more to the pool
	makeTx(true)

	// l1 derivation block building, blocklist disabled
	r2 := miner.generateWork(&generateParams{
		parentHash: parentBlock.Hash(),
		timestamp:  parentBlock.Time + 1,
		random:     common.HexToHash("0xcafebabe"),
		noTxs:      true,
		forceTime:  true,
		isPending:  false,
		txs:        types.Transactions{tx2},
	}, false)

	require.NoError(t, r2.err)

	// the mempool tx should not be included since we are omitting the mempool
	require.Equal(t, 1, len(r2.block.Transactions()), "block should have 1 transaction")
	require.Equal(t, 1, len(r2.receipts), "block should have 1 transaction receipts")
	require.False(t, tx2.Rejected(), "tx should not be rejected")

	// mempool block building, blocklist enabled
	// and the tx has a blocked fee-currency
	r3 := miner.generateWork(&generateParams{
		parentHash: parentBlock.Hash(),
		timestamp:  parentBlock.Time + 1,
		random:     common.HexToHash("0xcafebabe"),
		noTxs:      false,
		forceTime:  true,
		isPending:  false,
	}, false)
	require.NoError(t, r3.err)

	// third tx should not be included since it's using a blocked fee-currency
	require.Equal(t, 0, len(r3.block.Transactions()), "block should have 0 transactions")
}
