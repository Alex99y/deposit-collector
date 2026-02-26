package users

import (
	fmt "fmt"

	logger "deposit-collector/pkg/logger"
)

type UserHandler struct {
	logger *logger.Logger
}

func (h *UserHandler) CreateUser(externalID string) {
	h.logger.Info(fmt.Sprintf("Creating user %s", externalID))

	// TODO: Create user in database
}

func NewUserHandler(logger *logger.Logger) *UserHandler {
	return &UserHandler{logger: logger}
}
