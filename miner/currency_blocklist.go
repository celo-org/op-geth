package miner

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const hours uint64 = 60 * 60

const EvictionTimeoutSeconds uint64 = 2 * hours

type AddressBlocklist struct {
	mux        *sync.RWMutex
	currencies map[common.Address]*types.Header
	// fee-currencies blocked at headers with an older timestamp
	// will get evicted when evict() is called
	oldestHeader *types.Header

	// disabledCurrencies is the set of currencies for which the blocklist
	// functionality has been manually disabled.
	disabledCurrencies map[common.Address]struct{}
}

func NewAddressBlocklist() *AddressBlocklist {
	bl := &AddressBlocklist{
		mux:                &sync.RWMutex{},
		currencies:         map[common.Address]*types.Header{},
		oldestHeader:       nil,
		disabledCurrencies: make(map[common.Address]struct{}),
	}
	return bl
}

// Blocklist returns the current blocklist, a set of fee currency
// addresses mapped to their expiry Unix timestamp. Note that the parameter
// 'includeDisabled' determines if the map should contain fee currencies for
// which blocking has been manually disabled.
func (b *AddressBlocklist) Blocklist(includeDisabled bool) map[common.Address]uint64 {
	b.mux.RLock()
	defer b.mux.RUnlock()
	result := make(map[common.Address]uint64, len(b.currencies))
	for currency, addedHeader := range b.currencies {
		if _, disabled := b.disabledCurrencies[currency]; disabled && !includeDisabled {
			continue
		}
		result[currency] = addedHeader.Time + EvictionTimeoutSeconds
	}
	return result
}

// DisableBlocking disables blocking on the given currencies, for currencies that
// are already disabled this is a no-op.
func (b *AddressBlocklist) DisableBlocking(currencies []common.Address) {
	b.mux.Lock()
	defer b.mux.Unlock()
	for _, currency := range currencies {
		b.disabledCurrencies[currency] = struct{}{}
	}
}

// EnableBlocking enables blocking on the given currencies, for currencies that
// are already enabled this is a no-op.
func (b *AddressBlocklist) EnableBlocking(currencies []common.Address) {
	b.mux.Lock()
	defer b.mux.Unlock()
	for _, currency := range currencies {
		delete(b.disabledCurrencies, currency)
	}
}

// DisabledCurrencies returns the currencies for which blocking is currently
// manually disabled.
func (b *AddressBlocklist) DisabledCurrencies() []common.Address {
	b.mux.RLock()
	defer b.mux.RUnlock()
	result := make([]common.Address, 0, len(b.disabledCurrencies))
	for currency := range b.disabledCurrencies {
		result = append(result, currency)
	}
	return result
}

// BlockingEnabled returns whether blocking is enabled for the given currency.
func (b *AddressBlocklist) BlockingEnabled(currency common.Address) bool {
	b.mux.RLock()
	defer b.mux.RUnlock()
	_, disabled := b.disabledCurrencies[currency]
	return !disabled
}

// FilterAllowlist returns allowlist with any blocked and not disabled
// currencies removed. It accepts the latest header so that it may provide a
// view of the blocklist consistent with that header.
func (b *AddressBlocklist) FilterAllowlist(allowlist common.AddressSet, latest *types.Header) common.AddressSet {
	b.mux.RLock()
	defer b.mux.RUnlock()

	filtered := common.AddressSet{}
	for a := range allowlist {
		_, disabled := b.disabledCurrencies[a]
		if !b.isBlocked(a, latest) || disabled {
			filtered[a] = struct{}{}
		}
	}
	return filtered
}

func (b *AddressBlocklist) IsBlocked(currency common.Address, latest *types.Header) bool {
	b.mux.RLock()
	defer b.mux.RUnlock()

	return b.isBlocked(currency, latest)
}

func (b *AddressBlocklist) Remove(currency common.Address) bool {
	b.mux.Lock()
	defer b.mux.Unlock()

	h, ok := b.currencies[currency]
	if !ok {
		return false
	}
	delete(b.currencies, currency)
	if b.oldestHeader.Time >= h.Time {
		b.resetOldestHeader()
	}
	return ok
}

func (b *AddressBlocklist) Add(currency common.Address, head types.Header) bool {
	b.mux.Lock()
	defer b.mux.Unlock()

	_, existed := b.currencies[currency]
	if b.oldestHeader == nil || b.oldestHeader.Time > head.Time {
		b.oldestHeader = &head
	}
	b.currencies[currency] = &head
	return !existed
}

func (b *AddressBlocklist) Evict(latest *types.Header) []common.Address {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.evict(latest)
}

func (b *AddressBlocklist) resetOldestHeader() {
	if len(b.currencies) == 0 {
		b.oldestHeader = nil
		return
	}
	for _, v := range b.currencies {
		if b.oldestHeader == nil {
			b.oldestHeader = v
			continue
		}
		if v.Time < b.oldestHeader.Time {
			b.oldestHeader = v
		}
	}
}

func (b *AddressBlocklist) evict(latest *types.Header) []common.Address {
	evicted := []common.Address{}
	if latest == nil {
		return evicted
	}

	if b.oldestHeader == nil || !b.headerEvicted(b.oldestHeader, latest) {
		// nothing set yet
		return evicted
	}
	for feeCurrencyAddress, addedHeader := range b.currencies {
		if b.headerEvicted(addedHeader, latest) {
			delete(b.currencies, feeCurrencyAddress)
			evicted = append(evicted, feeCurrencyAddress)
		}
	}
	b.resetOldestHeader()
	return evicted
}

func (b *AddressBlocklist) headerEvicted(h, latest *types.Header) bool {
	return h.Time+EvictionTimeoutSeconds < latest.Time
}

func (b *AddressBlocklist) isBlocked(currency common.Address, latest *types.Header) bool {
	h, exists := b.currencies[currency]
	if !exists {
		return false
	}
	if latest == nil {
		// if no latest block provided to check eviction,
		// assume the currency is blocked
		return true
	}
	return !b.headerEvicted(h, latest)
}
