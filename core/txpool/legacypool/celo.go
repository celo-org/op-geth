package legacypool

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/celo/abigen"
	"github.com/ethereum/go-ethereum/core/state"
)

var (
	unitRate = big.NewRat(1, 1)
)

// IsWhitelisted checks if a given fee currency is whitelisted
func IsWhitelisted(exchangeRates common.ExchangeRates, feeCurrency *common.Address) bool {
	if feeCurrency == nil {
		return true
	}
	_, ok := exchangeRates[*feeCurrency]
	return ok
}

func TranslateValue(exchangeRates common.ExchangeRates, val *big.Int, fromFeeCurrency, toFeeCurrency *common.Address) (*big.Int, error) {
	// TODO: implement me
	return val, nil
}

func CurrencyBaseFee(exchangeRates common.ExchangeRates, feeCurrency *common.Address) *big.Int {
	// TODO: implement me
	return nil
}

func CurrencyBaseFeeAt(st *state.StateDB, feeCurrency *common.Address) *big.Int {
	var exchangeRates common.ExchangeRates
	return CurrencyBaseFee(exchangeRates, feeCurrency)
}

// Compares values in different currencies
// nil currency => native currency
func CompareValue(exchangeRates common.ExchangeRates, val1 *big.Int, feeCurrency1 *common.Address, val2 *big.Int, feeCurrency2 *common.Address) (int, error) {
	// Short circuit if the fee currency is the same.
	if areEqualAddresses(feeCurrency1, feeCurrency2) {
		return val1.Cmp(val2), nil
	}

	var exchangeRate1, exchangeRate2 *big.Rat
	var ok bool
	if feeCurrency1 == nil {
		exchangeRate1 = unitRate
	} else {
		exchangeRate1, ok = exchangeRates[*feeCurrency1]
		if !ok {
			return 0, fmt.Errorf("fee currency not registered: %s", feeCurrency1.Hex())
		}
	}

	if feeCurrency2 == nil {
		exchangeRate2 = unitRate
	} else {
		exchangeRate2, ok = exchangeRates[*feeCurrency2]
		if !ok {
			return 0, fmt.Errorf("fee currency not registered: %s", feeCurrency1.Hex())
		}
	}

	// Below code block is basically evaluating this comparison:
	// val1 * exchangeRate1.denominator / exchangeRate1.numerator < val2 * exchangeRate2.denominator / exchangeRate2.numerator
	// It will transform that comparison to this, to remove having to deal with fractional values.
	// val1 * exchangeRate1.denominator * exchangeRate2.numerator < val2 * exchangeRate2.denominator * exchangeRate1.numerator
	leftSide := new(big.Int).Mul(
		val1,
		new(big.Int).Mul(
			exchangeRate1.Denom(),
			exchangeRate2.Num(),
		),
	)
	rightSide := new(big.Int).Mul(
		val2,
		new(big.Int).Mul(
			exchangeRate2.Denom(),
			exchangeRate1.Num(),
		),
	)

	return leftSide.Cmp(rightSide), nil
}

func CompareValueAt(st *state.StateDB, val1 *big.Int, curr1 *common.Address, val2 *big.Int, curr2 *common.Address) int {
	// TODO: Get exchangeRates from statedb
	var exchangeRates common.ExchangeRates
	ret, err := CompareValue(exchangeRates, val1, curr1, val2, curr2)
	// Err should not be possible if the pruning of non whitelisted currencies
	// was made properly (and exchange rates are available)
	if err != nil {
		// TODO: LOG
		// Compare with no currencies (Panic could be an option too)
		r2, _ := CompareValue(exchangeRates, val1, nil, val2, nil)
		return r2
	}
	return ret
}

func areEqualAddresses(addr1, addr2 *common.Address) bool {
	return (addr1 == nil && addr2 == nil) || (addr1 != nil && addr2 != nil && *addr1 == *addr2)
}

func GetBalanceOf(backend *bind.ContractCaller, account common.Address, feeCurrency common.Address) (*big.Int, error) {
	token, err := abigen.NewFeeCurrencyCaller(feeCurrency, *backend)
	if err != nil {
		return nil, errors.New("failed to access fee currency token")
	}

	balance, err := token.BalanceOf(&bind.CallOpts{}, account)
	if err != nil {
		return nil, errors.New("failed to access token balance")
	}

	return balance, nil
}
