package transaction_service

import (
	system "deposit-collector/internal/system"

	uuid "github.com/google/uuid"
)

type DestinationDepositAddress = map[system.ChainPlatform]string
type WithdrawCollectorPrivateKeys = map[system.ChainPlatform]string

type baseDepositOperation struct {
	Amount         int64
	Address        string
	SequenceNumber uint32
	AccountID      uint32
}

type PendingDepositOperation struct {
	baseDepositOperation
	ID     uuid.UUID
	TxHash string
}

type GroupedPendingDepositOperation struct {
	baseDepositOperation
	OperationIDs []uuid.UUID
}

type PendingWithdrawalOperation struct {
	ID                 uuid.UUID
	Amount             int64
	DestinationAddress string
}
