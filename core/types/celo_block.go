package types

import (
	"io"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

//go:generate go run ../../rlp/rlpgen -type BeforeGingerbreadHeader --encoder --decoder -out gen_before_gingerbread_header_rlp.go
//go:generate go run ../../rlp/rlpgen -type AfterGingerbreadHeader --encoder --decoder -out gen_after_gingerbread_header_rlp.go

type IstanbulExtra rlp.RawValue

type BeforeGingerbreadHeader struct {
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
type AfterGingerbreadHeader Header

func (h *Header) DecodeRLP(s *rlp.Stream) error {
	var raw rlp.RawValue
	err := s.Decode(&raw)
	if err != nil {
		return err
	}

	preGingerbread, err := isPreGingerbreadHeader(raw)
	if err != nil {
		return err
	}

	if preGingerbread { // Address
		// Before gingerbread
		decodedHeader := BeforeGingerbreadHeader{}
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
		h.Difficulty = new(big.Int)
	} else {
		// After gingerbread
		decodedHeader := AfterGingerbreadHeader{}
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
		h.WithdrawalsHash = decodedHeader.WithdrawalsHash
		h.BlobGasUsed = decodedHeader.BlobGasUsed
		h.ExcessBlobGas = decodedHeader.ExcessBlobGas
		h.ParentBeaconRoot = decodedHeader.ParentBeaconRoot
	}

	return err
}

// EncodeRLP implements encodes the Header to an RLP data stream.
func (h *Header) EncodeRLP(w io.Writer) error {
	if h.IsPreGingerbread() {
		encodedHeader := BeforeGingerbreadHeader{
			ParentHash:  h.ParentHash,
			Coinbase:    h.Coinbase,
			Root:        h.Root,
			TxHash:      h.TxHash,
			ReceiptHash: h.ReceiptHash,
			Bloom:       h.Bloom,
			Number:      h.Number,
			GasUsed:     h.GasUsed,
			Time:        h.Time,
			Extra:       h.Extra,
		}

		return rlp.Encode(w, &encodedHeader)
	}

	// After gingerbread
	encodedHeader := AfterGingerbreadHeader{
		ParentHash:       h.ParentHash,
		UncleHash:        h.UncleHash,
		Coinbase:         h.Coinbase,
		Root:             h.Root,
		TxHash:           h.TxHash,
		ReceiptHash:      h.ReceiptHash,
		Bloom:            h.Bloom,
		Difficulty:       h.Difficulty,
		Number:           h.Number,
		GasLimit:         h.GasLimit,
		GasUsed:          h.GasUsed,
		Time:             h.Time,
		Extra:            h.Extra,
		MixDigest:        h.MixDigest,
		Nonce:            h.Nonce,
		BaseFee:          h.BaseFee,
		WithdrawalsHash:  h.WithdrawalsHash,
		BlobGasUsed:      h.BlobGasUsed,
		ExcessBlobGas:    h.ExcessBlobGas,
		ParentBeaconRoot: h.ParentBeaconRoot,
	}

	return rlp.Encode(w, &encodedHeader)
}

// isPreGingerbreadHeader introspects the header rlp to check the length of the
// second element of the list (the first element describes the list). Pre
// gingerbread the second element of a header is an address which is 20 bytes
// long, post gingerbread the second element is a hash which is 32 bytes long.
func isPreGingerbreadHeader(buf []byte) (bool, error) {
	var contentSize uint64
	var err error
	for i := 0; i < 3; i++ {
		buf, _, _, contentSize, err = rlp.ReadNext(buf)
		if err != nil {
			return false, err
		}
	}

	return contentSize == common.AddressLength, nil
}

// Returns if the header is a gingerbread header by looking at the gas limit.
func (h *Header) IsPreGingerbread() bool {
	// We check for a pre gingerbread header by looking for (GasLimit == 0)
	// here. We don't use Difficulty because we ensure that headers have a zero
	// difficulty, even if it's not set in the rlp encoded form (we do this
	// because the go ethereum codebase assumed non nil difficulties) and post
	// gingerbread difficulty is hardcoded to zero. Also testing for base fee
	// is not reliable because some older eth blocks had no base fee and they
	// are used in some tests.
	return h.GasLimit == 0
}
