import { assert } from "chai";
import "mocha";
import {
	parseAbi,
} from "viem";
import fs from "fs";
import { publicClient, walletClient } from "./viem_setup.mjs"

// Load compiled contract
const testContractJSON = JSON.parse(fs.readFileSync(process.env.COMPILED_TEST_CONTRACT, 'utf8'));

// check checks that the receipt has status success and that the transaction
// type matches the expected type, since viem sometimes mangles the type when
// building txs.
async function check(txHash, type) {
	const receipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
	assert.equal(receipt.status, "success", "receipt status 'failure'");
	const transaction = await publicClient.getTransaction({ hash: txHash });
	assert.equal(transaction.type, type, "transaction type does not match");
}

// sendTypedTransaction sends a transaction with the given type and an optional
// feeCurrency.
async function sendTypedTransaction(type, feeCurrency) {
	return await walletClient.sendTransaction({
		to: "0x00000000000000000000000000000000DeaDBeef",
		value: 1,
		type: type,
		feeCurrency: feeCurrency,
	});
}

// sendTypedSmartContractTransaction initiates a token transfer with the given type
// and an optional feeCurrency.
async function sendTypedSmartContractTransaction(type, feeCurrency) {
	const abi = parseAbi(['function transfer(address to, uint256 value) external returns (bool)']);
	return await walletClient.writeContract({
		abi: abi,
		address: process.env.TOKEN_ADDR,
		functionName: 'transfer',
		args: ['0x00000000000000000000000000000000DeaDBeef', 1n],
		type: type,
		feeCurrency: feeCurrency,
	});
}

// sendTypedCreateTransaction sends a create transaction with the given type
// and an optional feeCurrency.
async function sendTypedCreateTransaction(type, feeCurrency) {
	return await walletClient.deployContract({
		type: type,
		feeCurrency: feeCurrency,
		bytecode: testContractJSON.bytecode.object,
		abi: testContractJSON.abi,
		// The constructor args for the test contract at ../debug-fee-currency/DebugFeeCurrency.sol
		args: [1n, true, true, true],
	});
}

["legacy", "eip2930", "eip1559", "cip64"].forEach(function (type) {
	describe("viem smoke test, tx type " + type, () => {
		const feeCurrency = type == "cip64" ? process.env.FEE_CURRENCY : undefined;
		it("send tx", async () => {
			const send = await sendTypedTransaction(type, feeCurrency);
			await check(send, type);
		});
		it("send create tx", async () => {
			const create = await sendTypedCreateTransaction(type, feeCurrency);
			await check(create, type);
		});
		it("send contract interaction tx", async () => {
			const contract = await sendTypedSmartContractTransaction(type, feeCurrency);
			await check(contract, type);
		});
	});
});
describe("viem smoke test, unsupported txs", () => {
	// This test is failing because the produced transaction is of type cip64.
	// I guess this is a problem with the viem internals.
	it.skip("cip42", async () => {
		const type = "cip42";
		await sendTypedTransaction(type);
	});

	it("legacy tx with fee currency", async () => {
		const type = "legacy";
		try {
			const hash = await sendTypedTransaction(type, process.env.FEE_CURRENCY);

			// When using the devnet an exception is thrown from
			// sendTypedTransaction, on alfajores the transaction is submitted so we
			// get a hash but it's not included. So we validate here that the
			// transaction was not included.
			let blockNumber = await publicClient.getBlockNumber();
			const oldBlockNumber = blockNumber;
			while (blockNumber == oldBlockNumber) {
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

	it("legacy create tx with fee currency", async () => {
		const type = "legacy";
		try {
			const hash = await sendTypedCreateTransaction(type, process.env.FEE_CURRENCY);

			// When using the devnet an exception is thrown from
			// sendTypedTransaction, on alfajores the transaction is submitted so we
			// get a hash but it's not included. So we validate here that the
			// transaction was not included.
			let blockNumber = await publicClient.getBlockNumber();
			const oldBlockNumber = blockNumber;
			while (blockNumber == oldBlockNumber) {
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
