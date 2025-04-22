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
	"bytes"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCeloTransactionRLPEncodingDecoding tests the RLP encoding and decoding of Celo transactions
func TestCeloTransactionRLPEncodingDecoding(t *testing.T) {
	t.Parallel()

	var (
		chainId             = big.NewInt(params.CeloMainnetChainID)
		nonce               = uint64(10)
		gasPrice            = big.NewInt(1e5)
		gas                 = uint64(1e6)
		feeCurrency         = common.HexToAddress("0x2F25deB3848C207fc8E0c34035B3Ba7fC157602B")
		gatewayFee          = big.NewInt(1e7)
		gatewayFeeRecipient = common.HexToAddress("0x471EcE3750Da237f93B8E339c536989b8978a438")
		to                  = common.HexToAddress("0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045")
		value               = big.NewInt(1e8)
		data                = []byte{0x12, 0x34, 0x56, 0x78}
		v                   = common.Hex2Bytes("0x149fc")
		r                   = common.Hex2Bytes("0x444416344542ecfd5824c0173395cca148cfa58cf7572d81196314ad4f5bf1f1")
		s                   = common.Hex2Bytes("0x23f6fd845489499c1170c8d0bd745f9fd3b99c2f4c979891b94e6764a03dcef0")
		gasTipCap           = big.NewInt(1)
		gasFeeCap           = big.NewInt(1e9)
		accessListAddress   = common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")
		storageKey          = common.HexToHash("0x2ab2bf4c5cabc3000e2502e33470a863db2755809d7561237424a0eb373154c2")
	)

	mustEncodeToBytes := func(t *testing.T, value interface{}) []byte {
		t.Helper()

		data, err := rlp.EncodeToBytes(value)
		require.NoError(t, err)

		return data
	}

	tests := []struct {
		txType string
		tx     *Transaction
		bytes  []byte
	}{
		{
			txType: "Celo LegacyTx",
			tx: NewTx(&LegacyTx{
				Nonce:               nonce,
				GasPrice:            gasPrice,
				Gas:                 gas,
				FeeCurrency:         &feeCurrency,
				GatewayFeeRecipient: &gatewayFeeRecipient,
				GatewayFee:          gatewayFee,
				To:                  &to,
				Value:               value,
				Data:                data,
				V:                   new(big.Int).SetBytes(v),
				R:                   new(big.Int).SetBytes(r),
				S:                   new(big.Int).SetBytes(s),
				CeloLegacy:          true,
			}),
			bytes: mustEncodeToBytes(t, []interface{}{
				nonce,
				gasPrice,
				gas,
				feeCurrency,
				gatewayFeeRecipient,
				gatewayFee,
				to,
				value,
				data,
				v,
				r,
				s,
			}),
		},
		{
			txType: "CeloDynamicFeeTx",
			tx: NewTx(&CeloDynamicFeeTx{
				ChainID:             chainId,
				Nonce:               nonce,
				GasTipCap:           gasTipCap,
				GasFeeCap:           gasFeeCap,
				Gas:                 gas,
				FeeCurrency:         &feeCurrency,
				GatewayFeeRecipient: &gatewayFeeRecipient,
				GatewayFee:          gatewayFee,
				To:                  &to,
				Value:               value,
				Data:                data,
				AccessList: AccessList{
					{
						Address: accessListAddress,
						StorageKeys: []common.Hash{
							storageKey,
						},
					},
				},
				V: new(big.Int).SetBytes(v),
				R: new(big.Int).SetBytes(r),
				S: new(big.Int).SetBytes(s),
			}),
			bytes: append(
				[]byte{CeloDynamicFeeTxType},
				mustEncodeToBytes(t, []interface{}{
					chainId,
					nonce,
					gasTipCap,
					gasFeeCap,
					gas,
					feeCurrency,
					gatewayFeeRecipient,
					gatewayFee,
					to,
					value,
					data,
					[]interface{}{
						[]interface{}{
							accessListAddress,
							[]interface{}{
								storageKey,
							},
						},
					},
					v,
					r,
					s,
				})...,
			),
		},
		{
			txType: "CeloDynamicFeeTxV2",
			tx: NewTx(&CeloDynamicFeeTxV2{
				ChainID:   chainId,
				Nonce:     nonce,
				GasTipCap: gasTipCap,
				GasFeeCap: gasFeeCap,
				Gas:       gas,
				To:        &to,
				Value:     value,
				Data:      data,
				AccessList: AccessList{
					{
						Address: accessListAddress,
						StorageKeys: []common.Hash{
							storageKey,
						},
					},
				},
				FeeCurrency: &feeCurrency,
				V:           new(big.Int).SetBytes(v),
				R:           new(big.Int).SetBytes(r),
				S:           new(big.Int).SetBytes(s),
			}),
			bytes: append(
				[]byte{CeloDynamicFeeTxV2Type},
				mustEncodeToBytes(t, []interface{}{
					chainId,
					nonce,
					gasTipCap,
					gasFeeCap,
					gas,
					to,
					value,
					data,
					[]interface{}{
						[]interface{}{
							accessListAddress,
							[]interface{}{
								storageKey,
							},
						},
					},
					feeCurrency,
					v,
					r,
					s,
				})...,
			),
		},
	}

	testEncodeRLP := func(t *testing.T, tx *Transaction, expectedByte []byte) {
		t.Helper()

		var buf bytes.Buffer

		err := tx.EncodeRLP(&buf)
		require.NoError(t, err)

		assert.Equal(t, expectedByte, buf.Bytes())
	}

	testMarshalBinary := func(t *testing.T, tx *Transaction, expectedByte []byte) {
		t.Helper()

		data, err := tx.MarshalBinary()
		require.NoError(t, err)

		assert.Equal(t, expectedByte, data)
	}

	testDecodeRLP := func(t *testing.T, expectedTx *Transaction, bytes []byte) {
		t.Helper()

		tx := new(Transaction)
		err := rlp.DecodeBytes(bytes, tx)
		require.NoError(t, err)

		assert.Equal(t, expectedTx.inner, tx.inner)
	}

	testUnmarshalBinary := func(t *testing.T, expectedTx *Transaction, bytes []byte) {
		t.Helper()

		tx := new(Transaction)
		err := tx.UnmarshalBinary(bytes)
		require.NoError(t, err)

		assert.Equal(t, expectedTx.inner, tx.inner)
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s should encode to valid RLP bytes", test.txType), func(t *testing.T) {
			if _, ok := test.tx.inner.(*LegacyTx); ok {
				testEncodeRLP(t, test.tx, test.bytes)
			} else {
				testEncodeRLP(t, test.tx, mustEncodeToBytes(t, test.bytes))
			}
		})

		t.Run(fmt.Sprintf("%s should marshal to valid RLP bytes", test.txType), func(t *testing.T) {
			testMarshalBinary(t, test.tx, test.bytes)
		})

		t.Run(fmt.Sprintf("%s should decode RLP bytes to Transaction", test.txType), func(t *testing.T) {
			if _, ok := test.tx.inner.(*LegacyTx); ok {
				testDecodeRLP(t, test.tx, test.bytes)
			} else {
				testDecodeRLP(t, test.tx, mustEncodeToBytes(t, test.bytes))
			}
		})

		t.Run(fmt.Sprintf("%s should unmarshal valid RLP bytes to Transaction", test.txType), func(t *testing.T) {
			testUnmarshalBinary(t, test.tx, test.bytes)
		})
	}
}
