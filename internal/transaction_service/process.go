package transaction_service

import (
	errors "errors"

	memorycache "deposit-collector/internal/memory_cache"
	queue "deposit-collector/internal/queue"
	system "deposit-collector/internal/system"
	btc_utils "deposit-collector/pkg/crypto/btc"
	evm_utils "deposit-collector/pkg/crypto/evm"
	provider "deposit-collector/pkg/crypto/provider"
	utils "deposit-collector/pkg/utils"

	types "github.com/ethereum/go-ethereum/core/types"
)

type ProcessedDepositOperation struct {
	TokenAddress string
	Amount       int64
}

func processEVMDepositOperation(
	chainsCache *memorycache.ChainsCache,
	provider *provider.EVMProvider,
	operation *queue.DepositOperationEvent,
) (*ProcessedDepositOperation, error) {
	txInfo, err := provider.GetTxInfo(operation.DepositTxHash)
	if err != nil {
		return nil, err
	}

	latestBlockNumber, err := provider.GetLatestBlockNumber()
	if err != nil {
		return nil, err
	}

	return processEVMDepositTxInfo(
		txInfo,
		latestBlockNumber,
		provider.MinConfirmations,
		operation,
		chainsCache.GetTokenByChainNameAndTokenAddress,
	)
}

func processEVMDepositTxInfo(
	txInfo *provider.EvmTxInfo,
	latestBlockNumber uint64,
	minConfirmations int,
	operation *queue.DepositOperationEvent,
	lookupToken func(chainName string, tokenAddress string) *system.TokenAddress,
) (*ProcessedDepositOperation, error) {
	if txInfo.TxReceipt == nil || txInfo.TxReceipt.BlockNumber == nil {
		return nil, utils.NewCustomError("transaction not confirmed", false)
	}
	if txInfo.TxReceipt.Status != types.ReceiptStatusSuccessful {
		return nil, utils.NewCustomError("transaction failed on-chain", false)
	}

	receiptBlock := txInfo.TxReceipt.BlockNumber.Uint64()
	if latestBlockNumber < receiptBlock ||
		latestBlockNumber-receiptBlock < uint64(minConfirmations) {
		return nil, utils.NewCustomError("transaction not confirmed", false)
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
			return nil, utils.NewCustomError("no ERC20 transfer found", false)
		}
		tokenAddress = transfers[0].Token.Hex()
		if !transfers[0].Value.IsInt64() {
			return nil, utils.NewCustomError("ERC20 transfer amount exceeds int64", false)
		}
		amount = transfers[0].Value.Int64()
		txTargetAddress = transfers[0].To.Hex()

		tokenAddressInfo := lookupToken(
			operation.TargetChainName,
			tokenAddress,
		)
		if tokenAddressInfo == nil {
			return nil, utils.NewCustomError("token not found", false)
		}
	}

	if txTargetAddress != operation.TargetAddress {
		return nil, utils.NewCustomError(
			"invalid target address, expected: "+operation.TargetAddress+
				", got: "+txTargetAddress, false)
	}

	return &ProcessedDepositOperation{
		TokenAddress: tokenAddress,
		Amount:       amount,
	}, nil
}

func processBTCDepositOperation(
	provider provider.BitcoinProvider,
	operation *queue.DepositOperationEvent,
) (*ProcessedDepositOperation, error) {
	txInfo, err := provider.GetTxInfo(operation.DepositTxHash)

	if err != nil {
		return nil, err
	}

	confirmations := txInfo.Confirmations

	if confirmations <= provider.MinConfirmations {
		return nil, utils.NewCustomError("transaction not confirmed", false)
	}

	amount := int64(0)

	network := btc_utils.GetNetParamsByNetwork(provider.Network)

	if network == nil {
		return nil, errors.New("invalid bitcoin network")
	}

	for _, vout := range txInfo.Vout {
		addressFromScript, err := btc_utils.GetAddressFromScript(
			vout.ScriptPubKey.Hex, network,
		)
		if err != nil {
			return nil, err
		}
		if addressFromScript == operation.TargetAddress {
			satoshis, err := btc_utils.BitcoinToSatoshis(vout.Value)
			if err != nil {
				return nil, err
			}
			amount += satoshis
		}
	}

	// @TODO: check here min deposit value, so we wont process dust amount
	if amount < 10000 {
		return nil, errors.New("invalid received amount")
	}

	return &ProcessedDepositOperation{
		TokenAddress: "native",
		Amount:       amount,
	}, nil
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
		if evmProvider == nil {
			return nil, utils.NewCustomError("provider not found", false)
		}
		return processEVMDepositOperation(chainsCache, evmProvider, operation)
	case system.ChainPlatformBTC:
		btcProvider := providerPool.GetBitcoinProvider()
		if btcProvider == nil {
			return nil, utils.NewCustomError("provider not found", false)
		}
		return processBTCDepositOperation(*btcProvider, operation)
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
