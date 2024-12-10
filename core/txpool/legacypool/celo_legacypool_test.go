package legacypool

import (
	"crypto/ecdsa"
	"errors"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/txpool"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/triedb"
)

func celoConfig(baseFeeFloor uint64) *params.ChainConfig {
	cpy := *params.TestChainConfig
	config := &cpy
	ct := uint64(0)
	config.Cel2Time = &ct
	config.Celo = &params.CeloConfig{EIP1559BaseFeeFloor: baseFeeFloor}
	return config
}

var (
	baseFeeFloor        = 100
	celoTestChainConfig = celoConfig(uint64(baseFeeFloor))
	// worth half as much as native celo
	feeCurrencyOne = core.DevFeeCurrencyAddr
	// worth twice as much as native celo
	feeCurrencyTwo          = core.DevFeeCurrencyAddr2
	feeCurrencyIntrinsicGas = core.FeeCurrencyIntrinsicGas
)

func pricedCip64Transaction(
	nonce uint64,
	gasLimit uint64,
	gasFeeCap *big.Int,
	gasTipCap *big.Int,
	feeCurrency *common.Address,
	key *ecdsa.PrivateKey,
) *types.Transaction {
	tx, _ := types.SignTx(types.NewTx(&types.CeloDynamicFeeTxV2{
		Nonce:       nonce,
		To:          &common.Address{},
		Value:       big.NewInt(100),
		Gas:         gasLimit,
		GasFeeCap:   gasFeeCap,
		GasTipCap:   gasTipCap,
		FeeCurrency: feeCurrency,
		Data:        nil,
	}), types.LatestSigner(celoTestChainConfig), key)
	return tx
}

func newDBWithCeloGenesis(config *params.ChainConfig, fundedAddress common.Address) (state.Database, *types.Block) {
	gspec := &core.Genesis{
		Config: config,
		Alloc:  core.CeloGenesisAccounts(fundedAddress),
	}
	db := rawdb.NewMemoryDatabase()
	triedb := triedb.NewDatabase(db, triedb.HashDefaults)
	defer triedb.Close()
	block, err := gspec.Commit(db, triedb)
	if err != nil {
		panic(err)
	}
	return state.NewDatabase(triedb, nil), block
}

func setupCeloPoolWithConfig(config *params.ChainConfig) (*LegacyPool, *ecdsa.PrivateKey) {
	key, _ := crypto.GenerateKey()
	addr := crypto.PubkeyToAddress(key.PublicKey)

	ddb, genBlock := newDBWithCeloGenesis(config, addr)
	stateRoot := genBlock.Header().Root
	statedb, err := state.New(stateRoot, ddb)
	if err != nil {
		panic(err)
	}
	blockchain := newTestBlockChain(config, 10000000, statedb, new(event.Feed))
	pool := New(testTxPoolConfig, blockchain)

	block := blockchain.CurrentBlock()
	// inject the state-root from the genesis chain, so
	// that the fee-currency allocs are accessible from the state
	// and can be used to create the fee-currency context in the txpool
	block.Root = stateRoot
	if err := pool.Init(testTxPoolConfig.PriceLimit, block, makeAddressReserver()); err != nil {
		panic(err)
	}
	// wait for the pool to initialize
	<-pool.initDoneCh
	return pool, key
}

func TestBaseFeeFloorValidityCheck(t *testing.T) {
	t.Parallel()

	pool, key := setupCeloPoolWithConfig(celoTestChainConfig)
	defer pool.Close()

	// gas-price below base-fee-floor should return early
	// and thus raise an error in the validation
	t.Logf("fee-currency ctx: %v", pool.feeCurrencyContext.ExchangeRates)

	// the PriceLimit config is set to 1, so we need at least a tip of 1
	tx := pricedCip64Transaction(0, 21000, big.NewInt(98), big.NewInt(1), nil, key)

	if err, want := pool.validateTxBasics(tx, false), txpool.ErrGasPriceDoesNotExceedBaseFeeFloor; !errors.Is(err, want) {
		t.Errorf("want %v have %v", want, err)
	}
	// also test with fee currency conversion
	tx = pricedCip64Transaction(0, 21000+feeCurrencyIntrinsicGas, big.NewInt(198), big.NewInt(2), &feeCurrencyOne, key)
	if err, want := pool.validateTxBasics(tx, false), txpool.ErrGasPriceDoesNotExceedBaseFeeFloor; !errors.Is(err, want) {
		t.Errorf("want %v have %v", want, err)
	}

	// gas-price just at base-fee-floor should be valid
	tx = pricedCip64Transaction(0, 21000+feeCurrencyIntrinsicGas, big.NewInt(100), big.NewInt(1), nil, key)
	assert.NoError(t, pool.validateTxBasics(tx, false))
	// also test with fee currency conversion, increase nonce because of previous tx was valid
	tx = pricedCip64Transaction(1, 21000+feeCurrencyIntrinsicGas, big.NewInt(200), big.NewInt(2), &feeCurrencyOne, key)
	assert.NoError(t, pool.validateTxBasics(tx, false))
}
