package legacypool

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/txpool"
	"github.com/ethereum/go-ethereum/core/types"
)

func celoFilterWhitelisted(blockNumber *big.Int, list *list, all *lookup, fcv txpool.FeeCurrencyValidator) {
	removed := list.txs.Filter(func(tx *types.Transaction) bool {
		return txpool.IsFeeCurrencyTx(tx) && fcv.IsWhitelisted(tx.FeeCurrency(), blockNumber)
	})
	for _, tx := range removed {
		hash := tx.Hash()
		all.Remove(hash)
	}
}

func balanceMinusL1Cost(st *state.StateDB, l1Cost *big.Int,
	feeCurrency *common.Address, balance *big.Int,
	fcv txpool.FeeCurrencyValidator) *big.Int {
	currencyL1Cost := fcv.ToCurrencyValue(st, l1Cost, feeCurrency)
	return new(big.Int).Sub(balance, currencyL1Cost)
}

func celoFilterBalance(st *state.StateDB, addr common.Address, list *list, l1Cost *big.Int,
	gasLimit uint64,
	fcv txpool.FeeCurrencyValidator) (types.Transactions, types.Transactions) {

	// TODO: needs to filter out txs by gas limit, and by balance-l1cost for txs
	// disregarding currency.

	// drops, invalids := list.Filter(balance, gasLimit)
	return nil, nil
}
