package main

import (
	context "context"
	fmt "fmt"
	os "os"
	signal "os/signal"
	syscall "syscall"
	"time"

	config "deposit-collector/cmd/api/config"
	http "deposit-collector/cmd/api/http"
	handlers "deposit-collector/cmd/api/http/handlers"
	worker "deposit-collector/cmd/api/worker"
	memorycache "deposit-collector/internal/memory_cache"
	system "deposit-collector/internal/system"
	users "deposit-collector/internal/users"
	walletservices "deposit-collector/internal/wallet_services"
	logger "deposit-collector/pkg/logger"
	postgresql "deposit-collector/pkg/postgresql"
	utils "deposit-collector/pkg/utils"
)

func main() {
	logger := logger.NewLogger()

	apiConfig := config.GetAPIConfig(logger)

	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	db, err := postgresql.SetupPostgresConnection(apiConfig.PostgresURL)
	if err != nil {
		utils.FailOnError(logger, err, "error setting up postgres connection")
	}
	defer db.Close()

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

	systemRepository := system.NewSystemRepository(db)
	systemService := system.NewSystemService(systemRepository, logger)
	systemHandler := handlers.NewSystemHandler(systemService, logger)

	chainsCache, err := memorycache.NewChainsCache(systemRepository)
	if err != nil {
		utils.FailOnError(logger, err, "Error creating chains cache")
	}

	usersRepository := users.NewUsersRepository(appCtx, db)
	usersService := users.NewUserService(usersRepository, walletService, logger)
	usersHandler := handlers.NewUserHandler(
		usersService, chainsCache, publisher, logger,
	)

	serverDependencies := http.ServerDependencies{
		Logger:        logger,
		UsersHandler:  usersHandler,
		SystemHandler: systemHandler,
	}

	server := http.NewServer(serverDependencies)

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
