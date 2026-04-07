import { createWalletClient, http, createPublicClient, formatUnits } from "viem";
import { mainnet } from "viem/chains";
import { privateKeyToAccount } from "viem/accounts";
import ERC20Abi from "./abi/erc20.json";

import privateKeys from "../../common/privatekeys.json";
import { argv } from "process";

// USDC: 0x8fe7d924bd0ea1f30560237c15ea66d4d36d8162
// KKK: 0x89cff09053429eea9c440c56cdc69539bcfc8c19
// PermitAndTransfer: 0x92a2df6f945732aab8a26e0ca322b9efcdf3ef3f
const TOKEN = "0x8fe7d924bd0ea1f30560237c15ea66d4d36d8162" as const;

const toArg = argv[2];
const amountArg = argv[3];
if (!toArg || !amountArg) {
  console.error("Usage: npm run send:erc20 -- <to> <amount>");
  console.error("  <amount> is in whole tokens (uses on-chain decimals).");
  process.exit(1);
}

const wallet = privateKeyToAccount(privateKeys[0].privateKey as `0x${string}`);

const transport = http("http://localhost:8545");

const publicClient = createPublicClient({ chain: mainnet, transport });
const walletClient = createWalletClient({ chain: mainnet, transport, account: wallet });

const toAddress = toArg as `0x${string}`;
const amount = BigInt(amountArg);

const { request } = await publicClient.simulateContract({
  address: TOKEN,
  abi: ERC20Abi,
  functionName: "transfer",
  args: [toAddress, amount],
  account: wallet,
});

const tx = await walletClient.writeContract(request);

await publicClient.waitForTransactionReceipt({ hash: tx });

const balance = await publicClient.readContract({
  address: TOKEN,
  abi: ERC20Abi,
  functionName: "balanceOf",
  args: [toAddress],
});

console.log("Balance:", formatUnits(balance as bigint, 18));
console.log(tx);