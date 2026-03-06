package rabbitmq

import (
	logger "deposit-collector/pkg/logger"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
	conn   *amqp.Connection
	url    string
	logger *logger.Logger
}

func (r *RabbitMQClient) Reconnect() error {
	r.conn.Close()
	conn, err := amqp.Dial(r.url)
	if err != nil {
		return err
	}
	r.conn = conn
	return nil
}

func (r *RabbitMQClient) Close() error {
	return r.conn.Close()
}

func (s *RabbitMQClient) CreateChannel(prefetchCount int, prefetchSize int) (*amqp.Channel, error) {
	channel, err := s.conn.Channel()
	if err != nil {
		return nil, err
	}
	return channel, nil
}

func NewRabbitMQ(url string) (*RabbitMQClient, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	return &RabbitMQClient{conn: conn, url: url}, nil
}
