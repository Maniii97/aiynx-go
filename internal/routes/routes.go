package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Maniii97/aiynx-go/config"
	"github.com/Maniii97/aiynx-go/internal/handler"
	"github.com/Maniii97/aiynx-go/internal/middleware"
)

// SetupRoutes registers all middleware and route handlers on the Fiber app.
// Middleware is registered globally (applies to every request) before routes.
func SetupRoutes(app *fiber.App, h *handler.UserHandler, authH *handler.AuthHandler, cfg *config.Config) {
	// ── Global middleware — order matters: RequestID first so Logger can read it.
	app.Use(middleware.RequestID())
	app.Use(middleware.Logger())

	// ── Public routes (no auth required) ─────────────────────────────────────
	app.Get("/health", h.HealthCheck)
	app.Post("/auth/signup", authH.Signup)
	app.Post("/auth/login", authH.Login)

	// ── Authenticated routes (JWT required) ───────────────────────────────────
	auth := app.Group("/", middleware.AuthMiddleware(cfg))
	auth.Get("/users/me", h.GetMe)
	auth.Get("/users", h.ListUsers)
	auth.Get("/users/:id", h.GetUserByID)
	auth.Post("/users", h.CreateUser)
	auth.Put("/users/:id", h.UpdateUser)

	// ── Admin-only routes (JWT + admin role required) ─────────────────────────
	admin := app.Group("/admin", middleware.AuthMiddleware(cfg), middleware.RequireRole("admin"))
	admin.Delete("/users/:id", h.DeleteUser)
	admin.Get("/users", h.ListUsers) // admin-scoped view of all users
}
