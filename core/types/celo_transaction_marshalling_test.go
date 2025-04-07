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
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCeloTransactionMarshalUnmarshal tests that each Celo transactions marshal and unmarshal correctly
func TestCeloTransactionMarshalUnmarshal(t *testing.T) {
	t.Parallel()

	var (
		chainId                      = big.NewInt(params.CeloMainnetChainID)
		gingerbreadForkHeight int64  = 5
		signerBlockTime       uint64 = 10
		cel2Time              uint64 = 15

		key, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
		signer = makeCeloSigner(
			&params.ChainConfig{
				ChainID:          chainId,
				Cel2Time:         &cel2Time,
				GingerbreadBlock: big.NewInt(gingerbreadForkHeight),
			},
			signerBlockTime,
			NewEIP155Signer(chainId),
		)

		nonce               = uint64(10)
		gasPrice            = big.NewInt(1e5)
		gas                 = uint64(1e6)
		feeCurrency         = common.HexToAddress("0x2F25deB3848C207fc8E0c34035B3Ba7fC157602B")
		gatewayFee          = big.NewInt(1e7)
		gatewayFeeRecipient = common.HexToAddress("0x471EcE3750Da237f93B8E339c536989b8978a438")
		to                  = common.HexToAddress("0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045")
		value               = big.NewInt(1e8)
		data                = []byte{0x12, 0x34, 0x56, 0x78}
		gasTipCap           = big.NewInt(1)
		gasFeeCap           = big.NewInt(1e9)
		accessListAddress   = common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")
		storageKey          = common.HexToHash("0x2ab2bf4c5cabc3000e2502e33470a863db2755809d7561237424a0eb373154c2")
	)

	tests := []struct {
		txType         string
		isCeloTx       bool
		tx             *Transaction
		json           string
		requiredFields []string
	}{
		{
			txType:   "Ethereum LegacyTx",
			isCeloTx: false,
			tx: MustSignNewTx(key, signer, &LegacyTx{
				Nonce:      nonce,
				Gas:        gas,
				GasPrice:   gasPrice,
				To:         &to,
				Value:      value,
				Data:       data,
				CeloLegacy: false,
			}),
			json: `{
				"type": "0x0",
				"chainId": "0xa4ec",
				"nonce": "0xa",
				"gasPrice": "0x186a0",
				"gas": "0xf4240",
				"maxPriorityFeePerGas": null,
				"maxFeePerGas": null,
				"to": "0xd8da6bf26964af9d7eed9e03e53415d37aa96045",
				"value": "0x5f5e100",
				"input": "0x12345678",
				"v": "0x149fb",
				"r": "0xdfa5f33872c59990dbbe5474e164492f8581d35e73cd8fc04aa6a90c68db8edd",
				"s": "0x649e6ce1d912fec1df4db2433cad338b1426db45a8d4f17735ceb032fc889b91",
				"hash": "0xd03cc6400e90416de0301727b1d9113426a732c880b379fad51738df8ce1db76"
			}`,
			requiredFields: []string{"nonce", "gas", "gasPrice", "value", "input", "v", "r", "s"},
		},
		{
			txType:   "Celo LegacyTx",
			isCeloTx: true,
			tx: MustSignNewTx(key, signer, &LegacyTx{
				Nonce:               nonce,
				GasPrice:            gasPrice,
				Gas:                 gas,
				FeeCurrency:         &feeCurrency,
				GatewayFeeRecipient: &gatewayFeeRecipient,
				GatewayFee:          gatewayFee,
				To:                  &to,
				Value:               value,
				Data:                data,
				CeloLegacy:          true,
			}),
			json: `{
				"type": "0x0",
				"chainId": "0xa4ec",
				"nonce": "0xa",
				"gasPrice": "0x186a0",
				"gas": "0xf4240",
				"maxPriorityFeePerGas": null,
				"maxFeePerGas": null,
				"feeCurrency": "0x2f25deb3848c207fc8e0c34035b3ba7fc157602b",
				"gatewayFeeRecipient": "0x471ece3750da237f93b8e339c536989b8978a438",
				"gatewayFee": "0x989680",
				"to": "0xd8da6bf26964af9d7eed9e03e53415d37aa96045",
				"value": "0x5f5e100",
				"input": "0x12345678",
				"v": "0x149fb",
				"r": "0x7114020891fdc29e8f661593aaac8e732eb81600a305f8cbb2cc16a7c8e74344",
				"s": "0x5ac556fc86adc0ae139d15d030d708e0d01bc3070a4555b18de911b76ec39766",
				"hash": "0x98d9e286c319b9e349eef956d81e2377ab73bb603f21b950d0a78db638f1008e",
				"ethCompatible": false
			}`,
			requiredFields: []string{"nonce", "gas", "gasPrice", "value", "input", "v", "r", "s"},
		},
		{
			txType:   "CeloDynamicFeeTx",
			isCeloTx: true,
			tx: MustSignNewTx(key, signer, &CeloDynamicFeeTx{
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
			}),
			json: `{
				"type": "0x7c",
				"chainId": "0xa4ec",
				"nonce": "0xa",
				"maxPriorityFeePerGas": "0x1",
				"maxFeePerGas": "0x3b9aca00",
				"gasPrice": null,
				"gas": "0xf4240",
				"feeCurrency": "0x2f25deb3848c207fc8e0c34035b3ba7fc157602b",
				"gatewayFeeRecipient": "0x471ece3750da237f93b8e339c536989b8978a438",
				"gatewayFee": "0x989680",
				"to": "0xd8da6bf26964af9d7eed9e03e53415d37aa96045",
				"value": "0x5f5e100",
				"input": "0x12345678",
				"accessList": [
					{
						"address": "0xdac17f958d2ee523a2206206994597c13d831ec7",
						"storageKeys": [
							"0x2ab2bf4c5cabc3000e2502e33470a863db2755809d7561237424a0eb373154c2"
						]
					}
				],
				"v": "0x1",
				"r": "0x59156829e96e9bcac82a15e74dd9488adc24d86aee847943c6938bf4c7c5f8b2",
				"s": "0x14f49fe32fdae94ec24d275b36ff75b97ab34b4270c4074814042d50504f4c74",
				"hash": "0x36610786a28d7ae3b5e4d07210b3f287e0fa78aec4093bd1532ec9780ed260be"
			}`,
			requiredFields: []string{"chainId", "nonce", "gas", "maxPriorityFeePerGas", "maxFeePerGas", "value", "input", "v", "r", "s"},
		},
		{
			txType:   "CeloDynamicFeeTxV2",
			isCeloTx: true,
			tx: MustSignNewTx(key, signer, &CeloDynamicFeeTxV2{
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
			}),
			json: `{
				"type": "0x7b",
				"chainId": "0xa4ec",
				"nonce": "0xa",
				"maxPriorityFeePerGas": "0x1",
				"maxFeePerGas": "0x3b9aca00",
				"gasPrice": null,
				"gas": "0xf4240",
				"to": "0xd8da6bf26964af9d7eed9e03e53415d37aa96045",
				"value": "0x5f5e100",
				"input": "0x12345678",
				"accessList": [
					{
						"address": "0xdac17f958d2ee523a2206206994597c13d831ec7",
						"storageKeys": [
							"0x2ab2bf4c5cabc3000e2502e33470a863db2755809d7561237424a0eb373154c2"
						]
					}
				],
				"feeCurrency": "0x2f25deb3848c207fc8e0c34035b3ba7fc157602b",
				"v": "0x1",
				"r": "0x605b2b9125a17c346f4569382cb127695e92ebdeedbcf249ec86e98feb6da04",
				"s": "0xd98bff12452ccf6268d63ead95bafa45d67020a9a35bce504ac68dbc1a49812",
				"hash": "0x9d668d91172a442c72509003235c386d7a0b2bd29da4411271f8f58231d33611"
			}`,
			requiredFields: []string{"chainId", "nonce", "gas", "maxPriorityFeePerGas", "maxFeePerGas", "value", "input", "v", "r", "s"},
		},
	}

	// testMarshaling tests that the transaction marshals to the expected JSON
	testMarshaling := func(t *testing.T, tx *Transaction, expectedJson string, isCeloTxType bool) {
		t.Helper()

		txJsonOuter, err := json.Marshal(tx)
		require.NoError(t, err)

		txJsonInner, isCeloTx, err := celoTransactionMarshal(tx)
		require.NoError(t, err)

		assert.Equal(t, isCeloTxType, isCeloTx)

		if isCeloTx {
			// For Celo transaction types
			// Make sure that celoTransactionMarshal produces the same JSON output as Transaction.MarshalJSON
			assert.Equal(t, txJsonOuter, txJsonInner)
		}

		// Make sure the output JSON is as expected
		assert.JSONEq(t, expectedJson, string(txJsonOuter))
	}

	// testUnmarshaling tests that the transaction unmarshals to the expected Transaction
	testUnmarshaling := func(t *testing.T, expectedTx *Transaction, jsonData string) {
		t.Helper()

		tx := new(Transaction)

		err := json.Unmarshal([]byte(jsonData), tx)
		require.NoError(t, err)

		// Reassign the signature values because *hexutil.Big decodes "0x0" as nil for the `abs` field.
		// This causes a mismatch with `big.NewInt(0)`
		v2, r2, s2 := tx.inner.rawSignatureValues()
		tx.inner.setSignatureValues(
			chainId,
			new(big.Int).SetBytes(v2.Bytes()),
			new(big.Int).SetBytes(r2.Bytes()),
			new(big.Int).SetBytes(s2.Bytes()),
		)

		assert.Equal(t, expectedTx.inner, tx.inner)
	}

	// testUnmarshalMissingRequiredField tests that the transaction fails to unmarshal if a required field is missing
	testUnmarshalMissingRequiredField := func(t *testing.T, jsonData string, requiredFields []string) {
		t.Helper()

		for _, field := range requiredFields {
			// Create a copy of the JSON data and remove one of the required fields
			var jsonMap map[string]interface{}
			err := json.Unmarshal([]byte(jsonData), &jsonMap)
			require.NoError(t, err)

			delete(jsonMap, field)

			newJsonData, err := json.Marshal(jsonMap)
			require.NoError(t, err)

			// Attempt to unmarshal the JSON data
			tx := new(Transaction)
			err = json.Unmarshal(newJsonData, tx)
			assert.ErrorContains(t, err, fmt.Sprintf("missing required field '%s'", field))
		}
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s should marshal to valid JSON successfully", test.txType), func(t *testing.T) {
			testMarshaling(t, test.tx, test.json, test.isCeloTx)
		})

		t.Run(fmt.Sprintf("%s should unmarshal valid JSON successfully", test.txType), func(t *testing.T) {
			testUnmarshaling(t, test.tx, test.json)
		})

		t.Run(fmt.Sprintf("%s should fail to marshal if required fields are missing", test.txType), func(t *testing.T) {
			testUnmarshalMissingRequiredField(t, test.json, test.requiredFields)
		})
	}
}
