package users

import (
	time "time"

	system "deposit-collector/internal/system"

	uuid "github.com/google/uuid"
)

type StoredUser struct {
	ID         uuid.UUID `json:"userDbId"`
	ExternalID string    `json:"externalUserId"`
	AccountID  int       `json:"accountId"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type StoreAddressRequest struct {
	ExternalID     string
	Chain          system.ChainPlatform
	SequenceNumber int
}

type StoredAddress struct {
	Chain          system.ChainPlatform `json:"chainPlatform"`
	SequenceNumber int                  `json:"sequenceNumber"`
	Address        string               `json:"address"`
	CreatedAt      time.Time            `json:"createdAt"`
}

type CreateAddressRequest struct {
	UserID         uuid.UUID
	Address        string
	SequenceNumber int
	Chain          system.ChainPlatform
}

type StoredUserBalance struct {
	AvailableBalance            int64     `json:"availableBalance"`
	FrozenBalance               int64     `json:"frozenBalance"`
	BlockedBalanceForWithdrawal int64     `json:"blockedBalanceForWithdrawal"`
	UpdatedAt                   time.Time `json:"updatedAt"`
	UnitSymbol                  string    `json:"unitSymbol"`
	UnitName                    string    `json:"unitName"`
	TokenAddress                string    `json:"tokenAddress"`
}
