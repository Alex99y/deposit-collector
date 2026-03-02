package users

import (
	fmt "fmt"

	system "deposit-collector/internal/system"
	logger "deposit-collector/pkg/logger"
)

type UserService struct {
	usersRepository *UsersRepository
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
	switch chain {
	case system.ChainPlatformEVM:
		// Generate EVM address
	case system.ChainPlatformBTC:
		// Generate BTC address
	case system.ChainPlatformSOL:
		// Generate SOL address
	}
	return nil
}

func NewUserService(
	usersRepository *UsersRepository,
	logger *logger.Logger,
) *UserService {
	return &UserService{
		usersRepository: usersRepository,
		logger:          logger,
	}
}
