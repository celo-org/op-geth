package core

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/addresses"
	"github.com/ethereum/go-ethereum/contracts/celo"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

// Decode from file. Strip 0x prefix and whitespace (newline) if present.
func DecodeHex(hexbytes []byte) ([]byte, error) {
	// Strip 0x prefix and trailing newline
	hexbytes = bytes.TrimPrefix(bytes.TrimSpace(hexbytes), []byte("0x")) // strip 0x prefix

	// Decode hex string
	bytes := make([]byte, hex.DecodedLen(len(hexbytes)))
	_, err := hex.Decode(bytes, hexbytes)
	if err != nil {
		return nil, fmt.Errorf("DecodeHex: %w", err)
	}

	return bytes, nil
}

// Calculate address in evm mapping: keccak(key ++ mapping_slot)
func CalcMapAddr(slot common.Hash, key common.Hash) common.Hash {
	return crypto.Keccak256Hash(append(key.Bytes(), slot.Bytes()...))
}

// Increase a hash value by `i`, used for addresses in 32byte fields
func incHash(addr common.Hash, i int64) common.Hash {
	return common.BigToHash(new(big.Int).Add(addr.Big(), big.NewInt(i)))
}

var (
	DevPrivateKey, _  = crypto.HexToECDSA("2771aff413cac48d9f8c114fabddd9195a2129f3c2c436caa07e27bb7f58ead5")
	DevAddr           = common.HexToAddress("0x42cf1bbc38BaAA3c4898ce8790e21eD2738c6A4a")
	DevPrivateKey2, _ = crypto.HexToECDSA("fbc0c0a6b8e05a2770632982af1ea41bd444390b34476b52d57b0d455911a94c")
	DevAddr2          = common.HexToAddress("0xf280E427723B0ee6a1eF614ffFBDE15DB5fED5b1")

	DevFeeCurrencyAddr      = common.HexToAddress("0x000000000000000000000000000000000000ce16") // worth half as much as native CELO
	DevFeeCurrencyAddr2     = common.HexToAddress("0x000000000000000000000000000000000000ce17") // worth twice as much as native CELO
	DevBalance, _           = new(big.Int).SetString("100000000000000000000", 10)
	rateNumerator, _        = new(big.Int).SetString("2000000000000000000000000", 10)
	rateNumerator2, _       = new(big.Int).SetString("500000000000000000000000", 10)
	rateDenominator, _      = new(big.Int).SetString("1000000000000000000000000", 10)
	mockOracleAddr          = common.HexToAddress("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb0001")
	mockOracleAddr2         = common.HexToAddress("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb0002")
	mockOracleAddr3         = common.HexToAddress("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb0003")
	FaucetAddr              = common.HexToAddress("0xfcf982bb4015852e706100b14e21f947a5bb718e")
	FeeCurrencyIntrinsicGas = uint64(50000)
)

func CeloGenesisAccounts(fundedAddr common.Address) GenesisAlloc {
	// Initialize Bytecodes
	celoTokenBytecode, err := DecodeHex(celo.CeloTokenBytecodeRaw)
	if err != nil {
		panic(err)
	}
	feeCurrencyBytecode, err := DecodeHex(celo.FeeCurrencyBytecodeRaw)
	if err != nil {
		panic(err)
	}
	feeCurrencyDirectoryBytecode, err := DecodeHex(celo.FeeCurrencyDirectoryBytecodeRaw)
	if err != nil {
		panic(err)
	}
	mockOracleBytecode, err := DecodeHex(celo.MockOracleBytecodeRaw)
	if err != nil {
		panic(err)
	}

	var devBalance32 common.Hash
	DevBalance.FillBytes(devBalance32[:])

	arrayAtSlot2 := crypto.Keccak256Hash(common.HexToHash("0x2").Bytes())

	faucetBalance, ok := new(big.Int).SetString("500000000000000000000000000", 10) // 500M
	if !ok {
		panic("Could not set faucet balance!")
	}
	genesisAccounts := map[common.Address]GenesisAccount{
		addresses.MainnetAddresses.CeloToken: {
			Code:    celoTokenBytecode,
			Balance: big.NewInt(0),
		},
		DevFeeCurrencyAddr: {
			Code:    feeCurrencyBytecode,
			Balance: big.NewInt(0),
			Storage: map[common.Hash]common.Hash{
				CalcMapAddr(common.HexToHash("0x0"), common.BytesToHash(DevAddr.Bytes())):    devBalance32, // _balances[DevAddr]
				CalcMapAddr(common.HexToHash("0x0"), common.BytesToHash(DevAddr2.Bytes())):   devBalance32, // _balances[DevAddr2]
				CalcMapAddr(common.HexToHash("0x0"), common.BytesToHash(fundedAddr.Bytes())): devBalance32, // _balances[fund]
				common.HexToHash("0x2"): devBalance32, // _totalSupply
			},
		},
		DevFeeCurrencyAddr2: {
			Code:    feeCurrencyBytecode,
			Balance: big.NewInt(0),
			Storage: map[common.Hash]common.Hash{
				CalcMapAddr(common.HexToHash("0x0"), common.BytesToHash(DevAddr.Bytes())):    devBalance32, // _balances[DevAddr]
				CalcMapAddr(common.HexToHash("0x0"), common.BytesToHash(DevAddr2.Bytes())):   devBalance32, // _balances[DevAddr2]
				CalcMapAddr(common.HexToHash("0x0"), common.BytesToHash(fundedAddr.Bytes())): devBalance32, // _balances[fund]
				common.HexToHash("0x2"): devBalance32, // _totalSupply
			},
		},
		mockOracleAddr: {
			Code:    mockOracleBytecode,
			Balance: big.NewInt(0),
			Storage: map[common.Hash]common.Hash{
				common.HexToHash("0x0"): common.BigToHash(rateNumerator),
				common.HexToHash("0x1"): common.BigToHash(rateDenominator),
				common.HexToHash("0x3"): common.BytesToHash(DevFeeCurrencyAddr.Bytes()),
			},
		},
		mockOracleAddr2: {
			Code:    mockOracleBytecode,
			Balance: big.NewInt(0),
			Storage: map[common.Hash]common.Hash{
				common.HexToHash("0x0"): common.BigToHash(rateNumerator2),
				common.HexToHash("0x1"): common.BigToHash(rateDenominator),
				common.HexToHash("0x3"): common.BytesToHash(DevFeeCurrencyAddr2.Bytes()),
			},
		},
		mockOracleAddr3: {
			Code:    mockOracleBytecode,
			Balance: big.NewInt(0),
			// This oracle is available for tests of contracts outside the celo_genesis, so no initialization is done at this point
		},
		DevAddr: {
			Balance: DevBalance,
		},
		DevAddr2: {
			Balance: DevBalance,
		},
		FaucetAddr: {
			Balance: faucetBalance,
		},
		fundedAddr: {
			Balance: DevBalance,
		},
	}

	// FeeCurrencyDirectory
	devAddrOffset1 := common.Hash{}
	copy(devAddrOffset1[11:], DevAddr.Bytes())
	feeCurrencyDirectoryStorage := map[common.Hash]common.Hash{
		// owner, slot 0 offset 1
		common.HexToHash("0x0"): devAddrOffset1,
		// add entries to currencyList at slot 2
		common.HexToHash("0x2"):  common.HexToHash("0x2"),                         // array length 2
		arrayAtSlot2:             common.BytesToHash(DevFeeCurrencyAddr.Bytes()),  // FeeCurrency
		incHash(arrayAtSlot2, 1): common.BytesToHash(DevFeeCurrencyAddr2.Bytes()), // FeeCurrency2
	}
	// add entries to currencyConfig mapping
	addFeeCurrencyToStorage(DevFeeCurrencyAddr, mockOracleAddr, feeCurrencyDirectoryStorage)
	addFeeCurrencyToStorage(DevFeeCurrencyAddr2, mockOracleAddr2, feeCurrencyDirectoryStorage)
	genesisAccounts[addresses.MainnetAddresses.FeeCurrencyDirectory] = GenesisAccount{
		Code:    feeCurrencyDirectoryBytecode,
		Balance: big.NewInt(0),
		Storage: feeCurrencyDirectoryStorage,
	}

	return genesisAccounts
}

func addFeeCurrencyToStorage(feeCurrencyAddr common.Address, oracleAddr common.Address, storage map[common.Hash]common.Hash) {
	structStart := CalcMapAddr(common.HexToHash("0x1"), common.BytesToHash(feeCurrencyAddr.Bytes()))
	storage[structStart] = common.BytesToHash(oracleAddr.Bytes())                                   // oracle
	storage[incHash(structStart, 1)] = common.BigToHash(big.NewInt(int64(FeeCurrencyIntrinsicGas))) // intrinsicGas
}

// CeloDeveloperGenesisBlock returns the 'geth --dev' genesis block with Celo-specific configurations.
// It differs from the standard DeveloperGenesisBlock in that it:
// 1. Uses a more realistic chain configuration that mirrors mainnet settings
// 2. Includes Celo-specific genesis accounts and contract deployments
// 3. Sets up fee currency contracts and mock oracles for testing
//
// This ensures that development mode behaves as closely as possible to mainnet,
// making it more suitable for testing and development purposes.
func CeloDeveloperGenesisBlock(gasLimit uint64, faucet *common.Address) *Genesis {
	genesis := DeveloperGenesisBlock(gasLimit, faucet)

	// Set our own more realistic config
	config := *params.DevChainConfig
	genesis.Config = &config
	genesis.BaseFee = big.NewInt(int64(config.Celo.EIP1559BaseFeeFloor))

	// Add state from celoGenesisAccounts
	for addr, data := range CeloGenesisAccounts(common.HexToAddress("0x2")) {
		genesis.Alloc[addr] = data
	}

	return genesis
}
