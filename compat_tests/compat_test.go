package compat_tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"
)

type blockTransactions struct {
	Transactions []*types.Transaction `json:"transactions"`
}

func TestCompatibilityOfChain(t *testing.T) {
	dumpOutput := false
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	c, err := rpc.DialContext(ctx, "http://localhost:8545")
	require.NoError(t, err)
	startBlock := uint64(2800)
	amount := uint64(1000)
	incrementalLogs := make([]*types.Log, 0)
	for i := startBlock; i < startBlock+amount; i++ {
		res, err := rpcCall(c, dumpOutput, "eth_getBlockByNumber", hexutil.EncodeUint64(i), true)
		require.NoError(t, err)
		txs := blockTransactions{}
		err = json.Unmarshal(res, &txs)
		require.NoError(t, err)
		for _, tx := range txs.Transactions {
			_, err = rpcCall(c, dumpOutput, "eth_getTransactionByHash", tx.Hash())
			require.NoError(t, err)
			res, err = rpcCall(c, dumpOutput, "eth_getTransactionReceipt", tx.Hash())
			require.NoError(t, err)
			r := types.Receipt{}
			err = json.Unmarshal(res, &r)
			require.NoError(t, err)
			incrementalLogs = append(incrementalLogs, r.Logs...)
		}
	}

	// Get all logs for the range and compare with the logs extracted from receipts.
	ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	from := rpc.BlockNumber(startBlock)
	to := rpc.BlockNumber(amount + startBlock)
	res, err := rpcCall(c, dumpOutput, "eth_getLogs", filterQuery{
		FromBlock: &from,
		ToBlock:   &to,
	})
	require.NoError(t, err)
	var logs []*types.Log
	err = json.Unmarshal(res, &logs)
	require.NoError(t, err)
	require.Equal(t, len(incrementalLogs), len(logs))
	require.Equal(t, incrementalLogs, logs)
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
