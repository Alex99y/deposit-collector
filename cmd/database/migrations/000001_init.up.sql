CREATE TYPE chain_platform AS ENUM ('EVM', 'BTC', 'SOL');

CREATE TYPE operation_type AS ENUM ('deposit', 'withdraw');


-- Supported chains table stores the networks that will be used in the system
-- Example: Network is Ethereum. BIP44 ID is 60. Chain platform is EVM.

CREATE TABLE supported_chains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    network VARCHAR(100) UNIQUE NOT NULL,
    chain_platform chain_platform NOT NULL,
    bip44_coin_type INTEGER UNIQUE NOT NULL,
    -- EVM chain ID is the ID of the chain in the EVM network
    -- Example: 1 for Ethereum mainnet
    -- @TODO: This is only for EVM chains. This table should be generic for all chains and
    -- not have columns for specific chains.
    evm_chain_id INTEGER UNIQUE
);


-- Users table stores the users of the system
-- This system is not meant to manage users, it only associates external users with internal accounts.
-- Example: User with external ID 1234567890 has account ID 1.
-- Account ID is the index of the derivation path
-- Example: m/44'/60'/{account_id}'/0/0 for Ethereum

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- Account ID is the index of the derivation path
    -- Example: m/44'/60'/{account_id}'/0/0 for Ethereum
    account_id INTEGER UNIQUE NOT NULL,
    -- External ID is the ID of the user in the external system
    external_id VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT users_account_id_uk UNIQUE (account_id),
    CONSTRAINT users_external_id_uk UNIQUE (external_id)
);

CREATE INDEX idx_users_external_id ON users (external_id);


-- Token addresses table stores the addresses of the tokens that will be used in the system
-- Example: USDC token. Address is 0x1234567890123456789012345678901234567890.

CREATE TABLE token_addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    unit_name VARCHAR(25) NOT NULL,
    unit_symbol VARCHAR(10) NOT NULL,
    address VARCHAR(100) NOT NULL,
    -- ChainId references the supported_chains table because not all tokens are supported on all chains.
    chain_id UUID NOT NULL REFERENCES supported_chains(id),
    decimals INTEGER NOT NULL,
    UNIQUE (address, chain_id)
);


-- User balances table stores the balances of the users in the system
-- Example: User with ID 1234567890 has 100 USDC of available balance and 5 USDC of locked balance
-- locked balance is the balance that is not available for withdrawal. It can be because there is a pending withdrawal
-- or because the funds were blocked by the system for some reason.

CREATE TABLE user_balances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    available_balance BIGINT NOT NULL,
    locked_balance BIGINT DEFAULT 0 NOT NULL,
    token_address_id UUID NOT NULL REFERENCES token_addresses(id),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, token_address_id)
);


-- This is the main feature of the system. It stores the addresses of the users. Each address is unique for each one.
-- It is referenced by the user accountId and the sequenceNumber. SequenceNumber is the index of the derivation path
-- Example: m/44'/60'/{account_id}'/0/{sequence_number} for Ethereum

CREATE TABLE addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- Address is the public address of the user
    -- Example: 0x1234567890123456789012345678901234567890 for Ethereum
    address VARCHAR(100) NOT NULL,
    -- Sequence number is the index of the derivation path
    -- Example: m/44'/60'/0'/0/{sequence_number} for Ethereum
    sequence_number INTEGER NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id),
    chain chain_platform NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (address, chain)
);

CREATE INDEX idx_addresses_user_id ON addresses (user_id);
CREATE UNIQUE INDEX idx_addresses_user_address_chain ON addresses (user_id, address, chain);


-- Operations table stores the operations of the users in the system
-- Example: User with ID 1234567890 has deposited 100 USDC to address 0x1234567890123456789012345678901234567890 for Ethereum
-- Amount is the amount of the operation
-- Type is the type of the operation
-- Created at is the timestamp of the operation

CREATE TABLE operations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    address_id UUID NOT NULL REFERENCES addresses(id),
    token_address_id UUID NOT NULL REFERENCES token_addresses(id),
    amount BIGINT NOT NULL,
    type operation_type NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_operations_user_created ON operations (user_id, created_at DESC);