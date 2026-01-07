package types

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

var (
	// Historical txs accepted with wrong chain ID before validation fix. See https://github.com/celo-org/op-geth/issues/454
	sepoliaChainIDExceptionHash = common.HexToHash("0x4564b9903cfe18814ffc2696e1ad141d9cc3a549dc4f5726e15f7be2e0ccaa25") // block 12531083
	mainnetChainIDExceptionHash = common.HexToHash("0xd6bdf3261df7e7a4db6bbc486bf091eb62dfd2883e335c31219b6a37d3febca1") // block 53619115

	// deprecatedTxFuncs should be returned by forks that have deprecated support for a tx type.
	deprecatedTxFuncs = &txFuncs{
		signatureValues: func(tx *Transaction, sig []byte, signerChainID *big.Int) (r *big.Int, s *big.Int, v *big.Int, err error) {
			return nil, nil, nil, fmt.Errorf("%w %v", ErrDeprecatedTxType, tx.Type())
		},
		sender: func(tx *Transaction, signerChainID *big.Int) (common.Address, error) {
			if tx.IsCeloLegacy() {
				return common.Address{}, fmt.Errorf("%w %v %v", ErrDeprecatedTxType, tx.Type(), "(celo legacy)")
			}
			return common.Address{}, fmt.Errorf("%w %v", ErrDeprecatedTxType, tx.Type())
		},
	}

	// Although celo allowed unprotected transactions it never supported signing
	// them with signers retrieved by MakeSigner or LatestSigner (if you wanted
	// to make an unprotected transaction you needed to use the HomesteadSigner
	// directly), so signatureValues provides protected values,
	// but sender can accept unprotected transactions. See
	// https://github.com/celo-org/celo-blockchain/pull/1748/files and
	// https://github.com/celo-org/celo-blockchain/issues/1734 and
	// https://github.com/celo-org/celo-proposals/blob/master/CIPs/cip-0050.md
	celoLegacyTxFuncs = &txFuncs{
		signatureValues: func(tx *Transaction, sig []byte, signerChainID *big.Int) (r *big.Int, s *big.Int, v *big.Int, err error) {
			r, s, v = decodeSignature(sig)
			if signerChainID.Sign() != 0 {
				v = big.NewInt(int64(sig[64] + 35))
				signerChainMul := new(big.Int).Mul(signerChainID, big.NewInt(2))
				v.Add(v, signerChainMul)
			}
			return r, s, v, nil
		},
		sender: func(tx *Transaction, signerChainID *big.Int) (common.Address, error) {
			if tx.Protected() {
				if tx.ChainId().Cmp(signerChainID) != 0 {
					return common.Address{}, fmt.Errorf("%w: have %d want %d", ErrInvalidChainId, tx.ChainId(), signerChainID)
				}
				v, r, s := tx.RawSignatureValues()
				signerChainMul := new(big.Int).Mul(signerChainID, big.NewInt(2))
				v = new(big.Int).Sub(v, signerChainMul)
				v.Sub(v, big8)
				return recoverPlain(tx.inner.sigHash(signerChainID), r, s, v, true)
			} else {
				v, r, s := tx.RawSignatureValues()
				return recoverPlain(rlpHash(baseCeloLegacyTxSigningFields(tx)), r, s, v, true)
			}
		},
	}

	accessListTxFuncs = &txFuncs{
		signatureValues: func(tx *Transaction, sig []byte, signerChainID *big.Int) (r *big.Int, s *big.Int, v *big.Int, err error) {
			return NewEIP2930Signer(signerChainID).SignatureValues(tx, sig)
		},
		sender: func(tx *Transaction, signerChainID *big.Int) (common.Address, error) {
			// Historical chain ID bug exceptions - use tx's chain ID for signature recovery
			if isChainIDException(tx.Hash(), signerChainID) {
				return NewEIP2930Signer(tx.ChainId()).Sender(tx)
			}
			return NewEIP2930Signer(signerChainID).Sender(tx)
		},
	}

	dynamicFeeTxFuncs = &txFuncs{
		signatureValues: func(tx *Transaction, sig []byte, signerChainID *big.Int) (r *big.Int, s *big.Int, v *big.Int, err error) {
			return NewLondonSigner(signerChainID).SignatureValues(tx, sig)
		},
		sender: func(tx *Transaction, signerChainID *big.Int) (common.Address, error) {
			return NewLondonSigner(signerChainID).Sender(tx)
		},
	}

	celoDynamicFeeTxFuncs = &txFuncs{
		signatureValues: dynamicTxSigValues,
		sender:          dynamicTxSender,
	}

	// Custom signing functionality for CeloDynamicFeeTxV2 txs.
	celoDynamicFeeTxV2Funcs = &txFuncs{
		signatureValues: dynamicTxSigValues,
		sender:          dynamicTxSender,
	}
)

// txFuncs serves as a container to hold custom signing functionality for a transaction.
type txFuncs struct {
	signatureValues func(tx *Transaction, sig []byte, signerChainID *big.Int) (r *big.Int, s *big.Int, v *big.Int, err error)
	sender          func(tx *Transaction, signerChainID *big.Int) (common.Address, error)
}

// Returns the signature values for CeloDynamicFeeTxV2 transactions.
func dynamicTxSigValues(tx *Transaction, sig []byte, signerChainID *big.Int) (r *big.Int, s *big.Int, v *big.Int, err error) {
	// Check that chain ID of tx matches the signer. We also accept ID zero here,
	// because it indicates that the chain ID was not specified in the tx.
	chainID := tx.inner.chainID()
	if chainID.Sign() != 0 && chainID.Cmp(signerChainID) != 0 {
		return nil, nil, nil, ErrInvalidChainId
	}
	r, s, _ = decodeSignature(sig)
	v = big.NewInt(int64(sig[64]))
	return r, s, v, nil
}

// Returns the sender for CeloDynamicFeeTxV2 transactions.
func dynamicTxSender(tx *Transaction, signerChainID *big.Int) (common.Address, error) {
	if tx.ChainId().Cmp(signerChainID) != 0 {
		return common.Address{}, ErrInvalidChainId
	}
	V, R, S := tx.RawSignatureValues()
	// DynamicFee txs are defined to use 0 and 1 as their recovery
	// id, add 27 to become equivalent to unprotected Homestead signatures.
	V = new(big.Int).Add(V, big.NewInt(27))
	return recoverPlain(tx.inner.sigHash(signerChainID), R, S, V, true)
}

// Extracts the common signing fields for CeloLegacy transactions.
func baseCeloLegacyTxSigningFields(tx *Transaction) []interface{} {
	return []interface{}{
		tx.Nonce(),
		tx.GasPrice(),
		tx.Gas(),
		tx.FeeCurrency(),
		tx.GatewayFeeRecipient(),
		tx.GatewayFee(),
		tx.To(),
		tx.Value(),
		tx.Data(),
	}
}

// isChainIDException returns true if the tx hash is a known historical exception
// AND the signerChainID matches the network where the exception occurred.
func isChainIDException(txHash common.Hash, signerChainID *big.Int) bool {
	chainID := signerChainID.Uint64()
	return (txHash == sepoliaChainIDExceptionHash && chainID == params.CeloSepoliaChainID) ||
		(txHash == mainnetChainIDExceptionHash && chainID == params.CeloMainnetChainID)
}
