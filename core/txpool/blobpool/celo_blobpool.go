package blobpool

import (
	"github.com/ethereum/go-ethereum/contracts"
	"github.com/ethereum/go-ethereum/log"
)

func (pool *BlobPool) recreateCeloProperties() {
	head := pool.head.Load()
	pool.celoBackend = &contracts.CeloBackend{
		ChainConfig: pool.chain.Config(),
		State:       pool.state,
		BlockNumber: head.Number,
		Time:        head.Time,
	}
	currencyContext, err := contracts.GetFeeCurrencyContext(pool.celoBackend)
	if err != nil {
		log.Error("Error trying to get fee currency context in txpool.", "cause", err)
	}
	pool.feeCurrencyContext = currencyContext
}
