package worker

import (
	context "context"
	errors "errors"
	fmt "fmt"
	time "time"

	logger "deposit-collector/pkg/logger"
)

type Worker struct {
	id         string
	ctx        context.Context
	logger     logger.Logger
	interval   int64
	run        func()
	isRunning  bool
	hasStopped bool
}

func (w *Worker) Start() error {
	if w.isRunning {
		return errors.New("worker is already running")
	}
	w.hasStopped = false
	w.isRunning = true
	w.logger.Info(fmt.Sprintf("worker %s starting", w.id))

	go func() {
		defer func() {
			w.hasStopped = true
		}()

		for {
			select {
			case <-w.ctx.Done():
				return
			default:
			}

			if !w.isRunning {
				return
			}

			w.run()

			timer := time.NewTimer(time.Duration(w.interval) * time.Second)
			select {
			case <-w.ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
		}
	}()

	return nil
}

func (w *Worker) Stop() error {
	w.logger.Info(fmt.Sprintf("stopping worker %s...", w.id))
	w.isRunning = false

	for {
		if w.hasStopped {
			return nil
		}
		select {
		case <-w.ctx.Done():
			time.Sleep(100 * time.Millisecond)
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func NewWorker(
	id string,
	ctx context.Context,
	logger *logger.Logger,
	run func(),
	interval int64,
) *Worker {

	return &Worker{
		logger:   *logger,
		ctx:      ctx,
		id:       id,
		run:      run,
		interval: interval,
	}
}
