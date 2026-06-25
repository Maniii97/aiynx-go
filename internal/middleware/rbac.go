package middleware

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Maniii97/aiynx-go/internal/models"
)

// RequireRole returns middleware that permits only requests whose JWT role
// matches one of the supplied allowed roles.
// Must be called after AuthMiddleware so that c.Locals("user") is populated.
func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID, _ := c.Locals("requestId").(string)

		user, ok := c.Locals("user").(*models.AuthUser)
		if !ok || user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
				Error: models.ErrorDetail{
					Message:   "missing authentication token",
					Code:      "UNAUTHORIZED",
					RequestID: requestID,
				},
			})
		}

		for _, r := range roles {
			if user.Role == r {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(models.ErrorResponse{
			Error: models.ErrorDetail{
				Message:   "insufficient permissions",
				Code:      "FORBIDDEN",
				RequestID: requestID,
			},
		})
	}
}
