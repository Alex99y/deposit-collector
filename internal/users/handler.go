package users

import (
	json "encoding/json"

	worker "deposit-collector/cmd/api/worker"
	memorycache "deposit-collector/internal/memory_cache"
	system "deposit-collector/internal/system"
	crypto_utils "deposit-collector/pkg/crypto"
	httputils "deposit-collector/pkg/http"
	logger "deposit-collector/pkg/logger"

	fiber "github.com/gofiber/fiber/v3"
	requestid "github.com/gofiber/fiber/v3/middleware/requestid"
	uuid "github.com/google/uuid"
)

type UserHandler struct {
	userController *UserService
	chainCache     *memorycache.ChainsCache
	publisher      *worker.Publisher
	cryptoUtils    *crypto_utils.CryptoUtils
	logger         *logger.Logger
}

type CreateUserRequest struct {
	ExternalID string `json:"externalId" validate:"required"`
}

func (h *UserHandler) CreateUser(c fiber.Ctx) {
	user := new(CreateUserRequest)
	if err := c.Bind().Body(user); err != nil {
		_ = httputils.NewErrorResponse(
			c, fiber.StatusBadRequest, "invalid request body",
		)
		return
	}

	if user.ExternalID == "" {
		_ = httputils.NewErrorResponse(
			c, fiber.StatusBadRequest, "externalId is required",
		)
		return
	}

	err := h.userController.CreateUser(user.ExternalID)
	if err != nil {
		_ = httputils.NewServerErrorResponse(
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
		_ = httputils.NewErrorResponse(
			c, fiber.StatusBadRequest, err.Error(),
		)
		return
	}
	address, err := h.userController.GenerateAddress(
		c.Params("id"), h.cryptoUtils.GetBitcoinNetwork(),
		system.ChainPlatform(request.Chain),
	)
	if err != nil {
		_ = httputils.NewServerErrorResponse(c, h.logger, err)
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
		_ = httputils.NewServerErrorResponse(c, h.logger, err)
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
		_ = httputils.NewErrorResponse(
			c, fiber.StatusBadRequest, err.Error(),
		)
		return
	}
	supportedChain := h.chainCache.GetSupportedChainsByChainName(request.ChainName)
	if supportedChain == nil {
		_ = httputils.NewErrorResponse(
			c, fiber.StatusBadRequest, "chain not found",
		)
		return
	}
	userId, addressDbId, err := h.userController.FindUserIdsByAddress(
		request.Address,
	)
	// Invalid request body
	if err != nil {
		_ = httputils.NewServerErrorResponse(c, h.logger, err)
		return
	}
	// Address or chain name not found
	if userId == uuid.Nil || addressDbId == uuid.Nil {
		_ = httputils.NewErrorResponse(
			c, fiber.StatusBadRequest, "address not found",
		)
		return
	}
	requestUuId := requestid.FromContext(c)
	// Publish the deposit operation
	err = h.publisher.PublishDepositOperation(
		c.Context(),
		uuid.MustParse(requestUuId),
		userId,
		request.ChainName,
		request.Address,
		request.TxHash,
		addressDbId,
	)
	if err != nil {
		_ = httputils.NewServerErrorResponse(c, h.logger, err)
		return
	}
	c.Status(fiber.StatusAccepted)
	jsonData, _ := json.Marshal(map[string]string{
		"message": "Deposit request received. " +
			"If the tx is not finalized, it will be rejected by the system.",
		"id": requestUuId,
	})
	_, _ = c.Write(jsonData)
}

type RequestWithdrawRequest struct {
	ExternalID string `json:"externalId" validate:"required"`
	Amount     int64  `json:"amount" validate:"required"`
	ChainName  string `json:"chainName" validate:"required"`
	Address    string `json:"address" validate:"required"`
}

func (h *UserHandler) RequestWithdraw(c fiber.Ctx) {
	var request RequestWithdrawRequest
	if err := c.Bind().JSON(&request); err != nil {
		_ = httputils.NewErrorResponse(
			c, fiber.StatusBadRequest, err.Error(),
		)
		return
	}
	// Validate amount
	if request.Amount == 0 || request.Amount < 0 {
		_ = httputils.NewErrorResponse(
			c, fiber.StatusBadRequest,
			"amount is required and must be greater than 0",
		)
		return
	}
	// Validate chain name
	if request.ChainName == "" {
		_ = httputils.NewErrorResponse(
			c, fiber.StatusBadRequest, "chain name is required",
		)
		return
	}
	// Validate external ID
	if request.ExternalID == "" {
		_ = httputils.NewErrorResponse(
			c, fiber.StatusBadRequest, "external ID is required",
		)
		return
	}
	// Validate user exists
	user, err := h.userController.GetUserByExternalID(request.ExternalID)
	if err != nil {
		_ = httputils.NewServerErrorResponse(c, h.logger, err)
		return
	}
	if user == nil {
		_ = httputils.NewErrorResponse(c, fiber.StatusBadRequest, "user not found")
		return
	}
	// Validate ChainName
	chainSupported := h.chainCache.GetSupportedChainsByChainName(request.ChainName)
	if chainSupported == nil {
		_ = httputils.NewErrorResponse(
			c, fiber.StatusBadRequest, "chain not supported",
		)
		return
	}
	// Validate destination address
	if !h.cryptoUtils.ValidateAddress(
		request.Address, chainSupported.ChainPlatform,
	) {
		_ = httputils.NewErrorResponse(
			c, fiber.StatusBadRequest, "invalid address",
		)
		return
	}

	requestUuId := requestid.FromContext(c)

	// Publish the withdraw operation
	err = h.publisher.PublishWithdrawOperation(
		c.Context(),
		uuid.MustParse(requestUuId),
		user.ID,
		request.ChainName,
		request.Address,
		request.Amount,
	)
	if err != nil {
		_ = httputils.NewServerErrorResponse(c, h.logger, err)
		return
	}
	c.Status(fiber.StatusAccepted)
	jsonData, _ := json.Marshal(map[string]string{
		"message": "Withdraw request received. " +
			"If the tx is not finalized, it will be rejected by the system.",
		"id": requestUuId,
	})
	_, _ = c.Write(jsonData)
}

func NewUserHandler(
	usersService *UserService,
	chainCache *memorycache.ChainsCache,
	publisher *worker.Publisher,
	cryptoUtils *crypto_utils.CryptoUtils,
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
		cryptoUtils:    cryptoUtils,
		logger:         logger,
	}
}
