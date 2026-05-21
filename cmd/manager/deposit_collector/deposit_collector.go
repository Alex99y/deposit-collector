package deposit_collector

import (
	context "context"
	errors "errors"
	fmt "fmt"

	system "deposit-collector/internal/system"
	transaction_service "deposit-collector/internal/transaction_service"
	walletservices "deposit-collector/internal/wallet_services"
	provider "deposit-collector/pkg/crypto/provider"
	logger "deposit-collector/pkg/logger"
	worker "deposit-collector/pkg/worker"
)

type DepositCollectorWorker struct {
	chain                       system.SupportedChain
	tokens                      []system.TokenAddress
	destinationDepositAddresses transaction_service.DestinationDepositAddress
	transactionRepository       *transaction_service.TransactionRepository
	providerPool                *provider.ProviderPool
	walletServices              *walletservices.WalletServices
	logger                      *logger.Logger
}

func (w *DepositCollectorWorker) ProcessSetledDeposits() {
	for _, token := range w.tokens {
		result, err := transaction_service.CollectUnprocessedDeposits(
			w.chain,
			w.destinationDepositAddresses,
			token,
			w.providerPool,
			w.transactionRepository,
			w.walletServices,
		)
		if err != nil {
			w.logger.Error(
				fmt.Sprintf(
					"error collecting deposits for token %s in chain %s: %s",
					token.Address, w.chain.ChainName, err.Error(),
				),
			)
			w.logger.ErrorO(err)
			continue
		}
		if result == nil {
			continue
		}
		if result.TxHash != "" {
			w.logger.Info(
				fmt.Sprintf(
					"collected deposit for token %s with tx hash %s in chain %s",
					token.Address, result.TxHash, w.chain.ChainName,
				),
			)

			err = w.transactionRepository.MarkDepositOperationAsProcessed(
				result.OperationIDs,
			)
			if err != nil {
				w.logger.ErrorO(err)
			}
		}
	}
}

type DepositCollector struct {
	ctx                         context.Context
	logger                      *logger.Logger
	destinationDepositAddresses transaction_service.DestinationDepositAddress
	systemRepository            *system.SystemRepository
	transactionRepository       *transaction_service.TransactionRepository
	walletServices              *walletservices.WalletServices
	providerPool                *provider.ProviderPool
	workers                     []*worker.Worker
}

func (dc *DepositCollector) Start() error {
	chains, err := dc.systemRepository.GetSupportedChains()
	if err != nil {
		return err
	}

	for index := range chains {
		chain := chains[index]
		tokens, err := dc.systemRepository.GetTokenAddresses(
			system.GetTokenAddressesRequest{
				Chain: &chain.ChainName,
				Limit: 10000,
			},
		)
		if err != nil {
			return err
		}

		depositCollectorWorker := DepositCollectorWorker{
			chain:                       chain,
			tokens:                      tokens,
			destinationDepositAddresses: dc.destinationDepositAddresses,
			transactionRepository:       dc.transactionRepository,
			walletServices:              dc.walletServices,
			providerPool:                dc.providerPool,
			logger:                      dc.logger,
		}

		worker := worker.NewWorker(
			fmt.Sprintf("deposit-collector-%s-worker", chain.ChainName),
			dc.ctx,
			dc.logger,
			depositCollectorWorker.ProcessSetledDeposits,
			60,
		)

		err = worker.Start()
		if err != nil {
			return err
		}
		dc.workers = append(dc.workers, worker)
	}

	return nil
}

func (dc *DepositCollector) Stop() {
	for index := range dc.workers {
		err := dc.workers[index].Stop()
		if err != nil {
			dc.logger.ErrorO(err)
		}
	}
}

func NewDepositCollector(
	ctx context.Context,
	providerPool *provider.ProviderPool,
	destinationDepositAddresses transaction_service.DestinationDepositAddress,
	systemRepository *system.SystemRepository,
	transactionRepository *transaction_service.TransactionRepository,
	walletServices *walletservices.WalletServices,
	logger *logger.Logger,
) (*DepositCollector, error) {
	if systemRepository == nil {
		return nil, errors.New("invalid repository provided")
	}

	return &DepositCollector{
		ctx:                         ctx,
		systemRepository:            systemRepository,
		destinationDepositAddresses: destinationDepositAddresses,
		logger:                      logger,
		providerPool:                providerPool,
		transactionRepository:       transactionRepository,
		walletServices:              walletServices,
		workers:                     make([]*worker.Worker, 0),
	}, nil
}
