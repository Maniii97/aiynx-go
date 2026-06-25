package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RequestID ensures every request has a unique X-Request-ID.
//
// Priority:
//  1. If the incoming request already carries an X-Request-ID header
//     (e.g. set by a client SDK or upstream proxy), that value is reused.
//  2. Otherwise a new UUID-v4 is generated.
//
// The ID is stored in c.Locals("requestId") so that the Logger middleware
// and all handlers can read it without re-parsing the header.
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Get("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}
		c.Locals("requestId", id)
		c.Set("X-Request-ID", id)
		return c.Next()
	}
}
