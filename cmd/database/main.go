package main

import (
	config "deposit-collector/internal/config"
	logger "deposit-collector/pkg/logger"
	utils "deposit-collector/pkg/utils"

	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	logger := logger.NewLogger()
	commonConfig := config.GetCommonConfig(logger)

	logger.Info("Migrating database")
	m, err := migrate.New(
		"file://cmd/database/migrations",
		commonConfig.PostgresURL,
	)
	if err != nil {
		utils.FailOnError(logger, err, "Error creating migration")
	}
	if err := m.Up(); err != nil {
		utils.FailOnError(logger, err, "Error migrating database")
	}
}
