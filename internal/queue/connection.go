package queue

import (
	logger "deposit-collector/pkg/logger"
	rabbitmq "deposit-collector/pkg/rabbitmq"
	utils "deposit-collector/pkg/utils"
)

func GetQueueConnection(url string, logger *logger.Logger) *rabbitmq.RabbitMQ {
	rabbitmq, err := rabbitmq.NewRabbitMQ(url)
	if err != nil {
		utils.FailOnError(logger, err, "Error creating RabbitMQ connection")
	}
	return rabbitmq
}
