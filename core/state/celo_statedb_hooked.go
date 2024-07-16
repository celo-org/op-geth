// Copyright 2025 The Celo Authors
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

package state

import "github.com/ethereum/go-ethereum/core/tracing"

func (s *hookedStateDB) Hooks() *tracing.Hooks {
	return s.hooks
}

// SetHooks provides a way to dynamically modify the Hooks,
// since Hooks must be disabled during calls to `CreditFees` and `DebitFees`
// Ref: contracts/fee_currencies.go
func (s *hookedStateDB) SetHooks(hooks *tracing.Hooks) {
	if hooks != nil {
		s.hooks = hooks
	} else {
		s.hooks = new(tracing.Hooks)
	}
}
