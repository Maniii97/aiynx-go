package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/Maniii97/aiynx-go/internal/logger"
)

// Logger records the start time before delegating to the next handler,
// then logs method, path, status, and latency after the response is written.
// It picks up the request ID injected by the RequestID middleware.
func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		latency := time.Since(start).Milliseconds()
		requestID, _ := c.Locals("requestId").(string)

		logger.Log.Info("request completed",
			zap.String("request_id", requestID),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", c.Response().StatusCode()),
			zap.Int64("latency_ms", latency),
		)

		return err
	}
}
