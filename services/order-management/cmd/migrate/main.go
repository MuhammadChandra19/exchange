package main

import (
	"context"
	"flag"
	"log"

	migration "github.com/muhammadchandra19/exchange/pkg/migration-pg"
	"github.com/muhammadchandra19/exchange/pkg/postgresql"
	"github.com/muhammadchandra19/exchange/service/order-management/pkg/config"
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

	// Initialize PostgreSQL client
	pgClient, err := postgresql.NewClient(ctx, cfg.PostgreSQL)
	if err != nil {
		log.Fatalf("Failed to initialize PostgreSQL client: %v", err)
	}
	defer pgClient.Close()

	// Create migration runner
	migrationConfig := migration.Config{
		MigrationDir: "internal/infrastructure/postgresql/migrations",
		Schema:       "public",
		TableName:    "schema_migrations",
	}
	runner := migration.NewRunner(ctx, pgClient, migrationConfig)

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
