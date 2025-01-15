package ethapi

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
)

var (
	gasPriceMinimumABIJson = `[{"inputs":[{"internalType":"uint256","name":"gasPriceMinimum","type":"uint256"}],"name":"GasPriceMinimumUpdated","outputs":[],"type":"event"}]`
	gasPriceMinimumABI     abi.ABI
)

func init() {
	parsedAbi, _ := abi.JSON(strings.NewReader(gasPriceMinimumABIJson))
	gasPriceMinimumABI = parsedAbi
}

// PopulatePreGingerbreadBlockFields populates the baseFee and gasLimit fields of the block for pre-gingerbread blocks
func PopulatePreGingerbreadBlockFields(ctx context.Context, backend CeloBackend, block *types.Block) *types.Block {
	return block.WithSeal(
		PopulatePreGingerbreadHeaderFields(ctx, backend, block.Header()),
	)
}

// PopulatePreGingerbreadHeaderFields populates the baseFee and gasLimit fields of the header for pre-gingerbread blocks
func PopulatePreGingerbreadHeaderFields(ctx context.Context, backend CeloBackend, header *types.Header) *types.Header {
	// If the block is post-Gingerbread, return the header as is
	if backend.ChainConfig().IsGingerbread(header.Number) {
		return header
	}

	var (
		gasLimit *big.Int
		baseFee  *big.Int
		err      error
	)

	if chainId := backend.ChainConfig().ChainID; chainId != nil {
		gasLimit = retrievePreGingerbreadGasLimit(chainId.Uint64(), header.Number)
	}

	baseFee, err = rawdb.ReadPreGingerbreadBlockBaseFee(backend.ChainDb(), header.Hash())
	if err != nil {
		log.Debug("failed to load pre-Gingerbread block base fee from database", "block", header.Number.Uint64(), "err", err)
	}
	if baseFee == nil {
		// If the record is not found, get the values and store them
		baseFee, err = retrievePreGingerbreadBlockBaseFee(ctx, backend, header.Number)
		if err != nil {
			log.Debug("Not adding to RPC response, failed to retrieve pre-Gingerbread block base fee", "block", header.Number.Uint64(), "err", err)
		}

		// Store the base fee for future use
		if baseFee != nil {
			err = rawdb.WritePreGingerbreadBlockBaseFee(backend.ChainDb(), header.Hash(), baseFee)
			if err != nil {
				log.Debug("failed to write pre-Gingerbread block base fee", "block", header.Number.Uint64(), "err", err)
			}
		}
	}

	if baseFee != nil {
		header.BaseFee = baseFee
	}
	if gasLimit != nil {
		header.GasLimit = gasLimit.Uint64()
	}

	return header
}

// retrievePreGingerbreadGasLimit retrieves a gas limit at given height from hardcoded values
func retrievePreGingerbreadGasLimit(chainId uint64, height *big.Int) *big.Int {
	limits, ok := params.PreGingerbreadNetworkGasLimits[chainId]
	if !ok {
		log.Debug("Not adding gasLimit to RPC response, unknown network", "chainID", chainId)
		return nil
	}

	return new(big.Int).SetUint64(limits.Limit(height))
}

// retrievePreGingerbreadBlockBaseFee retrieves a base fee at given height from the previous block
func retrievePreGingerbreadBlockBaseFee(ctx context.Context, backend CeloBackend, height *big.Int) (*big.Int, error) {
	if height.Cmp(common.Big0) <= 0 {
		return nil, nil
	}

	prevHeight := height.Uint64() - 1
	prevBlock, err := backend.BlockByNumber(ctx, rpc.BlockNumber(prevHeight))
	if err != nil {
		return nil, err
	}
	if prevBlock == nil {
		return nil, fmt.Errorf("block #%d not found", prevHeight)
	}

	prevReceipts, err := backend.GetReceipts(ctx, prevBlock.Hash())
	if err != nil {
		return nil, err
	}

	numTxs, numReceipts := len(prevBlock.Transactions()), len(prevReceipts)
	if numReceipts <= numTxs {
		return nil, fmt.Errorf("receipts of block #%d don't contain system logs", prevHeight)
	}

	systemReceipt := prevReceipts[numTxs]
	for _, logRecord := range systemReceipt.Logs {
		if logRecord.Topics[0] != gasPriceMinimumABI.Events["GasPriceMinimumUpdated"].ID {
			continue
		}

		baseFee, err := parseGasPriceMinimumUpdated(logRecord.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to extract GasPriceMinimumUpdated event from system logs: %w", err)
		}

		return baseFee, nil
	}

	return nil, fmt.Errorf("an event GasPriceMinimumUpdated is not included in receipts of block #%d", prevHeight)
}

// parseGasPriceMinimumUpdated parses the data of GasPriceMinimumUpdated event
func parseGasPriceMinimumUpdated(data []byte) (*big.Int, error) {
	values, err := gasPriceMinimumABI.Unpack("GasPriceMinimumUpdated", data)
	if err != nil {
		return nil, err
	}

	// safe check, actually Unpack will parse first 32 bytes as a single value
	if len(values) != 1 {
		return nil, fmt.Errorf("unexpected format of values in GasPriceMinimumUpdated event")
	}

	baseFee, ok := values[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("unexpected base fee type in GasPriceMinimumUpdated event: expected *big.Int, got %T", values[0])
	}

	return baseFee, nil
}
