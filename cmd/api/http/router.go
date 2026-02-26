package http

import (
	handlers "deposit-collector/cmd/api/http/handlers"
	validations "deposit-collector/cmd/api/http/validations"

	logger "deposit-collector/pkg/logger"

	fiber "github.com/gofiber/fiber/v3"
	healthcheck "github.com/gofiber/fiber/v3/middleware/healthcheck"
)

type RouterDependencies struct {
	Logger       *logger.Logger
	UsersHandler *handlers.UserHandler
}

func RegisterRoutes(app *fiber.App, dependencies RouterDependencies) {
	app.Get("/health", healthcheck.New())

	apiV1 := app.Group("/api/v1")

	usersGroup := apiV1.Group("/users")
	usersGroup.Use(validations.ValidateContentType(validations.ContentTypeJSON))
	usersGroup.Get("/:id", dependencies.UsersHandler.GetUser)
	usersGroup.Post("/", dependencies.UsersHandler.CreateUser)
	usersGroup.Post("/:id/addresses", dependencies.UsersHandler.GenerateAddress)
	usersGroup.Post("/:id/deposits", dependencies.UsersHandler.ManualDeposit)
	usersGroup.Get("/:id/operations", dependencies.UsersHandler.GetUserOperations)

}
