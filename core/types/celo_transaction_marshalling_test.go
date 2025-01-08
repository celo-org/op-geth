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

		feeCurrencyAddress  = common.HexToAddress("0x2F25deB3848C207fc8E0c34035B3Ba7fC157602B")
		toAddress           = common.HexToAddress("0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045")
		gatewayFeeRecipient = common.HexToAddress("0x471EcE3750Da237f93B8E339c536989b8978a438")
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
				Nonce:      10,
				Gas:        1e6,
				GasPrice:   big.NewInt(1e7),
				To:         &toAddress,
				Value:      big.NewInt(1e8),
				Data:       []byte{0x11, 0x22, 0x33, 0x44},
				CeloLegacy: false,
			}),
			json: `{
				"type": "0x0",
				"chainId": "0xa4ec",
				"nonce": "0xa",
				"to": "0xd8da6bf26964af9d7eed9e03e53415d37aa96045",
				"gas": "0xf4240",
				"gasPrice": "0x989680",
				"maxPriorityFeePerGas": null,
				"maxFeePerGas": null,
				"value": "0x5f5e100",
				"input": "0x11223344",
				"v": "0x149fc",
				"r": "0x444416344542ecfd5824c0173395cca148cfa58cf7572d81196314ad4f5bf1f1",
				"s": "0x23f6fd845489499c1170c8d0bd745f9fd3b99c2f4c979891b94e6764a03dcef0",
				"hash": "0xf0051b6799141b669b18cf456ffb3509e089c00544878e74769819b345f00866"
			}`,
			requiredFields: []string{"nonce", "gas", "gasPrice", "value", "input", "v", "r", "s"},
		},
		{
			txType:   "Celo LegacyTx",
			isCeloTx: true,
			tx: MustSignNewTx(key, signer, &LegacyTx{
				Nonce:               10,
				Gas:                 1e6,
				GasPrice:            big.NewInt(1e7),
				FeeCurrency:         &feeCurrencyAddress,
				GatewayFeeRecipient: &gatewayFeeRecipient,
				GatewayFee:          big.NewInt(1e7),
				To:                  &toAddress,
				Value:               big.NewInt(1e8),
				Data:                []byte{0x11, 0x22, 0x33, 0x44},
				CeloLegacy:          true,
			}),
			json: `{
				"type": "0x0",
				"chainId": "0xa4ec",
				"nonce": "0xa",
				"to": "0xd8da6bf26964af9d7eed9e03e53415d37aa96045",
				"gas": "0xf4240",
				"gasPrice": "0x989680",
				"maxPriorityFeePerGas": null,
				"maxFeePerGas": null,
				"value": "0x5f5e100",
				"input": "0x11223344",
				"v": "0x149fb",
				"r": "0x8ce12cd818c57c73354a1d18b5075bed356170f615246a5e4f7ac3a6a6c2d4c8",
				"s": "0x64ab56cbf08dd6cf742f083084ed66167fc110192ee3365555e643f102b5cec7",
				"hash": "0xfa30135c37ab29654ed0f75c07d6ba75cbb5d6739ab3249731ae80f80237bd5a",
				"feeCurrency": "0x2f25deb3848c207fc8e0c34035b3ba7fc157602b",
				"ethCompatible": false,
				"gatewayFee": "0x989680",
				"gatewayFeeRecipient": "0x471ece3750da237f93b8e339c536989b8978a438"
			}`,
			requiredFields: []string{"nonce", "gas", "gasPrice", "value", "input", "v", "r", "s"},
		},
		{
			txType:   "CeloDynamicFeeTx",
			isCeloTx: true,
			tx: MustSignNewTx(key, signer, &CeloDynamicFeeTx{
				ChainID:             chainId,
				Nonce:               10,
				GasTipCap:           big.NewInt(1),
				GasFeeCap:           big.NewInt(5e9),
				Gas:                 500000,
				FeeCurrency:         &feeCurrencyAddress,
				GatewayFeeRecipient: &gatewayFeeRecipient,
				GatewayFee:          big.NewInt(1e7),
				To:                  &toAddress,
				Value:               big.NewInt(1e8),
				Data:                []byte{0x11, 0x22, 0x33, 0x44},
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
				"to": "0xd8da6bf26964af9d7eed9e03e53415d37aa96045",
				"gas": "0x7a120",
				"gasPrice": null,
				"maxPriorityFeePerGas": "0x1",
				"maxFeePerGas": "0x12a05f200",
				"value": "0x5f5e100",
				"input": "0x11223344",
				"accessList": [
					{
						"address": "0xdac17f958d2ee523a2206206994597c13d831ec7",
						"storageKeys": [
							"0x2ab2bf4c5cabc3000e2502e33470a863db2755809d7561237424a0eb373154c2"
						]
					}
		      ],
				"v": "0x0",
				"r": "0xe5c3c7490d804f15ab18a2c864eecd824cde653593c3e2bd1898d09bf0d59a51",
				"s": "0x2c49f096c89f45dc0e0b24b7ce6e9b2b9cf13e7ab331e2e09c496080b4a6af2e",
				"hash": "0x9a24ac05fb511923a40babf2826ff78ca71214f6315ef53476e7d3fd2aa51d18",
				"feeCurrency": "0x2f25deb3848c207fc8e0c34035b3ba7fc157602b",
				"gatewayFee": "0x989680",
				"gatewayFeeRecipient": "0x471ece3750da237f93b8e339c536989b8978a438"
			}`,
			requiredFields: []string{"chainId", "nonce", "gas", "maxPriorityFeePerGas", "maxFeePerGas", "value", "input", "v", "r", "s"},
		},
		{
			txType:   "CeloDynamicFeeTxV2",
			isCeloTx: true,
			tx: MustSignNewTx(key, signer, &CeloDynamicFeeTxV2{
				ChainID:     chainId,
				Nonce:       10,
				GasTipCap:   big.NewInt(1),
				GasFeeCap:   big.NewInt(5e9),
				Gas:         500000,
				FeeCurrency: &feeCurrencyAddress,
				To:          &toAddress,
				Value:       big.NewInt(1e8),
				Data:        []byte{0x11, 0x22, 0x33, 0x44},
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
				"type": "0x7b",
				"chainId": "0xa4ec",
				"nonce": "0xa",
				"to": "0xd8da6bf26964af9d7eed9e03e53415d37aa96045",
				"gas": "0x7a120",
				"gasPrice": null,
				"maxPriorityFeePerGas": "0x1",
				"maxFeePerGas": "0x12a05f200",
				"value": "0x5f5e100",
				"input": "0x11223344",
				"accessList": [
					{
						"address": "0xdac17f958d2ee523a2206206994597c13d831ec7",
						"storageKeys": [
							"0x2ab2bf4c5cabc3000e2502e33470a863db2755809d7561237424a0eb373154c2"
						]
					}
		       ],
				"v": "0x0",
				"r": "0xd0caa0257ad5e276c4164cc7cf033e95db5dc6f8c880c8b94aa132efd2378597",
				"s": "0x63f213eb18899c862f1f9b84033f940bbc88de28f9af7a7c50b39e36896d1f51",
				"hash": "0x8710502f18e464a4e44d6d660127c6173e1ba70ca5684e09d736d3b5e9e63e16",
				"feeCurrency": "0x2f25deb3848c207fc8e0c34035b3ba7fc157602b"
			}`,
			requiredFields: []string{"chainId", "nonce", "gas", "maxPriorityFeePerGas", "maxFeePerGas", "value", "input", "v", "r", "s"},
		},
		{
			txType:   "CeloDenominatedTx",
			isCeloTx: true,
			// Skip signing due to unsupported transaction type
			tx: NewTx(&CeloDenominatedTx{
				ChainID:             chainId,
				Nonce:               10,
				GasTipCap:           big.NewInt(1),
				GasFeeCap:           big.NewInt(5e9),
				Gas:                 500000,
				FeeCurrency:         &feeCurrencyAddress,
				MaxFeeInFeeCurrency: big.NewInt(1e8),
				To:                  &toAddress,
				Value:               big.NewInt(1e8),
				Data:                []byte{0x11, 0x22, 0x33, 0x44},
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
				"type": "0x7a",
				"chainId": "0xa4ec",
				"nonce": "0xa",
				"to": "0xd8da6bf26964af9d7eed9e03e53415d37aa96045",
				"gas": "0x7a120",
				"gasPrice": null,
				"maxPriorityFeePerGas": "0x1",
				"maxFeePerGas": "0x12a05f200",
				"maxFeeInFeeCurrency": "0x5f5e100",
				"value": "0x5f5e100",
				"input": "0x11223344",
				"accessList": [
					{
						"address": "0xdac17f958d2ee523a2206206994597c13d831ec7",
						"storageKeys": [
							"0x2ab2bf4c5cabc3000e2502e33470a863db2755809d7561237424a0eb373154c2"
						]
					}
		       ],
				"v": "0x0",
				"r": "0x0",
				"s": "0x0",
				"hash": "0x4812438d07f69839658264fd6bf9022ceb97e87f71ac7ebc56a463f550a8c065",
				"feeCurrency": "0x2f25deb3848c207fc8e0c34035b3ba7fc157602b"
			}`,
			requiredFields: []string{"chainId", "nonce", "gas", "maxPriorityFeePerGas", "maxFeePerGas", "value", "input", "v", "r", "s"},
		},
	}

	// testMarshaling tests that the transaction marshals to the expected JSON
	testMarshaling := func(t *testing.T, tx *Transaction, expectedJson string, isCeloTxType bool) {
		t.Helper()

		txJsonOuter, err := json.Marshal(tx)
		assert.NoError(t, err)

		txJsonInner, isCeloTx, err := celoTransactionMarshal(tx)
		assert.NoError(t, err)

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
		assert.NoError(t, err)

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
			assert.NoError(t, err)

			delete(jsonMap, field)

			newJsonData, err := json.Marshal(jsonMap)
			assert.NoError(t, err)

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
