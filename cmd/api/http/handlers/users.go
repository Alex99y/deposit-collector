package handlers

import (
	json "encoding/json"

	utils "deposit-collector/cmd/api/http/utils"
	users "deposit-collector/internal/users"
	logger "deposit-collector/pkg/logger"

	fiber "github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	userController *users.UserService
	logger         *logger.Logger
}

func (h *UserHandler) GetUser(c fiber.Ctx) {
	c.Status(fiber.StatusOK)
	_, _ = c.Write([]byte("ok"))
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

func (h *UserHandler) GenerateAddress(c fiber.Ctx) {
	c.Status(fiber.StatusOK)
	_, _ = c.Write([]byte("ok"))
}

func (h *UserHandler) ManualDeposit(c fiber.Ctx) {
	c.Status(fiber.StatusOK)
	_, _ = c.Write([]byte("ok"))
}

func (h *UserHandler) GetUserOperations(c fiber.Ctx) {
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
