package main

import (
	context "context"
	"fmt"

	queue "deposit-collector/internal/queue"
	config "deposit-collector/services/api/config"
	logger "deposit-collector/shared/logger"
	rabbitmq "deposit-collector/shared/rabbitmq"
	utils "deposit-collector/shared/utils"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := logger.NewLogger()
	apiConfig := config.GetAPIConfig(logger)
	rmq, err := queue.GetQueueConnection(apiConfig.RabbitMQURL)
	if err != nil {
		utils.FailOnError(logger, err, "Error creating RabbitMQ connection")
	}
	defer rmq.Close()
	operationsQueue, err := rabbitmq.GetQueue(
		rmq, string(queue.OperationsQueue), logger,
	)
	if err != nil {
		utils.FailOnError(logger, err, "Error creating operations queue")
	}
	err = operationsQueue.Publish(ctx, "Hello World")
	if err != nil {
		utils.FailOnError(logger, err, "Error publishing to operations queue")
	}
	fmt.Println("Message published to operations queue")
}
