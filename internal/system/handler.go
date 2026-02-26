package system

import (
	logger "deposit-collector/pkg/logger"
)

type SystemHandler struct {
	logger *logger.Logger
}

func (h *SystemHandler) GetSupportedChains() {

}

func (h *SystemHandler) GetSupportedTokens() {
}

func NewSystemController(logger *logger.Logger) *SystemHandler {
	return &SystemHandler{logger: logger}
}
