//go:build compat_test

package compat_tests

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math/big"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

// Note! That although it would be great to be able to compare the output of the
// respective celo and op-geth ethclient.Clients, unfortunately we are not able
// to pull in the celo client as a dependency since the cgo compilation results
// in duplicate symbol errors in the secp256k1 c code. E.g.:
//
// duplicate symbol '_secp256k1_ec_pubkey_tweak_mul' in: ...
//
// There are 33 such instances and it doesn't seem trivial to resolve. So we
// content ourselves with just calling the rpc endpoints of the celo node
// using the op-geth rpc client.

var (
	celoRpcURL        string
	opGethRpcURL      string
	gingerbreadBlocks = map[uint64]uint64{42220: 21616000, 62320: 18785000, 44787: 19814000}
)

func init() {
	// Define your custom flag
	flag.StringVar(&celoRpcURL, "celo-url", "", "celo rpc url")
	flag.StringVar(&opGethRpcURL, "op-geth-url", "", "op-geth rpc url")
}

type clients struct {
	celoEthclient, opEthclient *ethclient.Client
	celoClient, opClient       *rpc.Client
}

// Connects to a celo-blockchain and an op-geth node over rpc, selects the lowest head block and iterates from 0 to the
// lowest head block comparing all blocks, transactions receipts and logs.
//
// The test requires two flags to be set that provide the rpc urls to use, and the test is segregated from normal
// execution via a build tag. So to run it you would do:
//
// go test -v ./compat_test -tags compat_test -celo-url <celo rpc url> -op-geth-url <op-geth rpc url>
func TestCompatibilityOfChains(t *testing.T) {
	flag.Parse()

	if celoRpcURL == "" {
		t.Fatal("celo rpc url not set example usage:\n go test -v ./compat_test -tags compat_test -celo-url ws://localhost:9546 -op-geth-url ws://localhost:8546")
	}
	if opGethRpcURL == "" {
		t.Fatal("op-geth rpc url not set example usage:\n go test -v ./compat_test -tags compat_test -celo-url ws://localhost:9546 -op-geth-url ws://localhost:8546")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	clientOpts := []rpc.ClientOption{rpc.WithWebsocketMessageSizeLimit(1024 * 1024 * 256)}

	celoClient, err := rpc.DialOptions(context.Background(), celoRpcURL, clientOpts...)
	require.NoError(t, err)
	celoEthClient := ethclient.NewClient(celoClient)

	opClient, err := rpc.DialOptions(context.Background(), opGethRpcURL, clientOpts...)
	require.NoError(t, err)
	opEthClient := ethclient.NewClient(opClient)

	clients := &clients{
		celoEthclient: celoEthClient,
		opEthclient:   opEthClient,
		celoClient:    celoClient,
		opClient:      opClient,
	}

	celoChainID, err := celoEthClient.ChainID(ctx)
	require.NoError(t, err)
	opChainID, err := opEthClient.ChainID(ctx)
	require.NoError(t, err)
	require.Equal(t, celoChainID.Uint64(), opChainID.Uint64(), "chain ids of referenced chains differ")

	_, ok := gingerbreadBlocks[celoChainID.Uint64()]
	require.True(t, ok, "chain id %d not found in supported chainIDs %v", celoChainID.Uint64(), gingerbreadBlocks)

	latestCeloBlock, err := celoEthClient.BlockNumber(ctx)
	require.NoError(t, err)

	latestOpBlock, err := opEthClient.BlockNumber(ctx)
	require.NoError(t, err)

	// We take the lowest of the two blocks
	latestBlock := latestCeloBlock
	if latestOpBlock < latestCeloBlock {
		latestBlock = latestOpBlock
	}
	// We subtract 128 from the latest block to avoid handlig blocks where state
	// is present in celo since when state is present baseFeePerGas is set on
	// the celo block with a value, and we can't access that state from the
	// op-geth side
	endBlock := latestBlock - 128

	startBlock := uint64(0)
	batches := make(map[uint64]*batch)
	fmt.Printf("start block: %v, end block: %v\n", startBlock, endBlock)
	start := time.Now()
	prev := start
	var batchSize uint64 = 1000
	var count uint64
	resultChan := make(chan *blockResults, 100)

	longCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, longCtx := errgroup.WithContext(longCtx)
	g.SetLimit(5)

	g.Go(func() error {
		for i := startBlock; i <= endBlock; i++ {
			index := i
			g.Go(func() error {
				err := fetchBlockElements(clients, index, resultChan)
				if err != nil {
					fmt.Printf("block %d err: %v\n", index, err)
				}
				return err
			})
			select {
			case <-longCtx.Done():
				return longCtx.Err()
			default:
			}
		}
		return nil
	})

	g.Go(func() error {
		// Set up contiguous block tracking
		latestContiguousBlock := startBlock
		if startBlock > 0 {
			latestContiguousBlock = startBlock - 1
		}
		receivedNumbers := make(map[uint64]bool)
		for {
			select {
			case <-longCtx.Done():
				return longCtx.Err()
			case blockResult := <-resultChan:

				err := blockResult.Verify(celoChainID.Uint64())
				if err != nil {
					return fmt.Errorf("block verification failed: %w, failureBlock: %d ,latestContiguousBlock: %d", err, blockResult.blockNumber, latestContiguousBlock)
				}

				receivedNumbers[blockResult.blockNumber] = true
				for receivedNumbers[latestContiguousBlock+1] {
					delete(receivedNumbers, latestContiguousBlock)
					latestContiguousBlock++
				}

				// get the batch
				batchIndex := blockResult.blockNumber / batchSize
				b, ok := batches[batchIndex]
				if !ok {
					batchStart := batchIndex * batchSize
					b = newBatch(batchStart, batchStart+batchSize, clients.celoEthclient, clients.opEthclient)
					batches[batchIndex] = b
				}
				done, err := b.Process(blockResult)
				if err != nil {
					err := fmt.Errorf("batch %d procesing failed: %w, latestContiguousBlock %d", batchIndex, err, latestContiguousBlock)
					fmt.Printf("%v\n", err)
					return err
				}
				if done {
					delete(batches, batchIndex)
				}

				count++
				if count%batchSize == 0 {
					fmt.Printf("buffered objects %d\t\t current goroutines %v\t\tblocks %d\t\tlatestContiguous %d\t\telapsed: %v\ttotal: %v\n", len(resultChan), runtime.NumGoroutine(), count, latestContiguousBlock, time.Since(prev), time.Since(start))
					prev = time.Now()
				}
				// Check to see if we have processed all blocks
				if count == endBlock-startBlock+1 {
					fmt.Printf("buffered objects %d\t\t current goroutines %v\t\tblocks %d\t\tlatestContiguous %d\t\telapsed: %v\ttotal: %v\n", len(resultChan), runtime.NumGoroutine(), count, latestContiguousBlock, time.Since(prev), time.Since(start))
					return nil
				}
			}
		}
	})

	require.NoError(t, g.Wait())
}

type batch struct {
	start, end             uint64
	remaining              uint64
	incrementalLogs        [][]*types.Log
	celoClient, opClient   *ethclient.Client
	celoRawLogs, opRawLogs []json.RawMessage
	celoLogs, opLogs       []*types.Log
	logFetchErrGroup       *errgroup.Group
	logFetchContext        context.Context
	logFetchContextCancel  context.CancelFunc
}

func newBatch(start, end uint64, celoClient, opClient *ethclient.Client) *batch {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	g, ctx := errgroup.WithContext(ctx)
	b := &batch{
		start:                 start,
		end:                   end,
		remaining:             end - start,
		incrementalLogs:       make([][]*types.Log, end-start),
		celoClient:            celoClient,
		opClient:              opClient,
		logFetchErrGroup:      g,
		logFetchContext:       ctx,
		logFetchContextCancel: cancel,
	}
	b.fetchLogs()
	return b
}

func (b *batch) fetchLogs() {
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(b.start)),
		ToBlock:   big.NewInt(int64(b.end - 1)),
	}
	rawQuery := filterQuery{
		FromBlock: hexutil.Uint64(b.start),
		ToBlock:   hexutil.Uint64(b.end - 1),
	}
	b.logFetchErrGroup.Go(func() error {
		logs, err := b.celoClient.FilterLogs(b.logFetchContext, query)
		if err != nil {
			return err
		}
		logPointers := make([]*types.Log, len(logs))
		for i, log := range logs {
			logCopy := log
			logPointers[i] = &logCopy
		}
		b.celoLogs = logPointers
		return nil
	})
	b.logFetchErrGroup.Go(func() error {
		return rpcCall(b.logFetchContext, b.celoClient.Client(), &b.celoRawLogs, "eth_getLogs", rawQuery)
	})
	b.logFetchErrGroup.Go(func() error {
		logs, err := b.opClient.FilterLogs(b.logFetchContext, query)
		if err != nil {
			return err
		}
		logPointers := make([]*types.Log, len(logs))
		for i, log := range logs {
			logCopy := log
			logPointers[i] = &logCopy
		}
		b.opLogs = logPointers
		return nil
	})
	b.logFetchErrGroup.Go(func() error {
		return rpcCall(b.logFetchContext, b.opClient.Client(), &b.opRawLogs, "eth_getLogs", rawQuery)
	})
}

func (b *batch) Process(results *blockResults) (done bool, err error) {
	logs, err := results.Logs()
	if err != nil {
		return false, err
	}
	b.incrementalLogs[results.blockNumber-b.start] = logs
	b.remaining--
	if b.remaining == 0 {
		// copy all incremental logs into one slice of type []*types.Log
		allLogs := make([]*types.Log, 0)
		for _, logs := range b.incrementalLogs {
			allLogs = append(allLogs, logs...)
		}

		defer b.logFetchContextCancel()
		err = b.logFetchErrGroup.Wait()
		if err != nil {
			return false, err
		}

		err = EqualObjects(b.celoLogs, b.opLogs)
		if err != nil {
			return false, err
		}
		err = EqualObjects(b.celoRawLogs, b.opRawLogs)
		if err != nil {
			return false, err
		}

		// crosscheck the logs
		err = EqualObjects(len(b.celoLogs), len(allLogs))
		if err != nil {
			return false, err
		}
		err = EqualObjects(b.celoLogs, allLogs)
		if err != nil {
			return false, err
		}

		var unmarshaledCeloRawLogs []*types.Log
		err = jsonConvert(b.celoRawLogs, &unmarshaledCeloRawLogs)
		if err != nil {
			return false, err
		}

		err = EqualObjects(b.celoLogs, unmarshaledCeloRawLogs)
		if err != nil {
			return false, err
		}
	}
	return b.remaining == 0, nil
}

// Holds results retreived from the celo and op-geth clients rpc api, fields named raw were retrieved via a direcet rpc
// call whereas fields without raw were retrieved via the ethclient.
type blockResults struct {
	blockNumber uint64
	// danglingState bool

	// Blocks
	opBlockByNumber      *types.Block
	opBlockByHash        *types.Block
	celoRawBlockByNumber map[string]interface{}
	opRawBlockByNumber   map[string]interface{}
	celoRawBlockByHash   map[string]interface{}
	opRawBlockByHash     map[string]interface{}

	// Transactions
	celoTxs    []*types.Transaction
	opTxs      []*types.Transaction
	celoRawTxs []map[string]interface{}
	opRawTxs   []map[string]interface{}

	// Receipts
	celoReceipts    []*types.Receipt
	opReceipts      []*types.Receipt
	celoRawReceipts []map[string]interface{}
	opRawReceipts   []map[string]interface{}

	// BlockReceipts
	celoBlockReceipts    []*types.Receipt
	opBlockReceipts      []*types.Receipt
	celoRawBlockReceipts []map[string]interface{}
	opRawBlockReceipts   []map[string]interface{}

	// Block receipt (special receipt added by celo to capture system operations)
	celoRawBlockReceipt map[string]interface{}
	opRawBlockReceipt   map[string]interface{}
}

func (r *blockResults) verifyBlocks(chainID uint64) error {
	// Check block pairs
	makeBlockComparable(r.opBlockByNumber)
	makeBlockComparable(r.opBlockByHash)
	// Optimism blocks via ethclient
	err := EqualObjects(r.opBlockByNumber, r.opBlockByHash)
	if err != nil {
		return err
	}

	// Raw blocks by number
	err = filterCeloBlock(r.blockNumber, r.celoRawBlockByNumber, gingerbreadBlocks[chainID])
	if err != nil {
		return err
	}
	err = filterOpBlock(r.opRawBlockByNumber)
	if err != nil {
		return err
	}
	err = EqualObjects(r.celoRawBlockByNumber, r.opRawBlockByNumber)
	if err != nil {
		return err
	}

	// Raw blocks by hash
	err = filterCeloBlock(r.blockNumber, r.celoRawBlockByHash, gingerbreadBlocks[chainID])
	if err != nil {
		return err
	}
	err = filterOpBlock(r.opRawBlockByHash)
	if err != nil {
		return err
	}
	err = EqualObjects(r.celoRawBlockByHash, r.opRawBlockByHash)
	if err != nil {
		return err
	}

	// Cross check
	err = EqualObjects(r.celoRawBlockByNumber, r.celoRawBlockByHash)
	if err != nil {
		return err
	}

	// We can't easily convert blocks from the ethclient to a map[string]interface{} since they lack hydrated fields and
	// also due to the json conversion end up with null for unset fields (as opposed to just not having the field). So
	// instead we compare the hashes.
	err = EqualObjects(r.celoRawBlockByNumber["hash"].(string), r.opBlockByNumber.Hash().String())
	if err != nil {
		return err
	}
	celoRawTxs := r.celoRawBlockByNumber["transactions"].([]interface{})
	err = EqualObjects(len(celoRawTxs), len(r.opBlockByNumber.Transactions()))
	if err != nil {
		return err
	}
	for i := range celoRawTxs {
		celoTx := celoRawTxs[i].(map[string]interface{})
		err = EqualObjects(celoTx["hash"].(string), r.opBlockByNumber.Transactions()[i].Hash().String())
		if err != nil {
			return err
		}
	}
	return nil
}

func filterOpBlock(block map[string]interface{}) error {
	// We remove the following fields:
	// size: being the size of the rlp encoded block it differs between the two systems since the block structure is different
	// chainId (on transactions): celo didn't return chainId on legacy transactions, it did for other types but its a bit more involved to filter that per tx type.
	//
	// If the gas limit is zero (I.E. pre gingerbread) then we also remove the following fields:
	// gasLimit: since on the celo side we hardcoded pre-gingerbread gas limits but on the op-geth side we just have zero.
	// uncles, sha3Uncles, mixHash, nonce: since these are not present in the pre-gingerbread celo block.
	delete(block, "size")
	transactions, ok := block["transactions"].([]interface{})
	if ok {
		for _, tx := range transactions {
			txMap, ok := tx.(map[string]interface{})
			if ok {
				filterOpTx(txMap)
			}
		}
	}
	gasLimit, ok := block["gasLimit"].(string)
	if ok && gasLimit == "0x0" {
		delete(block, "uncles")
		delete(block, "sha3Uncles")
		delete(block, "mixHash")
		delete(block, "nonce")
		delete(block, "gasLimit")
	}
	return nil
}

func filterCeloBlock(blockNumber uint64, block map[string]interface{}, gingerbreadBlock uint64) error {
	// We remove the following fields:
	// size: being the size of the rlp encoded block it differs between the two systems since the block structure is different
	// randomness: we removed the concept of randomness for cel2 we filtered out the value in blocks during the migration.
	// epochSnarkData: same as randomness
	// gasLimit: removed for now since we don't have the value in the op-geth block so the op-geth block will just show 0, we may add this to op-geth later.
	// chainId (on transactions): celo didn't return chainId on legacy transactions, it did for other types but its a bit more involved to filter that per tx type.

	delete(block, "size")
	delete(block, "randomness")
	delete(block, "epochSnarkData")
	if blockNumber < gingerbreadBlock {
		// We hardcoded the gas limit in celo for pre-gingerbread blocks, we don't have that in op-geth so we remove it
		// from the celo block.
		delete(block, "gasLimit")
	}
	transactions, ok := block["transactions"].([]interface{})
	if ok {
		for _, tx := range transactions {
			txMap := tx.(map[string]interface{})
			filterCeloTx(txMap)
		}
	}

	// We need to filter out the istanbulAggregatedSeal from the extra data, since that was also filtered out during the migration process.
	extraData, ok := block["extraData"].(string)
	if !ok {
		return fmt.Errorf("extraData field not found or not a string in celo response")
	}

	extraDataBytes, err := hexutil.Decode(strings.TrimSpace(extraData))
	if err != nil {
		return fmt.Errorf("failed to hex decode extra data from celo response: %v", err)
	}

	if len(extraDataBytes) < IstanbulExtraVanity {
		return fmt.Errorf("invalid istanbul header extra-data length from res1 expecting at least %d but got %d", IstanbulExtraVanity, len(extraDataBytes))
	}

	istanbulExtra := &IstanbulExtra{}
	err = rlp.DecodeBytes(extraDataBytes[IstanbulExtraVanity:], istanbulExtra)
	if err != nil {
		return fmt.Errorf("failed to decode extra data from celo response: %v", err)
	}

	// Remove the istanbulAggregatedSeal from the extra data
	istanbulExtra.AggregatedSeal = IstanbulAggregatedSeal{}

	reEncodedExtra, err := rlp.EncodeToBytes(istanbulExtra)
	if err != nil {
		return fmt.Errorf("failed to re-encode extra data from celo response: %v", err)
	}
	finalEncodedString := hexutil.Encode(append(extraDataBytes[:IstanbulExtraVanity], reEncodedExtra...))

	block["extraData"] = finalEncodedString

	return nil
}

func (r *blockResults) verifyTransactions() error {
	makeTransactionsComparable(r.celoTxs)
	makeTransactionsComparable(r.opTxs)

	// We also need to take into account yparity (only set on op transactions which can cause different representations for the v value)
	// We overcome this by re-setting the v value of the celo txs to the op txs v value
	for _, tx := range r.celoTxs {
		// This sets the y value to be a big number where abs is nul rather than a zero length array if the numebr is zero.
		// It doesn't change the number but it does change the representation.
		types.SetYNullStyleBigIfZero(tx)
	}
	err := EqualObjects(r.celoTxs, r.opTxs)
	if err != nil {
		return err
	}
	// filter raw celo and op txs
	for i := range r.celoRawTxs {
		filterCeloTx(r.celoRawTxs[i])
	}
	for i := range r.opRawTxs {
		filterOpTx(r.opRawTxs[i])
	}
	err = EqualObjects(r.celoRawTxs, r.opRawTxs)
	if err != nil {
		return err
	}

	// cross check txs, unfortunately we can't easily do a direct comparison here so we compare number of txs and their
	// hashes.
	err = EqualObjects(len(r.celoTxs), len(r.celoRawTxs))
	if err != nil {
		return err
	}
	for i := range r.celoTxs {
		err = EqualObjects(r.celoTxs[i].Hash().String(), r.celoRawTxs[i]["hash"].(string))
		if err != nil {
			return err
		}
	}

	// Cross check the individually retrieved transactions with the transactions in the block
	return EqualObjects(r.celoTxs, []*types.Transaction(r.opBlockByNumber.Transactions()))
}

func (r *blockResults) verifyReceipts() error {
	err := EqualObjects(r.celoReceipts, r.opReceipts)
	if err != nil {
		return err
	}

	// filter the raw op receipts
	for i := range r.opReceipts {
		filterOpReceipt(r.opRawReceipts[i])
	}
	err = EqualObjects(len(r.celoRawReceipts), len(r.opRawReceipts))
	if err != nil {
		return err
	}
	for i := range r.celoRawReceipts {
		err = EqualObjects(r.celoRawReceipts[i], r.opRawReceipts[i])
		if err != nil {
			if r.celoRawReceipts[i]["effectiveGasPrice"] != nil && r.opRawReceipts[i]["effectiveGasPrice"] == nil {
				fmt.Printf("dangling state at block %d\n", r.blockNumber-1)
			} else {
				return err
			}
		}
	}
	// Cross check receipts, here we decode the raw receipts into receipt objects and compare them, since the raw
	// receipts are enriched with more fields than the receipt objects.
	var celoRawConverted []*types.Receipt
	jsonConvert(r.celoRawReceipts, &celoRawConverted)
	err = EqualObjects(celoRawConverted, r.celoReceipts)
	if err != nil {
		spew.Dump("celorawreceipts", r.celoRawBlockReceipts)
		return err
	}

	for _, opReceipt := range r.opRawReceipts {
		if opReceipt["cumulativeGasUsed"] == 0 {
			return fmt.Errorf("receipt with hash %q has CumulativeGasUsed == 0", opReceipt["transactionHash"])
		}
	}
	return nil
}

func (r *blockResults) verifyBlockReceipts() error {
	// Check block receipts pairs
	err := EqualObjects(r.celoBlockReceipts, r.opBlockReceipts)
	if err != nil {
		return err
	}

	// filter the raw op receipts
	for i := range r.opRawBlockReceipts {
		filterOpReceipt(r.opRawBlockReceipts[i])
	}
	err = EqualObjects(r.celoRawBlockReceipts, r.opRawBlockReceipts)
	if err != nil {
		return err
	}

	// Cross check receipts, here we decode the raw receipts into receipt objects and compare them, since the raw
	// receipts are enriched with more fields than the receipt objects.
	var celoRawConverted []*types.Receipt
	jsonConvert(r.celoRawBlockReceipts, &celoRawConverted)

	return EqualObjects(celoRawConverted, r.celoBlockReceipts)
}

func (r *blockResults) Verify(chainID uint64) error {
	// Cross check the tx and receipt effective gas price calculation
	for i, tx := range r.opRawTxs {
		EqualObjects(tx["effectiveGasPrice"], r.opRawReceipts[i]["effectiveGasPrice"])
	}

	err := r.verifyBlocks(chainID)
	if err != nil {
		return err
	}
	err = r.verifyTransactions()
	if err != nil {
		return err
	}
	err = r.verifyReceipts()
	if err != nil {
		return err
	}
	err = r.verifyBlockReceipts()
	if err != nil {
		return err
	}

	// Check the block receipt, we only have raw values for this because there is no method to retrieve it via the ethclient.
	// See https://docs.celo.org/developer/migrate/from-ethereum#core-contract-calls
	err = EqualObjects(r.celoRawBlockReceipt, r.opRawBlockReceipt)
	if err != nil {
		return err
	}

	toCrosscheckWithBlockReceipts := r.celoRawReceipts

	if r.celoRawBlockReceipt != nil {
		// Strangely we always return a block receipt even if there are no logs,
		// but we don't include it in the block receipts unless there are logs.
		if len(r.celoRawBlockReceipt["logs"].([]interface{})) > 0 {
			toCrosscheckWithBlockReceipts = append(toCrosscheckWithBlockReceipts, r.celoRawBlockReceipt)
		}
	}

	// Cross check block receipts with receipts
	err = EqualObjects(r.celoRawBlockReceipts, toCrosscheckWithBlockReceipts)
	if err != nil {
		return err
	}

	return nil
}

// Retreives all the logs for this block in order.
func (r *blockResults) Logs() ([]*types.Log, error) {
	logs := make([]*types.Log, 0)
	for _, receipt := range r.celoBlockReceipts {
		logs = append(logs, receipt.Logs...)
	}
	return logs, nil
}

func makeBlockComparable(b *types.Block) {
	// Blocks cache the hash
	b.Hash()
	makeTransactionsComparable(b.Transactions())
}
func makeTransactionsComparable(txs []*types.Transaction) {
	// Transactions in blocks cache the hash and also have a locally set time which varies
	for _, tx := range txs {
		tx.SetTime(time.Time{})
		tx.Hash()
	}
}

func fetchBlockElements(clients *clients, blockNumber uint64, resultChan chan *blockResults) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	blockNum := big.NewInt(int64(blockNumber))
	blockNumberHex := hexutil.EncodeUint64(blockNumber)

	var err error
	opBlockByNumber, err := clients.opEthclient.BlockByNumber(ctx, blockNum)
	if err != nil {
		return fmt.Errorf("%s: %w", "op.BlockByNumber", err)
	}
	blockHash := opBlockByNumber.Hash()
	l := len(opBlockByNumber.Transactions())

	results := &blockResults{
		blockNumber:       blockNumber,
		opBlockByNumber:   opBlockByNumber,
		celoTxs:           make([]*types.Transaction, l),
		opTxs:             make([]*types.Transaction, l),
		celoRawTxs:        make([]map[string]interface{}, l),
		opRawTxs:          make([]map[string]interface{}, l),
		celoReceipts:      make([]*types.Receipt, l),
		opReceipts:        make([]*types.Receipt, l),
		celoRawReceipts:   make([]map[string]interface{}, l),
		opRawReceipts:     make([]map[string]interface{}, l),
		celoBlockReceipts: make([]*types.Receipt, l),
		opBlockReceipts:   make([]*types.Receipt, l),
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(10)
	g.Go(func() error {
		var err error
		results.opBlockByHash, err = clients.opEthclient.BlockByHash(ctx, opBlockByNumber.Hash())
		if err != nil {
			return fmt.Errorf("%s: %w", "op.BlockByHash", err)
		}
		return nil
	})

	getBlockByNumber := "eth_getBlockByNumber"
	g.Go(func() error {
		err := rpcCall(ctx, clients.celoClient, &results.celoRawBlockByNumber, getBlockByNumber, blockNumberHex, true)
		if err != nil {
			return fmt.Errorf("%s: %w", getBlockByNumber, err)
		}
		return nil
	})
	g.Go(func() error {
		err := rpcCall(ctx, clients.opClient, &results.opRawBlockByNumber, getBlockByNumber, blockNumberHex, true)
		if err != nil {
			return fmt.Errorf("%s: %w", getBlockByNumber, err)
		}
		return nil
	})

	getBlockByHash := "eth_getBlockByHash"
	g.Go(func() error {
		err := rpcCall(ctx, clients.celoClient, &results.celoRawBlockByHash, getBlockByHash, blockHash.Hex(), true)
		if err != nil {
			return fmt.Errorf("%s: %w", getBlockByHash, err)
		}
		return nil
	})
	g.Go(func() error {
		err := rpcCall(ctx, clients.opClient, &results.opRawBlockByHash, getBlockByHash, blockHash.Hex(), true)
		if err != nil {
			return fmt.Errorf("%s: %w", getBlockByHash, err)
		}
		return nil
	})

	getBlockReceipt := "eth_getBlockReceipt"
	g.Go(func() error {
		err := rpcCall(ctx, clients.celoClient, &results.celoRawBlockReceipt, getBlockReceipt, blockHash.Hex())
		if err != nil {
			return fmt.Errorf("%s: %w", getBlockReceipt, err)
		}
		return nil
	})

	g.Go(func() error {
		err := rpcCall(ctx, clients.opClient, &results.opRawBlockReceipt, getBlockReceipt, blockHash.Hex())
		if err != nil {
			return fmt.Errorf("%s: %w", getBlockReceipt, err)
		}
		return nil
	})

	g.Go(func() error {
		celoBlockReceipts, err := clients.celoEthclient.BlockReceipts(ctx, rpc.BlockNumberOrHashWithHash(blockHash, true))
		if err != nil {
			return fmt.Errorf("BlockReceipts: %w", err)
		}
		results.celoBlockReceipts = celoBlockReceipts
		return nil
	})

	g.Go(func() error {
		opBlockReceipts, err := clients.opEthclient.BlockReceipts(ctx, rpc.BlockNumberOrHashWithHash(blockHash, true))
		if err != nil {
			return fmt.Errorf("BlockReceipts: %w", err)
		}
		results.opBlockReceipts = opBlockReceipts
		return nil
	})

	getBlockReceipts := "eth_getBlockReceipts"
	g.Go(func() error {
		err := rpcCall(ctx, clients.celoClient, &results.celoRawBlockReceipts, getBlockReceipts, blockHash.Hex())
		if err != nil {
			return fmt.Errorf("%s: %w", getBlockReceipts, err)
		}
		return nil
	})

	g.Go(func() error {
		err := rpcCall(ctx, clients.opClient, &results.opRawBlockReceipts, getBlockReceipts, blockHash.Hex())
		if err != nil {
			return fmt.Errorf("%s: %w", getBlockReceipts, err)
		}
		return nil
	})

	// For each transaction in blockByNumber we retrieve it using the
	// cethClient, ethClient, cClient, and client. Each transaction is
	// retrieved in its own goroutine and sets itself in the respective slice
	// that is accessible from its closure.
	for i, tx := range opBlockByNumber.Transactions() {
		h := tx.Hash()
		hexHash := h.Hex()
		index := i
		g.Go(func() error {
			celoTx, _, err := clients.celoEthclient.TransactionByHash(ctx, h)
			if err != nil {
				return fmt.Errorf("celoEthclient.TransactionByHash: %s, err: %w", h, err)
			}
			results.celoTxs[index] = celoTx
			return nil
		})

		g.Go(func() error {
			opTx, _, err := clients.opEthclient.TransactionByHash(ctx, h)
			if err != nil {
				return fmt.Errorf("opEthclient.TransactionByHash: %w", err)
			}
			results.opTxs[index] = opTx
			return nil
		})

		getTxByHash := "eth_getTransactionByHash"
		g.Go(func() error {
			err := rpcCall(ctx, clients.celoClient, &results.celoRawTxs[index], getTxByHash, hexHash)
			if err != nil {
				return fmt.Errorf("%s: %w", getTxByHash, err)
			}
			return nil
		})

		g.Go(func() error {
			err := rpcCall(ctx, clients.opClient, &results.opRawTxs[index], getTxByHash, hexHash)
			if err != nil {
				return fmt.Errorf("%s: %w", getTxByHash, err)
			}
			return nil
		})

		g.Go(func() error {
			celoReceipt, err := clients.celoEthclient.TransactionReceipt(ctx, h)
			if err != nil {
				return fmt.Errorf("celoEthclient.TransactionReceipt: %w", err)
			}
			results.celoReceipts[index] = celoReceipt
			return nil
		})

		g.Go(func() error {
			opReceipt, err := clients.opEthclient.TransactionReceipt(ctx, h)
			if err != nil {
				return fmt.Errorf("opEthclient.TransactionReceipt: %w", err)
			}
			results.opReceipts[index] = opReceipt
			return nil
		})

		getTransactionReceipt := "eth_getTransactionReceipt"
		g.Go(func() error {
			err := rpcCall(ctx, clients.celoClient, &results.celoRawReceipts[index], getTransactionReceipt, hexHash)
			if err != nil {
				return fmt.Errorf("%s: %w", getTransactionReceipt, err)
			}
			return nil
		})

		g.Go(func() error {
			err := rpcCall(ctx, clients.opClient, &results.opRawReceipts[index], getTransactionReceipt, hexHash)
			if err != nil {
				return fmt.Errorf("%s: %w", getTransactionReceipt, err)
			}
			return nil
		})
	}

	err = g.Wait()
	if err != nil {
		return err
	}

	resultChan <- results
	return err
}

func jsonConvert(in, out any) error {
	marshaled, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(marshaled, out)
}

func rpcCall(ctx context.Context, cl *rpc.Client, result interface{}, method string, args ...interface{}) error {
	err := cl.CallContext(ctx, result, method, args...)
	if err != nil {
		return err
	}
	// All other responses should be non null, but for block receipts since it's
	// not linked to a transaction we have to call the endpoint to find out if
	// one exists.
	if method != "eth_getBlockReceipt" {
		if result == "nil" {
			return fmt.Errorf("response for %v %v should not be null", method, args)
		}
	}
	return nil
}

func filterOpReceipt(receipt map[string]interface{}) {
	// Delete effective gas price fields that are nil, on the celo side we do not add them.
	v, ok := receipt["effectiveGasPrice"]
	if ok && v == nil {
		delete(receipt, "effectiveGasPrice")
	}
}

func filterOpTx(tx map[string]interface{}) {
	// Some txs on celo contain chainID all of them do on op, so we just remove it from both sides.
	delete(tx, "chainId")
	// Celo never returned yParity
	delete(tx, "yParity")
	// Since we unequivocally delete gatewayFee on the celo side we need to delete it here as well.
	delete(tx, "gatewayFee")
}

func filterCeloTx(tx map[string]interface{}) {
	// Some txs on celo contain chainID all of them do on op, so we just remove it from both sides.
	delete(tx, "chainId")
	// On the op side we now don't return ethCompatible when it's true, so we
	// remove it from the celo response in this case.
	txType := tx["type"].(string)
	if txType == "0x0" && tx["ethCompatible"].(bool) {
		delete(tx, "ethCompatible")
	}
	//It seems gateway fee is always added to all rpc transaction responses on celo because tx.GatewayFee returns 0 if
	//it's not set,even ethcompatible ones, this is confusing so we have removed this in the op code, so we need to make
	//sure the celo side matches.
	delete(tx, "gatewayFee")
	// Transaction at block 9476837 returns a nullGasPrice on celo but the op-geth side returns the gasFeeCap
	// delete(tx, "gasPrice")
}

type filterQuery struct {
	FromBlock hexutil.Uint64 `json:"fromBlock"`
	ToBlock   hexutil.Uint64 `json:"toBlock"`
}

var (
	IstanbulExtraVanity = 32 // Fixed number of extra-data bytes reserved for validator vanity
)

// IstanbulAggregatedSeal is the aggregated seal for Istanbul blocks
type IstanbulAggregatedSeal struct {
	// Bitmap is a bitmap having an active bit for each validator that signed this block
	Bitmap *big.Int
	// Signature is an aggregated BLS signature resulting from signatures by each validator that signed this block
	Signature []byte
	// Round is the round in which the signature was created.
	Round *big.Int
}

// IstanbulExtra is the extra-data for Istanbul blocks
type IstanbulExtra struct {
	// AddedValidators are the validators that have been added in the block
	AddedValidators []common.Address
	// AddedValidatorsPublicKeys are the BLS public keys for the validators added in the block
	AddedValidatorsPublicKeys [][96]byte
	// RemovedValidators is a bitmap having an active bit for each removed validator in the block
	RemovedValidators *big.Int
	// Seal is an ECDSA signature by the proposer
	Seal []byte
	// AggregatedSeal contains the aggregated BLS signature created via IBFT consensus.
	AggregatedSeal IstanbulAggregatedSeal
	// ParentAggregatedSeal contains and aggregated BLS signature for the previous block.
	ParentAggregatedSeal IstanbulAggregatedSeal
}
