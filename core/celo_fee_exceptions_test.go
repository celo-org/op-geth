package core

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
)

// TestFeeCurrencyMaxFeeExceptionTxHash rebuilds the canonical mainnet transaction from
// its on-chain fields and checks that it hashes to the pinned exception entry, so the
// exception cannot silently point at nothing.
//
// Celo Mainnet block 75046581, tx index 29.
func TestFeeCurrencyMaxFeeExceptionTxHash(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
