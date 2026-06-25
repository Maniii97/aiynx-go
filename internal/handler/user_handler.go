package handler

import (
	"errors"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/Maniii97/aiynx-go/internal/logger"
	"github.com/Maniii97/aiynx-go/internal/middleware"
	"github.com/Maniii97/aiynx-go/internal/models"
	"github.com/Maniii97/aiynx-go/internal/repository"
	"github.com/Maniii97/aiynx-go/internal/service"
)

// UserHandler holds the HTTP handlers for user-related endpoints.
// It is responsible for HTTP concerns only: parsing, validating input,
// calling the service, and writing the response. No business logic here.
type UserHandler struct {
	svc      service.UserService
	validate *validator.Validate
}

// NewUserHandler constructs a UserHandler with its dependencies.
func NewUserHandler(svc service.UserService, v *validator.Validate) *UserHandler {
	return &UserHandler{svc: svc, validate: v}
}

// requestID is a small helper to read the request-id from the response header
// (set by the RequestID middleware before handlers are called).
func requestID(c *fiber.Ctx) string {
	return c.GetRespHeader("X-Request-ID")
}

// HealthCheck handles GET /health.
func (h *UserHandler) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

// GetMe handles GET /users/me — reads user_id from the JWT, never from the URL.
func (h *UserHandler) GetMe(c *fiber.Ctx) error {
	authUser, err := middleware.GetAuthUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{
			Error: models.ErrorDetail{
				Message:   "unauthenticated",
				Code:      "UNAUTHORIZED",
				RequestID: requestID(c),
			},
		})
	}

	user, err := h.svc.GetUserByID(c.Context(), authUser.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
				Error: models.ErrorDetail{
					Message:   "user not found",
					Code:      "NOT_FOUND",
					RequestID: requestID(c),
				},
			})
		}
		logger.Log.Error("get me failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: models.ErrorDetail{
				Message:   "internal server error",
				Code:      "INTERNAL_ERROR",
				RequestID: requestID(c),
			},
		})
	}

	age := 0
	if user.Age != nil {
		age = *user.Age
	}

	return c.JSON(models.MeResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		DOB:   user.DOB,
		Age:   age,
		Role:  user.Role,
	})
}

// CreateUser handles POST /users.
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req models.SignupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: models.ErrorDetail{
				Message:   "failed to parse request body",
				Code:      "BAD_REQUEST",
				RequestID: requestID(c),
			},
		})
	}
	if err := h.validate.Struct(req); err != nil {
		logger.Log.Warn("validation error on create", zap.String("error", err.Error()))
		return c.Status(fiber.StatusUnprocessableEntity).JSON(models.ErrorResponse{
			Error: models.ErrorDetail{
				Message:   err.Error(),
				Code:      "VALIDATION_ERROR",
				RequestID: requestID(c),
			},
		})
	}

	user, err := h.svc.CreateUser(c.Context(), req)
	if err != nil {
		logger.Log.Error("create user failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: models.ErrorDetail{
				Message:   "internal server error",
				Code:      "INTERNAL_ERROR",
				RequestID: requestID(c),
			},
		})
	}
	return c.Status(fiber.StatusCreated).JSON(user)
}

// GetUserByID handles GET /users/:id.
func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: models.ErrorDetail{
				Message:   "invalid user ID",
				Code:      "INVALID_PARAM",
				RequestID: requestID(c),
			},
		})
	}

	user, err := h.svc.GetUserByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
				Error: models.ErrorDetail{
					Message:   "user not found",
					Code:      "NOT_FOUND",
					RequestID: requestID(c),
				},
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: models.ErrorDetail{
				Message:   "internal server error",
				Code:      "INTERNAL_ERROR",
				RequestID: requestID(c),
			},
		})
	}
	return c.JSON(user)
}

// UpdateUser handles PUT /users/:id.
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: models.ErrorDetail{
				Message:   "invalid user ID",
				Code:      "INVALID_PARAM",
				RequestID: requestID(c),
			},
		})
	}

	var req models.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: models.ErrorDetail{
				Message:   "failed to parse request body",
				Code:      "BAD_REQUEST",
				RequestID: requestID(c),
			},
		})
	}
	if err := h.validate.Struct(req); err != nil {
		logger.Log.Warn("validation error on update", zap.String("error", err.Error()))
		return c.Status(fiber.StatusUnprocessableEntity).JSON(models.ErrorResponse{
			Error: models.ErrorDetail{
				Message:   err.Error(),
				Code:      "VALIDATION_ERROR",
				RequestID: requestID(c),
			},
		})
	}

	user, err := h.svc.UpdateUser(c.Context(), id, req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
				Error: models.ErrorDetail{
					Message:   "user not found",
					Code:      "NOT_FOUND",
					RequestID: requestID(c),
				},
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: models.ErrorDetail{
				Message:   "internal server error",
				Code:      "INTERNAL_ERROR",
				RequestID: requestID(c),
			},
		})
	}
	return c.JSON(user)
}

// DeleteUser handles DELETE /users/:id.
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: models.ErrorDetail{
				Message:   "invalid user ID",
				Code:      "INVALID_PARAM",
				RequestID: requestID(c),
			},
		})
	}

	if err := h.svc.DeleteUser(c.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
				Error: models.ErrorDetail{
					Message:   "user not found",
					Code:      "NOT_FOUND",
					RequestID: requestID(c),
				},
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: models.ErrorDetail{
				Message:   "internal server error",
				Code:      "INTERNAL_ERROR",
				RequestID: requestID(c),
			},
		})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ListUsers handles GET /users?page=1&limit=10.
func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	result, err := h.svc.ListUsers(c.Context(), page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: models.ErrorDetail{
				Message:   "internal server error",
				Code:      "INTERNAL_ERROR",
				RequestID: requestID(c),
			},
		})
	}
	return c.JSON(result)
}
