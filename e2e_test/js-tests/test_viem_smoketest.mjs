import { assert } from "chai";
import "mocha";
import {
	createPublicClient,
	createWalletClient,
	http,
	defineChain,
	parseAbi,
	encodeFunctionData,
} from "viem";
import { celoAlfajores } from "viem/chains";
import { privateKeyToAccount } from "viem/accounts";
import fs from "fs";

// Load compiled contract
const testContractJSON = JSON.parse(fs.readFileSync(process.env.COMPILED_TEST_CONTRACT, 'utf8'));


// Setup up chain
const devChain = defineChain({
	...celoAlfajores,
	id: 1337,
	name: "local dev chain",
	network: "dev",
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
		default:
			return devChain
	};
})();

// Set up clients/wallet
const publicClient = createPublicClient({
	chain: chain,
	transport: http(),
});
const account = privateKeyToAccount(process.env.ACC_PRIVKEY);
const walletClient = createWalletClient({
	account,
	chain: chain,
	transport: http(),
});

// check checks that the receipt has status success and that the transaction
// type matches the expected type, since viem sometimes mangles the type when
// building txs.
async function check(txHash, type) {
	const receipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
	const transaction = await publicClient.getTransaction({ hash: txHash });
	assert.equal(transaction.type, type, "transaction type does not match");
	assert.equal(receipt.status, "success", "receipt status 'failure'");
}

// sendTypedTransaction sends a transaction with the given type and an optional
// feeCurrency.
async function sendTypedTransaction(type, feeCurrency){
		 return await walletClient.sendTransaction({
			to: "0x00000000000000000000000000000000DeaDBeef",
			value: 1,
			type: type,
			feeCurrency: feeCurrency,
		});
}

// sendTypedSmartContractTransaction initiates a token transfer with the given type
// and an optional feeCurrency.
async function sendTypedSmartContractTransaction(type, feeCurrency){
	const abi = parseAbi(['function transfer(address to, uint256 value) external returns (bool)']);
	const data = encodeFunctionData({
		abi: abi,
		functionName: 'transfer',
		args: ['0x00000000000000000000000000000000DeaDBeef', 1n]
	  });
	const hash = await walletClient.sendTransaction({
		type: type,
		to: process.env.TOKEN_ADDR,
		feeCurrency: feeCurrency,
		data:data,
	});
	return hash;
}

// sendTypedCreateTransaction sends a create transaction with the given type
// and an optional feeCurrency.
async function sendTypedCreateTransaction(type, feeCurrency){
		 return await walletClient.deployContract({
			type: type,
			feeCurrency: feeCurrency,
			bytecode: testContractJSON.bytecode.object,
			abi: testContractJSON.abi,
			 // The constructor args for the test contract at ../debug-fee-currency/DebugFeeCurrency.sol
			args:[1n, true, true, true],
		});
}

// verifyTypedTransactions is a helper function that submits a send and create
// transaction of the given type with optional feeCurrency and checks the
// results.
async function verifyTypedTransactions(type, feeCurrency){
		const send = await sendTypedTransaction(type, feeCurrency);
		await check(send, type);
		const create = await sendTypedCreateTransaction(type, feeCurrency);
		await check(create, type);
		const contract = await sendTypedSmartContractTransaction(type, feeCurrency);
		await check(contract, type);
}

describe("viem smoke test", () => {
	it("send legacy tx", async () => {
		const type = "legacy";
		await verifyTypedTransactions(type);
	});

	it("send eip2930 tx", async () => {
		const type = "eip2930";
		await verifyTypedTransactions(type);
	});

	it("send eip1559 tx", async () => {
		const type = "eip1559";
		await verifyTypedTransactions(type);
	});

	it("send cip64 tx", async () => {
		const type = "cip64";
		await verifyTypedTransactions(type, process.env.FEE_CURRENCY);
	});

	// This test is failing because the produced transaction is of type cip64.
	// I guess this is a problem with the viem internals.
	it.skip("cip42 not supported", async () => {
		const type = "cip42";
		await verifyTypedTransactions(type);
	});

	it("legacy tx with fee currency not supported", async () => {
		const type = "legacy";
		try {
			const hash = await sendTypedTransaction(type,process.env.FEE_CURRENCY);

			// When using the devnet an exception is thrown from
			// sendTypedTransaction, on alfajores the transaction is submitted so we
			// get a hash but it's not included. So we validate here that the
			// transaction was not included.
			let blockNumber = await publicClient.getBlockNumber();
			const oldBlockNumber = blockNumber;
			while(blockNumber == oldBlockNumber){
				// Sleep 100ms
				await new Promise(r => setTimeout(r, 100));
				blockNumber = await publicClient.getBlockNumber();
			}
			const tx = await publicClient.getTransaction({ hash: hash });
			assert.isNull(tx);
		} catch (error) {
			// Expect error to be thrown
			return
			// exceptionThrown += 1;
		}
			assert.fail("Managed to send unsupported legacy tx with fee currency");
	});

	it("legacy create tx with fee currency not supported", async () => {
		const type = "legacy";
		try {
			const hash = await sendTypedCreateTransaction(type,process.env.FEE_CURRENCY);

			// When using the devnet an exception is thrown from
			// sendTypedTransaction, on alfajores the transaction is submitted so we
			// get a hash but it's not included. So we validate here that the
			// transaction was not included.
			let blockNumber = await publicClient.getBlockNumber();
			const oldBlockNumber = blockNumber;
			while(blockNumber == oldBlockNumber){
				blockNumber = await publicClient.getBlockNumber();
			}
			const tx = await publicClient.getTransaction({ hash: hash });
			assert.isNull(tx);
		} catch (error) {
			// Expect error to be thrown
			return
			// exceptionThrown += 1;
		}
			assert.fail("Managed to send unsupported legacy tx with fee currency");
	});
});
