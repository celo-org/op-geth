package miner

import (
	"math"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
)

var (
	feeCurrency1 = common.BigToAddress(big.NewInt(1))
	feeCurrency2 = common.BigToAddress(big.NewInt(2))
	header       = types.Header{Time: 1111111111111}
)

func HeaderAfter(h types.Header, deltaSeconds int64) *types.Header {
	if h.Time > math.MaxInt64 {
		panic("int64 overflow")
	}
	t := int64(h.Time) + deltaSeconds
	if t < 0 {
		panic("uint64 underflow")
	}
	return &types.Header{Time: uint64(t)}
}

func TestBlocklistEviction(t *testing.T) {
	bl := NewAddressBlocklist()
	bl.Add(feeCurrency1, header)

	// latest header is before eviction time
	assert.True(t, bl.IsBlocked(feeCurrency1, HeaderAfter(header, int64(EvictionTimeoutSeconds)-1)))
	// latest header is after eviction time
	assert.False(t, bl.IsBlocked(feeCurrency1, HeaderAfter(header, int64(EvictionTimeoutSeconds)+1)))

	// check filter allowlist removes the currency from the allowlist
	assert.Equal(t, len(bl.FilterAllowlist(
		common.NewAddressSet(feeCurrency1),
		HeaderAfter(header, int64(EvictionTimeoutSeconds)-1)),
	), 0)

	// permanently delete the currency from the blocklist
	bl.Evict(HeaderAfter(header, int64(EvictionTimeoutSeconds)+1))

	// now the currency is removed from the cache, so the currency is not blocked even in earlier headers
	assert.False(t, bl.IsBlocked(feeCurrency1, HeaderAfter(header, int64(EvictionTimeoutSeconds)-1)))

	// check filter allowlist doesn't change the allowlist
	assert.Equal(t, len(bl.FilterAllowlist(
		common.NewAddressSet(feeCurrency1),
		HeaderAfter(header, int64(EvictionTimeoutSeconds)-1)),
	), 1)
}

func TestBlocklistAddAfterEviction(t *testing.T) {
	bl := NewAddressBlocklist()
	bl.Add(feeCurrency1, header)
	bl.Evict(HeaderAfter(header, int64(EvictionTimeoutSeconds)+1))

	header2 := HeaderAfter(header, 10)
	bl.Add(feeCurrency2, *header2)

	// make sure the feeCurrency2 behaves as expected
	assert.True(t, bl.IsBlocked(feeCurrency2, HeaderAfter(*header2, int64(EvictionTimeoutSeconds)-1)))
	assert.False(t, bl.IsBlocked(feeCurrency2, HeaderAfter(*header2, int64(EvictionTimeoutSeconds)+1)))
}

func TestBlocklistRemove(t *testing.T) {
	bl := NewAddressBlocklist()

	// Check that removing a fee currency from an empty blocklist doesn't panic.
	bl.Remove(feeCurrency1)

	// Check that removal of existing fee currency works.
	bl.Add(feeCurrency1, header)
	bl.Add(feeCurrency2, header)
	bl.Remove(feeCurrency1)

	assert.False(t, bl.IsBlocked(feeCurrency1, HeaderAfter(header, int64(EvictionTimeoutSeconds)-1)))
	assert.True(t, bl.IsBlocked(feeCurrency2, HeaderAfter(header, int64(EvictionTimeoutSeconds)-1)))
}

func TestBlocklistAddAfterRemove(t *testing.T) {
	bl := NewAddressBlocklist()
	bl.Add(feeCurrency1, header)
	bl.Remove(feeCurrency1)
	assert.False(t, bl.IsBlocked(feeCurrency1, HeaderAfter(header, int64(EvictionTimeoutSeconds)-1)))

	header2 := HeaderAfter(header, 10)
	bl.Add(feeCurrency2, *header2)

	// make sure the feeCurrency2 behaves as expected
	assert.True(t, bl.IsBlocked(feeCurrency2, HeaderAfter(*header2, int64(EvictionTimeoutSeconds)-1)))
	assert.False(t, bl.IsBlocked(feeCurrency2, HeaderAfter(*header2, int64(EvictionTimeoutSeconds)+1)))
}

func TestDisableEnableBlocking(t *testing.T) {
	bl := NewAddressBlocklist()
	bl.Add(feeCurrency1, header)
	bl.Add(feeCurrency2, header)

	assert.True(t, bl.BlockingEnabled(feeCurrency1))
	assert.True(t, bl.BlockingEnabled(feeCurrency2))

	allowlist := common.NewAddressSet(feeCurrency1, feeCurrency2)

	assert.Equal(t, common.NewAddressSet(), bl.FilterAllowlist(allowlist, &header))

	bl.DisableBlocking([]common.Address{feeCurrency1})
	assert.False(t, bl.BlockingEnabled(feeCurrency1))
	assert.Equal(t, common.NewAddressSet(feeCurrency1), bl.FilterAllowlist(allowlist, &header))

	bl.DisableBlocking([]common.Address{feeCurrency2})
	assert.False(t, bl.BlockingEnabled(feeCurrency2))
	assert.Equal(t, allowlist, bl.FilterAllowlist(allowlist, &header))

	bl.EnableBlocking([]common.Address{feeCurrency2, feeCurrency1})
	assert.True(t, bl.BlockingEnabled(feeCurrency2))
	assert.True(t, bl.BlockingEnabled(feeCurrency1))
	assert.Equal(t, common.NewAddressSet(), bl.FilterAllowlist(allowlist, &header))
}

func TestBlocklistRetrieval(t *testing.T) {
	bl := NewAddressBlocklist()
	bl.Add(feeCurrency1, header)
	bl.Add(feeCurrency2, header)

	expiryTime := header.Time + EvictionTimeoutSeconds

	allCurrencies := map[common.Address]uint64{
		feeCurrency1: expiryTime,
		feeCurrency2: expiryTime,
	}
	assert.Equal(t, allCurrencies, bl.Blocklist(true))
	assert.Equal(t, allCurrencies, bl.Blocklist(false))

	bl.DisableBlocking([]common.Address{feeCurrency1})
	assert.Equal(t, allCurrencies, bl.Blocklist(true))
	oneCurrency := map[common.Address]uint64{
		feeCurrency2: expiryTime,
	}
	assert.Equal(t, oneCurrency, bl.Blocklist(false))
}

func TestDisabledCurrenciesRetrieval(t *testing.T) {
	bl := NewAddressBlocklist()
	disabledCurrencies := []common.Address{feeCurrency1, feeCurrency2}

	assert.Empty(t, bl.DisabledCurrencies())

	bl.DisableBlocking(disabledCurrencies)
	assert.ElementsMatch(t, disabledCurrencies, bl.DisabledCurrencies())

	bl.EnableBlocking([]common.Address{feeCurrency1})
	assert.ElementsMatch(t, []common.Address{feeCurrency2}, bl.DisabledCurrencies())
}
