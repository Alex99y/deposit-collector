package config

import (
	fmt "fmt"

	config "deposit-collector/internal/config"
	logger "deposit-collector/pkg/logger"
	utils "deposit-collector/pkg/utils"
)

const (
	RPCFilePath         = "RPC_FILE_PATH"
	AllowMultiThreading = "ALLOW_MULTI_THREADING"
	MaxWorkers          = "MAX_WORKERS"
)

type ManagerConfig struct {
	config.CommonConfig
	RPCFilePath         string
	AllowMultiThreading bool
	MaxWorkers          int
}

func GetManagerConfig(logger *logger.Logger) *ManagerConfig {
	commonConfig := config.GetCommonConfig(logger)

	maxWorkers, err := utils.StringToInt(config.GetEnvOrDefault(MaxWorkers, "1"))
	if err != nil {
		utils.FailOnError(logger, err, "Error converting MaxWorkers to int")
	}

	if maxWorkers < 1 {
		utils.FailOnError(
			logger,
			fmt.Errorf("%s must be greater than 0", MaxWorkers),
			"",
		)
	}

	return &ManagerConfig{
		CommonConfig: *commonConfig,
		RPCFilePath:  config.GetEnvOrThrow(logger, RPCFilePath),
		AllowMultiThreading: config.GetEnvOrDefault(
			AllowMultiThreading, "false",
		) == "true",
		MaxWorkers: maxWorkers,
	}
}
