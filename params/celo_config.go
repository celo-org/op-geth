package params

import "math/big"

const (
	CeloMainnetChainID   = 42220
	CeloAlfajoresChainID = 44787
	CeloBaklavaChainID   = 62320
)

// GasLimits holds the gas limit changes for a given chain
type GasLimits struct {
	changes []LimitChange
}

type LimitChange struct {
	block    *big.Int
	gasLimit uint64
}

// Limit returns the gas limit at a given block number
func (g *GasLimits) Limit(block *big.Int) uint64 {
	// Grab the gas limit at block 0
	curr := g.changes[0].gasLimit
	for _, c := range g.changes[1:] {
		if block.Cmp(c.block) < 0 {
			return curr
		}
		curr = c.gasLimit
	}
	return curr
}

var (
	// Hardcoded set of gas limit changes derived from historical state of Celo L1 chain
	// Ported from celo-blockchain (https://github.com/celo-org/celo-blockchain/blob/master/params/config.go#L189-L220)
	mainnetGasLimits = &GasLimits{
		changes: []LimitChange{
			{big.NewInt(0), 20e6},
			{big.NewInt(3317), 10e6},
			{big.NewInt(3251772), 13e6},
			{big.NewInt(6137285), 20e6},
			{big.NewInt(13562578), 50e6},
			{big.NewInt(14137511), 20e6},
			{big.NewInt(21355415), 32e6},
		},
	}

	alfajoresGasLimits = &GasLimits{
		changes: []LimitChange{
			{big.NewInt(0), 20e6},
			{big.NewInt(912), 10e6},
			{big.NewInt(1392355), 130e6},
			{big.NewInt(1507905), 13e6},
			{big.NewInt(4581182), 20e6},
			{big.NewInt(11143973), 35e6},
		},
	}

	baklavaGasLimits = &GasLimits{
		changes: []LimitChange{
			{big.NewInt(0), 20e6},
			{big.NewInt(1230), 10e6},
			{big.NewInt(1713181), 130e6},
			{big.NewInt(1945003), 13e6},
			{big.NewInt(15158971), 20e6},
		},
	}

	PreGingerbreadNetworkGasLimits = map[uint64]*GasLimits{
		CeloMainnetChainID:   mainnetGasLimits,
		CeloAlfajoresChainID: alfajoresGasLimits,
		CeloBaklavaChainID:   baklavaGasLimits,
	}
)
