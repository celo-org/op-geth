#!/bin/bash
set -eo pipefail

source shared.sh
prepare_node

(cd debug-fee-currency && forge build)
export COMPILED_TEST_CONTRACT=../debug-fee-currency/out/DebugFeeCurrency.sol/DebugFeeCurrency.json
(cd js-tests && ./node_modules/mocha/bin/mocha.js test_viem_smoketest.mjs --timeout 25000 --exit)
echo  go test -v ./smoketest_unsupported -op-geth-url $ETH_RPC_URL -private-key $ACC_PRIVKEY -fee-currency $FEE_CURRENCY
go test -v ./smoketest_unsupported_txs -tags smoketest -op-geth-url $ETH_RPC_URL -private-key $ACC_PRIVKEY -fee-currency $FEE_CURRENCY
