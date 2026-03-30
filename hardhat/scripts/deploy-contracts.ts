import { network } from "hardhat";

const { viem, networkName } = await network.connect({ network: "node" });

const publicClient = await viem.getPublicClient();

console.log(`Deploying to "${networkName}" (chain id ${publicClient.chain.id})...`);

const usdc = await viem.deployContract("USDC");
console.log("USDC:", usdc.address);

const kkk = await viem.deployContract("KKK");
console.log("KKK:", kkk.address);

const permitAndTransfer = await viem.deployContract("PermitAndTransfer");
console.log("PermitAndTransfer:", permitAndTransfer.address);

console.log("Deployment successful.");
