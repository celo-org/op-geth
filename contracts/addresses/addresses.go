package addresses

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

type CeloAddresses struct {
	CeloToken            common.Address
	FeeHandler           common.Address
	FeeCurrencyDirectory common.Address
}

var (
	MainnetAddresses = &CeloAddresses{
		CeloToken:            common.HexToAddress("0x471ece3750da237f93b8e339c536989b8978a438"),
		FeeHandler:           common.HexToAddress("0xcd437749e43a154c07f3553504c68fbfd56b8778"),
		FeeCurrencyDirectory: common.HexToAddress("0x9212Fb72ae65367A7c887eC4Ad9bE310BAC611BF"), // TODO
	}

	AlfajoresAddresses = &CeloAddresses{
		CeloToken:            common.HexToAddress("0xF194afDf50B03e69Bd7D057c1Aa9e10c9954E4C9"),
		FeeHandler:           common.HexToAddress("0xEAaFf71AB67B5d0eF34ba62Ea06Ac3d3E2dAAA38"),
		FeeCurrencyDirectory: common.HexToAddress("0x9212Fb72ae65367A7c887eC4Ad9bE310BAC611BF"),
	}

	BaklavaAddresses = &CeloAddresses{
		CeloToken:            common.HexToAddress("0xdDc9bE57f553fe75752D61606B94CBD7e0264eF8"),
		FeeHandler:           common.HexToAddress("0xeed0A69c51079114C280f7b936C79e24bD94013e"),
		FeeCurrencyDirectory: common.HexToAddress("0xD59E1599F45e42Eb356202B2C714D6C7b734C034"),
	}
)

// GetAddresses returns the addresses for the given chainID.
func GetAddresses(chainID *big.Int) *CeloAddresses {
	// ChainID can be uninitialized in some tests
	if chainID == nil {
		return MainnetAddresses
	}

	switch chainID.Uint64() {
	case params.CeloAlfajoresChainID:
		return AlfajoresAddresses
	case params.CeloBaklavaChainID:
		return BaklavaAddresses
	default:
		return MainnetAddresses
	}
}
