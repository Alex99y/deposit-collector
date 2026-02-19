package main

import (
	context "context"
	fmt "fmt"
	os "os"
	signal "os/signal"
	syscall "syscall"
	time "time"

	config "deposit-collector/services/api/config"
	http "deposit-collector/services/api/http"
	logger "deposit-collector/shared/logger"
	utils "deposit-collector/shared/utils"
)

func main() {
	logger := logger.NewLogger()

	apiConfig := config.GetAPIConfig(logger)

	server := http.NewServer(logger, apiConfig.Port, apiConfig.Host)

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

	err := server.Shutdown(ctx)
	if err != nil {
		utils.FailOnError(logger, err, "error shutting down server")
	}

	<-ctx.Done()

	logger.Info("server exiting")
}
