package ethapi

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/internal/blocktest"
	"github.com/ethereum/go-ethereum/params"
	"github.com/status-im/keycard-go/hexutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// encodeGasPriceMinimumUpdatedEvent encodes the given gas price minimum value into 32 bytes event data
func encodeGasPriceMinimumUpdatedEvent(gasPriceMinimum *big.Int) []byte {
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
		name            string
		beforeDbBaseFee *big.Int // BaseFee to be stored in the database before the test
		afterDbBaseFee  *big.Int // BaseFee to be stored in the database after the test
		backendBaseFee  *big.Int
		header          *types.Header
		expected        *types.Header
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
			name:            "should return the header with baseFee and gasLimit retrieved from the database",
			beforeDbBaseFee: big.NewInt(10e4),
			afterDbBaseFee:  big.NewInt(10e4),
			header: &types.Header{
				Number: big.NewInt(10e3),
			},
			expected: &types.Header{
				Number:   big.NewInt(10e3),
				BaseFee:  big.NewInt(10e4),
				GasLimit: 1e7,
			},
		},
		{
			name:            "should return the header with baseFee and gasLimit retrieved from the backend",
			beforeDbBaseFee: nil,
			afterDbBaseFee:  big.NewInt(10e8),
			backendBaseFee:  big.NewInt(10e8),
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

			// set base fee into DB
			if test.beforeDbBaseFee != nil {
				err := rawdb.WritePreGingerbreadBlockBaseFee(backend.ChainDb(), headerHash, test.beforeDbBaseFee)
				require.NoError(t, err)
			}

			// set block & receipts for base fee
			if test.backendBaseFee != nil {
				prevHeader := &types.Header{
					Number: new(big.Int).Sub(test.header.Number, big.NewInt(1)),
				}
				prevBlock := types.NewBlock(
					prevHeader,
					nil,
					nil,
					hasher,
					types.DefaultBlockConfig,
				)
				backend.setBlock(prevBlock.Number().Int64(), prevBlock)
				backend.setReceipts(prevBlock.Hash(), types.Receipts{
					{
						Logs: []*types.Log{
							{
								Topics: []common.Hash{
									gasPriceMinimumABI.Events["GasPriceMinimumUpdated"].ID,
								},
								Data: encodeGasPriceMinimumUpdatedEvent(test.backendBaseFee),
							},
						},
					},
				})
			}

			// retrieve baseFee and gasLimit
			newHeader := PopulatePreGingerbreadHeaderFields(context.Background(), backend, test.header)
			assert.Equal(t, test.expected, newHeader)

			// check db data after the test
			dbData, err := rawdb.ReadPreGingerbreadBlockBaseFee(backend.ChainDb(), headerHash)
			if test.afterDbBaseFee != nil {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, "error retrieving pre gingerbread base fee for block")
			}
			assert.Equal(t, test.afterDbBaseFee, dbData)
		})
	}
}

// Test_retrievePreGingerbreadGasLimit checks the gas limit retrieval for pre-gingerbread blocks
func Test_retrievePreGingerbreadGasLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		chainId  uint64
		height   *big.Int
		expected *big.Int
	}{
		{
			name:     "should return latest gas limit value for Celo Mainnet",
			chainId:  params.CeloMainnetChainID,
			height:   big.NewInt(21355415),
			expected: big.NewInt(32e6),
		},
		{
			name:     "should return nil if chainId is unknown",
			chainId:  12345,
			height:   big.NewInt(10),
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gasLimit := retrievePreGingerbreadGasLimit(test.chainId, test.height)

			assert.Equal(t, test.expected, gasLimit)
		})
	}
}

// Test_retrievePreGingerbreadBlockBaseFee tests the base fee retrieval for pre-gingerbread blocks
func Test_retrievePreGingerbreadBlockBaseFee(t *testing.T) {
	t.Parallel()

	prevHeader := &types.Header{Number: big.NewInt(999)}
	hasher := blocktest.NewHasher()

	// encode GasPriceMinimumUpdated event body
	baseFee := big.NewInt(1_000_000)
	baseFeeEventData := encodeGasPriceMinimumUpdatedEvent(baseFee)

	tests := []struct {
		name     string
		blocks   map[int64]*types.Block
		receipts types.Receipts
		height   *big.Int
		expected *big.Int
		err      error
	}{
		{
			name:     "should return an error if previous block is not found",
			blocks:   nil,
			receipts: nil,
			height:   big.NewInt(1000),
			expected: nil,
			err:      fmt.Errorf("block #999 not found"),
		},
		{
			name: "should return an error if block receipt is empty",
			blocks: map[int64]*types.Block{
				999: types.NewBlock(
					prevHeader,
					nil,
					nil,
					hasher,
					types.DefaultBlockConfig,
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
					prevHeader,
					&types.Body{
						Transactions: []*types.Transaction{
							types.NewTx(&types.LegacyTx{
								Nonce: 0,
							}),
						},
					},
					nil,
					hasher,
					types.DefaultBlockConfig,
				),
			},
			receipts: types.Receipts{
				{
					TxHash: prevHeader.Hash(),
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
					prevHeader,
					&types.Body{
						Transactions: []*types.Transaction{
							types.NewTx(&types.LegacyTx{
								Nonce: 0,
							}),
						},
					},
					nil,
					hasher,
					types.DefaultBlockConfig,
				),
			},
			receipts: types.Receipts{
				{
					TxHash: prevHeader.Hash(),
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
			err:      fmt.Errorf("an event GasPriceMinimumUpdated is not included in receipts of block #999"),
		},
		{
			name: "should return base fee successfully",
			blocks: map[int64]*types.Block{
				999: types.NewBlock(
					prevHeader,
					&types.Body{
						Transactions: []*types.Transaction{
							types.NewTx(&types.LegacyTx{
								Nonce: 0,
							}),
						},
					},
					nil,
					hasher,
					types.DefaultBlockConfig,
				),
			},
			receipts: types.Receipts{
				{
					TxHash: prevHeader.Hash(),
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

			if test.err == nil {
				require.NoError(t, err)
			} else {
				assert.EqualError(t, err, test.err.Error())
			}

			assert.Equal(t, test.expected, baseFee)
		})
	}
}

// Test_parseGasPriceMinimumUpdated checks the gas price minimum updated event parsing
func Test_parseGasPriceMinimumUpdated(t *testing.T) {
	t.Parallel()

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

			if test.err == nil {
				require.NoError(t, err)
			} else {
				assert.EqualError(t, err, test.err.Error())
			}

			assert.Equal(t, test.result, result)
		})
	}
}
