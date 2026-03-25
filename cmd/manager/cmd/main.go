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
	memorycache "deposit-collector/internal/memory_cache"
	system "deposit-collector/internal/system"
	transaction_service "deposit-collector/internal/transaction_service"
	provider "deposit-collector/pkg/crypto/provider"
	logger "deposit-collector/pkg/logger"
	postgresql "deposit-collector/pkg/postgresql"
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

	db, err := postgresql.SetupPostgresConnection(managerConfig.PostgresURL)
	if err != nil {
		utils.FailOnError(logger, err, "Error creating postgres connection")
	}
	defer db.Close()

	providerPool := provider.NewProviderPool(
		managerConfig.RPCFilePath,
		ctx,
		logger,
	)

	systemRepository := system.NewSystemRepository(db)
	chainsCache, err := memorycache.NewChainsCache(systemRepository)
	if err != nil {
		utils.FailOnError(logger, err, "Error creating chains cache")
	}

	transactionRepository := transaction_service.NewTransactionRepository(db)
	transactionService := transaction_service.NewTransactionService(
		providerPool,
		transactionRepository,
		chainsCache,
	)

	logger.Info(fmt.Sprintf("starting manager with workers=%d", maxWorkers))

	workers := make([]*worker.Worker, maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		workers[i] = worker.NewWorker(rmq, transactionService, i, logger)
	}

	for _, worker := range workers {
		worker.Start(ctx)
		defer func() {
			_ = worker.Stop(ctx)
		}()
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
