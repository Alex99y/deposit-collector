package queue

import (
	logger "deposit-collector/shared/logger"
	rabbitmq "deposit-collector/shared/rabbitmq"
	utils "deposit-collector/shared/utils"
)

func GetQueueConnection(url string, logger *logger.Logger) *rabbitmq.RabbitMQ {
	rabbitmq, err := rabbitmq.NewRabbitMQ(url)
	if err != nil {
		utils.FailOnError(logger, err, "Error creating RabbitMQ connection")
	}
	return rabbitmq
}
