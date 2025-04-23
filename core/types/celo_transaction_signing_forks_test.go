package types

import (
	"testing"

	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
)

// Test_forks_activeForks tests that the correct forks are returned for a given block time and chain config
func Test_forks_activeForks(t *testing.T) {
	t.Parallel()

	cel2Time := uint64(1000)

	t.Run("Non-Celo", func(t *testing.T) {
		config := &params.ChainConfig{
			Cel2Time: nil,
		}
		assert.Equal(t, []fork(nil), celoForks.activeForks(1000, config))
	})

	t.Run("Celo1", func(t *testing.T) {
		config := &params.ChainConfig{
			Cel2Time: &cel2Time,
		}
		assert.Equal(t, []fork{&celoLegacy{}}, celoForks.activeForks(500, config))
	})

	t.Run("Celo2", func(t *testing.T) {
		config := &params.ChainConfig{
			Cel2Time: &cel2Time,
		}
		assert.Equal(t, []fork{&cel2{}, &celoLegacy{}}, celoForks.activeForks(1000, config))
	})
}
