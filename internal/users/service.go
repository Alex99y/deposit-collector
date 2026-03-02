package users

import (
	fmt "fmt"

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
		s.logger.Error(fmt.Sprintf("Error creating user %s: %s", externalID, err))
		return err
	}

	s.logger.Debug(fmt.Sprintf("User created with ID %s", externalID))
	return nil
}

func (s *UserService) GenerateAddress(
	externalID string,
	chain system.ChainPlatform,
) error {
	_, err := s.usersRepository.StoreAddress(
		&StoreAddressRequest{
			ExternalID: externalID,
			Chain:      chain,
		},
		func(sequenceNumber int) (string, error) {
			wallet, err := s.walletServices.GenerateWallet(
				externalID, uint32(sequenceNumber), 0, 0, chain,
			)
			if err != nil {
				return "", err
			}
			return wallet.GetAddress(), nil
		},
	)

	return err
}

func NewUserService(
	usersRepository *UsersRepository,
	walletServices *walletservices.WalletServices,
	logger *logger.Logger,
) *UserService {
	return &UserService{
		usersRepository: usersRepository,
		logger:          logger,
	}
}
