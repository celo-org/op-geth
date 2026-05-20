package beacon

import (
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/consensus/misc/eip1559"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/superchain"
)

type testHeaderChain struct {
	config *params.ChainConfig
	head   *types.Header
}

func (t *testHeaderChain) Config() *params.ChainConfig  { return t.config }
func (t *testHeaderChain) CurrentHeader() *types.Header { return t.head }
func (t *testHeaderChain) GetHeader(hash common.Hash, number uint64) *types.Header {
	if t.head.Number.Uint64() == number && t.head.Hash() == hash {
		return t.head
	}
	return nil
}
func (t *testHeaderChain) GetHeaderByNumber(number uint64) *types.Header {
	if t.head.Number.Uint64() == number {
		return t.head
	}
	return nil
}
func (t *testHeaderChain) GetHeaderByHash(hash common.Hash) *types.Header {
	if t.head.Hash() == hash {
		return t.head
	}
	return nil
}

func testJovianExtra(config *params.ChainConfig, timestamp uint64) []byte {
	minBaseFee := uint64(1)
	return eip1559.EncodeOptimismExtraData(config, timestamp, 250, 6, &minBaseFee)
}

func testCeloMainnetConfig(t *testing.T) *params.ChainConfig {
	t.Helper()

	chainCfg, err := superchain.GetChain(params.CeloMainnetChainID)
	if err != nil {
		t.Fatal(err)
	}
	opCfg, err := chainCfg.Config()
	if err != nil {
		t.Fatal(err)
	}
	config, err := params.LoadOPStackChainConfig(opCfg)
	if err != nil {
		t.Fatal(err)
	}
	return config
}

func TestVerifyHeadersDoesNotUseInvalidBatchParent(t *testing.T) {
	config := testCeloMainnetConfig(t)

	zero := uint64(0)
	beaconRoot := common.Hash{}
	withdrawalsRoot := types.EmptyWithdrawalsHash

	parent := &types.Header{
		Number:           big.NewInt(31_056_500),
		Difficulty:       common.Big0,
		UncleHash:        types.EmptyUncleHash,
		Time:             1_774_959_000,
		GasLimit:         30_000_000,
		GasUsed:          0,
		BaseFee:          big.NewInt(25_000_000_000),
		Extra:            testJovianExtra(config, 1_774_959_000),
		WithdrawalsHash:  &withdrawalsRoot,
		ParentBeaconRoot: &beaconRoot,
		ExcessBlobGas:    &zero,
		BlobGasUsed:      &zero,
	}
	first := &types.Header{
		ParentHash:       parent.Hash(),
		Number:           big.NewInt(31_056_501),
		Difficulty:       common.Big0,
		UncleHash:        types.EmptyUncleHash,
		Time:             1_774_959_002,
		GasLimit:         parent.GasLimit,
		GasUsed:          0,
		Extra:            testJovianExtra(config, 1_774_959_002),
		WithdrawalsHash:  &withdrawalsRoot,
		ParentBeaconRoot: &beaconRoot,
		ExcessBlobGas:    &zero,
		BlobGasUsed:      &zero,
	}
	second := &types.Header{
		ParentHash:       first.Hash(),
		Number:           big.NewInt(31_056_502),
		Difficulty:       common.Big0,
		UncleHash:        types.EmptyUncleHash,
		Time:             1_774_959_004,
		GasLimit:         first.GasLimit,
		GasUsed:          0,
		BaseFee:          big.NewInt(25_000_000_000),
		Extra:            testJovianExtra(config, 1_774_959_004),
		WithdrawalsHash:  &withdrawalsRoot,
		ParentBeaconRoot: &beaconRoot,
		ExcessBlobGas:    &zero,
		BlobGasUsed:      &zero,
	}

	chain := &testHeaderChain{config: config, head: parent}
	engine := New(ethash.NewFaker())
	_, results := engine.VerifyHeaders(chain, []*types.Header{first, second})

	select {
	case err := <-results:
		if err == nil || err.Error() != "header is missing baseFee" {
			t.Fatalf("first header error mismatch: have %v, want missing baseFee", err)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for first header result")
	}
	select {
	case err := <-results:
		if !errors.Is(err, consensus.ErrUnknownAncestor) {
			t.Fatalf("second header error mismatch: have %v, want %v", err, consensus.ErrUnknownAncestor)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for second header result")
	}
}
