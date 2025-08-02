package main

import (
	"context"
	"flag"
	"log"

	"github.com/muhammadchandra19/exchange/pkg/migration"
	"github.com/muhammadchandra19/exchange/pkg/questdb"
	"github.com/muhammadchandra19/exchange/services/market-data/pkg/config"
)

func main() {
	var (
		direction = flag.String("direction", "up", "Migration direction: up or down")
		steps     = flag.Int("steps", 0, "Number of steps to migrate (0 = all)")
	)
	flag.Parse()

	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize QuestDB client
	questdbClient, err := questdb.NewClient(ctx, cfg.QuestDB)
	if err != nil {
		log.Fatalf("Failed to initialize QuestDB client: %v", err)
	}
	defer questdbClient.Close()

	// Create migration runner
	migrationDir := "internal/infrastructure/questdb/migrations"
	runner := migration.NewRunner(ctx, questdbClient, migrationDir)

	// Ensure migration tracking table exists
	if err := runner.EnsureMigrationTable(); err != nil {
		log.Fatalf("Failed to create migration table: %v", err)
	}

	// Run migrations based on direction
	switch *direction {
	case "up":
		if err := runner.MigrateUp(*steps); err != nil {
			log.Fatalf("Failed to migrate up: %v", err)
		}
	case "down":
		if err := runner.MigrateDown(*steps); err != nil {
			log.Fatalf("Failed to migrate down: %v", err)
		}
	default:
		log.Fatalf("Invalid direction: %s. Use 'up' or 'down'", *direction)
	}

	log.Printf("Migration %s completed successfully", *direction)
}
