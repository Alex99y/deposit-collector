package worker

import (
	context "context"
	"fmt"

	queue "deposit-collector/internal/queue"
	logger "deposit-collector/pkg/logger"
	rabbitmq "deposit-collector/pkg/rabbitmq"
	utils "deposit-collector/pkg/utils"
)

type Worker struct {
	logger          *logger.Logger
	operationsQueue *queue.OperationQueue
	id              int
}

func (w *Worker) Start(ctx context.Context) {
	w.logger.Info(fmt.Sprintf("worker %d starting", w.id))
	go func() {
		err := w.run(ctx)
		if err != nil {
			w.logger.Error(fmt.Sprintf("error running worker %d: %v", w.id, err))
		}
	}()
}

func (w *Worker) Stop(ctx context.Context) error {
	w.logger.Info(fmt.Sprintf("worker %d stopping", w.id))
	return w.operationsQueue.Close()
}

func (w *Worker) run(ctx context.Context) error {
	err := w.operationsQueue.Consume(ctx, func(args *queue.OperationConsumerArgs) {
		operation := args.Message()
		w.logger.Info(fmt.Sprintf("Received operation: %+v", operation))
		err := args.Ack()
		if err != nil {
			w.logger.Error(fmt.Sprintf("Error acknowledging operation: %v", err))
		}
	})
	return err
}

func NewWorker(
	rmq *rabbitmq.RabbitMQClient,
	id int,
	logger *logger.Logger,
) *Worker {
	qq, err := rabbitmq.GetQueue(rmq, rabbitmq.ChannelArgs{
		PrefetchCount: 1,
		PrefetchSize:  0,
	}, rabbitmq.QueueArgs{
		Name:       string(queue.OperationsQueue),
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
	}, logger)
	if err != nil {
		utils.FailOnError(logger, err, "Error creating operations queue")
	}
	operationsQueue := queue.NewOperationsQueue(qq, logger)
	return &Worker{logger: logger, operationsQueue: operationsQueue, id: id}
}
