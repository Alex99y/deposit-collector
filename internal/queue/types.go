package queue

import (
	json "encoding/json"
	"errors"

	"github.com/google/uuid"
)

var ErrUnknownOperationType = errors.New("unknown operation type")

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
	TargetAddress     string
	DepositTxHash     string
}

type WithdrawOperationEvent struct {
	UserDbID         uuid.UUID
	TokenAddressDbId uuid.UUID
	TargetChainName  string
	TargetAddress    string
	WithdrawAmount   int64
}

type OperationEvent struct {
	RequestId     uuid.UUID
	OperationType OperationType
	OperationData []byte
}

func UnmarshalOperationData(operationEvent OperationEvent) (any, error) {
	switch operationEvent.OperationType {
	case OperationTypeDeposit:
		var depositOperation DepositOperationEvent
		err := json.Unmarshal(operationEvent.OperationData, &depositOperation)
		if err != nil {
			return nil, err
		}
		return depositOperation, nil
	case OperationTypeWithdraw:
		var withdrawOperation WithdrawOperationEvent
		err := json.Unmarshal(operationEvent.OperationData, &withdrawOperation)
		if err != nil {
			return nil, err
		}
		return withdrawOperation, nil
	}
	return nil, ErrUnknownOperationType
}
