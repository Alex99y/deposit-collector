package handlers

import (
	json "encoding/json"

	utils "deposit-collector/cmd/api/http/utils"
	system "deposit-collector/internal/system"
	users "deposit-collector/internal/users"
	logger "deposit-collector/pkg/logger"

	fiber "github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	userController *users.UserService
	logger         *logger.Logger
}

type CreateUserRequest struct {
	ExternalID string `json:"externalId" validate:"required"`
}

func (h *UserHandler) CreateUser(c fiber.Ctx) {
	user := new(CreateUserRequest)
	if err := c.Bind().Body(user); err != nil {
		utils.NewErrorResponse(
			c, fiber.StatusBadRequest, "invalid request body",
		)
		return
	}

	if user.ExternalID == "" {
		utils.NewErrorResponse(
			c, fiber.StatusBadRequest, "externalId is required",
		)
		return
	}

	err := h.userController.CreateUser(user.ExternalID)
	if err != nil {
		utils.NewErrorResponse(
			c, fiber.StatusBadRequest, "error creating user",
		)
		return
	}

	c.Status(fiber.StatusOK)
	jsonData, _ := json.Marshal(user)
	_, _ = c.Write(jsonData)
}

type GenerateAddressRequest struct {
	Chain string `json:"chain" validate:"required"`
}

func (h *UserHandler) GenerateAddress(c fiber.Ctx) {
	var request GenerateAddressRequest
	if err := c.Bind().JSON(&request); err != nil {
		utils.NewErrorResponse(
			c, fiber.StatusBadRequest, err.Error(),
		)
		return
	}
	address, err := h.userController.GenerateAddress(
		c.Params("id"), system.ChainPlatform(request.Chain),
	)
	if err != nil {
		utils.NewErrorResponse(
			c, fiber.StatusInternalServerError, err.Error(),
		)
		return
	}
	c.Status(fiber.StatusOK)
	jsonData, _ := json.Marshal(map[string]string{
		"address": address,
	})
	_, _ = c.Write(jsonData)
}

func (h *UserHandler) GetUserAddresses(c fiber.Ctx) {
	// @TODO: Filter by platform
	addresses, err := h.userController.GetUserAddresses(c.Params("id"))
	if err != nil {
		utils.NewErrorResponse(
			c, fiber.StatusInternalServerError, err.Error(),
		)
		return
	}
	c.Status(fiber.StatusOK)
	jsonData, _ := json.Marshal(addresses)
	_, _ = c.Write(jsonData)
}

func (h *UserHandler) ManualDeposit(c fiber.Ctx) {
	c.Status(fiber.StatusOK)
	_, _ = c.Write([]byte("ok"))
}

func NewUserHandler(
	usersService *users.UserService,
	logger *logger.Logger,
) *UserHandler {
	return &UserHandler{
		userController: usersService,
		logger:         logger,
	}
}
