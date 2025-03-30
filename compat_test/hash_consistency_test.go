//go:build compat_test

package compat_tests

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie"
	"golang.org/x/sync/errgroup"
	"math/big"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"
)

var (
	sha3Uncles = common.HexToHash("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347")
)

func TestHashConsistency(t *testing.T) {
	// TestHashCompatTest validates that the hashes obtained from RPC endpoint are consistent with the values calculated locally
	// It retrieves and compares the following hashes:
	// 1. TxHash
	// 2. TxRoot
	// 3. BlockHash
	// 4. ReceiptRoot
	// 5. WithdrawalsRoot
	flag.Parse()

	if opGethRpcURL == "" {
		t.Fatal("op-geth rpc url not set example usage:\n go test -v ./compat_test -tags compat_test -run TestHashCompatTest -op-geth-url ws://localhost:9994")
	}
	if blockInterval == 0 {
		t.Fatal("block interval must be positive integer:\n go test -v ./compat_test -tags compat_test -run TestHashCompatTest -op-geth-url ws://localhost:9994 -block-interval 1000000")
	}

	// Setup RPC clients for Celo2
	cel2Client, err := rpc.DialOptions(context.Background(), opGethRpcURL, clientOpts...)
	require.NoError(t, err)
	cel2EthClient := ethclient.NewClient(cel2Client)

	// Fetch Chain ID
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	chainId, err := cel2EthClient.ChainID(ctx)
	if err != nil {
		t.Fatalf("failed to get chain id: %v", err)
	}

	// Fetch chain config
	chainConfig, err := fetchChainConfig(t, int(chainId.Int64()))
	if err != nil {
		t.Fatalf("failed to fetch chain config: %v", err)
	}
	if chainConfig.Cel2Time == nil {
		t.Fatalf("Cel2Time is not set in chain config")
	}
	if chainConfig.ChainID.Cmp(chainId) != 0 {
		t.Fatalf("fetched chain ID (%s) and the chain ID in ChainConfig (%s) do not match", chainId.String(), chainConfig.ChainID.String())
	}

	// Fetch start and latest blocks
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	startBlockObj, err := cel2EthClient.BlockByNumber(ctx, big.NewInt(int64(startBlock)))
	if err != nil {
		t.Fatalf("failed to get the first block: %v", err)
	} else if startBlockObj == nil {
		t.Fatal("first block is nil")
	}
	if startBlockObj.Time() < *chainConfig.Cel2Time {
		t.Fatalf("start block time (%d) is less than Cel2Time (%d)", startBlockObj.Time(), *chainConfig.Cel2Time)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	latestBlock, err := cel2EthClient.BlockByNumber(ctx, nil)
	if err != nil {
		t.Fatalf("failed to get latest block: %v", err)
	} else if latestBlock == nil {
		t.Fatal("latest block is nil")
	}

	endBlock := latestBlock.NumberU64()

	t.Logf("Starting Hash Compatibility Test")
	t.Logf("\tChainID: %s", chainId.String())
	t.Logf("\tStart Block: %d", startBlock)
	t.Logf("\tEnd Block: %d", endBlock)
	t.Logf("\tBlock Count in Range: %d", endBlock-startBlock+1)
	t.Logf("\tBlock Interval: %d", blockInterval)
	t.Logf("\tEnable Random Block Test: %t\n", enableRandomBlockTest)

	outerCtx, outerCancel := context.WithCancel(context.Background())
	t.Cleanup(outerCancel)

	resultCh := make(chan *blockAndReceipts, 100)

	fetchingEg, jobCtx := errgroup.WithContext(outerCtx)
	fetchingEg.SetLimit(5)
	fetchingEg.Go(func() error {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))

		for height := startBlock; height <= endBlock; height += blockInterval {
			if isContextCanceled(jobCtx) {
				t.Logf("Context canceled, exiting fetching loop at height %d", height)
				return nil
			}

			fetchingEg.Go(func() error {
				var blockHeight uint64
				if enableRandomBlockTest {
					blockHeight = height + rng.Uint64()%blockInterval
					if blockHeight > endBlock {
						blockHeight = endBlock
					}
				} else {
					blockHeight = height
				}

				ctx, cancel := context.WithTimeout(jobCtx, 10*time.Second)
				defer cancel()

				res, err := fetchBlockAndReceipts(ctx, cel2Client, cel2EthClient, blockHeight)
				if err != nil {
					return err
				}

				t.Logf("Fetched data at height %d", res.height)

				select {
				case <-jobCtx.Done():
					return jobCtx.Err()
				case resultCh <- res:
				}

				return nil
			})
		}

		return nil
	})

	testingEg, jobCtx := errgroup.WithContext(outerCtx)
	testingEg.SetLimit(10)
	testingEg.Go(func() error {
		for result := range resultCh {
			if isContextCanceled(jobCtx) {
				t.Logf("Context canceled, exiting testing loop at height %d", result.header.Number.Uint64())
				return nil
			}

			testingEg.Go(func() error {
				err := result.Verify(t)
				if err != nil {
					return fmt.Errorf("failed to verify data at height %d: %w", result.header.Number.Uint64(), err)
				}

				t.Logf("Verified data at height %d", result.header.Number.Uint64())

				return nil
			})
		}

		return nil
	})

	// Wait for all fetching jobs to complete and then close the result channel
	if err := fetchingEg.Wait(); err != nil {
		t.Logf("failed to complete fetching block elements job: %v", err)
		outerCancel()
		t.Fail()
	}
	close(resultCh)

	// Wait for all testing jobs to complete
	if err := testingEg.Wait(); err != nil {
		t.Logf("failed to complete testing block elements job: %v", err)
		outerCancel()
		t.Fail()
	}
}

type blockAndReceipts struct {
	height             uint64
	rpcBlockHash       common.Hash
	rpcTxRoot          common.Hash
	rpcReceiptRoot     common.Hash
	rpcSha3Uncles      common.Hash
	rpcWithdrawalsRoot common.Hash
	rpcTxs             []interface{}
	header             *types.Header
	receipts           types.Receipts
	withdrawals        types.Withdrawals
}

// fetchBlockAndReceipts fetches the block and receipts for the given height from the RPC endpoint
func fetchBlockAndReceipts(ctx context.Context, cel2Client *rpc.Client, cel2EthClient *ethclient.Client, height uint64) (*blockAndReceipts, error) {
	var blockObj map[string]interface{}
	err := cel2Client.Call(&blockObj, "eth_getBlockByNumber", fmt.Sprintf("0x%x", height), true)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch block at height %d: %w", height, err)
	} else if blockObj == nil {
		return nil, fmt.Errorf("block #%d not found", height)
	}

	rawRpcBlockHash, ok := blockObj["hash"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to cast block hash from block at height %d", height)
	}
	rpcBlockHash := common.HexToHash(rawRpcBlockHash)

	rpcTxRoot, ok := blockObj["transactionsRoot"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to cast transactionsRoot from block at height %d", height)
	}

	rpcReceiptRoot, ok := blockObj["receiptsRoot"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to cast receiptsRoot from block at height %d", height)
	}

	rpcSha3Uncles, ok := blockObj["sha3Uncles"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to cast sha3Uncles from block at height %d", height)
	}

	rpcWithdrawalsRoot, ok := blockObj["withdrawalsRoot"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to cast withdrawalsRoot from block at height %d", height)
	}

	txsObj, ok := blockObj["transactions"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to cast transactions from block data at height %d", height)
	}

	block, err := cel2EthClient.BlockByHash(ctx, rpcBlockHash)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch block by hash %s: %w", rpcBlockHash, err)
	} else if block == nil {
		return nil, fmt.Errorf("block #%d not found", height)
	}

	receipts, err := cel2EthClient.BlockReceipts(ctx, rpc.BlockNumberOrHash{BlockHash: &rpcBlockHash})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch receipts for block %d: %v", height, err)
	} else if receipts == nil {
		return nil, fmt.Errorf("receipts for block #%d not found", height)
	}

	return &blockAndReceipts{
		height:             height,
		rpcBlockHash:       rpcBlockHash,
		rpcTxRoot:          common.HexToHash(rpcTxRoot),
		rpcReceiptRoot:     common.HexToHash(rpcReceiptRoot),
		rpcSha3Uncles:      common.HexToHash(rpcSha3Uncles),
		rpcWithdrawalsRoot: common.HexToHash(rpcWithdrawalsRoot),
		rpcTxs:             txsObj,
		header:             block.Header(),
		receipts:           receipts,
		withdrawals:        block.Withdrawals(),
	}, nil
}

// Verify compares the hashes obtained from RPC endpoint with the values calculated locally
func (b *blockAndReceipts) Verify(t *testing.T) error {
	t.Helper()

	failed := false

	// Block Hash Test
	calculatedBlockHash := b.header.Hash()
	if calculatedBlockHash != b.rpcBlockHash {
		t.Logf("Block Hash Mismatch, height=%d => calculated=%s, rpc=%s", b.height, calculatedBlockHash.Hex(), b.rpcBlockHash.Hex())
		failed = true
	}

	// Transactions Hash Test
	txs := make(types.Transactions, 0, len(b.rpcTxs))
	for idx, rawTxObj := range b.rpcTxs {
		txObj, ok := rawTxObj.(map[string]interface{})
		if !ok {
			return fmt.Errorf("failed to cast transaction from block data at height %d", b.height)
		}
		rpcHash := txObj["hash"].(string)

		jsonTx, err := json.Marshal(rawTxObj)
		if err != nil {
			return fmt.Errorf("failed to marshal tx: %v", err)
		}

		var tx types.Transaction
		if err := json.Unmarshal(jsonTx, &tx); err != nil {
			return fmt.Errorf("failed to unmarshal tx: %v", err)
		}
		txs = append(txs, &tx)

		calculatedHash := tx.Hash()
		if calculatedHash.Hex() != rpcHash {
			t.Logf("Tx Hash Mismatch, block height=%d, block hash=%s, tx index=%d, tx hash=%s, tx type=%s => calculated=%s, rpc=%s", b.height, b.rpcBlockHash, idx, rpcHash, getTxTypeName(tx.Type()), calculatedHash.Hex(), rpcHash)
			failed = true
		}
	}

	// Tx Root Test
	calculatedTxRoot := types.DeriveSha(txs, trie.NewStackTrie(nil))
	if calculatedTxRoot != b.rpcTxRoot {
		t.Logf("Tx Root Mismatch, block height=%d => calculated=%s, rpc=%s", b.height, calculatedTxRoot, b.rpcTxRoot)
		failed = true
	}

	// Block Receipt Root Test
	calculatedReceiptRoot := types.DeriveSha(b.receipts, trie.NewStackTrie(nil))
	if calculatedReceiptRoot != b.rpcReceiptRoot {
		t.Logf("Receipt Root Mismatch, block height=%d => calculated=%s, rpc=%s", b.height, calculatedReceiptRoot, b.rpcReceiptRoot)
		failed = true
	}

	// Withdrawal Root Test
	calculatedWithdrawalsRoot := types.DeriveSha(b.withdrawals, trie.NewStackTrie(nil))
	if calculatedWithdrawalsRoot != b.rpcWithdrawalsRoot {
		t.Logf("Withdrawal Root Mismatch, block height=%d => calculated=%s, rpc=%s", b.height, calculatedWithdrawalsRoot, b.rpcWithdrawalsRoot)
		failed = true
	}

	if b.rpcSha3Uncles != sha3Uncles {
		t.Logf("Sha3 Uncles Mismatch, block height=%d => calculated=%s, rpc=%s", b.height, sha3Uncles, b.rpcSha3Uncles)
		failed = true
	}

	if failed {
		return fmt.Errorf("block verification failed")
	}

	return nil
}

// getTxTypeName returns the name of the transaction type
func getTxTypeName(txType uint8) string {
	switch txType {
	case types.LegacyTxType:
		return "Legacy"
	case types.AccessListTxType:
		return "AccessList"
	case types.DynamicFeeTxType:
		return "DynamicFee"
	case types.BlobTxType:
		return "Blob"
	case types.SetCodeTxType:
		return "SetCode"
	case types.DepositTxType:
		return "Deposit"
	case types.CeloDynamicFeeTxType:
		return "CeloDynamicFee"
	case types.CeloDynamicFeeTxV2Type:
		return "CeloDynamicFeeV2"
	case types.CeloDenominatedTxType:
		return "CeloDenominated"
	default:
		return "Unknown"
	}
}

// fetchChainConfig fetches the chain config for the given chain ID from server
func fetchChainConfig(t *testing.T, chainId int) (cfg *params.ChainConfig, err error) {
	t.Helper()

	var url string
	switch chainId {
	case params.CeloMainnetChainID:
		url = "https://storage.googleapis.com/cel2-rollup-files/celo/genesis.json"
	case params.CeloAlfajoresChainID:
		url = "https://storage.googleapis.com/cel2-rollup-files/alfajores/genesis.json"
	case params.CeloBaklavaChainID:
		url = "https://storage.googleapis.com/cel2-rollup-files/baklava/genesis.json"
	default:
		return nil, fmt.Errorf("unsupported chain id: %d", chainId)
	}

	genesis := new(core.Genesis)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download genesis file: %v", err)
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			t.Logf("[WARN] failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code when fetching genesis file: %d", resp.StatusCode)
	}

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(genesis); err != nil {
		return nil, fmt.Errorf("failed to decode genesis file: %v", err)
	}

	return genesis.Config, nil
}
