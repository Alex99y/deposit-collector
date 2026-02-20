package rabbitmq

import (
	context "context"
	"fmt"
	time "time"

	logger "deposit-collector/pkg/logger"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Queue struct {
	rabbitmq *RabbitMQ
	queue    amqp.Queue
	logger   *logger.Logger
}

func GetQueue(
	rabbitmq *RabbitMQ,
	name string,
	logger *logger.Logger,
) (*Queue, error) {
	queue, err := rabbitmq.ch.QueueDeclare(
		name,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return nil, err
	}
	return &Queue{rabbitmq: rabbitmq, queue: queue, logger: logger}, nil
}

func (q *Queue) Name() string {
	return q.queue.Name
}

func (q *Queue) Publish(ctx context.Context, message []byte) error {
	return q.rabbitmq.ch.PublishWithContext(
		ctx,
		"",
		q.queue.Name,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        message,
		},
	)
}

type MessageMetadata struct {
	messageType     string
	messageEncoding string
	timestamp       time.Time
	expiration      string
}

func (m *MessageMetadata) GetMsgType() string      { return m.messageType }
func (m *MessageMetadata) GetMsgEncoding() string  { return m.messageEncoding }
func (m *MessageMetadata) GetTimestamp() time.Time { return m.timestamp }
func (m *MessageMetadata) GetExpiration() string   { return m.expiration }

type ConsumeArgs struct {
	id       string
	message  []byte
	metadata MessageMetadata
	ack      func() error
	nack     func() error
	reject   func() error
}

func (a *ConsumeArgs) Id() string                          { return a.id }
func (a *ConsumeArgs) RawMessage() []byte                  { return a.message }
func (a *ConsumeArgs) Ack() error                          { return a.ack() }
func (a *ConsumeArgs) Nack() error                         { return a.nack() }
func (a *ConsumeArgs) Reject() error                       { return a.reject() }
func (a *ConsumeArgs) GetMessageMetadata() MessageMetadata { return a.metadata }

type ConsumeCallback func(*ConsumeArgs)

func (q *Queue) handleDelivery(
	delivery amqp.Delivery,
	callback ConsumeCallback,
) {

	args := &ConsumeArgs{
		id:      delivery.MessageId,
		message: delivery.Body,
		metadata: MessageMetadata{
			messageType:     delivery.Type,
			messageEncoding: delivery.ContentEncoding,
			timestamp:       delivery.Timestamp,
			expiration:      delivery.Expiration,
		},
		ack:    func() error { return delivery.Ack(false) },
		nack:   func() error { return delivery.Nack(false, true) },
		reject: func() error { return delivery.Reject(false) },
	}
	callback(args)
	q.logger.Debug(
		fmt.Sprintf(
			"Delivery context done for message %s",
			delivery.MessageId,
		),
	)
}

func (q *Queue) Consume(ctx context.Context, callback ConsumeCallback) error {
	deliveries, err := q.rabbitmq.ch.ConsumeWithContext(
		ctx,
		q.queue.Name,
		"",
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return err
	}
	for delivery := range deliveries {
		go q.handleDelivery(delivery, callback)
	}

	return nil
}
