package compat_tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"
)

func TestCompatibilityOfChain(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	c, err := rpc.DialContext(ctx, "http://localhost:8545")
	require.NoError(t, err)
	for i := uint64(2800); i < 2800+100; i++ {
		res, err := rpcCall(c, "eth_getBlockByNumber", hexutil.EncodeUint64(i), true)
		require.NoError(t, err)
		b := types.Block{}
		err = json.Unmarshal(res, &b)
		require.NoError(t, err)
		for _, tx := range b.Transactions() {
			rpcCall(c, "eth_getTransactionByHash", tx.Hash())
		}
	}

	// // h := common.HexToHash("0xe1ee23cfb65ca96e96b68f22ddafd73f0f285f61afb980bc71532f2534815f54")
	// // try getting some logs
	// ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	// defer cancel()
	// logs, err := c.FilterLogs(ctx, celo.FilterQuery{
	// 	// FromBlock: big.NewInt(0),
	// 	// ToBlock:   big.NewInt(30000),
	// 	FromBlock: big.NewInt(2937),
	// 	ToBlock:   big.NewInt(2939),
	// 	// BlockHash: &h,
	// })
	// require.NoError(t, err)
	// fmt.Printf("num logs %d\n", len(logs))

}

func rpcCall(c *rpc.Client, method string, args ...interface{}) (json.RawMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour*1)
	defer cancel()
	var m json.RawMessage
	err := c.CallContext(ctx, &m, method, args...)
	if err != nil {
		return nil, err
	}
	dst := &bytes.Buffer{}
	err = json.Indent(dst, m, "", "  ")
	if err != nil {
		return nil, err
	}
	fmt.Printf("%v\n%v\n", method, dst.String())
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
