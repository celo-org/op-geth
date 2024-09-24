#!/bin/bash
#shellcheck disable=SC2086
set -eo pipefail

source shared.sh
source debug-fee-currency/lib.sh

tail -F -n0 geth.log >debug-fee-currency/geth.partial.log & # start log capture
(
	sleep 0.2
	fee_currency=$(deploy_fee_currency false true false)

	# trigger the first failed call to the CreditFees(), causing the
	# currency to get temporarily blocklisted.
	# initial tx should not succeed, should have required a replacement transaction.
	cip_64_tx $fee_currency 1 true | assert_cip_64_tx false

	sleep 2

	# since the fee currency is temporarily blocked,
	# this should NOT make the transaction execute anymore,
	# but invalidate the transaction earlier.
	# initial tx should not succeed, should have required a replacement transaction.
	cip_64_tx $fee_currency 1 true | assert_cip_64_tx false

	cleanup_fee_currency $fee_currency
)
sleep 0.5
kill %1 # stop log capture
# although we sent a transaction wih faulty fee-currency twice,
# the EVM call should have been executed only once
grep "" debug-fee-currency/geth.partial.log
if [ "$(grep -Ec "fee-currency EVM execution error, temporarily blocking fee-currency in local txpools .+ This DebugFeeCurrency always fails in \(old\) creditGasFees!" debug-fee-currency/geth.partial.log)" -ne 1 ]; then exit 1; fi
