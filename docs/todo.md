# TODO

- Add metrics and alerts to each module [WIP]
- Test, test, and more tests (yes, I know) [WIP]
- Transfer funds between users
- Add compliance logic to `Manager` so we do not accept/send funds from/to blacklisted addresses
- Ability to freeze funds or block accounts (yes, I'm sorry)
- Solana `Manager` implementation (`Processor` and `Collector` services)
- EVM `Manager` implementation (`Collector` service)
- Withdraw `Manager` implementation (for all platforms)
- Add an API token key to restrict access to the DC component (not a full auth system)
- Indexer: this is the most difficult part and involves complex implementation. A good idea is using ENVIO indexer services to monitor all supported blockchains.
- ChainsCache should have a TTL so cached information is refreshed every T minutes.