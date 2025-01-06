package ethapi

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/internal/blocktest"
	"github.com/ethereum/go-ethereum/params"
	"math/big"
	"testing"

	"github.com/status-im/keycard-go/hexutils"
	"github.com/stretchr/testify/assert"
)

// encodeGasPriceMinimumUpdatedEventBody encodes the given gas price minimum value into 32 bytes event data
func encodeGasPriceMinimumUpdatedEventBody(gasPriceMinimum *big.Int) []byte {
	gasPriceMinimumBytes := gasPriceMinimum.Bytes()
	gasPriceMinimumEventData := make([]byte, 32)
	copy(gasPriceMinimumEventData[32-len(gasPriceMinimumBytes):], gasPriceMinimumBytes)
	return gasPriceMinimumEventData
}

// TestPopulatePreGingerbreadHeaderFields tests the PopulatePreGingerbreadHeaderFields function
func TestPopulatePreGingerbreadHeaderFields(t *testing.T) {
	t.Parallel()

	hasher := blocktest.NewHasher()
	gingerBreadBeginsAt := big.NewInt(10e5)

	tests := []struct {
		name           string
		beforeDbData   *rawdb.PreGingerbreadFields // data to be stored in the database before the test
		afterDbData    *rawdb.PreGingerbreadFields // data to be stored in the database after the test
		backendBaseFee *big.Int
		header         *types.Header
		expected       *types.Header
	}{
		{
			name: "should return the same header for post-gingerbread header",
			header: &types.Header{
				Number:   big.NewInt(10e5),
				BaseFee:  big.NewInt(10e2),
				GasLimit: 10e3,
			},
			expected: &types.Header{
				Number:   big.NewInt(10e5),
				BaseFee:  big.NewInt(10e2),
				GasLimit: 10e3,
			},
		},
		{
			name: "should return the header with baseFee and gasLimit retrieved from the database",
			beforeDbData: &rawdb.PreGingerbreadFields{
				BaseFee:  big.NewInt(10e4),
				GasLimit: big.NewInt(10e5),
			},
			afterDbData: &rawdb.PreGingerbreadFields{
				BaseFee:  big.NewInt(10e4),
				GasLimit: big.NewInt(10e5),
			},
			header: &types.Header{
				Number: big.NewInt(10e3),
			},
			expected: &types.Header{
				Number:   big.NewInt(10e3),
				BaseFee:  big.NewInt(10e4),
				GasLimit: 10e5,
			},
		},
		{
			name: "should return the header with only baseFee retrieved from the database",
			beforeDbData: &rawdb.PreGingerbreadFields{
				BaseFee: big.NewInt(10e6),
			},
			afterDbData: &rawdb.PreGingerbreadFields{
				BaseFee:  big.NewInt(10e6),
				GasLimit: big.NewInt(0),
			},
			header: &types.Header{
				Number: big.NewInt(10e3),
			},
			expected: &types.Header{
				Number:  big.NewInt(10e3),
				BaseFee: big.NewInt(10e6),
			},
		},
		{
			name: "should return the header with only gasLimit retrieved from the database",
			beforeDbData: &rawdb.PreGingerbreadFields{
				GasLimit: big.NewInt(10e7),
			},
			afterDbData: &rawdb.PreGingerbreadFields{
				BaseFee:  big.NewInt(0),
				GasLimit: big.NewInt(10e7),
			},
			header: &types.Header{
				Number: big.NewInt(10e3),
			},
			expected: &types.Header{
				Number:   big.NewInt(10e3),
				GasLimit: 10e7,
			},
		},
		{
			name:         "should return the header with baseFee and gasLimit retrieved from the backend",
			beforeDbData: nil,
			afterDbData: &rawdb.PreGingerbreadFields{
				BaseFee:  big.NewInt(10e8),
				GasLimit: big.NewInt(20e6),
			},
			backendBaseFee: big.NewInt(10e8),
			header: &types.Header{
				Number: big.NewInt(1000),
			},
			expected: &types.Header{
				Number:   big.NewInt(1000),
				BaseFee:  big.NewInt(10e8),
				GasLimit: 20e6,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			headerHash := test.header.Hash()
			backend := newCeloBackendMock(&params.ChainConfig{
				ChainID:          big.NewInt(params.CeloMainnetChainID),
				GingerbreadBlock: gingerBreadBeginsAt,
			})

			// set data into database and backend
			if test.beforeDbData != nil {
				err := rawdb.WritePreGingerbreadAdditionalFields(backend.ChainDb(), headerHash, test.beforeDbData)
				assert.NoError(t, err)
			}
			if test.backendBaseFee != nil {
				prevHeader := &types.Header{
					Number: new(big.Int).Sub(test.header.Number, big.NewInt(1)),
				}
				prevBlock := types.NewBlock(
					prevHeader,
					nil,
					nil,
					hasher,
				)
				backend.setBlock(prevBlock.Number().Int64(), prevBlock)
				backend.setReceipts(prevBlock.Hash(), types.Receipts{
					{
						Logs: []*types.Log{
							{
								Topics: []common.Hash{
									gasPriceMinimumABI.Events["GasPriceMinimumUpdated"].ID,
								},
								Data: encodeGasPriceMinimumUpdatedEventBody(test.backendBaseFee),
							},
						},
					},
				})
			}

			// retrieve baseFee and gasLimit
			newHeader := PopulatePreGingerbreadHeaderFields(context.Background(), backend, test.header)

			assert.Equal(t, test.expected, newHeader)

			// check db data after the test
			dbData, err := rawdb.ReadPreGingerbreadAdditionalFields(backend.ChainDb(), headerHash)
			assert.NoError(t, err)
			assert.Equal(t, test.afterDbData, dbData)
		})
	}
}

// Test_retrievePreGingerbreadGasLimit checks the gas limit retrieval for pre-gingerbread blocks
func Test_retrievePreGingerbreadGasLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		chainId  *big.Int
		height   *big.Int
		expected *big.Int
	}{
		{
			name:     "should return nil for undefined chain",
			chainId:  nil,
			height:   big.NewInt(10e5),
			expected: nil,
		},
		{
			name:     "should return latest gas limit value for celo mainnet",
			chainId:  big.NewInt(params.CeloMainnetChainID),
			height:   big.NewInt(21355415),
			expected: big.NewInt(32e6),
		},
		{
			name:     "should return latest gas limit value for celo alfajores",
			chainId:  big.NewInt(params.CeloAlfajoresChainID),
			height:   big.NewInt(11143973),
			expected: big.NewInt(35e6),
		},
		{
			name:     "should return latest gas limit value for celo baklava",
			chainId:  big.NewInt(params.CeloBaklavaChainID),
			height:   big.NewInt(15158971),
			expected: big.NewInt(20e6),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			backend := newCeloBackendMock(&params.ChainConfig{
				ChainID: test.chainId,
			})

			gasLimit := retrievePreGingerbreadGasLimit(backend, test.height)

			assert.Equal(t, test.expected, gasLimit)
		})
	}
}

// Test_retrievePreGingerbreadBlockBaseFee tests the base fee retrieval for pre-gingerbread blocks
func Test_retrievePreGingerbreadBlockBaseFee(t *testing.T) {
	t.Parallel()

	header := &types.Header{Number: big.NewInt(999)}
	hasher := blocktest.NewHasher()

	// encode GasPriceMinimumUpdated event body
	baseFee := big.NewInt(1_000_000)
	baseFeeEventData := encodeGasPriceMinimumUpdatedEventBody(baseFee)

	tests := []struct {
		name     string
		blocks   map[int64]*types.Block
		receipts types.Receipts
		height   *big.Int
		expected *big.Int
		err      error
	}{
		{
			name:     "should return an error if block is not found",
			blocks:   nil,
			receipts: nil,
			height:   big.NewInt(1000),
			expected: nil,
			err:      fmt.Errorf("block #1000 not found"),
		},
		{
			name: "should return an error if block receipt is empty",
			blocks: map[int64]*types.Block{
				999: types.NewBlock(
					header,
					nil,
					nil,
					hasher,
				),
			},
			receipts: nil,
			height:   big.NewInt(1000),
			expected: nil,
			err:      fmt.Errorf("receipts of block #999 don't contain system logs"),
		},
		{
			name: "should return an error if block receipt doesn't contain system logs",
			blocks: map[int64]*types.Block{
				999: types.NewBlock(
					header,
					&types.Body{
						Transactions: []*types.Transaction{
							types.NewTx(&types.LegacyTx{
								Nonce: 0,
							}),
						},
					},
					nil,
					hasher,
				),
			},
			receipts: types.Receipts{
				{
					TxHash: header.Hash(),
					Logs:   nil,
				},
			},
			height:   big.NewInt(1000),
			expected: nil,
			err:      fmt.Errorf("receipts of block #999 don't contain system logs"),
		},
		{
			name: "should return an error if block receipt doesn't contain GasPriceMinimumUpdated event in system logs",
			blocks: map[int64]*types.Block{
				999: types.NewBlock(
					header,
					&types.Body{
						Transactions: []*types.Transaction{
							types.NewTx(&types.LegacyTx{
								Nonce: 0,
							}),
						},
					},
					nil,
					hasher,
				),
			},
			receipts: types.Receipts{
				{
					TxHash: header.Hash(),
					Logs:   nil,
				},
				{
					Logs: []*types.Log{
						{
							Topics: []common.Hash{
								common.HexToHash("0x123456"), // fake topic
							},
							Data: baseFeeEventData,
						},
					},
				},
			},
			height:   big.NewInt(1000),
			expected: nil,
			err:      fmt.Errorf("gas price minimum updated event is not included in a receipt of block #999"),
		},
		{
			name: "should return an error if block receipt doesn't contain GasPriceMinimumUpdated event in system logs",
			blocks: map[int64]*types.Block{
				999: types.NewBlock(
					header,
					&types.Body{
						Transactions: []*types.Transaction{
							types.NewTx(&types.LegacyTx{
								Nonce: 0,
							}),
						},
					},
					nil,
					hasher,
				),
			},
			receipts: types.Receipts{
				{
					TxHash: header.Hash(),
					Logs:   nil,
				},
				{
					Logs: []*types.Log{
						{
							Topics: []common.Hash{
								gasPriceMinimumABI.Events["GasPriceMinimumUpdated"].ID,
							},
							Data: baseFeeEventData,
						},
					},
				},
			},
			height:   big.NewInt(1000),
			expected: baseFee,
			err:      nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create a new backend mock with seed data
			backend := newCeloBackendMock(&params.ChainConfig{})
			for number, block := range test.blocks {
				backend.setBlock(number, block)
				backend.setReceipts(block.Hash(), test.receipts)
			}

			baseFee, err := retrievePreGingerbreadBlockBaseFee(context.Background(), backend, test.height)

			assert.Equal(t, test.expected, baseFee)
			if test.err == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, test.err.Error())
			}
		})
	}
}

// Test_parseGasPriceMinimumUpdated checks the gas price minimum updated event parsing
func Test_parseGasPriceMinimumUpdated(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		result *big.Int
		err    error
	}{
		{
			name:   "should parse gas price successfully",
			data:   hexutils.HexToBytes("00000000000000000000000000000000000000000000000000000000000f4240"),
			result: big.NewInt(1_000_000),
			err:    nil,
		},
		{
			name:   "should return error if data is not in the expected format",
			data:   hexutils.HexToBytes("123456"),
			result: nil,
			err:    errors.New("abi: cannot marshal in to go type: length insufficient 3 require 32"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := parseGasPriceMinimumUpdated(test.data)
			assert.Equal(t, test.result, result)

			if test.err == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, test.err.Error())
			}
		})
	}
}
