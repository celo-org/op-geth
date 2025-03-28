// Package main provides the entrypoint for the receipt_hash command
// This tool allows computing and verifying receipt root hashes either from:
// 1. A JSON file containing receipts
// 2. By querying an Ethereum RPC node for a specific block
//
// When querying an RPC node, it first tries to use the more efficient BlockReceipts
// method to fetch all receipts in a single call. If that fails or is not supported,
// it falls back to fetching receipts individually using TransactionReceipt.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/trie"
)

func main() {
	debug := false
	args := os.Args[1:]

	// Check for debug flag
	if len(args) > 0 && args[0] == "--debug" {
		debug = true
		args = args[1:]
	}

	if len(args) == 1 {
		// Original functionality - process JSON file
		processReceiptFile(args[0])
	} else if len(args) == 2 {
		// New functionality - fetch from RPC
		// This will use BlockReceipts when available for more efficient receipt fetching
		// and fall back to individual receipt fetching when BlockReceipts is not supported
		fetchAndProcessBlock(args[0], args[1], debug)
	} else {
		log.Fatal("Usage: \n" +
			"  receipt_hash_cmd [--debug] <path_to_receipt.json>\n" +
			"  receipt_hash_cmd [--debug] <rpc_url> <block_number>")
	}
}

func processReceiptFile(filePath string) {
	// Read the JSON file
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	// Parse the receipt
	var receipts types.Receipts
	if err := json.Unmarshal(data, &receipts); err != nil {
		log.Fatalf("Error parsing receipt: %v", err)
	}

	// Compute the receipt root hash
	receiptRoot := computeReceiptRoot(receipts)
	fmt.Printf("Receipt Root from file: %x\n", receiptRoot)
}

func fetchAndProcessBlock(rpcURL, blockNumberStr string, debug bool) {
	// Parse block numbertype
	blockNumber, err := strconv.ParseUint(blockNumberStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid block number: %v", err)
	}

	// Connect to the Ethereum client
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	defer client.Close()

	// Convert blockNumber to *big.Int
	blockNumberBig := new(big.Int).SetUint64(blockNumber)

	// Get the block
	block, err := client.BlockByNumber(context.Background(), blockNumberBig)
	if err != nil {
		log.Fatalf("Failed to fetch block: %v", err)
	}

	// Print block header info
	fmt.Printf("Block number: %d\n", blockNumber)
	fmt.Printf("Receipt Root from block header: %x\n", block.ReceiptHash())
	fmt.Printf("Number of transactions: %d\n", len(block.Transactions()))

	// Try to fetch receipts and compute the root locally
	fmt.Println("\nAttempting to fetch receipts and compute local root...")

	// Get receipts for the block, handling errors gracefully
	receipts, err := getReceiptsForBlock(client, block)
	if err != nil {
		fmt.Printf("⚠️ Warning: %v\n", err)
		return
	}

	// Debug info if requested
	if debug && len(receipts) > 0 {
		// fmt.Println("\n🔍 DEBUG: First receipt details:")
		for i := 0; i < len(receipts); i++ {
			printReceiptDetails(receipts[i])
		}

	}

	// Count and print transaction types
	txTypeCounts := countTxTypes(receipts)
	fmt.Println("\n📊 Transaction Types:")

	// Get sorted keys for consistent output
	var txTypes []uint8
	for txType := range txTypeCounts {
		txTypes = append(txTypes, txType)
	}
	//sort.Slice(txTypes, func(i, j int) bool { return txTypes[i] < txTypes[j] })

	// Print types in sorted order
	for _, txType := range txTypes {
		fmt.Printf("  Type %d (%s): %d transaction(s)\n",
			txType, getTxTypeDescription(txType), txTypeCounts[txType])
	}

	// Compute the receipt root locally
	localRoot := computeReceiptRoot(receipts)
	fmt.Printf("\nLocally computed receipt root: %x\n", localRoot)

	// Compare the roots
	if block.ReceiptHash() == localRoot {
		fmt.Println("✅ Receipt roots match!")
	} else {
		fmt.Println("❌ Receipt roots don't match!")
	}
}

func getReceiptsForBlock(client *ethclient.Client, block *types.Block) (types.Receipts, error) {
	// Create BlockNumberOrHash from block hash
	blockHash := block.Hash()
	blockNumberOrHash := rpc.BlockNumberOrHashWithHash(blockHash, true)

	// Try to get all receipts for the block in a single call
	receipts, err := client.BlockReceipts(context.Background(), blockNumberOrHash)
	fmt.Printf("  ⚠️ BlockReceipts failed: %v\n", err)
	fmt.Printf("  ℹ️ Falling back to individual receipt fetching...\n")
	if err != nil {
		// If BlockReceipts fails (possibly not supported by the node), fall back to individual receipt fetching
		fmt.Printf("  ⚠️ BlockReceipts failed: %v\n", err)
		fmt.Printf("  ℹ️ Falling back to individual receipt fetching...\n")
		return fetchIndividualReceipts(client, block)
	}

	// Check if we have a valid response
	if len(receipts) == 0 {
		// No receipts returned, try the fallback method
		fmt.Printf("  ⚠️ No receipts returned from BlockReceipts\n")
		fmt.Printf("  ℹ️ Falling back to individual receipt fetching...\n")
		return fetchIndividualReceipts(client, block)
	}

	// Check if receipts count matches transaction count
	txCount := len(block.Transactions())
	// Celo sometimes has an extra receipt for the block (known behavior)
	if len(receipts) != txCount && len(receipts) != txCount+1 {
		fmt.Printf("  ⚠️ Warning: Receipt count mismatch - got %d receipts for %d transactions\n",
			len(receipts), txCount)
	} else {
		fmt.Printf("  ℹ️ Successfully retrieved %d receipts with BlockReceipts\n", len(receipts))
	}

	return receipts, nil
}

// fetchIndividualReceipts fetches transaction receipts one by one
func fetchIndividualReceipts(client *ethclient.Client, block *types.Block) (types.Receipts, error) {
	var receipts types.Receipts
	var firstErr error
	var successCount int

	for i, tx := range block.Transactions() {
		receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			fmt.Printf("  ⚠️ Failed to get receipt for tx %d (%s): %v\n",
				i, tx.Hash().Hex()[:10]+"...", err)
			continue
		}
		successCount++
		receipts = append(receipts, receipt)
	}

	if len(receipts) == 0 && firstErr != nil {
		return nil, fmt.Errorf("couldn't get any receipts: %v", firstErr)
	}

	fmt.Printf("  ℹ️ Successfully retrieved %d/%d receipts individually\n", successCount, len(block.Transactions()))
	return receipts, nil
}

func computeReceiptRoot(receipts types.Receipts) common.Hash {
	// calculate the root before generating the proof or the proof will be invalid
	return types.DeriveSha(receipts, trie.NewStackTrie(nil))
}

func printReceiptDetails(receipt *types.Receipt) {
	// prnt receipt object as json
	// receiptJSON, err := receipt.MarshalJSON()
	// if err != nil {
	// 	fmt.Printf("Error marshalling receipt: %v\n", err)
	// 	return
	// }

	// fmt.Printf("Receipt JSON: %s\n", receiptJSON)

	fmt.Printf("  Type: %d\n", receipt.Type)
	fmt.Printf("  Status: %d\n", receipt.Status)
	fmt.Printf("  CumulativeGasUsed: %d\n", receipt.CumulativeGasUsed)
	fmt.Printf("  GasUsed: %d\n", receipt.GasUsed)
	fmt.Printf("  BlockHash: %s\n", receipt.BlockHash.Hex())
	fmt.Printf("  TxHash: %s\n", receipt.TxHash.Hex())
	fmt.Printf("  ContractAddress: %s\n", receipt.ContractAddress.Hex())
	fmt.Printf("  Logs: %d entries\n", len(receipt.Logs))

	// Try to decode the Celo-specific fields
	fmt.Printf("  Has BaseFee: %v\n", receipt.BaseFee != nil)
	if receipt.BaseFee != nil {
		fmt.Printf("  BaseFee: %s\n", receipt.BaseFee.String())
	}
}

// countTxTypes counts the number of each transaction type in the receipts
func countTxTypes(receipts types.Receipts) map[uint8]int {
	txTypes := make(map[uint8]int)
	for _, receipt := range receipts {
		fmt.Printf("Type: %d\n", receipt.Type)
		txTypes[receipt.Type]++
	}
	return txTypes
}

// getTxTypeDescription returns a human-readable description of a transaction type
func getTxTypeDescription(txType uint8) string {
	switch txType {
	case types.LegacyTxType: // 0x00
		return "Legacy (pre-EIP2718)"
	case types.AccessListTxType: // 0x01
		return "AccessList (EIP2930)"
	case types.DynamicFeeTxType: // 0x02
		return "DynamicFee (EIP1559)"
	case types.BlobTxType: // 0x03
		return "Blob (EIP4844)"
	case types.SetCodeTxType: // 0x04
		return "SetCode"
	case 0x79: // CeloDynamicFeeTxType
		return "CeloDynamicFee"
	case 0x7A: // CeloDynamicFeeTxV2Type
		return "CeloDynamicFeeV2"
	case 0x7D: // CeloDenominatedTxType
		return "CeloDenominated"
	default:
		return "Unknown"
	}
}
