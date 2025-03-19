#!/bin/bash
#shellcheck disable=SC2086
set -eo pipefail

source shared.sh

# Send token and check balance
balance_before=$(cast balance 0x000000000000000000000000000000000000dEaD)
cast send --private-key $ACC_PRIVKEY $TOKEN_ADDR 'transfer(address to, uint256 value) returns (bool)' 0x000000000000000000000000000000000000dEaD 100
balance_after=$(cast balance 0x000000000000000000000000000000000000dEaD)

# Use perl for arbitrary precision arithmetic
expected_balance=$(perl -e "use bignum; print $balance_before + 100")
if [ "$expected_balance" != "$balance_after" ]; then
    echo "Balance did not change as expected"
    echo "Expected: $expected_balance"
    echo "Actual: $balance_after"
    exit 1
fi
