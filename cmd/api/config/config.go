package config

import (
	"fmt"
	"strconv"

	config "deposit-collector/internal/config"
	logger "deposit-collector/pkg/logger"
	utils "deposit-collector/pkg/utils"
)

type APIConfig struct {
	config.CommonConfig
	Port int
	Host string
}

func GetAPIConfig(logger *logger.Logger) *APIConfig {
	commonConfig := config.GetCommonConfig(logger)

	port, err := strconv.Atoi(config.GetEnvOrDefault(Port, "8080"))
	if err != nil {
		utils.FailOnError(
			logger,
			err,
			fmt.Sprintf("Error converting %s to int", Port),
		)
	}
	return &APIConfig{
		CommonConfig: *commonConfig,
		Port:         port,
		Host:         config.GetEnvOrDefault(Host, "0.0.0.0"),
	}
}
