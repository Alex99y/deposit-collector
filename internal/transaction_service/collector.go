package transaction_service

import (
	errors "errors"
	math "math/big"

	system "deposit-collector/internal/system"
	walletservices "deposit-collector/internal/wallet_services"
	btc_utils "deposit-collector/pkg/crypto/btc"
	evm_utils "deposit-collector/pkg/crypto/evm"
	provider "deposit-collector/pkg/crypto/provider"
	wallet "deposit-collector/pkg/crypto/wallet"
	utils "deposit-collector/pkg/utils"

	common "github.com/ethereum/go-ethereum/common"
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

	if len(operations) == 0 {
		return nil, nil
	}

	operation := operations[0]

	switch tokenAddress.Address {
	case "native":
		// Native ETH sweep (EOA -> master).
	default:
		// ERC20 sweep is not implemented yet.
		return nil, nil
	}

	// 1) Generate the EOA wallet that controls `operation.Address`.
	// NOTE: the changeIndex is set to 0 for now.
	generatedWallet, err := walletServices.GenerateWallet(
		operation.AccountID,
		0,
		operation.SequenceNumber,
		btc_utils.MAINNET,
		system.ChainPlatformEVM,
	)
	if err != nil {
		return nil, err
	}

	evmWallet, ok := generatedWallet.(*wallet.EvmWallet)
	if !ok {
		return nil, errors.New("invalid wallet type for evm operation")
	}
	if evmWallet.GetAddress() != operation.Address {
		return nil, errors.New("wallet address does not match operation address")
	}

	// 2) TODO: check balance before sending/creating the txn.
	valueWei := math.NewInt(operation.Amount)
	if valueWei.Sign() <= 0 {
		return nil, errors.New("deposit amount must be > 0")
	}

	// 3) Send a normal EVM native transfer (EIP-1559).
	fromAddress := common.HexToAddress(evmWallet.GetAddress())
	toAddress := common.HexToAddress(destinationDepositAddress)

	nonce, err := provider.GetPendingNonce(fromAddress)
	if err != nil {
		return nil, err
	}

	tipCap, err := provider.SuggestGasTipCap()
	if err != nil {
		return nil, err
	}

	baseFeePerGas, err := provider.GetLatestBaseFeePerGas()
	if err != nil {
		return nil, err
	}

	gasLimit, err := provider.EstimateNativeGas(fromAddress, toAddress, valueWei)
	if err != nil {
		return nil, err
	}

	unsignedTx, err := evm_utils.BuildNativeTransferEIP1559Tx(
		provider.ChainID,
		fromAddress,
		nonce,
		toAddress,
		valueWei,
		gasLimit,
		tipCap,
		baseFeePerGas,
	)
	if err != nil {
		return nil, err
	}

	signedTxs, err := walletservices.SignEVMTransactions(
		[]walletservices.EVMSigner{
			{
				Wallet: *evmWallet,
				Txs: []*evm_utils.UnsignedEIP1559Tx{
					unsignedTx,
				},
			},
		})
	if err != nil {
		return nil, err
	}
	if len(signedTxs) == 0 {
		return nil, errors.New("no signed transactions generated")
	}

	txHash, err := provider.BroadcastSignedTransaction(signedTxs[0])
	if err != nil {
		return nil, err
	}

	// 4) Wait for at least 5 block confirmations.
	const confirmationsToWait uint64 = 5
	if err := provider.WaitForConfirmations(
		txHash, confirmationsToWait,
	); err != nil {
		return nil, err
	}

	return &CollectUnprocessedDepositsResult{
		TxHash:       txHash,
		OperationIDs: operation.OperationIDs,
	}, nil
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
		return nil, errors.New("SOL withdraw collector is not implemented yet")
	}
	return nil, errors.New("chain platform not supported")
}

type CollectUnprocessedWithdrawalsResult struct {
	TxHash       string
	OperationIDs []uuid.UUID
}

func collectEVMUnprocessedWithdrawals(
	tokenAddress system.TokenAddress,
	privateKey string,
	repository TransactionRepository,
	provider provider.EVMProvider,
) (*CollectUnprocessedWithdrawalsResult, error) {
	return nil, nil
}

func collectBTCUnprocessedWithdrawals(
	tokenAddress system.TokenAddress,
	privateKey string,
	repository TransactionRepository,
	provider provider.BitcoinProvider,
) (*CollectUnprocessedWithdrawalsResult, error) {
	if tokenAddress.Address != "native" {
		return nil, nil
	}

	signing, err := btc_utils.BitcoinWalletFromWithdrawCollectorKey(
		privateKey, provider.Network,
	)
	if err != nil {
		return nil, err
	}
	signerWallet := &wallet.BitcoinWallet{
		Address: signing.Address,
		WIF:     signing.WIF,
	}
	signerAddress := signing.Address

	confirmedBal, unconfirmedBal, err := provider.GetAddressBalanceSatoshis(
		signerAddress,
	)
	if err != nil {
		return nil, err
	}
	totalReported := confirmedBal + unconfirmedBal

	electrumUTXOs, err := provider.ListUnspentByAddress(signerAddress)
	if err != nil {
		return nil, err
	}
	if len(electrumUTXOs) == 0 {
		return nil, errors.New("no UTXOs available for signer address")
	}

	var utxoSum int64
	for _, u := range electrumUTXOs {
		utxoSum += u.Value
	}
	if utxoSum > totalReported {
		return nil, errors.New("UTXO sum exceeds reported script balance")
	}

	withdrawals, err := repository.GetUnprocessedWithdrawalsByTokenAddressID(
		tokenAddress.TokenAddressDbID,
		10,
	)
	if err != nil {
		return nil, err
	}
	if len(withdrawals) == 0 {
		return nil, nil
	}

	outputs := make([]btc_utils.TxOutput, 0, len(withdrawals))
	operationIDs := make([]uuid.UUID, 0, len(withdrawals))
	var totalWithdrawSats int64
	for _, w := range withdrawals {
		outputs = append(outputs, btc_utils.TxOutput{
			Address:    w.DestinationAddress,
			AmountSats: w.Amount,
		})
		operationIDs = append(operationIDs, w.ID)
		totalWithdrawSats += w.Amount
	}

	if totalReported < totalWithdrawSats {
		return nil, errors.New("on-chain balance is less than sum of withdrawals")
	}

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

	electrumEntries := make(
		[]btc_utils.ElectrumListUnspentEntry, len(electrumUTXOs),
	)
	for i, u := range electrumUTXOs {
		electrumEntries[i] = btc_utils.ElectrumListUnspentEntry{
			TxHash: u.TxHash,
			TxPos:  u.TxPos,
			Value:  u.Value,
		}
	}
	inputs, feeSats, changeAddress, err :=
		btc_utils.SelectBTCInputsAndFeeForWithdrawals(
			electrumEntries,
			outputs,
			satPerVByte,
			signerAddress,
			totalWithdrawSats,
		)
	if err != nil {
		return nil, err
	}

	totalIn := int64(0)
	for _, in := range inputs {
		totalIn += in.AmountSats
	}
	if totalIn < totalWithdrawSats+feeSats {
		return nil, errors.New("insufficient funds after fee and outputs")
	}
	if totalReported < totalWithdrawSats+feeSats {
		return nil, errors.New("reported balance does not cover withdrawals and fee")
	}

	createReq := btc_utils.CreateTransactionRequest{
		Network:       provider.Network,
		Inputs:        inputs,
		Outputs:       outputs,
		FeeSats:       feeSats,
		ChangeAddress: changeAddress,
	}

	created, err := btc_utils.CreateTransaction(createReq)
	if err != nil {
		return nil, err
	}

	signersInputs := make([]wallet.BitcoinTransactionInput, len(inputs))
	for i := range inputs {
		signersInputs[i] = wallet.BitcoinTransactionInput{
			Index:      i,
			AmountSats: inputs[i].AmountSats,
		}
	}
	signedTxHex, err := walletservices.SignTransactionWithInputs(
		provider.Network,
		created.RawHex,
		[]walletservices.BitcoinSigner{{
			Wallet: *signerWallet,
			Inputs: signersInputs,
		}},
	)
	if err != nil {
		return nil, err
	}

	txHash, err := provider.BroadcastTransaction(signedTxHex)
	if err != nil {
		return nil, err
	}

	// TODO: This process must be split into two steps:
	// 1. Calculate the txHash before broadcasting the transaction
	// 	a. Local txid: btc_utils.BitcoinTxIDFromSerializedHex(signedTxHex)
	// 	b. Broadcast the transaction and get the txHash
	//  c. Validate the txHash
	//  d. Store the txHash in the database
	// 2. In another process, we will confirm that the tx was committed
	// 	to the blockchain
	//  a. Update the processed_at with the current timestamp
	//  b. In case that the tx was not committed,
	//    we will remove the txHash from the registry
	if err := repository.MarkWithdrawalOperationAsProcessed(
		operationIDs, txHash,
	); err != nil {
		return nil, err
	}

	return &CollectUnprocessedWithdrawalsResult{
		TxHash:       txHash,
		OperationIDs: operationIDs,
	}, nil
}

func CollectUnprocessedWithdrawals(
	chain system.SupportedChain,
	privateKey string,
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
			tokenAddress, privateKey, *repository, *evmProvider,
		)
	case system.ChainPlatformBTC:
		btcProvider := providerPool.GetBitcoinProvider()
		if btcProvider == nil {
			return nil, utils.NewCustomError("provider not found", false)
		}
		return collectBTCUnprocessedWithdrawals(
			tokenAddress, privateKey, *repository, *btcProvider,
		)
	case system.ChainPlatformSOL:
		return nil, errors.New("SOL withdraw collector is not implemented yet")
	}
	return nil, errors.New("chain platform not supported")
}
