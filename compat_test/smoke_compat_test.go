//go:build compat_test

package compat_tests

import (
	"context"
	"flag"
	"math/rand"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

var (
	// CLI arguments
	blockInterval         uint64
	enableRandomBlockTest bool

	clientOpts = []rpc.ClientOption{
		rpc.WithWebsocketMessageSizeLimit(1024 * 1024 * 256),
	}
)

func init() {
	flag.Uint64Var(&blockInterval, "block-interval", 1_000_000, "number of blocks to skip between test iterations")
	flag.BoolVar(&enableRandomBlockTest, "random-block", false, "enable random block height within the range during test iterations")
}

func TestSmokeRPCCompatibilities(t *testing.T) {
	// TestSmokeRPCCompatibilities checks the compatibility of Celo L1 and Celo L2 RPC responses.
	// The test retrieves the following types of data at the beginning, the last L1 block, and at every one million blocks to verify compatibility.
	// 1. Block
	// 2. Transactions
	// 3. Transaction Receipts
	// 4. Block Receipts
	flag.Parse()

	if celoRpcURL == "" {
		t.Fatal("celo rpc url not set example usage:\n go test -v ./compat_test -tags compat_test -run TestSmokeRPCCompatibilities -celo-url ws://localhost:8546 -op-geth-url ws://localhost:9994")
	}
	if opGethRpcURL == "" {
		t.Fatal("op-geth rpc url not set example usage:\n go test -v ./compat_test -tags compat_test -run TestSmokeRPCCompatibilities -celo-url ws://localhost:8546 -op-geth-url ws://localhost:9994")
	}

	// Setup RPC clients for Celo L1 and Celo L2 server
	celoClient, err := rpc.DialOptions(context.Background(), celoRpcURL, clientOpts...)
	require.NoError(t, err)
	celoEthClient := ethclient.NewClient(celoClient)

	opClient, err := rpc.DialOptions(context.Background(), opGethRpcURL, clientOpts...)
	require.NoError(t, err)
	opEthClient := ethclient.NewClient(opClient)

	initCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
	t.Cleanup(cancel)

	// Fetch Chain IDs and make sure they're same
	celoChainId, err := celoEthClient.ChainID(initCtx)
	require.NoError(t, err)
	opChainId, err := opEthClient.ChainID(initCtx)
	require.NoError(t, err)
	require.Equal(t, celoChainId.Uint64(), opChainId.Uint64(), "chain ids of referenced chains differ")

	// Make sure the Chain ID is supported
	chainId := celoChainId.Uint64()
	_, ok := gingerbreadBlocks[chainId]
	require.True(t, ok, "chain id %d not found in supported chainIDs %v", chainId, gingerbreadBlocks)

	// Fetch the last block height in Celo L1
	lastCeloL1BlockHeight, err := celoEthClient.BlockNumber(initCtx)
	require.NoError(t, err)

	t.Logf("Last Celo L1 Block Height: %d", lastCeloL1BlockHeight)

	resultChan := make(chan *blockResults, 100)
	globalCtx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// Fetching RPC data job
	fetchingEg, jobCtx := errgroup.WithContext(globalCtx)
	fetchingEg.SetLimit(6)

	fetchingEg.Go(func() error {
		clients := &clients{
			celoEthclient: celoEthClient,
			opEthclient:   opEthClient,
			celoClient:    celoClient,
			opClient:      opClient,
		}

		// Fetch some random blocks between 0 to lastCeloL1BlockHeight
		for currentBlockNumber := uint64(0); currentBlockNumber < lastCeloL1BlockHeight; currentBlockNumber += blockInterval {
			// exit loop if context is canceled
			if isContextCanceled(jobCtx) {
				t.Logf("Context canceled, exiting fetching loop at height %d", currentBlockNumber)
				return nil
			}

			// decide block number to fetch between [currentBlockNumber, min(currentBlockNumber+blockInterval-1, lastCeloL1BlockHeight))
			offset := uint64(0)
			if enableRandomBlockTest {
				randomUpperBound := blockInterval
				if currentBlockNumber+randomUpperBound-1 >= lastCeloL1BlockHeight {
					randomUpperBound = lastCeloL1BlockHeight - currentBlockNumber
				}

				// Int63n returns a non-negative random number
				offset = uint64(rand.Int63n(int64(randomUpperBound)))
			}

			// Fetch data at the point of current range
			asyncFetchRpcData(t, jobCtx, fetchingEg, currentBlockNumber+offset, clients, resultChan)
		}

		// Fetch data at the last block
		asyncFetchRpcData(t, jobCtx, fetchingEg, lastCeloL1BlockHeight, clients, resultChan)

		return nil
	})

	// Testing fetched data job
	testingEg, jobCtx := errgroup.WithContext(globalCtx)
	testingEg.Go(func() error {
		for result := range resultChan {
			if isContextCanceled(jobCtx) {
				t.Logf("Context canceled, exiting testing loop at height %d", result.blockNumber)
				return nil
			}

			result := result
			testingEg.Go(func() error {
				err := result.Verify(chainId)
				if err != nil {
					t.Errorf("data at height %d err: %v\n", result.blockNumber, err)

					return err
				}

				t.Logf("Verified data at height %d", result.blockNumber)

				return nil
			})
		}

		return nil
	})

	// Wait for fetching and testing jobs to finish
	fetchingEg.Wait()
	close(resultChan) // close resultChan to signal testingEg to end
	testingEg.Wait()
}

func isContextCanceled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func asyncFetchRpcData(
	t *testing.T,
	ctx context.Context,
	eg *errgroup.Group,
	blockNumber uint64,
	clients *clients,
	resultChan chan *blockResults,
) {
	t.Helper()

	eg.Go(func() error {
		err := fetchBlockElements(ctx, clients, blockNumber, resultChan)
		if err != nil {
			t.Errorf("failed to fetch block elements at height %d: %v", blockNumber, err)

			return err
		}

		t.Logf("Fetched data at height %d", blockNumber)

		return nil
	})
}
