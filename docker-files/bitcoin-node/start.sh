#!/bin/sh
set -e

bitcoind -printtoconsole -regtest=1 -rpcport=18443 -rpcbind=0.0.0.0 -rpcallowip=0.0.0.0/0 &
BITCOIND_PID=$!
MINER_PID=""

cleanup() {
  echo "Stopping bitcoin-node..."
  if [ -n "$MINER_PID" ]; then
    kill "$MINER_PID" >/dev/null 2>&1 || true
    wait "$MINER_PID" 2>/dev/null || true
  fi
  bitcoin-cli -regtest -rpcuser=admin -rpcpassword=admin -rpcport=18443 stop >/dev/null 2>&1 || true
  wait "$BITCOIND_PID" 2>/dev/null || true
}

trap cleanup INT TERM

until bitcoin-cli -regtest -rpcuser=admin -rpcpassword=admin -rpcport=18443 getblockchaininfo >/dev/null 2>&1; do
  sleep 1
done

WALLET_NAME="miner"

# Bitcoin Core no longer creates a default wallet automatically.
# Try to load an existing wallet; if it doesn't exist, create it.
if ! bitcoin-cli -regtest -rpcuser=admin -rpcpassword=admin -rpcport=18443 loadwallet "$WALLET_NAME" >/dev/null 2>&1; then
  bitcoin-cli -regtest -rpcuser=admin -rpcpassword=admin -rpcport=18443 createwallet "$WALLET_NAME" >/dev/null
fi

ADDRESS=$MINER_ADDRESS
echo "Mining address: $ADDRESS"

(
  while true; do
    bitcoin-cli -regtest -rpcuser=admin -rpcpassword=admin -rpcport=18443 -rpcwallet="$WALLET_NAME" generatetoaddress 1 "$ADDRESS"
    sleep 2
  done
) &
MINER_PID=$!

wait "$BITCOIND_PID"