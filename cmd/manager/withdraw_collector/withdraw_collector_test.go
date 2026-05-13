package withdraw_collector

import (
	"context"
	"testing"

	system "deposit-collector/internal/system"
	transaction_service "deposit-collector/internal/transaction_service"
	walletservices "deposit-collector/internal/wallet_services"
	provider "deposit-collector/pkg/crypto/provider"
	logger "deposit-collector/pkg/logger"
)

func TestNewWithdrawCollectorStoresStartupDependencies(t *testing.T) {
	systemRepository := &system.SystemRepository{}
	providerPool := &provider.ProviderPool{}
	transactionRepository := &transaction_service.TransactionRepository{}
	walletServices := &walletservices.WalletServices{}
	keys := transaction_service.WithdrawCollectorPrivateKeys{
		system.ChainPlatformBTC: "btc-private-key",
	}

	collector := NewWithdrawCollector(
		context.Background(),
		keys,
		systemRepository,
		providerPool,
		transactionRepository,
		walletServices,
		logger.NewLogger(),
	)

	if collector.systemRepository != systemRepository {
		t.Fatal("system repository was not stored on withdraw collector")
	}
	if collector.collectorPrivateKeys[system.ChainPlatformBTC] != "btc-private-key" {
		t.Fatal("collector private keys were not stored on withdraw collector")
	}
}

func TestNewWorkerForChainUsesConfiguredPrivateKey(t *testing.T) {
	providerPool := &provider.ProviderPool{}
	transactionRepository := &transaction_service.TransactionRepository{}
	walletServices := &walletservices.WalletServices{}
	collector := &WithdrawCollector{
		transactionRepository: transactionRepository,
		providerPool:          providerPool,
		walletServices:        walletServices,
		logger:                logger.NewLogger(),
	}
	chain := system.SupportedChain{
		NewSupportedChainRequest: system.NewSupportedChainRequest{
			ChainName:     "bitcoin",
			ChainPlatform: system.ChainPlatformBTC,
		},
	}

	worker := collector.newWorkerForChain(chain, nil, "btc-private-key")

	if worker.privateKey != "btc-private-key" {
		t.Fatal("withdraw worker did not receive configured private key")
	}
	if worker.chain.ChainPlatform != system.ChainPlatformBTC {
		t.Fatal("withdraw worker did not receive configured chain")
	}
	if worker.providerPool != providerPool {
		t.Fatal("withdraw worker did not receive provider pool")
	}
	if worker.transactionRepository != transactionRepository {
		t.Fatal("withdraw worker did not receive transaction repository")
	}
	if worker.walletServices != walletServices {
		t.Fatal("withdraw worker did not receive wallet services")
	}
}
