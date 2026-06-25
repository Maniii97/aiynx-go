package models

// ─── Auth Request DTOs ───────────────────────────────────────────────────────

// SignupRequest is the body accepted by POST /auth/signup.
type SignupRequest struct {
	Name     string `json:"name"     validate:"required,min=1,max=255"`
	DOB      string `json:"dob"      validate:"required,dob_date"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,password_strength"`
}

// LoginRequest is the body accepted by POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// ─── Auth Response DTOs ──────────────────────────────────────────────────────

// SignupResponse is returned for a successful POST /auth/signup.
// NEVER include password_hash here.
type SignupResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// LoginResponse is returned for a successful POST /auth/login.
type LoginResponse struct {
	Token string `json:"token"`
	Role  string `json:"role"`
}

// MeResponse is returned by GET /users/me.
type MeResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	DOB   string `json:"dob"`
	Age   int    `json:"age"`
	Role  string `json:"role"`
}

// ─── Context DTO ─────────────────────────────────────────────────────────────

// AuthUser is injected into the Fiber context by the auth middleware.
// It carries only the minimal identity claims from the JWT — never raw user data.
type AuthUser struct {
	ID   int64  `json:"id"`
	Role string `json:"role"`
}
