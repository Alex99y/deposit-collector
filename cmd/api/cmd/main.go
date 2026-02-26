package main

import (
	context "context"
	fmt "fmt"
	os "os"
	signal "os/signal"
	syscall "syscall"
	time "time"

	config "deposit-collector/cmd/api/config"
	http "deposit-collector/cmd/api/http"
	handlers "deposit-collector/cmd/api/http/handlers"
	users "deposit-collector/internal/users"
	logger "deposit-collector/pkg/logger"
	postgresql "deposit-collector/pkg/postgresql"
	utils "deposit-collector/pkg/utils"
)

func main() {
	logger := logger.NewLogger()

	apiConfig := config.GetAPIConfig(logger)

	// Setup Postgres connection
	db, err := postgresql.SetupPostgresConnection(apiConfig.PostgresURL)
	if err != nil {
		utils.FailOnError(logger, err, "error setting up postgres connection")
	}
	defer db.Close()

	// Setup users services
	usersRepository := users.NewUsersRepository(db)
	usersService := users.NewUserService(usersRepository, logger)
	usersHandler := handlers.NewUserHandler(usersService, logger)

	// Setup server dependencies
	serverDependencies := http.ServerDependencies{
		Logger:       logger,
		UsersHandler: usersHandler,
	}

	server := http.NewServer(serverDependencies)

	go func() {
		logger.Info(
			fmt.Sprintf("starting server on %s:%d", apiConfig.Host, apiConfig.Port),
		)
		err := server.Start(apiConfig.Port, apiConfig.Host)
		if err != nil {
			utils.FailOnError(logger, err, "error starting server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.Info("shutdown server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		utils.FailOnError(logger, err, "error shutting down server")
	}

	<-ctx.Done()

	// rmq := queue.GetQueueConnection(apiConfig.RabbitMQURL, logger)
	// defer rmq.Close()
	// operationsQueue := queue.NewOperationsQueue(rmq, logger)
	// err := operationsQueue.PublishOperationEvent(ctx, queue.OperationEvent{
	// 	OperationType: queue.OperationTypeDeposit,
	// 	OperationData: queue.Operation{
	// 		Message: "Hello World Amigo!",
	// 	},
	// })
	// if err != nil {
	// 	utils.FailOnError(logger, err, "Error publishing to operations queue")
	// }

	logger.Info("server exiting")
}
