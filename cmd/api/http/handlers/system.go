package handlers

import (
	json "encoding/json"

	system "deposit-collector/internal/system"
	logger "deposit-collector/pkg/logger"

	fiber "github.com/gofiber/fiber/v3"
)

type SystemHandler struct {
	logger           *logger.Logger
	systemController *system.SystemHandler
}

func (h *SystemHandler) GetSupportedChains(c fiber.Ctx) {
	h.systemController.GetSupportedChains()
	c.Status(fiber.StatusOK)
	jsonData, _ := json.Marshal([]string{})
	_, _ = c.Write(jsonData)
}

func (h *SystemHandler) GetSupportedTokens(c fiber.Ctx) {
	h.systemController.GetSupportedTokens()
	c.Status(fiber.StatusOK)
	jsonData, _ := json.Marshal([]string{})
	_, _ = c.Write(jsonData)
}

func NewSystemHandler(logger *logger.Logger) *SystemHandler {
	return &SystemHandler{
		logger:           logger,
		systemController: system.NewSystemController(logger),
	}
}
