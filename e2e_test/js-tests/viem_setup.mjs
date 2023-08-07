import { assert } from "chai";
import "mocha";
import {
	createPublicClient,
	createWalletClient,
	http,
	webSocket,
	defineChain,
} from "viem";
import { celo, celoSepolia } from "viem/chains";
import { privateKeyToAccount } from "viem/accounts";

// Setup up chain
const devChain = defineChain({
	...celoSepolia,
	id: 1337,
	name: "local dev chain",
	rpcUrls: {
		default: {
			http: [process.env.ETH_RPC_URL],
			webSocket: [process.env.ETH_RPC_URL],
		},
	},
});

const celoMainnet = defineChain({
	...celo,
	rpcUrls: {
		default: {
			http: [process.env.ETH_RPC_URL],
			webSocket: [process.env.ETH_RPC_URL],
		},
	},
});

const chain = (() => {
	switch (process.env.NETWORK) {
		case 'celo-sepolia':
			return celoSepolia
		case 'mainnet':
			return celoMainnet
		default:
			return devChain
	};
})();

const transportForNetwork = (() => {
	switch (process.env.NETWORK) {
		case 'celo-sepolia':
		case 'mainnet':
			return webSocket(process.env.ETH_RPC_URL);
		default:
			return http(process.env.ETH_RPC_URL);
	};
})

// Set up clients/wallet
export const publicClient = createPublicClient({
	chain: chain,
	transport: transportForNetwork(),
});
export const account = privateKeyToAccount(process.env.ACC_PRIVKEY);
export const walletClient = createWalletClient({
	account,
	chain: chain,
	transport: transportForNetwork(),
});
