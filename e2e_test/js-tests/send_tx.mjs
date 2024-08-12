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

const [chainId, privateKey, feeCurrency] = process.argv.slice(2);
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

const account = privateKeyToAccount(privateKey);

const publicClient = createPublicClient({
  account,
  chain: devChain,
  transport: http(),
});
const walletClient = createWalletClient({
  account,
  chain: devChain,
  transport: http(),
});

async function getTransactionReceiptAfterWait(hash, numBlocks) {
  var count = 0;

  const res = new Promise((resolve) => {
    resolve();
  });
  const unwatch = publicClient.watchBlockNumber({
    onBlockNumber: () => {
      count++;
      if (count >= numBlocks) {
        res.resolve();
      }
    },
  });
  await res;
  unwatch();
  try {
    return await publicClient.getTransactionReceipt({ hash: hash });
  } catch (e) {
    if (e instanceof TransactionReceiptNotFoundError) {
      return undefined;
    }
    throw e;
  }
}

async function replaceTx(tx) {
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
    confirmations: 2,
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
  try {
    hash = await walletClient.sendRawTransaction({
      serializedTransaction: await walletClient.signTransaction(request),
    });
  } catch (e) {
    // direct revert
    console.log(
      JSON.stringify({
        success: false,
        error: e,
      }),
    );
    return;
  }

  var success = true;
  var receipt = await getTransactionReceiptAfterWait(hash, 2);
  if (!receipt) {
    receipt = await replaceTx(request);
    success = false;
  }
  // print for bash script wrapper return value
  console.log(
    JSON.stringify({
      success: success,
      error: null,
    }),
  );

  return receipt;
}
await main();
