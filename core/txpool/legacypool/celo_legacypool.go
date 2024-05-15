package legacypool

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

// filter Filters transactions from the given list, according to remaining balance (per currency)
// and gasLimit. Returns drops and invalid txs.
func (pool *LegacyPool) filter(list *list, addr common.Address, gasLimit uint64) (types.Transactions, types.Transactions) {
	// CELO: drop all transactions that no longer have a whitelisted currency
	dropsWhitelist, invalidsWhitelist := list.FilterWhitelisted(pool.currentRates)
	// Check from which currencies we need to get balances
	currenciesInList := list.FeeCurrencies()
	drops, invalids := list.Filter(pool.getBalances(addr, currenciesInList), gasLimit)
	totalDrops := append(dropsWhitelist, drops...)
	totalInvalids := append(invalidsWhitelist, invalids...)
	return totalDrops, totalInvalids
}

func (pool *LegacyPool) getBalances(address common.Address, currencies []common.Address) map[common.Address]*big.Int {
	balances := make(map[common.Address]*big.Int, len(currencies))
	for _, curr := range currencies {
		balance, err := contracts.GetFeeBalance(pool.celoBackend, address, &curr)
		if err != nil {
			log.Error(
				"Failed to retrieve fee-balance, assuming zero balance",
				"error", err,
				"account", address,
				"fee-currency", curr,
			)
			balance = new(big.Int)
		}
		balances[curr] = balance
	}
	return balances
}

func (pool *LegacyPool) recreateCeloProperties() {
	pool.celoBackend = &contracts.CeloBackend{
		ChainConfig: pool.chainconfig,
		State:       pool.currentState,
	}
	currentRates, err := contracts.GetExchangeRates(pool.celoBackend)
	if err != nil {
		log.Error("Error trying to get exchange rates in txpool.", "cause", err)
	}
	pool.currentRates = currentRates
}
