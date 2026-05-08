package processor

import (
	context "context"
	fmt "fmt"

	queue "deposit-collector/internal/queue"
	transaction_service "deposit-collector/internal/transaction_service"
	logger "deposit-collector/pkg/logger"
	rabbitmq "deposit-collector/pkg/rabbitmq"
	utils "deposit-collector/pkg/utils"
)

type Processor struct {
	logger             *logger.Logger
	transactionService *transaction_service.TransactionService
	operationsQueue    *queue.OperationQueue
	id                 int
}

func (dp *Processor) RunInBackground(ctx context.Context) {
	dp.logger.Info(fmt.Sprintf("worker %d starting", dp.id))
	go func() {
		err := dp.run(ctx)
		if err != nil {
			dp.logger.Error(fmt.Sprintf("error running worker %d: %v", dp.id, err))
		}
	}()
}

func (dp *Processor) Stop(ctx context.Context) error {
	dp.logger.Info(fmt.Sprintf("worker %d stopping", dp.id))
	return dp.operationsQueue.Close()
}

func (dp *Processor) run(ctx context.Context) error {
	err := dp.operationsQueue.Consume(
		ctx,
		func(args *queue.OperationConsumerArgs) {
			operation, err := args.OperationData()
			if err != nil {
				dp.logger.Error(
					fmt.Sprintf("Invalid operation data: %v", err),
				)
				_ = args.Reject()
				return
			}
			switch parsedOperation := operation.(type) {
			case queue.DepositOperationEvent:
				dp.logger.Info(
					"Received deposit operation id: " +
						args.OperationEvent.RequestId.String(),
				)
				err := dp.transactionService.ValidateAndStoreDepositOperation(
					&parsedOperation,
				)
				if customError, ok := utils.IsCustomError(err); ok {
					if !customError.IsRetryable() {
						dp.logger.Error(
							fmt.Sprintf("Error validating and storing deposit operation: %v",
								customError.Error()),
						)
						_ = args.Reject()
						return
					}
				}
				if err != nil {
					dp.logger.Error(
						fmt.Sprintf(
							"Error validating and storing deposit operation: %v",
							err,
						),
					)
					_ = args.Nack()
					return
				}
				dp.logger.Info(
					fmt.Sprintf(
						"Deposit operation validated and stored: %+v",
						parsedOperation.DepositTxHash,
					),
				)
				_ = args.Ack()
				return
			case queue.WithdrawOperationEvent:
				dp.logger.Info(
					"Received withdraw operation id: " +
						args.OperationEvent.RequestId.String(),
				)
				_ = args.Ack()
				return
			default:
				dp.logger.Error(fmt.Sprintf("Unknown operation type: %T", parsedOperation))
				_ = args.Reject()
				return
			}
		})
	return err
}

func NewProcessor(
	rmq *rabbitmq.RabbitMQClient,
	transactionService *transaction_service.TransactionService,
	id int,
	logger *logger.Logger,
) *Processor {
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
	return &Processor{
		logger:             logger,
		transactionService: transactionService,
		operationsQueue:    operationsQueue,
		id:                 id,
	}
}
