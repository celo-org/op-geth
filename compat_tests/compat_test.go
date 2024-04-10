package compat_tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"
)

type blockTransactions struct {
	Transactions []*types.Transaction `json:"transactions"`
}

type blockHash struct {
	Hash common.Hash `json:"hash"`
}

func TestCompatibilityOfChain(t *testing.T) {
	dumpOutput := false
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	c, err := rpc.DialContext(ctx, "http://localhost:8545")
	require.NoError(t, err)

	ec := ethclient.NewClient(c)
	ctx, cancel = context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()
	id, err := ec.ChainID(ctx)
	require.NoError(t, err)
	require.Greater(t, id.Uint64(), uint64(0))

	var res json.RawMessage
	startBlock := uint64(2800)
	amount := uint64(1000)
	incrementalLogs := make([]*types.Log, 0)
	for i := startBlock; i <= startBlock+amount; i++ {
		ctx, cancel = context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		b, err := ec.BlockByNumber(ctx, big.NewInt(int64(i)))
		require.NoError(t, err)
		h, err := ec.HeaderByNumber(ctx, big.NewInt(int64(i)))
		require.NoError(t, err)

		// Comparing the 2 headers directly doesn't work, they differ on unset
		// and nil big ints, which represent the same thing but are technically
		// not eaual. So instead we compare the marshalled output.
		bhMarshalled, err := json.Marshal(b.Header())
		require.NoError(t, err)
		hMarshalled, err := json.Marshal(h)
		require.NoError(t, err)
		require.Equal(t, bhMarshalled, hMarshalled)

		// Now get the header by hash and compare
		h2, err := ec.HeaderByHash(ctx, h.Hash())
		require.NoError(t, err)
		h2Marshalled, err := json.Marshal(h2)
		require.NoError(t, err)
		require.Equal(t, hMarshalled, h2Marshalled)

		// Now get the block by hash and compare
		b2, err := ec.BlockByHash(ctx, b.Hash())
		require.NoError(t, err)
		b2hMarshalled, err := json.Marshal(b2.Header())
		require.NoError(t, err)
		require.Equal(t, bhMarshalled, b2hMarshalled)

		bbMarshalled, err := json.Marshal(b.Body())
		require.NoError(t, err)
		b2bMarshalled, err := json.Marshal(b2.Body())
		require.NoError(t, err)
		require.Equal(t, bbMarshalled, b2bMarshalled)

		res, err = rpcCall(c, dumpOutput, "eth_getBlockByNumber", hexutil.EncodeUint64(i), true)
		require.NoError(t, err)
		// Check we got a block
		require.NotEqual(t, "null", string(res), "block %d should not be null", i)
		blockHash := blockHash{}
		err = json.Unmarshal(res, &blockHash)
		require.NoError(t, err)
		txs := blockTransactions{}
		err = json.Unmarshal(res, &txs)
		require.NoError(t, err)

		incrementalBlockReceipts := types.Receipts{}
		for i, tx := range txs.Transactions {
			// Compare transactions decoded from rpcCall with transactions from
			// the the block returned by ethclient and those directly retrieved
			// by ethclient. Comparing transactions directly does not work
			// because of differing private fields. The hash is calculated over
			// the data of the transaction and so serves as a good proxy for
			// equality.
			require.Equal(t, tx.Hash(), b.Transactions()[i].Hash())
			tx2, _, err := ec.TransactionByHash(ctx, tx.Hash())
			require.NoError(t, err)
			require.Equal(t, tx.Hash(), tx2.Hash())

			_, err = rpcCall(c, dumpOutput, "eth_getTransactionByHash", tx.Hash())
			require.NoError(t, err)
			res, err = rpcCall(c, dumpOutput, "eth_getTransactionReceipt", tx.Hash())
			require.NoError(t, err)
			r := types.Receipt{}
			err = json.Unmarshal(res, &r)
			require.NoError(t, err)

			// Check receipt decoded from RPC call matches that returned directly from ethclient.
			r2, err := ec.TransactionReceipt(ctx, tx.Hash())
			require.NoError(t, err)
			require.Equal(t, r, *r2)

			incrementalBlockReceipts = append(incrementalBlockReceipts, &r)
			incrementalLogs = append(incrementalLogs, r.Logs...)
		}
		// Get the Celo block receipt. See https://docs.celo.org/developer/migrate/from-ethereum#core-contract-calls
		res, err = rpcCall(c, dumpOutput, "eth_getBlockReceipt", blockHash.Hash)
		require.NoError(t, err)
		if string(res) != "null" {
			r := types.Receipt{}
			err = json.Unmarshal(res, &r)
			require.NoError(t, err)
			if len(r.Logs) > 0 {
				// eth_getBlockReceipt generates an empty receipt when there
				// are no logs, we want to avoid adding these here since the
				// same is not done in eth_gethBlockReceipts, the output of
				// which we will later compare against.
				incrementalBlockReceipts = append(incrementalBlockReceipts, &r)
			}
			incrementalLogs = append(incrementalLogs, r.Logs...)
		}

		blockReceipts := types.Receipts{}
		res, err = rpcCall(c, dumpOutput, "eth_getBlockReceipts", hexutil.EncodeUint64(i))
		require.NoError(t, err)
		err = json.Unmarshal(res, &blockReceipts)
		require.NoError(t, err)
		require.Equal(t, incrementalBlockReceipts, blockReceipts)
	}

	// Get all logs for the range and compare with the logs extracted from receipts.
	from := rpc.BlockNumber(startBlock)
	to := rpc.BlockNumber(amount + startBlock)
	res, err = rpcCall(c, dumpOutput, "eth_getLogs", filterQuery{
		FromBlock: &from,
		ToBlock:   &to,
	})
	require.NoError(t, err)
	var logs []*types.Log
	err = json.Unmarshal(res, &logs)
	require.NoError(t, err)
	require.Equal(t, len(incrementalLogs), len(logs))
	require.Equal(t, incrementalLogs, logs)

	// Compare logs with those retrived via ethclient
	logs2, err := ec.FilterLogs(ctx, ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(startBlock)),
		ToBlock:   big.NewInt(int64(startBlock + amount)),
	})
	// ethclient returns non pointers to logs, convert to pointer form
	logPointers := make([]*types.Log, len(logs2))
	for i := range logs2 {
		logPointers[i] = &logs2[i]
	}
	require.Equal(t, logPointers, logs)
}

type filterQuery struct {
	BlockHash *common.Hash     `json:"blockHash"`
	FromBlock *rpc.BlockNumber `json:"fromBlock"`
	ToBlock   *rpc.BlockNumber `json:"toBlock"`
	Addresses interface{}      `json:"address"`
	Topics    []interface{}    `json:"topics"`
}

func rpcCall(c *rpc.Client, dumpOutput bool, method string, args ...interface{}) (json.RawMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*1)
	defer cancel()
	var m json.RawMessage
	err := c.CallContext(ctx, &m, method, args...)
	if err != nil {
		return nil, err
	}
	if dumpOutput {
		dst := &bytes.Buffer{}
		err = json.Indent(dst, m, "", "  ")
		if err != nil {
			return nil, err
		}
		fmt.Printf("%v\n%v\n", method, dst.String())
	}
	return m, nil
}

// func TestCompatibilityOfChainWithDiff(t *testing.T) {

// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
// 	defer cancel()
// 	c, err := ethclient.DialContext(ctx, "http://localhost:8545")
// 	require.NoError(t, err)
// 	for i := int64(2800); i < 2800+100; i++ {
// 		ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
// 		defer cancel()
// 		b, err := c.BlockByNumber(ctx, big.NewInt(i))
// 		require.NoError(t, err)
// 		for _, tx := range b.Transactions() {
// 			ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
// 			defer cancel()
// 			, err := c.TransactionReceipt(ctx, tx.Hash())
// 			require.NoError(t, err)
// 		}
// 	}

// }
