package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum-optimism/superchain-registry/superchain"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

var newMainnetOpGenesis = `{
   "baseFeePerGas" : "0x3b9aca00",
   "difficulty" : "0x0",
   "extraData" : "0x424544524f434b",
   "gasLimit" : "0x1c9c380",
   "gasUsed" : "0x0",
   "logsBloom" : "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
   "miner" : "0x4200000000000000000000000000000000000011",
   "mixHash" : "0x0000000000000000000000000000000000000000000000000000000000000000",
   "nonce" : "0x0000000000000000",
   "number" : "0x645c277",
   "parentHash" : "0x21a168dfa5e727926063a28ba16fd5ee84c814e847c81a699c7a0ea551e4ca50",
   "receiptsRoot" : "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
   "sealFields" : [],
   "sha3Uncles" : "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
   "stateRoot" : "0x920314c198da844a041d63bf6cbe8b59583165fd2229d1b3f599da812fd424cb",
   "timestamp" : "0x647f5ea7",
   "totalDifficulty" : "0x0",
   "transactions" : [],
   "transactionsRoot" : "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
   "uncles" : []
}`

func LoadOPStackGenesis(chainID uint64) (*Genesis, error) {

	chConfig, ok := superchain.OPChains[chainID]
	if !ok {
		return nil, fmt.Errorf("unknown chain ID: %d", chainID)
	}

	cfg, err := params.LoadOPStackChainConfig(chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to load params.ChainConfig for chain %d: %w", chainID, err)
	}

	gen, err := superchain.LoadGenesis(chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to load genesis definition for chain %d: %w", chainID, err)
	}
	var genesis *Genesis

	genesis = &Genesis{
		Config:     cfg,
		Nonce:      gen.Nonce,
		Timestamp:  gen.Timestamp,
		ExtraData:  gen.ExtraData,
		GasLimit:   gen.GasLimit,
		Difficulty: (*big.Int)(gen.Difficulty),
		Mixhash:    common.Hash(gen.Mixhash),
		Coinbase:   common.Address(gen.Coinbase),
		Alloc:      make(GenesisAlloc),
		Number:     gen.Number,
		GasUsed:    gen.GasUsed,
		ParentHash: common.Hash(gen.ParentHash),
		BaseFee:    (*big.Int)(gen.BaseFee),
	}

	for addr, acc := range gen.Alloc {
		var code []byte
		if acc.CodeHash != ([32]byte{}) {
			dat, err := superchain.LoadContractBytecode(acc.CodeHash)
			if err != nil {
				return nil, fmt.Errorf("failed to load bytecode %s of address %s in chain %d: %w", acc.CodeHash, addr, chainID, err)
			}
			code = dat
		}
		var storage map[common.Hash]common.Hash
		if len(acc.Storage) > 0 {
			storage = make(map[common.Hash]common.Hash)
			for k, v := range acc.Storage {
				storage[common.Hash(k)] = common.Hash(v)
			}
		}
		bal := common.Big0
		if acc.Balance != nil {
			bal = (*big.Int)(acc.Balance)
		}
		genesis.Alloc[common.Address(addr)] = GenesisAccount{
			Code:    code,
			Storage: storage,
			Balance: bal,
			Nonce:   acc.Nonce,
		}
	}
	if gen.StateHash != nil {
		if len(gen.Alloc) > 0 {
			return nil, fmt.Errorf("chain definition unexpectedly contains both allocation (%d) and state-hash %s", len(gen.Alloc), *gen.StateHash)
		}
		genesis.StateHash = (*common.Hash)(gen.StateHash)
	}

	if chainID == 10 {
		h := types.Header{}
		if err := json.NewDecoder(bytes.NewBufferString(newMainnetOpGenesis)).Decode(&h); err != nil {
			return nil, fmt.Errorf("failed to decode genesis allocation of %d: %w", chainID, err)
		}
		genesis = &Genesis{
			Config:      cfg,
			Nonce:       h.Nonce.Uint64(),
			Timestamp:   h.Time,
			ExtraData:   h.Extra,
			GasLimit:    h.GasLimit,
			Difficulty:  h.Difficulty,
			Mixhash:     h.MixDigest,
			Coinbase:    h.Coinbase,
			Alloc:       nil,
			Number:      h.Number.Uint64(),
			GasUsed:     h.GasUsed,
			ParentHash:  h.ParentHash,
			BaseFee:     h.BaseFee,
			UncleHash:   h.UncleHash,
			TxHash:      h.TxHash,
			ReceiptHash: h.ReceiptHash,
			Bloom:       h.Bloom,
			StateHash:   &h.Root,
		}
	}
	genesisBlock := genesis.ToBlock()
	genesisBlockHash := genesisBlock.Hash()
	expectedHash := common.Hash([32]byte(chConfig.Genesis.L2.Hash))

	// Verify we correctly produced the genesis config by recomputing the genesis-block-hash,
	// and check the genesis matches the chain genesis definition.
	if chConfig.Genesis.L2.Number != genesisBlock.NumberU64() {
		switch chainID {
		case params.OPMainnetChainID:
			expectedHash = common.HexToHash("0x7ca38a1916c42007829c55e69d3e9a73265554b586a499015373241b8a3fa48b")
		case params.OPGoerliChainID:
			expectedHash = common.HexToHash("0xc1fc15cd51159b1f1e5cbc4b82e85c1447ddfa33c52cf1d98d14fba0d6354be1")
		default:
			return nil, fmt.Errorf("unknown stateless genesis definition for chain %d", chainID)
		}
	}
	if expectedHash != genesisBlockHash {
		return nil, fmt.Errorf("produced genesis with hash %s but expected %s", genesisBlockHash, expectedHash)
	}
	return genesis, nil
}
