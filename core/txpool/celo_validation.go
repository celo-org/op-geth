package txpool

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/exchange"
	"github.com/ethereum/go-ethereum/consensus/misc/eip1559"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
)

var (
	// ErrGasFeeCapBelowMinBaseFee is returned if the gas fee cap is below the minimum base fee
	ErrGasFeeCapBelowMinBaseFee = errors.New("gas fee cap is below the minimum base fee")
)

// AcceptSet is a set of accepted transaction types for a transaction subpool.
type AcceptSet = map[uint8]struct{}

// CeloValidationOptions define certain differences between transaction validation
// across the different pools without having to duplicate those checks.
// In comparison to the standard ValidationOptions, the Accept field has been
// changed to allow to test for CeloDynamicFeeTx types.
type CeloValidationOptions struct {
	Config *params.ChainConfig // Chain configuration to selectively validate based on current fork rules

	AcceptSet    AcceptSet // Set of transaction types that should be accepted for the calling pool
	MaxSize      uint64    // Maximum size of a transaction that the caller can meaningfully handle
	MaxBlobCount int       // Maximum number of blobs allowed per transaction
	MinTip       *big.Int  // Minimum gas tip needed to allow a transaction into the caller pool

	EffectiveGasCeil uint64 // if non-zero, a gas ceiling to enforce independent of the header's gaslimit value
	MaxTxGasLimit    uint64 // Maximum gas limit allowed per individual transaction
}

// NewAcceptSet creates a new AcceptSet with the types provided.
func NewAcceptSet(types ...uint8) AcceptSet {
	m := make(AcceptSet, len(types))
	for _, t := range types {
		m[t] = struct{}{}
	}
	return m
}

// Accepts returns true iff txType is accepted by this CeloValidationOptions.
func (cvo *CeloValidationOptions) Accepts(txType uint8) bool {
	_, ok := cvo.AcceptSet[txType]
	return ok
}

// CeloValidateTransaction is a helper method to check whether a transaction is valid
// according to the consensus rules, but does not check state-dependent validation
// (balance, nonce, etc).
//
// This check is public to allow different transaction pools to check the basic
// rules without duplicating code and running the risk of missed updates.
func CeloValidateTransaction(tx *types.Transaction, head *types.Header,
	signer types.Signer, opts *CeloValidationOptions, currencyCtx common.FeeCurrencyContext) error {
	if err := ValidateTransaction(tx, head, signer, opts, currencyCtx); err != nil {
		return err
	}

	if !common.IsCurrencyAllowed(currencyCtx.ExchangeRates, tx.FeeCurrency()) {
		return exchange.ErrUnregisteredFeeCurrency
	}

	// Determine the base fee floor based on the fork
	var baseFeeFloorNative *big.Int
	if opts.Config.IsJovian(head.Time) {
		// Post-Jovian: use OP minBaseFee from header
		_, _, minBaseFee := eip1559.DecodeOptimismExtraData(opts.Config, head.Time, head.Extra)
		if minBaseFee != nil && *minBaseFee > 0 {
			baseFeeFloorNative = new(big.Int).SetUint64(*minBaseFee)
		}
	} else if opts.Config.Celo != nil {
		// Pre-Jovian: use Celo config floor
		baseFeeFloorNative = new(big.Int).SetUint64(opts.Config.Celo.EIP1559BaseFeeFloor)
	}

	if baseFeeFloorNative != nil {
		baseFeeFloor, err := exchange.ConvertCeloToCurrency(
			currencyCtx.ExchangeRates,
			tx.FeeCurrency(),
			baseFeeFloorNative,
		)
		if err != nil {
			return err
		}
		// Check that the fee cap exceeds the base fee floor
		if baseFeeFloor.Cmp(tx.GasFeeCap()) == 1 {
			return ErrGasFeeCapBelowMinBaseFee
		}

		// Make sure that the effective gas tip at the base fee floor is at least the
		// requested min-tip.
		// The min-tip for local transactions is set to 0, we can skip checking here.
		if opts.MinTip != nil && opts.MinTip.Cmp(new(big.Int)) > 0 {
			// If not, this would never be included, so we can reject early.
			minTip, err := exchange.ConvertCeloToCurrency(currencyCtx.ExchangeRates, tx.FeeCurrency(), opts.MinTip)
			if err != nil {
				return err
			}
			if tx.EffectiveGasTipIntCmp(uint256.MustFromBig(minTip), uint256.MustFromBig(baseFeeFloor)) < 0 {
				return ErrUnderpriced
			}
		}
	}
	return nil
}
