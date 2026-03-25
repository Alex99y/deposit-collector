package worker

import (
	context "context"
	fmt "fmt"

	queue "deposit-collector/internal/queue"
	transaction_service "deposit-collector/internal/transaction_service"
	logger "deposit-collector/pkg/logger"
	rabbitmq "deposit-collector/pkg/rabbitmq"
	utils "deposit-collector/pkg/utils"
)

type Worker struct {
	logger             *logger.Logger
	transactionService *transaction_service.TransactionService
	operationsQueue    *queue.OperationQueue
	id                 int
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
		operation := args.OperationData()
		switch parsedOperation := operation.(type) {
		case queue.DepositOperationEvent:
			w.logger.Info(
				"Received deposit operation id: " +
					args.OperationEvent.RequestId.String(),
			)
			customError, err := w.transactionService.ValidateAndStoreDepositOperation(
				&parsedOperation,
			)
			if customError != nil {
				if customError.IsRetryable() {
					w.logger.Error(
						fmt.Sprintf("Error validating and storing deposit operation: %v",
							customError.ErrorMessage()),
					)
					_ = args.Reject()
					return
				}
			}
			if err != nil {
				w.logger.Error(
					fmt.Sprintf(
						"Error validating and storing deposit operation: %v",
						err,
					),
				)
				_ = args.Nack()
				return
			}
			w.logger.Info(
				fmt.Sprintf(
					"Deposit operation validated and stored: %+v",
					parsedOperation.DepositTxHash,
				),
			)
			_ = args.Ack()
			return
		case queue.WithdrawOperationEvent:
			w.logger.Info(
				"Received withdraw operation id: " +
					args.OperationEvent.RequestId.String(),
			)
			_ = args.Ack()
			return
		default:
			w.logger.Error(fmt.Sprintf("Unknown operation type: %T", parsedOperation))
			_ = args.Nack()
			return
		}
	})
	return err
}

func NewWorker(
	rmq *rabbitmq.RabbitMQClient,
	transactionService *transaction_service.TransactionService,
	id int,
	logger *logger.Logger,
) *Worker {
	if rmq == nil || transactionService == nil || logger == nil {
		panic("Invalid worker dependencies")
	}
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
	return &Worker{
		logger:             logger,
		transactionService: transactionService,
		operationsQueue:    operationsQueue,
		id:                 id,
	}
}
