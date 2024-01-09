package contracts

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/celo/abigen"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/log"
)

const (
	Thousand = 1000
	Million  = 1000 * 1000

	maxGasForDebitGasFeesTransactions  uint64 = 1 * Million
	maxGasForCreditGasFeesTransactions uint64 = 1 * Million
	// Default intrinsic gas cost of transactions paying for gas in alternative currencies.
	// Calculated to estimate 1 balance read, 1 debit, and 4 credit transactions.
	IntrinsicGasForAlternativeFeeCurrency uint64 = 50 * Thousand
)

var (
	tmpAddress = common.HexToAddress("0xce106a5")

	// ErrNonWhitelistedFeeCurrency is returned if the currency specified to use for the fees
	// isn't one of the currencies whitelisted for that purpose.
	ErrNonWhitelistedFeeCurrency = errors.New("non-whitelisted fee currency address")
)

// GetBalanceOf returns an account's balance on a given ERC20 currency
func GetBalanceOf(caller bind.ContractCaller, accountOwner common.Address, contractAddress common.Address) (result *big.Int, err error) {
	token, err := abigen.NewFeeCurrencyCaller(contractAddress, caller)
	if err != nil {
		return nil, fmt.Errorf("failed to access FeeCurrency: %w", err)
	}

	balance, err := token.BalanceOf(&bind.CallOpts{}, accountOwner)
	if err != nil {
		return nil, err
	}

	return balance, nil
}

func ConvertGoldToCurrency(exchangeRates map[common.Address]*big.Rat, feeCurrency *common.Address, goldAmount *big.Int) (*big.Int, error) {
	exchangeRate, ok := exchangeRates[*feeCurrency]
	if !ok {
		return nil, ErrNonWhitelistedFeeCurrency
	}
	return new(big.Int).Div(new(big.Int).Mul(goldAmount, exchangeRate.Num()), exchangeRate.Denom()), nil
}

func DebitFees(evm *vm.EVM, address common.Address, amount *big.Int, feeCurrency *common.Address) error {
	if amount.Cmp(big.NewInt(0)) == 0 {
		return nil
	}
	abi, err := abigen.FeeCurrencyMetaData.GetAbi()
	if err != nil {
		return err
	}
	// Solidity: function transfer(address to, uint256 amount) returns(bool)
	input, err := abi.Pack("transfer", tmpAddress, amount)
	if err != nil {
		return err
	}

	caller := vm.AccountRef(address)

	_, leftoverGas, err := evm.Call(caller, *feeCurrency, input, maxGasForDebitGasFeesTransactions, big.NewInt(0))
	gasUsed := maxGasForDebitGasFeesTransactions - leftoverGas
	log.Trace("DebitFees called", "feeCurrency", *feeCurrency, "gasUsed", gasUsed)
	return err
}

func CreditFees(
	evm *vm.EVM,
	from common.Address,
	feeRecipient common.Address,
	feeHandler common.Address,
	l1DataFeeReceiver common.Address,
	refund *big.Int,
	tipTxFee *big.Int,
	baseTxFee *big.Int,
	l1DataFee *big.Int,
	feeCurrency *common.Address) error {
	caller := vm.AccountRef(tmpAddress)
	leftoverGas := maxGasForCreditGasFeesTransactions

	abi, err := abigen.FeeCurrencyMetaData.GetAbi()
	if err != nil {
		return err
	}

	// Solidity: function transfer(address to, uint256 amount) returns(bool)
	transfer1Data, err := abi.Pack("transfer", feeHandler, baseTxFee)
	if err != nil {
		return fmt.Errorf("pack transfer base fee: %w", err)
	}
	_, leftoverGas, err = evm.Call(caller, *feeCurrency, transfer1Data, leftoverGas, big.NewInt(0))
	if err != nil {
		return fmt.Errorf("call transfer base fee: %w", err)
	}

	if tipTxFee.Cmp(common.Big0) == 1 {
		transfer2Data, err := abi.Pack("transfer", feeRecipient, tipTxFee)
		if err != nil {
			return fmt.Errorf("pack transfer tip: %w", err)
		}
		_, leftoverGas, err = evm.Call(caller, *feeCurrency, transfer2Data, leftoverGas, big.NewInt(0))
		if err != nil {
			return fmt.Errorf("call transfer tip: %w", err)
		}
	}

	if refund.Cmp(common.Big0) == 1 {
		transfer3Data, err := abi.Pack("transfer", from, refund)
		if err != nil {
			return fmt.Errorf("pack transfer refund: %w", err)
		}
		_, leftoverGas, err = evm.Call(caller, *feeCurrency, transfer3Data, leftoverGas, big.NewInt(0))
		if err != nil {
			return fmt.Errorf("call transfer refund: %w", err)
		}
	}

	if l1DataFee != nil {
		transfer4Data, err := abi.Pack("transfer", l1DataFeeReceiver, l1DataFee)
		if err != nil {
			return err
		}
		_, leftoverGas, err = evm.Call(caller, *feeCurrency, transfer4Data, leftoverGas, big.NewInt(0))
		if err != nil {
			return err
		}
	}

	gasUsed := maxGasForCreditGasFeesTransactions - leftoverGas
	log.Trace("creditFees called", "feeCurrency", *feeCurrency, "gasUsed", gasUsed)
	return nil
}
