package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/Maniii97/aiynx-go/config"
	"github.com/Maniii97/aiynx-go/internal/handler"
	"github.com/Maniii97/aiynx-go/internal/logger"
	"github.com/Maniii97/aiynx-go/internal/models"
	"github.com/Maniii97/aiynx-go/internal/repository"
	"github.com/Maniii97/aiynx-go/internal/routes"
	"github.com/Maniii97/aiynx-go/internal/service"
)

func main() {
	// ── 1. Config ────────────────────────────────────────────────────────────
	cfg := config.Load()

	// ── 2. Logger ────────────────────────────────────────────────────────────
	logger.Init(cfg.Env)
	defer logger.Log.Sync() //nolint:errcheck

	// ── 3. Database ───────────────────────────────────────────────────────────
	// pgxpool for connection pooling + health-check ping.
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Log.Fatal("failed to create db pool", zap.Error(err))
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		logger.Log.Fatal("database ping failed", zap.Error(err))
	}
	logger.Log.Info("database connected")

	// Wrap pgxpool as *sql.DB so SQLC's database/sql-based generated code works.
	// stdlib.OpenDBFromPool reuses the existing pool (no extra connections).
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	// ── 4. Validator ─────────────────────────────────────────────────────────
	v := validator.New()
	models.RegisterCustomValidators(v)

	// ── 5. Wire dependencies (Repository → Services → Handlers) ───────────────
	repo := repository.NewUserRepository(db)

	authSvc := service.NewAuthService(repo)
	userSvc := service.NewUserService(repo, authSvc)

	userHandler := handler.NewUserHandler(userSvc, v)
	authHandler := handler.NewAuthHandler(authSvc, v, cfg)

	// ── 6. Fiber app ──────────────────────────────────────────────────────────
	app := fiber.New(fiber.Config{
		// Global error handler — catches any panic or unhandled error from handlers.
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var fe *fiber.Error
			if errors.As(err, &fe) {
				code = fe.Code
			}
			requestID, _ := c.Locals("requestId").(string)
			logger.Log.Error("unhandled error", zap.Error(err), zap.String("request_id", requestID))
			return c.Status(code).JSON(models.ErrorResponse{
				Error: models.ErrorDetail{
					Message:   "internal server error",
					Code:      "INTERNAL_ERROR",
					RequestID: requestID,
				},
			})
		},
	})

	routes.SetupRoutes(app, userHandler, authHandler, cfg)

	// ── 7. Graceful shutdown ──────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine so we can block on the signal channel below.
	go func() {
		logger.Log.Info("server starting", zap.String("port", cfg.Port))
		if err := app.Listen(":" + cfg.Port); err != nil {
			logger.Log.Error("server listen error", zap.Error(err))
		}
	}()

	// Block until we receive a termination signal.
	<-quit
	logger.Log.Info("shutdown signal received, draining requests…")

	// fiber.Shutdown waits for in-flight requests to complete.
	if err := app.Shutdown(); err != nil {
		logger.Log.Error("error during server shutdown", zap.Error(err))
	}

	// DB pool is closed via defer above.
	logger.Log.Info("server exited cleanly")
}
