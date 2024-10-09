#!/usr/bin/env node
import {
  createWalletClient,
  createPublicClient,
  http,
  defineChain,
  TransactionReceiptNotFoundError,
} from "viem";
import { celoAlfajores } from "viem/chains";
import { privateKeyToAccount } from "viem/accounts";

const [chainId, privateKey, feeCurrency, waitBlocks, replaceTxAfterWait] =
  process.argv.slice(2);
const devChain = defineChain({
  ...celoAlfajores,
  id: parseInt(chainId, 10),
  name: "local dev chain",
  network: "dev",
  rpcUrls: {
    default: {
      http: ["http://127.0.0.1:8545"],
    },
  },
});

var chain;
switch(process.env.NETWORK) {
	case "alfajores":
		chain = celoAlfajores;
		break;
	default:
		chain = devChain;
		// Code to run if no cases match
};

const account = privateKeyToAccount(privateKey);

const publicClient = createPublicClient({
  account,
  chain: chain,
  transport: http(),
});
const walletClient = createWalletClient({
  account,
  chain: chain,
  transport: http(),
});

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
async function waitUntilBlock(blocknum) {
  var next = await publicClient.getBlockNumber({ cacheTime: 0 });
  while (next < blocknum) {
    await sleep(500);
    next = await publicClient.getBlockNumber({ cacheTime: 0 });
  }
}

async function getTransactionReceipt(hash) {
  try {
    return await publicClient.getTransactionReceipt({ hash: hash });
  } catch (e) {
    if (e instanceof TransactionReceiptNotFoundError) {
      return undefined;
    }
    throw e;
  }
}

async function replaceTransaction(tx) {
  const request = await walletClient.prepareTransactionRequest({
    account: tx.account,
    to: account.address,
    value: 0n,
    gas: 21000,
    nonce: tx.nonce,
    maxFeePerGas: tx.maxFeePerGas,
    maxPriorityFeePerGas: tx.maxPriorityFeePerGas + 1000n,
  });
  const hash = await walletClient.sendRawTransaction({
    serializedTransaction: await walletClient.signTransaction(request),
  });
  const receipt = await publicClient.waitForTransactionReceipt({
    hash: hash,
    confirmations: 1,
  });
  return receipt;
}

async function main() {
  const request = await walletClient.prepareTransactionRequest({
    account,
    to: "0x00000000000000000000000000000000DeaDBeef",
    value: 2n,
    gas: 90000,
    feeCurrency,
    maxFeePerGas: 2000000000n,
    maxPriorityFeePerGas: 0n,
  });

  var hash;

  var blocknum = await publicClient.getBlockNumber({ cacheTime: 0 });
  var replaced = false;
  try {
    hash = await walletClient.sendRawTransaction({
      serializedTransaction: await walletClient.signTransaction(request),
    });
  } catch (e) {
    // direct revert
    console.log(
      JSON.stringify({
        success: false,
        replaced: replaced,
        error: e,
      }),
    );
    return;
  }

  var success = true;
  var waitBlocksForReceipt = parseInt(waitBlocks);
  var receipt = await getTransactionReceipt(hash);
  while (waitBlocksForReceipt > 0) {
    await waitUntilBlock(blocknum + BigInt(1));
    waitBlocksForReceipt--;
    var receipt = await getTransactionReceipt(hash);
  }
  if (!receipt) {
    if (replaceTxAfterWait == "true") {
      receipt = await replaceTransaction(request);
    }
    success = false;
  }
  // print for bash script wrapper return value
  console.log(
    JSON.stringify({
      success: success,
      replaced: replaced,
      error: null,
    }),
  );

  return receipt;
}
await main();
