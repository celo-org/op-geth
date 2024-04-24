package legacypool

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/exchange"
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
		balances[curr] = contracts.GetFeeBalance(pool.celoBackend, address, &curr)
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

// compareWithRates compares the effective gas price of two transactions according to the exchange rates and
// the base fees in the transactions currencies.
func compareWithRates(a, b *types.Transaction, ratesAndFees *exchange.RatesAndFees) int {
	if ratesAndFees == nil {
		// During node startup the ratesAndFees might not be yet setup, compare nominally
		feeCapCmp := a.GasFeeCapCmp(b)
		if feeCapCmp != 0 {
			return feeCapCmp
		}
		return a.GasTipCapCmp(b)
	}
	rates := ratesAndFees.Rates
	if ratesAndFees.HasBaseFee() {
		tipA := a.EffectiveGasTipValue(ratesAndFees.GetBaseFeeIn(a.FeeCurrency()))
		tipB := b.EffectiveGasTipValue(ratesAndFees.GetBaseFeeIn(b.FeeCurrency()))
		c, _ := exchange.CompareValue(rates, tipA, a.FeeCurrency(), tipB, b.FeeCurrency())
		return c
	}

	// Compare fee caps if baseFee is not specified or effective tips are equal
	feeA := a.GasFeeCap()
	feeB := b.GasFeeCap()
	c, _ := exchange.CompareValue(rates, feeA, a.FeeCurrency(), feeB, b.FeeCurrency())
	if c != 0 {
		return c
	}

	// Compare tips if effective tips and fee caps are equal
	tipCapA := a.GasTipCap()
	tipCapB := b.GasTipCap()
	c, _ = exchange.CompareValue(rates, tipCapA, a.FeeCurrency(), tipCapB, b.FeeCurrency())
	return c
}
