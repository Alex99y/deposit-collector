package transaction_service

import (
	sql "database/sql"
	errors "errors"
	time "time"

	metrics "deposit-collector/internal/metrics"
	observability "deposit-collector/pkg/observability"

	uuid "github.com/google/uuid"
)

type TransactionRepository struct {
	db                *sql.DB
	repositoryMetrics *metrics.RepositoryMetrics
}

type StoredOperation struct {
	ExternalUserID string
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

func (r *TransactionRepository) observeQueryMetrics(
	operation string,
	status metrics.QueryStatus,
	stopTimer observability.StopTimer,
) {
	if r.repositoryMetrics == nil {
		return
	}
	_ = r.repositoryMetrics.IncrementDBQueryTotal(operation, string(status))
	_ = r.repositoryMetrics.ObserveDBQueryDuration(operation, stopTimer())
}

func (r *TransactionRepository) GetOperationByTxHash(
	txHash string,
) (StoredOperation, error) {
	const metricOperation = "get_operation_by_tx_hash"
	status := metrics.QUERY_STATUS_FAILED
	stopTimer := observability.StartTimer()
	defer func() {
		r.observeQueryMetrics(metricOperation, status, stopTimer)
	}()

	var operation StoredOperation
	q := `
SELECT o.amount, o.type, o.created_at, o.tx_hash,
u.external_id, ua.address, ua.chain,
ta.unit_name, ta.unit_symbol, ta.address AS token_address,
ta.decimals AS token_decimals
FROM operations AS o
JOIN users AS u ON o.user_id = u.id
JOIN user_addresses AS ua ON o.deposit_address_id = ua.id
JOIN token_addresses AS ta ON o.token_address_id = ta.id
WHERE o.tx_hash = $1
`

	err := r.db.QueryRow(q, txHash).Scan(
		&operation.Amount,
		&operation.Type,
		&operation.CreatedAt,
		&operation.TxHash,
		&operation.ExternalUserID,
		&operation.Address,
		&operation.Chain,
		&operation.UnitName,
		&operation.UnitSymbol,
		&operation.TokenAddress,
		&operation.TokenDecimals,
	)

	if err != nil {
		return StoredOperation{}, err
	}

	status = metrics.QUERY_STATUS_SUCCESS
	return operation, nil
}

func (r *TransactionRepository) ExistsOperationByTxHash(
	txHash string,
) (bool, error) {
	const metricOperation = "exists_operation_by_tx_hash"
	status := metrics.QUERY_STATUS_FAILED
	stopTimer := observability.StartTimer()
	defer func() {
		r.observeQueryMetrics(metricOperation, status, stopTimer)
	}()

	var exists bool
	q := `
SELECT EXISTS(SELECT 1 FROM operations WHERE tx_hash = $1)
`
	err := r.db.QueryRow(q, txHash).Scan(&exists)

	if err != nil {
		return false, err
	}
	status = metrics.QUERY_STATUS_SUCCESS
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
	const metricOperation = "endorse_deposit_operation"
	status := metrics.QUERY_STATUS_FAILED
	stopTimer := observability.StartTimer()
	defer func() {
		r.observeQueryMetrics(metricOperation, status, stopTimer)
	}()

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	insertOperationQuery := `
INSERT INTO operations (
user_id, deposit_address_id, token_address_id, amount, type, tx_hash
)
VALUES ($1, $2, $3, $4, $5, $6)
`
	result, err := tx.Exec(
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

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("no rows affected for insert operation")
	}

	insertBalanceQuery := `
INSERT INTO user_balances (user_id, token_address_id, available_balance)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, token_address_id) DO UPDATE SET
available_balance = user_balances.available_balance + EXCLUDED.available_balance,
updated_at = CURRENT_TIMESTAMP
`
	result, err = tx.Exec(
		insertBalanceQuery,
		userID,
		tokenAddressID,
		amount,
	)
	if err != nil {
		return err
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("no rows affected for insert balance")
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	status = metrics.QUERY_STATUS_SUCCESS
	return nil
}

func (r *TransactionRepository) GetUnprocessedDepositsByTokenAddressID(
	tokenAddressID uuid.UUID,
	limit int,
) ([]PendingDepositOperation, error) {
	const metricOperation = "get_unprocessed_deposits_by_token_address_id"
	status := metrics.QUERY_STATUS_FAILED
	stopTimer := observability.StartTimer()
	defer func() {
		r.observeQueryMetrics(metricOperation, status, stopTimer)
	}()

	q := `
SELECT o.id, o.amount, o.tx_hash, ua.address, ua.sequence_number, u.account_id
FROM operations o
JOIN user_addresses ua ON o.deposit_address_id = ua.id
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

	if err := rows.Err(); err != nil {
		return nil, err
	}
	status = metrics.QUERY_STATUS_SUCCESS
	return operations, nil
}

func (r *TransactionRepository) GetGroupedUnprocessedDepositsByTokenAddressID(
	tokenAddressID uuid.UUID,
	limit int,
) ([]GroupedPendingDepositOperation, error) {
	const metricOperation = "get_grouped_unprocessed_deposits_by_token_address_id"
	status := metrics.QUERY_STATUS_FAILED
	stopTimer := observability.StartTimer()
	defer func() {
		r.observeQueryMetrics(metricOperation, status, stopTimer)
	}()

	q := `SELECT
    u.account_id,
    ua.address,
	ua.sequence_number,
    SUM(o.amount) AS amount,
    ARRAY_AGG(o.id) AS operation_ids
FROM operations o
JOIN users u ON u.id = o.user_id
JOIN user_addresses ua ON ua.id = o.deposit_address_id
WHERE
    o.type = 'deposit'
    AND o.token_address_id = $1
    AND o.processed_at IS NULL
GROUP BY
    u.account_id,
    ua.address,
	ua.sequence_number
ORDER BY amount DESC
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

	if err := rows.Err(); err != nil {
		return nil, err
	}
	status = metrics.QUERY_STATUS_SUCCESS
	return operations, nil
}

func (r *TransactionRepository) MarkOperationAsProcessed(
	operationIDs []uuid.UUID,
	processedTxHash string,
) error {
	const metricOperation = "mark_operation_as_processed"
	status := metrics.QUERY_STATUS_FAILED
	stopTimer := observability.StartTimer()
	defer func() {
		r.observeQueryMetrics(metricOperation, status, stopTimer)
	}()

	q := `UPDATE operations
SET
    processed_at = NOW(),
    processed_tx_hash = $2
WHERE id = ANY($1::uuid[]);`

	_, err := r.db.Exec(q, operationIDs, processedTxHash)
	if err != nil {
		return err
	}
	status = metrics.QUERY_STATUS_SUCCESS
	return nil
}

func (r *TransactionRepository) EndorseWithdrawOperation(
	userID uuid.UUID,
	tokenAddressID uuid.UUID,
	withdrawAmount int64,
	withdrawDestinationAddress string,
) error {
	const metricOperation = "endorse_withdraw_operation"
	status := metrics.QUERY_STATUS_FAILED
	stopTimer := observability.StartTimer()
	defer func() {
		r.observeQueryMetrics(metricOperation, status, stopTimer)
	}()

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var userBalance int64
	query := `
SELECT available_balance
FROM user_balances
WHERE user_id = $1 AND token_address_id = $2 AND available_balance >= $3
FOR UPDATE
`
	err = tx.QueryRow(
		query, userID, tokenAddressID, withdrawAmount,
	).Scan(&userBalance)
	if err != nil {
		return err
	}

	if userBalance < withdrawAmount {
		return errors.New("insufficient balance for withdrawal")
	}

	insertOperationQuery := `
INSERT INTO operations (
user_id, withdraw_destination_address, token_address_id, amount, type
)
VALUES ($1, $2, $3, $4, 'withdraw')
`

	result, err := tx.Exec(
		insertOperationQuery,
		userID,
		withdrawDestinationAddress,
		tokenAddressID,
		withdrawAmount,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("no rows affected for insert operation")
	}

	updateWithdrawUserBalanceQuery := `
UPDATE user_balances
SET
available_balance = available_balance - $1,
blocked_balance_for_withdrawal = blocked_balance_for_withdrawal + $1,
updated_at = CURRENT_TIMESTAMP
WHERE user_id = $2 AND token_address_id = $3
`

	result, err = tx.Exec(
		updateWithdrawUserBalanceQuery,
		withdrawAmount,
		userID,
		tokenAddressID,
	)
	if err != nil {
		return err
	}
	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("no rows affected for update user balance")
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	status = metrics.QUERY_STATUS_SUCCESS

	return nil
}

func (r *TransactionRepository) GetUnprocessedWithdrawalsByTokenAddressID(
	tokenAddressID uuid.UUID,
	limit int,
) ([]PendingWithdrawalOperation, error) {
	const metricOperation = "get_unprocessed_withdrawals_by_token_address_id"
	status := metrics.QUERY_STATUS_FAILED
	stopTimer := observability.StartTimer()
	defer func() {
		r.observeQueryMetrics(metricOperation, status, stopTimer)
	}()

	q := `
SELECT o.id, o.amount, o.withdraw_destination_address
FROM operations o
WHERE o.token_address_id = $1 AND o.type = 'withdraw' AND o.processed_at IS NULL
ORDER BY o.created_at ASC
LIMIT $2
`

	rows, err := r.db.Query(q, tokenAddressID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	operations := make([]PendingWithdrawalOperation, 0)
	for rows.Next() {
		var operation PendingWithdrawalOperation
		err := rows.Scan(
			&operation.ID,
			&operation.Amount,
			&operation.DestinationAddress,
		)
		if err != nil {
			return nil, err
		}
		operations = append(operations, operation)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	status = metrics.QUERY_STATUS_SUCCESS
	return operations, nil
}

func (r *TransactionRepository) MarkWithdrawalOperationAsProcessed(
	operationIDs []uuid.UUID,
	processedTxHash string,
) error {
	// TODO: Accept map[uuid.UUID]string (operationId -> txHash) when batching
	// multiple withdrawals into one chain transaction (see docs/todo.md).
	const metricOperation = "mark_withdrawal_operation_as_processed"
	status := metrics.QUERY_STATUS_FAILED
	stopTimer := observability.StartTimer()
	defer func() {
		r.observeQueryMetrics(metricOperation, status, stopTimer)
	}()

	// @TODO: We should fist block the withdraw balance before marking the operation as processed.
	// This is to prevent duplicate withdrawals. Also, there should be another process to confirm the withdrawal.
	q := `UPDATE operations
SET
    processed_at = NOW(),
    processed_tx_hash = $2
WHERE id = ANY($1::uuid[]);`

	_, err := r.db.Exec(q, operationIDs, processedTxHash)
	if err != nil {
		return err
	}
	status = metrics.QUERY_STATUS_SUCCESS
	return nil
}

func NewTransactionRepository(
	db *sql.DB,
	repositoryMetrics *metrics.RepositoryMetrics,
) *TransactionRepository {
	return &TransactionRepository{
		db:                db,
		repositoryMetrics: repositoryMetrics,
	}
}
