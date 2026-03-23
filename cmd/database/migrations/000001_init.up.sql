CREATE TYPE chain_platform AS ENUM ('EVM', 'BTC', 'SOL');

CREATE TYPE operation_type AS ENUM ('deposit', 'withdraw');


-- Supported chains table stores the chains that will be used in the system
-- Example: Chain name is Ethereum. BIP44 ID is 60. Chain platform is EVM.

CREATE TABLE supported_chains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chain_name VARCHAR(100) NOT NULL,
    chain_platform chain_platform NOT NULL,
    -- EVM chain ID is the ID of the chain in the EVM chain
    -- Example: 1 for Ethereum mainnet
    -- @TODO: This is only for EVM chains. This table should be generic for all chains and
    -- not have columns for specific chains.
    evm_chain_id INTEGER,

    CONSTRAINT supported_chains_chain_name_uk UNIQUE (chain_name),
    CONSTRAINT supported_chains_evm_chain_id_uk UNIQUE (evm_chain_id)
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
    decimals INTEGER NOT NULL
);

CREATE INDEX idx_token_addresses_address_chain_id ON token_addresses (address, chain_id);


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
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX user_balances_user_id_token_address_id_uk ON user_balances (user_id, token_address_id);

-- This is the main feature of the system. It stores the addresses of the users. Each address is unique for each one.
-- It is referenced by the user accountId and the sequenceNumber. SequenceNumber is the index of the derivation path
-- Example: m/44'/60'/{account_id}'/0/{sequence_number} for Ethereum

CREATE TABLE user_addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- Address is the public address of the user
    -- Example: 0x1234567890123456789012345678901234567890 for Ethereum
    address VARCHAR(100) NOT NULL,
    -- Sequence number is the index of the derivation path
    -- Example: m/44'/60'/0'/0/{sequence_number} for Ethereum
    sequence_number INTEGER NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id),
    chain chain_platform NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX user_addresses_address_uk ON user_addresses (address);
CREATE UNIQUE INDEX user_addresses_user_chain_uk ON user_addresses (user_id, chain);

-- Operations table stores the operations of the users in the system
-- Example: User with ID 1234567890 has deposited 100 USDC to address 0x1234567890123456789012345678901234567890 for Ethereum
-- Amount is the amount of the operation
-- Type is the type of the operation
-- Created at is the timestamp of the operation

CREATE TABLE operations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    address_id UUID NOT NULL REFERENCES user_addresses(id),
    token_address_id UUID NOT NULL REFERENCES token_addresses(id),
    amount BIGINT NOT NULL,
    type operation_type NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
    tx_hash VARCHAR(100) NOT NULL
);

CREATE UNIQUE INDEX idx_operations_user_created ON operations (user_id, created_at DESC);
CREATE UNIQUE INDEX idx_operations_tx_hash ON operations (tx_hash);


-- Pending deposit operations table stores the references to the addresses that hold funds waiting to be transferred to
-- the master address. The master address is the protected, multisig, cold wallet address owned by the institution.
CREATE TABLE pending_deposit_operations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    address_id UUID NOT NULL REFERENCES user_addresses(id),
    accumulated_amount BIGINT NOT NULL,
    token_address_id UUID NOT NULL REFERENCES token_addresses(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMPTZ NULL,

    CONSTRAINT pending_deposit_operations_address_id_token_address_id_uk UNIQUE (address_id, token_address_id),
);

CREATE INDEX idx_pending_deposit_operations_token_address_id ON pending_deposit_operations (token_address_id, accumulated_amount DESC);
