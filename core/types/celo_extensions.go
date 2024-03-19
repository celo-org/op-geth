package types

import (
	"errors"
	"fmt"
	"io"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
)

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

func (h *Header) EncodeRLP(w io.Writer) error {
	if (h.UncleHash == common.Hash{}) {
		// Before gingerbread hardfork Celo did not include all of
		// Ethereum's header fields. In that case we must omit the new
		// fields from the header when encoding as RLP to maintain the same encoding and hashes.
		// `UncleHash` is a safe way to check, since it is the zero hash before
		// gingerbread and non-zero after.
		rlpFields := []interface{}{
			h.ParentHash,
			h.Coinbase,
			h.Root,
			h.TxHash,
			h.ReceiptHash,
			h.Bloom,
			h.Number,
			h.GasUsed,
			h.Time,
			h.Extra,
		}
		return rlp.Encode(w, rlpFields)
	} else {
		rlpFields := []interface{}{
			h.ParentHash,
			h.UncleHash,
			h.Coinbase,
			h.Root,
			h.TxHash,
			h.ReceiptHash,
			h.Bloom,
			h.Difficulty,
			h.Number,
			h.GasLimit,
			h.GasUsed,
			h.Time,
			h.Extra,
			h.MixDigest,
			h.Nonce,
			h.BaseFee,
		}
		return rlp.Encode(w, rlpFields)
	}
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

var (
	IstanbulExtraVanity = 32 // Fixed number of extra-data bytes reserved for validator vanity

	// ErrInvalidIstanbulHeaderExtra is returned if the length of extra-data is less than 32 bytes
	ErrInvalidIstanbulHeaderExtra = errors.New("invalid istanbul header extra-data")
)

type IstanbulAggregatedSeal struct {
	// Bitmap is a bitmap having an active bit for each validator that signed this block
	Bitmap *big.Int
	// Signature is an aggregated BLS signature resulting from signatures by each validator that signed this block
	Signature []byte
	// Round is the round in which the signature was created.
	Round *big.Int
}

type IstanbulExtra struct {
	// AddedValidators are the validators that have been added in the block
	AddedValidators []common.Address
	// AddedValidatorsPublicKeys are the BLS public keys for the validators added in the block
	AddedValidatorsPublicKeys []rlp.RawValue
	// RemovedValidators is a bitmap having an active bit for each removed validator in the block
	RemovedValidators *big.Int
	// Seal is an ECDSA signature by the proposer
	Seal []byte
	// AggregatedSeal contains the aggregated BLS signature created via IBFT consensus.
	AggregatedSeal IstanbulAggregatedSeal
	// ParentAggregatedSeal contains and aggregated BLS signature for the previous block.
	ParentAggregatedSeal IstanbulAggregatedSeal
}

// ExtractIstanbulExtra extracts all values of the IstanbulExtra from the header. It returns an
// error if the length of the given extra-data is less than 32 bytes or the extra-data can not
// be decoded.
func extractIstanbulExtra(h *Header) (*IstanbulExtra, error) {
	if len(h.Extra) < IstanbulExtraVanity {
		return nil, ErrInvalidIstanbulHeaderExtra
	}

	var istanbulExtra *IstanbulExtra
	err := rlp.DecodeBytes(h.Extra[IstanbulExtraVanity:], &istanbulExtra)
	if err != nil {
		return nil, err
	}
	return istanbulExtra, nil
}

// IstanbulFilteredHeader returns a filtered header which some information (like seal, aggregated signature)
// are clean to fulfill the Istanbul hash rules. It returns nil if the extra-data cannot be
// decoded/encoded by rlp.
func IstanbulFilteredHeader(h *Header, keepSeal bool) *Header {
	newHeader := CopyHeader(h)
	istanbulExtra, err := extractIstanbulExtra(newHeader)
	if err != nil {
		return nil
	}

	if !keepSeal {
		istanbulExtra.Seal = []byte{}
	}
	istanbulExtra.AggregatedSeal = IstanbulAggregatedSeal{}

	payload, err := rlp.EncodeToBytes(&istanbulExtra)
	if err != nil {
		return nil
	}

	newHeader.Extra = append(newHeader.Extra[:IstanbulExtraVanity], payload...)

	return newHeader
}
