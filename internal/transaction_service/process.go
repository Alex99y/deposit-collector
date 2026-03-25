package transaction_service

import (
	memorycache "deposit-collector/internal/memory_cache"
	queue "deposit-collector/internal/queue"
	system "deposit-collector/internal/system"
	evm_utils "deposit-collector/pkg/crypto/evm"
	provider "deposit-collector/pkg/crypto/provider"
	utils "deposit-collector/pkg/utils"
)

type ProcessedDepositOperation struct {
	TokenAddress string
	Amount       int64
}

func processEVMDepositOperation(
	provider *provider.EVMProvider,
	operation *queue.DepositOperationEvent,
) (*ProcessedDepositOperation, error) {
	txInfo, err := provider.GetTxInfo(operation.DepositTxHash)
	if err != nil {
		return nil, err
	}

	var tokenAddress string
	var amount int64
	var txTargetAddress string

	if len(txInfo.Input) == 0 {
		// Native transfer
		tokenAddress = "native"
		amount, err = utils.StringToInt64(txInfo.Amount)
		if err != nil {
			return nil, err
		}
		txTargetAddress = txInfo.To
	} else {
		// ERC20 transfer
		transfers := evm_utils.FindERC20Transfers(txInfo.TxReceipt)
		if len(transfers) == 0 {
			return nil, utils.NewError("no ERC20 transfer found")
		}
		tokenAddress = transfers[0].Token.Hex()
		amount = transfers[0].Value.Int64()
		txTargetAddress = transfers[0].To.Hex()
	}

	if txTargetAddress != operation.TargetAddress {
		return nil, utils.NewError(
			"invalid target address, expected: " + operation.TargetAddress +
				", got: " + txTargetAddress,
		)
	}

	return &ProcessedDepositOperation{
		TokenAddress: tokenAddress,
		Amount:       amount,
	}, nil
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
	providerPool *provider.ProviderPool,
	chainsCache *memorycache.ChainsCache,
	operation *queue.DepositOperationEvent,
) (*ProcessedDepositOperation, error) {
	chainPlatform := chainsCache.GetPlatformByChainName(
		operation.TargetChainName,
	)
	switch chainPlatform {
	case system.ChainPlatformEVM:
		evmProvider := providerPool.GetEVMProvider(operation.TargetChainName)
		return processEVMDepositOperation(evmProvider, operation)
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
