package core

import "github.com/ethereum/go-ethereum/common"

type FeeCurrency = common.Address

// MultiGasPool tracks the amount of gas available during execution
// of the transactions in a block per fee currency. The zero value is a pool
// with zero gas available.
type MultiGasPool struct {
	pools       map[FeeCurrency]*GasPool
	defaultPool *GasPool

	blockGasLimit uint64
	defaultLimit  float64
}

type FeeCurrencyLimitMapping = map[FeeCurrency]float64

// NewMultiGasPool creates a multi-fee currency gas pool and a default fallback
// pool for CELO
func NewMultiGasPool(
	blockGasLimit uint64,
	defaultLimit float64,
	limitsMapping FeeCurrencyLimitMapping,
) *MultiGasPool {
	pools := make(map[FeeCurrency]*GasPool, len(limitsMapping))
	// A special case for CELO which doesn't have a limit
	celoPool := new(GasPool).AddGas(blockGasLimit)
	mgp := &MultiGasPool{
		pools:         pools,
		defaultPool:   celoPool,
		blockGasLimit: blockGasLimit,
		defaultLimit:  defaultLimit,
	}
	for feeCurrency, fraction := range limitsMapping {
		mgp.getOrInitPool(feeCurrency, &fraction)
	}
	return mgp
}

func (mgp MultiGasPool) getOrInitPool(c FeeCurrency, fraction *float64) *GasPool {
	if gp, ok := mgp.pools[c]; ok {
		return gp
	}
	if fraction == nil {
		fraction = &mgp.defaultLimit
	}
	gp := new(GasPool).AddGas(
		uint64(float64(mgp.blockGasLimit) * *fraction),
	)
	mgp.pools[c] = gp
	return gp
}

// GetPool returns an initialised pool for the given fee currency or
// initialises and returns a new pool with a default limit.
// For a `nil` FeeCurrency value, it returns the default pool.
func (mgp MultiGasPool) GetPool(c *FeeCurrency) *GasPool {
	if c == nil {
		return mgp.defaultPool
	}
	// Use the default fraction here because the configured limits'
	// pools have been created already in the constructor.
	return mgp.getOrInitPool(*c, nil)
}
