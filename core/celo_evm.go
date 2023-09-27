package core

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/contracts/celo/abigen"
	contracts_config "github.com/ethereum/go-ethereum/contracts/config"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
)

func setCeloFieldsInBlockContext(blockContext *vm.BlockContext, header *types.Header, config *params.ChainConfig, statedb *state.StateDB) {
	if !config.IsCel2(header.Time) {
		return
	}

	stateCopy := statedb.Copy()
	caller := &CeloBackend{config, stateCopy}

	// Set goldTokenAddress
	registry, err := abigen.NewRegistryCaller(contracts_config.RegistrySmartContractAddress, caller)
	if err != nil {
		log.Error("Failed to access registry!", "err", err)
	}
	blockContext.GoldTokenAddress, err = registry.GetAddressForOrDie(&bind.CallOpts{}, contracts_config.GoldTokenRegistryId)
	if err != nil {
		log.Error("Failed to get address for GoldToken!", "err", err)
	}
}
