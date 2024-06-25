package contracts

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/contracts/addresses"
	"github.com/ethereum/go-ethereum/contracts/celo/abigen"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/log"
	"github.com/holiman/uint256"
)

const (
	IntrinsicGasForDepositAmount  uint64 = 50 * Thousand
	maxAllowedGasForDepositAmount uint64 = 3 * IntrinsicGasForAlternativeFeeCurrency
)

var celoTokenABI *abi.ABI

func init() {
	var err error
	celoTokenABI, err = abigen.CeloTokenMetaData.GetAbi()
	if err != nil {
		panic(err)
	}
}

func DepositAmount(evm *vm.EVM, amount *uint256.Int) error {
	leftoverGas, err := evm.CallWithABI(
		celoTokenABI, "depositAmount", addresses.CeloTokenAddress, maxAllowedGasForDepositAmount,
		// depositAmount(uint256 _depositAmount) parameters
		amount,
	)
	gasUsed := maxAllowedGasForDepositAmount - leftoverGas
	// TODO: Should we keep track of gasUsed ?
	evm.Context.GasUsedForDebit = gasUsed
	log.Trace("DepositAmount called", "amount", amount, "gasUsed", gasUsed)
	return err
}
