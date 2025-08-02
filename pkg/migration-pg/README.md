# PostgreSQL Migration Package

A reusable PostgreSQL migration package for the exchange monorepo that supports up/down migrations with step-based migration control.

## Features

- ✅ **Separate UP/DOWN files** - Clean separation of migration logic
- ✅ **Step-based migration** - Migrate up/down by specific number of steps  
- ✅ **Migration tracking** - Persistent tracking via `schema_migrations` table
- ✅ **PostgreSQL integration** - Direct integration with existing `pkg/postgresql`
- ✅ **Transaction safety** - Each migration runs in its own transaction
- ✅ **Rollback support** - Safe rollback with DOWN migrations
- ✅ **Reusable across services** - Any service can use the same migration logic

## Usage

### Basic Setup

```go
import (
    "context"
    "github.com/muhammadchandra19/exchange/pkg/migration-pg"
    "github.com/muhammadchandra19/exchange/pkg/postgresql"
)

func main() {
    ctx := context.Background()
    
    // Initialize your database client (PostgreSQL example)
    pgConfig := postgresql.Config{
        Host:     "localhost",
        Port:     5432,
        Database: "exchange",
        Username: "postgres",
        Password: "password",
    }
    
    pgClient, err := postgresql.NewClient(ctx, pgConfig)
    if err != nil {
        log.Fatal(err)
    }
    defer pgClient.Close()

    // Create migration runner directly with postgresql client
    migrationConfig := migration.Config{
        MigrationDir: "path/to/your/migrations",
        Schema:       "public",
        TableName:    "schema_migrations",
    }
    runner := migration.NewRunner(ctx, pgClient, migrationConfig)

    // Ensure migration tracking table exists
    if err := runner.EnsureMigrationTable(); err != nil {
        log.Fatal(err)
    }

    // Run migrations
    if err := runner.MigrateUp(0); err != nil { // 0 = all pending
        log.Fatal(err)
    }
}
```

### Migration File Format

Migration files use the naming convention: `YYYYMMDDHHMMSS_description.up.sql` and `YYYYMMDDHHMMSS_description.down.sql`

**Example UP file** (`20250125120000_create_orders_table.up.sql`):
```sql
-- Migration: create_orders_table
-- Created at: Sat Jan 25 12:00:00 WIB 2025

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(4) NOT NULL CHECK (side IN ('buy', 'sell')),
    order_type VARCHAR(10) NOT NULL CHECK (order_type IN ('limit', 'market')),
    price DECIMAL(20,8),
    quantity DECIMAL(20,8) NOT NULL CHECK (quantity > 0),
    filled_quantity DECIMAL(20,8) DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_orders_user_id ON orders (user_id);
CREATE INDEX idx_orders_symbol_status ON orders (symbol, status);
```

**Example DOWN file** (`20250125120000_create_orders_table.down.sql`):
```sql
-- Migration: create_orders_table
-- Created at: Sat Jan 25 12:00:00 WIB 2025

DROP TABLE IF EXISTS orders CASCADE;
```

### Methods

#### `NewRunner(ctx, pgClient, config) *Runner`
Creates a new migration runner instance with PostgreSQL client.

#### `EnsureMigrationTable() error`
Creates the `schema_migrations` tracking table if it doesn't exist.

#### `MigrateUp(steps int) error`
Applies pending migrations. Use `steps=0` to apply all pending migrations.

#### `MigrateDown(steps int) error`
Reverts applied migrations. `steps` must be > 0.

#### `LoadMigrations() ([]Migration, error)`
Loads all migration files from the migration directory.

#### `GetAppliedMigrations() (map[string]bool, error)`
Returns a map of applied migration IDs.

### Database Support

#### PostgreSQL Integration
The migration package directly uses the existing `pkg/postgresql` interfaces:
```go
// Uses postgresql.PostgreSQLClient directly - no adapter needed
runner := migration.NewRunner(ctx, pgClient, migrationConfig)
```

#### Transaction Safety
Each migration runs in its own transaction for data safety:
```go
// Each migration is wrapped in:
BEGIN;
-- Migration SQL here
-- Record migration success
COMMIT;
-- On error: automatic ROLLBACK
```

## Command Line Usage

The package is designed to be used with a command-line tool:

```bash
# Apply all pending migrations
go run cmd/migrate/main.go -direction=up -steps=0

# Apply next 3 migrations
go run cmd/migrate/main.go -direction=up -steps=3

# Rollback last 2 migrations  
go run cmd/migrate/main.go -direction=down -steps=2
```

## Integration with Other Services

This package can be used by any service in the monorepo. Simply:

1. Import the migration-pg package
2. Initialize your PostgreSQL client  
3. Create the migration runner with your migration directory
4. Use the migration methods

Since it reuses the existing `pkg/postgresql`, there's no additional setup needed.

## Migration Tracking

The package automatically creates and manages a `schema_migrations` table:

```sql
CREATE TABLE schema_migrations (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

## PostgreSQL Features

### Schema Support
```go
migrationConfig := migration.Config{
    MigrationDir: "migrations",
    Schema:       "order_management", // Custom schema
    TableName:    "migrations",       // Custom table name
}
```

### Advanced SQL Support
PostgreSQL migrations support advanced features:

```sql
-- Constraints and checks
ALTER TABLE orders ADD CONSTRAINT chk_positive_quantity 
    CHECK (quantity > 0);

-- Enum types
CREATE TYPE order_status AS ENUM ('pending', 'filled', 'cancelled');

-- JSON/JSONB support
ALTER TABLE orders ADD COLUMN metadata JSONB;
CREATE INDEX idx_orders_metadata_gin ON orders USING GIN (metadata);

-- Concurrent index creation (non-blocking)
CREATE INDEX CONCURRENTLY idx_orders_symbol ON orders (symbol);
```

## Makefile Integration

Each service can include migration commands in their Makefile:

```makefile
# Variables
MIGRATION_DIR=internal/infrastructure/postgresql/migrations

.PHONY: migrate
migrate: ## Run all pending migrations
	@go run cmd/migrate/main.go -direction=up -steps=0

.PHONY: migrate-up
migrate-up: ## Run pending migrations up. Usage: make migrate-up [steps=N]
	@if [ -z "$(steps)" ]; then \
		go run cmd/migrate/main.go -direction=up -steps=0; \
	else \
		go run cmd/migrate/main.go -direction=up -steps=$(steps); \
	fi

.PHONY: migrate-down  
migrate-down: ## Run migrations down. Usage: make migrate-down steps=N
	@go run cmd/migrate/main.go -direction=down -steps=$(steps)

.PHONY: migration
migration: ## Create a new migration file. Usage: make migration name=create_table
	@timestamp=$$(date +%Y%m%d%H%M%S); \
	basename="$${timestamp}_$(name)"; \
	upfile="$(MIGRATION_DIR)/$${basename}.up.sql"; \
	downfile="$(MIGRATION_DIR)/$${basename}.down.sql"; \
	mkdir -p $(MIGRATION_DIR); \
	echo "-- Migration: $(name)" > $${upfile}; \
	echo "-- Created at: $$(date)" >> $${upfile}; \
	echo "" >> $${upfile}; \
	echo "-- Write your UP migration SQL here" >> $${upfile}; \
	echo "-- Migration: $(name)" > $${downfile}; \
	echo "-- Created at: $$(date)" >> $${downfile}; \
	echo "" >> $${downfile}; \
	echo "-- Write your DOWN migration SQL here" >> $${downfile}; \
	echo "Migration files created:"; \
	echo "  UP:   $${upfile}"; \
	echo "  DOWN: $${downfile}"
```

### Usage Examples

```bash
# Apply all pending migrations
make migrate

# Apply specific number of migrations
make migrate-up steps=3

# Rollback migrations
make migrate-down steps=2

# Create new migration
make migration name=add_user_preferences
```

## Example Service Integration

```go
// services/order-management-service/cmd/migrate/main.go
package main

import (
    "context"
    "flag"
    "log"

    migration "github.com/muhammadchandra19/exchange/pkg/migration-pg"
    "github.com/muhammadchandra19/exchange/pkg/postgresql"
    "github.com/muhammadchandra19/exchange/services/order-management-service/pkg/config"
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
```

## Comparison with QuestDB Migration

| Feature | pkg/migration (QuestDB) | pkg/migration-pg (PostgreSQL) |
|---------|-------------------------|--------------------------------|
| Interface | ✅ Same simple interface | ✅ Same simple interface |
| Commands | ✅ Same commands | ✅ Same commands |
| File naming | ✅ Same pattern | ✅ Same pattern |
| Transaction safety | ⚠️ Basic | ✅ Full ACID transactions |
| Schema support | ❌ No schemas | ✅ PostgreSQL schemas |
| Advanced features | ❌ Limited | ✅ Constraints, enums, JSON |
| Rollback safety | ✅ Basic | ✅ Transaction-based |

## Best Practices

1. **Always provide DOWN migrations** for rollback capability
2. **Test migrations** on development database before production
3. **Use descriptive names** for migrations (e.g., `create_orders_table`)
4. **Keep migrations small** - one logical change per migration
5. **Use transactions** - leverage automatic transaction wrapping
6. **Create indexes concurrently** in production to avoid blocking

---

**Built to complement the existing `pkg/migration` (QuestDB) with the same simplicity and familiar interface.**