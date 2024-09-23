package params

import (
	"math/big"
)

const (
	CeloMainnetChainID   = 42220
	CeloBaklavaChainID   = 62320
	CeloAlfajoresChainID = 44787
	CeloSepoliaChainID   = 11142220
)

var (
	// This config should be kept up to date with our mainnet config so that the --dev flag produces
	// results as close as possible to mainnet.
	DevChainConfig = &ChainConfig{
		ChainID: big.NewInt(1337),

		// Ethereum forks
		HomesteadBlock:      big.NewInt(0),
		DAOForkBlock:        nil,
		DAOForkSupport:      false,
		EIP150Block:         big.NewInt(0),
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
		PetersburgBlock:     big.NewInt(0),
		IstanbulBlock:       big.NewInt(0),
		MuirGlacierBlock:    big.NewInt(0),
		BerlinBlock:         big.NewInt(0),
		LondonBlock:         big.NewInt(0),
		ArrowGlacierBlock:   big.NewInt(0),
		GrayGlacierBlock:    big.NewInt(0),
		MergeNetsplitBlock:  big.NewInt(0),
		ShanghaiTime:        newUint64(0),
		CancunTime:          newUint64(0),
		PragueTime:          nil,
		VerkleTime:          nil,

		// Optimism forks
		BedrockBlock: big.NewInt(0),
		RegolithTime: newUint64(0),
		CanyonTime:   newUint64(0),
		EcotoneTime:  newUint64(0),
		FjordTime:    newUint64(0),
		GraniteTime:  newUint64(0),
		HoloceneTime: nil,
		IsthmusTime:  nil,
		InteropTime:  nil,

		// Celo forks
		Cel2Time:         newUint64(0),
		GingerbreadBlock: big.NewInt(0),

		TerminalTotalDifficulty: big.NewInt(0),

		// Consensus engines
		Ethash: nil,
		Clique: nil,

		Optimism: &OptimismConfig{
			EIP1559Denominator:       400,
			EIP1559DenominatorCanyon: newUint64(400),
			EIP1559Elasticity:        5,
		},
		Celo: &CeloConfig{
			// 25000000000 is the base fee floor for mainnet, we use a lower value for dev mode
			// because the real value breaks many of our e2e tests.
			EIP1559BaseFeeFloor: 25000000000,
		},
	}
)
