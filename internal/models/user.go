package models

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// ─── Request DTOs ────────────────────────────────────────────────────────────

// CreateUserRequest is the body accepted by POST /users.
type CreateUserRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
	DOB  string `json:"dob"  validate:"required,dob_date"`
}

// UpdateUserRequest is the body accepted by PUT /users/:id.
type UpdateUserRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
	DOB  string `json:"dob"  validate:"required,dob_date"`
}

// ─── Response DTOs ───────────────────────────────────────────────────────────

// UserResponse is returned for single-user endpoints.
// Age is a pointer so it is omitted (omitempty) when not populated
// (e.g. create / update responses don't include age).
type UserResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	DOB  string `json:"dob"`
	Age  *int   `json:"age,omitempty"`
}

// PaginationMeta carries paging metadata in list responses.
type PaginationMeta struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// PaginatedUsersResponse wraps a slice of UserResponse with pagination metadata.
type PaginatedUsersResponse struct {
	Data []UserResponse `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

// ─── Custom Validators ───────────────────────────────────────────────────────

// RegisterCustomValidators registers application-specific validation tags on v.
// Call this once in main after creating the validator instance.
func RegisterCustomValidators(v *validator.Validate) {
	if err := v.RegisterValidation("dob_date", validateDOB); err != nil {
		panic("failed to register dob_date validator: " + err.Error())
	}
}

// validateDOB is the custom validator function for the "dob_date" tag.
// It enforces:
//  1. The value parses as a valid YYYY-MM-DD date.
//  2. The date is not in the future (UTC).
//  3. The date is not more than 130 years in the past (rolling window ).
func validateDOB(fl validator.FieldLevel) bool {
	raw := fl.Field().String()

	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return false
	}

	now := time.Now().UTC()
	// Truncate time-of-day so "today" is always valid.
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if t.After(today) {
		return false
	}

	// Rolling 130-year window: ages gracefully without a magic hardcoded year.
	minDate := time.Now().UTC().AddDate(-130, 0, 0)
	if t.Before(minDate) {
		return false
	}

	return true
}
