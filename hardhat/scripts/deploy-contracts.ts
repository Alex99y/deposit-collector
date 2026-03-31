import { network } from "hardhat";

// `network: "node"` is an in-process EDR chain; it is NOT the JSON-RPC server from `hardhat node`.
// Deploy via `localhost` so contracts exist on http://127.0.0.1:8545 (same RPC as send_erc20.ts).
const { viem, networkName } = await network.connect({ network: "localhost" });

const publicClient = await viem.getPublicClient();

console.log(`Deploying to "${networkName}" (chain id ${publicClient.chain.id})...`);

const usdc = await viem.deployContract("USDC");
console.log("USDC:", usdc.address);

const kkk = await viem.deployContract("KKK");
console.log("KKK:", kkk.address);

const permitAndTransfer = await viem.deployContract("PermitAndTransfer");
console.log("PermitAndTransfer:", permitAndTransfer.address);

console.log("Deployment successful.");
