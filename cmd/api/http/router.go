package http

import (
	handlers "deposit-collector/cmd/api/http/handlers"
	validations "deposit-collector/cmd/api/http/validations"

	logger "deposit-collector/pkg/logger"

	fiber "github.com/gofiber/fiber/v3"
	healthcheck "github.com/gofiber/fiber/v3/middleware/healthcheck"
)

func RegisterRoutes(app *fiber.App, logger *logger.Logger) {
	app.Get("/health", healthcheck.New())

	apiV1 := app.Group("/api/v1")

	userHandler := handlers.NewUserHandler(logger)

	usersGroup := apiV1.Group("/users")
	usersGroup.Use(validations.ValidateContentType(validations.ContentTypeJSON))
	usersGroup.Get("/:id", userHandler.GetUser)
	usersGroup.Post("/", userHandler.CreateUser)
	usersGroup.Post("/:id/addresses", userHandler.GenerateAddress)
	usersGroup.Post("/:id/deposits", userHandler.ManualDeposit)
	usersGroup.Get("/:id/operations", userHandler.GetUserOperations)

}
