package ethapi

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

func TestRPCMarshalHeaderAmsterdamFields(t *testing.T) {
	balHash := common.HexToHash("0x1234")
	slotNumber := uint64(123)
	header := &types.Header{
		Difficulty:          big.NewInt(0),
		Number:              big.NewInt(1),
		BlockAccessListHash: &balHash,
		SlotNumber:          &slotNumber,
	}
	result := RPCMarshalHeader(header, false)
	if got := result["blockAccessListHash"]; got != header.BlockAccessListHash {
		t.Fatalf("blockAccessListHash = %v, want %v", got, header.BlockAccessListHash)
	}
	if got := result["slotNumber"]; got != hexutil.Uint64(slotNumber) {
		t.Fatalf("slotNumber = %v, want %v", got, slotNumber)
	}
}
