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
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/exchange"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionEffectiveGasTipInCelo(t *testing.T) {
	t.Parallel()

	usdToken := common.HexToAddress("0x765de816845861e75a25fca122bb6898b8b1282a")
	eurToken := common.HexToAddress("0xd8763cba276a3738e6de85b4b3bf5fded6d6ca73")
	exchangeRates := common.ExchangeRates{
		usdToken: big.NewRat(2, 1), // 1 Celo ≒ 2 USD
	}

	getGasTip := func(t *testing.T, tx *Transaction, baseFee *big.Int) *big.Int {
		t.Helper()

		gasTipInCelo, err := tx.EffectiveGasTipInCelo(baseFee, exchangeRates)
		require.NoError(t, err)

		return gasTipInCelo
	}

	// Normal Tx
	t.Run("tx should return the difference between GasFeeCap and BaseFee when tx type is not CeloDynamicFeeTxV2", func(t *testing.T) {
		gasTip := getGasTip(t, NewTx(&DynamicFeeTx{
			GasFeeCap: big.NewInt(9e8),
			GasTipCap: big.NewInt(5e8),
		}), big.NewInt(5e8))

		assert.Equal(t, big.NewInt(4e8), gasTip)
	})

	t.Run("tx should return the GasTipCap when tx type is not CeloDynamicFeeTxV2", func(t *testing.T) {
		gasTip := getGasTip(t, NewTx(&DynamicFeeTx{
			GasFeeCap: big.NewInt(9e8),
			GasTipCap: big.NewInt(3e8),
		}), big.NewInt(5e8))

		assert.Equal(t, big.NewInt(3e8), gasTip)
	})

	// CeloDynamicFeeTxV2
	t.Run("tx should return the difference between GasFeeCap and BaseFee with conversions between Celo and USDT when the transaction is CeloDynamicFeeTxV2 with the specified fee currency", func(t *testing.T) {
		gasTip := getGasTip(t, NewTx(&CeloDynamicFeeTxV2{
			FeeCurrency: &usdToken,
			GasFeeCap:   big.NewInt(18e8), // USD
			GasTipCap:   big.NewInt(8e8),  // USD
		}), big.NewInt(6e8)) // Celo

		assert.Equal(t, big.NewInt(3e8), gasTip)
	})

	t.Run("tx should return the GasTipCap with conversions between Celo and USDT when the transaction is CeloDynamicFeeTxV2 with the specified fee currency", func(t *testing.T) {
		gasTip := getGasTip(t, NewTx(&CeloDynamicFeeTxV2{
			FeeCurrency: &usdToken,
			GasFeeCap:   big.NewInt(18e8), // USD
			GasTipCap:   big.NewInt(4e8),  // USD
		}), big.NewInt(5e8)) // Celo

		assert.Equal(t, big.NewInt(2e8), gasTip)
	})

	t.Run("tx should return GasTipCap with conversions between Celo and USDT when the transaction is CeloDynamicFeeTxV2 with the specified fee currency but the base fee is nil", func(t *testing.T) {
		gasTip := getGasTip(t, NewTx(&CeloDynamicFeeTxV2{
			FeeCurrency: &usdToken,
			GasFeeCap:   big.NewInt(18e8), // USD
			GasTipCap:   big.NewInt(6e8),  // USD
		}), nil)

		assert.Equal(t, big.NewInt(3e8), gasTip)
	})

	// Error cases
	t.Run("tx should return an error when the fee currency which is not listed in the exchange rates is specified", func(t *testing.T) {
		tx := NewTx(&CeloDynamicFeeTxV2{
			FeeCurrency: &eurToken,
			GasFeeCap:   big.NewInt(18e8),
			GasTipCap:   big.NewInt(8e8),
		})

		res, err := tx.EffectiveGasTipInCelo(big.NewInt(5e8), exchangeRates)
		assert.ErrorIs(t, err, exchange.ErrUnregisteredFeeCurrency)
		require.Nil(t, res)
	})
}
