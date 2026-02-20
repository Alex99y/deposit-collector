package config

import (
	config "deposit-collector/internal/config"
	logger "deposit-collector/pkg/logger"
)

type ManagerConfig struct {
	config.CommonConfig
}

func GetManagerConfig(logger *logger.Logger) *ManagerConfig {
	commonConfig := config.GetCommonConfig(logger)
	return &ManagerConfig{
		CommonConfig: *commonConfig,
	}
}
