-- Insert default chains
INSERT INTO supported_chains (
    chain_name, chain_platform, evm_chain_id
) VALUES (
    'bitcoin', 'BTC', NULL
);

INSERT INTO supported_chains (
    chain_name, chain_platform, evm_chain_id
) VALUES (
    'ethereum', 'EVM', 1
);

-- Insert native token addresses
INSERT INTO token_addresses (
    unit_name, unit_symbol, address, chain_id, decimals
) VALUES (
    'Bitcoin', 'BTC', 'native', (SELECT id FROM supported_chains WHERE chain_name = 'bitcoin'), 8
);

INSERT INTO token_addresses (
    unit_name, unit_symbol, address, chain_id, decimals
) VALUES (
    'Ethereum', 'ETH', 'native', (SELECT id FROM supported_chains WHERE chain_name = 'ethereum'), 18
);