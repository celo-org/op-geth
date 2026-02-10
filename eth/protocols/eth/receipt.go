// Copyright 2024 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package eth

import (
	"bytes"
	"fmt"
	"io"
	"iter"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

// Receipt is the representation of receipts for networking purposes.
type Receipt struct {
	TxType            byte
	PostStateOrStatus []byte
	GasUsed           uint64
	Logs              rlp.RawValue

	// OP-Stack additions for deposit receipts.
	// The `optional` tags are vestigial — auto-RLP is bypassed by the custom
	// EncodeRLP/DecodeRLP methods below, but we keep the tags for any callers
	// that still introspect the struct via reflection.
	DepositNonce          *uint64 `rlp:"optional"`
	DepositReceiptVersion *uint64 `rlp:"optional"`

	// Celo addition for CeloDynamicFeeTxV2 receipts.
	BaseFee *big.Int `rlp:"optional"`
}

func newReceipt(tr *types.Receipt) Receipt {
	r := Receipt{TxType: tr.Type, GasUsed: tr.CumulativeGasUsed}
	if tr.PostState != nil {
		r.PostStateOrStatus = tr.PostState
	} else {
		r.PostStateOrStatus = new(big.Int).SetUint64(tr.Status).Bytes()
	}
	r.Logs, _ = rlp.EncodeToBytes(tr.Logs)

	// OP-Stack addition for deposit receipts - these fields will be nil for non-deposit receipts
	r.DepositNonce = tr.DepositNonce
	r.DepositReceiptVersion = tr.DepositReceiptVersion

	// Celo addition for CeloDynamicFeeTxV2 receipts
	r.BaseFee = tr.BaseFee

	return r
}

// Celo addition for CeloDynamicFeeTxV2 receipts.
func (r *Receipt) maybeWriteCeloBaseFee(w *rlp.EncoderBuffer) {
	if r.BaseFee != nil {
		w.WriteBigInt(r.BaseFee)
	}
}

// OP-Stack addition for deposit receipts.
func (r *Receipt) maybeWriteDepositFields(w *rlp.EncoderBuffer, onlyWithVersion bool) {
	// Post-Regolith+pre-Canyon receipts may have been stored in DBs with
	// the deposit nonce but not the version.
	// And post-Regolith+pre-Canyon receipt hashes didn't include the deposit nonce, so we
	// need the onlyWithVersion variant for [encodeForHash] to detect this case.
	if onlyWithVersion && r.DepositReceiptVersion == nil {
		return
	}

	if r.DepositNonce != nil {
		w.WriteUint64(*r.DepositNonce)
		if r.DepositReceiptVersion != nil {
			w.WriteUint64(*r.DepositReceiptVersion)
		}
	}
}

// EncodeRLP implements rlp.Encoder. The encoding mirrors the original Celo
// `encodeForNetwork69`: trailing fields after Logs are gated by which optional
// fields are non-nil. For V2 receipts BaseFee is the only trailing field; for
// deposit receipts it's DepositNonce/Version. Auto-RLP can't express this
// because it would write nil-pointer placeholders ahead of any non-nil
// optional, which the decoder then can't disambiguate.
func (r *Receipt) EncodeRLP(w io.Writer) error {
	enc := rlp.NewEncoderBuffer(w)
	list := enc.List()
	enc.WriteUint64(uint64(r.TxType))
	enc.WriteBytes(r.PostStateOrStatus)
	enc.WriteUint64(r.GasUsed)
	enc.Write(r.Logs)
	r.maybeWriteDepositFields(&enc, false)
	r.maybeWriteCeloBaseFee(&enc)
	enc.ListEnd(list)
	return enc.Flush()
}

// DecodeRLP implements rlp.Decoder. It mirrors the original Celo
// `decode69`+`decodeInnerList`: trailing fields after Logs are gated by TxType.
func (r *Receipt) DecodeRLP(s *rlp.Stream) error {
	if _, err := s.List(); err != nil {
		return err
	}
	txType, err := s.Uint8()
	if err != nil {
		return fmt.Errorf("invalid txType: %w", err)
	}
	r.TxType = txType

	r.PostStateOrStatus, err = s.Bytes()
	if err != nil {
		return fmt.Errorf("invalid postStateOrStatus: %w", err)
	}
	r.GasUsed, err = s.Uint64()
	if err != nil {
		return fmt.Errorf("invalid gasUsed: %w", err)
	}
	r.Logs, err = s.Raw()
	if err != nil {
		return fmt.Errorf("invalid logs: %w", err)
	}

	// OP-Stack: optional deposit receipt fields (only for deposit receipts).
	if r.TxType == types.DepositTxType && s.MoreDataInList() {
		dn, err := s.Uint64()
		if err != nil {
			return fmt.Errorf("invalid depositNonce: %w", err)
		}
		r.DepositNonce = &dn
		if s.MoreDataInList() {
			drv, err := s.Uint64()
			if err != nil {
				return fmt.Errorf("invalid depositReceiptVersion: %w", err)
			}
			r.DepositReceiptVersion = &drv
		}
	}
	// Celo: optional base fee (only for CeloDynamicFeeTxV2 receipts).
	if r.TxType == types.CeloDynamicFeeTxV2Type && s.MoreDataInList() {
		var bf big.Int
		if err := s.Decode(&bf); err != nil {
			return fmt.Errorf("invalid baseFee: %w", err)
		}
		r.BaseFee = &bf
	}

	return s.ListEnd()
}

// encodeForHash encodes a receipt for the block receiptsRoot derivation.
func (r *Receipt) encodeForHash(bloomBuf *[6]byte, out *bytes.Buffer) {
	// For typed receipts, add the tx type.
	if r.TxType != 0 {
		out.WriteByte(r.TxType)
	}
	// Encode list = [postStateOrStatus, gasUsed, bloom, logs].
	w := rlp.NewEncoderBuffer(out)
	l := w.List()
	w.WriteBytes(r.PostStateOrStatus)
	w.WriteUint64(r.GasUsed)
	bloom := r.bloom(bloomBuf)
	w.WriteBytes(bloom[:])
	w.Write(r.Logs)
	// Celo addition: include base fee for CeloDynamicFeeTxV2 receipts.
	r.maybeWriteCeloBaseFee(&w)
	// Note that deposit fields must NOT be included in the receipt hash pre-Canyon,
	// which is detected by checking for the presence of the version field.
	r.maybeWriteDepositFields(&w, true)
	w.ListEnd(l)
	w.Flush()
}

// bloom computes the bloom filter of the receipt.
// Note this doesn't check the validity of encoding, and will produce an invalid filter
// for invalid input. This is acceptable for the purpose of this function, which is
// recomputing the receipt hash.
func (r *Receipt) bloom(buffer *[6]byte) types.Bloom {
	var b types.Bloom
	logsIter, err := rlp.NewListIterator(r.Logs)
	if err != nil {
		return b
	}
	for logsIter.Next() {
		log, _, _ := rlp.SplitList(logsIter.Value())
		address, log, _ := rlp.SplitString(log)
		b.AddWithBuffer(address, buffer)
		topicsIter, err := rlp.NewListIterator(log)
		if err != nil {
			return b
		}
		for topicsIter.Next() {
			topic, _, _ := rlp.SplitString(topicsIter.Value())
			b.AddWithBuffer(topic, buffer)
		}
	}
	return b
}

// decode assigns the fields of r by decoding the network format.
// It delegates to DecodeRLP, which knows how to read trailing fields gated
// by TxType (deposit nonce/version, or Celo BaseFee for V2 receipts).
func (r *Receipt) decode(input []byte) error {
	return rlp.DecodeBytes(input, r)
}

// ReceiptList is the block receipt list as downloaded by eth/69.
type ReceiptList struct {
	items rlp.RawList[Receipt]
}

// NewReceiptList creates a receipt list.
// This is slow, and exists for testing purposes.
func NewReceiptList(trs []*types.Receipt) *ReceiptList {
	rl := new(ReceiptList)
	for _, tr := range trs {
		r := newReceipt(tr)
		encoded, _ := rlp.EncodeToBytes(&r)
		rl.items.AppendRaw(encoded)
	}
	return rl
}

// DecodeRLP decodes a list receipts from the network format.
func (rl *ReceiptList) DecodeRLP(s *rlp.Stream) error {
	return rl.items.DecodeRLP(s)
}

// EncodeRLP encodes the list into the network format of eth/69.
func (rl *ReceiptList) EncodeRLP(w io.Writer) error {
	return rl.items.EncodeRLP(w)
}

// EncodeForStorage encodes a list of receipts for the database.
// It strips the first element (TxType) from each receipt's raw RLP and,
// for CeloDynamicFeeTxV2 receipts, prepends the empty-list marker (0xc0)
// expected by the storage encoding.
func (rl *ReceiptList) EncodeForStorage() (rlp.RawValue, error) {
	var out bytes.Buffer
	w := rlp.NewEncoderBuffer(&out)
	outer := w.List()
	it := rl.items.ContentIterator()
	for it.Next() {
		content, _, err := rlp.SplitList(it.Value())
		if err != nil {
			return nil, fmt.Errorf("bad receipt: %v", err)
		}
		txType, rest, err := rlp.SplitUint64(content)
		if err != nil {
			return nil, fmt.Errorf("bad receipt: %v", err)
		}
		inner := w.List()
		// Celo addition: re-add the empty-list marker for CeloDynamicFeeTxV2 receipts.
		if byte(txType) == types.CeloDynamicFeeTxV2Type {
			marker := w.List()
			w.ListEnd(marker)
		}
		w.Write(rest)
		w.ListEnd(inner)
	}
	if it.Err() != nil {
		return nil, fmt.Errorf("bad list: %v", it.Err())
	}
	w.ListEnd(outer)
	w.Flush()
	return out.Bytes(), nil
}

// Derivable returns a DerivableList, which can be used to decode
func (rl *ReceiptList) Derivable() types.DerivableList {
	var bloomBuf [6]byte
	return newDerivableRawList(&rl.items, func(data []byte, outbuf *bytes.Buffer) {
		var r Receipt
		if r.decode(data) == nil {
			r.encodeForHash(&bloomBuf, outbuf)
		}
	})
}

// blockReceiptsToNetwork takes a slice of rlp-encoded receipts, and transactions,
// and re-encodes them for the network protocol.
func blockReceiptsToNetwork(blockReceipts, blockBody rlp.RawValue) ([]byte, error) {
	txTypesIter, err := txTypesInBody(blockBody)
	if err != nil {
		return nil, fmt.Errorf("invalid block body: %v", err)
	}
	nextTxType, stopTxTypes := iter.Pull(txTypesIter)
	defer stopTxTypes()

	var (
		out   bytes.Buffer
		enc   = rlp.NewEncoderBuffer(&out)
		it, _ = rlp.NewListIterator(blockReceipts)
	)
	outer := enc.List()
	for i := 0; it.Next(); i++ {
		txType, _ := nextTxType()
		content, _, _ := rlp.SplitList(it.Value())

		// Celo addition: strip the empty list marker (0xc0) for CeloDynamicFeeTxV2 storage format.
		if txType == types.CeloDynamicFeeTxV2Type && len(content) > 0 && content[0] == 0xc0 {
			content = content[1:]
		}

		receiptList := enc.List()
		enc.WriteUint64(uint64(txType))
		enc.Write(content)
		enc.ListEnd(receiptList)
	}
	enc.ListEnd(outer)
	enc.Flush()
	return out.Bytes(), nil
}

// txTypesInBody parses the transactions list of an encoded block body, returning just the types.
func txTypesInBody(body rlp.RawValue) (iter.Seq[byte], error) {
	bodyFields, _, err := rlp.SplitList(body)
	if err != nil {
		return nil, err
	}
	txsIter, err := rlp.NewListIterator(bodyFields)
	if err != nil {
		return nil, err
	}
	return func(yield func(byte) bool) {
		for txsIter.Next() {
			var txType byte
			switch k, content, _, _ := rlp.Split(txsIter.Value()); k {
			case rlp.List:
				txType = 0
			case rlp.String:
				if len(content) > 0 {
					txType = content[0]
				}
			}
			if !yield(txType) {
				return
			}
		}
	}, nil
}
