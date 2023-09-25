package e2e_test

import (
	"context"
	"math/big"
	"os/exec"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/celo/abigen"
	"github.com/ethereum/go-ethereum/contracts/config"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"
)

func waitForTx(client *ethclient.Client, tx *types.Transaction) {
	for {
		receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
		if err == nil && receipt.Status == 1 {
			break
		}
		time.Sleep(100 * time.Microsecond)
	}
}

// waitForEndpoint attempts to connect to an RPC endpoint until it succeeds.
// Copied from cmd/geth/run_test.go in order to avoid changing existing file
func waitForEndpoint(t *testing.T, endpoint string, timeout time.Duration) {
	probe := func() bool {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		c, err := rpc.DialContext(ctx, endpoint)
		if c != nil {
			_, err = c.SupportedModules()
			c.Close()
		}
		return err == nil
	}

	start := time.Now()
	for {
		if probe() {
			return
		}
		if time.Since(start) > timeout {
			t.Fatal("endpoint", endpoint, "did not open within", timeout)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func getClient(ctx context.Context, t *testing.T) *ethclient.Client {
	cmd := exec.CommandContext(ctx, "../build/bin/geth", "--dev", "--http", "--http.api", "eth,web3,net")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	endpoint := "http://127.0.0.1:8545"
	waitForEndpoint(t, endpoint, 1000*time.Millisecond)
	client, err := ethclient.Dial(endpoint)
	if err != nil {
		t.Fatal(err)
	}

	return client
}

func TestTokenDuality(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := getClient(ctx, t)
	auth, err := bind.NewKeyedTransactorWithChainID(core.DevPrivateKey, big.NewInt(1337))
	if err != nil {
		t.Fatal(err)
	}

	// Verify that balance is zero before transfer
	targetAddress := common.HexToAddress("0x1dead")
	balance, err := client.BalanceAt(ctx, targetAddress, nil)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, int64(0), balance.Int64())

	// Initialize contract wrappers
	goldTokenAddress := common.HexToAddress("0xce12")
	goldToken, err := abigen.NewGoldToken(goldTokenAddress, client)
	if err != nil {
		t.Fatal(err)
	}
	registry, err := abigen.NewRegistry(config.RegistrySmartContractAddress, client)
	if err != nil {
		t.Fatal(err)
	}

	// Register GoldToken
	if _, err := registry.SetAddressFor(auth, "GoldToken", goldTokenAddress); err != nil {
		t.Fatal(err)
	}

	// Transfer and check result
	tx, err := goldToken.Transfer(auth, targetAddress, big.NewInt(101))
	if err != nil {
		t.Fatal(err)
	}
	waitForTx(client, tx)
	balance, err = client.BalanceAt(ctx, targetAddress, nil)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, int64(101), balance.Int64())
}
