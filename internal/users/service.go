package users

import (
	system "deposit-collector/internal/system"
	walletservices "deposit-collector/internal/wallet_services"
	logger "deposit-collector/pkg/logger"
)

type UserService struct {
	usersRepository *UsersRepository
	walletServices  *walletservices.WalletServices
	logger          *logger.Logger
}

func (s *UserService) CreateUser(externalID string) error {
	err := s.usersRepository.CreateUser(externalID)
	if err != nil {
		return err
	}

	s.logger.Debug("User created with ID " + externalID)
	return nil
}

func (s *UserService) GenerateAddress(
	externalID string,
	chain system.ChainPlatform,
) (string, error) {
	address, err := s.usersRepository.StoreAddress(
		&StoreAddressRequest{
			ExternalID: externalID,
			Chain:      chain,
		},
		func(userAccountID uint32, sequenceNumber uint32) (string, error) {
			wallet, err := s.walletServices.GenerateWallet(
				userAccountID, 0, sequenceNumber, chain,
			)
			if err != nil {
				return "", err
			}
			return wallet.GetAddress(), nil
		},
	)

	return address, err
}

func (s *UserService) AddressAndChainNameExists(
	address string,
	chainName string,
) (bool, error) {
	return s.usersRepository.AddressAndChainNameExists(address, chainName)
}

func (s *UserService) GetUserAddresses(
	externalID string,
) ([]StoredAddress, error) {
	return s.usersRepository.GetAddressesByExternalID(externalID)
}

func NewUserService(
	usersRepository *UsersRepository,
	walletServices *walletservices.WalletServices,
	logger *logger.Logger,
) *UserService {
	return &UserService{
		usersRepository: usersRepository,
		walletServices:  walletServices,
		logger:          logger,
	}
}
