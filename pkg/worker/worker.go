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
	w.isRunning = true
	w.logger.Info(fmt.Sprintf("worker %s starting", w.id))

	go func() {
		for {
			if !w.isRunning {
				w.hasStopped = true
				break
			}

			w.run()

			time.Sleep(time.Duration(w.interval) * time.Second)
		}
	}()

	return nil
}

func (w *Worker) Stop() error {
	w.logger.Info(fmt.Sprintf("stopping worker %s...", w.id))
	w.isRunning = false
	for {
		if w.hasStopped {
			break
		}
		time.Sleep(time.Duration(1) * time.Second)
	}
	return nil
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
