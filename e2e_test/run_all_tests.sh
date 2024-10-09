#!/bin/bash
set -eo pipefail

SCRIPT_DIR=$(readlink -f "$(dirname "$0")")
source "$SCRIPT_DIR/shared.sh"

TEST_GLOB=$1

if [ -z $NETWORK ]; then
    ## Start geth
    cd "$SCRIPT_DIR/.." || exit 1
    make geth
    trap 'kill %%' EXIT # kill bg job at exit
    build/bin/geth --dev --http --http.api eth,web3,net &>"$SCRIPT_DIR/geth.log" &
    
    # Wait for geth to be ready
    for _ in {1..10}; do
    	if cast block &>/dev/null; then
    		break
    	fi
    	sleep 0.2
    done
    
    ## Run tests
    echo Geth ready, start tests
    cd "$SCRIPT_DIR" || exit 1
fi

# There's a problem with geth return errors on the first transaction sent.
# See https://github.com/ethereum/web3.py/issues/3212
# To work around this, send a transaction before running tests
cast send --json --private-key "$ACC_PRIVKEY" "$TOKEN_ADDR" 'transfer(address to, uint256 value) returns (bool)' 0x000000000000000000000000000000000000dEaD 100 > /dev/null || true

failures=0
tests=0
for f in test_*"$TEST_GLOB"*; do
	if [[ -n $NETWORK ]]; then
		case $f in
		  # Skip tests that require a local network.
		  test_fee_currency_fails_on_credit.sh|test_fee_currency_fails_on_debit.sh|test_fee_currency_fails_intrinsic.sh)
		  continue
		  ;;
	    esac
	fi
	echo -e "\nRun $f"
	if "./$f"; then
		tput setaf 2 || true
		echo "PASS $f"
	else
		tput setaf 1 || true
		echo "FAIL $f ❌"
		((failures++)) || true
	fi
	tput sgr0 || true
	((tests++)) || true
done

## Final summary
echo
if [[ $failures -eq 0 ]]; then
	tput setaf 2 || true
	echo All $tests tests succeeded!
else
	tput setaf 1 || true
	echo $failures/$tests failed.
fi
tput sgr0 || true
exit $failures
