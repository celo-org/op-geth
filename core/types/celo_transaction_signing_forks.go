package types

import "github.com/ethereum/go-ethereum/params"

var (
	forks = []fork{&celoLegacy{}, &cel2{}}
)

// fork contains functionality to determine if it is active for a given block
// time and chain config. It also acts as a container for functionality related
// to transactions enabled or deprecated in that fork.
type fork interface {
	// active returns true if the fork is active at the given block time.
	active(blockTime uint64, config *params.ChainConfig) bool
	// equal returns true if the given fork is the same underlying type as this fork.
	equal(fork) bool
	// txFuncs returns the txFuncs for the given tx if it is supported by the
	// fork. If a fork deprecates a tx type then this function should return
	// deprecatedTxFuncs for that tx type.
	txFuncs(tx *Transaction) *txFuncs
}

// Cel2 is the fork marking the transition point from an L1 to an L2.
// It deprecates CeloDynamicFeeTxType and LegacyTxTypes with CeloLegacy set to true.
type cel2 struct{}

func (c *cel2) active(blockTime uint64, config *params.ChainConfig) bool {
	return config.IsCel2(blockTime)
}

func (c *cel2) equal(other fork) bool {
	_, ok := other.(*cel2)
	return ok
}

func (c *cel2) txFuncs(tx *Transaction) *txFuncs {
	t := tx.Type()
	switch {
	case t == LegacyTxType && tx.CeloLegacy():
		return deprecatedTxFuncs
	case t == CeloDenominatedTxType:
		return deprecatedTxFuncs
	}
	return nil
}

// celoLegacy isn't actually a fork, but a placeholder for all historical celo
// related forks occurring on the celo L1. We don't need to construct the full
// signer chain from the celo legacy project because we won't support
// historical transaction execution, so we just need to be able to derive the
// senders for historical transactions and since we assume that the historical
// data is correct we just need one blanket signer that can cover all legacy
// celo transactions, before the L2 transition point.
type celoLegacy struct{}

func (c *celoLegacy) active(blockTime uint64, config *params.ChainConfig) bool {
	return config.IsCel2(blockTime)
}

func (c *celoLegacy) equal(other fork) bool {
	_, ok := other.(*cel2)
	return ok
}

func (c *celoLegacy) txFuncs(tx *Transaction) *txFuncs {
	t := tx.Type()
	switch {
	case t == uint8(LegacyTxType) && tx.CeloLegacy():
		if tx.Protected() {
			return celoLegacyProtectedTxFuncs
		}
		return celoLegacyUnprotectedTxFuncs
	case t == CeloDynamicFeeTxV2Type:
		return celoDynamicFeeTxV2Funcs
	case t == CeloDenominatedTxType:
		return celoDenominatedTxFuncs
	}
	return nil
}
