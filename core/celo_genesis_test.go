package core

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
)

func TestToBlockWithRoot(t *testing.T) {
	nonMigrated := &Genesis{Config: &params.ChainConfig{}}
	var cel2Time uint64 = 1
	migrated := &Genesis{
		Config: &params.ChainConfig{
			Cel2Time: &cel2Time,
		},
	}

	t.Run("PreGingerbreadBlocksProducedWhenInitingGenesis", func(t *testing.T) {
		g := &Genesis{
			Config: &params.ChainConfig{
				GingerbreadBlock: big.NewInt(1),
			},
			initingGenesis: true,
		}
		b := g.toBlockWithRoot(common.Hash{}, common.Hash{})
		assert.True(t, b.Header().IsPreGingerbread())
	})
	t.Run("PreGingerbreadBlocksNotProducedWhenNotInitingGenesis", func(t *testing.T) {
		g := &Genesis{
			Config: &params.ChainConfig{
				GingerbreadBlock: big.NewInt(1),
			},
			initingGenesis: false,
		}
		b := g.toBlockWithRoot(common.Hash{}, common.Hash{})
		assert.False(t, b.Header().IsPreGingerbread())
	})
	t.Run("PreGingerbreadBlocksNotProducedWhenNotPreGingerbread", func(t *testing.T) {
		g := &Genesis{
			Config: &params.ChainConfig{
				GingerbreadBlock: big.NewInt(0),
			},
		}
		g.initingGenesis = false
		b := g.toBlockWithRoot(common.Hash{}, common.Hash{})
		assert.False(t, b.Header().IsPreGingerbread())
		g.initingGenesis = true
		b = g.toBlockWithRoot(common.Hash{}, common.Hash{})
		assert.False(t, b.Header().IsPreGingerbread())
	})
	t.Run("NonMigratedChainHasDefaultGasLimit&DifficultySet", func(t *testing.T) {
		b := nonMigrated.toBlockWithRoot(common.Hash{}, common.Hash{})
		assert.Equal(t, params.GenesisGasLimit, b.Header().GasLimit)
		assert.Equal(t, params.GenesisDifficulty, b.Header().Difficulty)
	})
	t.Run("MigratedChainGasLimitNotSet", func(t *testing.T) {
		b := migrated.toBlockWithRoot(common.Hash{}, common.Hash{})
		assert.Equal(t, uint64(0), b.Header().GasLimit)
	})
	t.Run("MigratedChainNilDifficultyInitialisedToZero", func(t *testing.T) {
		b := migrated.toBlockWithRoot(common.Hash{}, common.Hash{})
		assert.Equal(t, big.NewInt(0), b.Header().Difficulty)
	})
	t.Run("NonMigratedChainHasEmptyUncleHashSet", func(t *testing.T) {
		b := nonMigrated.toBlockWithRoot(common.Hash{}, common.Hash{})
		assert.Equal(t, types.EmptyUncleHash, b.Header().UncleHash)
	})
	t.Run("PreGingerbreadBlocksOnMigratedChainsDoNotHaveEmptyUncleHashSet", func(t *testing.T) {
		g := &Genesis{
			Config: &params.ChainConfig{
				Cel2Time:         &cel2Time,
				GingerbreadBlock: big.NewInt(1),
			},
		}
		b := g.toBlockWithRoot(common.Hash{}, common.Hash{})
		assert.Equal(t, common.Hash{}, b.Header().UncleHash)
	})
}
