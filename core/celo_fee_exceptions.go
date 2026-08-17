package core

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

var (
	// Historical tx accepted by celo-reth without the fee currency max fee check. See https://github.com/celo-org/celo-kona/issues/292
	mainnetMaxFeeExceptionHash = common.HexToHash("0xe9e0248cf5b02ce016b195690fbb168f08eeeff7b9b353a52ecf98f4229ba834") // block 75046581
)

// isFeeCurrencyMaxFeeException returns true if the tx hash is a known historical
// exception to the `balance >= gasLimit * maxFeePerGas` rule that canPayFee applies to
// fee currency balances AND the chainID matches the network where the exception
// occurred.
//
// celo-reth never implemented that rule for CIP-64 transactions: it requires the fee
// currency balance to cover only `gasLimit * effectiveGasPrice`, which is what the
// debitGasFees call charges. A transaction whose sender holds a balance between the two
// thresholds is therefore valid for celo-reth and invalid for op-geth. The transactions
// above were produced under the lenient rule and are canonical, so op-geth has to accept
// them to stay on the chain.
//
// Exempting a transaction only skips the max fee check; the fee currency debit itself
// still has to succeed, which is exactly the constraint celo-reth applies. This list
// mirrors the one in celo-reth and must stay in sync with it.
func isFeeCurrencyMaxFeeException(txHash common.Hash, chainID *big.Int) bool {
	if chainID == nil || !chainID.IsUint64() {
		return false
	}
	return txHash == mainnetMaxFeeExceptionHash && chainID.Uint64() == params.CeloMainnetChainID
}
