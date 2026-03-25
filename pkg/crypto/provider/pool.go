package provider

import (
	context "context"
	json "encoding/json"
	io "io"
	os "os"

	logger "deposit-collector/pkg/logger"
	utils "deposit-collector/pkg/utils"
)

type RpcConfig struct {
	Url              string `json:"url"`
	ChainID          int    `json:"chain_id"`
	MinConfirmations int    `json:"min_confirmations"`
}
type RpcConfigByChainPlatform struct {
	EVM RpcConfig `json:"EVM"`
	BTC RpcConfig `json:"BTC"`
	SOL RpcConfig `json:"SOL"`
}
type ProviderConfig struct {
	Rpc map[string]RpcConfigByChainPlatform `json:"rpc"`
}

type ProviderPool struct {
	evmProviders map[string]*EVMProvider
}

func (p *ProviderPool) GetEVMProvider(
	chainName string,
) *EVMProvider {
	return p.evmProviders[chainName]
}

func NewProviderPool(
	providerFilePath string,
	context context.Context,
	logger *logger.Logger,
) *ProviderPool {
	providerConfig := readProviderConfig(providerFilePath, logger)
	evmProvidersMap := make(map[string]*EVMProvider)
	for chainName, rpcConfig := range providerConfig.Rpc {
		evmProvidersMap[chainName] = NewEVMProvider(
			rpcConfig.EVM.Url,
			rpcConfig.EVM.ChainID,
			rpcConfig.EVM.MinConfirmations,
			context,
			logger,
		)
	}
	return &ProviderPool{evmProviders: evmProvidersMap}
}

func readProviderConfig(
	providerFilePath string,
	logger *logger.Logger,
) ProviderConfig {
	jsonFile, err := os.Open(providerFilePath)
	if err != nil {
		utils.FailOnError(logger, err, "Error opening provider config file")
	}
	defer jsonFile.Close()
	jsonBytes, err := io.ReadAll(jsonFile)
	if err != nil {
		utils.FailOnError(logger, err, "Error reading provider config file")
	}
	var providerConfig ProviderConfig
	err = json.Unmarshal(jsonBytes, &providerConfig)
	if err != nil {
		utils.FailOnError(logger, err, "Error unmarshalling provider config file")
	}
	return providerConfig
}
