package queue

import (
	rabbitmq "deposit-collector/shared/rabbitmq"
)

func GetQueueConnection(url string) (*rabbitmq.RabbitMQ, error) {
	rabbitmq, err := rabbitmq.NewRabbitMQ(url)
	if err != nil {
		return nil, err
	}
	return rabbitmq, nil
}
