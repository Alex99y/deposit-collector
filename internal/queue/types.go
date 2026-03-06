package queue

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
	TargetChainName string
	DepositTxHash   string
	TargetAddress   string
}

type WithdrawOperationEvent struct {
	SourceChainName string
	WithdrawTxHash  string
	SourceAddress   string
}

type OperationEvent struct {
	Id            string
	OperationType OperationType
	OperationData interface{}
}
