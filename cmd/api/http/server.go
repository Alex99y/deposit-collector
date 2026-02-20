package http

import (
	context "context"
	fmt "fmt"

	fiber "github.com/gofiber/fiber/v3"

	middlewares "deposit-collector/cmd/api/http/middlewares"
	logger "deposit-collector/pkg/logger"
)

type Server struct {
	httpServer *fiber.App
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown()
}

func (s *Server) Start(port int, host string) error {
	return s.httpServer.Listen(fmt.Sprintf("%s:%d", host, port))
}

func NewServer(logger *logger.Logger, port int, host string) *Server {
	app := fiber.New()
	app.Use(middlewares.AccessLog(logger))
	RegisterRoutes(app)
	return &Server{httpServer: app}
}
