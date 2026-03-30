#!/bin/bash

# Start hardhat node in the background
echo "Starting Hardhat node..."
npx hardhat node --config hardhat.config.ts &
HARDHAT_PID=$!

# Wait for hardhat to be ready
echo "Waiting for Hardhat node to be ready..."
sleep 5

# Build contracts
npx hardhat build
npx hardhat run scripts/deploy-contracts.ts

wait $HARDHAT_PID