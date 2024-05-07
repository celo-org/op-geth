package types

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/exchange"
)

// Returns the fee currency of the transaction if there is one.
func (tx *Transaction) FeeCurrency() *common.Address {
	var feeCurrency *common.Address
	switch t := tx.inner.(type) {
	case *CeloDynamicFeeTxV2:
		feeCurrency = t.FeeCurrency
	}
	return feeCurrency
}

// CompareWithRates compares the effective gas price of two transactions according to the exchange rates and
// the base fees in the transactions currencies.
func CompareWithRates(a, b *Transaction, ratesAndFees *exchange.RatesAndFees) int {
	if ratesAndFees == nil {
		// During node startup the ratesAndFees might not be yet setup, compare nominally
		feeCapCmp := a.GasFeeCapCmp(b)
		if feeCapCmp != 0 {
			return feeCapCmp
		}
		return a.GasTipCapCmp(b)
	}
	rates := ratesAndFees.Rates
	if ratesAndFees.HasBaseFee() {
		tipA := a.EffectiveGasTipValue(ratesAndFees.GetBaseFeeIn(a.FeeCurrency()))
		tipB := b.EffectiveGasTipValue(ratesAndFees.GetBaseFeeIn(b.FeeCurrency()))
		c, _ := exchange.CompareValue(rates, tipA, a.FeeCurrency(), tipB, b.FeeCurrency())
		return c
	}

	// Compare fee caps if baseFee is not specified or effective tips are equal
	feeA := a.inner.gasFeeCap()
	feeB := b.inner.gasFeeCap()
	c, _ := exchange.CompareValue(rates, feeA, a.FeeCurrency(), feeB, b.FeeCurrency())
	if c != 0 {
		return c
	}

	// Compare tips if effective tips and fee caps are equal
	tipCapA := a.inner.gasTipCap()
	tipCapB := b.inner.gasTipCap()
	c, _ = exchange.CompareValue(rates, tipCapA, a.FeeCurrency(), tipCapB, b.FeeCurrency())
	return c
}
