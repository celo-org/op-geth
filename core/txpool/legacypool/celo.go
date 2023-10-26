package legacypool

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/celo/abigen"
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
