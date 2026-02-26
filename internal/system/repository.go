package system

import (
	sql "database/sql"
	errors "errors"
	fmt "fmt"
	strings "strings"

	uuid "github.com/google/uuid"
)

type SystemRepository struct {
	db *sql.DB
}

func (r *SystemRepository) GetSupportedChains() ([]SupportedChain, error) {
	var chains []SupportedChain

	q := "SELECT id, network, chain_platform, bip44_id FROM supported_chains"

	rows, err := r.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var chain SupportedChain
		err := rows.Scan(
			&chain.ID,
			&chain.Network,
			&chain.ChainPlatform,
			&chain.BIP44ID,
		)
		if err != nil {
			return nil, err
		}
		chains = append(chains, chain)
	}

	return chains, nil
}

func (r *SystemRepository) AddNewSupportedChain(
	chain *NewSupportedChainRequest,
) error {
	q := `
INSERT INTO supported_chains (network, chain_platform, bip44_id)
VALUES ($1, $2, $3) ON CONFLICT DO NOTHING
`

	_, err := r.db.Exec(q, chain.Network, chain.ChainPlatform, chain.BIP44ID)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}

	return nil
}

func (r *SystemRepository) AddNewTokenAddress(
	tokenAddress *NewTokenAddressRequest,
) error {
	q := `
INSERT INTO token_addresses (
	unit_name, unit_symbol, address, chain_id, decimals
) VALUES (
	$1, $2, $3, (SELECT id FROM supported_chains WHERE network = $4), $5
) ON CONFLICT DO NOTHING
`

	_, err := r.db.Exec(
		q,
		tokenAddress.UnitName,
		tokenAddress.UnitSymbol,
		tokenAddress.Address,
		tokenAddress.Chain,
		tokenAddress.Decimals,
	)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}

	return nil
}

func (r *SystemRepository) GetTokenAddresses(
	filters GetTokenAddressesRequest,
) ([]TokenAddress, error) {
	var tokenAddresses []TokenAddress

	q := `
SELECT ta.id, ta.unit_name, ta.unit_symbol, ta.address, ta.decimals,
sc.id, sc.network, sc.chain_platform, sc.bip44_id
FROM token_addresses as ta
INNER JOIN supported_chains as sc ON ta.chain_id = sc.id
`
	where := []string{}
	args := []any{}

	addCond := func(cond string, v any) {
		args = append(args, v)
		where = append(where, fmt.Sprintf(cond, len(args))) // %d -> $N
	}

	if filters.Chain != nil {
		addCond("sc.network = $%d", *filters.Chain)
	}
	if filters.Address != nil {
		addCond("ta.address = $%d", *filters.Address)
	}
	if filters.UnitSymbol != nil {
		addCond("ta.unit_symbol = $%d", *filters.UnitSymbol)
	}

	if len(where) > 0 {
		q += "WHERE " + strings.Join(where, " AND ") + "\n"
	}

	if filters.Limit > 100 || filters.Limit < 1 {
		return nil, errors.New("limit must be between 1 and 100")
	} else {
		args = append(args, filters.Limit)
		q += fmt.Sprintf("LIMIT $%d\n", len(args))
	}

	if filters.Offset < 0 {
		return nil, errors.New("offset must be greater than 0")
	} else {
		args = append(args, filters.Offset)
		q += fmt.Sprintf("OFFSET $%d\n", len(args))
	}

	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tokenAddress TokenAddress
		var chain SupportedChain
		if err := rows.Scan(
			&tokenAddress.ID,
			&tokenAddress.UnitName,
			&tokenAddress.UnitSymbol,
			&tokenAddress.Address,
			&tokenAddress.Decimals,
			&chain.ID,
			&chain.Network,
			&chain.ChainPlatform,
			&chain.BIP44ID,
		); err != nil {
			return nil, err
		}
		tokenAddress.Chain = chain
		tokenAddresses = append(tokenAddresses, tokenAddress)
	}
	return tokenAddresses, rows.Err()
}

func (r *SystemRepository) GetTokenAddressByID(
	id uuid.UUID,
) (TokenAddress, error) {
	var tokenAddress TokenAddress
	var chain SupportedChain
	q := `
SELECT ta.id, ta.unit_name, ta.unit_symbol, ta.address, ta.decimals,
sc.id, sc.network, sc.chain_platform, sc.bip44_id
FROM token_addresses as ta
INNER JOIN supported_chains as sc ON ta.chain_id = sc.id
WHERE ta.id = $1
`

	err := r.db.QueryRow(q, id).Scan(
		&tokenAddress.ID,
		&tokenAddress.UnitName,
		&tokenAddress.UnitSymbol,
		&tokenAddress.Address,
		&tokenAddress.Decimals,
		&chain.ID,
		&chain.Network,
		&chain.ChainPlatform,
		&chain.BIP44ID,
	)
	if err != nil {
		return TokenAddress{}, err
	}

	tokenAddress.Chain = chain

	return tokenAddress, nil
}
func NewSystemRepository(db *sql.DB) *SystemRepository {
	return &SystemRepository{db: db}
}
