package addresses

import "github.com/ethereum/go-ethereum/common"

var (
	CeloTokenAddress            = common.HexToAddress("0x471ece3750da237f93b8e339c536989b8978a438")
	FeeHandlerAddress           = common.HexToAddress("0xcd437749e43a154c07f3553504c68fbfd56b8778")
	FeeCurrencyDirectoryAddress = common.HexToAddress("0x9212Fb72ae65367A7c887eC4Ad9bE310BAC611BF")

	CeloTokenAlfajoresAddress  = common.HexToAddress("0xF194afDf50B03e69Bd7D057c1Aa9e10c9954E4C9")
	FeeHandlerAlfajoresAddress = common.HexToAddress("0xEAaFf71AB67B5d0eF34ba62Ea06Ac3d3E2dAAA38")

	CeloTokenBaklavaAddress  = common.HexToAddress("0xdDc9bE57f553fe75752D61606B94CBD7e0264eF8")
	FeeHandlerBaklavaAddress = common.HexToAddress("0xeed0A69c51079114C280f7b936C79e24bD94013e")

	AlfajoresChainID uint64 = 44787
	BaklavaChainID   uint64 = 62320
)
