package handlers

import (
	json "encoding/json"

	utils "deposit-collector/cmd/api/http/utils"
	worker "deposit-collector/cmd/api/worker"
	memorycache "deposit-collector/internal/memory_cache"
	system "deposit-collector/internal/system"
	users "deposit-collector/internal/users"
	logger "deposit-collector/pkg/logger"

	fiber "github.com/gofiber/fiber/v3"
	uuid "github.com/google/uuid"
)

type UserHandler struct {
	userController *users.UserService
	chainCache     *memorycache.ChainsCache
	publisher      *worker.Publisher
	logger         *logger.Logger
}

type CreateUserRequest struct {
	ExternalID string `json:"externalId" validate:"required"`
}

func (h *UserHandler) CreateUser(c fiber.Ctx) {
	user := new(CreateUserRequest)
	if err := c.Bind().Body(user); err != nil {
		_ = utils.NewErrorResponse(
			c, fiber.StatusBadRequest, "invalid request body",
		)
		return
	}

	if user.ExternalID == "" {
		_ = utils.NewErrorResponse(
			c, fiber.StatusBadRequest, "externalId is required",
		)
		return
	}

	err := h.userController.CreateUser(user.ExternalID)
	if err != nil {
		_ = utils.NewServerErrorResponse(
			c, h.logger, err,
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
		_ = utils.NewErrorResponse(
			c, fiber.StatusBadRequest, err.Error(),
		)
		return
	}
	address, err := h.userController.GenerateAddress(
		c.Params("id"), system.ChainPlatform(request.Chain),
	)
	if err != nil {
		_ = utils.NewServerErrorResponse(c, h.logger, err)
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
		_ = utils.NewServerErrorResponse(c, h.logger, err)
		return
	}
	c.Status(fiber.StatusOK)
	jsonData, _ := json.Marshal(addresses)
	_, _ = c.Write(jsonData)
}

type ManualDepositRequest struct {
	Address   string `json:"address" validate:"required"`
	ChainName string `json:"chainName" validate:"required"`
	TxHash    string `json:"txHash" validate:"required"`
}

func (h *UserHandler) ManualDeposit(c fiber.Ctx) {
	var request ManualDepositRequest
	if err := c.Bind().JSON(&request); err != nil {
		_ = utils.NewErrorResponse(
			c, fiber.StatusBadRequest, err.Error(),
		)
		return
	}
	supportedChain := h.chainCache.GetSupportedChainsByChainName(request.ChainName)
	if supportedChain == nil {
		_ = utils.NewErrorResponse(
			c, fiber.StatusBadRequest, "chain not found",
		)
		return
	}
	userId, addressDbId, err := h.userController.FindUserIdsByAddress(
		request.Address,
	)
	// Invalid request body
	if err != nil {
		_ = utils.NewServerErrorResponse(c, h.logger, err)
		return
	}
	// Address or chain name not found
	if userId == uuid.Nil || addressDbId == uuid.Nil {
		_ = utils.NewErrorResponse(
			c, fiber.StatusBadRequest, "address not found",
		)
		return
	}
	requestId := uuid.New()
	// Publish the deposit operation
	err = h.publisher.PublishDepositOperation(
		c.Context(),
		requestId,
		userId,
		request.ChainName,
		request.Address,
		request.TxHash,
		addressDbId,
	)
	if err != nil {
		_ = utils.NewServerErrorResponse(c, h.logger, err)
		return
	}
	c.Status(fiber.StatusOK)
	jsonData, _ := json.Marshal(map[string]string{
		"message": "Deposit request received. " +
			"If the tx is not finalized, it will be rejected by the system.",
		"id": requestId.String(),
	})
	_, _ = c.Write(jsonData)
}

func NewUserHandler(
	usersService *users.UserService,
	chainCache *memorycache.ChainsCache,
	publisher *worker.Publisher,
	logger *logger.Logger,
) *UserHandler {
	if chainCache == nil || publisher == nil ||
		logger == nil || usersService == nil {
		panic("Invalid handler dependencies")
	}
	return &UserHandler{
		userController: usersService,
		chainCache:     chainCache,
		publisher:      publisher,
		logger:         logger,
	}
}
