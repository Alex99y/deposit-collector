# Configuration

This document describes the environment variables required by Deposit Collector.
Use it as a safe reference and **never commit real secrets** to source control.

## Example `.env` Template (Safe Values)

```bash
# Infrastructure, shared by all modules.
RABBITMQ_URL=amqp://<user>:<password>@<host>:5672/
POSTGRESQL_URL=postgres://<user>:<password>@<host>:5432/<database>
METRICS_PORT=2112
BITCOIN_NETWORK=local
WALLET_SEED=<hex_seed_or_mnemonic>

# API module
PORT=8000
HOST=0.0.0.0

# Manager module
ALLOW_MULTI_THREADING=true
MAX_WORKERS=2
RPC_FILE_PATH=config/local.json
DEPOSIT_COLLECTOR_INTERVAL=10s
DEPOSIT_COLLECTOR_EVM_DESTINATION_DEPOSIT_ADDRESS=<evm_destination_address>
DEPOSIT_COLLECTOR_BTC_DESTINATION_DEPOSIT_ADDRESS=<btc_destination_address>
DEPOSIT_COLLECTOR_SOL_DESTINATION_DEPOSIT_ADDRESS=<sol_destination_address>
WITHDRAW_COLLECTOR_EVM_PRIVATE_KEY=<evm_private_key>
WITHDRAW_COLLECTOR_BTC_PRIVATE_KEY=<btc_private_key>
WITHDRAW_COLLECTOR_SOL_PRIVATE_KEY=<sol_private_key>
```

## Variables

### Infrastructure

- `RABBITMQ_URL`: RabbitMQ connection string.
- `POSTGRESQL_URL`: PostgreSQL connection string used by the application.
- `METRICS_PORT`: Port where Prometheus/metrics endpoint is exposed. Defaults to `2112` if unset or empty.

### Network and Wallet

- `BITCOIN_NETWORK`: Bitcoin network mode (for example: `local`, `testnet`, `mainnet`).
- `WALLET_SEED`: Seed used to derive internal wallets. Treat as highly sensitive.

### API

- `PORT`: HTTP API port.
- `HOST`: Interface/IP where the API server listens.

### Manager

- `RPC_FILE_PATH`: Path to the chain RPC configuration file. Required (manager exits if unset).
- `ALLOW_MULTI_THREADING`: When set to the literal string `true`, enables concurrent processing in manager workers. Defaults to disabled.
- `MAX_WORKERS`: Number of worker routines when multithreading is enabled. Defaults to `1`; must be greater than zero.
- `DEPOSIT_COLLECTOR_INTERVAL`: Polling interval for deposit collection (for example: `10s`). Defaults to `10s` if unset or empty.
- `DEPOSIT_COLLECTOR_EVM_DESTINATION_DEPOSIT_ADDRESS`: Destination EVM address for collected funds. Optional; empty if unset.
- `DEPOSIT_COLLECTOR_BTC_DESTINATION_DEPOSIT_ADDRESS`: Destination BTC address for collected funds. Optional; empty if unset.
- `DEPOSIT_COLLECTOR_SOL_DESTINATION_DEPOSIT_ADDRESS`: Destination SOL address for collected funds. Optional; empty if unset.
- `WITHDRAW_COLLECTOR_EVM_PRIVATE_KEY`: Private key used for EVM withdraw collection. Highly sensitive. Optional; empty if unset.
- `WITHDRAW_COLLECTOR_BTC_PRIVATE_KEY`: Private key used for Bitcoin withdraw collection. Highly sensitive. Optional; empty if unset.
- `WITHDRAW_COLLECTOR_SOL_PRIVATE_KEY`: Private key used for Solana withdraw collection. Highly sensitive. Optional; empty if unset.

## Security Notes

- Keep real `.env` files out of git (add `.env` to `.gitignore`).
- Use separate credentials per environment (`local`, `staging`, `production`).
- Rotate keys/seeds immediately if they are ever exposed.
