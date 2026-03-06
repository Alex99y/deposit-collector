package queue

import (
	context "context"
	json "encoding/json"

	logger "deposit-collector/pkg/logger"
	rabbitmq "deposit-collector/pkg/rabbitmq"
)

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
	mandatory bool,
	immediate bool,
) error {
	message, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return q.queue.Publish(ctx, message, mandatory, immediate)
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

func (q *OperationQueue) Close() error {
	return q.queue.Close()
}

func NewOperationsQueue(
	queue *rabbitmq.Queue,
	logger *logger.Logger,
) *OperationQueue {
	return &OperationQueue{queue: queue, logger: logger}
}
