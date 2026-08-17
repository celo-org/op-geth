package core

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/consensus/beacon"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/contracts"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
)

// TestFeeCurrencyMaxFeeExceptionTxHash rebuilds the canonical mainnet transaction from
// its on-chain fields and checks that it hashes to the pinned exception entry, so the
// exception cannot silently point at nothing.
//
// Celo Mainnet block 75046581, tx index 29.
func TestFeeCurrencyMaxFeeExceptionTxHash(t *testing.T) {
	// Not parallel: TestFeeCurrencyMaxFeeExceptionSkipsTheCheck swaps out
	// feeCurrencyMaxFeeExceptions, which this test reads.

	require.Len(t, feeCurrencyMaxFeeExceptions, 1, "update this test when an entry is added")
	pinned := feeCurrencyMaxFeeExceptions[0]

	hexToBigInt := func(s string) *big.Int {
		b, ok := new(big.Int).SetString(s, 0)
		require.True(t, ok, "failed to parse %q", s)
		return b
	}

	to := common.HexToAddress("0x48065fbbe25f71c9282ddf5e1cd6d6a887483d5e")
	feeCurrency := common.HexToAddress("0x0E2A3e05bc9A16F5292A6170456A710cb89C6f72")
	data, err := hexutil.Decode("0xa9059cbb00000000000000000000000003d904993bf9a19e3fe78bed6bec435603d2e6150000000000000000000000000000000000000000000000000000000000105f46")
	require.NoError(t, err)

	tx := types.NewTx(&types.CeloDynamicFeeTxV2{
		ChainID:     big.NewInt(params.CeloMainnetChainID),
		Nonce:       0x1d,
		GasTipCap:   big.NewInt(150487100),   // 0x8f8403c
		GasFeeCap:   big.NewInt(29097136819), // 0x6c65312b3
		Gas:         200000,                  // 0x30d40
		To:          &to,
		Value:       big.NewInt(0),
		Data:        data,
		AccessList:  types.AccessList{},
		FeeCurrency: &feeCurrency,
		V:           big.NewInt(1),
		R:           hexToBigInt("0x3c6f794c1d0150430f72a84840900b64f3095bac2329b22d427cd2ce1845d5c5"),
		S:           hexToBigInt("0x6ae569be2db73db63ec20ebf618098c130949ca79b4a34cadcfe60cfb1992035"),
	})

	hash := tx.Hash()
	require.Equal(t, pinned.txHash, hash)
	require.Equal(t, uint64(params.CeloMainnetChainID), pinned.chainID)
	require.Equal(t, uint64(75046581), pinned.blockNumber)

	// The signature is part of the hash, so recovering the sender ties R/S to the
	// account that actually sent the transaction on mainnet.
	sender, err := types.Sender(types.LatestSignerForChainID(big.NewInt(params.CeloMainnetChainID)), tx)
	require.NoError(t, err)
	require.Equal(t, common.HexToAddress("0x1De6939e8A03DF7bDc970A951B67628d2D138eD9"), sender)

	// The rebuilt transaction is the one the exception is meant to cover, and it is
	// only exempt on the network and at the height it is canonical on.
	require.True(t, isFeeCurrencyMaxFeeException(hash, big.NewInt(params.CeloMainnetChainID), big.NewInt(75046581)))
	require.False(t, isFeeCurrencyMaxFeeException(hash, big.NewInt(params.CeloSepoliaChainID), big.NewInt(75046581)))
	require.False(t, isFeeCurrencyMaxFeeException(hash, big.NewInt(params.CeloMainnetChainID), big.NewInt(75046582)))

	// The sender must still not be able to afford the max fee, otherwise the exception
	// is pointless and the plain check would already let the transaction through.
	balance := big.NewInt(5295000000000000)
	maxFee := new(big.Int).Mul(big.NewInt(int64(tx.Gas())), tx.GasFeeCap())
	require.Equal(t, "5819427363800000", maxFee.String())
	require.Negative(t, balance.Cmp(maxFee))

	// ... while the amount actually debited is affordable.
	effectiveFee := new(big.Int).Mul(big.NewInt(int64(tx.Gas())), big.NewInt(12161345100))
	require.Equal(t, "2432269020000000", effectiveFee.String())
	require.Positive(t, balance.Cmp(effectiveFee))
}

func TestIsFeeCurrencyMaxFeeException(t *testing.T) {
	// Not parallel: TestFeeCurrencyMaxFeeExceptionSkipsTheCheck swaps out
	// feeCurrencyMaxFeeExceptions, which this test reads.

	require.Len(t, feeCurrencyMaxFeeExceptions, 1, "update this test when an entry is added")
	pinned := feeCurrencyMaxFeeExceptions[0]
	known := pinned.txHash
	chainID := new(big.Int).SetUint64(pinned.chainID)
	block := new(big.Int).SetUint64(pinned.blockNumber)

	require.True(t, isFeeCurrencyMaxFeeException(known, chainID, block))

	// Every part of the key has to match.
	require.False(t, isFeeCurrencyMaxFeeException(known, big.NewInt(params.CeloSepoliaChainID), block))
	require.False(t, isFeeCurrencyMaxFeeException(known, big.NewInt(1), block))
	require.False(t, isFeeCurrencyMaxFeeException(known, chainID, new(big.Int).Sub(block, common.Big1)))
	require.False(t, isFeeCurrencyMaxFeeException(known, chainID, new(big.Int).Add(block, common.Big1)))
	require.False(t, isFeeCurrencyMaxFeeException(known, chainID, common.Big0))

	// Messages synthesized from call arguments carry no transaction hash and must never
	// match an exception.
	require.False(t, isFeeCurrencyMaxFeeException(common.Hash{}, chainID, block))
	require.False(t, isFeeCurrencyMaxFeeException(common.HexToHash("0xdead"), chainID, block))

	// A nil or oversized chain ID or block number must not panic or match.
	huge := new(big.Int).Lsh(big.NewInt(1), 70)
	require.False(t, isFeeCurrencyMaxFeeException(known, nil, block))
	require.False(t, isFeeCurrencyMaxFeeException(known, chainID, nil))
	require.False(t, isFeeCurrencyMaxFeeException(known, huge, block))
	require.False(t, isFeeCurrencyMaxFeeException(known, chainID, huge))
}

// TestFeeCurrencyMaxFeeExceptionSkipsTheCheck is the behavioural counterpart to the unit
// tests above: it reproduces the mainnet halt and shows the exception clearing it.
//
// A CIP-64 transaction is built whose sender holds a fee currency balance in
// [gasLimit * effectiveGasPrice, gasLimit * maxFeePerGas) - the window where celo-reth and
// op-geth disagreed. The block containing it is generated with the exception pinned, so it
// is the canonical-but-lenient block op-geth has to accept, and then inserted twice:
//
//  1. with the exception pinned, the block is accepted and the fee currency debit still
//     runs, which is the constraint celo-reth applies;
//  2. without it, op-geth rejects the very same block - exactly the failure that stalled
//     mainnet at block 75046581.
func TestFeeCurrencyMaxFeeExceptionSkipsTheCheck(t *testing.T) {
	// Not parallel: this test swaps out the package-level exception table.
	defer func(saved []feeCurrencyMaxFeeException) {
		feeCurrencyMaxFeeExceptions = saved
	}(feeCurrencyMaxFeeExceptions)

	var (
		aa      = common.HexToAddress("0x000000000000000000000000000000000000aaaa")
		engine  = beacon.New(ethash.NewFaker())
		key1, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
		addr1   = crypto.PubkeyToAddress(key1.PublicKey)
		// Worth twice as much as native CELO, so the base fee halves when converted.
		feeCurrencyAddr = DevFeeCurrencyAddr2
		config          = *params.AllEthashProtocolChanges
		gspec           = &Genesis{
			Config: &config,
			Alloc:  CeloGenesisAccounts(addr1),
		}

		// The sender holds DevBalance of the fee currency. gasLimit * maxFeePerGas is
		// twice that, so the max fee check cannot pass, while the effective price is
		// small enough that the debit comfortably can.
		gasLimit  = uint64(100000)
		gasFeeCap = big.NewInt(2_000_000_000_000_000)
		gasTipCap = big.NewInt(2)
	)
	gspec.Config.Cel2Time = uint64ptr(0)
	gspec.Config.BedrockBlock = big.NewInt(0)
	gspec.Config.Optimism = &params.OptimismConfig{EIP1559Elasticity: 2, EIP1559Denominator: 8}
	gspec.Config.Celo = &params.CeloConfig{EIP1559BaseFeeFloor: 250000000}
	gspec.Config.Ethash = nil

	maxFee := new(big.Int).Mul(new(big.Int).SetUint64(gasLimit), gasFeeCap)
	require.Negative(t, DevBalance.Cmp(maxFee), "the max fee must be unaffordable, or the exception is untested")

	signer := types.LatestSigner(gspec.Config)

	// Generate the block with the exception in place, so that block production itself is
	// not what rejects the transaction.
	var tx *types.Transaction
	_, blocks, _ := GenerateChainWithGenesis(gspec, engine, 1, func(i int, b *BlockGen) {
		b.SetCoinbase(common.Address{1})

		var err error
		tx, err = types.SignTx(types.NewTx(&types.CeloDynamicFeeTxV2{
			ChainID:     gspec.Config.ChainID,
			Nonce:       0,
			To:          &aa,
			Gas:         gasLimit,
			GasFeeCap:   gasFeeCap,
			GasTipCap:   gasTipCap,
			FeeCurrency: &feeCurrencyAddr,
		}), signer, key1)
		require.NoError(t, err)

		feeCurrencyMaxFeeExceptions = []feeCurrencyMaxFeeException{{
			txHash:      tx.Hash(),
			chainID:     gspec.Config.ChainID.Uint64(),
			blockNumber: b.Number().Uint64(),
		}}

		b.AddTx(tx)
	})
	require.Len(t, blocks, 1)
	require.Len(t, blocks[0].Transactions(), 1, "the transaction must have made it into the block")

	insert := func() error {
		chain, err := NewBlockChain(rawdb.NewMemoryDatabase(), gspec, engine, DefaultConfig())
		require.NoError(t, err)
		defer chain.Stop()

		if _, err := chain.InsertChain(blocks); err != nil {
			return err
		}

		// The exception skips the max fee check and nothing else, so the fee currency
		// debit must still have happened.
		state, err := chain.State()
		require.NoError(t, err)
		head := chain.CurrentBlock()
		backend := contracts.CeloBackend{
			ChainConfig: chain.chainConfig,
			State:       state,
			BlockNumber: head.Number,
			Time:        head.Time,
		}
		balance, err := contracts.GetBalanceERC20(&backend, addr1, feeCurrencyAddr)
		require.NoError(t, err)

		paid := new(big.Int).Sub(DevBalance, balance)
		require.Positive(t, paid.Sign(), "the sender must have been debited")
		require.Negative(t, paid.Cmp(maxFee), "the sender must have paid less than the max fee it could not afford")
		return nil
	}

	// 1: the pinned transaction is accepted despite failing the max fee check.
	require.NoError(t, insert(), "the pinned block must be accepted")

	// 2: the same block without the exception is what stalled mainnet.
	feeCurrencyMaxFeeExceptions = nil
	err := insert()
	require.Error(t, err, "without the exception op-geth must reject the block")
	require.ErrorContains(t, err, ErrInsufficientFunds.Error())
	require.ErrorContains(t, err, "have "+DevBalance.String()+" want "+maxFee.String())
	require.ErrorContains(t, err, feeCurrencyAddr.Hex())
}
