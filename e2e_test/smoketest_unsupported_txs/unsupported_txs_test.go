//go:build smoketest

package smoketestunsupportedtxs

import (
	"context"
	"crypto/ecdsa"
	"flag"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
)

var (
	privateKey   string
	opGethRpcURL string
	feeCurrency  string
	deadAddr     = common.HexToAddress("0x00000000000000000000000000000000DeaDBeef")
)

func init() {
	// Define your custom flag
	flag.StringVar(&privateKey, "private-key", "", "private key of transaction sender")
	flag.StringVar(&opGethRpcURL, "op-geth-url", "", "op-geth rpc url")
	flag.StringVar(&feeCurrency, "fee-currency", "", "address of the fee currency to use")
}

func TestTxSendingFails(t *testing.T) {
	key, err := parsePrivateKey(privateKey)
	require.NoError(t, err)

	feeCurrencyAddr := common.HexToAddress(feeCurrency)

	client, err := ethclient.Dial(opGethRpcURL)
	require.NoError(t, err)

	chainId, err := client.ChainID(context.Background())
	require.NoError(t, err)

	// Get a signer that can sign deprecated txs, we need cel2 configured but not active yet.
	cel2Time := uint64(1)
	signer := types.MakeSigner(&params.ChainConfig{Cel2Time: &cel2Time, ChainID: chainId}, big.NewInt(0), 0)

	t.Run("CeloLegacy", func(t *testing.T) {
		gasPrice, err := client.SuggestGasPriceForCurrency(context.Background(), &feeCurrencyAddr)
		require.NoError(t, err)

		nonce, err := client.PendingNonceAt(context.Background(), crypto.PubkeyToAddress(key.PublicKey))
		require.NoError(t, err)

		txdata := &types.LegacyTx{
			Nonce:       nonce,
			To:          &deadAddr,
			Gas:         100_000,
			GasPrice:    gasPrice,
			FeeCurrency: &feeCurrencyAddr,
			Value:       big.NewInt(1),
			CeloLegacy:  true,
		}

		tx, err := types.SignNewTx(key, signer, txdata)
		require.NoError(t, err)

		// we expect this to fail because the tx is not supported.
		err = client.SendTransaction(context.Background(), tx)
		require.Error(t, err)
	})

	t.Run("CIP42", func(t *testing.T) {
		gasFeeCap, err := client.SuggestGasPriceForCurrency(context.Background(), &feeCurrencyAddr)
		require.NoError(t, err)

		nonce, err := client.PendingNonceAt(context.Background(), crypto.PubkeyToAddress(key.PublicKey))
		require.NoError(t, err)

		txdata := &types.CeloDynamicFeeTx{
			Nonce:       nonce,
			To:          &deadAddr,
			Gas:         100_000,
			GasFeeCap:   gasFeeCap,
			FeeCurrency: &feeCurrencyAddr,
			Value:       big.NewInt(1),
		}

		tx, err := types.SignNewTx(key, signer, txdata)
		require.NoError(t, err)

		// we expect this to fail because the tx is not supported.
		err = client.SendTransaction(context.Background(), tx)
		require.Error(t, err)
	})
}

func parsePrivateKey(privateKey string) (*ecdsa.PrivateKey, error) {
	if len(privateKey) >= 2 && privateKey[0] == '0' && (privateKey[1] == 'x' || privateKey[1] == 'X') {
		privateKey = privateKey[2:]
	}
	return crypto.HexToECDSA(privateKey)
}
