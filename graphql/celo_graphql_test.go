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

package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/exchange"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/beacon"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/eth/filters"
	"github.com/ethereum/go-ethereum/internal/celoapi"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCeloGQLService(t *testing.T, config *params.ChainConfig, stack *node.Node, gspec *core.Genesis, genBlocks int, genfunc func(i int, gen *core.BlockGen)) (*handler, []*types.Block, *eth.Ethereum) {
	ethConf := &ethconfig.Config{
		Genesis:        gspec,
		NetworkId:      gspec.Config.ChainID.Uint64(),
		TrieCleanCache: 5,
		TrieDirtyCache: 5,
		TrieTimeout:    60 * time.Minute,
		SnapshotCache:  5,
		RPCGasCap:      1000000,
		StateScheme:    rawdb.HashScheme,
	}
	var engine consensus.Engine = beacon.New(ethash.NewFaker())

	ethBackend, err := eth.New(stack, ethConf)
	if err != nil {
		t.Fatalf("could not create eth backend: %v", err)
	}
	chain, _ := core.GenerateChain(
		config,
		ethBackend.BlockChain().Genesis(),
		engine,
		ethBackend.ChainDb(),
		genBlocks,
		genfunc,
	)
	_, err = ethBackend.BlockChain().InsertChain(chain)
	if err != nil {
		t.Fatalf("could not create import blocks: %v", err)
	}

	// Set up handler
	filterSystem := filters.NewFilterSystem(ethBackend.APIBackend, filters.Config{})
	celoBackend := celoapi.NewCeloAPIBackend(ethBackend.APIBackend)
	handler, err := newHandler(stack, celoBackend, filterSystem, []string{}, []string{})
	if err != nil {
		t.Fatalf("could not create graphql service: %v", err)
	}
	return handler, chain, ethBackend
}

// TestCeloGraphQLTransactionGasFeePrices validates the gas fee prices of transactions in Cel2 via a GraphQL endpoint
// It tests that fee currency conversion and gas fee computation are done as expected
func TestCeloGraphQLTransactionGasFeePrices(t *testing.T) {
	stack := createNode(t)
	defer stack.Close()
	var (
		key, _  = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
		address = crypto.PubkeyToAddress(key.PublicKey)
		config  = *params.TestChainConfig
		genesis = &core.Genesis{
			Config:     &config,
			GasLimit:   11500000,
			Difficulty: big.NewInt(1048576),
			Alloc:      core.CeloGenesisAccounts(address),
		}
		signer = types.LatestSigner(genesis.Config)
	)
	// Enable all forks from genesis
	config.LondonBlock = big.NewInt(0)
	config.GingerbreadBlock = big.NewInt(0)
	config.ArrowGlacierBlock = big.NewInt(0)
	config.GrayGlacierBlock = big.NewInt(0)
	config.ShanghaiTime = &genesis.Timestamp
	config.CancunTime = &genesis.Timestamp
	config.TerminalTotalDifficultyPassed = true
	config.TerminalTotalDifficulty = common.Big0

	txParams := []struct {
		gasFeeCap *big.Int
		gasTipCap *big.Int
	}{
		// Case 1: GasFeeCap = GasTipCap
		// GasPrice and EffectiveGasPrice should return GasFeeCap
		// EffectiveTip should return (GasFeeCap - BaseFee)
		{
			gasFeeCap: big.NewInt(10 * params.GWei),
			gasTipCap: big.NewInt(10 * params.GWei),
		},
		// Case 2: GasFeeCap >> GasTipCap
		// GasPrice and EffectiveGasPrice should return (GasTipCap + BaseFee)
		// EffectiveGasTip should return GasTipCap
		{
			gasFeeCap: big.NewInt(1000 * params.GWei),
			gasTipCap: big.NewInt(10 * params.GWei),
		},
	}

	_, blocks, backend := newCeloGQLService(t, &config, stack, genesis, 10, func(i int, b *core.BlockGen) {
		b.SetPoS()
		if i != 9 {
			return
		}

		// Insert target transactions into Block #10
		for _, p := range txParams {
			b.AddTx(types.MustSignNewTx(key, signer, &types.CeloDynamicFeeTxV2{
				ChainID:     genesis.Config.ChainID,
				Nonce:       b.TxNonce(address),
				To:          &common.Address{},
				Gas:         80000,
				FeeCurrency: &core.DevFeeCurrencyAddr,
				GasFeeCap:   p.gasFeeCap,
				GasTipCap:   p.gasTipCap,
				Data:        []byte{},
			}))
		}
	})

	// start node
	if err := stack.Start(); err != nil {
		t.Fatalf("could not start node: %v", err)
	}

	query := `{"query": "{block(number:\"0xa\"){transactions{gasPrice,effectiveGasPrice,effectiveTip}}}","variables": null}`

	resp, err := http.Post(fmt.Sprintf("%s/graphql", stack.HTTPEndpoint()), "application/json", strings.NewReader(query))
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	require.NoError(t, err)

	// extract data
	respBody := map[string]interface{}{}
	err = json.Unmarshal(bodyBytes, &respBody)
	require.NoError(t, err)

	data, ok := respBody["data"].(map[string]interface{})
	require.True(t, ok)

	block, ok := data["block"].(map[string]interface{})
	require.True(t, ok)

	transactions, ok := block["transactions"].([]interface{})
	require.True(t, ok)
	require.Len(t, transactions, 2)

	// create expected values
	rates, err := backend.CeloAPIBackend().GetExchangeRates(context.Background(), rpc.BlockNumberOrHashWithNumber(rpc.BlockNumber(10)))
	require.NoError(t, err)

	block10 := blocks[9]
	baseFee := block10.BaseFee()
	baseFeeInCurrency, err := exchange.ConvertCeloToCurrency(rates, &core.DevFeeCurrencyAddr, baseFee)
	require.NoError(t, err)

	expectedValues := []map[string]interface{}{
		{
			"gasPrice":          "0x" + txParams[0].gasFeeCap.Text(16),
			"effectiveGasPrice": "0x" + txParams[0].gasFeeCap.Text(16),
			"effectiveTip":      "0x" + new(big.Int).Sub(txParams[0].gasFeeCap, baseFeeInCurrency).Text(16),
		},
		{
			"gasPrice":          "0x" + new(big.Int).Add(txParams[1].gasTipCap, baseFeeInCurrency).Text(16),
			"effectiveGasPrice": "0x" + new(big.Int).Add(txParams[1].gasTipCap, baseFeeInCurrency).Text(16),
			"effectiveTip":      "0x" + txParams[1].gasTipCap.Text(16),
		},
	}

	for idx, ex := range expectedValues {
		tx, ok := transactions[idx].(map[string]interface{})
		require.True(t, ok)

		for key, expected := range ex {
			actual, ok := tx[key].(string)
			require.True(t, ok, fmt.Sprintf(`tx doesn't have "%s"`, key))
			assert.Equal(t, expected, actual, fmt.Sprintf(`wrong %s, expected=%s, got=%s`, key, expected, actual))
		}
	}
}
