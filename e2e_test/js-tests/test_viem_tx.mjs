import { assert } from "chai";
import "mocha";
import {
	createPublicClient,
	createWalletClient,
	http,
	defineChain,
	encodeFunctionData,
	decodeFunctionResult,
	parseAbi,
} from "viem";
import { base, celoAlfajores } from "viem/chains";
import { privateKeyToAccount } from "viem/accounts";

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

// Returns the base fee per gas for the current block multiplied by 2 to account for any increase in the subsequent block.
async function getGasFees(publicClient, tip) {
	const b = await publicClient.getBlock();
	return [BigInt(b.baseFeePerGas) * 2n + tip, tip];
}

const testNonceBump = async (
	firstCap,
	firstCurrency,
	secondCap,
	secondCurrency,
	shouldReplace,
) => {
	const syncBarrierRequest = await walletClient.prepareTransactionRequest({
		account,
		to: "0x00000000000000000000000000000000DeaDBeef",
		value: 2,
		gas: 22000,
	});
	const firstTxHash = await walletClient.sendTransaction({
		account,
		to: "0x00000000000000000000000000000000DeaDBeef",
		value: 2,
		gas: 171000,
		maxFeePerGas: firstCap,
		maxPriorityFeePerGas: firstCap,
		nonce: syncBarrierRequest.nonce + 1,
		feeCurrency: firstCurrency,
	});
	var secondTxHash;
	try {
		secondTxHash = await walletClient.sendTransaction({
			account,
			to: "0x00000000000000000000000000000000DeaDBeef",
			value: 3,
			// dynaimcally retrieve intrinsic gas from feeCurrency directory "feeCurrencyConfig"
			gas: 171000,
			maxFeePerGas: secondCap,
			maxPriorityFeePerGas: secondCap,
			nonce: syncBarrierRequest.nonce + 1,
			feeCurrency: secondCurrency,
		});
	} catch (err) {
		// If shouldReplace, no error should be thrown
		// If shouldReplace == false, exactly the underpriced error should be thrown
		if (
			err.cause.details != "replacement transaction underpriced" ||
			shouldReplace
		) {
			throw err; // Only throw if unexpected error.
		}
	}
	const syncBarrierSignature =
		await walletClient.signTransaction(syncBarrierRequest);
	const barrierTxHash = await walletClient.sendRawTransaction({
		serializedTransaction: syncBarrierSignature,
	});
	await publicClient.waitForTransactionReceipt({ hash: barrierTxHash });
	if (shouldReplace) {
		// The new transaction was included.
		await publicClient.waitForTransactionReceipt({ hash: secondTxHash });
	} else {
		// The original transaction was not replaced.
		await publicClient.waitForTransactionReceipt({ hash: firstTxHash });
	}
};

describe("viem send tx", () => {
	// it("send basic tx and check receipt", async () => {
	// 	const request = await walletClient.prepareTransactionRequest({
	// 		account,
	// 		to: "0x00000000000000000000000000000000DeaDBeef",
	// 		value: 1,
	// 		gas: 21000,
	// 	});
	// 	const signature = await walletClient.signTransaction(request);
	// 	const hash = await walletClient.sendRawTransaction({
	// 		serializedTransaction: signature,
	// 	});
	// 	const receipt = await publicClient.waitForTransactionReceipt({ hash });
	// 	assert.equal(receipt.status, "success", "receipt status 'failure'");
	// }).timeout(20_000);

	// it("send basic tx using viem gas estimation and check receipt", async () => {
	// 	const request = await walletClient.prepareTransactionRequest({
	// 		account,
	// 		to: "0x00000000000000000000000000000000DeaDBeef",
	// 		value: 1,
	// 	});
	// 	const signature = await walletClient.signTransaction(request);
	// 	const hash = await walletClient.sendRawTransaction({
	// 		serializedTransaction: signature,
	// 	});
	// 	const receipt = await publicClient.waitForTransactionReceipt({ hash });
	// 	assert.equal(receipt.status, "success", "receipt status 'failure'");
	// }).timeout(20_000);

	// it("send fee currency tx with explicit gas fields and check receipt", async () => {
	// 	const [maxFeePerGas, tip] = await getGasFees(publicClient, 2n);
	// 	const request = await walletClient.prepareTransactionRequest({
	// 		account,
	// 		to: "0x00000000000000000000000000000000DeaDBeef",
	// 		value: 2,
	// 		gas: 171000,
	// 		feeCurrency: process.env.FEE_CURRENCY,
	// 		maxFeePerGas: maxFeePerGas,
	// 		maxPriorityFeePerGas: tip,
	// 	});
	// 	const signature = await walletClient.signTransaction(request);
	// 	const hash = await walletClient.sendRawTransaction({
	// 		serializedTransaction: signature,
	// 	});
	// 	const receipt = await publicClient.waitForTransactionReceipt({ hash });
	// 	assert.equal(receipt.status, "success", "receipt status 'failure'");
	// }).timeout(100_000);

	it("test gas price difference for fee currency", async () => {
		const request = await walletClient.prepareTransactionRequest({
			account,
			to: "0x00000000000000000000000000000000DeaDBeef",
			value: 2,
			gas: 171000,
			feeCurrency: process.env.FEE_CURRENCY,
		});

		// Get the raw gas price and maxPriorityFeePerGas
		const gasPriceNative = await publicClient.getGasPrice({});
		var maxPriorityFeePerGasNative =
			await publicClient.estimateMaxPriorityFeePerGas({});
		const block = await publicClient.getBlock({});

		// Check them against the base fee.
		assert.equal(
			BigInt(block.baseFeePerGas) + maxPriorityFeePerGasNative,
			gasPriceNative,
		);

		// viem's getGasPrice does not expose additional request parameters, but
		// Celo's override 'chain.fees.estimateFeesPerGas' action does. This will
		// call the eth_gasPrice and eth_maxPriorityFeePerGas methods with the
		// additional feeCurrency parameter internally, it also multiplies the
		// gasPriceInFeeCurrency by 12n/10n.
		var fees = await publicClient.estimateFeesPerGas({
			type: "eip1559",
			request: {
				feeCurrency: process.env.FEE_CURRENCY,
			},
		});

		// Get the exchange rates for the fee currency.
		const abi = parseAbi(['function getExchangeRate(address token) public view returns (uint256 numerator, uint256 denominator)']);
		const [numerator, denominator] = await publicClient.readContract({
			address: process.env.FEE_CURRENCY_DIRECTORY_ADDR,
			abi: abi,
			functionName: 'getExchangeRate',
			args: [process.env.FEE_CURRENCY],
		})

		// TODO fix this when viem is fixed - https://github.com/celo-org/viem/pull/20
		// The expected value for the max fee should be the (baseFeePerGas * multiplier) + maxPriorityFeePerGas
		// Instead what is currently returned is (maxFeePerGas * multiplier) + maxPriorityFeePerGas
		const maxPriorityFeeInFeeCurrency = (maxPriorityFeePerGasNative * numerator) / denominator;
    const maxFeeInFeeCurrency = ((block.baseFeePerGas +maxPriorityFeePerGasNative)*numerator)/denominator
		assert.equal(fees.maxFeePerGas, ((maxFeeInFeeCurrency*12n)/10n) + maxPriorityFeeInFeeCurrency);
		assert.equal(fees.maxPriorityFeePerGas, maxPriorityFeeInFeeCurrency);

		// check that the prepared transaction request uses the
		// converted gas price internally
		assert.equal(request.maxFeePerGas, fees.maxFeePerGas);
		assert.equal(request.maxPriorityFeePerGas, fees.maxPriorityFeePerGas);
	}).timeout(20_000);

	// it("send fee currency with gas estimation tx and check receipt", async () => {
	// 	const request = await walletClient.prepareTransactionRequest({
	// 		account,
	// 		to: "0x00000000000000000000000000000000DeaDBeef",
	// 		value: 2,
	// 		feeCurrency: process.env.FEE_CURRENCY,
	// 		maxFeePerGas: 50000000000n,
	// 		maxPriorityFeePerGas: 0n,
	// 	});
	// 	const signature = await walletClient.signTransaction(request);
	// 	const hash = await walletClient.sendRawTransaction({
	// 		serializedTransaction: signature,
	// 	});
	// 	const receipt = await publicClient.waitForTransactionReceipt({ hash });
	// 	assert.equal(receipt.status, "success", "receipt status 'failure'");
	// }).timeout(20_000);

	// it("send overlapping nonce tx in different currencies", async () => {
	// 	const priceBump = 1.1;
	// 	const rate = 2;
	// 	// Native to FEE_CURRENCY
	// 	const nativeCap = 30_000_000_000;
	// 	const bumpCurrencyCap = BigInt(Math.round(nativeCap * rate * priceBump));
	// 	const failToBumpCurrencyCap = BigInt(
	// 		Math.round(nativeCap * rate * priceBump) - 1,
	// 	);
	// 	// FEE_CURRENCY to Native
	// 	const currencyCap = 60_000_000_000;
	// 	const bumpNativeCap = BigInt(Math.round((currencyCap * priceBump) / rate));
	// 	const failToBumpNativeCap = BigInt(
	// 		Math.round((currencyCap * priceBump) / rate) - 1,
	// 	);
	// 	const tokenCurrency = process.env.FEE_CURRENCY;
	// 	const nativeCurrency = null;
	// 	await testNonceBump(
	// 		nativeCap,
	// 		nativeCurrency,
	// 		bumpCurrencyCap,
	// 		tokenCurrency,
	// 		true,
	// 	);
	// 	await testNonceBump(
	// 		nativeCap,
	// 		nativeCurrency,
	// 		failToBumpCurrencyCap,
	// 		tokenCurrency,
	// 		false,
	// 	);
	// 	await testNonceBump(
	// 		currencyCap,
	// 		tokenCurrency,
	// 		bumpNativeCap,
	// 		nativeCurrency,
	// 		true,
	// 	);
	// 	await testNonceBump(
	// 		currencyCap,
	// 		tokenCurrency,
	// 		failToBumpNativeCap,
	// 		nativeCurrency,
	// 		false,
	// 	);
	// }).timeout(20_000);

	// it("send tx with unregistered fee currency", async () => {
	// 	const request = await walletClient.prepareTransactionRequest({
	// 		account,
	// 		to: "0x00000000000000000000000000000000DeaDBeef",
	// 		value: 2,
	// 		gas: 171000,
	// 		feeCurrency: "0x000000000000000000000000000000000badc310",
	// 		maxFeePerGas: 1000000000n,
	// 		maxPriorityFeePerGas: 0n,
	// 	});
	// 	const signature = await walletClient.signTransaction(request);
	// 	try {
	// 		await walletClient.sendRawTransaction({
	// 			serializedTransaction: signature,
	// 		});
	// 		assert.fail("Failed to filter unregistered feeCurrency");
	// 	} catch (err) {
	// 		// TODO: find a better way to check the error type
	// 		if (err.cause.details.indexOf("unregistered fee-currency address") >= 0) {
	// 			// Test success
	// 		} else {
	// 			throw err;
	// 		}
	// 	}
	// }).timeout(20_000);
});
