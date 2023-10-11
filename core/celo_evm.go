package core

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/celo/abigen"
	contracts_config "github.com/ethereum/go-ethereum/contracts/config"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
)

// Returns the exchange rates for all gas currencies from CELO
func getExchangeRates(caller *CeloBackend, registry *abigen.RegistryCaller) (map[common.Address]*big.Rat, error) {
	exchangeRates := map[common.Address]*big.Rat{}
	whitelistAddress, err := registry.GetAddressForOrDie(&bind.CallOpts{}, contracts_config.FeeCurrencyWhitelistRegistryId)
	if err != nil {
		return exchangeRates, fmt.Errorf("Failed to get address for FeeCurrencyWhitelist: %w", err)
	}
	whitelist, err := abigen.NewFeeCurrencyWhitelistCaller(whitelistAddress, caller)
	if err != nil {
		return exchangeRates, fmt.Errorf("Failed to access FeeCurrencyWhitelist: %w", err)
	}
	oracleAddress, err := registry.GetAddressForOrDie(&bind.CallOpts{}, contracts_config.SortedOraclesRegistryId)
	if err != nil {
		return exchangeRates, fmt.Errorf("Failed to get address for SortedOracle: %w", err)
	}
	oracle, err := abigen.NewSortedOraclesCaller(oracleAddress, caller)
	if err != nil {
		return exchangeRates, fmt.Errorf("Failed to access SortedOracle: %w", err)
	}

	whitelistedTokens, err := whitelist.GetWhitelist(&bind.CallOpts{})
	if err != nil {
		return exchangeRates, fmt.Errorf("Failed to get whitelisted tokens: %w", err)
	}
	for _, tokenAddress := range whitelistedTokens {
		numerator, denominator, err := oracle.MedianRate(&bind.CallOpts{}, tokenAddress)
		if err != nil {
			log.Error("Failed to get medianRate for gas currency!", "err", err, "tokenAddress", tokenAddress)
			continue
		}
		if denominator.Sign() == 0 {
			log.Error("Bad exchange rate for fee currency", "tokenAddress", tokenAddress, "numerator", numerator, "denominator", denominator)
			continue
		}
		exchangeRates[tokenAddress] = big.NewRat(numerator.Int64(), denominator.Int64())
	}

	return exchangeRates, nil
}

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

	// Add fee currency exchange rates
	blockContext.ExchangeRates, err = getExchangeRates(caller, registry)
	if err != nil {
		log.Error("Error fetching exchange rates!", "err", err)
	}
}
