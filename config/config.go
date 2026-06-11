package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	DatabaseURL string
	Port        string
	Env         string
}

// Load reads .env (if present) then environment variables into a Config.
// It panics if DATABASE_URL is not set.
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

	return &Config{
		DatabaseURL: dbURL,
		Port:        port,
		Env:         env,
	}
}
