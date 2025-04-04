#!/bin/bash
#shellcheck disable=SC2086
set -eo pipefail

source shared.sh
source debug-fee-currency/lib.sh

fee_currency=$(deploy_fee_currency false false false 70000)
gas=$(estimate_tx 20 $fee_currency)

cleanup_fee_currency $fee_currency

# intrinsic of fee_currency: 70000
# intrinsic of tx: 21000
# total: 91000
if [ $gas -ne 91000 ]; then exit 1; fi

