#!/usr/bin/env node

// makeTestFromRPC generates a tracer test from an RPC transaction
// Usage: node makeTestFromRPC.js <txHash> <tracerName> [tracerConfig]
//
// Code formatted with `npx @biomejs/biome format --write eth/tracers/internal/tracetest/testdata/makeTestFromRPC.js`

const { execSync } = require("child_process");
const fs = require("fs");

function rpcCall(method, params) {
	// For cast rpc, we need to pass each parameter individually
	let paramArgs = "";
	if (params && params.length > 0) {
		paramArgs = params
			.map((p) => {
				if (typeof p === "string") {
					return `'${p}'`;
				} else {
					return `'${JSON.stringify(p)}'`;
				}
			})
			.join(" ");
	}

	const cmd = `cast rpc ${method} ${paramArgs}`;
	const result = execSync(cmd, { encoding: "utf8" });
	return JSON.parse(result);
}

function isCeloChain(chainId) {
	return (
		chainId === 42220 ||
		chainId === 44787 ||
		chainId === 62320 ||
		chainId === 11142220
	);
}

// Generate appropriate chain config based on chain ID and block number
function getChainConfig(chainId) {

	// Base config that all chains share
	const baseConfig = {
		chainId: chainId,
		homesteadBlock: 0,
		eip150Block: 0,
		eip155Block: 0,
		eip158Block: 0,
		byzantiumBlock: 0,
		constantinopleBlock: 0,
		petersburgBlock: 0,
		istanbulBlock: 0,
		muirGlacierBlock: 0,
		berlinBlock: 0,
		londonBlock: 0,
		arrowGlacierBlock: 0,
		grayGlacierBlock: 0,
		shanghaiTime: 0,
		terminalTotalDifficulty: 0,
		terminalTotalDifficultyPassed: true,
	};

	// Celo-specific config
	if (isCeloChain(chainId)) {
		// Celo networks (Mainnet, Sepolia)
		return {
			...baseConfig,
			cel2Time: 0,
			gingerbreadBlock: 0,
		};
	}

	// Default for other chains
	return baseConfig;
}

function main() {
	const args = process.argv.slice(2);
	if (args.length < 2) {
		console.error(
			`Usage: node makeTestFromRPC.js <txHash> <tracerName> [tracerConfig]
Example: node makeTestFromRPC.js $TXHASH prestateTracer > testdata/prestate_tracer/newtest.json
Example: node makeTestFromRPC.js $TXHASH callTracer '{\"withLog\": true}

Notes:
	* Ensure your RPC endpoint is set in the environment variable ETH_RPC_URL.
	* The fee currency context needs to be set manually if used by the tx.
	* The chain configuration is only a heuristic, but should work for most cases.`,
		);
		process.exit(1);
	}

	const txHash = args[0];
	const tracerName = args[1];
	const tracerConfig = args[2] ? JSON.parse(args[2]) : {};

	console.error(
		`Generating test for transaction ${txHash} with tracer ${tracerName}...`,
	);

	// Get transaction data
	const tx = rpcCall("eth_getTransactionByHash", [txHash]);
	if (!tx) {
		throw new Error(`Transaction ${txHash} not found`);
	}

	// Get block data
	const block = rpcCall("eth_getBlockByHash", [tx.blockHash, false]);
	if (!block) {
		throw new Error(`Block ${tx.blockHash} not found`);
	}

	// Get raw transaction
	const rawTx = rpcCall("eth_getRawTransactionByHash", [txHash]);

	// Get prestate
	console.error("Getting prestate...");
	const prestate = rpcCall("debug_traceTransaction", [
		txHash,
		{
			tracer: "prestateTracer",
		},
	]);

	// Get trace result
	console.error("Getting trace result...");
	const traceResult = rpcCall("debug_traceTransaction", [
		txHash,
		{
			tracer: tracerName,
			tracerConfig: tracerConfig,
		},
	]);

	// Clean up trace result (remove timing info)
	if (traceResult.time) {
		delete traceResult.time;
	}

	// Build genesis block from prestate
	const chainId = parseInt(tx.chainId, 16);
	const genesis = {
		alloc: prestate,
		config: getChainConfig(chainId),
		difficulty: block.difficulty,
		extraData: block.extraData,
		gasLimit: block.gasLimit,
		hash: block.hash,
		miner: block.miner,
		mixHash: block.mixHash,
		nonce: block.nonce,
		number: block.number,
		stateRoot: block.stateRoot,
		timestamp: block.timestamp,
		totalDifficulty: block.totalDifficulty || "0x1",
		withdrawals: block.withdrawals || [],
		withdrawalsRoot:
			block.withdrawalsRoot ||
			"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
	};

	// Convert hex strings to decimal strings for numbers
	genesis.gasLimit = parseInt(genesis.gasLimit, 16).toString();
	genesis.number = parseInt(genesis.number, 16).toString();
	genesis.timestamp = parseInt(genesis.timestamp, 16).toString();
	genesis.difficulty = parseInt(genesis.difficulty, 16).toString();
	genesis.totalDifficulty = parseInt(genesis.totalDifficulty, 16).toString();

	// Handle nonces in prestate
	for (const address in genesis.alloc) {
		const account = genesis.alloc[address];
		if (account.nonce) {
			account.nonce = parseInt(account.nonce, 16).toString();
		}
	}

	// Build execution context
	const context = {
		number: parseInt(block.number, 16).toString(),
		difficulty: genesis.difficulty,
		timestamp: parseInt(block.timestamp, 16).toString(),
		gasLimit: parseInt(block.gasLimit, 16).toString(),
		miner: block.miner,
	};

	// Add base fee if present (EIP-1559)
	if (block.baseFeePerGas) {
		context.baseFeePerGas = parseInt(block.baseFeePerGas, 16).toString();
	}

	// Add Celo-specific fee currency context
	if (isCeloChain(chainId)) {
		// For Celo networks, add a default fee currency context
		// This may need to be customized based on the specific transaction
		context.feeCurrencyContext = {
			exchangeRates: {},
			intrinsicGasCosts: {},
		};
	}

	// Build the test case
	const testCase = {
		genesis: genesis,
		context: context,
		input: rawTx,
		result: traceResult,
	};

	// Add tracer config if provided
	if (Object.keys(tracerConfig).length > 0) {
		testCase.tracerConfig = tracerConfig;
	}

	// Output the test
	console.log(JSON.stringify(testCase, null, 2));
}

main();
