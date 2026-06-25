package models

import (
	"time"
	"unicode"

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
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	DOB   string `json:"dob"`
	Age   *int   `json:"age,omitempty"`
	Email string `json:"email,omitempty"`
	Role  string `json:"role,omitempty"`
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
	if err := v.RegisterValidation("password_strength", validatePasswordStrength); err != nil {
		panic("failed to register password_strength validator: " + err.Error())
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
	return !t.Before(minDate)
}

// validatePasswordStrength enforces that a password has:
//  1. Minimum 8 characters
//  2. At least one uppercase letter
//  3. At least one digit
//  4. At least one special (non-letter, non-digit) character
//
// Uses the stdlib unicode package — no third-party password library.
func validatePasswordStrength(fl validator.FieldLevel) bool {
	pw := fl.Field().String()
	if len(pw) < 8 {
		return false
	}
	var hasUpper, hasDigit, hasSpecial bool
	for _, r := range pw {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		case !unicode.IsLetter(r) && !unicode.IsDigit(r):
			hasSpecial = true
		}
	}
	return hasUpper && hasDigit && hasSpecial
}
