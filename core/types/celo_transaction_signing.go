// Copyright 2024 The Celo Authors
// This file is part of the celo library.
//
// The celo library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The celo library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the celo library. If not, see <http://www.gnu.org/licenses/>.

package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

// type cel2Signer struct{ londonSigner }

// NewCel2Signer returns a signer that accepts
// - CIP-64 celo dynamic fee transaction (v2)
// - EIP-4844 blob transactions
// - EIP-1559 dynamic fee transactions
// - EIP-2930 access list transactions,
// - EIP-155 replay protected transactions, and
// - legacy Homestead transactions.
// func NewCel2Signer(chainId *big.Int) Signer {
// 	return cel2Signer{londonSigner{eip2930Signer{NewEIP155Signer(chainId)}}}
// }

// func (s cel2Signer) Sender(tx *Transaction) (common.Address, error) {
// 	if tx.Type() != CeloDynamicFeeTxV2Type && tx.Type() != CeloDenominatedTxType {
// 		return s.londonSigner.Sender(tx)
// 	}
// 	V, R, S := tx.RawSignatureValues()
// 	// DynamicFee txs are defined to use 0 and 1 as their recovery
// 	// id, add 27 to become equivalent to unprotected Homestead signatures.
// 	V = new(big.Int).Add(V, big.NewInt(27))
// 	if tx.ChainId().Cmp(s.chainId) != 0 {
// 		return common.Address{}, ErrInvalidChainId
// 	}
// 	return recoverPlain(s.Hash(tx), R, S, V, true)
// }

// func (s cel2Signer) Equal(s2 Signer) bool {
// 	x, ok := s2.(cel2Signer)
// 	return ok && x.chainId.Cmp(s.chainId) == 0
// }

// func (s cel2Signer) SignatureValues(tx *Transaction, sig []byte) (R, S, V *big.Int, err error) {
// 	if tx.Type() != CeloDynamicFeeTxV2Type && tx.Type() != CeloDenominatedTxType {
// 		return s.londonSigner.SignatureValues(tx, sig)
// 	}

// 	// Check that chain ID of tx matches the signer. We also accept ID zero here,
// 	// because it indicates that the chain ID was not specified in the tx.
// 	chainID := tx.inner.chainID()
// 	if chainID.Sign() != 0 && chainID.Cmp(s.chainId) != 0 {
// 		return nil, nil, nil, ErrInvalidChainId
// 	}
// 	R, S, _ = decodeSignature(sig)
// 	V = big.NewInt(int64(sig[64]))
// 	return R, S, V, nil
// }

// // Hash returns the hash to be signed by the sender.
// // It does not uniquely identify the transaction.
// func (s cel2Signer) Hash(tx *Transaction) common.Hash {
// 	if tx.Type() == CeloDynamicFeeTxV2Type {
// 		return prefixedRlpHash(
// 			tx.Type(),
// 			[]interface{}{
// 				s.chainId,
// 				tx.Nonce(),
// 				tx.GasTipCap(),
// 				tx.GasFeeCap(),
// 				tx.Gas(),
// 				tx.To(),
// 				tx.Value(),
// 				tx.Data(),
// 				tx.AccessList(),
// 				tx.FeeCurrency(),
// 			})
// 	}
// 	if tx.Type() == CeloDenominatedTxType {
// 		return prefixedRlpHash(
// 			tx.Type(),
// 			[]interface{}{
// 				s.chainId,
// 				tx.Nonce(),
// 				tx.GasTipCap(),
// 				tx.GasFeeCap(),
// 				tx.Gas(),
// 				tx.To(),
// 				tx.Value(),
// 				tx.Data(),
// 				tx.AccessList(),
// 				tx.FeeCurrency(),
// 				tx.MaxFeeInFeeCurrency(),
// 			})
// 	}
// 	return s.londonSigner.Hash(tx)
// }

// func makeCeloSigner(config *params.ChainConfig, blockNumber *big.Int, blockTime uint64, upstreamSigner Signer) Signer {
// 	signer := upstreamSigner
// 	switch {
// 	case config.IsCel2(blockTime):
// 		signer = NewCel2Signer(config.ChainID)
// 	}
// 	return signer
// }

type celoSigner struct {
	upstreamSigner Signer
	chainID        *big.Int
	activatedForks []forkID
}

// ChainID implements Signer.
func (c *celoSigner) ChainID() *big.Int {
	return c.chainID
}

// Equal implements Signer.
func (c *celoSigner) Equal(s Signer) bool {
	// Normally singers just check to see if the chainID and type are equal, because their logic is completely
	// hardcoded to a specific fork. In our case we need to also know that the two signers have matching latest forks.
	other, ok := s.(*celoSigner)
	return ok && c.ChainID() == other.ChainID() && c.latestFork().Equal(other.latestFork())
}

// Hash implements Signer.
func (c *celoSigner) Hash(tx *Transaction) common.Hash {
	if funcs := c.findTxFuncs(tx.Type()); funcs != nil {
		return funcs.hash(tx, c.ChainID())
	}
	return c.upstreamSigner.Hash(tx)
}

// Sender implements Signer.
func (c *celoSigner) Sender(tx *Transaction) (common.Address, error) {
	if funcs := c.findTxFuncs(tx.Type()); funcs != nil {
		return funcs.sender(tx, funcs.hash, c.ChainID())
	}
	return c.upstreamSigner.Sender(tx)
}

// SignatureValues implements Signer.
func (c *celoSigner) SignatureValues(tx *Transaction, sig []byte) (r *big.Int, s *big.Int, v *big.Int, err error) {
	if funcs := c.findTxFuncs(tx.Type()); funcs != nil {
		return funcs.signatureValues(tx, sig, c.ChainID())
	}
	return c.upstreamSigner.SignatureValues(tx, sig)
}

func (c *celoSigner) latestFork() forkID {
	return c.activatedForks[len(c.activatedForks)-1]
}

func (c *celoSigner) findTxFuncs(txType uint8) *txFuncs {
	// iterate in reverse over the activeForks and if any of them have a non nil txFuncs
	// for the tx type then return it.
	for i := len(c.activatedForks) - 1; i >= 0; i-- {
		if funcs := c.activatedForks[i].TxFuncs(txType); funcs != nil {
			return funcs
		}
	}
	return nil
}

func makeCeloSigner(chainConfig *params.ChainConfig, blockTime uint64, upstreamSigner Signer) Signer {
	s := &celoSigner{
		chainID:        chainConfig.ChainID,
		upstreamSigner: upstreamSigner,
	}
	// Iterate over forks and set the activated forks
	for i, fork := range forks {
		if fork.Active(blockTime, chainConfig) {
			s.activatedForks = forks[:i+1]
			break
		}
	}
	// If there are no active celo forks, return the upstream signer
	if s.activatedForks == nil {
		return upstreamSigner
	}
	return s
}

func latestCeloSigner(chainID *big.Int, upstreamSigner Signer) Signer {
	return &celoSigner{
		chainID:        chainID,
		upstreamSigner: upstreamSigner,
		activatedForks: forks,
	}
}

type forkID interface {
	Active(blockTime uint64, config *params.ChainConfig) bool
	Equal(forkID) bool
	TxFuncs(txType uint8) *txFuncs
}

type cel2 struct{}

func (c *cel2) Active(blockTime uint64, config *params.ChainConfig) bool {
	return config.IsCel2(blockTime)
}
func (c *cel2) Equal(other forkID) bool {
	_, ok := other.(*cel2)
	return ok
}
func (c *cel2) TxFuncs(txType uint8) *txFuncs {
	switch txType {
	case CeloDynamicFeeTxV2Type:
		return CeloDynamicFeeTxV2Funcs
	case CeloDenominatedTxType:
		return CeloDenominatedTxFuncs
	}
	return nil
}

// Returns the signature values for CeloDynamicFeeTxV2 and CeloDenominatedTx transactions.
func dynamicAndDenominatedTxSigValues(tx *Transaction, sig []byte, signerChainID *big.Int) (r *big.Int, s *big.Int, v *big.Int, err error) {
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

func dynamicAndDenominatedTxSender(tx *Transaction, hashFunc func(tx *Transaction, chainID *big.Int) common.Hash, signerChainID *big.Int) (common.Address, error) {
	V, R, S := tx.RawSignatureValues()
	// DynamicFee txs are defined to use 0 and 1 as their recovery
	// id, add 27 to become equivalent to unprotected Homestead signatures.
	V = new(big.Int).Add(V, big.NewInt(27))
	if tx.ChainId().Cmp(signerChainID) != 0 {
		return common.Address{}, ErrInvalidChainId
	}
	return recoverPlain(hashFunc(tx, signerChainID), R, S, V, true)
}

type txFuncs struct {
	hash            func(tx *Transaction, chainID *big.Int) common.Hash
	signatureValues func(tx *Transaction, sig []byte, signerChainID *big.Int) (r *big.Int, s *big.Int, v *big.Int, err error)
	sender          func(tx *Transaction, hashFunc func(tx *Transaction, chainID *big.Int) common.Hash, signerChainID *big.Int) (common.Address, error)
}

var (
	forks       = []forkID{&cel2{}}
	celoTxTypes = []uint8{CeloDynamicFeeTxType, CeloDynamicFeeTxV2Type, CeloDenominatedTxType}

	CeloDynamicFeeTxV2Funcs = &txFuncs{
		hash: func(tx *Transaction, chainID *big.Int) common.Hash {
			return prefixedRlpHash(
				tx.Type(),
				[]interface{}{
					chainID,
					tx.Nonce(),
					tx.GasTipCap(),
					tx.GasFeeCap(),
					tx.Gas(),
					tx.To(),
					tx.Value(),
					tx.Data(),
					tx.AccessList(),
					tx.FeeCurrency(),
				})
		},
		signatureValues: dynamicAndDenominatedTxSigValues,
		sender:          dynamicAndDenominatedTxSender,
	}
	CeloDenominatedTxFuncs = &txFuncs{
		hash: func(tx *Transaction, chainID *big.Int) common.Hash {
			return prefixedRlpHash(
				tx.Type(),
				[]interface{}{
					chainID,
					tx.Nonce(),
					tx.GasTipCap(),
					tx.GasFeeCap(),
					tx.Gas(),
					tx.To(),
					tx.Value(),
					tx.Data(),
					tx.AccessList(),
					tx.FeeCurrency(),
					tx.MaxFeeInFeeCurrency(),
				})
		},
		signatureValues: dynamicAndDenominatedTxSigValues,
		sender:          dynamicAndDenominatedTxSender,
	}
)

// CeloDynamicFeeTxFuncs = &txFuncs{
// 	hash: func(tx *Transaction, chainID *big.Int) common.Hash {
// 		inner := tx.inner.(*CeloDynamicFeeTx)
// 		return prefixedRlpHash(
// 			tx.Type(),
// 			[]interface{}{
// 				s.chainId,
// 				tx.Nonce(),
// 				tx.GasTipCap(),
// 				tx.GasFeeCap(),
// 				tx.Gas(),
// 				tx.FeeCurrency(),
// 				inner.GatewayFeeRecipient,
// 				inner.GatewayFee,
// 				tx.To(),
// 				tx.Value(),
// 				tx.Data(),
// 				tx.AccessList(),
// 			})
// 	},
// }
