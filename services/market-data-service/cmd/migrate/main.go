package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/muhammadchandra19/exchange/services/market-data-service/internal/infrastructure/questdb"
	"github.com/muhammadchandra19/exchange/services/market-data-service/pkg/config"
)

func main() {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize QuestDB client
	client, err := questdb.NewClient(ctx, cfg.QuestDB)
	if err != nil {
		log.Fatalf("Failed to initialize QuestDB client: %v", err)
	}
	defer client.Close()

	// Run migrations
	if err := runMigrations(ctx, client); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Migrations completed successfully")
}

func runMigrations(ctx context.Context, client questdb.QuestDBClient) error {
	migrationDir := "internal/infrastructure/questdb/migrations"

	files, err := filepath.Glob(filepath.Join(migrationDir, "*.sql"))
	if err != nil {
		return err
	}

	sort.Strings(files)

	for _, file := range files {
		log.Printf("Running migration: %s", file)

		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		if err := client.Exec(ctx, string(content)); err != nil {
			return err
		}

		log.Printf("Migration completed: %s", file)
	}

	return nil
}
