package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

const (
	// CeloDynamicFeeTxV2Type = 0x7c  old Celo tx type with gateway fee
	CeloDynamicFeeTxV2Type = 0x7b
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

func (tx *Transaction) MaxFeeInFeeCurrency() *big.Int {
	var maxFeeInFeeCurrency *big.Int
	switch t := tx.inner.(type) {
	case *CeloDenominatedTx:
		maxFeeInFeeCurrency = t.MaxFeeInFeeCurrency
	}
	return maxFeeInFeeCurrency
}
