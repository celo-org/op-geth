#!/bin/bash
#shellcheck disable=SC2034,SC2155,SC2086
set -xeo pipefail

source ./lib.sh

fee_currency=$(deploy_fee_currency $1 $2 $3)
cip_64_tx $fee_currency
