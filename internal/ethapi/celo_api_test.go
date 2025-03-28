package ethapi

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/exchange"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/internal/blocktest"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// tx fields
	nonce               uint64 = 1
	gasPrice                   = big.NewInt(1000)
	gasLimit            uint64 = 100000
	feeCurrency                = common.HexToAddress("0x0000000000000000000000000000000000000bbb")
	gatewayFee                 = big.NewInt(500)
	gatewayFeeRecipient        = common.HexToAddress("0x0000000000000000000000000000000000000ccc")
	to                         = common.HexToAddress("0x0000000000000000000000000000000000000aaa")
	value                      = big.NewInt(10)
	// block fields
	baseFee                 = big.NewInt(100)
	transactionIndex uint64 = 15
	blockhash               = common.HexToHash("0x6ba4a8c1bfe2619eb498e5296e81b1c393b13cba0198ed63dea0ee3aa619b073")
	blockNumber      uint64 = 100
	blockTime        uint64 = 100
)

func TestNewRPCTransactionLegacy(t *testing.T) {
	config := allEnabledChainConfig()
	// Set cel2 time to 2000 so that we don't activate the cel2 fork.
	var cel2Time uint64 = 2000
	config.Cel2Time = &cel2Time
	s := types.MakeSigner(config, new(big.Int).SetUint64(blockNumber), blockTime)

	key, err := crypto.GenerateKey()
	require.NoError(t, err)

	t.Run("WithCeloFields", func(t *testing.T) {
		tx := types.NewTx(&types.LegacyTx{
			Nonce:    nonce,
			GasPrice: gasPrice,
			Gas:      gasLimit,

			FeeCurrency:         &feeCurrency,
			GatewayFee:          gatewayFee,
			GatewayFeeRecipient: &gatewayFeeRecipient,

			To:    &to,
			Value: value,
			Data:  []byte{},

			CeloLegacy: true,
		})

		signed, err := types.SignTx(tx, s, key)
		require.NoError(t, err)
		rpcTx := newRPCTransaction(signed, blockhash, blockNumber, blockTime, transactionIndex, baseFee, config, nil)
		checkTxFields(t, signed, rpcTx, s, blockhash, blockNumber, transactionIndex, nil)
	})

	t.Run("WithoutCeloFields", func(t *testing.T) {
		tx := types.NewTx(&types.LegacyTx{
			Nonce:    nonce,
			GasPrice: gasPrice,
			Gas:      gasLimit,

			To:    &to,
			Value: value,
			Data:  []byte{},
		})
		signed, err := types.SignTx(tx, s, key)
		require.NoError(t, err)
		rpcTx := newRPCTransaction(signed, blockhash, blockNumber, blockTime, transactionIndex, baseFee, config, nil)
		checkTxFields(t, signed, rpcTx, s, blockhash, blockNumber, transactionIndex, nil)
	})
}

func TestNewRPCTransactionDynamicFee(t *testing.T) {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	feeCap := big.NewInt(1000)
	tipCap := big.NewInt(100)

	t.Run("PendingTransactions", func(t *testing.T) {
		// For pending transactions we expect the gas price to be the gas fee cap.
		gasFeeCap := func(t *testing.T, tx *types.Transaction, rpcTx *RPCTransaction) {
			assert.Equal(t, (*hexutil.Big)(feeCap), rpcTx.GasPrice)
		}
		overrides := map[string]func(*testing.T, *types.Transaction, *RPCTransaction){"gasPrice": gasFeeCap}
		config := allEnabledChainConfig()
		s := types.MakeSigner(config, new(big.Int).SetUint64(blockNumber), blockTime)

		// An empty bockhash signals pending transactions (I.E no mined block)
		blockhash := common.Hash{}
		t.Run("DynamicFeeTx", func(t *testing.T) {
			tx := types.NewTx(&types.DynamicFeeTx{
				ChainID:   config.ChainID,
				Nonce:     nonce,
				Gas:       gasLimit,
				GasFeeCap: feeCap,
				GasTipCap: tipCap,

				To:    &to,
				Value: value,
				Data:  []byte{},
			})

			signed, err := types.SignTx(tx, s, key)
			require.NoError(t, err)

			rpcTx := newRPCTransaction(signed, blockhash, blockNumber, blockTime, transactionIndex, baseFee, config, nil)
			checkTxFields(t, signed, rpcTx, s, blockhash, blockNumber, transactionIndex, overrides)
		})

		t.Run("CeloDynamicFeeTxV2", func(t *testing.T) {
			tx := types.NewTx(&types.CeloDynamicFeeTxV2{
				ChainID:     config.ChainID,
				Nonce:       nonce,
				Gas:         gasLimit,
				GasFeeCap:   feeCap,
				GasTipCap:   tipCap,
				FeeCurrency: &feeCurrency,

				To:    &to,
				Value: value,
				Data:  []byte{},
			})

			signed, err := types.SignTx(tx, s, key)
			require.NoError(t, err)

			rpcTx := newRPCTransaction(signed, blockhash, blockNumber, blockTime, transactionIndex, baseFee, config, nil)
			checkTxFields(t, signed, rpcTx, s, blockhash, blockNumber, transactionIndex, overrides)
		})
	})

	t.Run("PreGingerbreadMinedDynamicTxs", func(t *testing.T) {
		nilGasPrice := func(t *testing.T, tx *types.Transaction, rpcTx *RPCTransaction) {
			assert.Nil(t, rpcTx.GasPrice)
		}
		overrides := map[string]func(*testing.T, *types.Transaction, *RPCTransaction){"gasPrice": nilGasPrice}
		// For a pre gingerbread mined dynamic txs we expect the gas price to be unset, because without the state we
		// cannot retrieve the base fee, and we currently have no implementation in op-geth to handle retrieving the
		// base fee from state.
		config := allEnabledChainConfig()
		config.GingerbreadBlock = big.NewInt(200) // Setup config so that gingerbread is not active.
		cel2Time := uint64(1000)
		config.Cel2Time = &cel2Time // also deactivate cel2
		s := types.MakeSigner(config, new(big.Int).SetUint64(blockNumber), blockTime)

		t.Run("DynamicFeeTx", func(t *testing.T) {
			tx := types.NewTx(&types.DynamicFeeTx{
				ChainID:   config.ChainID,
				Nonce:     nonce,
				Gas:       gasLimit,
				GasFeeCap: feeCap,
				GasTipCap: tipCap,

				To:    &to,
				Value: value,
				Data:  []byte{},
			})

			signed, err := types.SignTx(tx, s, key)
			require.NoError(t, err)

			rpcTx := newRPCTransaction(signed, blockhash, blockNumber, blockTime, transactionIndex, baseFee, config, nil)
			checkTxFields(t, signed, rpcTx, s, blockhash, blockNumber, transactionIndex, overrides)
		})

		t.Run("CeloDynamicFeeTx", func(t *testing.T) {
			tx := types.NewTx(&types.CeloDynamicFeeTx{
				ChainID:             config.ChainID,
				Nonce:               nonce,
				Gas:                 gasLimit,
				GasFeeCap:           feeCap,
				GasTipCap:           tipCap,
				GatewayFee:          gatewayFee,
				GatewayFeeRecipient: &gatewayFeeRecipient,

				To:    &to,
				Value: value,
				Data:  []byte{},
			})

			signed, err := types.SignTx(tx, s, key)
			require.NoError(t, err)

			rpcTx := newRPCTransaction(signed, blockhash, blockNumber, blockTime, transactionIndex, baseFee, config, nil)
			checkTxFields(t, signed, rpcTx, s, blockhash, blockNumber, transactionIndex, overrides)
		})
	})

	t.Run("PostGingerbreadMinedDynamicTxsWithNativeFeeCurrency", func(t *testing.T) {
		// For a post gingerbread mined dynamic tx with a native fee currency we expect the gas price to be the
		// effective gas price calculated with the base fee available on the block.
		effectiveGasPrice := func(t *testing.T, tx *types.Transaction, rpcTx *RPCTransaction) {
			assert.Equal(t, (*hexutil.Big)(effectiveGasPrice(tx, baseFee)), rpcTx.GasPrice)
		}
		overrides := map[string]func(*testing.T, *types.Transaction, *RPCTransaction){"gasPrice": effectiveGasPrice}

		config := allEnabledChainConfig()
		s := types.MakeSigner(config, new(big.Int).SetUint64(blockNumber), blockTime)

		t.Run("DynamicFeeTx", func(t *testing.T) {
			tx := types.NewTx(&types.DynamicFeeTx{
				ChainID:   config.ChainID,
				Nonce:     nonce,
				Gas:       gasLimit,
				GasFeeCap: feeCap,
				GasTipCap: tipCap,

				To:    &to,
				Value: value,
				Data:  []byte{},
			})

			signed, err := types.SignTx(tx, s, key)
			require.NoError(t, err)

			rpcTx := newRPCTransaction(signed, blockhash, blockNumber, blockTime, transactionIndex, baseFee, config, nil)
			checkTxFields(t, signed, rpcTx, s, blockhash, blockNumber, transactionIndex, overrides)
		})

		t.Run("CeloDynamicFeeTx", func(t *testing.T) {
			// CeloDynamicFeeTxs are deprecated after cel2 so we need to ensure cel2time is not activated
			config := allEnabledChainConfig()
			cel2Time := uint64(1000)
			config.Cel2Time = &cel2Time
			s := types.MakeSigner(config, new(big.Int).SetUint64(blockNumber), blockTime)

			tx := types.NewTx(&types.CeloDynamicFeeTx{
				ChainID:             config.ChainID,
				Nonce:               nonce,
				Gas:                 gasLimit,
				GasFeeCap:           feeCap,
				GasTipCap:           tipCap,
				GatewayFee:          gatewayFee,
				GatewayFeeRecipient: &gatewayFeeRecipient,

				To:    &to,
				Value: value,
				Data:  []byte{},
			})

			signed, err := types.SignTx(tx, s, key)
			require.NoError(t, err)

			rpcTx := newRPCTransaction(signed, blockhash, blockNumber, blockTime, transactionIndex, baseFee, config, nil)
			checkTxFields(t, signed, rpcTx, s, blockhash, blockNumber, transactionIndex, overrides)
		})

		t.Run("CeloDynamicFeeTxV2", func(t *testing.T) {
			tx := types.NewTx(&types.CeloDynamicFeeTxV2{
				ChainID:   config.ChainID,
				Nonce:     nonce,
				Gas:       gasLimit,
				GasFeeCap: feeCap,
				GasTipCap: tipCap,

				To:    &to,
				Value: value,
				Data:  []byte{},
			})

			signed, err := types.SignTx(tx, s, key)
			require.NoError(t, err)

			rpcTx := newRPCTransaction(signed, blockhash, blockNumber, blockTime, transactionIndex, baseFee, config, nil)
			checkTxFields(t, signed, rpcTx, s, blockhash, blockNumber, transactionIndex, overrides)
		})

		// TODO unskip this when cip 66 txs are enabled currently they are not supporeted in the celo signer.
		t.Run("CeloDenominatedTx", func(t *testing.T) {
			t.Skip("CeloDenominatedTx is currently not supported in the celo signer")
			tx := types.NewTx(&types.CeloDenominatedTx{
				ChainID:             config.ChainID,
				Nonce:               nonce,
				Gas:                 gasLimit,
				GasFeeCap:           feeCap,
				GasTipCap:           tipCap,
				FeeCurrency:         &feeCurrency,
				MaxFeeInFeeCurrency: big.NewInt(100000),

				To:    &to,
				Value: value,
				Data:  []byte{},
			})

			signed, err := types.SignTx(tx, s, key)
			require.NoError(t, err)

			rpcTx := newRPCTransaction(signed, blockhash, blockNumber, blockTime, transactionIndex, baseFee, config, nil)
			checkTxFields(t, signed, rpcTx, s, blockhash, blockNumber, transactionIndex, overrides)
		})
	})

	t.Run("PostGingerbreadPreCel2MinedDynamicTxsWithNonNativeFeeCurrency", func(t *testing.T) {
		// For a post gingerbread mined dynamic txs with a non native fee currency we expect the gas price to be unset,
		// because without the state we cannot retrieve the base fee, and we currently have no implementation in op-geth
		// to handle retrieving the base fee from state.

		nilGasPrice := func(t *testing.T, tx *types.Transaction, rpcTx *RPCTransaction) {
			assert.Nil(t, rpcTx.GasPrice)
		}
		overrides := map[string]func(*testing.T, *types.Transaction, *RPCTransaction){"gasPrice": nilGasPrice}

		config := allEnabledChainConfig()
		cel2Time := uint64(1000)
		config.Cel2Time = &cel2Time // Deactivate cel2
		s := types.MakeSigner(config, new(big.Int).SetUint64(blockNumber), blockTime)

		t.Run("CeloDynamicFeeTx", func(t *testing.T) {
			// CeloDynamicFeeTxs are deprecated after cel2 so we need to ensure cel2time is not activated
			config := allEnabledChainConfig()
			cel2Time := uint64(1000)
			config.Cel2Time = &cel2Time
			s := types.MakeSigner(config, new(big.Int).SetUint64(blockNumber), blockTime)

			tx := types.NewTx(&types.CeloDynamicFeeTx{
				ChainID:             config.ChainID,
				Nonce:               nonce,
				Gas:                 gasLimit,
				GasFeeCap:           feeCap,
				GasTipCap:           tipCap,
				FeeCurrency:         &feeCurrency,
				GatewayFee:          gatewayFee,
				GatewayFeeRecipient: &gatewayFeeRecipient,

				To:    &to,
				Value: value,
				Data:  []byte{},
			})

			signed, err := types.SignTx(tx, s, key)
			require.NoError(t, err)

			rpcTx := newRPCTransaction(signed, blockhash, blockNumber, blockTime, transactionIndex, baseFee, config, nil)
			checkTxFields(t, signed, rpcTx, s, blockhash, blockNumber, transactionIndex, overrides)
		})

		t.Run("CeloDynamicFeeTxV2", func(t *testing.T) {
			tx := types.NewTx(&types.CeloDynamicFeeTxV2{
				ChainID:     config.ChainID,
				Nonce:       nonce,
				Gas:         gasLimit,
				GasFeeCap:   feeCap,
				GasTipCap:   tipCap,
				FeeCurrency: &feeCurrency,

				To:    &to,
				Value: value,
				Data:  []byte{},
			})

			signed, err := types.SignTx(tx, s, key)
			require.NoError(t, err)

			rpcTx := newRPCTransaction(signed, blockhash, blockNumber, blockTime, transactionIndex, baseFee, config, nil)
			checkTxFields(t, signed, rpcTx, s, blockhash, blockNumber, transactionIndex, overrides)
		})
	})

	t.Run("PostCel2MinedDynamicTxs", func(t *testing.T) {
		receipt := &types.Receipt{}
		receipt.EffectiveGasPrice = big.NewInt(1234)
		effectiveGasPrice := func(t *testing.T, tx *types.Transaction, rpcTx *RPCTransaction) {
			assert.Equal(t, (*hexutil.Big)(receipt.EffectiveGasPrice), rpcTx.GasPrice)
		}
		overrides := map[string]func(*testing.T, *types.Transaction, *RPCTransaction){"gasPrice": effectiveGasPrice}

		config := allEnabledChainConfig()
		s := types.MakeSigner(config, new(big.Int).SetUint64(blockNumber), blockTime)

		t.Run("CeloDynamicFeeTxV2", func(t *testing.T) {
			// For a pre gingerbread mined dynamic fee tx we expect the gas price to be unset.
			tx := types.NewTx(&types.CeloDynamicFeeTxV2{
				ChainID:     config.ChainID,
				Nonce:       nonce,
				Gas:         gasLimit,
				GasFeeCap:   feeCap,
				GasTipCap:   tipCap,
				FeeCurrency: &feeCurrency,

				To:    &to,
				Value: value,
				Data:  []byte{},
			})

			signed, err := types.SignTx(tx, s, key)
			require.NoError(t, err)

			rpcTx := newRPCTransaction(signed, blockhash, blockNumber, blockTime, transactionIndex, baseFee, config, receipt)
			checkTxFields(t, signed, rpcTx, s, blockhash, blockNumber, transactionIndex, overrides)
		})
	})
}

func allEnabledChainConfig() *params.ChainConfig {
	zeroTime := uint64(0)
	return &params.ChainConfig{
		ChainID:             big.NewInt(params.CeloAlfajoresChainID),
		HomesteadBlock:      big.NewInt(0),
		EIP150Block:         big.NewInt(0),
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
		PetersburgBlock:     big.NewInt(0),
		IstanbulBlock:       big.NewInt(0),
		MuirGlacierBlock:    big.NewInt(0),
		BerlinBlock:         big.NewInt(0),
		LondonBlock:         big.NewInt(0),
		ArrowGlacierBlock:   big.NewInt(0),
		GrayGlacierBlock:    big.NewInt(0),
		ShanghaiTime:        &zeroTime,
		CancunTime:          &zeroTime,
		RegolithTime:        &zeroTime,
		CanyonTime:          &zeroTime,
		EcotoneTime:         &zeroTime,
		FjordTime:           &zeroTime,
		Cel2Time:            &zeroTime,
		GingerbreadBlock:    big.NewInt(0),
		Optimism:            &params.OptimismConfig{},
	}
}

// checkTxFields for the most part checks that the fields of the rpcTx match those of the provided tx, it allows for
// overriding some checks by providing a map of fieldName -> overrideFunc.
func checkTxFields(
	t *testing.T,
	tx *types.Transaction,
	rpcTx *RPCTransaction,
	signer types.Signer,
	blockhash common.Hash,
	blockNumber uint64,
	transactionIndex uint64,
	overrides map[string]func(*testing.T, *types.Transaction, *RPCTransaction),
) {
	// Added fields (not part of the transaction type)
	//
	// If blockhash is empty it signifies a pending tx and for pending txs the block hash, block number and tx index are
	// not set on the rpcTx. on the result.
	if blockhash == (common.Hash{}) {
		assert.Nil(t, rpcTx.BlockHash)
		assert.Nil(t, rpcTx.BlockNumber)
		assert.Nil(t, rpcTx.TransactionIndex)
	} else {
		assert.Equal(t, &blockhash, rpcTx.BlockHash)
		assert.Equal(t, (*hexutil.Big)(big.NewInt(int64(blockNumber))), rpcTx.BlockNumber)
		assert.Equal(t, hexutil.Uint64(transactionIndex), *rpcTx.TransactionIndex)
	}

	from, err := types.Sender(signer, tx)
	require.NoError(t, err)

	assert.Equal(t, from, rpcTx.From)
	assert.Equal(t, hexutil.Uint64(tx.Gas()), rpcTx.Gas)
	assert.Equal(t, tx.To(), rpcTx.To)
	override, ok := overrides["gasPrice"]
	if ok {
		override(t, tx, rpcTx)
	} else {
		assert.Equal(t, (*hexutil.Big)(tx.GasPrice()), rpcTx.GasPrice)
	}
	switch tx.Type() {
	case types.DynamicFeeTxType, types.CeloDynamicFeeTxType, types.CeloDynamicFeeTxV2Type, types.CeloDenominatedTxType:
		assert.Equal(t, (*hexutil.Big)(tx.GasFeeCap()), rpcTx.GasFeeCap)
		assert.Equal(t, (*hexutil.Big)(tx.GasTipCap()), rpcTx.GasTipCap)
	default:
		assert.Nil(t, rpcTx.GasFeeCap)
		assert.Nil(t, rpcTx.GasTipCap)
	}
	assert.Equal(t, (*hexutil.Big)(tx.BlobGasFeeCap()), rpcTx.MaxFeePerBlobGas)
	assert.Equal(t, tx.Hash(), rpcTx.Hash)
	assert.Equal(t, (hexutil.Bytes)(tx.Data()), rpcTx.Input)
	assert.Equal(t, hexutil.Uint64(tx.Nonce()), rpcTx.Nonce)
	assert.Equal(t, tx.To(), rpcTx.To)
	assert.Equal(t, (*hexutil.Big)(tx.Value()), rpcTx.Value)
	assert.Equal(t, hexutil.Uint64(tx.Type()), rpcTx.Type)
	switch tx.Type() {
	case types.AccessListTxType, types.DynamicFeeTxType, types.CeloDynamicFeeTxType, types.CeloDynamicFeeTxV2Type, types.CeloDenominatedTxType, types.BlobTxType:
		assert.Equal(t, tx.AccessList(), *rpcTx.Accesses)
	default:
		assert.Nil(t, rpcTx.Accesses)
	}

	assert.Equal(t, (*hexutil.Big)(tx.ChainId()), rpcTx.ChainID)
	assert.Equal(t, tx.BlobHashes(), rpcTx.BlobVersionedHashes)

	v, r, s := tx.RawSignatureValues()
	assert.Equal(t, (*hexutil.Big)(v), rpcTx.V)
	assert.Equal(t, (*hexutil.Big)(r), rpcTx.R)
	assert.Equal(t, (*hexutil.Big)(s), rpcTx.S)

	switch tx.Type() {
	case types.AccessListTxType, types.DynamicFeeTxType, types.CeloDynamicFeeTxType, types.CeloDynamicFeeTxV2Type, types.CeloDenominatedTxType, types.BlobTxType:
		yparity := (hexutil.Uint64)(v.Sign())
		assert.Equal(t, &yparity, rpcTx.YParity)
	default:
		assert.Nil(t, rpcTx.YParity)
	}

	// optimism fields
	switch tx.Type() {
	case types.DepositTxType:
		assert.Equal(t, tx.SourceHash(), rpcTx.SourceHash)
		assert.Equal(t, tx.Mint(), rpcTx.Mint)
		assert.Equal(t, tx.IsSystemTx(), rpcTx.IsSystemTx)
	default:
		assert.Nil(t, rpcTx.SourceHash)
		assert.Nil(t, rpcTx.Mint)
		assert.Nil(t, rpcTx.IsSystemTx)
	}

	assert.Nil(t, rpcTx.DepositReceiptVersion)

	// celo fields
	assert.Equal(t, tx.FeeCurrency(), rpcTx.FeeCurrency)
	assert.Equal(t, (*hexutil.Big)(tx.MaxFeeInFeeCurrency()), rpcTx.MaxFeeInFeeCurrency)
	if tx.Type() == types.LegacyTxType && tx.IsCeloLegacy() {
		assert.Equal(t, false, *rpcTx.EthCompatible)
	} else {
		assert.Nil(t, rpcTx.EthCompatible)
	}
	assert.Equal(t, (*hexutil.Big)(tx.GatewayFee()), rpcTx.GatewayFee)
	assert.Equal(t, tx.GatewayFeeRecipient(), rpcTx.GatewayFeeRecipient)
}

// Test_isCelo1Block tests isCelo1Block function to determine whether the given block time
// corresponds to Celo1 chain based on the provided chain configuration
func Test_isCelo1Block(t *testing.T) {
	cel2Time := uint64(1000)

	t.Run("Non-Celo", func(t *testing.T) {
		res := isCelo1Block(&params.ChainConfig{
			Cel2Time: nil,
		}, 1000)

		assert.False(t, res)
	})

	t.Run("Celo1", func(t *testing.T) {
		res := isCelo1Block(&params.ChainConfig{
			Cel2Time: &cel2Time,
		}, 500)

		assert.True(t, res)
	})

	t.Run("Celo2", func(t *testing.T) {
		res := isCelo1Block(&params.ChainConfig{
			Cel2Time: &cel2Time,
		}, 1000)

		assert.False(t, res)
	})
}

// TestRPCMarshalBlock_Celo1TotalDifficulty tests the RPCMarshalBlock function, specifically for totalDifficulty field
// It validates the result has `totalDifficulty` field only if it's Celo1 block
func TestRPCMarshalBlock_Celo1TotalDifficulty(t *testing.T) {
	t.Parallel()

	blockTime := uint64(1000)
	block := types.NewBlock(&types.Header{Number: big.NewInt(100), Time: blockTime}, &types.Body{Transactions: []*types.Transaction{}}, nil, blocktest.NewHasher(), types.DefaultBlockConfig)

	marshalBlock := func(t *testing.T, config *params.ChainConfig) map[string]interface{} {
		t.Helper()

		resp, err := RPCMarshalBlock(context.Background(), block, false, false, config, testBackend{})
		if err != nil {
			require.NoError(t, err)
		}

		return resp
	}

	t.Run("Non-Celo", func(t *testing.T) {
		config := *params.MainnetChainConfig

		res := marshalBlock(t, &config)

		assert.Equal(t, nil, res["totalDifficulty"])
	})

	t.Run("Celo1", func(t *testing.T) {
		expected := (*hexutil.Big)(new(big.Int).Add(block.Number(), common.Big1))

		cel2Time := blockTime + 500
		config := *params.MainnetChainConfig
		config.Cel2Time = &cel2Time

		res := marshalBlock(t, &config)

		assert.Equal(t, expected, res["totalDifficulty"])
	})

	t.Run("Celo2", func(t *testing.T) {
		cel2Time := blockTime - 500
		config := *params.MainnetChainConfig
		config.Cel2Time = &cel2Time

		res := marshalBlock(t, &config)

		assert.Equal(t, nil, res["totalDifficulty"])
	})
}

// TestCheckTxFee ensures CheckTxFee validates whether the product of GasPrice and Gas
// does not exceed a predefined cap. Additionally, if FeeCurrency
// is provided, it ensures that GasPrice is correctly converted to
// the native currency before validation
func TestCheckTxFee(t *testing.T) {
	t.Parallel()

	var (
		txFeeCap = float64(1.5)
		rates    = common.ExchangeRates{
			core.DevFeeCurrencyAddr: big.NewRat(2, 1),
		}
		config = allEnabledChainConfig()
	)

	backend := newCeloBackendMock(config)
	backend.SetExchangeRates(rates)
	backend.SetRPCTxFeeCap(txFeeCap)

	t.Run("should allow transaction if fee does not exceed 1.5 Ether", func(t *testing.T) {
		err := CheckTxFee(context.Background(), backend, big.NewInt(params.Ether), 1, nil) // 1 Ether
		assert.NoError(t, err)
	})

	t.Run("should reject transaction if fee exceeds 1.5 Ether", func(t *testing.T) {
		err := CheckTxFee(context.Background(), backend, big.NewInt(params.Ether), 2, nil) // 2 Ether
		expected := fmt.Sprintf("tx fee (%.2f ether) exceeds the configured cap (%.2f ether)", 2.0, 1.5)
		assert.ErrorContains(t, err, expected)
	})

	t.Run("should allow transaction if fee currency is given and converted fee does not exceed 1.5 Ether", func(t *testing.T) {
		err := CheckTxFee(context.Background(), backend, big.NewInt(params.Ether), 2, &core.DevFeeCurrencyAddr) // 1 Ether
		assert.NoError(t, err)
	})

	t.Run("should reject transaction if fee currency is given and converted fee exceeds 1.5 Ether", func(t *testing.T) {
		err := CheckTxFee(context.Background(), backend, big.NewInt(params.Ether), 4, &core.DevFeeCurrencyAddr) // 2 Ether
		expected := fmt.Sprintf("tx fee (%.2f ether) exceeds the configured cap (%.2f ether)", 2.0, 1.5)
		assert.ErrorContains(t, err, expected)
	})

	t.Run("should reject transaction if fee currency is not registered", func(t *testing.T) {
		err := CheckTxFee(context.Background(), backend, big.NewInt(params.Ether), 4, &core.DevFeeCurrencyAddr2)
		assert.ErrorIs(t, err, exchange.ErrUnregisteredFeeCurrency)
	})
}

// TestMarshalReceipt tests the marshalReceipt function, which serializes a receipt object into a JSON-RPC response.
// It focuses on ensuring that this function returns the proper schema based on the transaction type.
func Test_marshalReceipt(t *testing.T) {
	config := allEnabledChainConfig()
	cel2Time := uint64(1000)
	config.Cel2Time = &cel2Time

	var (
		blockHash                    = common.BytesToHash([]byte{0x1})
		blockNumber           uint64 = 500
		cumulativeGasUsed     uint64 = 500000
		gasUsed               uint64 = 400000
		txIndex               uint   = 0
		gasTipCap                    = big.NewInt(100)
		gasFeeCap                    = big.NewInt(500)
		gas                   uint64 = 10000000
		l1GasPrice                   = big.NewInt(50000)
		l1GasUsed                    = big.NewInt(20000)
		l1Fee                        = big.NewInt(0)
		l1FeeScalar                  = big.NewFloat(0)
		l1BlobBaseFee                = big.NewInt(300000)
		l1BaseFeeScalar       uint64 = 0
		L1BlobBaseFeeScalar   uint64 = 0
		blobGasUsed           uint64 = 5000000
		blobGasPrice                 = big.NewInt(300000)
		from                         = common.HexToAddress("0x0000000000000000000000000000000000000bbb")
		sourceHash                   = common.BytesToHash([]byte{0xaa})
		depositNonce          uint64 = 100
		depositReceiptVersion uint64 = 1
		feeCurrency                  = core.DevFeeCurrencyAddr
		gatewayFeeRecipient          = common.HexToAddress("0x0000000000000000000000000000000000000eee")
		gatewayFee                   = big.NewInt(100)
		logs                         = []*types.Log{
			{
				Address: common.BytesToAddress([]byte{0x33}),
				Topics:  []common.Hash{common.HexToHash("dead")},
				Data:    []byte{0x01, 0x02, 0x03},
			},
		}
	)

	signer := types.MakeSigner(config, new(big.Int).SetUint64(blockNumber), blockTime)
	key, err := crypto.GenerateKey()
	require.NoError(t, err)

	receipt := types.Receipt{
		Type:                  types.DynamicFeeTxType,
		CumulativeGasUsed:     cumulativeGasUsed,
		Status:                types.ReceiptStatusSuccessful,
		GasUsed:               gasUsed,
		Logs:                  logs,
		BlockHash:             blockHash,
		BlockNumber:           big.NewInt(int64(blockNumber)),
		TransactionIndex:      txIndex,
		L1GasPrice:            l1GasPrice,
		L1GasUsed:             l1GasUsed,
		L1Fee:                 l1Fee,
		FeeScalar:             l1FeeScalar,
		L1BlobBaseFee:         l1BlobBaseFee,
		L1BaseFeeScalar:       &l1BaseFeeScalar,
		L1BlobBaseFeeScalar:   &L1BlobBaseFeeScalar,
		BlobGasUsed:           blobGasUsed,
		BlobGasPrice:          blobGasPrice,
		DepositNonce:          &depositNonce,
		DepositReceiptVersion: &depositReceiptVersion,
	}
	receipt.Bloom = types.CreateBloom(types.Receipts{&receipt})

	t.Run("LegacyTx receipt", func(t *testing.T) {
		tx, err := types.SignTx(types.NewTx(&types.LegacyTx{
			Nonce:    nonce,
			GasPrice: gasPrice,
			Gas:      gasLimit,
			To:       &to,
			Value:    value,
			Data:     []byte{},
		}), signer, key)
		require.NoError(t, err)

		// Disable Optimism Config
		config := *config
		config.Optimism = nil

		receipt := receipt
		receipt.TxHash = tx.Hash()

		result := marshalReceipt(&receipt, blockHash, blockNumber, signer, tx, int(txIndex), &config)

		checkReceiptFields(t, result, &receipt, tx, signer, blockHash, blockNumber, txIndex, false)
	})

	t.Run("DynamicFeeTx receipt", func(t *testing.T) {
		tx, err := types.SignTx(types.NewTx(&types.DynamicFeeTx{
			Nonce:     nonce,
			GasTipCap: gasTipCap,
			GasFeeCap: gasFeeCap,
			Gas:       gasLimit,
			To:        &to,
			Value:     value,
			Data:      []byte{},
		}), signer, key)
		require.NoError(t, err)

		receipt := receipt
		receipt.TxHash = tx.Hash()

		result := marshalReceipt(&receipt, blockHash, blockNumber, signer, tx, int(txIndex), config)

		checkReceiptFields(t, result, &receipt, tx, signer, blockHash, blockNumber, txIndex, true)
	})

	t.Run("DepositTx receipt", func(t *testing.T) {
		tx := types.NewTx(&types.DepositTx{
			SourceHash:          sourceHash,
			From:                from,
			To:                  &to,
			Value:               value,
			Gas:                 gas,
			IsSystemTransaction: true,
			Data:                []byte{},
		})
		require.NoError(t, err)

		receipt := receipt
		receipt.TxHash = tx.Hash()

		result := marshalReceipt(&receipt, blockHash, blockNumber, signer, tx, int(txIndex), config)

		checkReceiptFields(t, result, &receipt, tx, signer, blockHash, blockNumber, txIndex, true)
	})

	t.Run("CeloDynamicFeeTxV1 receipt", func(t *testing.T) {
		tx, err := types.SignTx(types.NewTx(&types.CeloDynamicFeeTx{
			ChainID:             big.NewInt(params.CeloAlfajoresChainID),
			Nonce:               nonce,
			GasTipCap:           gasTipCap,
			GasFeeCap:           gasFeeCap,
			Gas:                 gas,
			FeeCurrency:         &feeCurrency,
			GatewayFeeRecipient: &gatewayFeeRecipient,
			GatewayFee:          gatewayFee,
			To:                  &to,
			Value:               value,
			Data:                []byte{},
		}), signer, key)
		require.NoError(t, err)

		receipt := receipt
		receipt.TxHash = tx.Hash()

		result := marshalReceipt(&receipt, blockHash, blockNumber, signer, tx, int(txIndex), config)

		checkReceiptFields(t, result, &receipt, tx, signer, blockHash, blockNumber, txIndex, true)
	})

	t.Run("CeloDynamicFeeTxV2 receipt", func(t *testing.T) {
		tx, err := types.SignTx(types.NewTx(&types.CeloDynamicFeeTxV2{
			ChainID:     big.NewInt(params.CeloAlfajoresChainID),
			Nonce:       nonce,
			GasTipCap:   gasTipCap,
			GasFeeCap:   gasFeeCap,
			Gas:         gas,
			FeeCurrency: &feeCurrency,
			To:          &to,
			Value:       value,
			Data:        []byte{},
		}), signer, key)
		require.NoError(t, err)

		receipt := receipt
		receipt.TxHash = tx.Hash()

		result := marshalReceipt(&receipt, blockHash, blockNumber, signer, tx, int(txIndex), config)

		checkReceiptFields(t, result, &receipt, tx, signer, blockHash, blockNumber, txIndex, true)
	})
}

// checkReceiptFields is a helper function to perform a detailed comparison and verification of the contents of a transaction receipt
// retrieved from an RPC endpoint
// it specifically checks for the presence and values of special fields that are exported based on the transaction type
func checkReceiptFields(
	t *testing.T,
	rpcReceipt map[string]interface{},
	receipt *types.Receipt,
	tx *types.Transaction,
	signer types.Signer,
	blockhash common.Hash,
	blockNumber uint64,
	txIndex uint,
	isCel2 bool,
) {
	t.Helper()

	from, err := types.Sender(signer, tx)
	require.NoError(t, err)

	assert.Equal(t, hexutil.Uint(tx.Type()), rpcReceipt["type"])
	assert.Equal(t, blockhash, rpcReceipt["blockHash"])
	assert.Equal(t, hexutil.Uint64(blockNumber), rpcReceipt["blockNumber"])
	assert.Equal(t, tx.Hash(), rpcReceipt["transactionHash"])
	assert.Equal(t, hexutil.Uint64(txIndex), rpcReceipt["transactionIndex"])
	assert.Equal(t, from, rpcReceipt["from"])
	assert.Equal(t, tx.To(), rpcReceipt["to"])
	assert.Equal(t, hexutil.Uint64(receipt.GasUsed), rpcReceipt["gasUsed"])
	assert.Equal(t, hexutil.Uint64(receipt.CumulativeGasUsed), rpcReceipt["cumulativeGasUsed"])
	assert.Equal(t, receipt.Logs, rpcReceipt["logs"])
	assert.Equal(t, receipt.Bloom, rpcReceipt["logsBloom"])
	assert.Equal(t, hexutil.Uint(receipt.Status), rpcReceipt["status"])
	assert.Equal(t, receipt.Logs, rpcReceipt["logs"])

	if receipt.ContractAddress != (common.Address{}) {
		assert.Equal(t, receipt.ContractAddress, rpcReceipt["contractAddress"])
	} else {
		assert.Nil(t, rpcReceipt["contractAddress"])
	}

	if isCel2 && !tx.IsDepositTx() {
		assert.Equal(t, (*hexutil.Big)(receipt.L1GasPrice), rpcReceipt["l1GasPrice"])
		assert.Equal(t, (*hexutil.Big)(receipt.L1GasUsed), rpcReceipt["l1GasUsed"])
		assert.Equal(t, (*hexutil.Big)(receipt.L1Fee), rpcReceipt["l1Fee"])
		assert.Equal(t, receipt.FeeScalar.String(), rpcReceipt["l1FeeScalar"])
		assert.Equal(t, (*hexutil.Big)(receipt.L1BlobBaseFee), rpcReceipt["l1BlobBaseFee"])
		assert.Equal(t, hexutil.Uint64(*receipt.L1BaseFeeScalar), rpcReceipt["l1BaseFeeScalar"])
		assert.Equal(t, hexutil.Uint64(*receipt.L1BlobBaseFeeScalar), rpcReceipt["l1BlobBaseFeeScalar"])
	} else {
		assert.Nil(t, rpcReceipt["l1GasPrice"])
		assert.Nil(t, rpcReceipt["l1GasUsed"])
		assert.Nil(t, rpcReceipt["l1Fee"])
		assert.Nil(t, rpcReceipt["l1FeeScalar"])
		assert.Nil(t, rpcReceipt["l1BlobBaseFee"])
		assert.Nil(t, rpcReceipt["l1BaseFeeScalar"])
		assert.Nil(t, rpcReceipt["l1BlobBaseFeeScalar"])
	}

	if isCel2 && tx.IsDepositTx() {
		assert.Equal(t, hexutil.Uint64(*receipt.DepositNonce), rpcReceipt["depositNonce"])
		assert.Equal(t, hexutil.Uint64(*receipt.DepositReceiptVersion), rpcReceipt["depositReceiptVersion"])
	} else {
		assert.Nil(t, rpcReceipt["depositNonce"])
		assert.Nil(t, rpcReceipt["depositReceiptVersion"])
	}

	if isCel2 && tx.Type() == types.CeloDynamicFeeTxV2Type {
		assert.Equal(t, (*hexutil.Big)(receipt.BaseFee), rpcReceipt["baseFee"])
	} else {
		assert.Nil(t, rpcReceipt["baseFee"])
	}
}
