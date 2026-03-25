package provider

import (
	context "context"

	logger "deposit-collector/pkg/logger"
	utils "deposit-collector/pkg/utils"

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
		To:        txInfo.To().Hex(),
		Amount:    txInfo.Value().String(),
		Input:     txInfo.Data(),
		ChainID:   txInfo.ChainId().String(),
		Timestamp: txInfo.Time().String(),
		TxReceipt: tx,
	}, nil
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
