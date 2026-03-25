package transaction_service

import (
	memorycache "deposit-collector/internal/memory_cache"
	queue "deposit-collector/internal/queue"
	provider "deposit-collector/pkg/crypto/provider"
	logger "deposit-collector/pkg/logger"
	utils "deposit-collector/pkg/utils"
)

type TransactionService struct {
	providerPool *provider.ProviderPool
	chainsCache  *memorycache.ChainsCache
	repository   *TransactionRepository
	logger       *logger.Logger
}

func (s *TransactionService) ValidateAndStoreDepositOperation(
	operation *queue.DepositOperationEvent,
) (*utils.CustomError, error) {
	processedOperation, err := ProcessDepositOperation(
		s.providerPool,
		s.chainsCache,
		operation,
	)
	if err != nil {
		if err.Error() == "not found" {
			return utils.NewCustomError("operation not found", true), nil
		}
		return nil, err
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
		return nil, err
	}

	return nil, nil
}

func NewTransactionService(
	providerPool *provider.ProviderPool,
	repository *TransactionRepository,
	chainsCache *memorycache.ChainsCache,
	logger *logger.Logger,
) *TransactionService {
	if providerPool == nil || repository == nil || chainsCache == nil {
		panic("Invalid transaction service dependencies")
	}
	return &TransactionService{
		chainsCache:  chainsCache,
		providerPool: providerPool,
		repository:   repository,
		logger:       logger,
	}
}
