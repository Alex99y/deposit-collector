package api

import (
	context "context"
	fmt "fmt"
	time "time"

	middlewares "deposit-collector/internal/api/middlewares"
	validations "deposit-collector/internal/api/validations"
	metrics "deposit-collector/internal/metrics"
	system "deposit-collector/internal/system"
	users "deposit-collector/internal/users"
	logger "deposit-collector/pkg/logger"

	fiber "github.com/gofiber/fiber/v3"
	favicon "github.com/gofiber/fiber/v3/middleware/favicon"
	limiter "github.com/gofiber/fiber/v3/middleware/limiter"
	requestid "github.com/gofiber/fiber/v3/middleware/requestid"
	uuid "github.com/google/uuid"
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

type ServerDependencies struct {
	Logger        *logger.Logger
	Metrics       *metrics.Metrics
	UsersHandler  *users.UserHandler
	SystemHandler *system.SystemHandler
}

func NewServer(dependencies ServerDependencies) *Server {
	app := fiber.New(fiber.Config{
		StructValidator: validations.NewStructValidator(),
	})
	app.Use(middlewares.AccessLog(dependencies.Metrics, dependencies.Logger))
	app.Use(requestid.New(requestid.Config{
		Generator: func() string {
			return uuid.New().String()
		},
	}))
	app.Use(favicon.New())

	// TODO: Configure limiter
	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
	}))
	RegisterRoutes(app, RouterDependencies{
		Logger:        dependencies.Logger,
		UsersHandler:  dependencies.UsersHandler,
		SystemHandler: dependencies.SystemHandler,
	})
	return &Server{httpServer: app}
}
