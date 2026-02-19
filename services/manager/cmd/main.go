package main

import (
	context "context"

	queue "deposit-collector/internal/queue"
	config "deposit-collector/services/manager/config"
	logger "deposit-collector/shared/logger"
	rabbitmq "deposit-collector/shared/rabbitmq"
	utils "deposit-collector/shared/utils"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := logger.NewLogger()
	managerConfig := config.GetManagerConfig(logger)
	rmq, err := queue.GetQueueConnection(managerConfig.RabbitMQURL)
	if err != nil {
		utils.FailOnError(logger, err, "Error creating RabbitMQ connection")
	}
	defer rmq.Close()
	rmq.SetQos(1, 0, false)
	operationsQueue, err := rabbitmq.GetQueue(
		rmq, string(queue.OperationsQueue), logger,
	)
	if err != nil {
		utils.FailOnError(logger, err, "Error creating operations queue")
	}

	logger.Info("Waiting for messages from operations queue")
	forever := make(chan struct{})
	defer close(forever)
	err = operationsQueue.Consume(ctx, func(args *rabbitmq.ConsumeArgs) {
		if ctx.Err() != nil {
			logger.Error("Context cancelled, stopping consume")
			_ = args.Nack()
			return
		}
		logger.Info(args.Message())
		_ = args.Ack()
	})
	if err != nil {
		utils.FailOnError(logger, err, "Error consuming operations queue")
	}
	<-forever
}
