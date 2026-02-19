package config

import (
	"fmt"
	"os"
	"strconv"

	logger "deposit-collector/shared/logger"

	godotenv "github.com/joho/godotenv"

	utils "deposit-collector/shared/utils"
)

type CommonConfig struct {
	RabbitMQURL string
	MetricsPort int
}

func loadEnvFile(logger *logger.Logger) {
	err := godotenv.Load(".env")
	if err != nil {
		logger.Warn(fmt.Sprintf("Error loading .env file: %v", err))
	}
}

func GetEnvOrDefault(env string, defaultValue string) string {
	if value := os.Getenv(env); value != "" {
		return value
	}
	return defaultValue
}

func GetEnvOrThrow(logger *logger.Logger, env string) string {
	value := os.Getenv(env)
	if value == "" {
		utils.FailOnError(
			logger,
			fmt.Errorf("environment variable %s is not set", env),
			"Environment variable is not set",
		)
	}
	return value
}

func GetCommonConfig(logger *logger.Logger) *CommonConfig {
	loadEnvFile(logger)
	metricsPort, err := strconv.Atoi(GetEnvOrDefault(MetricsPort, "9090"))
	if err != nil {
		utils.FailOnError(
			logger,
			err,
			fmt.Sprintf("Error converting %s to int", MetricsPort),
		)
	}
	return &CommonConfig{
		RabbitMQURL: GetEnvOrThrow(logger, RabbitMQURL),
		MetricsPort: metricsPort,
	}
}
