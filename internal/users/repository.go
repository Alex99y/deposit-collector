package users

import (
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
	db *sql.DB
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

		err := r.db.QueryRow(q, externalID).Scan(&id, &accountID)
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

	err := r.db.QueryRow(q, externalID).Scan(
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
SELECT a.id, a.address, a.sequence_number, a.user_id, a.chain, a.created_at
FROM addresses a
INNER JOIN users u ON a.user_id = u.id
WHERE u.external_id = $1`

	rows, err := r.db.Query(q, externalID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var address StoredAddress
		err := rows.Scan(
			&address.ID,
			&address.Address,
			&address.SequenceNumber,
			&address.UserID,
			&address.Chain,
			&address.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, address)
	}

	return addresses, nil
}

func (r *UsersRepository) CreateAddress(
	address *CreateAddressRequest,
) (*uuid.UUID, error) {

	q := `
INSERT INTO addresses (address, sequence_number, user_id, chain)
VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING RETURNING id
`

	var id uuid.UUID

	err := r.db.QueryRow(
		q, address.Address, address.SequenceNumber, address.UserID, address.Chain,
	).Scan(&id)

	// If the address already exists, return nil
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &id, nil
}

func NewUsersRepository(db *sql.DB) *UsersRepository {
	return &UsersRepository{db: db}
}
