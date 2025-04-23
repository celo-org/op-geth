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

func copyBigInt(b *big.Int) *big.Int {
	if b == nil {
		return nil
	}
	return new(big.Int).Set(b)
}
