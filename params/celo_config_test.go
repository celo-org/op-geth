package params

import (
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestGasLimits_Limit(t *testing.T) {
	subTest := func(t *testing.T, name string, chainId uint64, limits []LimitChange) {
		t.Run(name, func(t *testing.T) {
			for i, l := range limits {
				beginningHeight := l.block
				beginningLimit := PreGingerbreadNetworkGasLimits[chainId].Limit(beginningHeight)
				assert.Equal(t, l.gasLimit, beginningLimit, "gas limit at block %d (%s)", beginningHeight, name)

				if i < len(limits)-1 {
					endHeight := new(big.Int).Sub(limits[i+1].block, big.NewInt(1))
					endLimit := PreGingerbreadNetworkGasLimits[chainId].Limit(endHeight)
					assert.Equal(t, l.gasLimit, endLimit, "gas limit at block %d (%s)", endHeight, name)
				}
			}
		})
	}

	subTest(t, "mainnet", CeloMainnetChainID, mainnetGasLimits.changes)
	subTest(t, "alfajores", CeloAlfajoresChainID, alfajoresGasLimits.changes)
	subTest(t, "baklava", CeloBaklavaChainID, baklavaGasLimits.changes)
}
