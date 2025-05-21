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
	// Post ecotone we need to provide a system transaction to set L1 gas params.
	// This mechanism doesn't work in the STF, since no L2 contracts are pre-deployed.
	// The l1 gas-parameters are defaulting to 0, which is coincidentally what we expect in Celo.
	// This is why we don't need to set realistic values in the system transaction here, except for
	// the `Data` field, which is used by `types.Receipts.DeriveFields` and expected to be all 0 to match
	// what is deducted in the STF.
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
