package contracts

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
)

// CeloBackend provide a partial ContractBackend implementation, so that we can
// access core contracts during block processing.
type CeloBackend struct {
	ChainConfig *params.ChainConfig
	State       vm.StateDB
	BlockNumber *big.Int
	Time        uint64
}

// ContractCaller implementation

func (b *CeloBackend) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	return b.State.GetCode(contract), nil
}

func (b *CeloBackend) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	// Ensure message is initialized properly.
	if call.Gas == 0 {
		// Chosen to be the same as ethconfig.Defaults.RPCGasCap
		call.Gas = 50000000
	}
	if call.Value == nil {
		call.Value = new(big.Int)
	}
	if blockNumber == nil {
		blockNumber = b.BlockNumber
	}
	blockCtx := vm.BlockContext{
		BlockNumber: blockNumber,
		Time:        b.Time,
		Random:      &common.Hash{}, // Setting this is important since it is used to set IsMerge
	}
	vmConfig := vm.Config{}

	// While StaticCall does not actually change state, it changes the
	// access lists. We don't want this to add any access list changes, so
	// we do a snapshot+revert.
	snapshot := b.State.Snapshot()
	evm := vm.NewEVM(blockCtx, b.State, b.ChainConfig, vmConfig)
	ret, _, err := evm.StaticCall(evm.Origin, *call.To, call.Data, call.Gas)
	b.State.RevertToSnapshot(snapshot)

	return ret, err
}

// Get a vm.EVM object of you can't use the abi wrapper via the ContractCaller interface
// This is usually the case when executing functions that modify state.
func (b *CeloBackend) NewEVM(feeCurrencyContext *common.FeeCurrencyContext) *vm.EVM {
	blockCtx := vm.BlockContext{
		BlockNumber: b.BlockNumber,
		Time:        b.Time,
		Transfer: func(state vm.StateDB, from common.Address, to common.Address, value *uint256.Int) {
			if value.Cmp(common.U2560) != 0 {
				panic("Non-zero transfers not implemented, yet.")
			}
		},
		Random: &common.Hash{}, // Setting this is important since it is used to set IsMerge
	}
	if feeCurrencyContext != nil {
		blockCtx.FeeCurrencyContext = *feeCurrencyContext
	}
	vmConfig := vm.Config{}
	return vm.NewEVM(blockCtx, b.State, b.ChainConfig, vmConfig)
}
