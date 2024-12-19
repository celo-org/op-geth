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

func PopulatePreGingerbreadBlockFields(ctx context.Context, backend CeloBackend, block *types.Block) *types.Block {
	newHeader := PopulatePreGingerbreadHeaderFields(ctx, backend, block.Header())
	block = block.WithSeal(newHeader)
	return block
}

func PopulatePreGingerbreadHeaderFields(ctx context.Context, backend CeloBackend, header *types.Header) *types.Header {
	isGingerbread := backend.ChainConfig().IsGingerbread(header.Number)
	if isGingerbread {
		return header
	}

	var (
		baseFee  *big.Int
		gasLimit *big.Int
	)

	fields, err := rawdb.ReadPreGingerbreadAdditionalFields(backend.ChainDb(), header.Hash())
	if fields != nil {
		// If the record is found, use the values from the record
		baseFee = fields.BaseFee
		gasLimit = fields.GasLimit
	} else {
		if err != nil {
			log.Debug("failed to read pre-gingerbread fields", "err", err)
		}

		// If the record is not found, get the values and store them
		retrievedBaseFee, err := retrievePreGingerbreadBlockBaseFee(ctx, backend, header.Number)
		if err != nil {
			log.Debug("Not adding baseFeePerGas to RPC response, failed to retrieve gas price minimum", "block", header.Number.Uint64(), "err", err)
		} else {
			baseFee = retrievedBaseFee
		}

		gasLimit = new(big.Int).SetUint64(params.PreGingerbreadNetworkGasLimits[backend.ChainConfig().ChainID.Uint64()].Limit(header.Number))

		err = rawdb.WritePreGingerbreadAdditionalFields(backend.ChainDb(), header.Hash(), &rawdb.PreGingerbreadAdditionalFields{
			BaseFee:  baseFee,
			GasLimit: gasLimit,
		})
		if err != nil {
			log.Debug("failed to write pre-gingerbread fields", "err", err)
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

func retrievePreGingerbreadBlockBaseFee(ctx context.Context, backend CeloBackend, height *big.Int) (*big.Int, error) {
	if height.Cmp(common.Big0) <= 0 {
		return common.Big0, nil
	}

	prevBlock, err := backend.BlockByNumber(ctx, rpc.BlockNumber(height.Uint64()-1))
	if err != nil {
		return nil, err
	}
	if prevBlock == nil {
		return nil, fmt.Errorf("block #%d not found", height.Int64())
	}

	prevReceipts, err := backend.GetReceipts(ctx, prevBlock.Hash())
	if err != nil {
		return nil, err
	}

	numTxs, numReceipts := len(prevBlock.Transactions()), len(prevReceipts)
	if numReceipts <= numTxs {
		return nil, fmt.Errorf("receipts of block #%d do not contain system logs", height.Int64())
	}

	systemReceipt := prevReceipts[numTxs]
	for _, logRecord := range systemReceipt.Logs {
		baseFee, err := parseGasPriceMinimumUpdated(logRecord.Data)
		if err == nil {
			return baseFee, nil
		}
	}

	return nil, fmt.Errorf("gas price minimum updated event is not included in a receipt of block #%d", height.Int64())
}

func parseGasPriceMinimumUpdated(data []byte) (*big.Int, error) {
	values, err := gasPriceMinimumABI.Unpack("GasPriceMinimumUpdated", data)
	if err != nil {
		return nil, err
	}

	if len(values) != 1 {
		return nil, fmt.Errorf("unexpected format of values in GasPriceMinimumUpdated event")
	}

	baseFee, ok := values[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("unexpected base fee type in GasPriceMinimumUpdated event: expected *big.Int, got %T", values[0])
	}

	return baseFee, nil
}
