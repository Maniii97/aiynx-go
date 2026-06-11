package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RequestID generates a UUID for every request, stores it in Fiber locals,
// and injects it as the X-Request-ID response header.
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := uuid.New().String()
		c.Locals("requestId", id)
		c.Set("X-Request-ID", id)
		return c.Next()
	}
}
