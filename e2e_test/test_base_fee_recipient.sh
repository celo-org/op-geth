#!/bin/bash
#shellcheck disable=SC2086
set -eo pipefail
set -x

source shared.sh

# Send token and check balance
balance_before=$(cast balance $FEE_HANDLER)
tx_json=$(cast send --json --private-key $ACC_PRIVKEY $TOKEN_ADDR 'transfer(address to, uint256 value) returns (bool)' 0x000000000000000000000000000000000000dEaD 100)
echo "tx json: $tx_json"
tx_hash=$(echo $tx_json | jq -r '.transactionHash')
echo "tx hash: $tx_hash"
cast tx $tx_hash
cast receipt $tx_hash
balance_after=$(cast balance $FEE_HANDLER)
echo "Balance change: $balance_before -> $balance_after"
# TODO(Alec) calculate expected balance change
[[ $((balance_before + 100)) -eq $balance_after ]] || (echo "Balance did not change as expected"; exit 1)