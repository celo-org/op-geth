#!/bin/bash
#shellcheck disable=SC2086
set -eo pipefail

source shared.sh
source debug-fee-currency/lib.sh

# Expect that the debitGasFees fails during tx submission
#
fee_currency=$(deploy_fee_currency true false false)
# this fails during the RPC call, since the DebitFees() is part of the pre-validation
cip_64_tx $fee_currency 1 false | assert_cip_64_tx false "fee-currency internal error"

cleanup_fee_currency $fee_currency
