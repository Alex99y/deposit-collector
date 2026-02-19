package main

import (
	context "context"
	"fmt"

	queue "deposit-collector/internal/queue"
	config "deposit-collector/services/api/config"
	logger "deposit-collector/shared/logger"
	utils "deposit-collector/shared/utils"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := logger.NewLogger()
	apiConfig := config.GetAPIConfig(logger)
	rmq := queue.GetQueueConnection(apiConfig.RabbitMQURL, logger)
	defer rmq.Close()
	operationsQueue := queue.NewOperationsQueue(rmq, logger)
	err := operationsQueue.PublishOperationEvent(ctx, queue.OperationEvent{
		OperationType: queue.OperationTypeDeposit,
		OperationData: queue.Operation{
			Message: "Hello World Amigo!",
		},
	})
	if err != nil {
		utils.FailOnError(logger, err, "Error publishing to operations queue")
	}
	fmt.Println("Message published to operations queue")
}
