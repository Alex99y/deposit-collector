package http

import (
	context "context"
	fmt "fmt"
	time "time"

	fiber "github.com/gofiber/fiber/v3"
	favicon "github.com/gofiber/fiber/v3/middleware/favicon"
	limiter "github.com/gofiber/fiber/v3/middleware/limiter"
	requestid "github.com/gofiber/fiber/v3/middleware/requestid"

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
	app.Use(requestid.New())
	app.Use(favicon.New())

	// TODO: Configure limiter
	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
	}))
	RegisterRoutes(app, logger)
	return &Server{httpServer: app}
}
