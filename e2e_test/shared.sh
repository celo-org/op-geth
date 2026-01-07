#!/bin/bash
#shellcheck disable=SC2034  # unused vars make sense in a shared file

SCRIPT_DIR=$(readlink -f "$(dirname "$0")")
export SCRIPT_DIR
export TERM="${TERM:-xterm-256color}"

case $NETWORK in
# Get these values by querying the registry:
# for contract in GoldToken FeeHandler FeeCurrencyDirectory
#      cast call 0x000000000000000000000000000000000000ce10 "getAddressForStringOrDie(string calldata identifier) returns (address)" $contract
#  end
mainnet)
	export ETH_RPC_URL=wss://forno.celo.org/ws
	export TOKEN_ADDR=0x471EcE3750Da237f93B8E339c536989b8978a438
	export FEE_HANDLER=0xcD437749E43A154C07F3553504c68fBfD56B8778
	export FEE_CURRENCY=0xD8763CBa276a3738E6DE85b4b3bF5FDed6D6cA73
	export FEE_CURRENCY_DIRECTORY_ADDR=0x15F344b9E6c3Cb6F0376A36A64928b13F62C6276
	echo "Using mainnet network"
	;;
celo-sepolia)
	export ETH_RPC_URL=wss://forno.celo-sepolia.celo-testnet.org/ws
	export TOKEN_ADDR=0x471EcE3750Da237f93B8E339c536989b8978a438
	export FEE_HANDLER=0xcD437749E43A154C07F3553504c68fBfD56B8778
	export FEE_CURRENCY=0x6B172e333e2978484261D7eCC3DE491E79764BbC
	export FEE_CURRENCY_DIRECTORY_ADDR=0x9212Fb72ae65367A7c887eC4Ad9bE310BAC611BF
	echo "Using Celo Sepolia network"

	case $CURRENCY in
        EUR)
			echo "Set FEE_CURRENCY to cEUR address"
          	export FEE_CURRENCY=0x6B172e333e2978484261D7eCC3DE491E79764BbC
          	;;
        USD)
			echo "Set FEE_CURRENCY to cUSD address"
          	export FEE_CURRENCY=0xEF4d55D6dE8e8d73232827Cd1e9b2F2dBb45bC80
          	;;	
        REAL)
		  	echo "Set FEE_CURRENCY to cREAL address"
          	export FEE_CURRENCY=0x13d68A1Bf4a8cB7d9feF54EF70401871b666269c
          	;;
        '')
			echo "Set FEE_CURRENCY to cEUR address"
        	export FEE_CURRENCY=0x6B172e333e2978484261D7eCC3DE491E79764BbC
        	;;
    esac
	;;
'')
	export ETH_RPC_URL=http://127.0.0.1:8545
	export TOKEN_ADDR=0x471ece3750da237f93b8e339c536989b8978a438
	export FEE_HANDLER=0xcd437749e43a154c07f3553504c68fbfd56b8778
	export FEE_CURRENCY=0x000000000000000000000000000000000000ce16
	export FEE_CURRENCY2=0x000000000000000000000000000000000000ce17
	export FEE_CURRENCY_DIRECTORY_ADDR=0x15F344b9E6c3Cb6F0376A36A64928b13F62C6276
	echo "Using local network"
	;;
esac

export ACC_ADDR=0x42cf1bbc38BaAA3c4898ce8790e21eD2738c6A4a
export ACC_PRIVKEY=0x2771aff413cac48d9f8c114fabddd9195a2129f3c2c436caa07e27bb7f58ead5
export REGISTRY_ADDR=0x000000000000000000000000000000000000ce10
export ORACLE3=0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb0003

export FIXIDITY_1=1000000000000000000000000
export ZERO_ADDRESS=0x0000000000000000000000000000000000000000

prepare_node() {
    (
        cd js-tests || exit 1
        # re-install when .package-lock.json is missing　or package-lock.json is newer than it
        if [[ ! -f node_modules/.package-lock.json ]] || 
           [[ package-lock.json -nt node_modules/.package-lock.json ]]; then
            npm ci
        fi
    )
}
