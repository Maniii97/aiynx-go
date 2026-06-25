package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

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

	// Discover all migration files in order.
	migrationsDir := filepath.Join("db", "migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("Unable to read migrations directory: %v\n", err)
	}

	// Collect .sql files and sort lexicographically (000001 < 000002 …).
	var files []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".sql" {
			files = append(files, filepath.Join(migrationsDir, e.Name()))
		}
	}
	sort.Strings(files)

	for _, path := range files {
		fmt.Printf("Applying migration: %s\n", path)
		content, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("Unable to read migration file %s: %v\n", path, err)
		}
		if _, err = conn.Exec(ctx, string(content)); err != nil {
			log.Fatalf("Failed to execute migration %s: %v\n", path, err)
		}
		fmt.Printf("  ✓ applied %s\n", filepath.Base(path))
	}

	fmt.Println("All migrations applied successfully!")
}
