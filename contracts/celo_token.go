package contracts

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/contracts/addresses"
	"github.com/ethereum/go-ethereum/contracts/celo/abigen"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/log"
	"github.com/holiman/uint256"
)

var celoTokenABI *abi.ABI

// The `depositAmount` gascost was traced at 29465.
// doubling the allowed gas for this internal call
// will give it some slack
var maxGasForWithdrawAmount uint64 = 30 * Thousand

func init() {
	var err error
	celoTokenABI, err = abigen.CeloTokenMetaData.GetAbi()
	if err != nil {
		panic(err)
	}
}

// Debits transaction fees from the transaction sender and stores them in the temporary address
func DepositAmount(evm *vm.EVM, amount *uint256.Int) error {
	if amount.CmpUint64(0) == 0 {
		return nil
	}

	tokenAddress := addresses.CeloTokenAddress
	if evm.ChainConfig().ChainID != nil && evm.ChainConfig().ChainID.Uint64() == addresses.AlfajoresChainID {
		tokenAddress = addresses.CeloTokenAlfajoresAddress
	}

	gasCallLimit := 2 * maxGasForWithdrawAmount
	leftoverGas, err := evm.CallWithABI(
		celoTokenABI, "depositAmount", tokenAddress, gasCallLimit,
		// depositAmount(uint256 _withdrawAmount) parameters
		amount,
	)
	if err != nil {
		err = fmt.Errorf("depositAmount call reverted: %w", err)
	}
	gasUsed := gasCallLimit - leftoverGas
	if gasUsed > maxGasForWithdrawAmount {
		log.Warn("DepositAmount used more gas than expected", "max-exptected-gas", maxGasForWithdrawAmount, "gas-used", gasUsed)
	}

	log.Trace("DepositAmount called", "amount", *amount, "gasUsed", gasUsed)
	return err
}
