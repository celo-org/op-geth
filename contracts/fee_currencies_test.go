package contracts

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/triedb"
)

func TestGetFeeBalanceReturnsZeroForInvalidCurrency(t *testing.T) {
	db := state.NewDatabase(triedb.NewDatabase(rawdb.NewMemoryDatabase(), triedb.HashDefaults), nil)
	stateDB, _ := state.New(common.Hash{}, db)

	backend := &CeloBackend{
		ChainConfig: params.TestChainConfig,
		State:       stateDB,
	}

	bogusAddr := common.HexToAddress("0x0000000000000000000000000000000000000099")
	account := common.HexToAddress("0x0000000000000000000000000000000000000001")

	balance := GetFeeBalance(backend, account, &bogusAddr)
	if balance == nil {
		t.Fatal("GetFeeBalance returned nil, expected zero")
	}
	if balance.Cmp(new(big.Int)) != 0 {
		t.Fatalf("GetFeeBalance returned %s, expected 0", balance.String())
	}
}
