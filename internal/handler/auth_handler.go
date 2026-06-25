package handler

import (
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/Maniii97/aiynx-go/config"
	"github.com/Maniii97/aiynx-go/internal/logger"
	"github.com/Maniii97/aiynx-go/internal/models"
	"github.com/Maniii97/aiynx-go/internal/service"
)

// AuthHandler holds the HTTP handlers for authentication endpoints.
type AuthHandler struct {
	authSvc  service.AuthService
	validate *validator.Validate
	cfg      *config.Config
}

// NewAuthHandler constructs an AuthHandler with its dependencies.
func NewAuthHandler(authSvc service.AuthService, v *validator.Validate, cfg *config.Config) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, validate: v, cfg: cfg}
}

// Signup handles POST /auth/signup.
func (h *AuthHandler) Signup(c *fiber.Ctx) error {
	requestID := c.GetRespHeader("X-Request-ID")

	var req models.SignupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: models.ErrorDetail{
				Message:   "failed to parse request body",
				Code:      "BAD_REQUEST",
				RequestID: requestID,
			},
		})
	}

	if err := h.validate.Struct(req); err != nil {
		logger.Log.Warn("signup validation error", zap.String("error", err.Error()))
		return c.Status(fiber.StatusUnprocessableEntity).JSON(models.ErrorResponse{
			Error: models.ErrorDetail{
				Message:   err.Error(),
				Code:      "VALIDATION_ERROR",
				RequestID: requestID,
			},
		})
	}

	resp, err := h.authSvc.Signup(c.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrEmailConflict) {
			return c.Status(fiber.StatusConflict).JSON(models.ErrorResponse{
				Error: models.ErrorDetail{
					Message:   "email already registered",
					Code:      "CONFLICT",
					RequestID: requestID,
				},
			})
		}
		logger.Log.Error("signup failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: models.ErrorDetail{
				Message:   "internal server error",
				Code:      "INTERNAL_ERROR",
				RequestID: requestID,
			},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// Login handles POST /auth/login.
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	requestID := c.GetRespHeader("X-Request-ID")

	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: models.ErrorDetail{
				Message:   "failed to parse request body",
				Code:      "BAD_REQUEST",
				RequestID: requestID,
			},
		})
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(models.ErrorResponse{
			Error: models.ErrorDetail{
				Message:   err.Error(),
				Code:      "VALIDATION_ERROR",
				RequestID: requestID,
			},
		})
	}

	resp, err := h.authSvc.Login(c.Context(), req, h.cfg)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
				Error: models.ErrorDetail{
					Message:   "invalid email or password",
					Code:      "INVALID_CREDENTIALS",
					RequestID: requestID,
				},
			})
		}
		logger.Log.Error("login failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: models.ErrorDetail{
				Message:   "internal server error",
				Code:      "INTERNAL_ERROR",
				RequestID: requestID,
			},
		})
	}

	// Set HttpOnly cookie; Secure flag depends on environment.
	expiry := time.Duration(h.cfg.JWTExpiryHours) * time.Hour
	secureCookie := h.cfg.Env != "development"
	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    resp.Token,
		HTTPOnly: true,
		Secure:   secureCookie,
		SameSite: "Lax",
		Expires:  time.Now().Add(expiry),
	})

	return c.Status(fiber.StatusOK).JSON(resp)
}
