#!/bin/bash
#shellcheck disable=SC2034,SC2155,SC2086
set -xeo pipefail

# args:
# 	$1: failOnDebit (bool):
# 		if true, this will make the DebugFeeCurrenc.DebitFees() call fail with a revert
# 	$2: failOnCredit (bool)
# 		if true, this will make the DebugFeeCurrenc.CreditFees() call fail with a revert
# 	$3: highGasOnCredit (bool)
# 		if true, this will make the DebugFeeCurrenc.CreditFees() call use
# 		a high amount of gas
# returns:
# 	deployed fee-currency address
function deploy_fee_currency() {
	(
		local fee_currency=$(
			forge create --root "$SCRIPT_DIR/debug-fee-currency" --contracts "$SCRIPT_DIR/debug-fee-currency" --private-key $ACC_PRIVKEY DebugFeeCurrency.sol:DebugFeeCurrency --constructor-args '100000000000000000000000000' $1 $2 $3 --json | jq .deployedTo -r
		)
		if [ -z "${fee_currency}" ]; then
			exit 1
		fi
		cast send --private-key $ACC_PRIVKEY $ORACLE3 'setExchangeRate(address, uint256, uint256)' $fee_currency 2ether 1ether &>/dev/null
		cast send --private-key $ACC_PRIVKEY $FEE_CURRENCY_DIRECTORY_ADDR 'setCurrencyConfig(address, address, uint256)' $fee_currency $ORACLE3 60000 &>/dev/null
		echo "$fee_currency"
	)
}

# args:
# 	$1: feeCurrencyAddress (string):
# 		which fee-currency address to use for the default CIP-64 transaction
function cip_64_tx() {
	$SCRIPT_DIR/js-tests/send_tx.mjs "$(cast chain-id)" $ACC_PRIVKEY $1
}

# use this function to assert the cip_64_tx return value, by using a pipe like
# `cip_64_tx "$fee-currency" | assert_cip_64_tx true`
#
# args:
# 	$1: success (string):
# 		expected success value, "true" for when the cip-64 tx should have succeeded, "false" if not
# 	$2: error-regex (string):
# 		expected RPC return-error value regex to grep for, use "null", "" or unset value if no error is assumed.
function assert_cip_64_tx() {
	local value
	read -r value
	local expected_error="$2"

	if [ "$(echo "$value" | jq .success)" != "$1" ]; then
		exit 1
	fi
	if [ -z "$expected_error" ]; then
		expected_error="null"
	fi
	echo "$value" | jq .error | grep -qE "$expected_error"
}
