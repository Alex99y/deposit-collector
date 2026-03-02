package users

import (
	"context"
	sql "database/sql"
	time "time"

	postgresql "deposit-collector/pkg/postgresql"

	uuid "github.com/google/uuid"
)

const (
	ukAccountID  = "users_account_id_uk"
	ukExternalID = "users_external_id_uk"
)

type UsersRepository struct {
	ctx context.Context
	db  *sql.DB
}

func (r *UsersRepository) CreateUser(
	externalID string,
) error {

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

	return user, nil
}

func (r *UsersRepository) GetAddressesByExternalID(
	externalID string,
) ([]StoredAddress, error) {
	var addresses []StoredAddress

	q := `
SELECT ua.address, ua.sequence_number, ua.user_id, ua.chain, ua.created_at
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
			&address.UserID,
			&address.CreatedAt,
		)
		address.ExternalID = externalID
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, address)
	}

	return addresses, nil
}

func (r *UsersRepository) StoreAddress(
	request *StoreAddressRequest,
	getAddressFromSequenceNumber func(sequenceNumber int) (string, error),
) (*uuid.UUID, error) {
	tx, err := r.db.BeginTx(r.ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var userID uuid.UUID
	err = tx.QueryRowContext(
		r.ctx,
		"SELECT id FROM users WHERE external_id = $1 FOR UPDATE",
		request.ExternalID,
	).Scan(&userID)
	if err != nil {
		// User not found
		return nil, err
	}

	var sequenceNumber int
	querySequenceNumber := `
SELECT COALESCE(MAX(sequence_number), -1) + 1
FROM user_addresses
WHERE user_id = $1 AND chain = $2
	`
	err = tx.QueryRowContext(
		r.ctx, querySequenceNumber, userID, request.Chain,
	).Scan(&sequenceNumber)
	if err != nil {
		return nil, err
	}

	addressString, err := getAddressFromSequenceNumber(sequenceNumber)
	if err != nil {
		return nil, err
	}

	var addressID uuid.UUID
	insertAddressQuery := `
INSERT INTO user_addresses (address, sequence_number, user_id, chain)
VALUES ($1, $2, $3, $4)
RETURNING id
`
	err = tx.QueryRowContext(
		r.ctx, insertAddressQuery,
		addressString, sequenceNumber, userID, request.Chain,
	).Scan(&addressID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &addressID, nil
}

func NewUsersRepository(
	ctx context.Context,
	db *sql.DB,
) *UsersRepository {
	return &UsersRepository{ctx: ctx, db: db}
}
