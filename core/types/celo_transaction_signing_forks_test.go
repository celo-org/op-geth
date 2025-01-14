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

// Test_forks_findTxFuncs tests that the correct txFuncs are returned for a given transaction type and chain config
func Test_forks_findTxFuncs(t *testing.T) {
	t.Parallel()

	cel2Time := uint64(1000)
	config := &params.ChainConfig{
		Cel2Time: &cel2Time,
	}

	celo2Forks := forks(celoForks.activeForks(cel2Time, config))
	t.Run("Celo LegacyTx in Celo2", func(t *testing.T) {
		assert.Equal(t, deprecatedTxFuncs, celo2Forks.findTxFuncs(NewTx(&LegacyTx{CeloLegacy: true})))
	})

	t.Run("Ethereum LegacyTx in Celo2", func(t *testing.T) {
		assert.Equal(t, (*txFuncs)(nil), celo2Forks.findTxFuncs(NewTx(&LegacyTx{CeloLegacy: false})))
	})

	t.Run("AccessListTx in Celo2", func(t *testing.T) {
		assert.Equal(t, accessListTxFuncs, celo2Forks.findTxFuncs(NewTx(&AccessListTx{})))
	})

	t.Run("DynamicFeeTx in Celo2", func(t *testing.T) {
		assert.Equal(t, dynamicFeeTxFuncs, celo2Forks.findTxFuncs(NewTx(&DynamicFeeTx{})))
	})

	t.Run("CeloDynamicFeeTx in Celo2", func(t *testing.T) {
		assert.Equal(t, deprecatedTxFuncs, celo2Forks.findTxFuncs(NewTx(&CeloDynamicFeeTx{})))
	})

	t.Run("CeloDynamicFeeTxV2 in Celo2", func(t *testing.T) {
		assert.Equal(t, celoDynamicFeeTxV2Funcs, celo2Forks.findTxFuncs(NewTx(&CeloDynamicFeeTxV2{})))
	})

	celo1Forks := forks(celoForks.activeForks(uint64(100), config))
	t.Run("Celo LegacyTx in Celo1", func(t *testing.T) {
		assert.Equal(t, celoLegacyTxFuncs, celo1Forks.findTxFuncs(NewTx(&LegacyTx{CeloLegacy: true})))
	})

	t.Run("Ethereum LegacyTx in Celo1", func(t *testing.T) {
		assert.Equal(t, (*txFuncs)(nil), celo1Forks.findTxFuncs(NewTx(&LegacyTx{CeloLegacy: false})))
	})

	t.Run("AccessListTx in Celo1", func(t *testing.T) {
		assert.Equal(t, accessListTxFuncs, celo1Forks.findTxFuncs(NewTx(&AccessListTx{})))
	})

	t.Run("DynamicFeeTx in Celo1", func(t *testing.T) {
		assert.Equal(t, dynamicFeeTxFuncs, celo1Forks.findTxFuncs(NewTx(&DynamicFeeTx{})))
	})

	t.Run("CeloDynamicFeeTx in Celo1", func(t *testing.T) {
		assert.Equal(t, celoDynamicFeeTxFuncs, celo1Forks.findTxFuncs(NewTx(&CeloDynamicFeeTx{})))
	})

	t.Run("CeloDynamicFeeTxV2 in Celo1", func(t *testing.T) {
		assert.Equal(t, celoDynamicFeeTxV2Funcs, celo1Forks.findTxFuncs(NewTx(&CeloDynamicFeeTxV2{})))
	})

	t.Run("CeloDenominatedTx in Celo1", func(t *testing.T) {
		assert.Equal(t, (*txFuncs)(nil), celo1Forks.findTxFuncs(NewTx(&CeloDenominatedTx{})))
	})
}
