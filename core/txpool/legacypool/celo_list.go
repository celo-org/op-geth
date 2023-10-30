package legacypool

import (
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/txpool"
	"github.com/ethereum/go-ethereum/core/types"
)

type celo_list struct {
	list      *list
	totalCost map[common.Address]*big.Int

	// Pointer reference to inner list
	txs *sortedMap
}

func newCeloList(strict bool) *celo_list {
	inner_list := newList(strict)
	return &celo_list{
		list:      inner_list,
		totalCost: make(map[common.Address]*big.Int),

		txs: inner_list.txs,
	}
}

func (c *celo_list) TotalCostFor(feeCurrency *common.Address) *big.Int {
	if feeCurrency == nil {
		return c.list.totalcost
	}
	if tc, ok := c.totalCost[*feeCurrency]; ok {
		return tc
	}
	return new(big.Int)
}

// TotalCost Returns the total cost for transactions with the same fee currency.
func (c *celo_list) TotalCost(tx *types.Transaction) *big.Int {
	if !txpool.IsFeeCurrencyTx(tx) {
		return c.list.totalcost
	}
	if tc, ok := c.totalCost[*tx.FeeCurrency()]; ok {
		return tc
	}
	return new(big.Int)
}

func (c *celo_list) addTotalCost(tx *types.Transaction) {
	if txpool.IsFeeCurrencyTx(tx) {
		feeCurrency := tx.FeeCurrency()
		if _, ok := c.totalCost[*feeCurrency]; !ok {
			c.totalCost[*feeCurrency] = big.NewInt(0)
		}
		c.totalCost[*feeCurrency].Add(c.totalCost[*feeCurrency], tx.Cost())
	} else {
		c.list.totalcost.Add(c.list.totalcost, tx.Cost())
	}
}

func (c *celo_list) subTotalCost(txs types.Transactions) {
	for _, tx := range txs {
		if txpool.IsFeeCurrencyTx(tx) {
			feeCurrency := tx.FeeCurrency()
			c.totalCost[*feeCurrency].Sub(c.totalCost[*feeCurrency], tx.Cost())
		} else {
			c.list.totalcost.Sub(c.list.totalcost, tx.Cost())
		}
	}
}

// FirstElement Returns the first element from the list, that is, the lowest nonce.
func (c *celo_list) FirstElement() *types.Transaction {
	return c.list.txs.FirstElement()
}

func (c *celo_list) FilterWhitelisted(st *state.StateDB, all *lookup, fcv txpool.FeeCurrencyValidator) {
	removed := c.list.txs.Filter(func(tx *types.Transaction) bool {
		return txpool.IsFeeCurrencyTx(tx) && fcv.IsWhitelisted(st, tx.FeeCurrency())
	})
	c.subTotalCost(removed)
	for _, tx := range removed {
		hash := tx.Hash()
		all.Remove(hash)
	}
}

func balanceMinusL1Cost(st *state.StateDB, l1Cost *big.Int,
	feeCurrency *common.Address, addr common.Address,
	fcv txpool.FeeCurrencyValidator) *big.Int {
	balance := fcv.Balance(st, addr, feeCurrency)
	currencyL1Cost := fcv.ToCurrencyValue(st, l1Cost, feeCurrency)
	return new(big.Int).Sub(balance, currencyL1Cost)
}

func (c *celo_list) FilterBalance(st *state.StateDB, addr common.Address, l1Cost *big.Int,
	gasLimit uint64,
	fcv txpool.FeeCurrencyValidator) (types.Transactions, types.Transactions) {

	balanceNative := balanceMinusL1Cost(st, l1Cost, nil, addr, fcv)
	balances := make(map[common.Address]*big.Int)

	// Filter out all the transactions above the account's funds
	removed := c.list.txs.Filter(func(tx *types.Transaction) bool {
		var feeCurrency *common.Address = nil
		var costLimit *big.Int = nil
		if txpool.IsFeeCurrencyTx(tx) {
			feeCurrency = tx.FeeCurrency()
			if _, ok := balances[*feeCurrency]; !ok {
				balances[*feeCurrency] = balanceMinusL1Cost(st, l1Cost, feeCurrency, addr, fcv)
			}
			costLimit = balances[*feeCurrency]
		} else {
			costLimit = balanceNative
		}
		return tx.Gas() > gasLimit || tx.Cost().Cmp(costLimit) > 0
	})
	if len(removed) == 0 {
		return nil, nil
	}
	var invalids types.Transactions
	// If the list was strict, filter anything above the lowest nonce
	if c.list.strict {
		lowest := uint64(math.MaxUint64)
		for _, tx := range removed {
			if nonce := tx.Nonce(); lowest > nonce {
				lowest = nonce
			}
		}
		invalids = c.list.txs.filter(func(tx *types.Transaction) bool { return tx.Nonce() > lowest })
	}
	// Reset total cost
	c.subTotalCost(removed)
	c.subTotalCost(invalids)
	c.list.txs.reheap()
	return removed, invalids
}

// Forwarded methods

// Get retrieves the current transactions associated with the given nonce.
func (c *celo_list) Get(nonce uint64) *types.Transaction {
	return c.list.txs.Get(nonce)
}

// Contains returns whether the  list contains a transaction
// with the provided nonce.
func (c *celo_list) Contains(nonce uint64) bool {
	return c.list.Contains(nonce)
}

// Add tries to insert a new transaction into the list, returning whether the
// transaction was accepted, and if yes, any previous transaction it replaced.
//
// If the new transaction is accepted into the list, the lists' cost and gas
// thresholds are also potentially updated.
func (c *celo_list) Add(tx *types.Transaction, priceBump uint64, l1CostFn txpool.L1CostFunc) (bool, *types.Transaction) {
	oldNativeTotalCost := big.NewInt(0).Set(c.list.totalcost)
	added, oldTx := c.list.Add(tx, priceBump, l1CostFn)
	if !added {
		return false, nil
	}
	if !txpool.IsFeeCurrencyTx(tx) && oldTx != nil && !txpool.IsFeeCurrencyTx(oldTx) {
		// both the tx and the replacement are native currency, nothing to do
		return true, oldTx
	}
	// undo change in native totalcost
	c.list.totalcost.Set(oldNativeTotalCost)
	// Recalculate
	c.addTotalCost(tx)
	// TODO: Add rollup cost, translated to the feecurrency of the tx

	// Remove replaced tx cost
	if oldTx != nil {
		c.subTotalCost(types.Transactions{oldTx})
	}
	return added, oldTx

}

// Forward removes all transactions from the list with a nonce lower than the
// provided threshold. Every removed transaction is returned for any post-removal
// maintenance.
func (c *celo_list) Forward(threshold uint64) types.Transactions {
	txs := c.list.txs.Forward(threshold)
	c.subTotalCost(txs)
	return txs
}

// Cap places a hard limit on the number of items, returning all transactions
// exceeding that limit.
func (c *celo_list) Cap(threshold int) types.Transactions {
	txs := c.list.txs.Cap(threshold)
	c.subTotalCost(txs)
	return txs
}

// Remove deletes a transaction from the maintained list, returning whether the
// transaction was found, and also returning any transaction invalidated due to
// the deletion (strict mode only).
func (c *celo_list) Remove(tx *types.Transaction) (bool, types.Transactions) {
	return c.list.Remove(tx)
}

// Ready retrieves a sequentially increasing list of transactions starting at the
// provided nonce that is ready for processing. The returned transactions will be
// removed from the list.
//
// Note, all transactions with nonces lower than start will also be returned to
// prevent getting into and invalid state. This is not something that should ever
// happen but better to be self correcting than failing!
func (c *celo_list) Ready(start uint64) types.Transactions {
	return c.list.Ready(start)
}

// Len returns the length of the transaction list.
func (c *celo_list) Len() int {
	return c.list.Len()
}

// Empty returns whether the list of transactions is empty or not.
func (c *celo_list) Empty() bool {
	return c.list.Empty()
}

// Flatten creates a nonce-sorted slice of transactions based on the loosely
// sorted internal representation. The result of the sorting is cached in case
// it's requested again before any modifications are made to the contents.
func (c *celo_list) Flatten() types.Transactions {
	return c.list.Flatten()
}

// LastElement returns the last element of a flattened list, thus, the
// transaction with the highest nonce
func (c *celo_list) LastElement() *types.Transaction {
	return c.list.LastElement()
}
