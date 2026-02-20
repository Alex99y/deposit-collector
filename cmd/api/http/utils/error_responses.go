package utils

import (
	fiber "github.com/gofiber/fiber/v3"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func NewErrorResponse(c fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(ErrorResponse{Message: message})
}
