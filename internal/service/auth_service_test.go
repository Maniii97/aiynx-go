package service_test

import (
	"testing"

	"github.com/Maniii97/aiynx-go/config"
	"github.com/Maniii97/aiynx-go/internal/service"
)

// ─── Password helpers ─────────────────────────────────────────────────────────

func TestHashPassword_RoundTrip(t *testing.T) {
	password := "MyP@ssword1"

	hash, err := service.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword() returned empty string")
	}
	// Hash must not equal the plaintext.
	if hash == password {
		t.Error("HashPassword() returned the plaintext password unchanged")
	}
}

func TestCheckPassword_Correct(t *testing.T) {
	password := "MyP@ssword1"
	hash, _ := service.HashPassword(password)

	if !service.CheckPassword(password, hash) {
		t.Error("CheckPassword() returned false for a correct password")
	}
}

func TestCheckPassword_Wrong(t *testing.T) {
	hash, _ := service.HashPassword("MyP@ssword1")

	if service.CheckPassword("WrongPassword1!", hash) {
		t.Error("CheckPassword() returned true for an incorrect password")
	}
}

// ─── JWT ─────────────────────────────────────────────────────────────────────

func testConfig() *config.Config {
	return &config.Config{
		JWTSecret:      "test-secret-key-that-is-32-chars!!", // 36 chars
		JWTExpiryHours: 24,
		Env:            "development",
	}
}

func TestGenerateToken_ValidToken(t *testing.T) {
	cfg := testConfig()

	token, err := service.GenerateToken(42, "user", cfg)
	if err != nil {
		t.Fatalf("GenerateToken() error: %v", err)
	}
	if token == "" {
		t.Fatal("GenerateToken() returned empty string")
	}

	claims, err := service.ParseToken(token, cfg)
	if err != nil {
		t.Fatalf("ParseToken() error on valid token: %v", err)
	}
	if claims.UserID != 42 {
		t.Errorf("ParseToken() UserID = %d; want 42", claims.UserID)
	}
	if claims.Role != "user" {
		t.Errorf("ParseToken() Role = %q; want %q", claims.Role, "user")
	}
}

func TestGenerateToken_ExpiredToken(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:      "test-secret-key-that-is-32-chars!!",
		JWTExpiryHours: 0, // zero-hour expiry → already expired
		Env:            "development",
	}

	// We need to sign with a negative expiry to simulate an expired token.
	// Use a 1-nanosecond window so it expires immediately.
	// Build the token manually using the same function — the expiry will be in
	// the past by the time ParseToken is called.
	// Easiest: temporarily override expiry to -1h.
	expiredCfg := &config.Config{
		JWTSecret:      cfg.JWTSecret,
		JWTExpiryHours: -1, // negative → token already expired
		Env:            "development",
	}

	token, err := service.GenerateToken(1, "user", expiredCfg)
	if err != nil {
		t.Fatalf("GenerateToken() error: %v", err)
	}

	_, err = service.ParseToken(token, cfg)
	if err == nil {
		t.Error("ParseToken() should have returned an error for an expired token")
	}
}

func TestGenerateToken_TamperedToken(t *testing.T) {
	cfg := testConfig()

	token, _ := service.GenerateToken(1, "user", cfg)

	// Flip the last character to tamper with the signature.
	tampered := token[:len(token)-1] + "X"

	_, err := service.ParseToken(tampered, cfg)
	if err == nil {
		t.Error("ParseToken() should have returned an error for a tampered token")
	}
}

func TestGenerateToken_WrongSecret(t *testing.T) {
	cfg := testConfig()
	wrongCfg := &config.Config{
		JWTSecret:      "completely-different-secret-key!",
		JWTExpiryHours: 24,
		Env:            "development",
	}

	token, _ := service.GenerateToken(1, "user", cfg)

	_, err := service.ParseToken(token, wrongCfg)
	if err == nil {
		t.Error("ParseToken() should have returned an error when using wrong secret")
	}
}


