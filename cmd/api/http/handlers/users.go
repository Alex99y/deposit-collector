package handlers

import (
	json "encoding/json"

	utils "deposit-collector/cmd/api/http/utils"
	controllers "deposit-collector/internal/http/controllers"
	logger "deposit-collector/pkg/logger"

	fiber "github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	userController *controllers.UserHandler
	logger         *logger.Logger
}

func (h *UserHandler) GetUser(c fiber.Ctx) {
	c.Status(fiber.StatusOK)
	_, _ = c.Write([]byte("ok"))
}

type CreateUserRequest struct {
	ExternalID string `json:"external_id" validate:"required"`
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
			c, fiber.StatusBadRequest, "external_id is required",
		)
		return
	}

	h.userController.CreateUser(user.ExternalID)

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

func NewUserHandler(logger *logger.Logger) *UserHandler {
	return &UserHandler{
		userController: controllers.NewUserHandler(logger),
		logger:         logger,
	}
}
