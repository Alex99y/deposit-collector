# Deposit Collector

This is a personal project created to learn and improve my experience with popular technologies and the Web3 ecosystem.  
This code is not intended for production use; treat it as an example/PoC.  
The project is still a work in progress.

Deposit Collector (DC) is a component designed to be integrated into a complete user-managed service (for example, a fintech or bank platform).  
This service maps external users from the integrating system to DC users.  
DC users can generate as many custodial accounts as needed across EVM, BTC, and SOL networks.

Users can deposit funds into previously generated DC custodial accounts, withdraw funds, or transfer funds between users.  
DC monitors blockchains for incoming deposits and processes them.  
Incoming deposits are handled by a module called `Manager`, which validates and approves them before forwarding funds to an external cold wallet after a short delay.

## Index

[Configuration](./docs/configs.md)  
[Improvements](./docs/improvements.md)  
[TODO](./docs/todo.md)  

## Software Requirements
- Go (>= 1.25.7)
- Postgresql (>= 17.8)
- Rabbitmq (>= 4.2.4)
- ElectrumX indexer
- Envio (Soon)

## LICENSE

This software is distributed under the Apache License.
