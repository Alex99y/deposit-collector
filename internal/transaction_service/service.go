package transaction_service

import (
	memorycache "deposit-collector/internal/memory_cache"
	queue "deposit-collector/internal/queue"
	users "deposit-collector/internal/users"
	provider "deposit-collector/pkg/crypto/provider"
	logger "deposit-collector/pkg/logger"
	postgresql "deposit-collector/pkg/postgresql"
	utils "deposit-collector/pkg/utils"
)

type TransactionService struct {
	providerPool    *provider.ProviderPool
	chainsCache     *memorycache.ChainsCache
	usersRepository *users.UsersRepository
	repository      *TransactionRepository
	logger          *logger.Logger
}

func (s *TransactionService) ValidateAndStoreDepositOperation(
	operation *queue.DepositOperationEvent,
) error {
	processedOperation, err := ProcessDepositOperation(
		s.providerPool,
		s.chainsCache,
		operation,
	)
	if customError, ok := utils.IsCustomError(err); ok {
		return customError
	}
	if err != nil {
		if err.Error() == "not found" {
			return utils.NewCustomError("operation not found", false)
		}
		return err
	}

	tokenAddressInfo := s.chainsCache.GetTokenAddressByChainNameAndTokenAddress(
		operation.TargetChainName,
		processedOperation.TokenAddress,
	)

	if tokenAddressInfo == nil {
		return utils.NewCustomError("token address not found", false)
	}

	err = s.repository.EndorseDepositOperation(
		operation.UserDbID,
		operation.TargetAddressDbId,
		tokenAddressInfo.TokenAddressDbID,
		processedOperation.Amount,
		operation.DepositTxHash,
	)

	if err != nil {
		if _, ok := postgresql.UniqueConstraintViolation(err); ok {
			return utils.NewCustomError("operation already endorsed", false)
		}
		return err
	}

	return nil
}

func (s *TransactionService) ValidateAndStoreWithdrawOperation(
	operation *queue.WithdrawOperationEvent,
) error {
	if operation.WithdrawAmount <= 0 {
		return utils.NewCustomError("withdraw amount must be greater than 0", false)
	}
	return s.repository.EndorseWithdrawOperation(
		operation.UserDbID,
		operation.TokenAddressDbId,
		operation.WithdrawAmount,
		operation.TargetAddress,
	)
}

func NewTransactionService(
	providerPool *provider.ProviderPool,
	usersRepository *users.UsersRepository,
	repository *TransactionRepository,
	chainsCache *memorycache.ChainsCache,
	logger *logger.Logger,
) *TransactionService {
	if providerPool == nil || usersRepository == nil ||
		repository == nil || chainsCache == nil || logger == nil {
		panic("Invalid transaction service dependencies")
	}
	return &TransactionService{
		chainsCache:     chainsCache,
		usersRepository: usersRepository,
		providerPool:    providerPool,
		repository:      repository,
		logger:          logger,
	}
}
