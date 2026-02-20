CREATE TYPE chain_platform AS ENUM ('EVM', 'BTC', 'SOL');

CREATE TYPE operation_type AS ENUM ('deposit', 'withdraw');

CREATE TYPE network_type AS ENUM ('ethereum', 'base', 'avalanche', 'polygon', 'bitcoin', 'solana');

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- Account ID is the index of the derivation path
    -- Example: m/44'/60'/{account_id}'/0/0 for Ethereum
    account_id SERIAL UNIQUE,
    -- External ID is the ID of the user in the external system
    external_id VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_external_id ON users (external_id);

CREATE TABLE token_addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    unit_name VARCHAR(25) NOT NULL,
    unit_symbol VARCHAR(10) NOT NULL,
    address VARCHAR(100) NOT NULL,
    chain network_type NOT NULL,
    decimals INTEGER NOT NULL,
    UNIQUE (address, chain)
);

CREATE TABLE user_balances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    available_balance BIGINT NOT NULL,
    locked_balance BIGINT DEFAULT 0 NOT NULL,
    token_address_id UUID NOT NULL REFERENCES token_addresses(id),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, token_address_id)
);

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