package transaction_service

import (
	sql "database/sql"
	time "time"

	uuid "github.com/google/uuid"
)

type TransactionRepository struct {
	db *sql.DB
}

type StoredOperation struct {
	ExternalUserID uuid.UUID
	Amount         int64
	Type           string
	CreatedAt      time.Time
	TxHash         string
	Address        string
	Chain          string
	TokenAddress   string
	UnitName       string
	UnitSymbol     string
	TokenDecimals  int
}

func (r *TransactionRepository) GetOperationByTxHash(
	txHash string,
) (StoredOperation, error) {
	var operation StoredOperation
	q := `
SELECT o.amount, o.type, o.created_at, o.tx_hash,
u.external_id, ua.address, ua.chain,
ta.unit_name, ta.unit_symbol, ta.address AS token_address,
ta.decimals AS token_decimals
FROM operations AS o
JOIN users AS u ON operations.user_id = u.id
JOIN user_addresses AS ua ON operations.address_id = ua.id
JOIN token_addresses AS ta ON operations.token_address_id = ta.id
WHERE tx_hash = $1
`

	err := r.db.QueryRow(q, txHash).Scan(
		&operation.ExternalUserID,
		&operation.Amount,
		&operation.Type,
		&operation.CreatedAt,
		&operation.TxHash,
		&operation.Address,
		&operation.Chain,
		&operation.TokenAddress,
		&operation.UnitName,
		&operation.UnitSymbol,
		&operation.TokenDecimals,
	)

	if err != nil {
		return StoredOperation{}, err
	}

	return operation, nil
}

func (r *TransactionRepository) ExistsOperationByTxHash(
	txHash string,
) (bool, error) {
	var exists bool
	q := `
SELECT EXISTS(SELECT 1 FROM operations WHERE tx_hash = $1)
`
	err := r.db.QueryRow(q, txHash).Scan(&exists)

	if err != nil {
		return false, err
	}
	return exists, nil
}

/*
EndorseDepositOperation is the main function to endorse a deposit operation.
It will:
- Select the user ID, address ID and token address ID
- Insert the operation into the operations table
- Insert/update the balance into the user_balances table
- Insert/update the deposit operation into the pending_deposit_operations table
*/
func (r *TransactionRepository) EndorseDepositOperation(
	userID uuid.UUID,
	addressID uuid.UUID,
	tokenAddressID uuid.UUID,
	amount int64,
	txHash string,
) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	insertOperationQuery := `
INSERT INTO operations (
user_id, address_id, token_address_id, amount, type, tx_hash
)
VALUES ($1, $2, $3, $4, $5, $6)
`
	_, err = tx.Exec(
		insertOperationQuery,
		userID,
		addressID,
		tokenAddressID,
		amount,
		"deposit",
		txHash,
	)
	if err != nil {
		return err
	}

	insertBalanceQuery := `
INSERT INTO user_balances (user_id, token_address_id, available_balance)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, token_address_id) DO UPDATE SET
available_balance = user_balances.available_balance + EXCLUDED.available_balance,
updated_at = CURRENT_TIMESTAMP
`
	_, err = tx.Exec(
		insertBalanceQuery,
		userID,
		tokenAddressID,
		amount,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *TransactionRepository) GetUnprocessedDepositsByTokenAddressID(
	tokenAddressID uuid.UUID,
	limit int,
) ([]PendingDepositOperation, error) {
	q := `
SELECT o.id, o.amount, o.tx_hash, ua.address, ua.sequence_number, u.account_id
FROM operations o
JOIN user_addresses ua ON o.address_id = ua.id
JOIN users u ON o.user_id = u.id
WHERE o.token_address_id = $1 AND o.type = 'deposit' AND o.processed_at IS NULL
ORDER BY o.amount DESC
LIMIT $2
`

	rows, err := r.db.Query(q, tokenAddressID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	operations := make([]PendingDepositOperation, 0)
	for rows.Next() {
		var operation PendingDepositOperation
		err := rows.Scan(
			&operation.ID,
			&operation.Amount,
			&operation.TxHash,
			&operation.Address,
			&operation.SequenceNumber,
			&operation.AccountID,
		)
		if err != nil {
			return nil, err
		}
		operations = append(operations, operation)
	}

	return operations, nil
}

func (r *TransactionRepository) GetGroupedUnprocessedDepositsByTokenAddressID(
	tokenAddressID uuid.UUID,
	limit int,
) ([]GroupedPendingDepositOperation, error) {
	q := `SELECT
    u.account_id,
    ua.address,
	ua.sequence_number,
    SUM(o.amount) AS amount,
    ARRAY_AGG(o.id) AS operation_ids
FROM operations o
JOIN users u ON u.id = o.user_id
JOIN user_addresses ua ON ua.id = o.address_id
WHERE
    o.type = 'deposit'
    AND o.token_address_id = $1
    AND o.processed_at IS NULL
ORDER BY o.amount DESC
GROUP BY
    o.token_address_id
LIMIT $2;`

	rows, err := r.db.Query(q, tokenAddressID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	operations := make([]GroupedPendingDepositOperation, 0)
	for rows.Next() {
		var operation GroupedPendingDepositOperation
		err := rows.Scan(
			&operation.AccountID,
			&operation.Address,
			&operation.SequenceNumber,
			&operation.Amount,
			&operation.OperationIDs,
		)
		if err != nil {
			return nil, err
		}
		operations = append(operations, operation)
	}

	return operations, nil
}

func (r *TransactionRepository) MarkOperationAsProcessed(
	operationIDs []uuid.UUID,
	processedTxHash string,
) error {
	q := `UPDATE operations
SET
    processed_at = NOW(),
    processed_tx_hash = $2
WHERE id = ANY($1::uuid[]);`

	_, err := r.db.Exec(q, operationIDs, processedTxHash)
	if err != nil {
		return err
	}
	return nil
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}
