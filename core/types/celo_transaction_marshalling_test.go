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
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
)

func TestCeloTransactionMarshal(t *testing.T) {
	t.Parallel()

	var (
		gingerbreadForkHeight int64  = 5
		signerBlockTime       uint64 = 10
		cel2Time              uint64 = 15

		key, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
		signer = makeCeloSigner(
			&params.ChainConfig{
				ChainID:          big.NewInt(params.CeloMainnetChainID),
				Cel2Time:         &cel2Time,
				GingerbreadBlock: big.NewInt(gingerbreadForkHeight),
			},
			signerBlockTime,
			NewEIP155Signer(big.NewInt(params.CeloMainnetChainID)),
		)

		feeCurrencyAddress  = common.HexToAddress("0x2F25deB3848C207fc8E0c34035B3Ba7fC157602B")
		toAddress           = common.HexToAddress("0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045")
		gatewayFeeRecipient = common.HexToAddress("0x471EcE3750Da237f93B8E339c536989b8978a438")
		accessListAddress   = common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")
		storageKey          = common.HexToHash("0x2ab2bf4c5cabc3000e2502e33470a863db2755809d7561237424a0eb373154c2")
	)

	tests := []struct {
		name         string
		tx           *Transaction
		expectedJson string
	}{
		{
			name: "Celo LegacyTx",
			tx: MustSignNewTx(key, signer, &LegacyTx{
				Nonce:               10,
				Gas:                 1e6,
				FeeCurrency:         &feeCurrencyAddress,
				GatewayFeeRecipient: &gatewayFeeRecipient,
				GatewayFee:          big.NewInt(1e7),
				To:                  &toAddress,
				Value:               big.NewInt(1e8),
				Data:                []byte{0x11, 0x22, 0x33, 0x44},
				CeloLegacy:          true,
			}),
			expectedJson: `{
				"type": "0x0",
				"chainId": "0xa4ec",
				"nonce": "0xa",
				"to": "0xd8da6bf26964af9d7eed9e03e53415d37aa96045",
				"gas": "0xf4240",
				"gasPrice": null,
				"maxPriorityFeePerGas": null,
				"maxFeePerGas": null,
				"value": "0x5f5e100",
				"input": "0x11223344",
				"v": "0x149fc",
				"r": "0x87e31aaf469f90072f46a87f48a60c6833f08216c0976e83d0f6d07ee16c2944",
				"s": "0x6366cf4f800913df4db35d6e3360b7bf3db087aff63291ab7bfbe6bef4865bc9",
				"hash": "0xa956b0aa70bc8e92ba260bb6865b46f36d8fc60cd72aaf472a3b2badc4379638",
				"feeCurrency": "0x2f25deb3848c207fc8e0c34035b3ba7fc157602b",
				"ethCompatible": false,
				"gatewayFee": "0x989680",
				"gatewayFeeRecipient": "0x471ece3750da237f93b8e339c536989b8978a438"
			}`,
		},
		{
			name: "CeloDynamicFeeTx",
			tx: MustSignNewTx(key, signer, &CeloDynamicFeeTx{
				ChainID:             big.NewInt(params.CeloMainnetChainID),
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
			expectedJson: `{
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
		},
		{
			name: "CeloDynamicFeeTxV2",
			tx: MustSignNewTx(key, signer, &CeloDynamicFeeTxV2{
				ChainID:     big.NewInt(params.CeloMainnetChainID),
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
			expectedJson: `{
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
		},
		{
			name: "CeloDenominatedTx",
			// Skip signing due to unsupported transaction type
			tx: NewTx(&CeloDenominatedTx{
				ChainID:             big.NewInt(params.CeloMainnetChainID),
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
			expectedJson: `{
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
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			txJsonOuter, err := json.Marshal(test.tx)
			assert.NoError(t, err)

			txJsonInner, isCeloTx, err := celoTransactionMarshal(test.tx)
			assert.NoError(t, err)

			// Make sure the transaction is unmarshalled by celoTransactionMarshal
			assert.True(t, isCeloTx)

			// Make sure Transaction.MarshalJSON returns the output of celoTransactionMarshal
			assert.Equal(t, txJsonOuter, txJsonInner)

			// Make sure the output JSON is as expected
			assert.JSONEq(t, test.expectedJson, string(txJsonOuter))
		})
	}
}
