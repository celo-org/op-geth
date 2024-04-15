package contracts

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
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
	blockCtx := vm.BlockContext{BlockNumber: blockNumber, Time: 0}
	txCtx := vm.TxContext{}
	vmConfig := vm.Config{}

	readOnlyStateDB := ReadOnlyStateDB{StateDB: b.State}
	evm := vm.NewEVM(blockCtx, txCtx, &readOnlyStateDB, b.ChainConfig, vmConfig)
	ret, _, err := evm.StaticCall(vm.AccountRef(evm.Origin), *call.To, call.Data, call.Gas)

	return ret, err
}

// Get a vm.EVM object of you can't use the abi wrapper via the ContractCaller interface
// This is usually the case when executing functions that modify state.
func (b *CeloBackend) NewEVM() *vm.EVM {
	blockCtx := vm.BlockContext{BlockNumber: new(big.Int), Time: 0,
		Transfer: func(state vm.StateDB, from common.Address, to common.Address, value *big.Int) {
			if value.Cmp(common.Big0) != 0 {
				panic("Non-zero transfers not implemented, yet.")
			}
		},
	}
	txCtx := vm.TxContext{}
	vmConfig := vm.Config{}
	return vm.NewEVM(blockCtx, txCtx, b.State, b.ChainConfig, vmConfig)
}

// GetFeeBalance returns the account's balance from the specified feeCurrency
// (if feeCurrency is nil or ZeroAddress, native currency balance is returned).
func (b *CeloBackend) GetFeeBalance(account common.Address, feeCurrency *common.Address) *big.Int {
	if feeCurrency == nil || *feeCurrency == common.ZeroAddress {
		return b.State.GetBalance(account)
	}
	balance, err := GetBalanceERC20(b, account, *feeCurrency)
	if err != nil {
		log.Error("Error while trying to get ERC20 balance:", "cause", err, "contract", feeCurrency.Hex(), "account", account.Hex())
	}
	return balance
}
