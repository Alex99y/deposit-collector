package queue

import (
	context "context"
	json "encoding/json"

	logger "deposit-collector/pkg/logger"
	rabbitmq "deposit-collector/pkg/rabbitmq"
	utils "deposit-collector/pkg/utils"
)

type QueueName string
type OperationType string

const (
	OperationTypeWithdraw OperationType = "withdraw-operation"
	OperationTypeDeposit  OperationType = "deposit-operation"
)

const (
	OperationsQueue QueueName = "operations-queue"
)

type Operation struct {
	Message string
}

type OperationEvent struct {
	OperationType OperationType
	OperationData Operation
}

type OperationConsumerArgs struct {
	rabbitmq.ConsumeArgs
}

func (a *OperationConsumerArgs) Message() OperationEvent {
	rawMessage := a.RawMessage()
	var operationEvent OperationEvent
	err := json.Unmarshal(rawMessage, &operationEvent)
	if err != nil {
		return OperationEvent{}
	}
	return operationEvent
}

type OperationQueue struct {
	queue  *rabbitmq.Queue
	logger *logger.Logger
}

func (q *OperationQueue) PublishOperationEvent(
	ctx context.Context,
	event OperationEvent,
) error {
	message, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return q.queue.Publish(ctx, message)
}

type ConsumeCallback func(*OperationConsumerArgs)

func (q *OperationQueue) Consume(
	ctx context.Context,
	callback ConsumeCallback,
) error {
	return q.queue.Consume(ctx, func(args *rabbitmq.ConsumeArgs) {
		operationConsumerArgs := &OperationConsumerArgs{
			ConsumeArgs: *args,
		}
		callback(operationConsumerArgs)
	})
}

func NewOperationsQueue(
	rmq *rabbitmq.RabbitMQ,
	logger *logger.Logger,
) *OperationQueue {
	operationsQueue, err := rabbitmq.GetQueue(
		rmq, string(OperationsQueue), logger,
	)
	if err != nil {
		utils.FailOnError(logger, err, "Error creating operations queue")
	}
	return &OperationQueue{queue: operationsQueue, logger: logger}
}
