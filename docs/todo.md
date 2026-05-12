# TODO

- Add metrics and alerts to each module [WIP]
- Test, test, and more tests (yes, I know) [WIP]
- Transfer funds between users
- Add compliance logic to `Manager` so we do not accept/send funds from/to blacklisted addresses
- Ability to freeze funds or block accounts (yes, I'm sorry)
- Solana `Manager` implementation (`Processor` and `Collector` services)
- EVM `Manager` implementation (`Collector` service)
    - Only for ERC-20 tokens, native transfer is implemented
- Withdraw `Manager` implementation
    - Bitcoin withdraw processor is done. Bitcoin withdraw collector is still in WIP.
    - **BTC withdraw collector — two-phase persistence (see `collectBTCUnprocessedWithdrawals`):** (1) Compute `txid` locally before broadcast (use `btc_utils.BitcoinTxIDFromSerializedHex` on the signed raw hex), broadcast, validate the returned hash against that `txid`, then persist `txHash` / mark operations in the DB. (2) In a separate process, confirm the tx is committed on-chain; set `processed_at` when confirmed, or if the tx never commits / is dropped, clear the stored `txHash` from the registry as appropriate.
    - For EVM and SOL, are not implemented yet.
- **Security — withdraw authorization:** Withdraw endorsement today trusts AMQP payload fields (`UserDbID`, `TokenAddressDbId`, `TargetAddress`, amount). A forged message could debit a victim’s balance and send funds to an attacker-controlled destination if someone can publish to the queue. Mitigate by re-resolving/authorizing the user, token, and destination from an authenticated request (or an internal idempotent server-side record created only after auth), or by verifying a signed command at the `Manager` boundary; do not treat the queue payload alone as proof of intent.
- Add an API token key to restrict access to the DC component (not a full auth system)
- Indexer: this is the most difficult part and involves complex implementation. A good idea is using ENVIO indexer services to monitor all supported blockchains.
- ChainsCache should have a TTL so cached information is refreshed every T minutes.