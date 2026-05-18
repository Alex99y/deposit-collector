package queue

import (
	context "context"
	json "encoding/json"
	fmt "fmt"

	logger "deposit-collector/pkg/logger"
	rabbitmq "deposit-collector/pkg/rabbitmq"
)

type OperationConsumerArgs struct {
	rabbitmq.ConsumeArgs
	OperationEvent OperationEvent
}

func (a *OperationConsumerArgs) OperationData() (any, error) {
	return UnmarshalOperationData(a.OperationEvent)
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
		var operationEvent OperationEvent
		err := json.Unmarshal(args.RawMessage(), &operationEvent)
		if err != nil {
			q.logger.Error(fmt.Sprintf("error unmarshalling operation event: %v", err))
			_ = args.Reject()
			return
		}
		callback(&OperationConsumerArgs{
			ConsumeArgs:    *args,
			OperationEvent: operationEvent,
		})
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
