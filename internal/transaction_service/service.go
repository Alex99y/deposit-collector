package transaction_service

import (
	memorycache "deposit-collector/internal/memory_cache"
	queue "deposit-collector/internal/queue"
	provider "deposit-collector/pkg/crypto/provider"
)

type TransactionService struct {
	providerPool *provider.ProviderPool
	chainsCache  *memorycache.ChainsCache
	repository   *TransactionRepository
}

func (s *TransactionService) ValidateAndStoreDepositOperation(
	operation *queue.DepositOperationEvent,
) error {
	processedOperation, err := ProcessDepositOperation(
		s.chainsCache,
		operation,
	)
	if err != nil {
		return err
	}

	tokenAddressInfo := s.chainsCache.GetTokenAddressByChainNameAndTokenAddress(
		operation.TargetChainName,
		processedOperation.TokenAddress,
	)

	err = s.repository.EndorseDepositOperation(
		operation.UserDbID,
		operation.TargetAddressDbId,
		tokenAddressInfo.TokenAddressDbID,
		processedOperation.Amount,
		operation.DepositTxHash,
	)
	if err != nil {
		return err
	}

	return nil
}

func NewTransactionService(
	providerPool *provider.ProviderPool,
	repository *TransactionRepository,
	chainsCache *memorycache.ChainsCache,
) *TransactionService {
	return &TransactionService{
		providerPool: providerPool,
		repository:   repository,
	}
}
