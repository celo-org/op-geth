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

	tests := []struct {
		name      string
		config    *params.ChainConfig
		blockTime uint64
		expected  []fork
	}{
		{
			name: "Non-Celo",
			config: &params.ChainConfig{
				Cel2Time: nil,
			},
			blockTime: 1000,
			expected:  nil,
		},
		{
			name: "Celo1",
			config: &params.ChainConfig{
				Cel2Time: &cel2Time,
			},
			blockTime: 500,
			expected:  []fork{&celoLegacy{}},
		},
		{
			name: "Celo2",
			config: &params.ChainConfig{
				Cel2Time: &cel2Time,
			},
			blockTime: 1000,
			expected:  []fork{&cel2{}, &celoLegacy{}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := celoForks.activeForks(test.blockTime, test.config)
			assert.Equal(t, test.expected, result)
		})
	}
}

// Test_forks_findTxFuncs tests that the correct txFuncs are returned for a given transaction type and chain config
func Test_forks_findTxFuncs(t *testing.T) {
	t.Parallel()

	cel2Time := uint64(1000)
	config := &params.ChainConfig{
		Cel2Time: &cel2Time,
	}

	test := []struct {
		name      string
		blockTime uint64
		txData    TxData
		expected  *txFuncs
	}{
		// Celo2
		{
			name:      "Celo LegacyTx in Celo2",
			blockTime: 2000,
			txData: &LegacyTx{
				CeloLegacy: true,
			},
			expected: deprecatedTxFuncs,
		},
		{
			name:      "Ethereum LegacyTx in Celo2",
			blockTime: 2000,
			txData: &LegacyTx{
				CeloLegacy: false,
			},
			expected: nil,
		},
		{
			name:      "AccessListTx in Celo2",
			blockTime: 2000,
			txData:    &AccessListTx{},
			expected:  accessListTxFuncs,
		},
		{
			name:      "DynamicFeeTx in Celo2 ",
			blockTime: 2000,
			txData:    &DynamicFeeTx{},
			expected:  dynamicFeeTxFuncs,
		},
		{
			name:      "CeloDynamicFeeTx in Celo2",
			blockTime: 2000,
			txData:    &CeloDynamicFeeTx{},
			expected:  deprecatedTxFuncs,
		},
		{
			name:      "CeloDynamicFeeTxV2 in Celo2",
			blockTime: 2000,
			txData:    &CeloDynamicFeeTxV2{},
			expected:  celoDynamicFeeTxV2Funcs,
		},
		{
			name:      "CeloDenominatedTx in Celo2",
			blockTime: 2000,
			txData:    &CeloDenominatedTx{},
			expected:  nil,
		},
		// Celo1
		{
			name:      "Celo LegacyTx in Celo1",
			blockTime: 100,
			txData: &LegacyTx{
				CeloLegacy: true,
			},
			expected: celoLegacyTxFuncs,
		},
		{
			name:      "Ethereum LegacyTx in Celo1",
			blockTime: 100,
			txData: &LegacyTx{
				CeloLegacy: false,
			},
			expected: nil,
		},
		{
			name:      "AccessListTx in Celo1",
			blockTime: 100,
			txData:    &AccessListTx{},
			expected:  accessListTxFuncs,
		},
		{
			name:      "DynamicFeeTx in Celo1",
			blockTime: 100,
			txData:    &DynamicFeeTx{},
			expected:  dynamicFeeTxFuncs,
		},
		{
			name:      "CeloDynamicFeeTx in Celo1",
			blockTime: 100,
			txData:    &CeloDynamicFeeTx{},
			expected:  celoDynamicFeeTxFuncs,
		},
		{
			name:      "CeloDynamicFeeTxV2 in Celo1",
			blockTime: 100,
			txData:    &CeloDynamicFeeTxV2{},
			expected:  celoDynamicFeeTxV2Funcs,
		},
		{
			name:      "CeloDenominatedTx in Celo1",
			blockTime: 100,
			txData:    &CeloDenominatedTx{},
			expected:  nil,
		},
	}

	for _, test := range test {
		t.Run(test.name, func(t *testing.T) {
			forks := forks(celoForks.activeForks(test.blockTime, config))

			result := forks.findTxFuncs(NewTx(test.txData))

			assert.Equal(t, test.expected, result)
		})
	}
}
