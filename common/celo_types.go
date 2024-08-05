package common

import (
	"math/big"
)

var (
	ZeroAddress = BytesToAddress([]byte{})
)

type ExchangeRates = map[Address]*big.Rat
type IntrinsicGasCosts = map[Address]uint64

type FeeCurrencyContext struct {
	ExchangeRates     ExchangeRates
	IntrinsicGasCosts IntrinsicGasCosts
}

func MaxAllowedIntrinsicGasCost(i IntrinsicGasCosts, feeCurrency *Address) (uint64, bool) {
	intrinsicGas, ok := CurrencyIntrinsicGasCost(i, feeCurrency)
	if !ok {
		return 0, false
	}
	// allow the contract to overshoot 2 times the deducted intrinsic gas
	// during execution
	return intrinsicGas * 3, true
}

func CurrencyIntrinsicGasCost(i IntrinsicGasCosts, feeCurrency *Address) (uint64, bool) {
	if feeCurrency == nil {
		return 0, true
	}
	gasCost, ok := i[*feeCurrency]
	if !ok {
		return 0, false
	}
	return gasCost, true
}

func CurrencyWhitelist(exchangeRates ExchangeRates) []Address {
	addrs := make([]Address, len(exchangeRates))
	i := 0
	for k := range exchangeRates {
		addrs[i] = k
		i++
	}
	return addrs
}

func IsCurrencyWhitelisted(exchangeRates ExchangeRates, feeCurrency *Address) bool {
	if feeCurrency == nil {
		return true
	}

	// Check if fee currency is registered
	_, ok := exchangeRates[*feeCurrency]
	return ok
}

func AreSameAddress(a, b *Address) bool {
	// both are nil or point to the same address
	if a == b {
		return true
	}
	// if only one is nil
	if a == nil || b == nil {
		return false
	}
	// if they point to the same
	return *a == *b
}
