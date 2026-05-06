package system

import (
	sql "database/sql"
	errors "errors"
	fmt "fmt"
	strings "strings"

	metrics "deposit-collector/internal/metrics"
	observability "deposit-collector/pkg/observability"
	postgresql "deposit-collector/pkg/postgresql"

	uuid "github.com/google/uuid"
)

type SystemRepository struct {
	db            *sql.DB
	systemMetrics *metrics.SystemMetrics
}

func (r *SystemRepository) observeQueryMetrics(
	operation string,
	status metrics.QueryStatus,
	stopTimer observability.StopTimer,
) {
	if r.systemMetrics == nil {
		return
	}
	_ = r.systemMetrics.IncrementSystemDBQueryTotal(operation, string(status))
	_ = r.systemMetrics.ObserveSystemDBQueryDuration(operation, stopTimer())
}

func (r *SystemRepository) GetSupportedChains() ([]SupportedChain, error) {
	const operation = "get_supported_chains"
	status := metrics.QUERY_STATUS_FAILED
	stopTimer := observability.StartTimer()
	defer r.observeQueryMetrics(operation, status, stopTimer)

	var chains []SupportedChain

	q := `
SELECT id, chain_name, chain_platform, evm_chain_id
FROM supported_chains
`

	rows, err := r.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var chain SupportedChain
		err := rows.Scan(
			&chain.ChainDbID,
			&chain.ChainName,
			&chain.ChainPlatform,
			&chain.EVMChainID,
		)
		if err != nil {
			return nil, err
		}
		chains = append(chains, chain)
	}

	if chains == nil && rows.Err() == nil {
		status = metrics.QUERY_STATUS_SUCCESS
		return []SupportedChain{}, nil
	}

	status = metrics.QUERY_STATUS_SUCCESS
	return chains, nil
}

func (r *SystemRepository) AddNewSupportedChain(
	chain *NewSupportedChainRequest,
) error {
	const operation = "add_new_supported_chain"
	status := metrics.QUERY_STATUS_FAILED
	stopTimer := observability.StartTimer()
	defer r.observeQueryMetrics(operation, status, stopTimer)

	q := `
INSERT INTO supported_chains (
	chain_name, chain_platform, evm_chain_id
) VALUES ($1, $2, $3)
`

	_, err := r.db.Exec(
		q,
		strings.ToLower(chain.ChainName),
		strings.ToUpper(string(chain.ChainPlatform)),
		chain.EVMChainID,
	)
	if err == sql.ErrNoRows {
		status = metrics.QUERY_STATUS_SUCCESS
		return nil
	}
	if _, ok := postgresql.UniqueConstraintViolation(err); ok {
		return errors.New("chain already exists")
	}
	if err != nil {
		return err
	}

	status = metrics.QUERY_STATUS_SUCCESS
	return nil
}

func (r *SystemRepository) AddNewTokenAddress(
	tokenAddress *NewTokenAddressRequest,
) error {
	const operation = "add_new_token_address"
	status := metrics.QUERY_STATUS_FAILED
	stopTimer := observability.StartTimer()
	defer r.observeQueryMetrics(operation, status, stopTimer)

	q := `
INSERT INTO token_addresses (
	unit_name, unit_symbol, address, chain_id, decimals
) VALUES (
	$1, $2, $3, (SELECT id FROM supported_chains WHERE chain_name = $4), $5
)
`

	_, err := r.db.Exec(
		q,
		tokenAddress.UnitName,
		strings.ToUpper(tokenAddress.UnitSymbol),
		strings.ToLower(tokenAddress.Address),
		strings.ToLower(tokenAddress.ChainName),
		tokenAddress.Decimals,
	)
	if err == sql.ErrNoRows {
		status = metrics.QUERY_STATUS_SUCCESS
		return nil
	}
	if _, ok := postgresql.UniqueConstraintViolation(err); ok {
		return errors.New("token address already exists")
	}
	if err != nil {
		return err
	}

	status = metrics.QUERY_STATUS_SUCCESS
	return nil
}

func (r *SystemRepository) GetTokenAddresses(
	filters GetTokenAddressesRequest,
) ([]TokenAddress, error) {
	const operation = "get_token_addresses"
	status := metrics.QUERY_STATUS_FAILED
	stopTimer := observability.StartTimer()
	defer r.observeQueryMetrics(operation, status, stopTimer)

	var tokenAddresses []TokenAddress

	q := `
SELECT ta.id as token_address_db_id, ta.unit_name, ta.unit_symbol, ta.address,
ta.decimals, sc.id as chain_db_id, sc.chain_name, sc.chain_platform,
sc.evm_chain_id 
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
		addCond("sc.chain_name = $%d", *filters.Chain)
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

	if filters.Limit < 1 {
		return nil, errors.New("limit must be greater than 1")
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
			&tokenAddress.TokenAddressDbID,
			&tokenAddress.UnitName,
			&tokenAddress.UnitSymbol,
			&tokenAddress.Address,
			&tokenAddress.Decimals,
			&chain.ChainDbID,
			&chain.ChainName,
			&chain.ChainPlatform,
			&chain.EVMChainID,
		); err != nil {
			return nil, err
		}
		tokenAddress.Chain = chain
		tokenAddresses = append(tokenAddresses, tokenAddress)
	}

	if tokenAddresses == nil && rows.Err() == nil {
		status = metrics.QUERY_STATUS_SUCCESS
		return []TokenAddress{}, nil
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	status = metrics.QUERY_STATUS_SUCCESS
	return tokenAddresses, nil
}

func (r *SystemRepository) GetTokenAddressByID(
	id uuid.UUID,
) (TokenAddress, error) {
	const operation = "get_token_address_by_id"
	status := metrics.QUERY_STATUS_FAILED
	stopTimer := observability.StartTimer()
	defer r.observeQueryMetrics(operation, status, stopTimer)

	var tokenAddress TokenAddress
	var chain SupportedChain
	q := `
SELECT ta.unit_name, ta.unit_symbol, ta.address, ta.decimals,
sc.chain_name, sc.chain_platform, sc.evm_chain_id
FROM token_addresses as ta
INNER JOIN supported_chains as sc ON ta.chain_id = sc.id
WHERE ta.id = $1
`

	err := r.db.QueryRow(q, id).Scan(
		&tokenAddress.UnitName,
		&tokenAddress.UnitSymbol,
		&tokenAddress.Address,
		&tokenAddress.Decimals,
		&chain.ChainName,
		&chain.ChainPlatform,
		&chain.EVMChainID,
	)
	if err != nil {
		return TokenAddress{}, err
	}

	tokenAddress.Chain = chain

	status = metrics.QUERY_STATUS_SUCCESS
	return tokenAddress, nil
}
func NewSystemRepository(
	db *sql.DB,
	systemMetrics *metrics.SystemMetrics,
) *SystemRepository {
	return &SystemRepository{
		db:            db,
		systemMetrics: systemMetrics,
	}
}
