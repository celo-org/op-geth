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

	// Minimal initialization, might need to be extended when CeloBackend
	// is used in more places. Also initializing blockNumber and time with
	// 0 works now, but will break once we add hardforks at a later time.
	if blockNumber == nil {
		blockNumber = common.Big0
	}
	blockCtx := vm.BlockContext{
		BlockNumber: blockNumber,
		Time:        0,
		Random:      &common.Hash{}, // Setting this is important since it is used to set IsMerge
	}
	txCtx := vm.TxContext{}
	vmConfig := vm.Config{}

	// We can use b.State here without copying or making a snapshot because
	// StaticCall won't change the state. It reverts on all state modifying
	// operations.
	// The "touch" that is caused by a StaticCall is not relevant because
	// no Celo chain contains empty accounts. Per EIP-7523, all chains that
	// had the Spurious Dragon fork enabled at genesis don't have empty
	// accounts, making touches irrelevant.
	// We only have to filter out the access list tracking, which is done
	// by ReadOnlyStateDB.
	readOnlyStateDB := ReadOnlyStateDB{StateDB: b.State}
	evm := vm.NewEVM(blockCtx, txCtx, &readOnlyStateDB, b.ChainConfig, vmConfig)
	ret, _, err := evm.StaticCall(vm.AccountRef(evm.Origin), *call.To, call.Data, call.Gas)

	return ret, err
}

// Get a vm.EVM object of you can't use the abi wrapper via the ContractCaller interface
// This is usually the case when executing functions that modify state.
func (b *CeloBackend) NewEVM(feeCurrencyContext *common.FeeCurrencyContext) *vm.EVM {
	blockCtx := vm.BlockContext{
		BlockNumber: new(big.Int),
		Time:        0,
		Transfer: func(state vm.StateDB, from common.Address, to common.Address, value *uint256.Int) {
			if value.Cmp(common.U2560) != 0 {
				panic("Non-zero transfers not implemented, yet.")
			}
		},
	}
	if feeCurrencyContext != nil {
		blockCtx.FeeCurrencyContext = *feeCurrencyContext
	}
	txCtx := vm.TxContext{}
	vmConfig := vm.Config{}
	return vm.NewEVM(blockCtx, txCtx, b.State, b.ChainConfig, vmConfig)
}
