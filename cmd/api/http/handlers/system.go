package handlers

import (
	json "encoding/json"

	utils "deposit-collector/cmd/api/http/utils"
	system "deposit-collector/internal/system"
	logger "deposit-collector/pkg/logger"
	commonUtils "deposit-collector/pkg/utils"

	fiber "github.com/gofiber/fiber/v3"
)

type SystemHandler struct {
	logger        *logger.Logger
	systemService *system.SystemService
}

func (h *SystemHandler) GetSupportedChains(c fiber.Ctx) {
	chains, err := h.systemService.GetSupportedChains()
	if err != nil {
		_ = utils.NewServerErrorResponse(c, h.logger, err)
		return
	}
	c.Status(fiber.StatusOK)
	jsonData, _ := json.Marshal(chains)
	_, _ = c.Write(jsonData)
}

type GetSupportedTokensQuery struct {
	Chain      *string `query:"chain"`
	Address    *string `query:"address"`
	UnitSymbol *string `query:"unitSymbol"`
	Limit      int     `query:"limit,default:100" validate:"min=1,max=100"`
	Offset     int     `query:"offset,default:0" validate:"min=0"`
}

func (h *SystemHandler) GetSupportedTokens(c fiber.Ctx) {
	query := new(GetSupportedTokensQuery)
	if err := c.Bind().Query(query); err != nil {
		if commonUtils.ContainsAny(err.Error(), "user not found") {
			_ = utils.NewErrorResponse(
				c, fiber.StatusNotFound, "user not found",
			)
			return
		}
		_ = utils.NewErrorResponse(
			c, fiber.StatusBadRequest, "invalid request body",
		)
		return
	}

	tokens, err := h.systemService.GetSupportedTokens(
		system.GetTokenAddressesRequest{
			Chain:      query.Chain,
			Address:    query.Address,
			UnitSymbol: query.UnitSymbol,
			Limit:      query.Limit,
			Offset:     query.Offset,
		},
	)

	if err != nil {
		_ = utils.NewServerErrorResponse(c, h.logger, err)
		return
	}
	c.Status(fiber.StatusOK)
	jsonData, _ := json.Marshal(tokens)
	_, _ = c.Write(jsonData)
}

type AddNewSupportedChainRequest struct {
	ChainName     string `json:"chainName" validate:"required"`
	ChainPlatform string `json:"chainPlatform" validate:"required"`
	EVMChainID    int    `json:"evmChainId" validate:"required,min=0"`
}

func (h *SystemHandler) AddNewSupportedChain(c fiber.Ctx) {
	var request AddNewSupportedChainRequest
	if err := c.Bind().JSON(&request); err != nil {
		_ = utils.NewErrorResponse(
			c, fiber.StatusBadRequest, err.Error(),
		)
		return
	}

	if err := system.ValidateChainPlatform(request.ChainPlatform); err != nil {
		_ = utils.NewErrorResponse(
			c, fiber.StatusBadRequest, "invalid chain platform",
		)
		return
	}
	err := h.systemService.AddNewSupportedChain(&system.NewSupportedChainRequest{
		ChainName:     request.ChainName,
		ChainPlatform: system.ChainPlatform(request.ChainPlatform),
		EVMChainID:    request.EVMChainID,
	})
	if err != nil {
		_ = utils.NewServerErrorResponse(c, h.logger, err)
		return
	}
	c.Status(fiber.StatusOK)
	jsonData, _ := json.Marshal(request)
	_, _ = c.Write(jsonData)
}

type AddNewTokenAddressRequest struct {
	UnitName   string `json:"unitName" validate:"required"`
	UnitSymbol string `json:"unitSymbol" validate:"required"`
	Address    string `json:"address" validate:"required"`
	ChainName  string `json:"chainName" validate:"required"`
	Decimals   int    `json:"decimals" validate:"required"`
}

func (h *SystemHandler) AddNewTokenAddress(c fiber.Ctx) {
	var request AddNewTokenAddressRequest
	if err := c.Bind().JSON(&request); err != nil {
		_ = utils.NewErrorResponse(
			c, fiber.StatusBadRequest, err.Error(),
		)
		return
	}
	err := h.systemService.AddNewTokenAddress(&system.NewTokenAddressRequest{
		BaseTokenAddress: system.BaseTokenAddress{
			UnitName:   request.UnitName,
			UnitSymbol: request.UnitSymbol,
			Address:    request.Address,
			Decimals:   request.Decimals,
		},
		ChainName: request.ChainName,
	})
	if err != nil {
		_ = utils.NewServerErrorResponse(c, h.logger, err)
		return
	}
	c.Status(fiber.StatusOK)
	jsonData, _ := json.Marshal(request)
	_, _ = c.Write(jsonData)
}

func NewSystemHandler(
	systemService *system.SystemService,
	logger *logger.Logger,
) *SystemHandler {
	return &SystemHandler{
		systemService: systemService,
		logger:        logger,
	}
}
