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

type StoreAddressRequest struct {
	ExternalID     string
	Chain          system.ChainPlatform
	SequenceNumber int
}

type StoredAddress struct {
	StoreAddressRequest
	UserID    uuid.UUID
	Address   string
	CreatedAt time.Time
}

type CreateAddressRequest struct {
	UserID         uuid.UUID
	Address        string
	SequenceNumber int
	Chain          system.ChainPlatform
}
