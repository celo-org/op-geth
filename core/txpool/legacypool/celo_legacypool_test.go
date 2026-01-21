package legacypool

import (
	"crypto/ecdsa"
	"errors"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/misc/eip1559"
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
	jt := uint64(0)
	config.Cel2Time = &ct
	config.JovianTime = &jt
	config.Celo = &params.CeloConfig{EIP1559BaseFeeFloor: baseFeeFloor}
	return config
}

var (
	// worth half as much as native celo
	feeCurrencyOne = core.DevFeeCurrencyAddr
	// worth twice as much as native celo
	feeCurrencyTwo          = core.DevFeeCurrencyAddr2
	feeCurrencyIntrinsicGas = core.FeeCurrencyIntrinsicGas
	defaultBaseFeeFloor     = 100
	defaultChainConfig      = celoConfig(uint64(defaultBaseFeeFloor))
	preJovianChainConfig    *params.ChainConfig
)

func init() {
	preJovianChainConfig = celoConfig(uint64(defaultBaseFeeFloor))
	preJovianChainConfig.JovianTime = nil
}

func pricedCip64Transaction(
	config *params.ChainConfig,
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
	}), types.LatestSigner(config), key)
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

func setupCeloPoolWithConfig(config *params.ChainConfig, minBaseFee ...uint64) (*LegacyPool, *ecdsa.PrivateKey) {
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
	// Encode minBaseFee in ExtraData (defaults to 0)
	mbf := uint64(0)
	if len(minBaseFee) > 0 {
		mbf = minBaseFee[0]
	}
	block.Extra = eip1559.EncodeJovianExtraData(250, 6, mbf)

	if err := pool.Init(testTxPoolConfig.PriceLimit, block, newReserver()); err != nil {
		panic(err)
	}
	// wait for the pool to initialize
	<-pool.initDoneCh
	return pool, key
}

func TestBelowBaseFeeFloorValidityCheck(t *testing.T) {
	t.Parallel()

	pool, key := setupCeloPoolWithConfig(preJovianChainConfig)
	defer pool.Close()

	// gas-price below base-fee-floor should return early
	// and thus raise an error in the validation

	// We need to ensure that the tip cap fulfils the min tip requirement since that is checked first in ValidateTransaction.
	tx := pricedCip64Transaction(preJovianChainConfig, 0, 21000, big.NewInt(99), big.NewInt(1), nil, key)
	if err, want := pool.addRemoteSync(tx), txpool.ErrGasFeeCapBelowMinBaseFee; !errors.Is(err, want) {
		t.Errorf("want %v have %v", want, err)
	}
	// also test with fee currency conversion
	tx = pricedCip64Transaction(preJovianChainConfig, 0, 21000+feeCurrencyIntrinsicGas, big.NewInt(198), big.NewInt(2), &feeCurrencyOne, key)
	if err, want := pool.addRemoteSync(tx), txpool.ErrGasFeeCapBelowMinBaseFee; !errors.Is(err, want) {
		t.Errorf("want %v have %v", want, err)
	}
	tx = pricedCip64Transaction(preJovianChainConfig, 0, 21000+feeCurrencyIntrinsicGas, big.NewInt(48), big.NewInt(1), &feeCurrencyTwo, key)
	if err, want := pool.addRemoteSync(tx), txpool.ErrGasFeeCapBelowMinBaseFee; !errors.Is(err, want) {
		t.Errorf("want %v have %v", want, err)
	}
}

func TestAboveBaseFeeFloorValidityCheck(t *testing.T) {
	t.Parallel()

	pool, key := setupCeloPoolWithConfig(preJovianChainConfig)
	defer pool.Close()

	// gas-price just at base-fee-floor should be valid,
	// this also adds the required min-tip of 1
	tx := pricedCip64Transaction(preJovianChainConfig, 0, 21000, big.NewInt(101), big.NewInt(1), nil, key)
	assert.NoError(t, pool.addRemote(tx))
	// also test with fee currency conversion, increase nonce because of previous tx was valid
	tx = pricedCip64Transaction(preJovianChainConfig, 1, 21000+feeCurrencyIntrinsicGas, big.NewInt(202), big.NewInt(2), &feeCurrencyOne, key)
	assert.NoError(t, pool.addRemote(tx))
	tx = pricedCip64Transaction(preJovianChainConfig, 2, 21000+feeCurrencyIntrinsicGas, big.NewInt(51), big.NewInt(1), &feeCurrencyTwo, key)
	assert.NoError(t, pool.addRemote(tx))
}

func TestBelowMinTipValidityCheck(t *testing.T) {
	t.Parallel()

	pool, key := setupCeloPoolWithConfig(preJovianChainConfig)
	defer pool.Close()

	// the min-tip is set to 1 per default

	// Gas-price just at base-fee-floor should be valid,
	// the effective gas-price would also pass the min-tip restriction of 1.
	// However the explicit gas-tip-cap at 0 should reject the transaction.
	tx := pricedCip64Transaction(preJovianChainConfig, 0, 21000, big.NewInt(101), big.NewInt(0), nil, key)
	if err, want := pool.addRemote(tx), txpool.ErrTxGasPriceTooLow; !errors.Is(err, want) {
		t.Errorf("want %v have %v", want, err)
	}
	tx = pricedCip64Transaction(preJovianChainConfig, 0, 21000+feeCurrencyIntrinsicGas, big.NewInt(202), big.NewInt(0), &feeCurrencyOne, key)
	if err, want := pool.addRemote(tx), txpool.ErrTxGasPriceTooLow; !errors.Is(err, want) {
		t.Errorf("want %v have %v", want, err)
	}

	// This passes the check that only checks the actual gas-tip-cap value for the min-tip that was
	// tested above.
	// Now the effective gas-tip should still be below the min-tip, since we consume everything
	// for the base fee floor and thus the tx should get rejected.
	tx = pricedCip64Transaction(preJovianChainConfig, 0, 21000, big.NewInt(100), big.NewInt(1), nil, key)
	if err, want := pool.addRemote(tx), txpool.ErrUnderpriced; !errors.Is(err, want) {
		t.Errorf("want %v have %v", want, err)
	}
	tx = pricedCip64Transaction(preJovianChainConfig, 0, 21000+feeCurrencyIntrinsicGas, big.NewInt(200), big.NewInt(2), &feeCurrencyOne, key)
	if err, want := pool.addRemote(tx), txpool.ErrUnderpriced; !errors.Is(err, want) {
		t.Errorf("want %v have %v", want, err)
	}
}

func TestExpectMinTipRoundingFeeCurrency(t *testing.T) {
	t.Parallel()

	pool, key := setupCeloPoolWithConfig(preJovianChainConfig)
	defer pool.Close()

	// the min-tip is set to 1 per default

	// even though the gas-tip-cap as well as the effective gas tip at the base-fee-floor
	// is 0, the transaction is still accepted.
	// This is because at a min-tip requirement of 1, a more valuable currency than native
	// token will get rounded down to a min-tip of 0 during conversion.
	tx := pricedCip64Transaction(preJovianChainConfig, 0, 21000+feeCurrencyIntrinsicGas, big.NewInt(50), big.NewInt(0), &feeCurrencyTwo, key)
	assert.NoError(t, pool.addRemote(tx))

	// set the required min-tip to 10
	pool.SetGasTip(big.NewInt(10))

	// but as soon as we increase the min-tip, the check rejects a gas-tip-cap that is too low after conversion
	tx = pricedCip64Transaction(preJovianChainConfig, 0, 21000+feeCurrencyIntrinsicGas, big.NewInt(100), big.NewInt(4), &feeCurrencyTwo, key)
	if err, want := pool.addRemote(tx), txpool.ErrTxGasPriceTooLow; !errors.Is(err, want) {
		t.Errorf("want %v have %v", want, err)
	}
}

// Verify that transactions with gas price below the Celo base fee floor are
// accepted when Jovian is active and minBaseFee is 0, because the Celo floor
// validation is skipped in favor of OP's minBaseFee mechanism.
func TestBaseFeeFloorNotEnforcedPostJovian(t *testing.T) {
	t.Parallel()

	// Use minBaseFee=0 so that transactions with low fees are accepted
	// (Celo floor is not enforced, and OP minBaseFee=0 means no OP floor either)
	pool, key := setupCeloPoolWithConfig(defaultChainConfig, 0)
	defer pool.Close()

	// Transaction with gas price below Celo base fee floor should be ACCEPTED
	// because Jovian is active and Celo floor validation is not enforced.
	// The gas fee cap of 99 is below the defaultBaseFeeFloor of 100.
	tx := pricedCip64Transaction(defaultChainConfig, 0, 21000, big.NewInt(99), big.NewInt(1), nil, key)
	assert.NoError(t, pool.addRemote(tx))

	// Also test with fee currency - gas price below converted floor should be accepted
	tx = pricedCip64Transaction(defaultChainConfig, 1, 21000+feeCurrencyIntrinsicGas, big.NewInt(198), big.NewInt(2), &feeCurrencyOne, key)
	assert.NoError(t, pool.addRemote(tx))
}

// Verify that transactions with gas fee cap below the OP minBaseFee are rejected
// when Jovian is active.
func TestBelowMinBaseFeeRejected(t *testing.T) {
	t.Parallel()

	// Set minBaseFee=100
	pool, key := setupCeloPoolWithConfig(defaultChainConfig, 100)
	defer pool.Close()

	// Transaction with gas fee cap below minBaseFee should be REJECTED
	// The gas fee cap of 99 is below the minBaseFee of 100.
	tx := pricedCip64Transaction(defaultChainConfig, 0, 21000, big.NewInt(99), big.NewInt(1), nil, key)
	if err, want := pool.addRemoteSync(tx), txpool.ErrGasFeeCapBelowMinBaseFee; !errors.Is(err, want) {
		t.Errorf("want %v have %v", want, err)
	}

	// Also test with fee currency conversion
	// feeCurrencyOne is worth half as much as native CELO, so minBaseFee=100 converts to 200
	// A gas fee cap of 199 should be rejected
	tx = pricedCip64Transaction(defaultChainConfig, 0, 21000+feeCurrencyIntrinsicGas, big.NewInt(199), big.NewInt(2), &feeCurrencyOne, key)
	if err, want := pool.addRemoteSync(tx), txpool.ErrGasFeeCapBelowMinBaseFee; !errors.Is(err, want) {
		t.Errorf("want %v have %v", want, err)
	}

	// feeCurrencyTwo is worth twice as much as native CELO, so minBaseFee=100 converts to 50
	// A gas fee cap of 49 should be rejected
	tx = pricedCip64Transaction(defaultChainConfig, 0, 21000+feeCurrencyIntrinsicGas, big.NewInt(49), big.NewInt(1), &feeCurrencyTwo, key)
	if err, want := pool.addRemoteSync(tx), txpool.ErrGasFeeCapBelowMinBaseFee; !errors.Is(err, want) {
		t.Errorf("want %v have %v", want, err)
	}
}

// Verify that transactions with gas fee cap at or above the OP minBaseFee are accepted
// when Jovian is active.
func TestAboveMinBaseFeeAccepted(t *testing.T) {
	t.Parallel()

	// Set minBaseFee=100
	pool, key := setupCeloPoolWithConfig(defaultChainConfig, 100)
	defer pool.Close()

	// Transaction with gas fee cap at minBaseFee should be ACCEPTED
	// The gas fee cap of 101 (100 + 1 for min tip) is at/above the minBaseFee of 100.
	tx := pricedCip64Transaction(defaultChainConfig, 0, 21000, big.NewInt(101), big.NewInt(1), nil, key)
	assert.NoError(t, pool.addRemote(tx))

	// Also test with fee currency conversion
	// feeCurrencyOne is worth half as much as native CELO, so minBaseFee=100 converts to 200
	// A gas fee cap of 202 (200 + 2 for min tip) should be accepted
	tx = pricedCip64Transaction(defaultChainConfig, 1, 21000+feeCurrencyIntrinsicGas, big.NewInt(202), big.NewInt(2), &feeCurrencyOne, key)
	assert.NoError(t, pool.addRemote(tx))

	// feeCurrencyTwo is worth twice as much as native CELO, so minBaseFee=100 converts to 50
	// A gas fee cap of 51 (50 + 1 for min tip) should be accepted
	tx = pricedCip64Transaction(defaultChainConfig, 2, 21000+feeCurrencyIntrinsicGas, big.NewInt(51), big.NewInt(1), &feeCurrencyTwo, key)
	assert.NoError(t, pool.addRemote(tx))
}
