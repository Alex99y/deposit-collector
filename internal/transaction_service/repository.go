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

func (r *TransactionRepository) GetDepositOperationByTxHash(
	txHash string,
) (StoredOperation, error) {
	const metricOperation = "get_deposit_operation_by_tx_hash"
	status := metrics.QUERY_STATUS_FAILED
	stopTimer := observability.StartTimer()
	defer func() {
		r.observeQueryMetrics(metricOperation, status, stopTimer)
	}()

	var operation StoredOperation
	q := `
SELECT user_operations.created_at, users.external_id,
deposit_operations.amount, deposit_operations.tx_hash,
user_addresses.address, user_addresses.chain,
token_addresses.unit_name, token_addresses.unit_symbol,
token_addresses.address AS token_address,
token_addresses.decimals AS token_decimals
FROM deposit_operations
JOIN users ON deposit_operations.user_id = users.id
JOIN user_addresses ON deposit_operations.address_id = user_addresses.id
JOIN user_operations ON deposit_operations.user_operation_id = user_operations.id
JOIN token_addresses ON deposit_operations.token_address_id = token_addresses.id
WHERE deposit_operations.tx_hash = $1
`

	err := r.db.QueryRow(q, txHash).Scan(
		&operation.CreatedAt,
		&operation.ExternalUserID,
		&operation.Amount,
		&operation.TxHash,
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

func (r *TransactionRepository) ExistsDepositOperationByTxHash(
	txHash string,
) (bool, error) {
	const metricOperation = "exists_deposit_operation_by_tx_hash"
	status := metrics.QUERY_STATUS_FAILED
	stopTimer := observability.StartTimer()
	defer func() {
		r.observeQueryMetrics(metricOperation, status, stopTimer)
	}()

	var exists bool
	q := `
SELECT EXISTS(SELECT 1 FROM deposit_operations WHERE tx_hash = $1)
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

	insertQuery := `
WITH inserted_user_operation AS (
INSERT INTO user_operations (user_id, type, status)
VALUES ($1, 'deposit', 'pending')
RETURNING id
)
INSERT INTO deposit_operations (
user_operation_id, token_address_id, amount, address_id, tx_hash
)
SELECT id, $2, $3, $4, $5 FROM inserted_user_operation
`

	result, err := tx.Exec(
		insertQuery,
		userID, tokenAddressID, amount, addressID, txHash,
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

	query := `
SELECT user_operations.id, deposit_operations.amount, deposit_operations.tx_hash,
user_addresses.address, user_addresses.sequence_number, users.account_id
FROM deposit_operations
JOIN user_operations ON user_operations.id = deposit_operations.user_operation_id
JOIN user_addresses ON user_addresses.id = deposit_operations.address_id
JOIN users ON users.id = user_operations.user_id
WHERE deposit_operations.token_address_id = $1 AND user_operations.status = 'pending'
ORDER BY deposit_operations.amount DESC
LIMIT $2;
`

	rows, err := r.db.Query(query, tokenAddressID, limit)
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

	query := `
SELECT users.account_id, user_addresses.address, user_addresses.sequence_number,
SUM(deposit_operations.amount) AS amount,
ARRAY_AGG(user_operations.id) AS operation_ids
FROM deposit_operations
JOIN user_operations ON user_operations.id = deposit_operations.user_operation_id
JOIN user_addresses ON user_addresses.id = deposit_operations.address_id
JOIN users ON users.id = user_operations.user_id
WHERE deposit_operations.token_address_id = $1 AND user_operations.status = 'pending'
GROUP BY users.account_id, user_addresses.address, user_addresses.sequence_number
ORDER BY amount DESC
LIMIT $2;
`

	rows, err := r.db.Query(query, tokenAddressID, limit)
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

func (r *TransactionRepository) MarkDepositOperationAsProcessed(
	depositUserOperationIDs []uuid.UUID,
) error {
	const metricOperation = "mark_deposit_operation_as_processed"
	status := metrics.QUERY_STATUS_FAILED
	stopTimer := observability.StartTimer()
	defer func() {
		r.observeQueryMetrics(metricOperation, status, stopTimer)
	}()

	q := `UPDATE user_operations
SET
    update_at = NOW(),
    status = 'processed'
WHERE id = ANY($1::uuid[]);`

	_, err := r.db.Exec(q, depositUserOperationIDs)
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

	insertQuery := `
WITH inserted_user_operation AS (
INSERT INTO user_operations (user_id, type, status)
VALUES ($1, 'withdraw', 'pending')
RETURNING id
)
INSERT INTO withdraw_operations (user_operation_id, token_address_id, amount, destination_address)
SELECT id, $2, $3, $4 FROM inserted_user_operation
`

	result, err := tx.Exec(
		insertQuery,
		userID,
		tokenAddressID,
		withdrawAmount,
		withdrawDestinationAddress,
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
SELECT user_operations.id, withdraw_operations.amount, withdraw_operations.destination_address
FROM withdraw_operations
JOIN user_operations ON user_operations.id = withdraw_operations.user_operation_id
WHERE withdraw_operations.token_address_id = $1 AND user_operations.status = 'pending'
ORDER BY withdraw_operations.amount ASC
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

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// @TODO: We should fist block the withdraw balance before marking the operation as processed.
	// This is to prevent duplicate withdrawals. Also, there should be another process to confirm the withdrawal.
	queryUpdateUserOperations := `UPDATE user_operations
SET
    update_at = NOW(),
    status = 'processed'
WHERE id = ANY($1::uuid[]);`

	_, err = tx.Exec(queryUpdateUserOperations, operationIDs)
	if err != nil {
		return err
	}

	queryUpdateWithdrawOperations := `UPDATE withdraw_operations
SET
	tx_hash = $1
FROM user_operations
WHERE user_operations.id = ANY($2::uuid[]);`

	_, err = tx.Exec(queryUpdateWithdrawOperations, processedTxHash, operationIDs)
	if err != nil {
		return err
	}

	queryUpdateUserBalances := `UPDATE user_balances
SET
	blocked_balance_for_withdrawal = blocked_balance_for_withdrawal - withdraw_operations.amount
FROM withdraw_operations
WHERE withdraw_operations.user_operation_id = ANY($1::uuid[]);`

	_, err = tx.Exec(queryUpdateUserBalances, operationIDs)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
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
