package transaction_service

import (
	errors "errors"
	fmt "fmt"

	system "deposit-collector/internal/system"
	walletservices "deposit-collector/internal/wallet_services"
	btc_utils "deposit-collector/pkg/crypto/btc"
	provider "deposit-collector/pkg/crypto/provider"
	wallet "deposit-collector/pkg/crypto/wallet"
	utils "deposit-collector/pkg/utils"

	uuid "github.com/google/uuid"
)

type CollectUnprocessedDepositsResult struct {
	TxHash       string
	OperationIDs []uuid.UUID
}

func collectEVMUnprocessedDeposits(
	tokenAddress system.TokenAddress,
	destinationDepositAddress string,
	repository TransactionRepository,
	provider provider.EVMProvider,
	walletServices walletservices.WalletServices,
) (*CollectUnprocessedDepositsResult, error) {
	operations, err := repository.GetGroupedUnprocessedDepositsByTokenAddressID(
		tokenAddress.TokenAddressDbID,
		1,
	)
	if err != nil {
		return nil, err
	}
	for _, operation := range operations {
		fmt.Println(operation)
	}
	return nil, nil
}

func collectBTCUnprocessedDeposits(
	tokenAddress system.TokenAddress,
	destinationDepositAddress string,
	repository TransactionRepository,
	provider provider.BitcoinProvider,
	walletServices walletservices.WalletServices,
) (*CollectUnprocessedDepositsResult, error) {
	// @TODO: Validate the unspent outputs of the user,
	// so we wont collect deposits that are already spent
	operations, err := repository.GetUnprocessedDepositsByTokenAddressID(
		tokenAddress.TokenAddressDbID,
		// Up to 10 inputs per transaction
		10,
	)

	if err != nil {
		return nil, err
	}
	if len(operations) == 0 {
		return nil, nil
	}

	network := btc_utils.GetNetParamsByNetwork(provider.Network)

	if network == nil {
		return nil, errors.New("invalid bitcoin network")
	}

	totalAmountInSats := int64(0)
	inputs := make([]btc_utils.UTXOInput, 0)
	signers := make([]walletservices.BitcoinSigner, 0)
	operationIDs := make([]uuid.UUID, 0)
	// 1. Iterate over the operations and get the tx info
	for _, operation := range operations {
		txInfo, err := provider.GetTxInfo(operation.TxHash)
		if err != nil {
			return nil, err
		}
		// 2. Iterate over each output of the transaction
		for _, vout := range txInfo.Vout {
			addressFromScript, err := btc_utils.GetAddressFromScript(
				vout.ScriptPubKey.Hex, network,
			)
			if err != nil {
				return nil, err
			}
			if addressFromScript == operation.Address {
				// 3. Generate a new wallet for the operation
				generatedWallet, err := walletServices.GenerateWallet(
					operation.AccountID,
					0,
					operation.SequenceNumber,
					provider.Network,
					system.ChainPlatformBTC,
				)
				if err != nil {
					return nil, err
				}
				btcWallet, ok := generatedWallet.(*wallet.BitcoinWallet)
				if !ok {
					return nil, errors.New("invalid wallet type for bitcoin operation")
				}
				if btcWallet.GetAddress() != operation.Address {
					return nil, errors.New("wallet address does not match operation address")
				}
				amountSats, err := btc_utils.BitcoinToSatoshis(vout.Value)
				if err != nil {
					return nil, err
				}
				// 4. Add the input to the inputs array
				inputs = append(inputs, btc_utils.UTXOInput{
					TxHash:     operation.TxHash,
					Vout:       uint32(vout.N),
					AmountSats: amountSats,
					Sequence:   btc_utils.DefaultInputSequence,
				})
				inputIndex := len(inputs) - 1
				signers = append(signers, walletservices.BitcoinSigner{
					Wallet: *btcWallet,
					Inputs: []wallet.BitcoinTransactionInput{
						{
							Index:      inputIndex,
							AmountSats: amountSats,
						},
					},
				})
				totalAmountInSats += amountSats
				operationIDs = append(operationIDs, operation.ID)
			}
		}
	}

	if len(inputs) == 0 {
		return nil, errors.New("no inputs found")
	}

	// 5. Calculate the fee for the transaction
	minFeePerKB, err := provider.GetMinFeePerKB(3)
	if err != nil {
		return nil, err
	}
	minFeePerKbInSats, err := btc_utils.BitcoinToSatoshis(minFeePerKB)
	if err != nil {
		return nil, err
	}

	if minFeePerKbInSats < btc_utils.MIN_FEE_PER_KB_IN_SATS {
		minFeePerKbInSats = btc_utils.MIN_FEE_PER_KB_IN_SATS
	}
	satPerVByte := float64(minFeePerKbInSats) / 1000
	totalTxFeeInSats, err := btc_utils.CalculateFee(
		btc_utils.CalculateFeeRequest{
			InputCount:         len(inputs),
			OutputCount:        1,
			SignaturesPerInput: 1,
			FeeRateSatPerVByte: satPerVByte,
			IncludeChange:      false,
		},
	)
	if err != nil {
		return nil, err
	}

	// 6. Create the outputs for the transaction
	destinationAmountInSats := totalAmountInSats - totalTxFeeInSats
	if destinationAmountInSats <= 0 {
		return nil, errors.New("destination amount is less than or equal to 0")
	}
	outputs := make([]btc_utils.TxOutput, 0)
	outputs = append(outputs, btc_utils.TxOutput{
		Address:    destinationDepositAddress,
		AmountSats: destinationAmountInSats,
	})
	createTransactionRequest := btc_utils.CreateTransactionRequest{
		Network: provider.Network,
		Inputs:  inputs,
		Outputs: outputs,
		FeeSats: totalTxFeeInSats,
	}

	// 7. Create the transaction
	createTransactionResult, err := btc_utils.CreateTransaction(
		createTransactionRequest,
	)
	if err != nil {
		return nil, err
	}

	// 8. Sign the transaction
	signedTxHex, err := walletservices.SignTransactionWithInputs(
		provider.Network,
		createTransactionResult.RawHex,
		signers,
	)
	if err != nil {
		return nil, err
	}

	// 9. Broadcast the transaction
	txHash, err := provider.BroadcastTransaction(signedTxHex)
	if err != nil {
		return nil, err
	}

	return &CollectUnprocessedDepositsResult{
		TxHash:       txHash,
		OperationIDs: operationIDs,
	}, nil
}

func collectSOLUnprocessedDeposits(
	tokenAddress system.TokenAddress,
	repository TransactionRepository,
) (*CollectUnprocessedDepositsResult, error) {
	return nil, nil
}

func CollectUnprocessedDeposits(
	chain system.SupportedChain,
	destinationDepositAddresses DestinationDepositAddress,
	tokenAddress system.TokenAddress,
	providerPool *provider.ProviderPool,
	repository *TransactionRepository,
	walletServices *walletservices.WalletServices,
) (*CollectUnprocessedDepositsResult, error) {
	switch chain.ChainPlatform {
	case system.ChainPlatformEVM:
		evmProvider := providerPool.GetEVMProvider(chain.ChainName)
		if evmProvider == nil {
			return nil, utils.NewCustomError("provider not found", false)
		}
		destinationDepositAddress := destinationDepositAddresses[chain.ChainPlatform]
		return collectEVMUnprocessedDeposits(
			tokenAddress, destinationDepositAddress, *repository,
			*evmProvider, *walletServices,
		)
	case system.ChainPlatformBTC:
		btcProvider := providerPool.GetBitcoinProvider()
		if btcProvider == nil {
			return nil, utils.NewCustomError("provider not found", false)
		}
		destinationDepositAddress := destinationDepositAddresses[chain.ChainPlatform]
		return collectBTCUnprocessedDeposits(
			tokenAddress, destinationDepositAddress,
			*repository, *btcProvider, *walletServices,
		)
	case system.ChainPlatformSOL:
		return collectSOLUnprocessedDeposits(tokenAddress, *repository)
	}
	return nil, nil
}

type CollectUnprocessedWithdrawalsResult struct {
	TxHash       string
	OperationIDs []uuid.UUID
}

func collectEVMUnprocessedWithdrawals(
	tokenAddress system.TokenAddress,
	repository TransactionRepository,
	provider provider.EVMProvider,
	walletServices walletservices.WalletServices,
) (*CollectUnprocessedWithdrawalsResult, error) {
	return nil, nil
}

func collectBTCUnprocessedWithdrawals(
	tokenAddress system.TokenAddress,
	repository TransactionRepository,
	provider provider.BitcoinProvider,
	walletServices walletservices.WalletServices,
) (*CollectUnprocessedWithdrawalsResult, error) {
	return nil, nil
}

func collectSOLUnprocessedWithdrawals(
	tokenAddress system.TokenAddress,
	repository TransactionRepository,
) (*CollectUnprocessedWithdrawalsResult, error) {
	return nil, nil
}

func CollectUnprocessedWithdrawals(
	chain system.SupportedChain,
	tokenAddress system.TokenAddress,
	providerPool *provider.ProviderPool,
	repository *TransactionRepository,
	walletServices *walletservices.WalletServices,
) (*CollectUnprocessedWithdrawalsResult, error) {
	switch chain.ChainPlatform {
	case system.ChainPlatformEVM:
		evmProvider := providerPool.GetEVMProvider(chain.ChainName)
		if evmProvider == nil {
			return nil, utils.NewCustomError("provider not found", false)
		}
		return collectEVMUnprocessedWithdrawals(
			tokenAddress, *repository, *evmProvider, *walletServices,
		)
	case system.ChainPlatformBTC:
		btcProvider := providerPool.GetBitcoinProvider()
		if btcProvider == nil {
			return nil, utils.NewCustomError("provider not found", false)
		}
		return collectBTCUnprocessedWithdrawals(
			tokenAddress, *repository, *btcProvider, *walletServices,
		)
	case system.ChainPlatformSOL:
		return collectSOLUnprocessedWithdrawals(tokenAddress, *repository)
	}
	return nil, nil
}
