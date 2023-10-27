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
	list *list
}

func newCeloList(strict bool) *celo_list {
	return &celo_list{
		list: newList(strict),
	}
}

func (c *celo_list) FilterWhitelisted(blockNumber *big.Int, all *lookup, fcv txpool.FeeCurrencyValidator) {
	removed := c.list.txs.Filter(func(tx *types.Transaction) bool {
		return txpool.IsFeeCurrencyTx(tx) && fcv.IsWhitelisted(tx.FeeCurrency(), blockNumber)
	})
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
		if txpool.IsFeeCurrencyTx(tx) && tx.FeeCurrency() != nil {
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
	c.list.subTotalCost(removed)
	c.list.subTotalCost(invalids)
	c.list.txs.reheap()
	return removed, invalids
}

// Forwarded methods

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
	return c.list.Add(tx, priceBump, l1CostFn)
}

// Forward removes all transactions from the list with a nonce lower than the
// provided threshold. Every removed transaction is returned for any post-removal
// maintenance.
func (c *celo_list) Forward(threshold uint64) types.Transactions {
	return c.list.Forward(threshold)
}

// Filter removes all transactions from the list with a cost or gas limit higher
// than the provided thresholds. Every removed transaction is returned for any
// post-removal maintenance. Strict-mode invalidated transactions are also
// returned.
//
// This method uses the cached costcap and gascap to quickly decide if there's even
// a point in calculating all the costs or if the balance covers all. If the threshold
// is lower than the costgas cap, the caps will be reset to a new high after removing
// the newly invalidated transactions.
func (c *celo_list) Filter(costLimit *big.Int, gasLimit uint64) (types.Transactions, types.Transactions) {
	// If all transactions are below the threshold, short circuit
	if c.list.costcap.Cmp(costLimit) <= 0 && c.list.gascap <= gasLimit {
		return nil, nil
	}
	c.list.costcap = new(big.Int).Set(costLimit) // Lower the caps to the thresholds
	c.list.gascap = gasLimit

	// Filter out all the transactions above the account's funds
	removed := c.list.txs.Filter(func(tx *types.Transaction) bool {
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
	c.list.subTotalCost(removed)
	c.list.subTotalCost(invalids)
	c.list.txs.reheap()
	return removed, invalids
}

// Cap places a hard limit on the number of items, returning all transactions
// exceeding that limit.
func (c *celo_list) Cap(threshold int) types.Transactions {

	txs := c.list.txs.Cap(threshold)
	c.list.subTotalCost(txs)
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
