package http

import (
	"deposit-collector/services/api/http/handlers"

	fiber "github.com/gofiber/fiber/v3"
)

func RegisterRoutes(app *fiber.App) {
	app.Get("/health", handlers.Health)

	app.Get("/users/:id", handlers.GetUser)
	app.Post("/users", handlers.CreateUser)
	app.Post("/users/:id/addresses", handlers.GenerateAddress)
	app.Post("/users/:id/deposits", handlers.ManualDeposit)
	app.Get("/users/:id/operations", handlers.GetUserOperations)
}
