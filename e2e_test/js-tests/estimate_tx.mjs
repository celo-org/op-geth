#!/usr/bin/env node
import { publicClient, account } from "./viem_setup.mjs"

const [chainId, celoValue, feeCurrency] = process.argv.slice(2);

async function main() {
  let bigCeloValue = BigInt(celoValue)

  let result = await publicClient.estimateGas({
    account,
    to: "0x00000000000000000000000000000000DeaDBeef",
    value: bigCeloValue,
    feeCurrency
  });
  console.log(result.toString())

  return result;
}

await main();
process.exit(0);
