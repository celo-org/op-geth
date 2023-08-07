#!/bin/bash
#shellcheck disable=SC2086
set -eo pipefail

source shared.sh
source debug-fee-currency/lib.sh

tail -F -n0 geth.log >debug-fee-currency/geth.disable_and_enable_blocking.log & # start log capture
trap 'kill %%' EXIT # kill bg tail job on exit
(
	sleep 0.2
	fee_currency=$(deploy_fee_currency false false true)

	# Disable and Enable blocking for the fee currency
	disable_block_list_fee_currency $fee_currency
	enable_block_list_fee_currency $fee_currency

	# trigger the first failed call to the CreditFees(), causing the
	# currency to get temporarily blocklisted.
	cip_64_tx $fee_currency 1 true 2 | assert_cip_64_tx false

	sleep 2

	# since the fee currency is temporarily blocked,
	# this should NOT make the transaction execute anymore,
	# but invalidate the transaction earlier.
	cip_64_tx $fee_currency 1 true 2 | assert_cip_64_tx false

	cleanup_fee_currency $fee_currency
)
sleep 0.5
# Even though we send the faulty fee‑currency transaction twice,
# the execution error should occur only once.
if [ "$(grep -Ec "fee-currency EVM execution error.+fee-currency contract error during internal EVM call: surpassed maximum allowed intrinsic gas for CreditFees\(\) in fee-currency" debug-fee-currency/geth.disable_and_enable_blocking.log)" -ne 1 ]; then exit 1; fi
