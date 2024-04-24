package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
)

const (
	// CeloDynamicFeeTxType = 0x7c  old Celo tx type with gateway fee
	CeloDynamicFeeTxType = 0x7b
)

// Returns the fee currency of the transaction if there is one.
func (tx *Transaction) FeeCurrency() *common.Address {
	var feeCurrency *common.Address
	switch t := tx.inner.(type) {
	case *CeloDynamicFeeTx:
		feeCurrency = t.FeeCurrency
	}
	return feeCurrency
}

func celoTransactionMarshal(tx *Transaction) ([]byte, bool, error) {
	var enc txJSON
	// These are set for all tx types.
	enc.Hash = tx.Hash()
	enc.Type = hexutil.Uint64(tx.Type())
	switch itx := tx.inner.(type) {
	case *CeloDynamicFeeTx:
		enc.ChainID = (*hexutil.Big)(itx.ChainID)
		enc.Nonce = (*hexutil.Uint64)(&itx.Nonce)
		enc.To = tx.To()
		enc.Gas = (*hexutil.Uint64)(&itx.Gas)
		enc.MaxFeePerGas = (*hexutil.Big)(itx.GasFeeCap)
		enc.MaxPriorityFeePerGas = (*hexutil.Big)(itx.GasTipCap)
		enc.FeeCurrency = itx.FeeCurrency
		enc.Value = (*hexutil.Big)(itx.Value)
		enc.Input = (*hexutil.Bytes)(&itx.Data)
		enc.AccessList = &itx.AccessList
		enc.V = (*hexutil.Big)(itx.V)
		enc.R = (*hexutil.Big)(itx.R)
		enc.S = (*hexutil.Big)(itx.S)
	default:
		return nil, false, nil
	}
	bytes, err := json.Marshal(&enc)
	return bytes, true, err
}

func celoTransactionUnmarshal(dec txJSON, inner *TxData) (bool, error) {
	switch dec.Type {
	case CeloDynamicFeeTxType:
		var itx CeloDynamicFeeTx
		*inner = &itx
		if dec.ChainID == nil {
			return true, errors.New("missing required field 'chainId' in transaction")
		}
		itx.ChainID = (*big.Int)(dec.ChainID)
		if dec.Nonce == nil {
			return true, errors.New("missing required field 'nonce' in transaction")
		}
		itx.Nonce = uint64(*dec.Nonce)
		if dec.To != nil {
			itx.To = dec.To
		}
		if dec.Gas == nil {
			return true, errors.New("missing required field 'gas' for txdata")
		}
		itx.Gas = uint64(*dec.Gas)
		if dec.MaxPriorityFeePerGas == nil {
			return true, errors.New("missing required field 'maxPriorityFeePerGas' for txdata")
		}
		itx.GasTipCap = (*big.Int)(dec.MaxPriorityFeePerGas)
		if dec.MaxFeePerGas == nil {
			return true, errors.New("missing required field 'maxFeePerGas' for txdata")
		}
		itx.GasFeeCap = (*big.Int)(dec.MaxFeePerGas)
		if dec.Value == nil {
			return true, errors.New("missing required field 'value' in transaction")
		}
		itx.FeeCurrency = dec.FeeCurrency
		itx.Value = (*big.Int)(dec.Value)
		if dec.Input == nil {
			return true, errors.New("missing required field 'input' in transaction")
		}
		itx.Data = *dec.Input
		if dec.V == nil {
			return true, errors.New("missing required field 'v' in transaction")
		}
		if dec.AccessList != nil {
			itx.AccessList = *dec.AccessList
		}
		itx.V = (*big.Int)(dec.V)
		if dec.R == nil {
			return true, errors.New("missing required field 'r' in transaction")
		}
		itx.R = (*big.Int)(dec.R)
		if dec.S == nil {
			return true, errors.New("missing required field 's' in transaction")
		}
		itx.S = (*big.Int)(dec.S)
		withSignature := itx.V.Sign() != 0 || itx.R.Sign() != 0 || itx.S.Sign() != 0
		if withSignature {
			if err := sanityCheckSignature(itx.V, itx.R, itx.S, false); err != nil {
				return true, err
			}
		}
	default:
		return false, nil
	}

	return true, nil
}

func celoDecodeTyped(b []byte) (TxData, bool, error) {
	var inner TxData
	switch b[0] {
	case CeloDynamicFeeTxType:
		inner = new(CeloDynamicFeeTx)
	default:
		return nil, false, nil
	}
	err := inner.decode(b[1:])
	return inner, true, err
}

type cel2Signer struct{ londonSigner }

// NewCel2Signer returns a signer that accepts
// - CIP-64 celo dynamic fee transaction (v2)
// - EIP-4844 blob transactions
// - EIP-1559 dynamic fee transactions
// - EIP-2930 access list transactions,
// - EIP-155 replay protected transactions, and
// - legacy Homestead transactions.
func NewCel2Signer(chainId *big.Int) Signer {
	return cel2Signer{londonSigner{eip2930Signer{NewEIP155Signer(chainId)}}}
}

func (s cel2Signer) Sender(tx *Transaction) (common.Address, error) {
	if tx.Type() != CeloDynamicFeeTxType {
		return s.londonSigner.Sender(tx)
	}
	V, R, S := tx.RawSignatureValues()
	// DynamicFee txs are defined to use 0 and 1 as their recovery
	// id, add 27 to become equivalent to unprotected Homestead signatures.
	V = new(big.Int).Add(V, big.NewInt(27))
	if tx.ChainId().Cmp(s.chainId) != 0 {
		return common.Address{}, ErrInvalidChainId
	}
	return recoverPlain(s.Hash(tx), R, S, V, true)
}

func (s cel2Signer) Equal(s2 Signer) bool {
	x, ok := s2.(cel2Signer)
	return ok && x.chainId.Cmp(s.chainId) == 0
}

func (s cel2Signer) SignatureValues(tx *Transaction, sig []byte) (R, S, V *big.Int, err error) {
	txdata, ok := tx.inner.(*CeloDynamicFeeTx)
	if !ok {
		return s.londonSigner.SignatureValues(tx, sig)
	}
	// Check that chain ID of tx matches the signer. We also accept ID zero here,
	// because it indicates that the chain ID was not specified in the tx.
	if txdata.ChainID.Sign() != 0 && txdata.ChainID.Cmp(s.chainId) != 0 {
		return nil, nil, nil, ErrInvalidChainId
	}
	R, S, _ = decodeSignature(sig)
	V = big.NewInt(int64(sig[64]))
	return R, S, V, nil
}

// Hash returns the hash to be signed by the sender.
// It does not uniquely identify the transaction.
func (s cel2Signer) Hash(tx *Transaction) common.Hash {
	if tx.Type() == CeloDynamicFeeTxType {
		return prefixedRlpHash(
			tx.Type(),
			[]interface{}{
				s.chainId,
				tx.Nonce(),
				tx.GasTipCap(),
				tx.GasFeeCap(),
				tx.Gas(),
				tx.To(),
				tx.Value(),
				tx.Data(),
				tx.AccessList(),
				tx.FeeCurrency(),
			})
	}
	return s.londonSigner.Hash(tx)
}

func makeCeloSigner(config *params.ChainConfig, blockNumber *big.Int, blockTime uint64) (Signer, bool) {
	if config.IsCel2(blockTime) {
		return NewCel2Signer(config.ChainID), true
	}
	return nil, false
}

func latestCeloSigner(config *params.ChainConfig) (Signer, bool) {
	if config.ChainID != nil && config.Cel2Time != nil {
		return NewCel2Signer(config.ChainID), true
	}
	return nil, false
}

// CeloDynamicFeeTx represents a CIP-64 transaction.
type CeloDynamicFeeTx struct {
	ChainID    *big.Int
	Nonce      uint64
	GasTipCap  *big.Int
	GasFeeCap  *big.Int
	Gas        uint64
	To         *common.Address `rlp:"nil"` // nil means contract creation
	Value      *big.Int
	Data       []byte
	AccessList AccessList

	FeeCurrency *common.Address `rlp:"nil"` // nil means native currency

	// Signature values
	V *big.Int `json:"v" gencodec:"required"`
	R *big.Int `json:"r" gencodec:"required"`
	S *big.Int `json:"s" gencodec:"required"`
}

// copy creates a deep copy of the transaction data and initializes all fields.
func (tx *CeloDynamicFeeTx) copy() TxData {
	cpy := &CeloDynamicFeeTx{
		Nonce:       tx.Nonce,
		To:          copyAddressPtr(tx.To),
		Data:        common.CopyBytes(tx.Data),
		Gas:         tx.Gas,
		FeeCurrency: copyAddressPtr(tx.FeeCurrency),
		// These are copied below.
		AccessList: make(AccessList, len(tx.AccessList)),
		Value:      new(big.Int),
		ChainID:    new(big.Int),
		GasTipCap:  new(big.Int),
		GasFeeCap:  new(big.Int),
		V:          new(big.Int),
		R:          new(big.Int),
		S:          new(big.Int),
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
func (tx *CeloDynamicFeeTx) txType() byte           { return CeloDynamicFeeTxType }
func (tx *CeloDynamicFeeTx) chainID() *big.Int      { return tx.ChainID }
func (tx *CeloDynamicFeeTx) accessList() AccessList { return tx.AccessList }
func (tx *CeloDynamicFeeTx) data() []byte           { return tx.Data }
func (tx *CeloDynamicFeeTx) gas() uint64            { return tx.Gas }
func (tx *CeloDynamicFeeTx) gasFeeCap() *big.Int    { return tx.GasFeeCap }
func (tx *CeloDynamicFeeTx) gasTipCap() *big.Int    { return tx.GasTipCap }
func (tx *CeloDynamicFeeTx) gasPrice() *big.Int     { return tx.GasFeeCap }
func (tx *CeloDynamicFeeTx) value() *big.Int        { return tx.Value }
func (tx *CeloDynamicFeeTx) nonce() uint64          { return tx.Nonce }
func (tx *CeloDynamicFeeTx) to() *common.Address    { return tx.To }
func (tx *CeloDynamicFeeTx) isSystemTx() bool       { return false }

func (tx *CeloDynamicFeeTx) effectiveGasPrice(dst *big.Int, baseFee *big.Int) *big.Int {
	if baseFee == nil {
		return dst.Set(tx.GasFeeCap)
	}
	tip := dst.Sub(tx.GasFeeCap, baseFee)
	if tip.Cmp(tx.GasTipCap) > 0 {
		tip.Set(tx.GasTipCap)
	}
	return tip.Add(tip, baseFee)
}

func (tx *CeloDynamicFeeTx) rawSignatureValues() (v, r, s *big.Int) {
	return tx.V, tx.R, tx.S
}

func (tx *CeloDynamicFeeTx) setSignatureValues(chainID, v, r, s *big.Int) {
	tx.ChainID, tx.V, tx.R, tx.S = chainID, v, r, s
}

func (tx *CeloDynamicFeeTx) encode(b *bytes.Buffer) error {
	return rlp.Encode(b, tx)
}

func (tx *CeloDynamicFeeTx) decode(input []byte) error {
	return rlp.DecodeBytes(input, tx)
}
