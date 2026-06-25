package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/Maniii97/aiynx-go/config"
	"github.com/Maniii97/aiynx-go/internal/models"
	"github.com/Maniii97/aiynx-go/internal/repository"
)

// ─── Sentinel errors ──────────────────────────────────────────────────────────

// ErrEmailConflict is returned when an email is already registered.
var ErrEmailConflict = errors.New("email already exists")

// ErrInvalidCredentials is returned when email/password don't match.
var ErrInvalidCredentials = errors.New("invalid credentials")

// ─── JWT Claims ───────────────────────────────────────────────────────────────

// Claims defines the JWT payload.
type Claims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// ─── Password helpers ─────────────────────────────────────────────────────────

// HashPassword hashes a plaintext password with bcrypt (DefaultCost = 10).
// Never log or return the returned hash directly — treat it as an opaque blob.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword returns true when the plaintext password matches the stored hash.
func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// ─── JWT helpers ──────────────────────────────────────────────────────────────

// GenerateToken creates a signed HS256 JWT for the given user.
func GenerateToken(userID int64, role string, cfg *config.Config) (string, error) {
	expiry := time.Duration(cfg.JWTExpiryHours) * time.Hour
	now := time.Now()

	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

// ParseToken validates a JWT string and returns the embedded claims.
// Returns an error if the token is expired, tampered, or otherwise invalid.
func ParseToken(tokenStr string, cfg *config.Config) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

// ─── AuthService ──────────────────────────────────────────────────────────────

// AuthService defines the business-logic interface for authentication.
type AuthService interface {
	Signup(ctx context.Context, req models.SignupRequest) (models.SignupResponse, error)
	Login(ctx context.Context, req models.LoginRequest, cfg *config.Config) (models.LoginResponse, error)
}

type authServiceImpl struct {
	repo repository.UserRepository
}

// NewAuthService constructs an AuthService with the given repository.
func NewAuthService(repo repository.UserRepository) AuthService {
	return &authServiceImpl{repo: repo}
}

func (s *authServiceImpl) Signup(ctx context.Context, req models.SignupRequest) (models.SignupResponse, error) {
	// Check for duplicate email.
	if _, err := s.repo.GetByEmail(ctx, req.Email); err == nil {
		return models.SignupResponse{}, ErrEmailConflict
	}

	dob, _ := time.Parse("2006-01-02", req.DOB) // already validated upstream

	hash, err := HashPassword(req.Password)
	if err != nil {
		return models.SignupResponse{}, err
	}

	role := "user" // default role; only admin can assign admin role

	user, err := s.repo.CreateWithAuth(ctx, req.Name, dob, req.Email, hash, role)
	if err != nil {
		return models.SignupResponse{}, err
	}

	return models.SignupResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}, nil
}

func (s *authServiceImpl) Login(ctx context.Context, req models.LoginRequest, cfg *config.Config) (models.LoginResponse, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		// Don't reveal whether the email exists.
		return models.LoginResponse{}, ErrInvalidCredentials
	}

	if !CheckPassword(req.Password, user.PasswordHash) {
		return models.LoginResponse{}, ErrInvalidCredentials
	}

	token, err := GenerateToken(user.ID, user.Role, cfg)
	if err != nil {
		return models.LoginResponse{}, err
	}

	return models.LoginResponse{
		Token: token,
		Role:  user.Role,
	}, nil
}
