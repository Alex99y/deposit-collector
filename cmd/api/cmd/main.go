package main

import (
	context "context"
	fmt "fmt"
	os "os"
	signal "os/signal"
	syscall "syscall"
	time "time"

	config "deposit-collector/cmd/api/config"
	worker "deposit-collector/cmd/api/worker"
	api "deposit-collector/internal/api"
	memorycache "deposit-collector/internal/memory_cache"
	metrics "deposit-collector/internal/metrics"
	system "deposit-collector/internal/system"
	users "deposit-collector/internal/users"
	walletservices "deposit-collector/internal/wallet_services"
	crypto_utils "deposit-collector/pkg/crypto"
	logger "deposit-collector/pkg/logger"
	observability "deposit-collector/pkg/observability"
	postgresql "deposit-collector/pkg/postgresql"
	utils "deposit-collector/pkg/utils"
)

func main() {
	logger := logger.NewLogger()

	// Read config from env
	apiConfig := config.GetAPIConfig(logger)

	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	// Setup postgres connection
	db, err := postgresql.SetupPostgresConnection(apiConfig.PostgresURL)
	if err != nil {
		utils.FailOnError(logger, err, "error setting up postgres connection")
	}
	defer db.Close()

	promRegistry := observability.NewPrometheusRegistry()

	promMetrics := observability.NewPrometheusMetrics(
		observability.PrometheusMetricsConfig{
			Registerer: promRegistry,
		},
	)

	if promMetrics == nil {
		utils.FailOnError(logger, err, "error creating prometheus metrics")
	}

	apiMetrics, err := metrics.NewApiMetrics(promMetrics)
	if err != nil {
		utils.FailOnError(logger, err, "error creating metrics")
	}
	repositoryMetrics, err := metrics.NewRepositoryMetrics(promMetrics)
	if err != nil {
		utils.FailOnError(logger, err, "error creating repository metrics")
	}

	// Setup API services
	walletService := walletservices.NewWalletServices(
		apiConfig.WalletSeed, logger,
	)

	publisher := worker.NewPublisher(appCtx, apiConfig.RabbitMQURL, logger)
	err = publisher.Start(appCtx)
	if err != nil {
		utils.FailOnError(logger, err, "Error starting publisher")
	}
	logger.Info("publisher started")
	defer publisher.Close()

	systemRepository := system.NewSystemRepository(db, repositoryMetrics)
	systemService := system.NewSystemService(systemRepository, logger)
	systemHandler := system.NewSystemHandler(systemService, logger)

	chainsCache, err := memorycache.NewChainsCache(systemRepository)
	if err != nil {
		utils.FailOnError(logger, err, "Error creating chains cache")
	}

	cryptoUtils := crypto_utils.NewCryptoUtils(apiConfig.BitcoinNetwork)

	usersRepository := users.NewUsersRepository(appCtx, db, repositoryMetrics)
	usersService := users.NewUserService(usersRepository, walletService, logger)
	usersHandler := users.NewUserHandler(
		usersService, chainsCache, publisher, cryptoUtils, logger,
	)

	serverDependencies := api.ServerDependencies{
		Logger:        logger,
		UsersHandler:  usersHandler,
		SystemHandler: systemHandler,
		Metrics:       apiMetrics,
	}

	// Setup HTTP servers
	promServer := observability.NewPrometheusServer(
		apiConfig.MetricsPort,
		promMetrics,
		logger,
	)

	if promServer == nil {
		utils.FailOnError(logger, err, "error creating prometheus server")
	}

	err = promServer.Start()
	if err != nil {
		utils.FailOnError(logger, err, "error starting prometheus server")
	}
	defer func() {
		if err := promServer.Stop(); err != nil {
			logger.Error(fmt.Sprintf("error stopping prometheus server: %v", err))
		}
	}()

	server := api.NewServer(serverDependencies)

	serverErrCh := make(chan error, 1)

	go func() {
		logger.Info(
			fmt.Sprintf("starting server on %s:%d", apiConfig.Host, apiConfig.Port),
		)
		serverErrCh <- server.Start(apiConfig.Port, apiConfig.Host)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case sig := <-quit:
		logger.Info(fmt.Sprintf("shutdown server ... signal=%s", sig))
	case err := <-serverErrCh:
		if err != nil {
			utils.FailOnError(logger, err, "error starting server")
		}
		return
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(), 20*time.Second,
	)
	defer shutdownCancel()

	err = server.Shutdown(shutdownCtx)
	if err != nil {
		utils.FailOnError(logger, err, "error shutting down server")
	}

	publisher.Close()

	logger.Info("server exiting")
}
