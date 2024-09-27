package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/rlp"
)

type celoDynamicReceiptRLP struct {
	PostStateOrStatus []byte
	CumulativeGasUsed uint64
	Bloom             Bloom
	Logs              []*Log
	// BaseFee was introduced as mandatory in Cel2 ONLY for the CeloDynamicFeeTxs
	BaseFee *big.Int `rlp:"optional"`
}

type CeloDynamicFeeStoredReceiptRLP struct {
	CeloDynamicReceiptMarker []interface{} // Marker to distinguish this from storedReceiptRLP
	PostStateOrStatus        []byte
	CumulativeGasUsed        uint64
	Logs                     []*Log
	BaseFee                  *big.Int `rlp:"optional"`
}

// Detect CeloDynamicFee receipts by looking at the first list element
// To distinguish these receipts from the very similar normal receipts, an
// empty list is added as the first element of the RLP-serialized struct.
func IsCeloDynamicFeeReceipt(blob []byte) bool {
	listHeaderSize := 1 // Length of the list header representing the struct in bytes
	if blob[0] > 0xf7 {
		listHeaderSize += int(blob[0]) - 0xf7
	}
	firstListElement := blob[listHeaderSize] // First byte of first list element
	return firstListElement == 0xc0
}

func decodeStoredCeloDynamicFeeReceiptRLP(r *ReceiptForStorage, blob []byte) error {
	var stored CeloDynamicFeeStoredReceiptRLP
	if err := rlp.DecodeBytes(blob, &stored); err != nil {
		return err
	}
	if err := (*Receipt)(r).setStatus(stored.PostStateOrStatus); err != nil {
		return err
	}
	r.CumulativeGasUsed = stored.CumulativeGasUsed
	r.Logs = stored.Logs
	r.Bloom = CreateBloom(Receipts{(*Receipt)(r)})
	r.BaseFee = stored.BaseFee
	return nil
}
