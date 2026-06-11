package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Maniii97/aiynx-go/internal/handler"
	"github.com/Maniii97/aiynx-go/internal/middleware"
)

// SetupRoutes registers all middleware and route handlers on the Fiber app.
// Middleware is registered globally (applies to every request) before routes.
func SetupRoutes(app *fiber.App, h *handler.UserHandler) {
	// Global middleware — order matters: RequestID first so Logger can read it.
	app.Use(middleware.RequestID())
	app.Use(middleware.Logger())

	// Health check (used by Docker HEALTHCHECK and CI smoke tests).
	app.Get("/health", h.HealthCheck)

	// User resource.
	app.Post("/users", h.CreateUser)
	app.Get("/users", h.ListUsers)
	app.Get("/users/:id", h.GetUserByID)
	app.Put("/users/:id", h.UpdateUser)
	app.Delete("/users/:id", h.DeleteUser)
}
