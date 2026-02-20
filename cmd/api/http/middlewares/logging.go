package middlewares

import (
	"fmt"
	"time"

	logger "deposit-collector/pkg/logger"

	fiber "github.com/gofiber/fiber/v3"
	requestid "github.com/gofiber/fiber/v3/middleware/requestid"
)

func AccessLog(logger *logger.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		status := c.Response().StatusCode()
		lat := time.Since(start)

		if err != nil {
			if fe, ok := err.(*fiber.Error); ok {
				status = fe.Code
			} else {
				status = fiber.StatusInternalServerError
			}
			_ = c.Status(status)
		}

		logger.Info(
			fmt.Sprintf("http_request [%d %s %s] %dms %s id: %s",
				status,
				c.Method(),
				c.Path(),
				lat.Milliseconds(),
				c.IP(),
				requestid.FromContext(c),
			))

		return err
	}
}
