package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	DatabaseURL    string
	Port           string
	Env            string
	JWTSecret      string
	JWTExpiryHours int
}

// Load reads .env (if present) then environment variables into a Config.
// It panics if DATABASE_URL or JWT_SECRET is not set / too short.
func Load() *Config {
	// Best-effort: load .env if it exists. Ignore error so the app works
	// when env vars are injected by the platform (e.g. Docker, CI).
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}
	if len(jwtSecret) < 32 {
		log.Fatal("JWT_SECRET must be at least 32 characters long")
	}

	jwtExpiryHours := 24
	if raw := os.Getenv("JWT_EXPIRY_HOURS"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			jwtExpiryHours = parsed
		}
	}

	return &Config{
		DatabaseURL:    dbURL,
		Port:           port,
		Env:            env,
		JWTSecret:      jwtSecret,
		JWTExpiryHours: jwtExpiryHours,
	}
}
