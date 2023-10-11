#!/bin/bash
#shellcheck disable=SC2086
set -eo pipefail
set -x

source shared.sh

# Register fee currency whitelist and sorted oracles
cast send --private-key $ACC_PRIVKEY $REGISTRY_ADDR 'function setAddressFor(string calldata identifier, address addr) external' FeeCurrencyWhitelist $FEE_CURRENCY_WHITELIST_ADDR
cast send --private-key $ACC_PRIVKEY $REGISTRY_ADDR 'function setAddressFor(string calldata identifier, address addr) external' SortedOracles $SORTED_ORACLES_ADDR
# Add our account as oracle and submit value
cast send --private-key $ACC_PRIVKEY $SORTED_ORACLES_ADDR 'function addOracle(address token, address oracleAddress) external' $FEE_CURRENCY $ACC_ADDR
cast send --private-key $ACC_PRIVKEY $SORTED_ORACLES_ADDR 'function report(address token, uint256 value, address lesserKey, address greaterKey) external' $FEE_CURRENCY $FIXIDITY_1 $ZERO_ADDRESS $ZERO_ADDRESS

# Debug suggestions:
# cast call $FEE_CURRENCY_WHITELIST_ADDR 'function getWhitelist() external view returns (address[] memory)'
# cast logs 'event OracleReported(address indexed token, address indexed oracle, uint256 timestamp, uint256 value)'
# cast call $SORTED_ORACLES_ADDR 'function medianRate(address token) external view returns (uint256, uint256)' $FEE_CURRENCY

# Send token and check balance (TODO: add fee currency)
cast send --private-key $ACC_PRIVKEY --value 10 0x000000000000000000000000000000000000dEaD
