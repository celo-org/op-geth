package legacypool

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func (pool *LegacyPool) filter(list *celo_list, addr common.Address, l1Cost *big.Int, gasLimit uint64) (types.Transactions, types.Transactions) {
	st := pool.currentState
	fcv := pool.feeCurrencyValidator
	// CELO: drop all transactions that no longer have a whitelisted currency
	dropsWhitelist := list.FilterWhitelisted(st, pool.all, fcv)

	drops, invalids := list.FilterBalance(st, addr, l1Cost, gasLimit,
		fcv)
	totalDrops := make(types.Transactions, 0, len(dropsWhitelist)+len(drops))
	totalDrops = append(totalDrops, dropsWhitelist...)
	totalDrops = append(totalDrops, drops...)
	return totalDrops, invalids
}
