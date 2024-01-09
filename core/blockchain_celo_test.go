// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	fee_currencies "github.com/ethereum/go-ethereum/contracts"
	contracts "github.com/ethereum/go-ethereum/contracts/celo"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

// TestNativeTransferWithFeeCurrency tests the following:
//
//  1. A transaction whose gasFeeCap is greater than the baseFee is valid.
//  2. Gas accounting for celo fee currency transactions is correct.
//  3. Only the transaction's tip will be received by the coinbase.
//  4. The transaction sender pays for both the tip and baseFee.
//  5. The base fee goes to the fee handler.
func TestNativeTransferWithFeeCurrency(t *testing.T) {
	testNativeTransferWithFeeCurrency(t, rawdb.HashScheme)
	testNativeTransferWithFeeCurrency(t, rawdb.PathScheme)
}

func testNativeTransferWithFeeCurrency(t *testing.T, scheme string) {
	var (
		aa     = common.HexToAddress("0x000000000000000000000000000000000000aaaa")
		engine = ethash.NewFaker()

		// A sender who makes transactions, has some funds
		key1, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
		addr1   = crypto.PubkeyToAddress(key1.PublicKey)
		config  = *params.AllEthashProtocolChanges
		funds   = DevBalance
		gspec   = &Genesis{
			Config: &config,
			Alloc:  celoGenesisAccounts(addr1),
		}
	)
	gspec.Config.Cel2Time = uint64ptr(0)

	signer := types.LatestSigner(gspec.Config)

	_, blocks, _ := GenerateChainWithGenesis(gspec, engine, 1, func(i int, b *BlockGen) {
		b.SetCoinbase(common.Address{1})

		txdata := &types.CeloDynamicFeeTx{
			ChainID:     gspec.Config.ChainID,
			Nonce:       0,
			To:          &aa,
			Gas:         100000,
			GasFeeCap:   newGwei(5),
			GasTipCap:   big.NewInt(2),
			Data:        []byte{},
			FeeCurrency: &FeeCurrencyAddr,
		}
		tx := types.NewTx(txdata)
		tx, _ = types.SignTx(tx, signer, key1)

		b.AddTx(tx)
	})
	chain, err := NewBlockChain(rawdb.NewMemoryDatabase(), DefaultCacheConfigWithScheme(scheme), gspec, nil, engine, vm.Config{}, nil, nil)
	if err != nil {
		t.Fatalf("failed to create tester chain: %v", err)
	}
	defer chain.Stop()

	if n, err := chain.InsertChain(blocks); err != nil {
		t.Fatalf("block %d: failed to insert into chain: %v", n, err)
	}

	block := chain.GetBlockByNumber(1)

	// 1+2: Ensure correct gas amount is deducted
	expectedGas := uint64(71000)
	if block.GasUsed() != expectedGas {
		t.Fatalf("incorrect amount of gas spent: expected %d, got %d", expectedGas, block.GasUsed())
	}

	state, _ := chain.State()

	backend := CeloBackend{
		chainConfig: chain.chainConfig,
		state:       state,
	}
	exchangeRates, err := getExchangeRates(&backend)
	if err != nil {
		t.Fatal("could not get exchange rates")
	}
	baseFeeInFeeCurrency, _ := fee_currencies.ConvertGoldToCurrency(exchangeRates, &FeeCurrencyAddr, block.BaseFee())
	actual, _ := fee_currencies.GetBalanceOf(&backend, block.Coinbase(), FeeCurrencyAddr)

	// 3: Ensure that miner received only the tx's tip.
	expected := new(big.Int).SetUint64(block.GasUsed() * block.Transactions()[0].GasTipCap().Uint64())
	if actual.Cmp(expected) != 0 {
		t.Fatalf("miner balance incorrect: expected %d, got %d", expected, actual)
	}

	// 4: Ensure the tx sender paid for the gasUsed * (tip + block baseFee).
	actual, _ = fee_currencies.GetBalanceOf(&backend, addr1, FeeCurrencyAddr)
	actual = new(big.Int).Sub(funds, actual)
	expected = new(big.Int).SetUint64(block.GasUsed() * (block.Transactions()[0].GasTipCap().Uint64() + baseFeeInFeeCurrency.Uint64()))
	if actual.Cmp(expected) != 0 {
		t.Fatalf("sender balance incorrect: expected %d, got %d", expected, actual)
	}

	// 5: Check that base fee has been moved to the fee handler.
	actual, _ = fee_currencies.GetBalanceOf(&backend, contracts.FeeHandlerAddress, FeeCurrencyAddr)
	expected = new(big.Int).SetUint64(block.GasUsed() * baseFeeInFeeCurrency.Uint64())
	if actual.Cmp(expected) != 0 {
		t.Fatalf("fee handler balance incorrect: expected %d, got %d", expected, actual)
	}
}