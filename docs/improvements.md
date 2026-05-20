# Improvements

- Move the address generator to `Manager` so we do not expose `WALLET_SEED`, which is highly sensitive. `Manager` should not expose ports to external networks.
- Restructure folders so TypeScript and Go code are not mixed. We can evaluate whether BTC scripts should be merged with EVM scripts.
- Change module import path from `deposit-collector` to `github.com/alex99y/deposit-collector`.
- Canonical module naming: `Manager`, `Collector`, and `Processor`.
- ChainsCache should have a TTL so cached information is refreshed every T minutes.