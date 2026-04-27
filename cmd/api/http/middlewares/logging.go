package middlewares

import (
	fmt "fmt"
	strconv "strconv"

	metrics "deposit-collector/internal/metrics"
	logger "deposit-collector/pkg/logger"
	observability "deposit-collector/pkg/observability"

	fiber "github.com/gofiber/fiber/v3"
	requestid "github.com/gofiber/fiber/v3/middleware/requestid"
)

func AccessLog(
	metrics *metrics.Metrics,
	logger *logger.Logger,
) fiber.Handler {
	return func(c fiber.Ctx) error {
		stopTimer := observability.StartTimer()
		err := c.Next()

		status := c.Response().StatusCode()
		lat := stopTimer()

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

		_ = metrics.IncrementAPIRequestsCount(
			c.Method(), c.FullPath(),
		)
		_ = metrics.ObserveAPIRequestsDuration(
			c.FullPath(), strconv.Itoa(status), lat,
		)
		_ = metrics.IncrementAPIRequestsStatus(
			c.Method(), c.FullPath(), strconv.Itoa(status),
		)

		return err
	}
}
