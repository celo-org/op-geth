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
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

var (
	ErrDeprecatedTxType = errors.New("deprecated transaction type")
	forks               = []forkID{&cel2{}}
)

// celoSigner acts as an overlay signer that handles celo specific signing
// functionality and hands off to an upstream signer for any other transaction
// types. Unlike the signers in the go-ethereum library, the celoSigner is
// configured with a list of forks that determine it's singing capabilities so
// there should not be a need to create any further signers to handle celo
// specific transaction types.
type celoSigner struct {
	upstreamSigner Signer
	chainID        *big.Int
	activatedForks []forkID
}

// makeCeloSigner creates a new celoSigner that is configured to handle all
// celo forks that are active at the given block time. If there are no active
// celo forks the upstream signer will be returned.
func makeCeloSigner(chainConfig *params.ChainConfig, blockTime uint64, upstreamSigner Signer) Signer {
	s := &celoSigner{
		chainID:        chainConfig.ChainID,
		upstreamSigner: upstreamSigner,
	}
	// Iterate over forks and set the activated forks
	for i, fork := range forks {
		if fork.active(blockTime, chainConfig) {
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

// latestCeloSigner creates a new celoSigner that is configured to handle all
// celo forks for non celo transaction types it will delegate to the given
// upstream signer.
func latestCeloSigner(chainID *big.Int, upstreamSigner Signer) Signer {
	return &celoSigner{
		chainID:        chainID,
		upstreamSigner: upstreamSigner,
		activatedForks: forks,
	}
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

// Hash implements Signer.
func (c *celoSigner) Hash(tx *Transaction) common.Hash {
	if funcs := c.findTxFuncs(tx.Type()); funcs != nil {
		return funcs.hash(tx, c.ChainID())
	}
	return c.upstreamSigner.Hash(tx)
}

// findTxFuncs returns the txFuncs for the given tx type if it is supported by
// one of the active forks. note that this mechanism can be used to deprecate
// support for tx types in future forks, by having forks reutrn
// deprecatedTxFuncs for a tx type.
func (c *celoSigner) findTxFuncs(txType uint8) *txFuncs {
	// iterate in reverse over the activeForks and if any of them have a non nil txFuncs
	// for the tx type then return it.
	for i := len(c.activatedForks) - 1; i >= 0; i-- {
		if funcs := c.activatedForks[i].txFuncs(txType); funcs != nil {
			return funcs
		}
	}
	return nil
}

// ChainID implements Signer.
func (c *celoSigner) ChainID() *big.Int {
	return c.chainID
}

// Equal implements Signer.
func (c *celoSigner) Equal(s Signer) bool {
	// Normally singers just check to see if the chainID and type are equal,
	// because their logic is hardcoded to a specific fork. In our case we need
	// to also know that the two signers have matching latest forks.
	other, ok := s.(*celoSigner)
	return ok && c.ChainID() == other.ChainID() && c.latestFork().equal(other.latestFork())
}

func (c *celoSigner) latestFork() forkID {
	return c.activatedForks[len(c.activatedForks)-1]
}

// forkID represents a fork. It contains functionality to determine if it is
// active for a given block time and chain config and also acts as a container
// for functionality related to transactions enabled in that fork.
type forkID interface {
	// active returns true if the fork is active at the given block time.
	active(blockTime uint64, config *params.ChainConfig) bool
	// equal returns true if the given fork is the same underlying type as this fork.
	equal(forkID) bool
	// txFuncs returns the txFuncs for the given tx type if it is supported by
	// the fork. If a fork deprecates a tx type then this function should
	// return deprecatedTxFuncs for that tx type.
	txFuncs(txType uint8) *txFuncs
}

// Cel2 is the fork marking the transition point from an L1 to an L2. At
// present it provides support for all historical celo tx types.
type cel2 struct{}

func (c *cel2) active(blockTime uint64, config *params.ChainConfig) bool {
	return config.IsCel2(blockTime)
}

func (c *cel2) equal(other forkID) bool {
	_, ok := other.(*cel2)
	return ok
}

func (c *cel2) txFuncs(txType uint8) *txFuncs {
	switch txType {
	case CeloDynamicFeeTxV2Type:
		return celoDynamicFeeTxV2Funcs
	case CeloDenominatedTxType:
		return celoDenominatedTxFuncs
	}
	return nil
}
