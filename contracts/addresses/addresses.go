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
		FeeCurrencyDirectory: common.HexToAddress("0x15F344b9E6c3Cb6F0376A36A64928b13F62C6276"),
	}

	CeloSepoliaAddresses = &CeloAddresses{
		CeloToken:            common.HexToAddress("0x471EcE3750Da237f93B8E339c536989b8978a438"),
		FeeHandler:           common.HexToAddress("0xcD437749E43A154C07F3553504c68fBfD56B8778"),
		FeeCurrencyDirectory: common.HexToAddress("0x9212Fb72ae65367A7c887eC4Ad9bE310BAC611BF"),
	}
)

// GetAddresses returns the addresses for the given chainID or
// nil if not found.
func GetAddresses(chainID *big.Int) *CeloAddresses {
	// ChainID can be uninitialized in some tests
	if chainID == nil {
		return nil
	}
	switch chainID.Uint64() {
	case params.CeloSepoliaChainID:
		return CeloSepoliaAddresses
	case params.CeloMainnetChainID:
		return MainnetAddresses
	default:
		return nil
	}
}

// GetAddressesOrDefault returns the addresses for the given chainID or
// the Mainnet addresses if none are found.
func GetAddressesOrDefault(chainID *big.Int, defaultValue *CeloAddresses) *CeloAddresses {
	addresses := GetAddresses(chainID)
	if addresses == nil {
		return defaultValue
	}
	return addresses
}
