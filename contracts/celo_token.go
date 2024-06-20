package contracts

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/contracts/addresses"
	"github.com/ethereum/go-ethereum/contracts/celo/abigen"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/log"
	"github.com/holiman/uint256"
)

var celoTokenABI *abi.ABI

// TODO: find an upper limit for this call
// this will mainly prevent implementation errors
// in the L2's CeloToken contract
var maxGasForWithdrawAmount uint64 = 20 * 50 * Thousand

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

	leftoverGas, err := evm.CallWithABI(
		celoTokenABI, "depositAmount", tokenAddress, maxGasForWithdrawAmount,
		// depositAmount(uint256 _withdrawAmount) parameters
		amount,
	)
	gasUsed := maxGasForWithdrawAmount - leftoverGas
	//TODO: how to handle gas?
	log.Trace("DepositAmount called", "amount", *amount, "gasUsed", gasUsed)
	return err
}
