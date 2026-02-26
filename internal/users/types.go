package users

import (
	"time"

	system "deposit-collector/internal/system"

	uuid "github.com/google/uuid"
)

type StoredUser struct {
	ID         uuid.UUID
	ExternalID string
	AccountID  int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type StoredAddress struct {
	ID             uuid.UUID
	Address        string
	SequenceNumber int
	UserID         uuid.UUID
	Chain          system.ChainPlatform
	CreatedAt      time.Time
}

type CreateAddressRequest struct {
	UserID         uuid.UUID
	Address        string
	SequenceNumber int
	Chain          system.ChainPlatform
}
