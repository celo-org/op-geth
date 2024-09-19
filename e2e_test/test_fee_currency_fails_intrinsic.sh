#!/bin/bash
#shellcheck disable=SC2086
set -eo pipefail

source shared.sh

# Expect that the creditGasFees failed and is logged by geth
tail -f -n0 geth.log >debug-fee-currency/geth.intrinsic.log & # start log capture
(cd debug-fee-currency && ./deploy_and_send_tx.sh false false true)
sleep 0.5
kill %1 # stop log capture
grep "surpassed maximum allowed intrinsic gas for CreditFees() in fee-currency: out of gas" debug-fee-currency/geth.intrinsic.log
