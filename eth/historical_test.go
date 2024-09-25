// Copyright 2016 The go-ethereum Authors
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

package eth

import (
	"context"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/internal/ethapi"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
)

var (
	testBalance = big.NewInt(2e15)
)

var genesis = &core.Genesis{
	Config:    params.Cel2TestConfig(9010),
	Alloc:     types.GenesisAlloc{testAddr: {Balance: testBalance}},
	ExtraData: []byte("test genesis"),
	Timestamp: 9000,
	BaseFee:   big.NewInt(params.InitialBaseFee),
}

// var genesis = &core.Genesis{
// 	Config:    params.OptimismTestConfig,
// 	Alloc:     types.GenesisAlloc{testAddr: {Balance: testBalance}},
// 	ExtraData: []byte("test genesis"),
// 	Timestamp: 9000,
// 	BaseFee:   big.NewInt(params.InitialBaseFee),
// }

// OptimismTestConfig = func() *ChainConfig {
// 	conf := *AllCliqueProtocolChanges // copy the config
// 	conf.Clique = nil
// 	conf.TerminalTotalDifficultyPassed = true
// 	conf.BedrockBlock = big.NewInt(5)
// 	conf.Optimism = &OptimismConfig{EIP1559Elasticity: 50, EIP1559Denominator: 10}
// 	return &conf
// }()

// var depositTx = types.NewTx(&types.DepositTx{
// 	Value: big.NewInt(12),
// 	Gas:   params.TxGas + 2000,
// 	To:    &common.Address{2},
// 	Data:  make([]byte, 500),
// })

var testTx1 = types.MustSignNewTx(testKey, types.LatestSigner(genesis.Config), &types.LegacyTx{
	Nonce:    0,
	Value:    big.NewInt(12),
	GasPrice: big.NewInt(params.InitialBaseFee),
	Gas:      params.TxGas,
	To:       &common.Address{2},
})

var testTx2 = types.MustSignNewTx(testKey, types.LatestSigner(genesis.Config), &types.LegacyTx{
	Nonce:    1,
	Value:    big.NewInt(8),
	GasPrice: big.NewInt(params.InitialBaseFee),
	Gas:      params.TxGas,
	To:       &common.Address{2},
})

func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	if number.Sign() >= 0 {
		return hexutil.EncodeBig(number)
	}
	// It's negative.
	if number.IsInt64() {
		return rpc.BlockNumber(number.Int64()).String()
	}
	// It's negative and large, which is invalid.
	return fmt.Sprintf("<invalid %d>", number)
}

func toCallArg(msg ethereum.CallMsg) interface{} {
	arg := map[string]interface{}{
		"from": msg.From,
		"to":   msg.To,
	}
	if len(msg.Data) > 0 {
		arg["input"] = hexutil.Bytes(msg.Data)
	}
	if msg.Value != nil {
		arg["value"] = (*hexutil.Big)(msg.Value)
	}
	if msg.Gas != 0 {
		arg["gas"] = hexutil.Uint64(msg.Gas)
	}
	if msg.GasPrice != nil {
		arg["gasPrice"] = (*hexutil.Big)(msg.GasPrice)
	}
	if msg.GasFeeCap != nil {
		arg["maxFeePerGas"] = (*hexutil.Big)(msg.GasFeeCap)
	}
	if msg.GasTipCap != nil {
		arg["maxPriorityFeePerGas"] = (*hexutil.Big)(msg.GasTipCap)
	}
	if msg.AccessList != nil {
		arg["accessList"] = msg.AccessList
	}
	if msg.BlobGasFeeCap != nil {
		arg["maxFeePerBlobGas"] = (*hexutil.Big)(msg.BlobGasFeeCap)
	}
	if msg.BlobHashes != nil {
		arg["blobVersionedHashes"] = msg.BlobHashes
	}
	return arg
}

type mockHistoricalBackend struct{}

func (m *mockHistoricalBackend) Call(ctx context.Context, args ethapi.TransactionArgs, blockNrOrHash rpc.BlockNumberOrHash, overrides *ethapi.StateOverride) (hexutil.Bytes, error) {
	num, ok := blockNrOrHash.Number()
	if ok && num == 1 {
		return hexutil.Bytes("test"), nil
	}
	return nil, ethereum.NotFound
}

func (m *mockHistoricalBackend) EstimateGas(ctx context.Context, args ethapi.TransactionArgs, blockNrOrHash *rpc.BlockNumberOrHash) (hexutil.Uint64, error) {
	num, ok := blockNrOrHash.Number()
	if ok && num == 1 {
		return hexutil.Uint64(12345), nil
	}
	return 0, ethereum.NotFound
}

func (m *mockHistoricalBackend) TraceBlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (interface{}, error) {
	if blockNr == 1 {
		return "traceBlockByNumberResult", nil
	}
	return nil, ethereum.NotFound
}

func (m *mockHistoricalBackend) TraceBlockByHash(ctx context.Context, blockHash common.Hash) (interface{}, error) {
	if blockHash == common.HexToHash("0x1") {
		return "traceBlockByHashResult", nil
	}
	return nil, ethereum.NotFound
}

func (m *mockHistoricalBackend) TraceCall(ctx context.Context, args ethapi.TransactionArgs, blockNrOrHash rpc.BlockNumberOrHash, overrides *ethapi.StateOverride) (interface{}, error) {
	num, ok := blockNrOrHash.Number()
	if ok && num == 1 {
		return "traceCallResult", nil
	}
	return nil, ethereum.NotFound
}

func (m *mockHistoricalBackend) TraceTransaction(ctx context.Context, txHash common.Hash) (interface{}, error) {
	if txHash == common.HexToHash("0x1") {
		return "traceTransactionResult", nil
	}
	return nil, ethereum.NotFound
}

func (m *mockHistoricalBackend) GetBalance(ctx context.Context, account common.Address, blockNrOrHash rpc.BlockNumberOrHash) (*hexutil.Big, error) {
	num, ok := blockNrOrHash.Number()
	if ok && num == 1 {
		return (*hexutil.Big)(big.NewInt(1000)), nil
	}
	return nil, ethereum.NotFound
}

func (m *mockHistoricalBackend) GetProof(ctx context.Context, account common.Address, storageKeys []common.Hash, blockNrOrHash rpc.BlockNumberOrHash) (*ethapi.AccountResult, error) {
	num, ok := blockNrOrHash.Number()
	if ok && num == 1 {
		return &ethapi.AccountResult{
			Nonce: hexutil.Uint64(12345),
		}, nil
	}
	return nil, ethereum.NotFound
}

func (m *mockHistoricalBackend) GetCode(ctx context.Context, account common.Address, blockNrOrHash rpc.BlockNumberOrHash) (hexutil.Bytes, error) {
	num, ok := blockNrOrHash.Number()
	if ok && num == 1 {
		return hexutil.Bytes("testGetCode"), nil
	}
	return nil, ethereum.NotFound
}

func (m *mockHistoricalBackend) GetStorageAt(ctx context.Context, account common.Address, key common.Hash, blockNrOrHash rpc.BlockNumberOrHash) (hexutil.Bytes, error) {
	num, ok := blockNrOrHash.Number()
	if ok && num == 1 {
		return hexutil.Bytes("testGetStorageAt"), nil
	}
	return nil, ethereum.NotFound
}

type accessListResult struct {
	Accesslist *types.AccessList `json:"accessList"`
	Error      string            `json:"error,omitempty"`
	GasUsed    hexutil.Uint64    `json:"gasUsed"`
}

func (m *mockHistoricalBackend) CreateAccessList(ctx context.Context, args ethapi.TransactionArgs, blockNrOrHash rpc.BlockNumberOrHash, overrides *ethapi.StateOverride) (*accessListResult, error) {
	num, ok := blockNrOrHash.Number()
	if ok && num == 1 {
		return &accessListResult{
			Accesslist: &types.AccessList{},
			GasUsed:    12345,
		}, nil
	}
	return nil, ethereum.NotFound
}

func (m *mockHistoricalBackend) GetTransactionCount(ctx context.Context, account common.Address, blockNrOrHash rpc.BlockNumberOrHash) (hexutil.Uint64, error) {
	num, ok := blockNrOrHash.Number()
	if ok && num == 1 {
		return hexutil.Uint64(12345), nil
	}
	return 0, ethereum.NotFound
}

func newMockHistoricalBackend(t *testing.T) string {
	s := rpc.NewServer()
	err := node.RegisterApis([]rpc.API{
		{
			Namespace:     "debug",
			Service:       new(mockHistoricalBackend),
			Public:        true,
			Authenticated: false,
		},
		{
			Namespace:     "eth",
			Service:       new(mockHistoricalBackend),
			Public:        true,
			Authenticated: false,
		},
	}, nil, s)
	if err != nil {
		t.Fatalf("error creating mock historical backend: %v", err)
	}

	hdlr := node.NewHTTPHandlerStack(s, []string{"*"}, []string{"*"}, nil)
	mux := http.NewServeMux()
	mux.Handle("/", hdlr)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("error creating mock historical backend listener: %v", err)
	}

	go func() {
		httpS := &http.Server{Handler: mux}
		httpS.Serve(listener)

		t.Cleanup(func() {
			httpS.Shutdown(context.Background())
		})
	}()

	return fmt.Sprintf("http://%s", listener.Addr().String())
}

func newTestBackend(t *testing.T) (*node.Node, []*types.Block) {
	histAddr := newMockHistoricalBackend(t) // TODO(Alec)

	// Generate test chain
	blocks := generateTestChain(genesis, 10)

	// Create node
	n, err := node.New(&node.Config{})
	if err != nil {
		t.Fatalf("can't create new node: %v", err)
	}

	// Create Ethereum Service
	config := &ethconfig.Config{Genesis: genesis}

	config.RollupHistoricalRPC = histAddr
	config.RollupHistoricalRPCTimeout = time.Second * 5
	ethservice, err := New(n, config)
	if err != nil {
		t.Fatalf("can't create new ethereum service: %v", err)
	}

	// Import the test chain.
	if err := n.Start(); err != nil {
		t.Fatalf("can't start test node: %v", err)
	}
	if _, err := ethservice.BlockChain().InsertChain(blocks[1:]); err != nil {
		t.Fatalf("can't import test blocks: %v", err)
	}

	// Ensure the tx indexing is fully generated
	for ; ; time.Sleep(time.Millisecond * 100) {
		progress, err := ethservice.BlockChain().TxIndexProgress()
		if err == nil && progress.Done() {
			break
		}
	}

	return n, blocks
}

func generateTestChain(genesis *core.Genesis, length int) []*types.Block {
	generate := func(i int, g *core.BlockGen) {
		g.OffsetTime(5)
		g.SetExtra([]byte("test"))
		g.SetDifficulty(big.NewInt(0))
		g.BaseFee().Set(genesis.BaseFee)
		// if i == 1 {
		// 	g.AddTx(testTx1)
		// 	g.AddTx(testTx2)
		// }
	}
	_, blocks, _ := core.GenerateChainWithGenesis(genesis, ethash.NewFaker(), length, generate)
	return append([]*types.Block{genesis.ToBlock()}, blocks...)
}

func TestHistoricalRPCService(t *testing.T) {
	backend, _ := newTestBackend(t)
	client := backend.Attach()

	defer backend.Close()
	defer client.Close()

	// _supportedModules, err := client.SupportedModules()
	// if err != nil {
	// 	t.Fatalf("unexpected error fetching modules: %v", err)
	// }
	// log.Info("Supported modules", "modules", _supportedModules)

	// backend.RegisterAPIs([]rpc.API{
	// 	{
	// 		Namespace:     "eth",
	// 		Service:       new(eth.EthAPIBackend),
	// 		Public:        true,
	// 		Authenticated: false,
	// 	},
	// 	{
	// 		Namespace:     "debug",
	// 		Service:       new(eth.EthAPIBackend),
	// 		Public:        true,
	// 		Authenticated: false,
	// 	},
	// })

	tests := map[string]struct {
		args []interface{}
		test func(t *testing.T, client *rpc.Client, method string, args []interface{})
	}{
		"eth_call": {
			args: []interface{}{toCallArg(ethereum.CallMsg{From: testAddr, To: &common.Address{}, Gas: 21000, Value: big.NewInt(1)}), rpc.BlockNumberOrHashWithNumber(1)},
			test: func(t *testing.T, client *rpc.Client, method string, args []interface{}) {
				var result hexutil.Bytes
				err := client.CallContext(context.Background(), &result, method, args...)
				if err != nil {
					t.Fatalf("RPC call %s failed: %v", method, err)
				}
				expectedResult := "test"
				if string(result) != expectedResult {
					t.Errorf("RPC call %s returned unexpected result: got %v, want %v", method, result, expectedResult)
				}
			},
		},
		// "eth_estimateGas": {
		// 	args: []interface{}{toCallArg(ethereum.CallMsg{From: testAddr, To: &common.Address{}, Gas: 21000, Value: big.NewInt(1)}), rpc.BlockNumberOrHashWithNumber(1)},
		// 	test: func(t *testing.T, client *rpc.Client, method string, args []interface{}) {
		// 		var result hexutil.Uint64
		// 		err := client.CallContext(context.Background(), &result, method, args...)
		// 		if err != nil {
		// 			t.Fatalf("RPC call %s failed: %v", method, err)
		// 		}
		// 		expectedResult := uint64(12345)
		// 		if uint64(result) != expectedResult {
		// 			t.Errorf("RPC call %s returned unexpected result: got %v, want %v", method, result.String(), expectedResult)
		// 		}
		// 	},
		// },
		// "eth_getBalance": {
		// 	args: []interface{}{testAddr, rpc.BlockNumberOrHashWithNumber(1)},
		// 	test: func(t *testing.T, client *rpc.Client, method string, args []interface{}) {
		// 		var result hexutil.Big
		// 		err := client.CallContext(context.Background(), &result, method, args...)
		// 		if err != nil {
		// 			t.Fatalf("RPC call %s failed: %v", method, err)
		// 		}
		// 		expectedResult := big.NewInt(1000)
		// 		if result.ToInt().Cmp(expectedResult) != 0 {
		// 			t.Errorf("RPC call %s returned unexpected result: got %v, want %v", method, result.String(), expectedResult)
		// 		}
		// 	},
		// },
		// "eth_getProof": {
		// 	args: []interface{}{testAddr, []common.Hash{}, rpc.BlockNumberOrHashWithNumber(1)},
		// 	test: func(t *testing.T, client *rpc.Client, method string, args []interface{}) {
		// 		var result ethapi.AccountResult
		// 		err := client.CallContext(context.Background(), &result, method, args...)
		// 		if err != nil {
		// 			t.Fatalf("RPC call %s failed: %v", method, err)
		// 		}
		// 		expectedResult := ethapi.AccountResult{
		// 			Nonce: hexutil.Uint64(12345),
		// 		}
		// 		if result.Nonce != expectedResult.Nonce {
		// 			t.Errorf("RPC call %s returned unexpected result: got %v, want %v", method, result, nil)
		// 		}
		// 	},
		// },
		// "eth_getCode": {
		// 	args: []interface{}{testAddr, rpc.BlockNumberOrHashWithNumber(1)},
		// 	test: func(t *testing.T, client *rpc.Client, method string, args []interface{}) {
		// 		var result hexutil.Bytes
		// 		err := client.CallContext(context.Background(), &result, method, args...)
		// 		if err != nil {
		// 			t.Fatalf("RPC call %s failed: %v", method, err)
		// 		}
		// 		expectedResult := "testGetCode"
		// 		if string(result) != expectedResult {
		// 			t.Errorf("RPC call %s returned unexpected result: got %v, want %v", method, result, expectedResult)
		// 		}
		// 	},
		// },
		// "eth_getStorageAt": {
		// 	args: []interface{}{testAddr, common.Hash{}, rpc.BlockNumberOrHashWithNumber(1)},
		// 	test: func(t *testing.T, client *rpc.Client, method string, args []interface{}) {
		// 		var result hexutil.Bytes
		// 		err := client.CallContext(context.Background(), &result, method, args...)
		// 		if err != nil {
		// 			t.Fatalf("RPC call %s failed: %v", method, err)
		// 		}
		// 		expectedResult := "testGetStorageAt"
		// 		if string(result) != expectedResult {
		// 			t.Errorf("RPC call %s returned unexpected result: got %v, want %v", method, result, expectedResult)
		// 		}
		// 	},
		// },
		// "eth_createAccessList": {
		// 	args: []interface{}{toCallArg(ethereum.CallMsg{From: testAddr, To: &common.Address{}, Gas: 21000, Value: big.NewInt(1)}), rpc.BlockNumberOrHashWithNumber(1)},
		// 	test: func(t *testing.T, client *rpc.Client, method string, args []interface{}) {
		// 		var result accessListResult
		// 		err := client.CallContext(context.Background(), &result, method, args...)
		// 		if err != nil {
		// 			t.Fatalf("RPC call %s failed: %v", method, err)
		// 		}
		// 		if result.GasUsed != 12345 {
		// 			t.Errorf("RPC call %s returned unexpected result: got %v, want %v", method, result, nil)
		// 		}
		// 	},
		// },
		// "eth_getTransactionCount": {
		// 	args: []interface{}{testAddr, rpc.BlockNumberOrHashWithNumber(1)},
		// 	test: func(t *testing.T, client *rpc.Client, method string, args []interface{}) {
		// 		var result hexutil.Uint64
		// 		err := client.CallContext(context.Background(), &result, method, args...)
		// 		if err != nil {
		// 			t.Fatalf("RPC call %s failed: %v", method, err)
		// 		}
		// 		expectedResult := hexutil.Uint64(12345)
		// 		if result != expectedResult {
		// 			t.Errorf("RPC call %s returned unexpected result: got %v, want %v", method, result.String(), expectedResult)
		// 		}
		// 	},
		// },
		// "debug_traceBlockByNumber": {
		// 	args: []interface{}{rpc.BlockNumber(1), tracers.TraceConfig{}},
		// 	test: func(t *testing.T, client *rpc.Client, method string, args []interface{}) {
		// 		var result interface{}

		// 		err := client.CallContext(context.Background(), &result, method, args...)
		// 		if err != nil {
		// 			t.Fatalf("RPC call %s failed: %v", method, err)
		// 		}
		// 		if result != nil {
		// 			t.Errorf("RPC call %s returned unexpected result: got %v, want %v", method, result, nil)
		// 		}
		// 	},
		// },
		// "debug_traceBlockByHash": {
		// 	args: []interface{}{common.HexToHash("0x1")},
		// 	test: func(t *testing.T, client *rpc.Client, method string, args []interface{}) {
		// 		var result interface{}
		// 		err := client.CallContext(context.Background(), &result, method, args...)
		// 		if err != nil {
		// 			t.Fatalf("RPC call %s failed: %v", method, err)
		// 		}
		// 		if result != nil {
		// 			t.Errorf("RPC call %s returned unexpected result: got %v, want %v", method, result, nil)
		// 		}
		// 	},
		// },
		// "debug_traceCall": {
		// 	args: []interface{}{toCallArg(ethereum.CallMsg{From: testAddr, To: &common.Address{}, Gas: 21000, Value: big.NewInt(1)}), rpc.BlockNumberOrHashWithNumber(1)},
		// 	test: func(t *testing.T, client *rpc.Client, method string, args []interface{}) {
		// 		var result interface{}
		// 		err := client.CallContext(context.Background(), &result, method, args...)
		// 		if err != nil {
		// 			t.Fatalf("RPC call %s failed: %v", method, err)
		// 		}
		// 		if result != nil {
		// 			t.Errorf("RPC call %s returned unexpected result: got %v, want %v", method, result, nil)
		// 		}
		// 	},
		// },
		// "debug_traceTransaction": {
		// 	args: []interface{}{common.HexToHash("0x1")},
		// 	test: func(t *testing.T, client *rpc.Client, method string, args []interface{}) {
		// 		var result interface{}
		// 		err := client.CallContext(context.Background(), &result, method, args...)
		// 		if err != nil {
		// 			t.Fatalf("RPC call %s failed: %v", method, err)
		// 		}
		// 		if result != nil {
		// 			t.Errorf("RPC call %s returned unexpected result: got %v, want %v", method, result, nil)
		// 		}
		// 	},
		// },
	}

	for method, tt := range tests {
		t.Run(method, func(t *testing.T) {
			tt.test(t, client, method, tt.args)
		})
	}
}
