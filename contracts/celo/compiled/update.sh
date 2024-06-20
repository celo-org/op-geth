#!/bin/bash

SCRIPT_DIR=$(readlink -f "$(dirname "$0")")
CONTRACTS_DIR=${CELO_OPTIMISM_REPO:-~/optimism}/packages/contracts-bedrock

(cd "$CONTRACTS_DIR" && forge build)

for contract in CeloRegistry FeeCurrency Proxy; do
	contract_json="$CONTRACTS_DIR/forge-artifacts/$contract.sol/$contract.json"
	jq .abi "$contract_json" >"$SCRIPT_DIR/$contract.abi"
	jq .deployedBytecode.object -r "$contract_json" >"$SCRIPT_DIR/$contract.bin-runtime"
done

# These should go into the optimism repo, but since they are not there yet,
# let's get them from the celo-monorepo for now.
CONTRACTS_DIR=${CELO_MONOREPO:-~/celo-monorepo}/packages/protocol

(cd "$CONTRACTS_DIR" && forge build)
for contract in GoldToken FeeCurrencyDirectory IFeeCurrencyDirectory MockOracle; do
	contract_json="$CONTRACTS_DIR/out/$contract.sol/$contract.json"
	jq .abi "$contract_json" >"$SCRIPT_DIR/$contract.abi"
	jq .deployedBytecode.object -r "$contract_json" >"$SCRIPT_DIR/$contract.bin-runtime"
done
