package legacypool

import (
	"github.com/holiman/uint256"
	"github.com/tenderly/net-celo/common"
	"github.com/tenderly/net-celo/contracts"
	"github.com/tenderly/net-celo/core/types"
	"github.com/tenderly/net-celo/log"
)

// filter Filters transactions from the given list, according to remaining balance (per currency)
// and gasLimit. Returns drops and invalid txs.
func (pool *LegacyPool) filter(list *list, addr common.Address, gasLimit uint64) (types.Transactions, types.Transactions) {
	// CELO: drop all transactions that no longer have a registered currency
	dropsAllowlist, invalidsAllowlist := list.FilterAllowlisted(pool.feeCurrencyContext.ExchangeRates)
	// Check from which currencies we need to get balances
	currenciesInList := list.FeeCurrencies()
	drops, invalids := list.Filter(pool.getBalances(addr, currenciesInList), gasLimit)
	totalDrops := append(dropsAllowlist, drops...)
	totalInvalids := append(invalidsAllowlist, invalids...)
	return totalDrops, totalInvalids
}

func (pool *LegacyPool) getBalances(address common.Address, currencies []common.Address) map[common.Address]*uint256.Int {
	balances := make(map[common.Address]*uint256.Int, len(currencies))
	for _, curr := range currencies {
		balances[curr] = uint256.MustFromBig(contracts.GetFeeBalance(pool.celoBackend, address, &curr))
	}
	return balances
}

func (pool *LegacyPool) recreateCeloProperties() {
	pool.celoBackend = &contracts.CeloBackend{
		ChainConfig: pool.chainconfig,
		State:       pool.currentState,
	}
	feeCurrencyContext, err := contracts.GetFeeCurrencyContext(pool.celoBackend)
	if err != nil {
		log.Error("Error trying to get fee currency context in txpool.", "cause", err)
	}

	pool.feeCurrencyContext = feeCurrencyContext
}
