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

// TODO: actually it should be pretty clear how expensive the call is
var maxGasForIncreaseWithdraw uint64 = 3 * 50 * Thousand

func init() {
	var err error
	celoTokenABI, err = abigen.CeloTokenMetaData.GetAbi()
	if err != nil {
		panic(err)
	}
}

// Debits transaction fees from the transaction sender and stores them in the temporary address
func DecreaseWithdrawn(evm *vm.EVM, amount *uint256.Int) error {
	if amount.CmpUint64(0) == 0 {
		return nil
	}

	tokenAddress := addresses.CeloTokenAddress
	if evm.ChainConfig().ChainID != nil && evm.ChainConfig().ChainID.Uint64() == addresses.AlfajoresChainID {
		tokenAddress = addresses.CeloTokenAlfajoresAddress
	}

	leftoverGas, err := evm.CallWithABI(
		celoTokenABI, "decreaseWithdrawn", tokenAddress, maxGasForIncreaseWithdraw,
		// decreaseWithdrawn(uint256 value) parameters
		amount,
	)
	gasUsed := maxGasForIncreaseWithdraw - leftoverGas
	//TODO: how to handle gas?
	log.Trace("DecreaseWithdrawn called", "amount", *amount, "gasUsed", gasUsed)
	return err
}
