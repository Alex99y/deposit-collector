package worker

import (
	context "context"
	errors "errors"
	fmt "fmt"
	"sync"
	time "time"

	queue "deposit-collector/internal/queue"
	logger "deposit-collector/pkg/logger"
	rabbitmq "deposit-collector/pkg/rabbitmq"
	utils "deposit-collector/pkg/utils"
)

type publishReq struct {
	ctx        context.Context
	routingKey string
	body       queue.OperationEvent

	mandatory  bool
	immediate  bool
	persistent bool

	// respuesta a la request (para poder esperar ACK del broker)
	resp chan error
}

type Publisher struct {
	isRunning   bool
	mu          sync.Mutex
	rabbitMQURL string
	rmq         *rabbitmq.RabbitMQClient
	logger      *logger.Logger

	operationsQueue *queue.OperationQueue

	reqCh chan publishReq

	done chan struct{}
}

func (p *Publisher) Start(ctx context.Context) error {
	// Make sure to run only one instance of the publisher
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.isRunning {
		return errors.New("publisher already running")
	}
	p.isRunning = true

	go func() {
		p.run()
	}()
	return nil
}

func (p *Publisher) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.isRunning {
		return
	}
	p.isRunning = false

	close(p.done)

	if p.operationsQueue != nil {
		_ = p.operationsQueue.Close()
		p.operationsQueue = nil
	}

	if p.rmq != nil {
		_ = p.rmq.Close()
		p.rmq = nil
	}
}

func (p *Publisher) run() {
	backoff := time.Second

	for {
		select {
		case <-p.done:
			return

		case req := <-p.reqCh:
			err := p.operationsQueue.PublishOperationEvent(
				req.ctx,
				req.body,
				req.mandatory,
				req.immediate,
			)

			if err != nil && isConnOrChannelClosed(err) {
				p.logger.Error(
					fmt.Sprintf(
						"channel/conn closed, reconnecting: %v",
						err,
					),
				)

				if recErr := p.reconnectLoop(&backoff); recErr != nil {
					req.resp <- recErr
					continue
				}

				err = p.operationsQueue.PublishOperationEvent(
					req.ctx,
					req.body,
					req.mandatory,
					req.immediate,
				)
			}

			req.resp <- err
		}
	}
}

func (p *Publisher) PublishDepositOperation(
	ctx context.Context,
	id string,
	targetChainName string,
	despositTxHash string,
	targetAddress string,
) error {
	req := publishReq{
		ctx:        ctx,
		routingKey: "deposit.operation",
		body: queue.OperationEvent{
			Id:            id,
			OperationType: queue.OperationTypeDeposit,
			OperationData: queue.DepositOperationEvent{
				TargetChainName: targetChainName,
				DepositTxHash:   despositTxHash,
				TargetAddress:   targetAddress,
			},
		},
		mandatory:  false,
		immediate:  false,
		persistent: true,
		resp:       make(chan error),
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.done:
		return errors.New("publisher closed")
	case p.reqCh <- req:
	}

	select {
	case err := <-req.resp:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *Publisher) reconnectLoop(backoff *time.Duration) error {
	for {
		select {
		case <-p.done:
			return errors.New("publisher closed")
		default:
		}

		rmq, operationsQueue, err := connect(p.rabbitMQURL, p.logger)

		if err == nil {
			p.rmq = rmq
			p.operationsQueue = operationsQueue
			*backoff = time.Second
			p.logger.Info("[publisher] reconnected")
			return nil
		}

		time.Sleep(*backoff)
		if *backoff < 15*time.Second {
			*backoff *= 2
		}
	}
}

func isConnOrChannelClosed(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return utils.ContainsAny(s,
		"channel/connection is not open",
		"connection closed",
		"exception (504)",
		"EOF",
	)
}

func connect(
	rabbitMQURL string,
	logger *logger.Logger,
) (*rabbitmq.RabbitMQClient, *queue.OperationQueue, error) {

	rmq, err := rabbitmq.NewRabbitMQ(rabbitMQURL)
	if err != nil {
		return nil, nil, utils.NewError(
			fmt.Sprintf("error creating RabbitMQ client: %v", err),
		)
	}

	publishQueue, err := rabbitmq.GetQueue(
		rmq,
		rabbitmq.ChannelArgs{
			PrefetchCount: 1,
			PrefetchSize:  0,
		},
		rabbitmq.QueueArgs{
			Name:       string(queue.OperationsQueue),
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Args:       nil,
		},
		logger,
	)

	operationsQueue := queue.NewOperationsQueue(publishQueue, logger)
	if err != nil {
		return nil, nil, utils.NewError(
			fmt.Sprintf("error creating publish queue: %v", err),
		)
	}

	return rmq, operationsQueue, nil
}

func NewPublisher(
	ctx context.Context,
	rabbitMQURL string,
	logger *logger.Logger,
) *Publisher {

	rmq, operationsQueue, err := connect(rabbitMQURL, logger)
	if err != nil {
		utils.FailOnError(logger, err, "Error creating publisher")
	}

	return &Publisher{
		rabbitMQURL:     rabbitMQURL,
		rmq:             rmq,
		operationsQueue: operationsQueue,
		logger:          logger,
		reqCh:           make(chan publishReq),
		done:            make(chan struct{}),
	}
}
