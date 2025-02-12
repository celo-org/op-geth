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
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/trie"
)

type celoTestTxData struct {
	feeCurrency *common.Address
	priorityFee int64
	gasLimit    uint64
}

type celoOpTestBackend struct {
	opTestBackend
	rates common.ExchangeRates
}

func (b *celoOpTestBackend) GetExchangeRates(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (common.ExchangeRates, error) {
	return b.rates, nil
}

func newCeloOpTestBackend(t *testing.T, txs []celoTestTxData, rates common.ExchangeRates) *celoOpTestBackend {
	t.Helper()

	var (
		key, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
		signer = types.LatestSigner(params.TestChainConfig)
	)
	// only the most recent block is considered for optimism priority fee suggestions, so this is
	// where we add the test transactions
	ts := []*types.Transaction{}
	rs := []*types.Receipt{}
	header := types.Header{}
	header.GasLimit = blockGasLimit
	var nonce uint64
	for _, tx := range txs {
		var txData types.TxData
		if tx.feeCurrency == nil {
			txData = &types.DynamicFeeTx{
				ChainID:   params.TestChainConfig.ChainID,
				Nonce:     nonce,
				To:        &common.Address{},
				Gas:       params.TxGas,
				GasFeeCap: big.NewInt(100 * params.GWei),
				GasTipCap: big.NewInt(tx.priorityFee),
				Data:      []byte{},
			}
		} else {
			txData = &types.CeloDynamicFeeTxV2{
				ChainID:     params.TestChainConfig.ChainID,
				Nonce:       nonce,
				To:          &common.Address{},
				Gas:         params.TxGas,
				FeeCurrency: tx.feeCurrency,
				GasFeeCap:   big.NewInt(100 * params.GWei),
				GasTipCap:   big.NewInt(tx.priorityFee),
				Data:        []byte{},
			}
		}

		t := types.MustSignNewTx(key, signer, txData)
		ts = append(ts, t)
		r := types.Receipt{}
		r.GasUsed = tx.gasLimit
		header.GasUsed += r.GasUsed
		rs = append(rs, &r)
		nonce++
	}
	hasher := trie.NewStackTrie(nil)
	b := types.NewBlock(&header, &types.Body{Transactions: ts}, nil, hasher, types.DefaultBlockConfig)
	return &celoOpTestBackend{opTestBackend: opTestBackend{block: b, receipts: rs}, rates: rates}
}

// TestCeloSuggestOptimismPriorityFee verifies that the suggested priority fee calculation correctly accounts for
// currency conversion in Cel2. This test is similar to TestSuggestOptimismPriorityFee but focuses
// specifically on the accuracy of fee currency conversions in different transaction scenarios.
func TestCeloSuggestOptimismPriorityFee(t *testing.T) {
	minSuggestion := new(big.Int).SetUint64(1e8 * params.Wei)
	cases := []struct {
		txdata []celoTestTxData
		want   *big.Int
	}{
		{
			// block well under capacity, expect min priority fee suggestion
			txdata: []celoTestTxData{{&core.DevFeeCurrencyAddr, params.GWei, 21000}},
			want:   minSuggestion,
		},
		{
			// 2 txs, still under capacity, expect min priority fee suggestion
			txdata: []celoTestTxData{{nil, params.GWei, 21000}, {&core.DevFeeCurrencyAddr, params.GWei, 21000}},
			want:   minSuggestion,
		},
		{
			// 2 txs w same priority fee (1 gwei), but second tx puts it right over capacity
			txdata: []celoTestTxData{{nil, params.GWei, 21000}, {&core.DevFeeCurrencyAddr, 2 * params.GWei, 21001}},
			want:   big.NewInt(1.1 * params.GWei), // 10 percent over 1 gwei, the median
		},
		{
			// 3 txs, full block. return 10% over the median tx (10 gwei * 10% == 11 gwei)
			txdata: []celoTestTxData{{nil, 10 * params.GWei, 21000}, {nil, 1 * params.GWei, 21000}, {&core.DevFeeCurrencyAddr, 100 * params.GWei, 21000}},
			want:   big.NewInt(11 * params.GWei), // 10 percent over 10 gwei, the median
		},
	}

	rates := common.ExchangeRates{
		core.DevFeeCurrencyAddr: big.NewRat(2, 1),
	}

	for i, c := range cases {
		backend := newCeloOpTestBackend(t, c.txdata, rates)
		oracle := NewOracle(backend, Config{MinSuggestedPriorityFee: minSuggestion}, big.NewInt(params.GWei))
		got := oracle.SuggestOptimismPriorityFee(context.Background(), backend.block.Header(), backend.block.Hash())
		if got.Cmp(c.want) != 0 {
			t.Errorf("Gas price mismatch for test case %d: want %d, got %d", i, c.want, got)
		}
	}
}
