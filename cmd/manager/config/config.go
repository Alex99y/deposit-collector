package config

import (
	config "deposit-collector/internal/config"
	logger "deposit-collector/pkg/logger"
)

const RPCFilePath = "RPC_FILE_PATH"

type ManagerConfig struct {
	config.CommonConfig
	RPCFilePath string
}

func GetManagerConfig(logger *logger.Logger) *ManagerConfig {
	commonConfig := config.GetCommonConfig(logger)
	return &ManagerConfig{
		CommonConfig: *commonConfig,
		RPCFilePath:  config.GetEnvOrThrow(logger, RPCFilePath),
	}
}
