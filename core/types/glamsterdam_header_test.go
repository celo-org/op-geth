package types

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
)

func TestAmsterdamHeaderRPCAndHash(t *testing.T) {
	zero := uint64(0)
	beaconRoot := common.HexToHash("0x01")
	balHash := common.HexToHash("0x02")
	slotNumber := uint64(123)
	h := &Header{
		UncleHash:        EmptyUncleHash,
		Difficulty:       big.NewInt(0),
		Number:           big.NewInt(456),
		BaseFee:          big.NewInt(100),
		WithdrawalsHash:  &EmptyWithdrawalsHash,
		BlobGasUsed:      &zero,
		ExcessBlobGas:    &zero,
		ParentBeaconRoot: &beaconRoot,
		RequestsHash:     &EmptyRequestsHash,
	}
	data, err := json.Marshal(h)
	if err != nil {
		t.Fatal(err)
	}
	var rpc map[string]any
	if err := json.Unmarshal(data, &rpc); err != nil {
		t.Fatal(err)
	}
	rpc["blockAccessListHash"] = balHash.Hex()
	rpc["slotNumber"] = "0x7b"
	data, err = json.Marshal(rpc)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Header
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	enc, err := rlp.EncodeToBytes([]any{
		h.ParentHash, h.UncleHash, h.Coinbase, h.Root, h.TxHash, h.ReceiptHash,
		h.Bloom, h.Difficulty, h.Number, h.GasLimit, h.GasUsed, h.Time,
		h.Extra, h.MixDigest, h.Nonce, h.BaseFee, *h.WithdrawalsHash,
		zero, zero, beaconRoot, *h.RequestsHash, balHash, slotNumber,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := crypto.Keccak256Hash(enc)
	if got := decoded.Hash(); got != want {
		t.Fatalf("Amsterdam header hash = %s, want %s", got, want)
	}
	if decoded.BlockAccessListHash == nil || *decoded.BlockAccessListHash != balHash {
		t.Fatalf("decoded blockAccessListHash = %v", decoded.BlockAccessListHash)
	}
	if decoded.SlotNumber == nil || *decoded.SlotNumber != slotNumber {
		t.Fatalf("decoded slotNumber = %v", decoded.SlotNumber)
	}

	encoded, err := json.Marshal(&decoded)
	if err != nil {
		t.Fatal(err)
	}
	var marshaled map[string]any
	if err := json.Unmarshal(encoded, &marshaled); err != nil {
		t.Fatal(err)
	}
	if marshaled["blockAccessListHash"] != balHash.Hex() || marshaled["slotNumber"] != "0x7b" {
		t.Fatalf("marshaled Amsterdam fields = %v, %v", marshaled["blockAccessListHash"], marshaled["slotNumber"])
	}

	roundTrip, err := rlp.EncodeToBytes(&decoded)
	if err != nil {
		t.Fatal(err)
	}
	if string(roundTrip) != string(enc) {
		t.Fatal("Amsterdam header RLP differs from the expected field sequence")
	}
	var fromRLP Header
	if err := rlp.DecodeBytes(roundTrip, &fromRLP); err != nil {
		t.Fatal(err)
	}
	if fromRLP.Hash() != want {
		t.Fatalf("RLP round trip hash = %s, want %s", fromRLP.Hash(), want)
	}
	copy := CopyHeader(&decoded)
	if copy.BlockAccessListHash == decoded.BlockAccessListHash || copy.SlotNumber == decoded.SlotNumber {
		t.Fatal("CopyHeader shared Amsterdam field pointers")
	}
}
