package queue

import (
	context "context"
	json "encoding/json"

	rabbitmq "deposit-collector/shared/rabbitmq"
)

type QueueName string
type OperationType string

const (
	OperationTypeCreateUser      OperationType = "create-user"
	OperationTypeGenerateAddress OperationType = "generate-address"
)

const (
	OperationsQueue QueueName = "operations-queue"
)

type Operation struct {
	message string
}

type OperationEvent struct {
	OperationType OperationType
	OperationData Operation
}

func PublishOperationEvent(
	ctx context.Context,
	queue *rabbitmq.Queue,
	event OperationEvent,
) error {
	message, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return queue.Publish(ctx, message)
}
