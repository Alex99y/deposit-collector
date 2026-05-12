package config

import (
	fmt "fmt"
	time "time"

	config "deposit-collector/internal/config"
	system "deposit-collector/internal/system"
	tx_service "deposit-collector/internal/transaction_service"
	logger "deposit-collector/pkg/logger"
	utils "deposit-collector/pkg/utils"
)

const (
	RPCFilePath                                  = "RPC_FILE_PATH"
	AllowMultiThreading                          = "ALLOW_MULTI_THREADING"
	MaxWorkers                                   = "MAX_WORKERS"
	DepositCollectorInterval                     = "DEPOSIT_COLLECTOR_INTERVAL"
	DepositCollectorEVMDestinationDepositAddress = "DEPOSIT_COLLECTOR_EVM_DESTINATION_DEPOSIT_ADDRESS"
	DepositCollectorBTCDestinationDepositAddress = "DEPOSIT_COLLECTOR_BTC_DESTINATION_DEPOSIT_ADDRESS"
	DepositCollectorSOLDestinationDepositAddress = "DEPOSIT_COLLECTOR_SOL_DESTINATION_DEPOSIT_ADDRESS"
	WithdrawCollectorEVMPrivateKey               = "WITHDRAW_COLLECTOR_EVM_PRIVATE_KEY"
	WithdrawCollectorBTCPrivateKey               = "WITHDRAW_COLLECTOR_BTC_PRIVATE_KEY"
	WithdrawCollectorSOLPrivateKey               = "WITHDRAW_COLLECTOR_SOL_PRIVATE_KEY"
)

type ManagerConfig struct {
	config.CommonConfig
	RPCFilePath                          string
	AllowMultiThreading                  bool
	MaxWorkers                           int
	DepositCollectorInterval             time.Duration
	DepositCollectorDestinationAddresses tx_service.DestinationDepositAddress
	WithdrawCollectorPrivateKeys         tx_service.WithdrawCollectorPrivateKeys
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

	depositCollectorInterval, err := time.ParseDuration(
		config.GetEnvOrDefault(DepositCollectorInterval, "10s"),
	)
	if err != nil {
		utils.FailOnError(
			logger, err, "Error converting DepositCollectorInterval to duration",
		)
	}

	depositCollectorEVMDepositAddress :=
		config.GetEnvOrDefault(DepositCollectorEVMDestinationDepositAddress, "")
	depositCollectorBTCDepositAddress :=
		config.GetEnvOrDefault(DepositCollectorBTCDestinationDepositAddress, "")
	depositCollectorSOLDepositAddress :=
		config.GetEnvOrDefault(DepositCollectorSOLDestinationDepositAddress, "")

	withdrawCollectorEVMPrivateKey :=
		config.GetEnvOrDefault(WithdrawCollectorEVMPrivateKey, "")
	withdrawCollectorBTCPrivateKey :=
		config.GetEnvOrDefault(WithdrawCollectorBTCPrivateKey, "")
	withdrawCollectorSOLPrivateKey :=
		config.GetEnvOrDefault(WithdrawCollectorSOLPrivateKey, "")

	return &ManagerConfig{
		CommonConfig: *commonConfig,
		RPCFilePath:  config.GetEnvOrThrow(logger, RPCFilePath),
		AllowMultiThreading: config.GetEnvOrDefault(
			AllowMultiThreading, "false",
		) == "true",
		MaxWorkers:               maxWorkers,
		DepositCollectorInterval: depositCollectorInterval,
		DepositCollectorDestinationAddresses: tx_service.DestinationDepositAddress{
			system.ChainPlatformEVM: depositCollectorEVMDepositAddress,
			system.ChainPlatformBTC: depositCollectorBTCDepositAddress,
			system.ChainPlatformSOL: depositCollectorSOLDepositAddress,
		},
		WithdrawCollectorPrivateKeys: tx_service.WithdrawCollectorPrivateKeys{
			system.ChainPlatformEVM: withdrawCollectorEVMPrivateKey,
			system.ChainPlatformBTC: withdrawCollectorBTCPrivateKey,
			system.ChainPlatformSOL: withdrawCollectorSOLPrivateKey,
		},
	}
}
