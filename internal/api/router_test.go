package api

import (
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestRegisterRoutesIncludesWithdrawEndpoint(t *testing.T) {
	app := fiber.New()

	RegisterRoutes(app, RouterDependencies{})

	if !hasRoute(app.GetRoutes(true), fiber.MethodPost, "/api/v1/withdraw") {
		t.Fatal("expected POST /api/v1/withdraw route to be registered")
	}
}

func hasRoute(routes []fiber.Route, method, path string) bool {
	for _, route := range routes {
		if route.Method == method && route.Path == path {
			return true
		}
	}
	return false
}
