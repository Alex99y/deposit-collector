# Configuration

This document describes the environment variables required by Deposit Collector.
Use it as a safe reference and **never commit real secrets** to source control.

## Example `.env` Template (Safe Values)

```bash
# Infrastructure, shared by all modules.
RABBITMQ_URL=amqp://<user>:<password>@<host>:5672/
POSTGRESQL_URL=postgres://<user>:<password>@<host>:5432/<database>
METRICS_PORT=9090
BITCOIN_NETWORK=local
WALLET_SEED=<hex_seed_or_mnemonic>

# API module
PORT=8000
HOST=0.0.0.0

# Manager module
ALLOW_MULTI_THREADING=true
MAX_WORKERS=2
RPC_FILE_PATH=config/local.json
EVM_FEE_PAYER_PRIVATE_KEY=<evm_private_key>
DEPOSIT_COLLECTOR_INTERVAL=10s
DEPOSIT_COLLECTOR_EVM_DESTINATION_DEPOSIT_ADDRESS=<evm_destination_address>
DEPOSIT_COLLECTOR_BTC_DESTINATION_DEPOSIT_ADDRESS=<btc_destination_address>
DEPOSIT_COLLECTOR_SOL_DESTINATION_DEPOSIT_ADDRESS=<sol_destination_address>
```

## Variables

### Infrastructure

- `RABBITMQ_URL`: RabbitMQ connection string.
- `POSTGRESQL_URL`: PostgreSQL connection string used by the application.
- `METRICS_PORT`: Port where Prometheus/metrics endpoint is exposed.

### Network and Wallet

- `BITCOIN_NETWORK`: Bitcoin network mode (for example: `local`, `testnet`, `mainnet`).
- `WALLET_SEED`: Seed used to derive internal wallets. Treat as highly sensitive.

### API

- `PORT`: HTTP API port.
- `HOST`: Interface/IP where the API server listens.

### Manager

- `ALLOW_MULTI_THREADING`: Enables concurrent processing in manager workers.
- `MAX_WORKERS`: Number of worker routines when multithreading is enabled.
- `RPC_FILE_PATH`: Path to the chain RPC configuration file.
- `EVM_FEE_PAYER_PRIVATE_KEY`: Private key used to pay EVM gas fees. Highly sensitive.
- `DEPOSIT_COLLECTOR_INTERVAL`: Polling interval for deposit collection (for example: `10s`).
- `DEPOSIT_COLLECTOR_EVM_DESTINATION_DEPOSIT_ADDRESS`: Destination EVM address for collected funds.
- `DEPOSIT_COLLECTOR_BTC_DESTINATION_DEPOSIT_ADDRESS`: Destination BTC address for collected funds.
- `DEPOSIT_COLLECTOR_SOL_DESTINATION_DEPOSIT_ADDRESS`: Destination SOL address for collected funds.

## Security Notes

- Keep real `.env` files out of git (add `.env` to `.gitignore`).
- Use separate credentials per environment (`local`, `staging`, `production`).
- Rotate keys/seeds immediately if they are ever exposed.
