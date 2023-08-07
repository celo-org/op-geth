#!/bin/bash
#shellcheck disable=SC2086
set -eo pipefail

source shared.sh
source debug-fee-currency/lib.sh

tail -F -n0 geth.log >debug-fee-currency/geth.unblock_fee_currency.log & # start log capture
trap 'kill %%' EXIT # kill bg tail job on exit
(
	sleep 0.2
	fee_currency=$(deploy_fee_currency false false true)

	# trigger the first failed call to the CreditFees(), causing the
	# currency to get temporarily blocklisted
	cip_64_tx $fee_currency 1 true 2 | assert_cip_64_tx false

	sleep 2
	# Unblock the currency via the admin API
	unblock_fee_currency $fee_currency

	# since the fee currency was unblocked,
	# this should be executed in miner
	cip_64_tx $fee_currency 1 true 2 | assert_cip_64_tx false

	cleanup_fee_currency $fee_currency
)
sleep 0.5
# expect the execution error to appear twice
if [ "$(grep -Ec "fee-currency EVM execution error.+fee-currency contract error during internal EVM call: surpassed maximum allowed intrinsic gas for CreditFees\(\) in fee-currency" debug-fee-currency/geth.unblock_fee_currency.log)" -ne 2 ]; then exit 1; fi
