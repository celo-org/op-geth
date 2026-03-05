package types

import (
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests that by default the celo legacy signer will sign transactions in a protected manner.
func TestProtectedCeloLegacyTxSigning(t *testing.T) {
	tx := newCeloTx(t)
	// Configure config and block time to enable the celoLegacy signer, legacy
	// transactions are deprecated after cel2
	cel2Time := uint64(2000)
	config := &params.ChainConfig{
		ChainID:  big.NewInt(10000),
		Cel2Time: &cel2Time,
	}
	number := new(big.Int).SetUint64(100)
	blockTime := uint64(1000)
	s := MakeSigner(config, number, blockTime)

	senderKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	signed, err := SignTx(tx, s, senderKey)
	require.NoError(t, err)

	// Check the sender just to be sure that the signing worked correctly
	actualSender, err := Sender(s, signed)
	require.NoError(t, err)
	require.Equal(t, crypto.PubkeyToAddress(senderKey.PublicKey), actualSender)
	// Validate that the transaction is protected
	require.True(t, signed.Protected())
}

// Tests that the celo legacy signer can still derive the sender of an unprotected transaction.
func TestUnprotectedCeloLegacyTxSenderDerivation(t *testing.T) {
	tx := newCeloTx(t)
	// Configure config and block time to enable the celoLegacy signer, legacy
	// transactions are deprecated after cel2
	cel2Time := uint64(2000)
	config := &params.ChainConfig{
		ChainID:  big.NewInt(10000),
		Cel2Time: &cel2Time,
	}
	number := new(big.Int).SetUint64(100)
	blockTime := uint64(1000)
	s := MakeSigner(config, number, blockTime)
	u := &unprotectedSigner{config.ChainID}

	senderKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	// Sign unprotected
	signed, err := SignTx(tx, u, senderKey)
	require.NoError(t, err)

	// Check that the sender can be derived with the signer from MakeSigner
	actualSender, err := Sender(s, signed)
	require.NoError(t, err)
	require.Equal(t, crypto.PubkeyToAddress(senderKey.PublicKey), actualSender)
	// Validate that the transaction is not protected
	require.False(t, signed.Protected())
}

func newCeloTx(t *testing.T) *Transaction {
	return NewTx(&LegacyTx{
		Nonce:    1,
		GasPrice: new(big.Int).SetUint64(10000),
		Gas:      100000,

		FeeCurrency:         randomAddress(t),
		GatewayFee:          new(big.Int).SetUint64(100),
		GatewayFeeRecipient: randomAddress(t),

		To:    randomAddress(t),
		Value: new(big.Int).SetUint64(1000),
		Data:  []byte{},

		CeloLegacy: true,
	})
}

// TestAccessListTxWrongChainId verifies that AccessListTx with a wrong chain ID
// returns ErrInvalidChainId when calling Sender.
// Regression test for https://github.com/celo-org/op-geth/issues/454
func TestAccessListTxWrongChainId(t *testing.T) {
	t.Parallel()

	// Create a signer with CeloMainnetChainID
	signerChainID := big.NewInt(params.CeloMainnetChainID)
	cel2Time := uint64(1000)
	chainConfig := *params.TestChainConfig
	chainConfig.ChainID = signerChainID
	chainConfig.Cel2Time = &cel2Time
	signer := MakeSigner(&chainConfig, big.NewInt(1), cel2Time+1)

	// Create an AccessListTx with a DIFFERENT chain ID
	wrongChainID := big.NewInt(999)
	signerKey, _ := crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	to := common.HexToAddress("0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045")

	txData := &AccessListTx{
		ChainID:  wrongChainID,
		Nonce:    10,
		GasPrice: big.NewInt(1e9),
		Gas:      500000,
		To:       &to,
		Value:    big.NewInt(1e18),
		Data:     []byte{0x11, 0x22},
	}

	// Sign with a signer that uses the wrong chain ID (to create a valid signature for that chain)
	wrongChainSigner := NewEIP2930Signer(wrongChainID)
	signedTx, err := SignNewTx(signerKey, wrongChainSigner, txData)
	require.NoError(t, err)

	// Now try to recover sender using the correct chain ID signer
	// This SHOULD fail with ErrInvalidChainId because tx.ChainId() != signerChainID
	_, err = signer.Sender(signedTx)
	require.ErrorIs(t, err, ErrInvalidChainId, "expected ErrInvalidChainId when tx chain ID doesn't match signer chain ID")
}

// TestChainIDExceptionTxs verifies that the historical transactions that were
// accepted with wrong chain IDs (before the validation fix) can still have
// their sender recovered correctly.
// See https://github.com/celo-org/op-geth/issues/454
func TestChainIDExceptionTxs(t *testing.T) {
	t.Parallel()

	hexToBigInt := func(hex string) *big.Int {
		hex = strings.TrimPrefix(hex, "0x")
		b, _ := new(big.Int).SetString(hex, 16)
		return b
	}

	tests := []struct {
		name           string
		signerChainID  int64       // The correct network chain ID
		txChainID      int64       // The wrong chain ID in the tx
		expectedHash   common.Hash // Expected tx hash (must match exception)
		expectedSender common.Address
		txData         *AccessListTx
	}{
		{
			name:           "Celo Sepolia exception tx",
			signerChainID:  params.CeloSepoliaChainID, // 11142220
			txChainID:      11162320,                  // Wrong chain ID in the tx
			expectedHash:   common.HexToHash("0x4564b9903cfe18814ffc2696e1ad141d9cc3a549dc4f5726e15f7be2e0ccaa25"),
			expectedSender: common.HexToAddress("0xEaD1E32D7D783f37d464d223Aa4E4C88f02Ae52f"),
			txData: &AccessListTx{
				ChainID:    big.NewInt(11162320),
				Nonce:      0,
				GasPrice:   big.NewInt(25001000000),
				Gas:        21000,
				To:         ptrTo(common.HexToAddress("0x52BCbd8Bf68EE24A15adcD05951a49aE6c168A14")),
				Value:      big.NewInt(1),
				Data:       []byte{},
				AccessList: AccessList{},
				V:          big.NewInt(0),
				R:          hexToBigInt("0xc96e8be6c653d8eb6f03842ffdc29347745c8122893f9cc9b64809d1bc49302d"),
				S:          hexToBigInt("0x49b51c8d25cff880327495cb1f322ebfbcb42151e9b617466eee2c737737f259"),
			},
		},
		{
			name:           "Celo Mainnet exception tx",
			signerChainID:  params.CeloMainnetChainID, // 42220
			txChainID:      44787,                     // Alfajores chain ID - wrong!
			expectedHash:   common.HexToHash("0xd6bdf3261df7e7a4db6bbc486bf091eb62dfd2883e335c31219b6a37d3febca1"),
			expectedSender: common.HexToAddress("0x5421E1E4fcc98AF133059723BA049B4ea1e9DE91"),
			txData: &AccessListTx{
				ChainID:    big.NewInt(44787),
				Nonce:      5,
				GasPrice:   big.NewInt(30000000000),
				Gas:        30000,
				To:         ptrTo(common.HexToAddress("0xC04b2FFAcc30C7FE19741E27ea150ccCc212e072")),
				Value:      big.NewInt(200000000000000),
				Data:       []byte{},
				AccessList: AccessList{},
				V:          big.NewInt(1),
				R:          hexToBigInt("0x197400aceb14cacc9a75710ebb4d3cba85538fc96a2b254d51a0db742c24ad08"),
				S:          hexToBigInt("0x2c18b58821fe02d107ff1fa9f4fd157bafb571bb83867d56618de3e1045141bb"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the transaction
			tx := NewTx(tt.txData)

			// Verify the hash matches the expected exception hash
			require.Equal(t, tt.expectedHash, tx.Hash(), "tx hash must match exception hash")

			// Create a signer for the correct network chain ID
			cel2Time := uint64(1000)
			chainConfig := *params.TestChainConfig
			chainConfig.ChainID = big.NewInt(tt.signerChainID)
			chainConfig.Cel2Time = &cel2Time
			signer := MakeSigner(&chainConfig, big.NewInt(1), cel2Time+1)

			// Verify that sender can be recovered despite chain ID mismatch
			sender, err := signer.Sender(tx)
			require.NoError(t, err, "sender recovery should succeed for exception tx")
			require.Equal(t, tt.expectedSender, sender, "recovered sender must match expected")
		})
	}
}

func ptrTo[T any](v T) *T {
	return &v
}

func randomAddress(t *testing.T) *common.Address {
	addr := common.Address{}
	_, err := rand.Read(addr[:])
	require.NoError(t, err)
	return &addr
}

// This signer mimics Homestead signing but for Celo transactions
type unprotectedSigner struct {
	chainID *big.Int
}

// ChainID implements Signer.
func (u *unprotectedSigner) ChainID() *big.Int {
	return u.chainID
}

// Equal implements Signer.
func (u *unprotectedSigner) Equal(Signer) bool {
	panic("unimplemented")
}

// Hash implements Signer.
func (u *unprotectedSigner) Hash(tx *Transaction) common.Hash {
	return rlpHash(baseCeloLegacyTxSigningFields(tx))
}

// Sender implements Signer.
func (u *unprotectedSigner) Sender(tx *Transaction) (common.Address, error) {
	panic("unimplemented")
}

// SignatureValues implements Signer.
func (u *unprotectedSigner) SignatureValues(tx *Transaction, sig []byte) (r *big.Int, s *big.Int, v *big.Int, err error) {
	r, s, v = decodeSignature(sig)
	return r, s, v, nil
}

type testCelo1Tx struct {
	data    TxData
	rawHash common.Hash
	hash    common.Hash
}

type celo1TxFixtures struct {
	signerKey     *ecdsa.PrivateKey
	signerAddress common.Address

	legacyTx           *testCelo1Tx
	accessListTx       *testCelo1Tx
	dynamicFeeTx       *testCelo1Tx
	celoLegacyTx       *testCelo1Tx
	celoDynamicFeeTx   *testCelo1Tx
	celoDynamicFeeTxV2 *testCelo1Tx
}

// createTestCelo1TxFixtures generates a set of test fixtures for transactions
// whose signatures are generated using celo-blockchain codebase
func createTestCelo1TxFixtures(t *testing.T) celo1TxFixtures {
	t.Helper()

	hexToBigInt := func(t *testing.T, hex string) *big.Int {
		t.Helper()

		hex = strings.TrimPrefix(hex, "0x")
		b, ok := new(big.Int).SetString(hex, 16)
		require.True(t, ok)

		return b
	}

	signerKey, _ := crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	signerAddress := crypto.PubkeyToAddress(signerKey.PublicKey)

	// common tx fields
	var (
		chainId    = big.NewInt(params.CeloMainnetChainID)
		nonce      = uint64(10)
		gasPrice   = big.NewInt(1e9)
		gasTipCap  = big.NewInt(1e7)
		gasFeeCap  = big.NewInt(1e10)
		gas        = uint64(5e5)
		to         = common.HexToAddress("0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045")
		value      = big.NewInt(1e18)
		data       = []byte{0x11, 0x22, 0x33, 0x44, 0x55}
		accessList = AccessList{
			{
				Address: common.HexToAddress("0xcebA9300f2b948710d2653dD7B07f33A8B32118C"),
				StorageKeys: []common.Hash{
					common.HexToHash("0xd6aca1be9729c13d677335161321649cccae6a591554772516700f986f942eaa"),
				},
			},
		}
		feeCurrency         = common.HexToAddress("0x2F25deB3848C207fc8E0c34035B3Ba7fC157602B")
		gatewayFeeRecipient = common.HexToAddress("0xde0B295669a9FD93d5F28D9Ec85E40f4cb697BAe")
		gatewayFee          = big.NewInt(1e8)
	)

	return celo1TxFixtures{
		signerKey:     signerKey,
		signerAddress: signerAddress,
		legacyTx: &testCelo1Tx{
			data: &LegacyTx{
				CeloLegacy: false,
				Nonce:      nonce,
				GasPrice:   gasPrice,
				Gas:        gas,
				To:         &to,
				Value:      value,
				Data:       data,
				V:          hexToBigInt(t, "0x0149fb"),
				R:          hexToBigInt(t, "0x0fb7068a40c34c3f6f6336175a9bfb9827dd1c86d3eace5b3827483ae31bffd9"),
				S:          hexToBigInt(t, "0x5e1df88b81d99356646e396497f7d9890845d58fd173ff353b13f76da7339a30"),
			},
			rawHash: common.HexToHash("0xb7e83f2a9024d3cfbcfe6376714ae2b8cd795d976179da06a330e8dc365be4ef"),
			hash:    common.HexToHash("0xb31d1e63ce610b3283ec26b81a17d2ba74a8ce1a0f56db44e85350c617a3aa04"),
		},
		accessListTx: &testCelo1Tx{
			data: &AccessListTx{
				ChainID:    chainId,
				Nonce:      nonce,
				GasPrice:   gasPrice,
				Gas:        gas,
				To:         &to,
				Value:      value,
				Data:       data,
				AccessList: accessList,
				V:          hexToBigInt(t, "0x00"),
				R:          hexToBigInt(t, "0xe010ac4c6b6be34cb7f569cb3d12f8646d8f9dfed04e528566211fd39bdc0703"),
				S:          hexToBigInt(t, "0x1ce2cf6ab68d6659dc29a1b0572dcc7baf78fb76a7597a1701da89cbebc74889"),
			},
			rawHash: common.HexToHash("0x0c93f7cea365d2312a9cfda5ff5e3896fd0f16b9711fe05c32c32d28b70b2aa7"),
			hash:    common.HexToHash("0x6fb3acc8403309f731beffbfb176f193b9ccd7d38828fe8017b0568901072dd3"),
		},
		dynamicFeeTx: &testCelo1Tx{
			data: &DynamicFeeTx{
				ChainID:    chainId,
				Nonce:      nonce,
				GasTipCap:  gasTipCap,
				GasFeeCap:  gasFeeCap,
				Gas:        gas,
				To:         &to,
				Value:      value,
				Data:       data,
				AccessList: accessList,
				V:          hexToBigInt(t, "0x00"),
				R:          hexToBigInt(t, "0x176f7aa3080bd805f07ad8a0c04027fa80c58f68a692c5facea50946a43ab5a9"),
				S:          hexToBigInt(t, "0x4706fbb2fc6163ebd5fd8bfcb66ac33dc0f832aea2f5806c89c0526211ad40c3"),
			},
			rawHash: common.HexToHash("0xc5c8e34888c0c761465dc1b0c192b3ea1ce533901f9d8dda0d43ccad54a1be63"),
			hash:    common.HexToHash("0x2c2ba8e9556c02f7db0d804a518ff0417a820d1ae4c85736aee93e41ec60c7fe"),
		},
		celoLegacyTx: &testCelo1Tx{
			data: &LegacyTx{
				CeloLegacy:          true,
				Nonce:               nonce,
				GasPrice:            gasPrice,
				Gas:                 gas,
				To:                  &to,
				Value:               value,
				Data:                data,
				FeeCurrency:         &feeCurrency,
				GatewayFeeRecipient: &gatewayFeeRecipient,
				GatewayFee:          gatewayFee,
				V:                   hexToBigInt(t, "0x0149fc"),
				R:                   hexToBigInt(t, "0xe78161be77dbde1b366b4431a95213a1d470273ec6cbe815cc851eec5924ba8d"),
				S:                   hexToBigInt(t, "0x20673a2a5ecbc09dd7e9d3762fc3a0c594984f4346d91a9f1259c78a0a2b0eec"),
			},
			rawHash: common.HexToHash("0x537a22034ab9b0cf0c6cfe3bbd27767903d23aba622217e7b8734d56f6c07a1d"),
			hash:    common.HexToHash("0xc66c343adfaecc4af7fcf38cf65d88ae936b8678b7a05d40e8912754292bbf57"),
		},
		celoDynamicFeeTx: &testCelo1Tx{
			data: &CeloDynamicFeeTx{
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
				AccessList:          accessList,
				V:                   hexToBigInt(t, "0x00"),
				R:                   hexToBigInt(t, "0x5ebfb28def4a6b115bcd8351f2056870fcaf0a36bd1045b56c40fa4856148830"),
				S:                   hexToBigInt(t, "0x28a04e99788825c37aae057e9800a6f39b670df97174b963c8e5d71e430934c6"),
			},
			rawHash: common.HexToHash("0xd8c3813a9471e747d12838b5ae51792de38b21a9c62bf5cce2133e28d37c1d8f"),
			hash:    common.HexToHash("0xcf5cfbb82684b1ffbcfc9a29736269cb42438fbd05bd9c6f4586559c9b87576a"),
		},
		celoDynamicFeeTxV2: &testCelo1Tx{
			data: &CeloDynamicFeeTxV2{
				ChainID:     chainId,
				Nonce:       nonce,
				GasTipCap:   gasTipCap,
				GasFeeCap:   gasFeeCap,
				Gas:         gas,
				To:          &to,
				Value:       value,
				Data:        data,
				AccessList:  accessList,
				FeeCurrency: &feeCurrency,
				V:           hexToBigInt(t, "0x00"),
				R:           hexToBigInt(t, "0x577c470cfda044ee14c5c5133de4a25c57854f885d10a16f0758e3401cdfc89d"),
				S:           hexToBigInt(t, "0x09facd14e350f964627ef212e5083233dff364f4aacd20067fedbefe0f5f673c"),
			},
			rawHash: common.HexToHash("0xbbf484aa35ab3783badf4278ec24d1e217984c45457ae9683ba46a73f761b7e9"),
			hash:    common.HexToHash("0xc41349a437ad7937105b51831ee245a5742ff32266a893d0f4f23ca451faca6d"),
		},
	}
}

// TestCeloSigner_Celo1TxRecovery verifies that transaction signatures created using celo-blockchain codebase
// can be correctly validated and recovered by signer of Celo2
func TestCeloSigner_Celo1TxRecovery(t *testing.T) {
	t.Parallel()

	cel2Time := uint64(1000)
	chainConfig := *params.TestChainConfig
	chainConfig.ChainID = big.NewInt(params.CeloMainnetChainID)
	chainConfig.Cel2Time = &cel2Time
	signer := MakeSigner(&chainConfig, big.NewInt(1), cel2Time-1)

	fixtures := createTestCelo1TxFixtures(t)

	tests := []struct {
		name                 string
		tx                   *Transaction
		expectedError        error
		expectedRawTxHash    common.Hash
		expectedSignedTxHash common.Hash
	}{
		{
			name:                 "LegacyTx",
			tx:                   NewTx(fixtures.legacyTx.data),
			expectedError:        nil,
			expectedRawTxHash:    fixtures.legacyTx.rawHash,
			expectedSignedTxHash: fixtures.legacyTx.hash,
		},
		{
			name:                 "AccessListTx",
			tx:                   NewTx(fixtures.accessListTx.data),
			expectedError:        nil,
			expectedRawTxHash:    fixtures.accessListTx.rawHash,
			expectedSignedTxHash: fixtures.accessListTx.hash,
		},
		{
			name:                 "DynamicFeeTx",
			tx:                   NewTx(fixtures.dynamicFeeTx.data),
			expectedError:        nil,
			expectedRawTxHash:    fixtures.dynamicFeeTx.rawHash,
			expectedSignedTxHash: fixtures.dynamicFeeTx.hash,
		},
		{
			name:                 "CeloLegacyTx",
			tx:                   NewTx(fixtures.celoLegacyTx.data),
			expectedError:        nil,
			expectedRawTxHash:    fixtures.celoLegacyTx.rawHash,
			expectedSignedTxHash: fixtures.celoLegacyTx.hash,
		},
		{
			name:                 "CeloDynamicFeeTx",
			tx:                   NewTx(fixtures.celoDynamicFeeTx.data),
			expectedError:        nil,
			expectedRawTxHash:    fixtures.celoDynamicFeeTx.rawHash,
			expectedSignedTxHash: fixtures.celoDynamicFeeTx.hash,
		},
		{
			name:                 "CeloDynamicFeeTxV2",
			tx:                   NewTx(fixtures.celoDynamicFeeTxV2.data),
			expectedError:        nil,
			expectedRawTxHash:    fixtures.celoDynamicFeeTxV2.rawHash,
			expectedSignedTxHash: fixtures.celoDynamicFeeTxV2.hash,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rawTxHash := signer.Hash(test.tx)
			signedTxHash := test.tx.Hash()

			recoveredSender, err := signer.Sender(test.tx)
			if test.expectedError == nil {
				require.NoError(t, err)
				assert.Equal(t, fixtures.signerAddress, recoveredSender)
			} else {
				assert.Equal(t, common.ZeroAddress, recoveredSender)
				assert.ErrorIs(t, test.expectedError, err)
			}

			assert.Equal(t, test.expectedRawTxHash, rawTxHash)
			assert.Equal(t, test.expectedSignedTxHash, signedTxHash)
		})
	}
}

// TestCeloSigner_SigningAndValidation tests the following:
// 1. A transaction can be correctly signed and validated in Celo2
// 2. Unsupported transaction types for Celo2 return appropriate error
func TestCeloSigner_SignAndRecovery(t *testing.T) {
	t.Parallel()

	cel2Time := uint64(1000)
	chainConfig := *params.TestChainConfig
	chainConfig.ChainID = big.NewInt(params.CeloMainnetChainID)
	chainConfig.Cel2Time = &cel2Time
	signer := MakeSigner(&chainConfig, big.NewInt(10), cel2Time)

	fixtures := createTestCelo1TxFixtures(t)

	tests := []struct {
		name                         string
		txData                       TxData
		expectedRawTxHash            common.Hash
		expectedSignedTxHash         common.Hash
		expectedSignatureValuesError error
		expectedSenderError          error
	}{
		// Ethereum Tx types
		{
			name:                         "LegacyTx",
			txData:                       fixtures.legacyTx.data,
			expectedRawTxHash:            fixtures.legacyTx.rawHash,
			expectedSignedTxHash:         fixtures.legacyTx.hash,
			expectedSenderError:          nil,
			expectedSignatureValuesError: nil,
		},
		{
			name:                         "AccessListTx",
			txData:                       fixtures.accessListTx.data,
			expectedRawTxHash:            fixtures.accessListTx.rawHash,
			expectedSignedTxHash:         fixtures.accessListTx.hash,
			expectedSenderError:          nil,
			expectedSignatureValuesError: nil,
		},
		{
			name:                         "DynamicFeeTx",
			txData:                       fixtures.dynamicFeeTx.data,
			expectedRawTxHash:            fixtures.dynamicFeeTx.rawHash,
			expectedSignedTxHash:         fixtures.dynamicFeeTx.hash,
			expectedSenderError:          nil,
			expectedSignatureValuesError: nil,
		},
		// Celo Tx types
		{
			name:                         "CeloLegacyTx",
			txData:                       fixtures.celoLegacyTx.data,
			expectedRawTxHash:            fixtures.celoLegacyTx.hash, // NOTE: deprecatedTxFuncs just returns tx hash
			expectedSignedTxHash:         common.Hash{},
			expectedSenderError:          ErrDeprecatedTxType,
			expectedSignatureValuesError: ErrDeprecatedTxType,
		},
		{
			name:                         "CeloDynamicFeeTx",
			txData:                       fixtures.celoDynamicFeeTx.data,
			expectedRawTxHash:            fixtures.celoDynamicFeeTx.hash, // NOTE: deprecatedTxFuncs just returns tx hash
			expectedSignedTxHash:         common.Hash{},
			expectedSenderError:          ErrDeprecatedTxType,
			expectedSignatureValuesError: ErrDeprecatedTxType,
		},
		{
			name:                         "CeloDynamicFeeTxV2",
			txData:                       fixtures.celoDynamicFeeTxV2.data,
			expectedRawTxHash:            fixtures.celoDynamicFeeTxV2.rawHash,
			expectedSignedTxHash:         fixtures.celoDynamicFeeTxV2.hash,
			expectedSenderError:          nil,
			expectedSignatureValuesError: nil,
		},
	}

	// testHash tests that Signer's Hash function returns the expected hash
	testHash := func(t *testing.T, tx TxData, expectedRawHash common.Hash) {
		t.Helper()

		rawTx := NewTx(tx)
		rawTxHash := signer.Hash(rawTx)
		assert.Equal(t, expectedRawHash, rawTxHash)
	}

	// testSender tests that Signer's Sender function recovers the expected address
	testSender := func(t *testing.T, tx TxData, expectedError error) {
		t.Helper()

		rawTx := NewTx(tx)
		sender, err := signer.Sender(rawTx)
		if expectedError == nil {
			require.NoError(t, err)
			assert.Equal(t, fixtures.signerAddress, sender)
		} else {
			require.ErrorAs(t, expectedError, &err)
		}
	}

	// testSignatureValues tests that Signer's SignatureValues function generates the expected signature values
	// Also tests that the generated signature and the recovered sender are correct
	testSignatureValues := func(t *testing.T, txData TxData, expectedSignedTxHash common.Hash, expectedSignTxError, expectedSenderError error) {
		t.Helper()

		// Extract expected values from the original Celo1 transaction
		expectedV, expectedR, expectedS := txData.rawSignatureValues()

		// Create new transaction without signatures
		txData = txData.copy()
		txData.setSignatureValues(chainConfig.ChainID, nil, nil, nil)

		// Sign transaction
		signedTx, err := SignNewTx(fixtures.signerKey, signer, txData)
		if expectedSignTxError != nil {
			require.ErrorAs(t, expectedSignTxError, &err)
			return
		}

		require.NoError(t, err)

		v, r, s := signedTx.RawSignatureValues()
		assert.Equal(t, expectedV, v)
		assert.Equal(t, expectedR, r)
		assert.Equal(t, expectedS, s)

		// Make sure the generated signature can be recovered correctly
		sender, err := signer.Sender(signedTx)
		if expectedSenderError != nil {
			require.ErrorAs(t, expectedSenderError, err)
			return
		}

		require.NoError(t, err)
		assert.Equal(t, fixtures.signerAddress, sender)

		// Check transaction Hash
		assert.Equal(t, expectedSignedTxHash, signedTx.Hash())
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testHash(t, test.txData, test.expectedRawTxHash)
			testSender(t, test.txData, test.expectedSenderError)
			testSignatureValues(t, test.txData, test.expectedSignedTxHash, test.expectedSignatureValuesError, test.expectedSenderError)
		})
	}
}

// TestLatestSigner_Celo ensures that LatestSigner selects the correct
// base signer for Celo L2 immediately after launch and after each
// network upgrade
func TestLatestSigner_Celo(t *testing.T) {
	newUint64Ptr := func(x uint64) *uint64 { return &x }

	t.Run("returns celoSigner with LondonSigner right after Celo L2 release", func(t *testing.T) {
		cfg := &params.ChainConfig{
			ChainID:     big.NewInt(params.CeloMainnetChainID),
			Optimism:    &params.OptimismConfig{},
			Cel2Time:    newUint64Ptr(1742957258),
			EIP155Block: big.NewInt(0),
			BerlinBlock: big.NewInt(31056500),
			LondonBlock: big.NewInt(31056500),
			CancunTime:  newUint64Ptr(1742957258),
			PragueTime:  nil,
			IsthmusTime: nil,
		}

		got := LatestSigner(cfg)

		assert.Equal(t, &celoSigner{
			upstreamSigner: NewLondonSigner(cfg.ChainID),
			chainID:        cfg.ChainID,
			activatedForks: celoForks,
		}, got)
	})

	t.Run("returns celoSigner with IsthmusSigner once Isthmus upgrade is active", func(t *testing.T) {
		cfg := &params.ChainConfig{
			ChainID:     big.NewInt(params.CeloMainnetChainID),
			Optimism:    &params.OptimismConfig{},
			Cel2Time:    newUint64Ptr(1742957258),
			EIP155Block: big.NewInt(0),
			BerlinBlock: big.NewInt(31056500),
			LondonBlock: big.NewInt(31056500),
			CancunTime:  newUint64Ptr(1742957258),
			PragueTime:  newUint64Ptr(1754006400),
			IsthmusTime: newUint64Ptr(1754006400),
		}

		got := LatestSigner(cfg)

		assert.Equal(t, &celoSigner{
			upstreamSigner: NewIsthmusSigner(cfg.ChainID),
			chainID:        cfg.ChainID,
			activatedForks: celoForks,
		}, got)
	})
}
