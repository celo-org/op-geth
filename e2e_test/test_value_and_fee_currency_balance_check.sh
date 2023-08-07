#!/bin/bash
#shellcheck disable=SC2086
set -eo pipefail
set -x

source shared.sh
source debug-fee-currency/lib.sh

TEST_ACCOUNT_ADDR=0xEa787f769d66B5C131319f262F07254790985BdC
TEST_ACCOUNT_PRIVKEY=0xd36ad839c0bc4bfd8c718a3219591a791871dafad2391149153e6abb43a777fd

fee_currency=$(deploy_fee_currency false false false)

# Send 2.5e15 fee currency to test account
cast send --private-key $ACC_PRIVKEY $fee_currency 'transfer(address to, uint256 value) returns (bool)' $TEST_ACCOUNT_ADDR 2500000000000000
# Send 1e18 celo to test account
cast send --private-key $ACC_PRIVKEY $TOKEN_ADDR 'transfer(address to, uint256 value) returns (bool)' $TEST_ACCOUNT_ADDR 1000000000000000000

# balanceFeeCurrency=2.5e15, txCost=2.25e15, balanceCelo=1e18, valueCelo=1e15
# this should succed because the value and the txCost should not be added (with the bug the total cost could be 3.25e15 will fail because the balanceFeeCurrency is 2.5e15)
$SCRIPT_DIR/js-tests/send_tx.mjs "$(cast chain-id)" $TEST_ACCOUNT_PRIVKEY $fee_currency 1 false 1000000000000000 | assert_cip_64_tx true ""

cleanup_fee_currency $fee_currency