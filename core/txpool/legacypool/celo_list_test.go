package legacypool

import (
	"container/heap"
	"math/big"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

// Tests that transactions can be added to strict lists and list contents and
// nonce boundaries are correctly maintained.
func TestStrictCeloListAdd(t *testing.T) {
	// Generate a list of transactions to insert
	key, _ := crypto.GenerateKey()

	txs := make(types.Transactions, 1024)
	for i := 0; i < len(txs); i++ {
		txs[i] = transaction(uint64(i), 0, key)
	}
	// Insert the transactions in a random order
	list := newCeloList(true)
	for _, v := range rand.Perm(len(txs)) {
		list.Add(txs[v], DefaultConfig.PriceBump, nil)
	}
	// Verify internal state
	if len(list.list.txs.items) != len(txs) {
		t.Errorf("transaction count mismatch: have %d, want %d", len(list.txs.items), len(txs))
	}
	for i, tx := range txs {
		if list.list.txs.items[tx.Nonce()] != tx {
			t.Errorf("item %d: transaction mismatch: have %v, want %v", i, list.txs.items[tx.Nonce()], tx)
		}
	}
}

func BenchmarkCeloListAdd(b *testing.B) {
	// Generate a list of transactions to insert
	key, _ := crypto.GenerateKey()

	txs := make(types.Transactions, 100000)
	for i := 0; i < len(txs); i++ {
		txs[i] = transaction(uint64(i), 0, key)
	}
	// Insert the transactions in a random order
	priceLimit := big.NewInt(int64(DefaultConfig.PriceLimit))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		list := newCeloList(true)
		for _, v := range rand.Perm(len(txs)) {
			list.Add(txs[v], DefaultConfig.PriceBump, nil)
			list.list.Filter(priceLimit, DefaultConfig.PriceBump) // TODO: change to actual celo_list benchmark
		}
	}
}

// Priceheap tests

func legacytx(price int) *types.Transaction {
	return types.NewTx(&types.LegacyTx{GasPrice: big.NewInt(int64(price))})
}

type testNominalTxComparator struct{}

func (t *testNominalTxComparator) GasFeeCapCmp(a *types.Transaction, b *types.Transaction) int {
	return a.GasFeeCapCmp(b)
}

func (t *testNominalTxComparator) GasTipCapCmp(a *types.Transaction, b *types.Transaction) int {
	return a.GasTipCapCmp(b)
}

func (t *testNominalTxComparator) EffectiveGasTipCmp(a *types.Transaction, b *types.Transaction, baseFee *big.Int) int {
	return a.EffectiveGasTipCmp(b, baseFee)
}

func newTestPriceHeap() *priceHeap {
	return &priceHeap{
		txComparator: &testNominalTxComparator{},
	}
}

func TestLegacyPushes(t *testing.T) {
	m := newTestPriceHeap()
	heap.Push(m, legacytx(100))
	heap.Push(m, legacytx(50))
	heap.Push(m, legacytx(200))
	heap.Push(m, legacytx(75))
	assert.Equal(t, 4, m.Len())
	v := heap.Pop(m)
	tm, _ := v.(*types.Transaction)
	assert.Equal(t, big.NewInt(50), tm.GasPrice())
	assert.Equal(t, 3, m.Len())
}
