import { createWalletClient, http, parseEther } from "viem";
import { mainnet } from "viem/chains";
import { privateKeyToAccount } from "viem/accounts";

import privateKeys from "../../common/privatekeys.json";
import { argv } from "process";

const wallet = privateKeyToAccount(privateKeys[0].privateKey as `0x${string}`);

const client = createWalletClient({
  chain: mainnet,
  transport: http("http://localhost:8545"),
  account: wallet,
});

const tx = await client.sendTransaction({
  to: argv[2] as `0x${string}`,
  value: parseEther(argv[3]),
});

console.log(tx);