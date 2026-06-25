package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/Maniii97/aiynx-go/config"
	"github.com/Maniii97/aiynx-go/internal/models"
	"github.com/Maniii97/aiynx-go/internal/service"
)

// AuthMiddleware validates the JWT from the Authorization header or auth_token cookie.
// On success it injects an *models.AuthUser into c.Locals("user").
// On failure it returns 401 with the standard error envelope.
func AuthMiddleware(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID, _ := c.Locals("requestId").(string)

		tokenStr := extractToken(c)
		if tokenStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
				Error: models.ErrorDetail{
					Message:   "missing authentication token",
					Code:      "UNAUTHORIZED",
					RequestID: requestID,
				},
			})
		}

		claims, err := service.ParseToken(tokenStr, cfg)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
				Error: models.ErrorDetail{
					Message:   "invalid or expired token",
					Code:      "UNAUTHORIZED",
					RequestID: requestID,
				},
			})
		}

		c.Locals("user", &models.AuthUser{
			ID:   claims.UserID,
			Role: claims.Role,
		})
		return c.Next()
	}
}

// extractToken pulls the raw JWT string from the Authorization header (Bearer scheme)
// and falls back to the auth_token cookie if the header is absent.
func extractToken(c *fiber.Ctx) string {
	header := c.Get("Authorization")
	if header != "" {
		parts := strings.SplitN(header, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return parts[1]
		}
	}
	return c.Cookies("auth_token")
}

// GetAuthUser retrieves the authenticated user from Fiber context locals.
// Returns an error if the middleware was not applied or the token was invalid.
func GetAuthUser(c *fiber.Ctx) (*models.AuthUser, error) {
	user, ok := c.Locals("user").(*models.AuthUser)
	if !ok || user == nil {
		return nil, errors.New("unauthenticated")
	}
	return user, nil
}
