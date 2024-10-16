package types

import (
	"bytes"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

const CeloDenominatedTxType = 0x7a

type CeloDenominatedTx struct {
	ChainID    *big.Int
	Nonce      uint64
	GasTipCap  *big.Int
	GasFeeCap  *big.Int
	Gas        uint64
	To         *common.Address `rlp:"nil"` // nil means contract creation
	Value      *big.Int
	Data       []byte
	AccessList AccessList

	FeeCurrency         *common.Address
	MaxFeeInFeeCurrency *big.Int

	// Signature values
	V *big.Int `json:"v" gencodec:"required"`
	R *big.Int `json:"r" gencodec:"required"`
	S *big.Int `json:"s" gencodec:"required"`
}

// copy creates a deep copy of the transaction data and initializes all fields.
func (tx *CeloDenominatedTx) copy() TxData {
	cpy := &CeloDenominatedTx{
		Nonce:       tx.Nonce,
		To:          copyAddressPtr(tx.To),
		Data:        common.CopyBytes(tx.Data),
		Gas:         tx.Gas,
		FeeCurrency: copyAddressPtr(tx.FeeCurrency),
		// These are copied below.
		MaxFeeInFeeCurrency: new(big.Int),
		AccessList:          make(AccessList, len(tx.AccessList)),
		Value:               new(big.Int),
		ChainID:             new(big.Int),
		GasTipCap:           new(big.Int),
		GasFeeCap:           new(big.Int),
		V:                   new(big.Int),
		R:                   new(big.Int),
		S:                   new(big.Int),
	}
	if tx.MaxFeeInFeeCurrency != nil {
		cpy.MaxFeeInFeeCurrency.Set(tx.MaxFeeInFeeCurrency)
	}
	copy(cpy.AccessList, tx.AccessList)
	if tx.Value != nil {
		cpy.Value.Set(tx.Value)
	}
	if tx.ChainID != nil {
		cpy.ChainID.Set(tx.ChainID)
	}
	if tx.GasTipCap != nil {
		cpy.GasTipCap.Set(tx.GasTipCap)
	}
	if tx.GasFeeCap != nil {
		cpy.GasFeeCap.Set(tx.GasFeeCap)
	}
	if tx.V != nil {
		cpy.V.Set(tx.V)
	}
	if tx.R != nil {
		cpy.R.Set(tx.R)
	}
	if tx.S != nil {
		cpy.S.Set(tx.S)
	}
	return cpy
}

// accessors for innerTx.
func (tx *CeloDenominatedTx) txType() byte           { return CeloDenominatedTxType }
func (tx *CeloDenominatedTx) chainID() *big.Int      { return tx.ChainID }
func (tx *CeloDenominatedTx) accessList() AccessList { return tx.AccessList }
func (tx *CeloDenominatedTx) data() []byte           { return tx.Data }
func (tx *CeloDenominatedTx) gas() uint64            { return tx.Gas }
func (tx *CeloDenominatedTx) gasFeeCap() *big.Int    { return tx.GasFeeCap }
func (tx *CeloDenominatedTx) gasTipCap() *big.Int    { return tx.GasTipCap }
func (tx *CeloDenominatedTx) gasPrice() *big.Int     { return tx.GasFeeCap }
func (tx *CeloDenominatedTx) value() *big.Int        { return tx.Value }
func (tx *CeloDenominatedTx) nonce() uint64          { return tx.Nonce }
func (tx *CeloDenominatedTx) to() *common.Address    { return tx.To }
func (tx *CeloDenominatedTx) isSystemTx() bool       { return false }

func (tx *CeloDenominatedTx) effectiveGasPrice(dst *big.Int, baseFee *big.Int) *big.Int {
	if baseFee == nil {
		return dst.Set(tx.GasFeeCap)
	}
	tip := dst.Sub(tx.GasFeeCap, baseFee)
	if tip.Cmp(tx.GasTipCap) > 0 {
		tip.Set(tx.GasTipCap)
	}
	return tip.Add(tip, baseFee)
}

func (tx *CeloDenominatedTx) rawSignatureValues() (v, r, s *big.Int) {
	return tx.V, tx.R, tx.S
}

func (tx *CeloDenominatedTx) setSignatureValues(chainID, v, r, s *big.Int) {
	tx.ChainID, tx.V, tx.R, tx.S = chainID, v, r, s
}

func (tx *CeloDenominatedTx) encode(b *bytes.Buffer) error {
	return rlp.Encode(b, tx)
}

func (tx *CeloDenominatedTx) decode(input []byte) error {
	return rlp.DecodeBytes(input, tx)
}
