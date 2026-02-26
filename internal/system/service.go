package system

import (
	fmt "fmt"

	logger "deposit-collector/pkg/logger"
)

type SystemService struct {
	logger           *logger.Logger
	systemRepository *SystemRepository
}

func (s *SystemService) GetSupportedChains() ([]SupportedChain, error) {
	chains, err := s.systemRepository.GetSupportedChains()
	if err != nil {
		s.logger.Error(fmt.Sprintf("Error getting supported chains: %s", err))
		return nil, err
	}
	return chains, nil
}

func (s *SystemService) AddNewSupportedChain(
	chain *NewSupportedChainRequest,
) error {
	err := s.systemRepository.AddNewSupportedChain(chain)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Error adding new supported chain: %s", err))
		return err
	}
	return nil
}

func (s *SystemService) GetSupportedTokens(
	filters GetTokenAddressesRequest,
) ([]TokenAddress, error) {
	tokens, err := s.systemRepository.GetTokenAddresses(filters)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func (s *SystemService) AddNewTokenAddress(
	tokenAddress *NewTokenAddressRequest,
) error {
	err := s.systemRepository.AddNewTokenAddress(tokenAddress)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	return nil
}

func NewSystemService(
	systemRepository *SystemRepository,
	logger *logger.Logger,
) *SystemService {
	return &SystemService{
		systemRepository: systemRepository,
		logger:           logger,
	}
}
