package core

import (
	"github.com/ethereum/go-ethereum/common"
)

type FeeCurrency = common.Address

// MultiGasPool tracks the amount of gas available during execution
// of the transactions in a block per fee currency. The zero value is a pool
// with zero gas available.
type MultiGasPool struct {
	pools       map[FeeCurrency]*GasPool
	defaultPool *GasPool
}

type FeeCurrencyLimitMapping = map[FeeCurrency]float64

// NewMultiGasPool creates a multi-fee currency gas pool and a default fallback
// pool for CELO
func NewMultiGasPool(
	blockGasLimit uint64,
	allowlist common.AddressSet,
	defaultLimit float64,
	limitsMapping FeeCurrencyLimitMapping,
	isDeriving bool,
) *MultiGasPool {
	multiGasPool := &MultiGasPool{
		defaultPool: new(GasPool).AddGas(blockGasLimit),
	}

	// we want to deactivate the separate pools for
	// fee-currencies when we are deriving from l1.
	// The parameters are locally configurable,
	// and we don't want them to differ
	// from the "consensus", i.e. the ones
	// that the sequencer used for building.
	if isDeriving {
		return multiGasPool
	}
	multiGasPool.pools = make(map[FeeCurrency]*GasPool, len(allowlist))

	for currency := range allowlist {
		fraction, ok := limitsMapping[currency]
		if !ok {
			fraction = defaultLimit
		}
		multiGasPool.pools[currency] = new(GasPool).AddGas(
			uint64(float64(blockGasLimit) * fraction),
		)
	}

	return multiGasPool
}

// PoolFor returns a configured pool for the given fee currency or the default
// one otherwise, the returned boolean is true when a custom fee currency gas pool is returned.
func (mgp MultiGasPool) PoolFor(feeCurrency *FeeCurrency) (*GasPool, bool) {
	if feeCurrency == nil || mgp.pools[*feeCurrency] == nil {
		return mgp.defaultPool, false
	}

	return mgp.pools[*feeCurrency], true
}

func (mgp MultiGasPool) Copy() *MultiGasPool {
	pools := make(map[FeeCurrency]*GasPool, len(mgp.pools))
	for fc, gp := range mgp.pools {
		gpCpy := *gp
		pools[fc] = &gpCpy
	}
	gpCpy := *mgp.defaultPool
	return &MultiGasPool{
		pools:       pools,
		defaultPool: &gpCpy,
	}
}
