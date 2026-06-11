package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Maniii97/aiynx-go/config"
	"github.com/jackc/pgx/v5"
)

func main() {
	// Load config (loads .env automatically via godotenv.Load())
	cfg := config.Load()

	fmt.Println("Connecting to database...")
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(ctx)

	// Read migration file
	migrationPath := filepath.Join("db", "migrations", "000001_create_users.sql")
	fmt.Printf("Reading migration file: %s\n", migrationPath)
	content, err := os.ReadFile(migrationPath)
	if err != nil {
		log.Fatalf("Unable to read migration file: %v\n", err)
	}

	fmt.Println("Applying migration...")
	_, err = conn.Exec(ctx, string(content))
	if err != nil {
		log.Fatalf("Failed to execute migration: %v\n", err)
	}

	fmt.Println("Migration applied successfully!")
}
