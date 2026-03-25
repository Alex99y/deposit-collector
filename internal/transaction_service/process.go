package transaction_service

import (
	memorycache "deposit-collector/internal/memory_cache"
	queue "deposit-collector/internal/queue"
	system "deposit-collector/internal/system"
)

type ProcessedDepositOperation struct {
	TokenAddress string
	Amount       int64
}

func processEVMDepositOperation(
	operation *queue.DepositOperationEvent,
) (*ProcessedDepositOperation, error) {
	return nil, nil
}

func processBTCDepositOperation(
	operation *queue.DepositOperationEvent,
) (*ProcessedDepositOperation, error) {
	return nil, nil
}

func processSOLDepositOperation(
	operation *queue.DepositOperationEvent,
) (*ProcessedDepositOperation, error) {
	return nil, nil
}

func ProcessDepositOperation(
	chainsCache *memorycache.ChainsCache,
	operation *queue.DepositOperationEvent,
) (*ProcessedDepositOperation, error) {
	chainPlatform := chainsCache.GetPlatformByChainName(
		operation.TargetChainName,
	)
	switch chainPlatform {
	case system.ChainPlatformEVM:
		return processEVMDepositOperation(operation)
	case system.ChainPlatformBTC:
		return processBTCDepositOperation(operation)
	case system.ChainPlatformSOL:
		return processSOLDepositOperation(operation)
	}
	return nil, nil
}

func ProcessWithdrawOperation(
	operation *queue.WithdrawOperationEvent,
) error {
	return nil
}
