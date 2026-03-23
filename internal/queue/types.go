package queue

import "github.com/google/uuid"

type QueueName string
type OperationType string

const (
	OperationTypeWithdraw OperationType = "withdraw-operation"
	OperationTypeDeposit  OperationType = "deposit-operation"
)

const (
	OperationsQueue QueueName = "operations-queue"
)

type DepositOperationEvent struct {
	UserDbID          uuid.UUID
	TargetAddressDbId uuid.UUID
	TargetChainName   string
	DepositTxHash     string
}

type WithdrawOperationEvent struct {
	UserDbID          uuid.UUID
	SourceChainName   string
	SourceAddressDbId uuid.UUID
}

type OperationEvent struct {
	RequestId     uuid.UUID
	OperationType OperationType
	OperationData interface{}
}
