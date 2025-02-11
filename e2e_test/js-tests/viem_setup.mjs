import { assert } from "chai";
import "mocha";
import {
	createPublicClient,
	createWalletClient,
	http,
	defineChain,
} from "viem";
import { celoAlfajores } from "viem/chains";
import { privateKeyToAccount } from "viem/accounts";

// Setup up chain
const devChain = defineChain({
	...celoAlfajores,
	id: 1337,
	name: "local dev chain",
	rpcUrls: {
		default: {
			http: [process.env.ETH_RPC_URL],
		},
	},
});

const celoBaklava = defineChain({
	...celoAlfajores,
	id: 62320,
	name: "baklava",
	rpcUrls: {
		default: {
			http: [process.env.ETH_RPC_URL],
		},
	},
});

const chain = (() => {
	switch (process.env.NETWORK) {
		case 'alfajores':
			return celoAlfajores
		case 'baklava':
			return celoBaklava
		default:
			return devChain
	};
})();

// Set up clients/wallet
export const publicClient = createPublicClient({
	chain: chain,
	transport: http(),
});
export const account = privateKeyToAccount(process.env.ACC_PRIVKEY);
export const walletClient = createWalletClient({
	account,
	chain: chain,
	transport: http(),
});