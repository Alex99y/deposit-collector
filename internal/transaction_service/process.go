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
	chainsCache *memorycache.ChainsCache,
	provider *provider.EVMProvider,
	operation *queue.DepositOperationEvent,
) (*ProcessedDepositOperation, *utils.CustomError, error) {
	txInfo, err := provider.GetTxInfo(operation.DepositTxHash)
	if err != nil {
		return nil, nil, err
	}

	latestBlockNumber, err := provider.GetLatestBlockNumber()
	if err != nil {
		return nil, nil, err
	}

	if txInfo.TxReceipt.BlockNumber.Int64()+int64(provider.MinConfirmations) <
		int64(latestBlockNumber) {
		return nil, utils.NewCustomError("transaction not confirmed", false), nil
	}

	var tokenAddress string
	var amount int64
	var txTargetAddress string

	if len(txInfo.Input) == 0 {
		// Native transfer
		tokenAddress = "native"
		amount, err = utils.StringToInt64(txInfo.Amount)
		if err != nil {
			return nil, nil, err
		}
		txTargetAddress = txInfo.To
	} else {
		// ERC20 transfer
		transfers := evm_utils.FindERC20Transfers(txInfo.TxReceipt)
		if len(transfers) == 0 {
			return nil, utils.NewCustomError("no ERC20 transfer found", false), nil
		}
		tokenAddress = transfers[0].Token.Hex()
		amount = transfers[0].Value.Int64()
		txTargetAddress = transfers[0].To.Hex()

		tokenAddressInfo := chainsCache.GetTokenByChainNameAndTokenAddress(
			operation.TargetChainName,
			tokenAddress,
		)
		if tokenAddressInfo == nil {
			return nil, utils.NewCustomError("token not found", false), nil
		}
	}

	if txTargetAddress != operation.TargetAddress {
		return nil, utils.NewCustomError(
			"invalid target address, expected: "+operation.TargetAddress+
				", got: "+txTargetAddress,
			false,
		), nil
	}

	return &ProcessedDepositOperation{
		TokenAddress: tokenAddress,
		Amount:       amount,
	}, nil, nil
}

func processBTCDepositOperation(
	operation *queue.DepositOperationEvent,
) (*ProcessedDepositOperation, *utils.CustomError, error) {
	return nil, nil, nil
}

func processSOLDepositOperation(
	operation *queue.DepositOperationEvent,
) (*ProcessedDepositOperation, *utils.CustomError, error) {
	return nil, nil, nil
}

func ProcessDepositOperation(
	providerPool *provider.ProviderPool,
	chainsCache *memorycache.ChainsCache,
	operation *queue.DepositOperationEvent,
) (*ProcessedDepositOperation, *utils.CustomError, error) {
	chainPlatform := chainsCache.GetPlatformByChainName(
		operation.TargetChainName,
	)
	switch chainPlatform {
	case system.ChainPlatformEVM:
		evmProvider := providerPool.GetEVMProvider(operation.TargetChainName)
		return processEVMDepositOperation(chainsCache, evmProvider, operation)
	case system.ChainPlatformBTC:
		return processBTCDepositOperation(operation)
	case system.ChainPlatformSOL:
		return processSOLDepositOperation(operation)
	}
	return nil, nil, nil
}

func ProcessWithdrawOperation(
	operation *queue.WithdrawOperationEvent,
) error {
	return nil
}
