package handlers

import (
	json "encoding/json"

	utils "deposit-collector/cmd/api/http/utils"
	system "deposit-collector/internal/system"
	logger "deposit-collector/pkg/logger"

	fiber "github.com/gofiber/fiber/v3"
)

type SystemHandler struct {
	logger        *logger.Logger
	systemService *system.SystemService
}

func (h *SystemHandler) GetSupportedChains(c fiber.Ctx) {
	chains, err := h.systemService.GetSupportedChains()
	if err != nil {
		utils.NewErrorResponse(
			c, fiber.StatusInternalServerError, "error getting supported chains",
		)
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
		utils.NewErrorResponse(
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
		utils.NewErrorResponse(
			c, fiber.StatusInternalServerError, err.Error(),
		)
		return
	}
	c.Status(fiber.StatusOK)
	jsonData, _ := json.Marshal(tokens)
	_, _ = c.Write(jsonData)
}

type AddNewSupportedChainRequest struct {
	Network       string `json:"network" validate:"required"`
	ChainPlatform string `json:"chainPlatform" validate:"required"`
	BIP44CoinType int    `json:"bip44CoinType" validate:"required,min=0"`
	EVMChainID    int    `json:"evmChainId" validate:"required,min=0"`
}

func (h *SystemHandler) AddNewSupportedChain(c fiber.Ctx) {
	var request AddNewSupportedChainRequest
	if err := c.Bind().JSON(&request); err != nil {
		utils.NewErrorResponse(
			c, fiber.StatusBadRequest, err.Error(),
		)
		return
	}

	if err := system.ValidateChainPlatform(request.ChainPlatform); err != nil {
		utils.NewErrorResponse(
			c, fiber.StatusBadRequest, "invalid chain platform",
		)
		return
	}
	err := h.systemService.AddNewSupportedChain(&system.NewSupportedChainRequest{
		Network:       request.Network,
		ChainPlatform: system.ChainPlatform(request.ChainPlatform),
		BIP44CoinType: request.BIP44CoinType,
		EVMChainID:    request.EVMChainID,
	})
	if err != nil {
		utils.NewErrorResponse(
			c, fiber.StatusInternalServerError, err.Error(),
		)
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
	Network    string `json:"network" validate:"required"`
	Decimals   int    `json:"decimals" validate:"required"`
}

func (h *SystemHandler) AddNewTokenAddress(c fiber.Ctx) {
	var request AddNewTokenAddressRequest
	if err := c.Bind().JSON(&request); err != nil {
		utils.NewErrorResponse(
			c, fiber.StatusBadRequest, err.Error(),
		)
		return
	}
	err := h.systemService.AddNewTokenAddress(&system.NewTokenAddressRequest{
		UnitName:   request.UnitName,
		UnitSymbol: request.UnitSymbol,
		Address:    request.Address,
		Network:    request.Network,
		Decimals:   request.Decimals,
	})
	if err != nil {
		utils.NewErrorResponse(
			c, fiber.StatusInternalServerError, "error adding new token address",
		)
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
