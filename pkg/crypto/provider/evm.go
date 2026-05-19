package provider

import (
	context "context"
	errors "errors"
	math "math/big"
	time "time"

	logger "deposit-collector/pkg/logger"
	utils "deposit-collector/pkg/utils"

	ethereum "github.com/ethereum/go-ethereum"
	common "github.com/ethereum/go-ethereum/common"
	types "github.com/ethereum/go-ethereum/core/types"
	ethclient "github.com/ethereum/go-ethereum/ethclient"
)

type EVMProvider struct {
	client           *ethclient.Client
	context          context.Context
	ChainID          int
	MinConfirmations int
}

type EvmTxInfo struct {
	From      string
	To        string
	Amount    string
	ChainID   string
	TxHash    string
	Timestamp string
	Input     []byte
	TxReceipt *types.Receipt
}

func transactionRecipientHex(tx *types.Transaction) string {
	if tx == nil || tx.To() == nil {
		return ""
	}
	return tx.To().Hex()
}

func confirmedSuccessfulReceipt(
	receipt *types.Receipt,
	latestBlock uint64,
	confirmations uint64,
) (bool, error) {
	if receipt == nil || receipt.BlockNumber == nil {
		return false, nil
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return false, errors.New("evm transaction failed on-chain")
	}

	receiptBlock := receipt.BlockNumber.Uint64()
	if latestBlock < receiptBlock {
		return false, nil
	}
	return latestBlock-receiptBlock >= confirmations, nil
}

func (p *EVMProvider) GetLatestBlockNumber() (uint64, error) {
	blockNumber, err := p.client.BlockNumber(p.context)
	if err != nil {
		return 0, err
	}
	return blockNumber, nil
}

func (p *EVMProvider) GetTxInfo(txHash string) (*EvmTxInfo, error) {
	tx, err := p.client.TransactionReceipt(p.context, common.HexToHash(txHash))
	if err != nil {
		return nil, err
	}

	txInfo, _, err := p.client.TransactionByHash(p.context, tx.TxHash)
	if err != nil {
		return nil, err
	}

	signer := types.LatestSignerForChainID(txInfo.ChainId())
	from, err := signer.Sender(txInfo)
	if err != nil {
		return nil, err
	}

	return &EvmTxInfo{
		TxHash:    txHash,
		From:      from.Hex(),
		To:        transactionRecipientHex(txInfo),
		Amount:    txInfo.Value().String(),
		Input:     txInfo.Data(),
		ChainID:   txInfo.ChainId().String(),
		Timestamp: txInfo.Time().String(),
		TxReceipt: tx,
	}, nil
}

func (p *EVMProvider) GetPendingNonce(address common.Address) (uint64, error) {
	return p.client.PendingNonceAt(p.context, address)
}

func (p *EVMProvider) SuggestGasTipCap() (*math.Int, error) {
	return p.client.SuggestGasTipCap(p.context)
}

func (p *EVMProvider) GetLatestBaseFeePerGas() (*math.Int, error) {
	header, err := p.client.HeaderByNumber(p.context, nil)
	if err != nil {
		return nil, err
	}
	if header.BaseFee == nil {
		return nil, errors.New("chain does not appear to support EIP-1559 (missing baseFee)")
	}
	return header.BaseFee, nil
}

func (p *EVMProvider) EstimateNativeGas(from common.Address, to common.Address, valueWei *math.Int) (uint64, error) {
	if valueWei == nil || valueWei.Sign() <= 0 {
		return 0, errors.New("valueWei must be > 0")
	}
	return p.client.EstimateGas(
		p.context,
		ethereum.CallMsg{
			From:  from,
			To:    &to,
			Value: valueWei,
		},
	)
}

// EstimateContractGas estimates gas for a contract call (e.g. ERC-20 transfer).
func (p *EVMProvider) EstimateContractGas(
	from common.Address,
	contract common.Address,
	valueWei *math.Int,
	data []byte,
) (uint64, error) {
	msg := ethereum.CallMsg{
		From: from,
		To:   &contract,
		Data: data,
	}
	if valueWei != nil {
		msg.Value = valueWei
	}
	return p.client.EstimateGas(p.context, msg)
}

func (p *EVMProvider) BroadcastSignedTransaction(signedTx *types.Transaction) (string, error) {
	if signedTx == nil {
		return "", errors.New("signedTx is required")
	}
	if err := p.client.SendTransaction(p.context, signedTx); err != nil {
		return "", err
	}
	return signedTx.Hash().Hex(), nil
}

func (p *EVMProvider) WaitForConfirmations(txHash string, confirmations uint64) error {
	if confirmations == 0 {
		return nil
	}

	for {
		latestBlock, err := p.GetLatestBlockNumber()
		if err != nil {
			return err
		}

		txInfo, err := p.GetTxInfo(txHash)
		if err == nil && txInfo != nil && txInfo.TxReceipt != nil {
			confirmed, err := confirmedSuccessfulReceipt(
				txInfo.TxReceipt,
				latestBlock,
				confirmations,
			)
			if err != nil {
				return err
			}
			if confirmed {
				return nil
			}
		}

		time.Sleep(3 * time.Second)
	}
}

func NewEVMProvider(
	url string,
	chainID int,
	minConfirmations int,
	context context.Context,
	logger *logger.Logger,
) *EVMProvider {
	client, err := ethclient.Dial(url)
	if err != nil {
		utils.FailOnError(logger, err, "Error creating EVM provider")
	}
	return &EVMProvider{
		client:           client,
		context:          context,
		ChainID:          chainID,
		MinConfirmations: minConfirmations,
	}
}
