package legacypool

import (
	"container/heap"
	"math/big"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/common"
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

func celotx(fee int, tip int) *types.Transaction {
	return celotxcurrency(fee, tip, nil)
}

func celotxcurrency(fee int, tip int, currency *common.Address) *types.Transaction {
	return types.NewTx(&types.CeloDynamicFeeTx{
		GasFeeCap:   big.NewInt(int64(fee)),
		GasTipCap:   big.NewInt(int64(tip)),
		FeeCurrency: currency,
	})
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

type testMultiplierTxComparator struct {
	mults map[common.Address]int
}

func testCmpMults(vA *big.Int, mA int, vB *big.Int, mB int) int {
	if mA == 0 {
		mA = 1
	}
	if mB == 0 {
		mB = 1
	}
	rA := int(vA.Uint64()) * mA
	rB := int(vB.Uint64()) * mB
	if (rA - rB) < 0 {
		return -1
	}
	if (rA - rB) > 0 {
		return 1
	}
	return 0
}

func (t *testMultiplierTxComparator) GasFeeCapCmp(a *types.Transaction, b *types.Transaction) int {
	return testCmpMults(a.GasFeeCap(), t.mults[*a.FeeCurrency()], b.GasFeeCap(), t.mults[*b.FeeCurrency()])
}

func (t *testMultiplierTxComparator) GasTipCapCmp(a *types.Transaction, b *types.Transaction) int {
	return testCmpMults(a.GasTipCap(), t.mults[*a.FeeCurrency()], b.GasTipCap(), t.mults[*b.FeeCurrency()])
}

func (t *testMultiplierTxComparator) EffectiveGasTipCmp(a *types.Transaction, b *types.Transaction, baseFee *big.Int) int {
	return testCmpMults(a.EffectiveGasTipValue(baseFee), t.mults[*a.FeeCurrency()], b.EffectiveGasTipValue(baseFee), t.mults[*b.FeeCurrency()])
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

func TestCeloPushes(t *testing.T) {
	m := newTestPriceHeap()
	heap.Push(m, celotx(100, 0))
	heap.Push(m, celotx(50, 3))
	heap.Push(m, celotx(200, 0))
	heap.Push(m, celotx(75, 0))
	assert.Equal(t, 4, m.Len())
	v := heap.Pop(m)
	tm, _ := v.(*types.Transaction)
	assert.Equal(t, big.NewInt(50), tm.GasFeeCap())
	assert.Equal(t, big.NewInt(3), tm.GasTipCap())
	assert.Equal(t, 3, m.Len())
}

func TestCurrencyAdds(t *testing.T) {
	c1 := common.BigToAddress(big.NewInt(2))
	c2 := common.BigToAddress(big.NewInt(3))
	tmc := &testMultiplierTxComparator{
		mults: map[common.Address]int{
			c1: 2,
			c2: 3,
		}}
	m := newTestPriceHeap()
	m.txComparator = tmc
	heap.Push(m, celotxcurrency(100, 0, &c1)) // 200
	heap.Push(m, celotxcurrency(250, 0, &c2)) // 750
	heap.Push(m, celotxcurrency(50, 0, &c1))  // 100
	heap.Push(m, celotxcurrency(75, 0, &c2))  // 225
	heap.Push(m, celotxcurrency(200, 0, &c1)) // 400

	assert.Equal(t, 5, m.Len())

	tm := heap.Pop(m).(*types.Transaction)
	assert.Equal(t, big.NewInt(50), tm.GasPrice())
	assert.Equal(t, 4, m.Len())

	tm2 := heap.Pop(m).(*types.Transaction)
	assert.Equal(t, big.NewInt(100), tm2.GasPrice())
}
