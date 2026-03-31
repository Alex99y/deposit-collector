#!/bin/bash

# Start hardhat node in the background
echo "Starting Hardhat node..."
npx hardhat node --config hardhat.config.ts &
HARDHAT_PID=$!

# Wait until JSON-RPC accepts connections (separate from in-process `edr-simulated` used by `hardhat run` alone).
echo "Waiting for Hardhat node to be ready..."
for i in $(seq 1 30); do
  if curl -s -X POST -H 'Content-Type: application/json' \
    --data '{"jsonrpc":"2.0","method":"eth_chainId","id":1}' \
    http://127.0.0.1:8545 >/dev/null 2>&1; then
    break
  fi
  sleep 0.5
done

# Build contracts
npx hardhat build
npx hardhat run scripts/deploy-contracts.ts

wait $HARDHAT_PID