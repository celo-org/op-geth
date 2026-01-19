// Copyright 2024 The celo Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/stateless"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/superchain"
	"github.com/urfave/cli/v2"
)

var commandWitness = &cli.Command{
	Name:  "witness",
	Usage: "Execute a witness and verify the result",
	Subcommands: []*cli.Command{
		{
			Name:   "execute",
			Usage:  "Execute a witness file against a block",
			Action: executeWitness,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "witness",
					Usage:    "Path to witness JSON file",
					Required: true,
				},
				&cli.StringFlag{
					Name:  "block",
					Usage: "Path to block JSON file (if not provided, fetches from RPC)",
				},
				&cli.StringFlag{
					Name:  "rpc",
					Usage: "RPC endpoint to fetch block data",
					Value: "https://forno.celo.org",
				},
				&cli.Uint64Flag{
					Name:  "block-number",
					Usage: "Block number to execute (defaults to block after witness header)",
				},
				&cli.Uint64Flag{
					Name:  "chainid",
					Usage: "Chain ID (42220 for Celo mainnet, 44787 for Alfajores)",
					Value: 42220,
				},
			},
		},
	},
}

// rpcBlock is the JSON structure returned by eth_getBlockByNumber
type rpcBlock struct {
	Hash             common.Hash         `json:"hash"`
	ParentHash       common.Hash         `json:"parentHash"`
	UncleHash        common.Hash         `json:"sha3Uncles"`
	Coinbase         common.Address      `json:"miner"`
	Root             common.Hash         `json:"stateRoot"`
	TxHash           common.Hash         `json:"transactionsRoot"`
	ReceiptHash      common.Hash         `json:"receiptsRoot"`
	Bloom            types.Bloom         `json:"logsBloom"`
	Difficulty       *hexutil.Big        `json:"difficulty"`
	Number           *hexutil.Big        `json:"number"`
	GasLimit         hexutil.Uint64      `json:"gasLimit"`
	GasUsed          hexutil.Uint64      `json:"gasUsed"`
	Time             hexutil.Uint64      `json:"timestamp"`
	Extra            hexutil.Bytes       `json:"extraData"`
	MixDigest        common.Hash         `json:"mixHash"`
	Nonce            types.BlockNonce    `json:"nonce"`
	BaseFee          *hexutil.Big        `json:"baseFeePerGas"`
	WithdrawalsHash  *common.Hash        `json:"withdrawalsRoot"`
	BlobGasUsed      *hexutil.Uint64     `json:"blobGasUsed"`
	ExcessBlobGas    *hexutil.Uint64     `json:"excessBlobGas"`
	ParentBeaconRoot *common.Hash        `json:"parentBeaconBlockRoot"`
	RequestsHash     *common.Hash        `json:"requestsHash"`
	Transactions     []json.RawMessage   `json:"transactions"`
	Uncles           []common.Hash       `json:"uncles"`
	Withdrawals      []*types.Withdrawal `json:"withdrawals"`
}

func executeWitness(ctx *cli.Context) error {
	witnessPath := ctx.String("witness")
	blockPath := ctx.String("block")
	rpcURL := ctx.String("rpc")
	chainID := ctx.Uint64("chainid")

	// Load witness from JSON file
	witnessData, err := os.ReadFile(witnessPath)
	if err != nil {
		return fmt.Errorf("failed to read witness file: %w", err)
	}

	var execWitness stateless.ExecutionWitness
	if err := json.Unmarshal(witnessData, &execWitness); err != nil {
		return fmt.Errorf("failed to parse witness JSON: %w", err)
	}

	// Convert to internal witness format
	witness, err := stateless.FromExecutionWitness(&execWitness)
	if err != nil {
		return fmt.Errorf("failed to convert witness: %w", err)
	}

	fmt.Printf("Loaded witness:\n")
	fmt.Printf("  Headers: %d\n", len(witness.Headers))
	fmt.Printf("  Codes: %d\n", len(witness.Codes))
	fmt.Printf("  State nodes: %d\n", len(witness.State))
	fmt.Printf("  Pre-state root: %s\n", witness.Root().Hex())

	// Determine block number to execute
	blockNum := ctx.Uint64("block-number")
	if blockNum == 0 && blockPath == "" {
		// Default to the block after the parent header in the witness
		if len(witness.Headers) == 0 {
			return fmt.Errorf("witness has no headers")
		}
		blockNum = witness.Headers[0].Number.Uint64() + 1
	}

	var block *types.Block
	var expectedStateRoot, expectedReceiptRoot common.Hash

	if blockPath != "" {
		// Load block from JSON file
		block, expectedStateRoot, expectedReceiptRoot, err = loadBlockFromJSON(blockPath)
		if err != nil {
			return fmt.Errorf("failed to load block from JSON: %w", err)
		}
	} else {
		// Fetch block from RPC
		fmt.Printf("  Fetching block %d from RPC...\n", blockNum)
		block, err = fetchBlockFromRPC(rpcURL, blockNum)
		if err != nil {
			return fmt.Errorf("failed to fetch block: %w", err)
		}
		expectedStateRoot = block.Root()
		expectedReceiptRoot = block.ReceiptHash()
	}

	fmt.Printf("\nBlock info:\n")
	fmt.Printf("  Number: %d\n", block.NumberU64())
	fmt.Printf("  Hash: %s\n", block.Hash().Hex())
	fmt.Printf("  Parent hash: %s\n", block.ParentHash().Hex())
	fmt.Printf("  Transactions: %d\n", len(block.Transactions()))
	fmt.Printf("  Gas used: %d\n", block.GasUsed())
	fmt.Printf("  Expected state root: %s\n", expectedStateRoot.Hex())
	fmt.Printf("  Expected receipt root: %s\n", expectedReceiptRoot.Hex())

	// Verify parent hash matches witness
	if len(witness.Headers) > 0 {
		witnessParentHash := witness.Headers[0].Hash()
		if block.ParentHash() != witnessParentHash {
			return fmt.Errorf("parent hash mismatch: block has %s, witness has %s",
				block.ParentHash().Hex(), witnessParentHash.Hex())
		}
		fmt.Printf("\n✓ Parent hash matches witness\n")
	}

	// Get chain config
	config, err := getChainConfig(chainID)
	if err != nil {
		return fmt.Errorf("failed to get chain config: %w", err)
	}

	// Create a block copy with zeroed roots for stateless execution
	statelessBlock := createStatelessBlock(block)

	// Execute stateless
	fmt.Printf("\nExecuting stateless...\n")
	vmConfig := vm.Config{}
	stateRoot, receiptRoot, err := core.ExecuteStateless(config, vmConfig, statelessBlock, witness)
	if err != nil {
		fmt.Printf("\n❌ Stateless execution failed: %v\n", err)
		fmt.Printf("\nThis may indicate an invalid transaction in the block.\n")
		fmt.Printf("Analyzing transactions for chain ID mismatches...\n\n")
		for i, tx := range block.Transactions() {
			txChainID := tx.ChainId()
			if txChainID != nil && txChainID.Uint64() != chainID && txChainID.Uint64() != 0 {
				fmt.Printf("  ⚠ Tx %d [%s]: chain ID %d (expected %d)\n",
					i, tx.Hash().Hex()[:18]+"...", txChainID.Uint64(), chainID)
				fmt.Printf("    Type: %d, From: (cannot recover with wrong chain ID)\n", tx.Type())
			}
		}
		return fmt.Errorf("stateless execution failed: %w", err)
	}

	fmt.Printf("\nResults:\n")
	fmt.Printf("  Computed state root:   %s\n", stateRoot.Hex())
	fmt.Printf("  Expected state root:   %s\n", expectedStateRoot.Hex())
	fmt.Printf("  Computed receipt root: %s\n", receiptRoot.Hex())
	fmt.Printf("  Expected receipt root: %s\n", expectedReceiptRoot.Hex())

	// Compare results
	stateMatch := stateRoot == expectedStateRoot
	receiptMatch := receiptRoot == expectedReceiptRoot

	if stateMatch {
		fmt.Printf("\n✓ State root matches!\n")
	} else {
		fmt.Printf("\n✗ State root MISMATCH!\n")
	}

	if receiptMatch {
		fmt.Printf("✓ Receipt root matches!\n")
	} else {
		fmt.Printf("✗ Receipt root MISMATCH!\n")
	}

	if stateMatch && receiptMatch {
		fmt.Printf("\n✓ Witness execution successful!\n")
		return nil
	}

	return fmt.Errorf("witness verification failed")
}

func loadBlockFromJSON(path string) (*types.Block, common.Hash, common.Hash, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, common.Hash{}, common.Hash{}, err
	}

	var rb rpcBlock
	if err := json.Unmarshal(data, &rb); err != nil {
		return nil, common.Hash{}, common.Hash{}, fmt.Errorf("failed to parse block JSON: %w", err)
	}

	// Parse transactions
	txs := make([]*types.Transaction, len(rb.Transactions))
	for i, txData := range rb.Transactions {
		var tx types.Transaction
		if err := tx.UnmarshalJSON(txData); err != nil {
			return nil, common.Hash{}, common.Hash{}, fmt.Errorf("failed to parse transaction %d: %w", i, err)
		}
		txs[i] = &tx
	}

	// Build header
	header := &types.Header{
		ParentHash:       rb.ParentHash,
		UncleHash:        rb.UncleHash,
		Coinbase:         rb.Coinbase,
		Root:             common.Hash{}, // Zero for stateless
		TxHash:           rb.TxHash,
		ReceiptHash:      common.Hash{}, // Zero for stateless
		Bloom:            rb.Bloom,
		Difficulty:       (*big.Int)(rb.Difficulty),
		Number:           (*big.Int)(rb.Number),
		GasLimit:         uint64(rb.GasLimit),
		GasUsed:          uint64(rb.GasUsed),
		Time:             uint64(rb.Time),
		Extra:            rb.Extra,
		MixDigest:        rb.MixDigest,
		Nonce:            rb.Nonce,
		BaseFee:          (*big.Int)(rb.BaseFee),
		WithdrawalsHash:  rb.WithdrawalsHash,
		ParentBeaconRoot: rb.ParentBeaconRoot,
		RequestsHash:     rb.RequestsHash,
	}
	if rb.BlobGasUsed != nil {
		header.BlobGasUsed = (*uint64)(rb.BlobGasUsed)
	}
	if rb.ExcessBlobGas != nil {
		header.ExcessBlobGas = (*uint64)(rb.ExcessBlobGas)
	}

	block := types.NewBlockWithHeader(header).WithBody(types.Body{
		Transactions: txs,
		Withdrawals:  rb.Withdrawals,
	})

	return block, rb.Root, rb.ReceiptHash, nil
}

func fetchBlockFromRPC(rpcURL string, blockNum uint64) (*types.Block, error) {
	rpcClient, err := rpc.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC: %w", err)
	}
	defer rpcClient.Close()

	client := ethclient.NewClient(rpcClient)

	block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(blockNum)))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch block %d: %w", blockNum, err)
	}

	return block, nil
}

func createStatelessBlock(block *types.Block) *types.Block {
	// Create a block copy with zeroed roots for stateless execution
	header := &types.Header{
		ParentHash:       block.ParentHash(),
		UncleHash:        block.UncleHash(),
		Coinbase:         block.Coinbase(),
		Root:             common.Hash{}, // Zero for stateless
		TxHash:           block.TxHash(),
		ReceiptHash:      common.Hash{}, // Zero for stateless
		Bloom:            block.Bloom(),
		Difficulty:       block.Difficulty(),
		Number:           block.Number(),
		GasLimit:         block.GasLimit(),
		GasUsed:          block.GasUsed(),
		Time:             block.Time(),
		Extra:            block.Extra(),
		MixDigest:        block.MixDigest(),
		Nonce:            types.EncodeNonce(block.Nonce()),
		BaseFee:          block.BaseFee(),
		WithdrawalsHash:  block.Header().WithdrawalsHash,
		BlobGasUsed:      block.Header().BlobGasUsed,
		ExcessBlobGas:    block.Header().ExcessBlobGas,
		ParentBeaconRoot: block.Header().ParentBeaconRoot,
		RequestsHash:     block.Header().RequestsHash,
	}

	return types.NewBlockWithHeader(header).WithBody(types.Body{
		Transactions: block.Transactions(),
		Uncles:       block.Uncles(),
		Withdrawals:  block.Withdrawals(),
	})
}

func getChainConfig(chainID uint64) (*params.ChainConfig, error) {
	// Special case for Celo mainnet which is not in the superchain registry
	if chainID == params.CeloMainnetChainID {
		return getCeloMainnetConfig(), nil
	}

	// Try to load from superchain registry
	for _, chain := range superchain.Chains {
		cfg, err := chain.Config()
		if err != nil {
			continue
		}
		if cfg.ChainID == chainID {
			return params.LoadOPStackChainConfig(cfg)
		}
	}
	return nil, fmt.Errorf("chain ID %d not found in superchain registry", chainID)
}

// getCeloMainnetConfig returns the chain config for Celo mainnet
// This is hardcoded since Celo mainnet is not in the superchain registry
func getCeloMainnetConfig() *params.ChainConfig {
	// Celo mainnet fork timestamps
	// Cel2 (L2 migration) activated at genesis of L2
	cel2Time := uint64(0)
	canyonTime := uint64(0)
	ecotoneTime := uint64(0)
	fjordTime := uint64(0)
	graniteTime := uint64(0)
	holoceneTime := uint64(0)
	isthmusTime := params.CeloMainnetIsthmusTimestamp

	return &params.ChainConfig{
		ChainID: big.NewInt(params.CeloMainnetChainID),

		// Ethereum forks - all activated at genesis for L2
		HomesteadBlock:      common.Big0,
		DAOForkBlock:        nil,
		DAOForkSupport:      false,
		EIP150Block:         common.Big0,
		EIP155Block:         common.Big0,
		EIP158Block:         common.Big0,
		ByzantiumBlock:      common.Big0,
		ConstantinopleBlock: common.Big0,
		PetersburgBlock:     common.Big0,
		IstanbulBlock:       common.Big0,
		MuirGlacierBlock:    common.Big0,
		BerlinBlock:         common.Big0,
		LondonBlock:         common.Big0,
		ArrowGlacierBlock:   common.Big0,
		GrayGlacierBlock:    common.Big0,
		MergeNetsplitBlock:  common.Big0,
		ShanghaiTime:        &canyonTime,   // Shanghai activates with Canyon
		CancunTime:          &ecotoneTime,  // Cancun activates with Ecotone
		PragueTime:          &isthmusTime,  // Prague activates with Isthmus
		VerkleTime:          nil,

		// Optimism forks
		BedrockBlock: common.Big0,
		RegolithTime: &cel2Time,
		CanyonTime:   &canyonTime,
		EcotoneTime:  &ecotoneTime,
		FjordTime:    &fjordTime,
		GraniteTime:  &graniteTime,
		HoloceneTime: &holoceneTime,
		IsthmusTime:  &isthmusTime,
		JovianTime:   nil,
		InteropTime:  nil,

		// Celo forks
		Cel2Time:         &cel2Time,
		GingerbreadBlock: common.Big0,

		TerminalTotalDifficulty: common.Big0,

		// Consensus engines
		Ethash: nil,
		Clique: nil,

		Optimism: &params.OptimismConfig{
			EIP1559Elasticity:        5,
			EIP1559Denominator:       400,
			EIP1559DenominatorCanyon: newUint64(400),
		},
		Celo: &params.CeloConfig{
			EIP1559BaseFeeFloor: 25000000000, // 25 gwei
		},
	}
}

func newUint64(n uint64) *uint64 {
	return &n
}
