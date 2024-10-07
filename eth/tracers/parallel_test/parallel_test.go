package paralleltest_test

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth/tracers"
	_ "github.com/ethereum/go-ethereum/eth/tracers/js"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	errStateNotFound = errors.New("state not found")
	errBlockNotFound = errors.New("block not found")
)

func TestTraceBlockParallel(t *testing.T) {
	t.Parallel()

	// Initialize test accounts
	accounts := newAccounts(64)
	alloc := make(types.GenesisAlloc)
	for _, a := range accounts {
		alloc[a.addr] = types.Account{Balance: big.NewInt(params.Ether)}
	}
	genesis := &core.Genesis{
		Config: params.TestChainConfig,
		Alloc:  alloc,
	}
	genBlocks := 10
	signer := types.HomesteadSigner{}
	backend := newTestBackend(t, genBlocks, genesis, func(i int, b *core.BlockGen) {
		for _, a := range accounts {
			// Transfer from account[0] to account[1]
			//    value: 1000 wei
			//    fee:   0 wei
			tx, _ := types.SignTx(types.NewTx(&types.LegacyTx{
				Nonce:    uint64(i),
				To:       &a.addr,
				Value:    big.NewInt(1000),
				Gas:      params.TxGas,
				GasPrice: b.BaseFee(),
				Data:     nil}),
				signer, a.key)
			b.AddTx(tx)
		}
	})
	defer backend.chain.Stop()
	api := tracers.NewAPI(backend)
	tracer := "{data: [], fault: function(log) {}, step: function(log) { if(log.op.toString() == 'CALL') this.data.push(log.stack.peek(0)); }, result: function() { return this.data; }}"
	config := &tracers.TraceConfig{
		Tracer: &tracer,
	}
	for i := 1; i <= genBlocks; i++ {
		result, err := api.TraceBlockByNumber(context.Background(), rpc.BlockNumber(i), config)
		require.NoError(t, err)
		require.NotNil(t, result)
	}
}

func newAccounts(n int) (accounts []Account) {
	for i := 0; i < n; i++ {
		key, _ := crypto.GenerateKey()
		addr := crypto.PubkeyToAddress(key.PublicKey)
		accounts = append(accounts, Account{key: key, addr: addr})
	}
	slices.SortFunc(accounts, func(a, b Account) int { return a.addr.Cmp(b.addr) })
	return accounts
}

type Account struct {
	key  *ecdsa.PrivateKey
	addr common.Address
}

// newTestBackend creates a new test backend. OBS: After test is done, teardown must be
// invoked in order to release associated resources.
func newTestBackend(t *testing.T, n int, gspec *core.Genesis, generator func(i int, b *core.BlockGen)) *testBackend {
	mock := new(mockHistoricalBackend)
	historicalAddr := newMockHistoricalBackend(t, mock)

	historicalClient, err := rpc.Dial(historicalAddr)
	if err != nil {
		t.Fatalf("error making historical client: %v", err)
	}

	backend := &testBackend{
		chainConfig:    gspec.Config,
		engine:         ethash.NewFaker(),
		chaindb:        rawdb.NewMemoryDatabase(),
		historical:     historicalClient,
		mockHistorical: mock,
	}
	// Generate blocks for testing
	_, blocks, _ := core.GenerateChainWithGenesis(gspec, backend.engine, n, generator)

	// Import the canonical chain
	cacheConfig := &core.CacheConfig{
		TrieCleanLimit:    256,
		TrieDirtyLimit:    256,
		TrieTimeLimit:     5 * time.Minute,
		SnapshotLimit:     0,
		TrieDirtyDisabled: true, // Archive mode
	}
	chain, err := core.NewBlockChain(backend.chaindb, cacheConfig, gspec, nil, backend.engine, vm.Config{}, nil, nil)
	if err != nil {
		t.Fatalf("failed to create tester chain: %v", err)
	}
	if n, err := chain.InsertChain(blocks); err != nil {
		t.Fatalf("block %d: failed to insert into chain: %v", n, err)
	}
	backend.chain = chain
	return backend
}

func (b *testBackend) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	return b.chain.GetHeaderByHash(hash), nil
}

func (b *testBackend) HeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Header, error) {
	if number == rpc.PendingBlockNumber || number == rpc.LatestBlockNumber {
		return b.chain.CurrentHeader(), nil
	}
	return b.chain.GetHeaderByNumber(uint64(number)), nil
}

func (b *testBackend) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	return b.chain.GetBlockByHash(hash), nil
}

func (b *testBackend) BlockByNumber(ctx context.Context, number rpc.BlockNumber) (*types.Block, error) {
	if number == rpc.PendingBlockNumber || number == rpc.LatestBlockNumber {
		return b.chain.GetBlockByNumber(b.chain.CurrentBlock().Number.Uint64()), nil
	}
	return b.chain.GetBlockByNumber(uint64(number)), nil
}

func (b *testBackend) GetTransaction(ctx context.Context, txHash common.Hash) (bool, *types.Transaction, common.Hash, uint64, uint64, error) {
	tx, hash, blockNumber, index := rawdb.ReadTransaction(b.chaindb, txHash)
	return tx != nil, tx, hash, blockNumber, index, nil
}

func (b *testBackend) RPCGasCap() uint64 {
	return 25000000
}

func (b *testBackend) ChainConfig() *params.ChainConfig {
	return b.chainConfig
}

func (b *testBackend) Engine() consensus.Engine {
	return b.engine
}

func (b *testBackend) ChainDb() ethdb.Database {
	return b.chaindb
}

func (b *testBackend) StateAtBlock(ctx context.Context, block *types.Block, reexec uint64, base *state.StateDB, readOnly bool, preferDisk bool) (*state.StateDB, tracers.StateReleaseFunc, error) {
	statedb, err := b.chain.StateAt(block.Root())
	if err != nil {
		return nil, nil, errStateNotFound
	}
	if b.refHook != nil {
		b.refHook()
	}
	release := func() {
		if b.relHook != nil {
			b.relHook()
		}
	}
	return statedb, release, nil
}

func (b *testBackend) StateAtTransaction(ctx context.Context, block *types.Block, txIndex int, reexec uint64) (*types.Transaction, vm.BlockContext, *state.StateDB, tracers.StateReleaseFunc, error) {
	parent := b.chain.GetBlock(block.ParentHash(), block.NumberU64()-1)
	if parent == nil {
		return nil, vm.BlockContext{}, nil, nil, errBlockNotFound
	}
	statedb, release, err := b.StateAtBlock(ctx, parent, reexec, nil, true, false)
	if err != nil {
		return nil, vm.BlockContext{}, nil, nil, errStateNotFound
	}
	if txIndex == 0 && len(block.Transactions()) == 0 {
		return nil, vm.BlockContext{}, statedb, release, nil
	}
	// Recompute transactions up to the target index.
	signer := types.MakeSigner(b.chainConfig, block.Number(), block.Time())
	for idx, tx := range block.Transactions() {
		context := core.NewEVMBlockContext(block.Header(), b.chain, nil, b.chainConfig, statedb)
		msg, _ := core.TransactionToMessage(tx, signer, block.BaseFee(), context.FeeCurrencyContext.ExchangeRates)
		txContext := core.NewEVMTxContext(msg)
		if idx == txIndex {
			return tx, context, statedb, release, nil
		}
		vmenv := vm.NewEVM(context, txContext, statedb, b.chainConfig, vm.Config{})
		if _, err := core.ApplyMessage(vmenv, msg, new(core.GasPool).AddGas(tx.Gas())); err != nil {
			return nil, vm.BlockContext{}, nil, nil, fmt.Errorf("transaction %#x failed: %v", tx.Hash(), err)
		}
		statedb.Finalise(vmenv.ChainConfig().IsEIP158(block.Number()))
	}
	return nil, vm.BlockContext{}, nil, nil, fmt.Errorf("transaction index %d out of range for block %#x", txIndex, block.Hash())
}

func (b *testBackend) HistoricalRPCService() *rpc.Client {
	return b.historical
}

type testBackend struct {
	chainConfig *params.ChainConfig
	engine      consensus.Engine
	chaindb     ethdb.Database
	chain       *core.BlockChain

	refHook func() // Hook is invoked when the requested state is referenced
	relHook func() // Hook is invoked when the requested state is released

	historical     *rpc.Client
	mockHistorical *mockHistoricalBackend
}

type mockHistoricalBackend struct {
	mock.Mock
}

// mockHistoricalBackend does not have a TraceCall, because pre-bedrock there is no debug_traceCall available

func (m *mockHistoricalBackend) TraceBlockByNumber(ctx context.Context, number rpc.BlockNumber, config *tracers.TraceConfig) ([]*tracers.TxTraceResult, error) {
	ret := m.Mock.MethodCalled("TraceBlockByNumber", number, config)
	return ret[0].([]*tracers.TxTraceResult), *ret[1].(*error)
}

func (m *mockHistoricalBackend) ExpectTraceBlockByNumber(number rpc.BlockNumber, config *tracers.TraceConfig, out []*tracers.TxTraceResult, err error) {
	m.Mock.On("TraceBlockByNumber", number, config).Once().Return(out, &err)
}

func (m *mockHistoricalBackend) TraceTransaction(ctx context.Context, hash common.Hash, config *tracers.TraceConfig) (interface{}, error) {
	ret := m.Mock.MethodCalled("TraceTransaction", hash, config)
	return ret[0], *ret[1].(*error)
}

func (m *mockHistoricalBackend) ExpectTraceTransaction(hash common.Hash, config *tracers.TraceConfig, out interface{}, err error) {
	jsonOut, _ := json.Marshal(out)
	m.Mock.On("TraceTransaction", hash, config).Once().Return(json.RawMessage(jsonOut), &err)
}

func newMockHistoricalBackend(t *testing.T, backend *mockHistoricalBackend) string {
	s := rpc.NewServer()
	err := node.RegisterApis([]rpc.API{
		{
			Namespace:     "debug",
			Service:       backend,
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
