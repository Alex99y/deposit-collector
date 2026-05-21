package withdraw_collector

import (
	context "context"
	errors "errors"
	fmt "fmt"
	strings "strings"

	system "deposit-collector/internal/system"
	transaction_service "deposit-collector/internal/transaction_service"
	walletservices "deposit-collector/internal/wallet_services"
	provider "deposit-collector/pkg/crypto/provider"
	logger "deposit-collector/pkg/logger"
	worker "deposit-collector/pkg/worker"
)

type WithdrawCollectorWorker struct {
	chain                 system.SupportedChain
	tokens                []system.TokenAddress
	privateKey            string
	transactionRepository *transaction_service.TransactionRepository
	providerPool          *provider.ProviderPool
	walletServices        *walletservices.WalletServices
	logger                *logger.Logger
}

func (w *WithdrawCollectorWorker) ProcessWithdrawals() {
	for _, token := range w.tokens {
		result, err := transaction_service.CollectUnprocessedWithdrawals(
			w.chain,
			w.privateKey,
			token,
			w.providerPool,
			w.transactionRepository,
			w.walletServices,
		)
		if err != nil {
			w.logger.Error(
				fmt.Sprintf("error collecting withdrawals for token %s in chain %s: %s",
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
				fmt.Sprintf("collected withdrawal for token %s with tx hash %s in chain %s",
					token.Address, result.TxHash, w.chain.ChainName,
				),
			)
			err = w.transactionRepository.MarkWithdrawalOperationAsProcessed(
				result.OperationIDs,
				result.TxHash,
			)
			if err != nil {
				w.logger.ErrorO(err)
			}
		}
	}
}

type WithdrawCollector struct {
	ctx                   context.Context
	logger                *logger.Logger
	collectorPrivateKeys  transaction_service.WithdrawCollectorPrivateKeys
	transactionRepository *transaction_service.TransactionRepository
	providerPool          *provider.ProviderPool
	walletServices        *walletservices.WalletServices
	systemRepository      *system.SystemRepository
	workers               []*worker.Worker
}

func (w *WithdrawCollector) Start() error {
	if w.systemRepository == nil {
		return errors.New("invalid system repository provided")
	}

	chains, err := w.systemRepository.GetSupportedChains()
	if err != nil {
		return err
	}

	for index := range chains {
		chain := chains[index]
		privateKey := strings.TrimSpace(w.collectorPrivateKeys[chain.ChainPlatform])
		if privateKey == "" {
			w.logger.Warn(
				fmt.Sprintf(
					"withdraw collector private key not configured for chain platform %s; skipping chain %s",
					chain.ChainPlatform,
					chain.ChainName,
				),
			)
			continue
		}

		tokens, err := w.systemRepository.GetTokenAddresses(
			system.GetTokenAddressesRequest{
				Chain: &chain.ChainName,
				Limit: 10000,
			},
		)
		if err != nil {
			return err
		}

		withdrawCollectorWorker := WithdrawCollectorWorker{
			chain:                 chain,
			tokens:                tokens,
			privateKey:            privateKey,
			transactionRepository: w.transactionRepository,
			providerPool:          w.providerPool,
			walletServices:        w.walletServices,
			logger:                w.logger,
		}

		worker := worker.NewWorker(
			fmt.Sprintf("withdraw-collector-%s-worker", chain.ChainName),
			w.ctx,
			w.logger,
			withdrawCollectorWorker.ProcessWithdrawals,
			60,
		)
		err = worker.Start()
		if err != nil {
			return err
		}
		w.workers = append(w.workers, worker)
	}
	return nil
}

func (w *WithdrawCollector) Stop() {
	for index := range w.workers {
		err := w.workers[index].Stop()
		if err != nil {
			w.logger.ErrorO(err)
		}
	}
}

func NewWithdrawCollector(
	ctx context.Context,
	collectorPrivateKeys transaction_service.WithdrawCollectorPrivateKeys,
	systemRepository *system.SystemRepository,
	providerPool *provider.ProviderPool,
	transactionRepository *transaction_service.TransactionRepository,
	walletServices *walletservices.WalletServices,
	logger *logger.Logger,
) *WithdrawCollector {
	return &WithdrawCollector{
		ctx:                   ctx,
		logger:                logger,
		systemRepository:      systemRepository,
		transactionRepository: transactionRepository,
		providerPool:          providerPool,
		walletServices:        walletServices,
		workers:               make([]*worker.Worker, 0),
		collectorPrivateKeys:  collectorPrivateKeys,
	}
}
