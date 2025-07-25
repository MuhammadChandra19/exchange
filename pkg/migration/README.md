# Migration Package

A reusable database migration package for the exchange monorepo that supports up/down migrations with step-based migration control.

## Features

- ✅ **Separate UP/DOWN files** - Clean separation of migration logic
- ✅ **Step-based migration** - Migrate up/down by specific number of steps  
- ✅ **Migration tracking** - Persistent tracking via `schema_migrations` table
- ✅ **QuestDB integration** - Direct integration with existing `pkg/questdb`
- ✅ **Rollback support** - Safe rollback with DOWN migrations
- ✅ **Reusable across services** - Any service can use the same migration logic

## Usage

### Basic Setup

```go
import (
    "context"
    "github.com/muhammadchandra19/exchange/pkg/migration"
    "github.com/muhammadchandra19/exchange/pkg/questdb"
)

func main() {
    ctx := context.Background()
    
    // Initialize your database client (QuestDB example)
    questdbClient, err := questdb.NewClient(ctx, config)
    if err != nil {
        log.Fatal(err)
    }
    defer questdbClient.Close()

    // Create migration runner directly with questdb client
    migrationDir := "path/to/your/migrations"
    runner := migration.NewRunner(ctx, questdbClient, migrationDir)

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

**Example UP file** (`20250725211616_add_indexes.up.sql`):
```sql
-- Migration: add_indexes
-- Created at: Sat Jul 25 21:16:16 WIB 2025

CREATE TABLE users (
    id STRING,
    name STRING,
    email STRING
);
```

**Example DOWN file** (`20250725211616_add_indexes.down.sql`):
```sql
-- Migration: add_indexes
-- Created at: Sat Jul 25 21:16:16 WIB 2025

DROP TABLE IF EXISTS users;
```

### Methods

#### `NewRunner(ctx, dbClient, migrationDir) *Runner`
Creates a new migration runner instance.

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

#### QuestDB Integration
The migration package directly uses the existing `pkg/questdb` interfaces:
```go
// Uses questdb.QuestDBClient directly - no adapter needed
runner := migration.NewRunner(ctx, questdbClient, migrationDir)
```

#### Custom Database Integration
To support other databases, update the Runner to accept your database client interface directly, or create a wrapper that implements the required methods (`Exec`, `Query`, `Close`).

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

1. Import the migration package
2. Initialize your QuestDB client  
3. Create the migration runner with your migration directory
4. Use the migration methods

Since it reuses the existing `pkg/questdb`, there's no additional setup needed.

## Migration Tracking

The package automatically creates and manages a `schema_migrations` table:

```sql
CREATE TABLE schema_migrations (
    id STRING,              -- Migration ID (filename without extension)
    name STRING,            -- Migration name (description part)
    applied_at TIMESTAMP    -- When the migration was applied
) TIMESTAMP(applied_at) PARTITION BY DAY;
``` 