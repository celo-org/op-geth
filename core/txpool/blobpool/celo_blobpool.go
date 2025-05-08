package blobpool

import (
	"github.com/tenderly/net-celo/contracts"
	"github.com/tenderly/net-celo/log"
)

func (pool *BlobPool) recreateCeloProperties() {
	pool.celoBackend = &contracts.CeloBackend{
		ChainConfig: pool.chain.Config(),
		State:       pool.state,
	}
	currencyContext, err := contracts.GetFeeCurrencyContext(pool.celoBackend)
	if err != nil {
		log.Error("Error trying to get fee currency context in txpool.", "cause", err)
	}
	pool.feeCurrencyContext = currencyContext
}
