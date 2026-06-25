package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/Maniii97/aiynx-go/config"
	"github.com/Maniii97/aiynx-go/internal/handler"
	"github.com/Maniii97/aiynx-go/internal/logger"
	"github.com/Maniii97/aiynx-go/internal/middleware"
	"github.com/Maniii97/aiynx-go/internal/models"
	"github.com/Maniii97/aiynx-go/internal/service"
)

// TestMain initialises the global Zap logger so handlers don't panic on logger.Log.Warn.
func TestMain(m *testing.M) {
	logger.Init("development")
	os.Exit(m.Run())
}


// ─── Minimal stubs ────────────────────────────────────────────────────────────

// stubAuthService implements service.AuthService with configurable responses.
type stubAuthService struct {
	signupResp models.SignupResponse
	signupErr  error
	loginResp  models.LoginResponse
	loginErr   error
}

func (s *stubAuthService) Signup(_ context.Context, _ models.SignupRequest) (models.SignupResponse, error) {
	return s.signupResp, s.signupErr
}
func (s *stubAuthService) Login(_ context.Context, _ models.LoginRequest, _ *config.Config) (models.LoginResponse, error) {
	return s.loginResp, s.loginErr
}

// ─── Test helpers ──────────────────────────────────────────────────────────────

func testCfg() *config.Config {
	return &config.Config{
		JWTSecret:      "test-secret-key-that-is-32-chars!!",
		JWTExpiryHours: 24,
		Env:            "development",
		Port:           "3000",
	}
}

func newApp(cfg *config.Config) (*fiber.App, *handler.AuthHandler) {
	v := validator.New()
	models.RegisterCustomValidators(v)

	authSvc := &stubAuthService{
		signupResp: models.SignupResponse{ID: 1, Name: "Alice", Email: "alice@example.com", Role: "user"},
		loginResp:  models.LoginResponse{Token: "tok", Role: "user"},
	}
	authH := handler.NewAuthHandler(authSvc, v, cfg)
	return fiber.New(), authH
}

// ─── Signup tests ─────────────────────────────────────────────────────────────

func TestSignup_WeakPassword_Returns422(t *testing.T) {
	cfg := testCfg()
	app, authH := newApp(cfg)
	app.Post("/auth/signup", authH.Signup)

	body, _ := json.Marshal(map[string]string{
		"name":     "Alice",
		"dob":      "1990-05-10",
		"email":    "alice@example.com",
		"password": "weak", // no uppercase, no digit, no special, too short
	})

	req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("status = %d; want 422", resp.StatusCode)
	}
}

func TestSignup_DuplicateEmail_Returns409(t *testing.T) {
	cfg := testCfg()
	v := validator.New()
	models.RegisterCustomValidators(v)

	// stub that returns ErrEmailConflict
	authSvc := &stubAuthService{signupErr: service.ErrEmailConflict}
	authH := handler.NewAuthHandler(authSvc, v, cfg)

	app := fiber.New()
	app.Post("/auth/signup", authH.Signup)

	body, _ := json.Marshal(map[string]string{
		"name":     "Alice",
		"dob":      "1990-05-10",
		"email":    "alice@example.com",
		"password": "StrongPass1!",
	})

	req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("status = %d; want 409", resp.StatusCode)
	}
}

// ─── GET /users/me tests ──────────────────────────────────────────────────────

// stubUserService implements just enough for GetMe via handler.
type stubUserService struct {
	resp models.UserResponse
	err  error
}

func (s *stubUserService) CreateUser(_ context.Context, _ models.SignupRequest) (models.UserResponse, error) {
	return models.UserResponse{}, nil
}
func (s *stubUserService) GetUserByID(_ context.Context, _ int64) (models.UserResponse, error) {
	return s.resp, s.err
}
func (s *stubUserService) UpdateUser(_ context.Context, _ int64, _ models.UpdateUserRequest) (models.UserResponse, error) {
	return models.UserResponse{}, nil
}
func (s *stubUserService) DeleteUser(_ context.Context, _ int64) error { return nil }
func (s *stubUserService) ListUsers(_ context.Context, _, _ int) (models.PaginatedUsersResponse, error) {
	return models.PaginatedUsersResponse{}, nil
}

func newMeApp(cfg *config.Config, userSvc service.UserService) *fiber.App {
	v := validator.New()
	models.RegisterCustomValidators(v)
	userH := handler.NewUserHandler(userSvc, v)

	app := fiber.New()
	app.Use(middleware.RequestID())
	// Protected group
	auth := app.Group("/", middleware.AuthMiddleware(cfg))
	auth.Get("/users/me", userH.GetMe)
	return app
}

func TestGetMe_NoToken_Returns401(t *testing.T) {
	cfg := testCfg()
	age := 35
	userSvc := &stubUserService{
		resp: models.UserResponse{ID: 1, Name: "Alice", Email: "alice@example.com", DOB: "1990-05-10", Age: &age, Role: "user"},
	}
	app := newMeApp(cfg, userSvc)

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d; want 401", resp.StatusCode)
	}
}

func TestGetMe_ValidToken_Returns200(t *testing.T) {
	cfg := testCfg()
	age := 35
	userSvc := &stubUserService{
		resp: models.UserResponse{ID: 1, Name: "Alice", Email: "alice@example.com", DOB: "1990-05-10", Age: &age, Role: "user"},
	}
	app := newMeApp(cfg, userSvc)

	token, err := service.GenerateToken(1, "user", cfg)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d; want 200", resp.StatusCode)
	}
}

func TestGetMe_ExpiredToken_Returns401(t *testing.T) {
	cfg := testCfg()
	expiredCfg := &config.Config{
		JWTSecret:      cfg.JWTSecret,
		JWTExpiryHours: -1, // already expired
		Env:            "development",
	}
	userSvc := &stubUserService{}
	app := newMeApp(cfg, userSvc)

	token, _ := service.GenerateToken(1, "user", expiredCfg)

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d; want 401", resp.StatusCode)
	}
}

// ─── RequireRole middleware tests ─────────────────────────────────────────────

func newRBACApp(cfg *config.Config) *fiber.App {
	app := fiber.New()
	app.Use(middleware.RequestID())

	admin := app.Group("/admin", middleware.AuthMiddleware(cfg), middleware.RequireRole("admin"))
	admin.Delete("/users/:id", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})
	return app
}

func TestRequireRole_CorrectRole_Passes(t *testing.T) {
	cfg := testCfg()
	app := newRBACApp(cfg)

	token, _ := service.GenerateToken(1, "admin", cfg)

	req := httptest.NewRequest(http.MethodDelete, "/admin/users/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d; want 204", resp.StatusCode)
	}
}

func TestRequireRole_WrongRole_Returns403(t *testing.T) {
	cfg := testCfg()
	app := newRBACApp(cfg)

	token, _ := service.GenerateToken(1, "user", cfg) // user role, not admin

	req := httptest.NewRequest(http.MethodDelete, "/admin/users/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("status = %d; want 403", resp.StatusCode)
	}
}

// keep the compiler happy with the time import
var _ = time.Now
