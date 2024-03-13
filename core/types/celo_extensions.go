package types

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
)

type IstanbulExtra rlp.RawValue

type beforeGingerbreadHeader struct {
	ParentHash  common.Hash    `json:"parentHash"       gencodec:"required"`
	Coinbase    common.Address `json:"miner"            gencodec:"required"`
	Root        common.Hash    `json:"stateRoot"        gencodec:"required"`
	TxHash      common.Hash    `json:"transactionsRoot" gencodec:"required"`
	ReceiptHash common.Hash    `json:"receiptsRoot"     gencodec:"required"`
	Bloom       Bloom          `json:"logsBloom"        gencodec:"required"`
	Number      *big.Int       `json:"number"           gencodec:"required"`
	GasUsed     uint64         `json:"gasUsed"          gencodec:"required"`
	Time        uint64         `json:"timestamp"        gencodec:"required"`
	Extra       []byte         `json:"extraData"        gencodec:"required"`

	// Used to cache deserialized istanbul extra data
	extraLock  sync.Mutex
	extraValue *IstanbulExtra
	extraError error
}

// This type is required to avoid an infinite loop when decoding
type afterGingerbreadHeader Header

func (h *Header) DecodeRLP(s *rlp.Stream) error {
	_, size, _ := s.Kind()
	var raw rlp.RawValue
	err := s.Decode(&raw)
	if err != nil {
		return err
	}
	headerSize := len(raw) - int(size)
	numElems, err := rlp.CountValues(raw[headerSize:])
	if err != nil {
		return err
	}
	if numElems == 10 {
		// Before gingerbread
		decodedHeader := beforeGingerbreadHeader{}
		err = rlp.DecodeBytes(raw, &decodedHeader)

		h.ParentHash = decodedHeader.ParentHash
		h.Coinbase = decodedHeader.Coinbase
		h.Root = decodedHeader.Root
		h.TxHash = decodedHeader.TxHash
		h.ReceiptHash = decodedHeader.ReceiptHash
		h.Bloom = decodedHeader.Bloom
		h.Number = decodedHeader.Number
		h.GasUsed = decodedHeader.GasUsed
		h.Time = decodedHeader.Time
		h.Extra = decodedHeader.Extra
	} else {
		// After gingerbread
		decodedHeader := afterGingerbreadHeader{}
		err = rlp.DecodeBytes(raw, &decodedHeader)

		h.ParentHash = decodedHeader.ParentHash
		h.UncleHash = decodedHeader.UncleHash
		h.Coinbase = decodedHeader.Coinbase
		h.Root = decodedHeader.Root
		h.TxHash = decodedHeader.TxHash
		h.ReceiptHash = decodedHeader.ReceiptHash
		h.Bloom = decodedHeader.Bloom
		h.Difficulty = decodedHeader.Difficulty
		h.Number = decodedHeader.Number
		h.GasLimit = decodedHeader.GasLimit
		h.GasUsed = decodedHeader.GasUsed
		h.Time = decodedHeader.Time
		h.Extra = decodedHeader.Extra
		h.MixDigest = decodedHeader.MixDigest
		h.Nonce = decodedHeader.Nonce
		h.BaseFee = decodedHeader.BaseFee
	}

	return err
}

type CeloBody struct {
	Transactions []*Transaction

	// These fields were custom additions to the celo block body, but neither
	// are required in CEL2, so we "skip" decoding them by setting their value
	// to rlp.RawValue.
	Randomness     rlp.RawValue
	EpochSnarkData rlp.RawValue
}

// This type is required to avoid an infinite loop when decoding
type opBody Body

func (b *Body) DecodeRLP(s *rlp.Stream) error {
	// Celo block bodies differ from op block bodies in that the second field
	// is Randomness which is a struct containing 2 hashes and Randomness is
	// never nil. We can use this to detect celo block bodies, because in
	// contrast optimism has a list of Headers as the second field which is
	// either empty or with one header is far larger than 2 hashes (size 66
	// each hash is 1 byte specifying the length in bytes and then 32 bytes for
	// the hash)

	var data rlp.RawValue
	err := s.Decode(&data)
	if err != nil {
		return err
	}
	var remaining []byte = data

	kind, tagSize, _, err := rlp.ReadKind(remaining)
	if err != nil {
		panic(fmt.Sprintf("Error decoding block body: %v", err))
	}
	// At the top level we expect a list
	if kind != rlp.List {
		panic(fmt.Sprintf("Expecting encoded body to be an rlp list, instead got %v", kind))
	}
	// We skip over the list tag so we can dig into the list data.
	remaining = remaining[tagSize:]

	kind, tagSize, contentSize, err := rlp.ReadKind(remaining)
	if err != nil {
		panic(fmt.Sprintf("Error decoding block body: %v", err))
	}
	// Now we expect a list of transactions which we wish to skip over.
	if kind != rlp.List {
		panic(fmt.Sprintf("Expecting encoded body to be an rlp list, instead got %v", kind))
	}
	remaining = remaining[tagSize+contentSize:]

	// Now we expect either a list of headers or the randomness which will have tagSize 2 and content size 66.
	kind, tagSize, contentSize, err = rlp.ReadKind(remaining)
	if err != nil {
		panic(fmt.Sprintf("Error decoding block body: %v", err))
	}
	if kind != rlp.List {
		panic(fmt.Sprintf("Expecting encoded body to be an rlp list, instead got %v", kind))
	}
	// celo block body
	if tagSize == 2 && contentSize == 66 {
		body := new(CeloBody)
		if err := rlp.DecodeBytes(data, body); err != nil {
			log.Error("Invalid block body 1 RLP", "err", err)
			return nil
		}
		b.Transactions = body.Transactions
	} else {
		body := new(opBody)
		if err := rlp.DecodeBytes(data, body); err != nil {
			log.Error("Invalid block body 2 RLP", "err", err)
			return nil
		}
		b.Transactions = body.Transactions
		b.Uncles = body.Uncles
		b.Withdrawals = body.Withdrawals
	}
	return nil
}
