package catalyst

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const (
	// PostEcotoneGasParamsLength is the length of gas parameters after the Ecotone fork
	// See - types.extractL1GasParamsPostEcotone
	PostEcotoneGasParamsLength = 164
)

func (c *SimulatedBeacon) payloadGasLimit() *uint64 {
	if c.eth.BlockChain().Config().Optimism == nil {
		return nil
	}
	// If Optimism config is set we need to set the gas limit in the payload attributes.
	return &c.eth.BlockChain().CurrentBlock().GasLimit
}

func (c *SimulatedBeacon) payloadSystemTransaction() ([][]byte, error) {
	// Post ecotone we need to provide a system transaction to set L1 gas params, we don't actually need to
	// set realistic values for the gas params, since in Celo we ensure L1 gas fees of zero.
	// However this is required in order to be able to correctly derive receipt fields.
	// See types.Receipts.DeriveFields.
	if c.eth.BlockChain().Config().Optimism != nil && c.eth.BlockChain().Config().IsEcotone(c.eth.BlockChain().CurrentBlock().Time) {
		sysTx := &types.DepositTx{
			SourceHash:          common.Hash{},
			From:                common.Address{},
			To:                  &common.Address{},
			Mint:                nil,
			Value:               big.NewInt(0),
			Gas:                 50000,
			IsSystemTransaction: true,
			Data:                make([]byte, PostEcotoneGasParamsLength),
		}

		l1Tx := types.NewTx(sysTx)
		systemTxBytes, err := l1Tx.MarshalBinary()
		if err != nil {
			return nil, err
		}
		return [][]byte{systemTxBytes}, nil
	}
	return nil, nil
}
