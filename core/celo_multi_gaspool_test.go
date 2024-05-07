package core

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestMultiCurrencyGasPool(t *testing.T) {
	blockGasLimit := uint64(1_000)
	subGasAmount := 100

	cUSDToken := common.HexToAddress("0x765DE816845861e75A25fCA122bb6898B8B1282a")
	cEURToken := common.HexToAddress("0xD8763CBa276a3738E6DE85b4b3bF5FDed6D6cA73")

	testCases := []struct {
		name                string
		feeCurrency         *FeeCurrency
		defaultLimit        float64
		limits              FeeCurrencyLimitMapping
		defaultPoolExpected bool
		expectedValue       uint64
	}{
		{
			name:                "Empty mapping, CELO uses default pool",
			feeCurrency:         nil,
			defaultLimit:        0.9,
			limits:              map[FeeCurrency]float64{},
			defaultPoolExpected: true,
			expectedValue:       900, // blockGasLimit - subGasAmount
		},
		{
			name:         "Non-empty mapping, CELO uses default pool",
			feeCurrency:  nil,
			defaultLimit: 0.9,
			limits: map[FeeCurrency]float64{
				cUSDToken: 0.5,
			},
			defaultPoolExpected: true,
			expectedValue:       900, // blockGasLimit - subGasAmount
		},
		{
			name:                "Empty mapping, currency fallbacks to the default limit",
			feeCurrency:         &cUSDToken,
			defaultLimit:        0.9,
			limits:              map[FeeCurrency]float64{},
			defaultPoolExpected: false,
			expectedValue:       800, // blockGasLimit * defaultLimit- subGasAmount
		},
		{
			name:         "Non-empty mapping, currency uses default limit",
			feeCurrency:  &cEURToken,
			defaultLimit: 0.9,
			limits: map[FeeCurrency]float64{
				cUSDToken: 0.5,
			},
			defaultPoolExpected: false,
			expectedValue:       800, // blockGasLimit * defaultLimit - subGasAmount
		},
		{
			name:         "Non-empty mapping, configured currency uses configured limits",
			feeCurrency:  &cUSDToken,
			defaultLimit: 0.9,
			limits: map[FeeCurrency]float64{
				cUSDToken: 0.5,
			},
			defaultPoolExpected: false,
			expectedValue:       400, // blockGasLimit * 0.5 - subGasAmount
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			mgp := NewMultiGasPool(
				blockGasLimit,
				c.defaultLimit,
				c.limits,
			)

			pool := mgp.GetPool(c.feeCurrency)
			pool.SubGas(uint64(subGasAmount))

			if c.defaultPoolExpected {
				result := mgp.GetPool(nil).Gas()
				if result != c.expectedValue {
					t.Error("Default pool expected", c.expectedValue, "got", result)
				}
			} else {
				result := mgp.GetPool(c.feeCurrency).Gas()

				if result != c.expectedValue {
					t.Error(
						"Expected pool", c.feeCurrency, "value", c.expectedValue,
						"got", result,
					)
				}
			}
		})
	}
}
