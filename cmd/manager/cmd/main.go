package main

import (
	context "context"
	fmt "fmt"
	os "os"
	signal "os/signal"
	runtime "runtime"
	syscall "syscall"

	config "deposit-collector/cmd/manager/config"
	deposit_collector "deposit-collector/cmd/manager/deposit_collector"
	worker "deposit-collector/cmd/manager/deposit_processor"
	memorycache "deposit-collector/internal/memory_cache"
	system "deposit-collector/internal/system"
	transaction_service "deposit-collector/internal/transaction_service"
	walletservices "deposit-collector/internal/wallet_services"
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
		managerConfig.BitcoinNetwork,
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
		logger,
	)

	logger.Info(fmt.Sprintf("starting manager with workers=%d", maxWorkers))

	workers := make([]*worker.DepositProcessor, maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		workers[i] = worker.NewDepositProcessor(rmq, transactionService, i, logger)
	}

	for _, worker := range workers {
		worker.RunInBackground(ctx)
	}

	walletServices := walletservices.NewWalletServices(
		managerConfig.WalletSeed, logger,
	)

	depositCollector, err := deposit_collector.NewDepositCollector(
		ctx,
		providerPool,
		managerConfig.DepositCollectorDestinationAddresses,
		systemRepository,
		transactionRepository,
		walletServices,
		logger,
	)
	if err != nil {
		utils.FailOnError(logger, err, "Error creating deposit collector")
	}

	if err := depositCollector.Start(); err != nil {
		utils.FailOnError(logger, err, "Error starting deposit collector")
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case sig := <-quit:
		logger.Info(fmt.Sprintf("shutdown manager ... signal=%s", sig))
		cancel()
	case <-ctx.Done():
		logger.Info("manager exiting")
	}

	depositCollector.Stop()
	for _, worker := range workers {
		_ = worker.Stop(ctx)
	}
}
