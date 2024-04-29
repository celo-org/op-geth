package rawdb

import (
	"bytes"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
)

// Extra hash comparison is necessary since ancient database only maintains
// the canonical data.
func headerHash(data []byte) common.Hash {
	header := new(types.Header)
	if err := rlp.Decode(bytes.NewReader(data), header); err != nil {
		log.Error("Error decoding stored block header", "err", err)
	}

	return header.Hash()
}
