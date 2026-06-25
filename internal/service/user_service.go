package service

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/Maniii97/aiynx-go/internal/logger"
	"github.com/Maniii97/aiynx-go/internal/models"
	"github.com/Maniii97/aiynx-go/internal/repository"
)

// UserService defines the business-logic interface for user operations.
type UserService interface {
	CreateUser(ctx context.Context, req models.SignupRequest) (models.UserResponse, error)
	GetUserByID(ctx context.Context, id int64) (models.UserResponse, error)
	UpdateUser(ctx context.Context, id int64, req models.UpdateUserRequest) (models.UserResponse, error)
	DeleteUser(ctx context.Context, id int64) error
	ListUsers(ctx context.Context, page, limit int) (models.PaginatedUsersResponse, error)
}

type userServiceImpl struct {
	repo     repository.UserRepository
	authSvc  AuthService
}

// NewUserService constructs a UserService with the given repository.
// authSvc is used for password hashing when creating users via POST /users.
func NewUserService(repo repository.UserRepository, authSvc AuthService) UserService {
	return &userServiceImpl{repo: repo, authSvc: authSvc}
}

// CalculateAge returns the age in full years as of today (UTC).
// It correctly handles the case where the birthday has not yet occurred this calendar year.
// This is a pure function with no side effects — safe to unit-test without DB.
func CalculateAge(dob time.Time) int {
	today := time.Now().UTC()
	years := today.Year() - dob.Year()
	// Birthday hasn't happened yet this calendar year.
	if today.Month() < dob.Month() ||
		(today.Month() == dob.Month() && today.Day() < dob.Day()) {
		years--
	}
	return years
}

// ─── Service Methods ─────────────────────────────────────────────────────────

func (s *userServiceImpl) CreateUser(ctx context.Context, req models.SignupRequest) (models.UserResponse, error) {
	dob, _ := time.Parse("2006-01-02", req.DOB) // already validated upstream

	hash, err := HashPassword(req.Password)
	if err != nil {
		logger.Log.Error("failed to hash password", zap.Error(err))
		return models.UserResponse{}, err
	}

	role := "user"

	user, err := s.repo.CreateWithAuth(ctx, req.Name, dob, req.Email, hash, role)
	if err != nil {
		logger.Log.Error("failed to create user", zap.Error(err))
		return models.UserResponse{}, err
	}

	logger.Log.Info("user created", zap.Int64("user_id", user.ID))
	return models.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		DOB:   user.Dob.Format("2006-01-02"),
		Email: user.Email,
		Role:  user.Role,
		// Age intentionally omitted from create response per spec.
	}, nil
}

func (s *userServiceImpl) GetUserByID(ctx context.Context, id int64) (models.UserResponse, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			logger.Log.Warn("user not found", zap.Int64("user_id", id))
		} else {
			logger.Log.Error("failed to get user", zap.Int64("user_id", id), zap.Error(err))
		}
		return models.UserResponse{}, err
	}

	logger.Log.Info("user fetched", zap.Int64("user_id", user.ID))
	age := CalculateAge(user.Dob)
	return models.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		DOB:   user.Dob.Format("2006-01-02"),
		Age:   &age,
		Email: user.Email,
		Role:  user.Role,
	}, nil
}

func (s *userServiceImpl) UpdateUser(ctx context.Context, id int64, req models.UpdateUserRequest) (models.UserResponse, error) {
	dob, _ := time.Parse("2006-01-02", req.DOB) // already validated upstream

	user, err := s.repo.Update(ctx, id, req.Name, dob)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			logger.Log.Warn("user not found for update", zap.Int64("user_id", id))
		} else {
			logger.Log.Error("failed to update user", zap.Int64("user_id", id), zap.Error(err))
		}
		return models.UserResponse{}, err
	}

	logger.Log.Info("user updated", zap.Int64("user_id", user.ID))
	return models.UserResponse{
		ID:   user.ID,
		Name: user.Name,
		DOB:  user.Dob.Format("2006-01-02"),
		// Age intentionally omitted from update response per spec.
	}, nil
}

func (s *userServiceImpl) DeleteUser(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		logger.Log.Error("failed to delete user", zap.Int64("user_id", id), zap.Error(err))
		return err
	}
	logger.Log.Info("user deleted", zap.Int64("user_id", id))
	return nil
}

func (s *userServiceImpl) ListUsers(ctx context.Context, page, limit int) (models.PaginatedUsersResponse, error) {
	// Clamp pagination params.
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset := int32((page - 1) * limit)

	users, err := s.repo.List(ctx, int32(limit), offset)
	if err != nil {
		logger.Log.Error("failed to list users", zap.Error(err))
		return models.PaginatedUsersResponse{}, err
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		logger.Log.Error("failed to count users", zap.Error(err))
		return models.PaginatedUsersResponse{}, err
	}

	data := make([]models.UserResponse, 0, len(users))
	for _, u := range users {
		age := CalculateAge(u.Dob)
		data = append(data, models.UserResponse{
			ID:    u.ID,
			Name:  u.Name,
			DOB:   u.Dob.Format("2006-01-02"),
			Age:   &age,
			Email: u.Email,
			Role:  u.Role,
		})
	}

	return models.PaginatedUsersResponse{
		Data: data,
		Meta: models.PaginationMeta{
			Page:  page,
			Limit: limit,
			Total: int(total),
		},
	}, nil
}
