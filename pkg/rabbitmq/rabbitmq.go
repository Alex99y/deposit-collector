package rabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewRabbitMQ(url string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &RabbitMQ{conn: conn, ch: ch}, nil
}

func (r *RabbitMQ) Close() {
	_ = r.ch.Close()
	_ = r.conn.Close()
}

func (r *RabbitMQ) SetQos(prefetchCount int, prefetchSize int, global bool) {
	_ = r.ch.Qos(prefetchCount, prefetchSize, global)
}
