import { assert } from "chai";
import "mocha";
import {
	parseAbi,
} from "viem";
import { publicClient, walletClient } from "./viem_setup.mjs"

const TX_GAS = 21000;

// Returns the base fee per gas for the current block multiplied by 2 to account for any increase in the subsequent block.
async function getGasFees(publicClient, tip, feeCurrency) {
	const rate = await getRate(feeCurrency);
	const b = await publicClient.getBlock();
	let tipInFeeCurrency = rate.toFeeCurrency(tip);
	if (tipInFeeCurrency === 0n) {
		// The tip must be at least native 1 wei for the tx to be included. No
		// matter what the exchange rate for a fee currency is, if the fee currency
		// tip is zero, we can't reach 1 wei if the fee currency tip is zero. So
		// increase the tip to at least one.
		tipInFeeCurrency = 1n;
	}
	return [rate.toFeeCurrency(b.baseFeePerGas) + tipInFeeCurrency, tipInFeeCurrency];
}

const testNonceBump = async (
	firstCap,
	firstCurrency,
	secondCap,
	secondCurrency,
	shouldReplace,
) => {
	const syncBarrierRequest = await walletClient.prepareTransactionRequest({

		to: "0x00000000000000000000000000000000DeaDBeef",
		value: 2,
		gas: 22000,
	});
	const firstTxHash = await walletClient.sendTransaction({
		to: "0x00000000000000000000000000000000DeaDBeef",
		value: 2,
		gas: await getIntrinsicGasForFeeCurrency(TX_GAS, firstCurrency),
		maxFeePerGas: firstCap,
		maxPriorityFeePerGas: firstCap,
		nonce: syncBarrierRequest.nonce + 1,
		feeCurrency: firstCurrency,
	});
	var secondTxHash;
	try {
		secondTxHash = await walletClient.sendTransaction({
			to: "0x00000000000000000000000000000000DeaDBeef",
			value: 3,
			gas: await getIntrinsicGasForFeeCurrency(TX_GAS, secondCurrency),
			maxFeePerGas: secondCap,
			maxPriorityFeePerGas: secondCap,
			nonce: syncBarrierRequest.nonce + 1,
			feeCurrency: secondCurrency,
		});
	} catch (err) {
		// If shouldReplace, no error should be thrown
		// If shouldReplace == false, exactly the underpriced error should be thrown
		if (
			!err.cause.details.includes("replacement transaction underpriced") ||
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
	it("send basic tx and check receipt", async () => {
		const request = await walletClient.prepareTransactionRequest({
			to: "0x00000000000000000000000000000000DeaDBeef",
			value: 1,
			gas: TX_GAS,
		});
		const signature = await walletClient.signTransaction(request);
		const hash = await walletClient.sendRawTransaction({
			serializedTransaction: signature,
		});
		const receipt = await publicClient.waitForTransactionReceipt({ hash });
		assert.equal(receipt.status, "success", "receipt status 'failure'");
	}).timeout(10_000);

	it("send basic tx using viem gas estimation and check receipt", async () => {
		const request = await walletClient.prepareTransactionRequest({
			to: "0x00000000000000000000000000000000DeaDBeef",
			value: 1,
		});
		const signature = await walletClient.signTransaction(request);
		const hash = await walletClient.sendRawTransaction({
			serializedTransaction: signature,
		});
		const receipt = await publicClient.waitForTransactionReceipt({ hash });
		assert.equal(receipt.status, "success", "receipt status 'failure'");
	}).timeout(10_000);

	it("send fee currency tx with explicit gas fields and check receipt", async () => {
		const [maxFeePerGas, tip] = await getGasFees(publicClient, 2n, process.env.FEE_CURRENCY);
		const request = await walletClient.prepareTransactionRequest({
			to: "0x00000000000000000000000000000000DeaDBeef",
			value: 2,
			gas: await getIntrinsicGasForFeeCurrency(TX_GAS, process.env.FEE_CURRENCY),
			feeCurrency: process.env.FEE_CURRENCY,
			maxFeePerGas: maxFeePerGas,
			maxPriorityFeePerGas: tip,
		});
		const signature = await walletClient.signTransaction(request);
		const hash = await walletClient.sendRawTransaction({
			serializedTransaction: signature,
		});
		const receipt = await publicClient.waitForTransactionReceipt({ hash });
		assert.equal(receipt.status, "success", "receipt status 'failure'");
	}).timeout(10_000);

	it("send fee currency tx using viem gas estimation and check receipt", async () => {
		const request = await walletClient.prepareTransactionRequest({
			to: "0x00000000000000000000000000000000DeaDBeef",
			value: 2,
			feeCurrency: process.env.FEE_CURRENCY,
		});
		const signature = await walletClient.signTransaction(request);
		const hash = await walletClient.sendRawTransaction({
			serializedTransaction: signature,
		});
		const receipt = await publicClient.waitForTransactionReceipt({ hash });
		assert.equal(receipt.status, "success", "receipt status 'failure'");
	}).timeout(10_000);

	it("test gas price difference for fee currency", async () => {
		const request = await walletClient.prepareTransactionRequest({
			to: "0x00000000000000000000000000000000DeaDBeef",
			value: 2,
			gas: await getIntrinsicGasForFeeCurrency(TX_GAS, process.env.FEE_CURRENCY),
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
		// additional feeCurrency parameter internally, it also multiplies the base
		// fee component of the maxFeePerGas by a multiplier which by default is
		// 1.2 or (12n/10n).
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
		});

		// The expected value for the max fee should be the (baseFeePerGas * multiplier) + maxPriorityFeePerGas
		const maxPriorityFeeInFeeCurrency = (maxPriorityFeePerGasNative * numerator) / denominator;
		const maxFeeInFeeCurrency = ((block.baseFeePerGas) * numerator) / denominator;
		assert.equal(fees.maxFeePerGas, ((maxFeeInFeeCurrency * 12n) / 10n) + maxPriorityFeeInFeeCurrency);
		assert.equal(fees.maxPriorityFeePerGas, maxPriorityFeeInFeeCurrency);

		// check that the prepared transaction request uses the
		// converted gas price internally
		assert.equal(request.maxFeePerGas, fees.maxFeePerGas);
		assert.equal(request.maxPriorityFeePerGas, fees.maxPriorityFeePerGas);
	}).timeout(10_000);

	// The goal is this test is to ensure that fee currencies are correctly
	// taken into account when performing tx replacements. As such we want the
	// prices that we use for the failed and successful tx replacements to be
	// close to the threshold value, such that an invalid currency conversion is
	// more liable to result in a failure.
	it("send overlapping nonce tx in different currencies", async () => {
		// Note the threshold for a price bump to be accepted is 10%, i.e >= oldPrice * 1.1
		const priceBump = 1.1; // minimum bump percentage to replace a transaction
		const priceNearBump = 1.09; // slightly lower percentage than the price bump

		const rate = await getRate(process.env.FEE_CURRENCY);
		// Native to FEE_CURRENCY
		const nativeCap = 30_000_000_000;
		const bumpCurrencyCap = rate.toFeeCurrency(BigInt(Math.round(nativeCap * priceBump)));
		const failToBumpCurrencyCap = rate.toFeeCurrency(BigInt(
			Math.round(nativeCap * priceNearBump)
		));
		const tokenCurrency = process.env.FEE_CURRENCY;
		const nativeCurrency = null;
		await testNonceBump(
			nativeCap,
			nativeCurrency,
			bumpCurrencyCap,
			tokenCurrency,
			true,
		);
		await testNonceBump(
			nativeCap,
			nativeCurrency,
			failToBumpCurrencyCap,
			tokenCurrency,
			false,
		);

		// FEE_CURRENCY to Native
		const currencyCap = 60_000_000_000;
		const bumpNativeCap = rate.toNative(BigInt(Math.round(currencyCap * priceBump)));
		const failToBumpNativeCap = rate.toNative(BigInt(
			Math.round(currencyCap * priceNearBump)
		));
		await testNonceBump(
			currencyCap,
			tokenCurrency,
			bumpNativeCap,
			nativeCurrency,
			true,
		);
		await testNonceBump(
			currencyCap,
			tokenCurrency,
			failToBumpNativeCap,
			nativeCurrency,
			false,
		);
	}).timeout(60_000);

	it("send tx with unregistered fee currency", async () => {
		const request = await walletClient.prepareTransactionRequest({
			to: "0x00000000000000000000000000000000DeaDBeef",
			value: 2,
			gas: await getIntrinsicGasForFeeCurrency(TX_GAS, process.env.FEE_CURRENCY),
			feeCurrency: "0x000000000000000000000000000000000badc310",
			maxFeePerGas: 1000000000n,
			maxPriorityFeePerGas: 1n,
		});
		const signature = await walletClient.signTransaction(request);
		try {
			await walletClient.sendRawTransaction({
				serializedTransaction: signature,
			});
			assert.fail("Failed to filter unregistered feeCurrency");
		} catch (err) {
			// TODO: find a better way to check the error type
			if (err.cause.details.indexOf("unregistered fee-currency address") >= 0) {
				// Test success
			} else {
				throw err;
			}
		}
	}).timeout(10_000);

	it("send fee currency tx with just high enough gas price", async function () {
		// The idea of this test is to check that the fee currency is taken into
		// account by the server. We do this by using a fee currency that has a
		// value greater than celo, so that the base fee in fee currency becomes a
		// number significantly lower than the base fee in celo. If the server
		// doesn't take into account the fee currency then it will reject the
		// transaction because the maxFeePerGas will be too low.

		// If we are running local tests we use FEE_CURRENCY2 since it is worth
		// double the value of celo, otherwise we use FEE_CURRENCY which is USDC
		// end currently worth roughly double the value of celo.
		const fc = process.env.NETWORK == null ? process.env.FEE_CURRENCY2 : process.env.FEE_CURRENCY;
		const rate = await getRate(fc);
		const block = await publicClient.getBlock({});
		// We increment the base fee by 10% to cover the case where the base fee increases next block.
		const convertedBaseFee = rate.toFeeCurrency(block.baseFeePerGas * 11n/10n);

		// This test assumes the fee currency is more valuable than CELO, so the
		// base fee converted into the fee currency is LOWER than the native CELO
		// base fee. If the exchange rate makes the fee currency equal/cheaper,
		// that assumption breaks and the test becomes invalid,
		// so we skip to avoid a false failure.
		if (rate.toFeeCurrency(1n) >= 1n) {
			this.skip();
			return;
		}

		// Check that the converted base fee value is still below the native base
		// fee value, if this check fails we will need to consider an alternative
		// fee currency to USDC for network tests.
		if (convertedBaseFee >= block.baseFeePerGas) {
			assert.fail(`Converted base fee (${convertedBaseFee}) not less than native base fee (${block.baseFeePerGas})`);
		}
		const maxFeePerGas = convertedBaseFee + 2n;
		const request = await walletClient.prepareTransactionRequest({
			to: "0x00000000000000000000000000000000DeaDBeef",
			value: 2,
			gas: await getIntrinsicGasForFeeCurrency(TX_GAS, fc),
			feeCurrency: fc,
			maxFeePerGas: maxFeePerGas,
			maxPriorityFeePerGas: 2n,
		});
		const signature = await walletClient.signTransaction(request);
		const hash = await walletClient.sendRawTransaction({
			serializedTransaction: signature,
		});
		const receipt = await publicClient.waitForTransactionReceipt({ hash });
		assert.equal(receipt.status, "success", "receipt status 'failure'");
		assert.isAtMost(Number(receipt.effectiveGasPrice), Number(maxFeePerGas), "effective gas price is too high");
		assert.isAbove(Number(receipt.effectiveGasPrice), Number(maxFeePerGas) * 0.7, "effective gas price is too low");
	}).timeout(10_000);

	it("zero tip tx rejected", async () => {
		const gasPrice = await publicClient.getGasPrice();
		let request = await walletClient.prepareTransactionRequest({
			to: "0x00000000000000000000000000000000DeaDBeef",
			gas: TX_GAS,
			maxFeePerGas: gasPrice,
			maxPriorityFeePerGas: 0n,
		});
		await expectTxFail(request, "gas tip cap 0");
	}).timeout(10_000);

});

// expectTxFail expects the transaction to fail with an error that contains the given errorString.
async function expectTxFail(txRequest, errorString) {
	const signedTx = await walletClient.signTransaction(txRequest);
	try {
		await walletClient.sendRawTransaction({
			serializedTransaction: signedTx,
		});
	} catch (err) {
		// Combine multiple error properties to get comprehensive error information
		const errorMessage = [
			err.shortMessage,
			err.details,
			err.cause?.message,
			err.cause?.details
		].filter(Boolean).join(' | ');
		if (!errorMessage.includes(errorString)) {
			assert.fail(`Expected error to contain "${errorString}", but got: ${errorMessage}`);
		}
		return;
	}
	assert.fail("expecting transaction sending to fail")
}

async function getRate(feeCurrencyAddress) {
	const abi = parseAbi(['function getExchangeRate(address token) public view returns (uint256 numerator, uint256 denominator)']);
	const [numerator, denominator] = await publicClient.readContract({
		address: process.env.FEE_CURRENCY_DIRECTORY_ADDR,
		abi: abi,
		functionName: 'getExchangeRate',
		args: [feeCurrencyAddress],
	});
	return {
		toFeeCurrency: (v) => (v * numerator) / denominator,
		toNative: (v) => (v * denominator) / numerator,
	};
}

// getIntrinsicGasForFeeCurrency calculates intrinsic gas by adding extra intrinsic gas to the base intrinsic gas and additional cost
async function getIntrinsicGasForFeeCurrency(baseIntrinsicGas, feeCurrency) {
	if (!feeCurrency) return BigInt(baseIntrinsicGas)
	const extraFee = await getExtraCustomFeeCurrencyIntrinsicGas(feeCurrency)
	return BigInt(baseIntrinsicGas) + extraFee;
}

// getExtraCustomFeeCurrencyIntrinsicGas retrieves extra intrinsic gas from the smart contract for custom fee currency.
async function getExtraCustomFeeCurrencyIntrinsicGas(feeCurrency) {
	const abi = parseAbi(['function getCurrencyConfig(address token) public view returns (address oracle, uint256 intrinsicGas)']);
	const [_, intrinsicGas] = await publicClient.readContract({
		address: process.env.FEE_CURRENCY_DIRECTORY_ADDR,
		abi: abi,
		functionName: 'getCurrencyConfig',
		args: [feeCurrency],
	});
	return intrinsicGas;
}
