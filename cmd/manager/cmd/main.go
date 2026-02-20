package main

import (
	context "context"

	config "deposit-collector/cmd/manager/config"
	queue "deposit-collector/internal/queue"
	logger "deposit-collector/pkg/logger"
	utils "deposit-collector/pkg/utils"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := logger.NewLogger()
	managerConfig := config.GetManagerConfig(logger)
	rmq := queue.GetQueueConnection(managerConfig.RabbitMQURL, logger)
	defer rmq.Close()
	rmq.SetQos(1, 0, false)
	operationsQueue := queue.NewOperationsQueue(rmq, logger)
	logger.Info("Waiting for messages from operations queue")
	forever := make(chan struct{})
	defer close(forever)
	err := operationsQueue.Consume(ctx, func(args *queue.OperationConsumerArgs) {
		if ctx.Err() != nil {
			logger.Error("Context cancelled, stopping consume")
			_ = args.Nack()
			return
		}
		logger.Info(args.Message().OperationData.Message)
		_ = args.Ack()
	})
	if err != nil {
		utils.FailOnError(logger, err, "Error consuming operations queue")
	}
	<-forever
}
