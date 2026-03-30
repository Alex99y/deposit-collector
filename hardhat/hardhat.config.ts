import hardhatViem from "@nomicfoundation/hardhat-viem";
import { defineConfig } from "hardhat/config";
import privateKeys from "../common/privatekeys.json";

export default defineConfig({
  plugins: [hardhatViem],
  paths: {
    sources: {
      solidity: ["contracts/src"],
    },
  },
  solidity: {
    version: "0.8.28",
  },
  networks: {
    node: {
      type: "edr-simulated",
      chainId: 1,
      blockGasLimit: 30_000_000, // Default value
      initialBaseFeePerGas: 1,
      loggingEnabled: false,
      accounts: privateKeys.map((wallet: { privateKey: string }) => ({
        privateKey: wallet.privateKey,
        balance: '10000000000000000000000', // 10,000 ETH
      })),
      mining: {
        auto: true,
        interval: 1000,
      }
    }
  }
});
