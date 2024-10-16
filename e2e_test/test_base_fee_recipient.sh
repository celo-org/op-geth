#!/bin/bash
#shellcheck disable=SC2086
set -eo pipefail
set -x

source shared.sh

# Send token and check balance
tx_json=$(cast send --json --private-key $ACC_PRIVKEY $TOKEN_ADDR 'transfer(address to, uint256 value) returns (bool)' 0x000000000000000000000000000000000000dEaD 100)
block_number=$(echo $tx_json | jq -r '.blockNumber' | cast to-dec)
block=$(cast block --json --full $block_number)
gas_used=$(echo $block | jq -r '.gasUsed' | cast to-dec)
base_fee=$(echo $block | jq -r '.baseFeePerGas' | cast to-dec)
if [[ -n $NETWORK  ]]; then
	# Every block in a non dev system contains a system tx that pays nothing for gas so we must subtract this from the total gas per block
	system_tx_hash=$(echo $block | jq -r '.transactions[] | select(.from == "0xDeaDDEaDDeAdDeAdDEAdDEaddeAddEAdDEAd0001") | .hash')
	system_tx_gas=$(cast receipt --json $system_tx_hash | jq -r '.cumulativeGasUsed' | cast to-dec)
	gas_used=$((gas_used - system_tx_gas))
fi
balance_before=$(cast balance --block $((block_number-1)) $FEE_HANDLER)
balance_after=$(cast balance --block $block_number $FEE_HANDLER)
balance_change=$((balance_after - balance_before))
expected_balance_change=$((base_fee * gas_used))
[[ $expected_balance_change -eq $balance_change ]] || (
  echo "Balance did not change as expected"
  exit 1
)
