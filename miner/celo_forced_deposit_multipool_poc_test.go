package miner

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/beacon"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/consensus/misc/eip1559"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
)

// TestCeloForcedDepositGasLimitPoisonsNativeMultiGasPool is a regression test
// for the forced-deposit MultiGasPool debit. The native (non-fee-currency) pool
// must be charged actual gas used, not the (possibly inflated) deposit gas
// limit, so that native CELO txpool selection is not throttled by unused gas.
func TestCeloForcedDepositGasLimitPoisonsNativeMultiGasPool(t *testing.T) {
	build := func(t *testing.T, depositGas uint64) *types.Block {
		t.Helper()

		db := rawdb.NewMemoryDatabase()
		miner, backend := newTestWorker(t, jovianConfig(), beacon.New(ethash.NewFaker()), db, 0)
		parent := backend.chain.CurrentBlock()
		timestamp := parent.Time + 12
		l1BlockAddr := types.L1BlockAddr
		deposit := jovianDepositTx(testDAFootprintGasScalar)
		deposit.SourceHash = common.HexToHash("0x01")
		deposit.From = testUserAddress
		deposit.To = &l1BlockAddr
		deposit.Value = big.NewInt(0)
		deposit.Gas = depositGas

		result := miner.generateWork(&generateParams{
			parentHash:    parent.Hash(),
			timestamp:     timestamp,
			withdrawals:   types.Withdrawals{},
			beaconRoot:    new(common.Hash),
			gasLimit:      ptr(uint64(1_000_000)),
			txs:           types.Transactions{types.NewTx(deposit)},
			eip1559Params: eip1559.EncodeHolocene1559Params(250, 6),
			minBaseFee:    new(uint64),
		}, false)
		require.NoError(t, result.err)
		require.NotNil(t, result.block)
		return result.block
	}

	control := build(t, 100_000)
	t.Logf("control: txs=%d gasUsed=%d gasRemaining=%d", len(control.Transactions()), control.GasUsed(), control.GasLimit()-control.GasUsed())
	require.Len(t, control.Transactions(), 2)

	poisoned := build(t, 1_000_000)
	t.Logf("poisoned: txs=%d gasUsed=%d gasRemaining=%d", len(poisoned.Transactions()), poisoned.GasUsed(), poisoned.GasLimit()-poisoned.GasUsed())
	require.Len(t, poisoned.Transactions(), 2)
	require.Greater(t, poisoned.GasLimit()-poisoned.GasUsed(), params.TxGas)
}

func TestCeloForcedDepositGasLimitCapsNativeThroughput(t *testing.T) {
	const nativeTxCount = 700

	simpleNativeTxs := func(startNonce, count uint64) types.Transactions {
		signer := types.LatestSigner(params.TestChainConfig)
		txs := make(types.Transactions, 0, count)
		for nonce := startNonce; nonce < startNonce+count; nonce++ {
			txs = append(txs, types.MustSignNewTx(testBankKey, signer, &types.LegacyTx{
				Nonce:    nonce,
				To:       &testUserAddress,
				Value:    big.NewInt(1000),
				Gas:      params.TxGas,
				GasPrice: big.NewInt(params.InitialBaseFee),
			}))
		}
		return txs
	}

	build := func(t *testing.T, userDepositGas uint64) *types.Block {
		t.Helper()

		db := rawdb.NewMemoryDatabase()
		miner, backend := newTestWorker(t, jovianConfig(), beacon.New(ethash.NewFaker()), db, 0)
		for _, err := range backend.txPool.Add(simpleNativeTxs(1, nativeTxCount), false) {
			require.NoError(t, err)
		}

		parent := backend.chain.CurrentBlock()
		l1BlockAddr := types.L1BlockAddr
		l1Attributes := jovianDepositTx(testDAFootprintGasScalar)
		l1Attributes.SourceHash = common.HexToHash("0x01")
		l1Attributes.From = common.HexToAddress("0xdeaddeaddeaddeaddeaddeaddeaddeaddead0001")
		l1Attributes.To = &l1BlockAddr
		l1Attributes.Value = big.NewInt(0)
		l1Attributes.Gas = 1_000_000

		userDepositTo := testRecipient
		userDeposit := &types.DepositTx{
			SourceHash: common.HexToHash("0x02"),
			From:       testUserAddress,
			To:         &userDepositTo,
			Value:      big.NewInt(0),
			Gas:        userDepositGas,
		}

		result := miner.generateWork(&generateParams{
			parentHash:    parent.Hash(),
			timestamp:     parent.Time + 12,
			withdrawals:   types.Withdrawals{},
			beaconRoot:    new(common.Hash),
			gasLimit:      ptr(uint64(30_000_000)),
			txs:           types.Transactions{types.NewTx(l1Attributes), types.NewTx(userDeposit)},
			eip1559Params: eip1559.EncodeHolocene1559Params(250, 6),
			minBaseFee:    new(uint64),
		}, false)
		require.NoError(t, result.err)
		require.NotNil(t, result.block)
		return result.block
	}

	control := build(t, 100_000)
	poisoned := build(t, 20_000_000)
	controlNativeTxs := len(control.Transactions()) - 2
	poisonedNativeTxs := len(poisoned.Transactions()) - 2

	t.Logf("control: nativeTxs=%d gasUsed=%d gasRemaining=%d", controlNativeTxs, control.GasUsed(), control.GasLimit()-control.GasUsed())
	t.Logf("poisoned: nativeTxs=%d gasUsed=%d gasRemaining=%d", poisonedNativeTxs, poisoned.GasUsed(), poisoned.GasLimit()-poisoned.GasUsed())

	// With the fix in place, a forced deposit carrying a large gas limit but
	// consuming little gas must not reduce native CELO throughput: the native
	// MultiGasPool is charged actual gas used, matching the real gas pool.
	require.Equal(t, controlNativeTxs, poisonedNativeTxs)
}
