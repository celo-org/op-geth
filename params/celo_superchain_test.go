package params

import (
	"testing"

	"github.com/ethereum/go-ethereum/superchain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadChainConfig(t *testing.T, chainID uint64) *ChainConfig {
	t.Helper()
	chain, err := superchain.GetChain(chainID)
	require.NoError(t, err)
	chConfig, err := chain.Config()
	require.NoError(t, err)
	cfg, err := LoadOPStackChainConfig(chConfig)
	require.NoError(t, err)
	return cfg
}

func TestNonCeloChainHasNoCeloFields(t *testing.T) {
	t.Run("OP Mainnet", func(t *testing.T) {
		cfg := loadChainConfig(t, OPMainnetChainID)
		assert.Nil(t, cfg.Cel2Time)
		assert.Nil(t, cfg.Celo)
	})

	t.Run("Base Mainnet", func(t *testing.T) {
		cfg := loadChainConfig(t, BaseMainnetChainID)
		assert.Nil(t, cfg.Cel2Time)
		assert.Nil(t, cfg.Celo)
	})
}

func TestCeloChainHasCeloFields(t *testing.T) {
	t.Run("Celo Mainnet", func(t *testing.T) {
		cfg := loadChainConfig(t, CeloMainnetChainID)
		assert.NotNil(t, cfg.Cel2Time)
		require.NotNil(t, cfg.Celo)
		assert.NotZero(t, cfg.Celo.EIP1559BaseFeeFloor)
	})

	t.Run("Celo Sepolia", func(t *testing.T) {
		cfg := loadChainConfig(t, CeloSepoliaChainID)
		assert.NotNil(t, cfg.Cel2Time)
		require.NotNil(t, cfg.Celo)
		assert.NotZero(t, cfg.Celo.EIP1559BaseFeeFloor)
	})
}
