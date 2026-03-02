package http

import (
	handlers "deposit-collector/cmd/api/http/handlers"
	validations "deposit-collector/cmd/api/http/validations"

	logger "deposit-collector/pkg/logger"

	fiber "github.com/gofiber/fiber/v3"
	healthcheck "github.com/gofiber/fiber/v3/middleware/healthcheck"
)

type RouterDependencies struct {
	Logger        *logger.Logger
	UsersHandler  *handlers.UserHandler
	SystemHandler *handlers.SystemHandler
}

func RegisterRoutes(app *fiber.App, dependencies RouterDependencies) {
	app.Get("/health", healthcheck.New())

	apiV1 := app.Group("/api/v1")

	// Users group
	usersGroup := apiV1.Group("/users")
	usersGroup.Post(
		"/",
		validations.ValidateContentType(validations.ContentTypeJSON),
		dependencies.UsersHandler.CreateUser,
	)
	usersGroup.Get(
		"/:id/addresses",
		dependencies.UsersHandler.GetUserAddresses,
	)
	usersGroup.Post(
		"/:id/addresses",
		validations.ValidateContentType(validations.ContentTypeJSON),
		dependencies.UsersHandler.GenerateAddress,
	)

	// System group
	systemGroup := apiV1.Group("/system")
	systemGroup.Get("/chains", dependencies.SystemHandler.GetSupportedChains)
	systemGroup.Get("/tokens", dependencies.SystemHandler.GetSupportedTokens)
	systemGroup.Post(
		"/chains",
		validations.ValidateContentType(validations.ContentTypeJSON),
		dependencies.SystemHandler.AddNewSupportedChain,
	)
	systemGroup.Post(
		"/tokens",
		validations.ValidateContentType(validations.ContentTypeJSON),
		dependencies.SystemHandler.AddNewTokenAddress,
	)
}
