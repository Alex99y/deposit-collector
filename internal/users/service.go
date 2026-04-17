package users

import (
	system "deposit-collector/internal/system"
	walletservices "deposit-collector/internal/wallet_services"
	btc_utils "deposit-collector/pkg/crypto/btc"
	logger "deposit-collector/pkg/logger"

	uuid "github.com/google/uuid"
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
	network btc_utils.NETWORK,
	chain system.ChainPlatform,
) (string, error) {
	address, err := s.usersRepository.StoreAddress(
		&StoreAddressRequest{
			ExternalID: externalID,
			Chain:      chain,
		},
		func(userAccountID uint32, sequenceNumber uint32) (string, error) {
			wallet, err := s.walletServices.GenerateWallet(
				userAccountID, 0, sequenceNumber, network, chain,
			)
			if err != nil {
				return "", err
			}
			return wallet.GetAddress(), nil
		},
	)

	return address, err
}

func (s *UserService) FindUserIdsByAddress(
	address string,
) (uuid.UUID, uuid.UUID, error) {
	return s.usersRepository.FindUserIDAndAddressIDByAddress(
		address,
	)
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
