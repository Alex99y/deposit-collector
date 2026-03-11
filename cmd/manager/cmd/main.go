package main

import (
	context "context"
	fmt "fmt"
	os "os"
	signal "os/signal"
	runtime "runtime"
	syscall "syscall"

	config "deposit-collector/cmd/manager/config"
	worker "deposit-collector/cmd/manager/worker"
	logger "deposit-collector/pkg/logger"
	rabbitmq "deposit-collector/pkg/rabbitmq"
	utils "deposit-collector/pkg/utils"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := logger.NewLogger()
	managerConfig := config.GetManagerConfig(logger)

	rmq, err := rabbitmq.NewRabbitMQ(managerConfig.RabbitMQURL)
	if err != nil {
		utils.FailOnError(logger, err, "Error creating RabbitMQ client")
	}

	var maxWorkers int
	totalCpu := runtime.NumCPU()
	if managerConfig.AllowMultiThreading {
		if managerConfig.MaxWorkers > totalCpu {
			maxWorkers = totalCpu
		} else {
			maxWorkers = managerConfig.MaxWorkers
		}
	} else {
		maxWorkers = 1
	}

	logger.Info(fmt.Sprintf("starting manager with workers=%d", maxWorkers))

	workers := make([]*worker.Worker, maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		workers[i] = worker.NewWorker(rmq, i, logger)
	}

	for _, worker := range workers {
		worker.Start(ctx)
		defer worker.Stop(ctx)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case sig := <-quit:
		logger.Info(fmt.Sprintf("shutdown manager ... signal=%s", sig))
	case <-ctx.Done():
		logger.Info("manager exiting")
	}
}
