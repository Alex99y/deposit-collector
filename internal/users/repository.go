package users

import (
	context "context"
	sql "database/sql"
	errors "errors"
	fmt "fmt"
	time "time"

	metrics "deposit-collector/internal/metrics"
	observability "deposit-collector/pkg/observability"
	postgresql "deposit-collector/pkg/postgresql"

	uuid "github.com/google/uuid"
)

const (
	ukAccountID  = "users_account_id_uk"
	ukExternalID = "users_external_id_uk"
)

type UsersRepository struct {
	ctx               context.Context
	db                *sql.DB
	repositoryMetrics *metrics.RepositoryMetrics
}

func (r *UsersRepository) observeQueryMetrics(
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

func (r *UsersRepository) CreateUser(
	externalID string,
) error {
	const operation = "create_user"
	status := metrics.QUERY_STATUS_FAILED
	stopTimer := observability.StartTimer()
	defer r.observeQueryMetrics(operation, status, stopTimer)

	q := `
INSERT INTO users (external_id, account_id)
SELECT $1, COALESCE(MAX(account_id), 0) + 1
FROM users RETURNING id, account_id`

	// TODO: Hardcoded value for now. Should be configurable.
	const maxRetries = 5

	// Retry logic to handle unique violation on account_id.
	// We want account_id to be unique and a sequential index.
	// So probably we will have race conditions here and we need to retry.
	var lastErr error
	for i := 0; i <= maxRetries; i++ {
		var id uuid.UUID
		var accountID int64

		err := r.db.QueryRowContext(r.ctx, q, externalID).Scan(&id, &accountID)
		if err == nil {
			status = metrics.QUERY_STATUS_SUCCESS
			return nil
		}

		if c, ok := postgresql.UniqueConstraintViolation(err); ok {
			switch c {
			case ukAccountID:
				// Unique violation on account_id, retry
				time.Sleep(time.Duration(5*(i+1)) * time.Millisecond)
				continue
			case ukExternalID:
				// The user already exists, return nil
				status = metrics.QUERY_STATUS_SUCCESS
				return nil
			default:
				// Unique violation on other column, return error
				return err
			}
		}

		return err
	}

	return lastErr
}

func (r *UsersRepository) GetUserByExternalID(
	externalID string,
) (StoredUser, error) {
	const operation = "get_user_by_external_id"
	status := metrics.QUERY_STATUS_FAILED
	stopTimer := observability.StartTimer()
	defer r.observeQueryMetrics(operation, status, stopTimer)

	var user StoredUser

	q := `
SELECT id, external_id, account_id, created_at, updated_at \
FROM users
WHERE external_id = $1
`

	err := r.db.QueryRowContext(r.ctx, q, externalID).Scan(
		&user.ID,
		&user.ExternalID,
		&user.AccountID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return StoredUser{}, err
	}

	status = metrics.QUERY_STATUS_SUCCESS
	return user, nil
}

func (r *UsersRepository) GetAddressesByExternalID(
	externalID string,
) ([]StoredAddress, error) {
	const operation = "get_addresses_by_external_id"
	status := metrics.QUERY_STATUS_FAILED
	stopTimer := observability.StartTimer()
	defer r.observeQueryMetrics(operation, status, stopTimer)

	var addresses []StoredAddress

	q := `
SELECT ua.address, ua.sequence_number, ua.chain, ua.created_at
FROM user_addresses ua
INNER JOIN users u ON ua.user_id = u.id
WHERE u.external_id = $1`

	rows, err := r.db.QueryContext(r.ctx, q, externalID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var address StoredAddress
		err := rows.Scan(
			&address.Address,
			&address.SequenceNumber,
			&address.Chain,
			&address.CreatedAt,
		)

		if err != nil {
			return nil, err
		}
		addresses = append(addresses, address)
	}

	if addresses == nil && rows.Err() == nil {
		status = metrics.QUERY_STATUS_SUCCESS
		return []StoredAddress{}, nil
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	status = metrics.QUERY_STATUS_SUCCESS
	return addresses, nil
}

func (r *UsersRepository) StoreAddress(
	request *StoreAddressRequest,
	getAddressFromSequenceNumber func(
		userAccountID uint32,
		sequenceNumber uint32,
	) (string, error),
) (string, error) {
	const operation = "store_address"
	status := metrics.QUERY_STATUS_FAILED
	stopTimer := observability.StartTimer()
	defer r.observeQueryMetrics(operation, status, stopTimer)

	tx, err := r.db.BeginTx(r.ctx, nil)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var userID uuid.UUID
	var userAccountID uint32
	err = tx.QueryRowContext(
		r.ctx,
		"SELECT id, account_id FROM users WHERE external_id = $1 FOR UPDATE",
		request.ExternalID,
	).Scan(&userID, &userAccountID)
	if err != nil {
		// User not found
		return "", errors.New("user not found: " + request.ExternalID)
	}

	var sequenceNumber uint32
	querySequenceNumber := `
SELECT COALESCE(MAX(sequence_number), -1) + 1
FROM user_addresses
WHERE user_id = $1 AND chain = $2
	`
	err = tx.QueryRowContext(
		r.ctx, querySequenceNumber, userID, request.Chain,
	).Scan(&sequenceNumber)
	if err != nil {
		return "", errors.New("error getting sequence number: " + err.Error())
	}

	addressString, err := getAddressFromSequenceNumber(
		userAccountID, sequenceNumber,
	)
	if err != nil {
		return "", errors.New(
			"error getting address from sequence number: " + err.Error(),
		)
	}

	var addressID uuid.UUID
	insertAddressQuery := `
INSERT INTO user_addresses (address, sequence_number, user_id, chain)
VALUES ($1, $2, $3, $4)
RETURNING id
`
	fmt.Printf("Address %s\n", addressString)
	err = tx.QueryRowContext(
		r.ctx, insertAddressQuery,
		addressString, sequenceNumber, userID, request.Chain,
	).Scan(&addressID)

	if err != nil {
		return "", errors.New("error inserting address: " + err.Error())
	}

	if err := tx.Commit(); err != nil {
		return "", errors.New("error committing transaction: " + err.Error())
	}
	status = metrics.QUERY_STATUS_SUCCESS
	return addressString, nil
}

func (r *UsersRepository) FindUserIDAndAddressIDByAddress(
	address string,
) (uuid.UUID, uuid.UUID, error) {
	const operation = "find_user_id_and_address_id_by_address"
	status := metrics.QUERY_STATUS_FAILED
	stopTimer := observability.StartTimer()
	defer r.observeQueryMetrics(operation, status, stopTimer)

	var userDbId uuid.UUID
	var addressDbId uuid.UUID
	err := r.db.QueryRowContext(
		r.ctx,
		"SELECT id, user_id FROM user_addresses WHERE address = $1",
		address,
	).Scan(&addressDbId, &userDbId)
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, err
	}
	status = metrics.QUERY_STATUS_SUCCESS
	return userDbId, addressDbId, nil
}

func NewUsersRepository(
	ctx context.Context,
	db *sql.DB,
	repositoryMetrics *metrics.RepositoryMetrics,
) *UsersRepository {
	return &UsersRepository{
		ctx:               ctx,
		db:                db,
		repositoryMetrics: repositoryMetrics,
	}
}
