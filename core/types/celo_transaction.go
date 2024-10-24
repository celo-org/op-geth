// Copyright 2024 The Celo Authors
// This file is part of the celo library.
//
// The celo library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The celo library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the celo library. If not, see <http://www.gnu.org/licenses/>.

package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/exchange"
)

// IsCeloLegacy returns true if the transaction is a legacy celo transaction.
// I.E. it has the fields feeCurrency, gatewayFee and gatewayFeeRecipient.
func (tx *Transaction) IsCeloLegacy() bool {
	switch t := tx.inner.(type) {
	case *LegacyTx:
		return t.CeloLegacy
	}
	return false
}

// FeeCurrency returns the fee currency of the transaction if there is one.
func (tx *Transaction) FeeCurrency() *common.Address {
	var feeCurrency *common.Address
	switch t := tx.inner.(type) {
	case *CeloDynamicFeeTx:
		feeCurrency = t.FeeCurrency
	case *CeloDynamicFeeTxV2:
		feeCurrency = t.FeeCurrency
	case *CeloDenominatedTx:
		feeCurrency = t.FeeCurrency
	case *LegacyTx:
		feeCurrency = t.FeeCurrency
	}
	return feeCurrency
}

// GatewayFee returns the gateway fee of the transaction if there is one.
// Note: this is here to support serving legacy transactions over the RPC, it should not be used in new code.
func (tx *Transaction) GatewayFee() *big.Int {
	var gatewayFee *big.Int
	switch t := tx.inner.(type) {
	case *CeloDynamicFeeTx:
		gatewayFee = t.GatewayFee
	case *LegacyTx:
		gatewayFee = t.GatewayFee
	}
	return gatewayFee
}

// GatewayFeeRecipient returns the gateway fee recipient of the transaction if there is one.
// Note: this is here to support serving legacy transactions over the RPC, it should not be used in new code.
func (tx *Transaction) GatewayFeeRecipient() *common.Address {
	var gatewayFeeRecipient *common.Address
	switch t := tx.inner.(type) {
	case *CeloDynamicFeeTx:
		gatewayFeeRecipient = t.GatewayFeeRecipient
	case *LegacyTx:
		gatewayFeeRecipient = t.GatewayFeeRecipient
	}
	return gatewayFeeRecipient
}

// MaxFeeInFeeCurrency returns the maximum fee in the fee currency of the transaction if there is one.
func (tx *Transaction) MaxFeeInFeeCurrency() *big.Int {
	var maxFeeInFeeCurrency *big.Int
	switch t := tx.inner.(type) {
	case *CeloDenominatedTx:
		maxFeeInFeeCurrency = t.MaxFeeInFeeCurrency
	}
	return maxFeeInFeeCurrency
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

func copyBigInt(b *big.Int) *big.Int {
	if b == nil {
		return nil
	}
	return new(big.Int).Set(b)
}

// SetYNullStyleBigIfZero is for use by tests only, it sets the Y (AKA V) field of the transaction to big.NewInt(0)
// which internally ensures that the abs field of the big int is null as opposed to an empty slice. The reason for doing
// this is to facilitate direct deep equal comparisons of transactions, which although they may share the same value for
// V have different internal representations.
func SetYNullStyleBigIfZero(tx *Transaction) {
	switch itx := tx.inner.(type) {
	case *DynamicFeeTx:
		if itx.V.Sign() == 0 {
			itx.V = big.NewInt(0)
		}
	case *AccessListTx:
		if itx.V.Sign() == 0 {
			itx.V = big.NewInt(0)
		}
	}
}
