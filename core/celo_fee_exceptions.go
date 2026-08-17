package core

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

// feeCurrencyMaxFeeException pins a single transaction that is exempt from the fee
// currency max fee check. The (chainID, blockNumber) pair is part of the key so that the
// exemption cannot apply anywhere but the canonical block it was found in.
type feeCurrencyMaxFeeException struct {
	txHash      common.Hash
	chainID     uint64
	blockNumber uint64
}

// feeCurrencyMaxFeeExceptions lists the CIP-64 transactions that were included in a
// canonical block even though the sender's fee currency balance did not cover
// `gasLimit * maxFeePerGas`. See https://github.com/celo-org/celo-kona/issues/292
//
// celo-reth never implemented that rule for CIP-64 transactions: it required the fee
// currency balance to cover only `gasLimit * effectiveGasPrice`, which is what the
// debitGasFees call charges. A transaction whose sender holds a balance between the two
// thresholds is therefore valid for celo-reth and invalid for op-geth. The transactions
// below were produced under the lenient rule and are canonical, so op-geth has to accept
// them to stay on the chain.
//
// This list mirrors CIP64_MAX_FEE_EXCEPTIONS in celo-reth and must stay in sync with it,
// including the (txHash, chainID, blockNumber) key: a client that exempts a transaction
// the other one rejects reintroduces the divergence this list exists to paper over.
//
// celo-reth now enforces the rule too, so the list is closed for any block it has
// already processed under the check. It may still gain entries for blocks between
// 75046581 and that deployment.
var feeCurrencyMaxFeeExceptions = []feeCurrencyMaxFeeException{
	// Celo Mainnet block 75046581, tx index 29 - 200000 gas * 29097136819 max fee
	// = 5819427363800000 required in 0x0E2A3e05bc9A16F5292A6170456A710cb89C6f72, while
	// sender 0x1De6939e8A03DF7bDc970A951B67628d2D138eD9 held 5295000000000000. The
	// 12161345100 effective price made the actual debit 2432269020000000, affordable.
	{
		txHash:      common.HexToHash("0xe9e0248cf5b02ce016b195690fbb168f08eeeff7b9b353a52ecf98f4229ba834"),
		chainID:     params.CeloMainnetChainID,
		blockNumber: 75046581,
	},
}

// isFeeCurrencyMaxFeeException returns true if the transaction is a known historical
// exception to the `balance >= gasLimit * maxFeePerGas` rule that canPayFee applies to
// fee currency balances, and is being applied on the network and at the block height
// where the exception occurred.
//
// Exempting a transaction only skips the max fee check; the fee currency debit itself
// still has to succeed, which is exactly the constraint celo-reth applies.
func isFeeCurrencyMaxFeeException(txHash common.Hash, chainID, blockNumber *big.Int) bool {
	if chainID == nil || !chainID.IsUint64() || blockNumber == nil || !blockNumber.IsUint64() {
		return false
	}
	chain, block := chainID.Uint64(), blockNumber.Uint64()
	for _, e := range feeCurrencyMaxFeeExceptions {
		// Compare the cheap fields first, so that ordinary transactions never reach the
		// hash comparison.
		if e.chainID == chain && e.blockNumber == block && e.txHash == txHash {
			return true
		}
	}
	return false
}
