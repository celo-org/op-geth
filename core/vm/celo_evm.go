package vm

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// Call function from ABI and check revert message after call.
// ABIs can be found at contracts/celo/abigen, e.g. abigen.FeeCurrencyMetaData.GetAbi().
// args are passed through to the EVM function.
func (evm *EVM) CallWithABI(contractABI *abi.ABI, funcName string, addr common.Address, gas uint64, args ...interface{}) (leftOverGas uint64, err error) {
	caller := AccountRef(common.ZeroAddress)
	input, err := contractABI.Pack(funcName, args...)
	if err != nil {
		return 0, fmt.Errorf("pack %s: %w", funcName, err)
	}

	ret, leftOverGas, err := evm.Call(caller, addr, input, gas, big.NewInt(0))
	if err != nil {
		revertReason, err2 := abi.UnpackRevert(ret)
		if err2 == nil {
			return 0, fmt.Errorf("%s reverted: %s", funcName, revertReason)
		}
	}

	return leftOverGas, err
}
