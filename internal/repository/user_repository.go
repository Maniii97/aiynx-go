package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sqlcdb "github.com/Maniii97/aiynx-go/db/sqlc"
)

// ErrNotFound is returned when a requested user does not exist in the database.
var ErrNotFound = errors.New("user not found")

// UserRepository defines the data-access interface for users.
// Keeping it as an interface makes the service layer testable via mocks.
type UserRepository interface {
	Create(ctx context.Context, name string, dob time.Time) (sqlcdb.User, error)
	GetByID(ctx context.Context, id int64) (sqlcdb.User, error)
	Update(ctx context.Context, id int64, name string, dob time.Time) (sqlcdb.User, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int32) ([]sqlcdb.User, error)
	Count(ctx context.Context) (int64, error)
}

type pgxUserRepository struct {
	queries *sqlcdb.Queries
}

// NewUserRepository creates a UserRepository backed by a *sql.DB
// (obtained via pgx/v5/stdlib so we keep pgx as the driver while satisfying
// the database/sql-based SQLC interface).
func NewUserRepository(db *sql.DB) UserRepository {
	return &pgxUserRepository{queries: sqlcdb.New(db)}
}

func (r *pgxUserRepository) Create(ctx context.Context, name string, dob time.Time) (sqlcdb.User, error) {
	return r.queries.CreateUser(ctx, sqlcdb.CreateUserParams{
		Name: name,
		Dob:  dob,
	})
}

func (r *pgxUserRepository) GetByID(ctx context.Context, id int64) (sqlcdb.User, error) {
	user, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sqlcdb.User{}, ErrNotFound
		}
		return sqlcdb.User{}, err
	}
	return user, nil
}

func (r *pgxUserRepository) Update(ctx context.Context, id int64, name string, dob time.Time) (sqlcdb.User, error) {
	user, err := r.queries.UpdateUser(ctx, sqlcdb.UpdateUserParams{
		ID:   id,
		Name: name,
		Dob:  dob,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sqlcdb.User{}, ErrNotFound
		}
		return sqlcdb.User{}, err
	}
	return user, nil
}

func (r *pgxUserRepository) Delete(ctx context.Context, id int64) error {
	return r.queries.DeleteUser(ctx, id)
}

func (r *pgxUserRepository) List(ctx context.Context, limit, offset int32) ([]sqlcdb.User, error) {
	return r.queries.ListUsers(ctx, sqlcdb.ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (r *pgxUserRepository) Count(ctx context.Context) (int64, error) {
	return r.queries.CountUsers(ctx)
}
