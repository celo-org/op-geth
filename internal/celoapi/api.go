package celoapi

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/internal/ethapi"
)

type Ethereum interface {
	BlockChain() *core.BlockChain
}

type CeloAPI struct {
	ethAPI      *ethapi.EthereumAPI
	celoBackend ethapi.CeloBackend
	eth         Ethereum
}

func NewCeloAPI(e Ethereum, b ethapi.CeloBackend) *CeloAPI {
	return &CeloAPI{
		ethAPI:      ethapi.NewEthereumAPI(b),
		celoBackend: b,
		eth:         e,
	}
}

func (c *CeloAPI) convertedCurrencyValue(ctx context.Context, v *hexutil.Big, feeCurrency *common.Address) (*hexutil.Big, error) {
	if feeCurrency != nil {
		// retrieve the latest head
		header := c.eth.BlockChain().CurrentBlock()
		if header != nil {
			return nil, errors.New("no latest header retrieved")
		}
		convertedTipCap, err := c.celoBackend.ConvertToCurrency(ctx, header.Hash(), v.ToInt(), feeCurrency)
		if err != nil {
			return nil, err
		}
		v = (*hexutil.Big)(convertedTipCap)
	}
	return v, nil
}

// GasPrice wraps the original JSON RPC `eth_gasPrice` and adds an additional
// optional parameter `feeCurrency` for fee-currency conversion.
// When `feeCurrency` is not given, then the original JSON RPC method is called without conversion.
func (c *CeloAPI) GasPrice(ctx context.Context, feeCurrency *common.Address) (*hexutil.Big, error) {
	tipcap, err := c.ethAPI.GasPrice(ctx)
	if err != nil {
		return nil, err
	}
	// Between the call to `ethapi.GasPrice` and the call to fetch and convert the rates,
	// there is a chance of a state-change. This means that gas-price suggestion is calculated
	// based on state of block x, while the currency conversion could be calculated based on block
	// x+1.
	// However, a similar race condition is present in the `ethapi.GasPrice` method itself.
	return c.convertedCurrencyValue(ctx, tipcap, feeCurrency)
}

// MaxPriorityFeePerGas wraps the original JSON RPC `eth_maxPriorityFeePerGas` and adds an additional
// optional parameter `feeCurrency` for fee-currency conversion.
// When `feeCurrency` is not given, then the original JSON RPC method is called without conversion.
func (c *CeloAPI) MaxPriorityFeePerGas(ctx context.Context, feeCurrency *common.Address) (*hexutil.Big, error) {
	tipcap, err := c.ethAPI.MaxPriorityFeePerGas(ctx)
	if err != nil {
		return nil, err
	}
	// Between the call to `ethapi.MaxPriorityFeePerGas` and the call to fetch and convert the rates,
	// there is a chance of a state-change. This means that gas-price suggestion is calculated
	// based on state of block x, while the currency conversion could be calculated based on block
	// x+1.
	return c.convertedCurrencyValue(ctx, tipcap, feeCurrency)
}
