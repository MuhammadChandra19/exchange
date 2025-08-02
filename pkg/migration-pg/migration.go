package migrationpg

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/muhammadchandra19/exchange/pkg/postgresql"
)

// Migration represents a database migration
type Migration struct {
	ID        string
	Name      string
	Timestamp time.Time
	UpSQL     string
	DownSQL   string
}

// Runner handles PostgreSQL migration execution
type Runner struct {
	client       postgresql.PostgreSQLClient
	ctx          context.Context
	migrationDir string
	schema       string
	tableName    string
}

// Config for migration runner
type Config struct {
	MigrationDir string
	Schema       string // PostgreSQL schema name (default: "public")
	TableName    string // Migration table name (default: "schema_migrations")
}

// NewRunner creates a new migration runner for PostgreSQL
func NewRunner(ctx context.Context, client postgresql.PostgreSQLClient, config Config) *Runner {
	if config.Schema == "" {
		config.Schema = "public"
	}
	if config.TableName == "" {
		config.TableName = "schema_migrations"
	}

	return &Runner{
		client:       client,
		ctx:          ctx,
		migrationDir: config.MigrationDir,
		schema:       config.Schema,
		tableName:    config.TableName,
	}
}

// EnsureMigrationTable creates the schema_migrations table if it doesn't exist
func (r *Runner) EnsureMigrationTable() error {
	createTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.%s (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);
	`, r.schema, r.tableName)

	_, err := r.client.Exec(r.ctx, createTableSQL)
	return err
}

// GetAppliedMigrations returns a map of applied migration IDs
func (r *Runner) GetAppliedMigrations() (map[string]bool, error) {
	applied := make(map[string]bool)

	query := fmt.Sprintf("SELECT id FROM %s.%s ORDER BY applied_at", r.schema, r.tableName)
	rows, err := r.client.Query(r.ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		applied[id] = true
	}

	return applied, nil
}

// LoadMigrations loads all migration files from the migration directory
func (r *Runner) LoadMigrations() ([]Migration, error) {
	// Look for .up.sql files to identify migration base names
	upFiles, err := filepath.Glob(filepath.Join(r.migrationDir, "*.up.sql"))
	if err != nil {
		return nil, err
	}

	sort.Strings(upFiles)

	var migrations []Migration
	for _, upFile := range upFiles {
		migration, err := r.parseMigrationFiles(upFile)
		if err != nil {
			return nil, fmt.Errorf("failed to parse migration %s: %v", upFile, err)
		}
		migrations = append(migrations, migration)
	}

	return migrations, nil
}

// parseMigrationFiles parses UP and DOWN migration files
func (r *Runner) parseMigrationFiles(upFilePath string) (Migration, error) {
	// Read UP file
	upContent, err := os.ReadFile(upFilePath)
	if err != nil {
		return Migration{}, err
	}

	// Determine base name and construct down file path
	fileName := filepath.Base(upFilePath)
	id := strings.TrimSuffix(fileName, ".up.sql")
	downFilePath := strings.Replace(upFilePath, ".up.sql", ".down.sql", 1)

	// Parse timestamp from filename (format: YYYYMMDDHHMMSS_name)
	parts := strings.SplitN(id, "_", 2)
	timestampStr := parts[0]
	var name string
	if len(parts) > 1 {
		name = parts[1]
	} else {
		name = id
	}

	timestamp, err := time.Parse("20060102150405", timestampStr)
	if err != nil {
		// Fallback for files like "001_initial"
		timestamp = time.Unix(0, 0)
	}

	upSQL := strings.TrimSpace(string(upContent))

	// Read DOWN file if it exists
	var downSQL string
	if downContent, err := os.ReadFile(downFilePath); err == nil {
		downSQL = strings.TrimSpace(string(downContent))
	}

	return Migration{
		ID:        id,
		Name:      name,
		Timestamp: timestamp,
		UpSQL:     upSQL,
		DownSQL:   downSQL,
	}, nil
}

// MigrateUp applies pending migrations
func (r *Runner) MigrateUp(steps int) error {
	migrations, err := r.LoadMigrations()
	if err != nil {
		return err
	}

	applied, err := r.GetAppliedMigrations()
	if err != nil {
		return err
	}

	var toApply []Migration
	for _, migration := range migrations {
		if !applied[migration.ID] {
			toApply = append(toApply, migration)
		}
	}

	if steps > 0 && len(toApply) > steps {
		toApply = toApply[:steps]
	}

	for _, migration := range toApply {
		fmt.Printf("Applying migration: %s\n", migration.ID)

		if migration.UpSQL == "" {
			fmt.Printf("Warning: No UP SQL found for migration %s\n", migration.ID)
			continue
		}

		// Execute migration in transaction
		err := postgresql.WithTx(r.ctx, r.client, func(txCtx context.Context) error {
			if _, err := r.client.Exec(txCtx, migration.UpSQL); err != nil {
				return err
			}

			// Record migration as applied
			recordSQL := fmt.Sprintf(
				"INSERT INTO %s.%s (id, name, applied_at) VALUES ($1, $2, NOW())",
				r.schema, r.tableName,
			)
			_, err := r.client.Exec(txCtx, recordSQL, migration.ID, migration.Name)
			return err
		})

		if err != nil {
			return fmt.Errorf("failed to apply migration %s: %v", migration.ID, err)
		}

		fmt.Printf("Applied migration: %s\n", migration.ID)
	}

	return nil
}

// MigrateDown reverts applied migrations
func (r *Runner) MigrateDown(steps int) error {
	if steps <= 0 {
		return fmt.Errorf("steps must be greater than 0 for down migrations")
	}

	migrations, err := r.LoadMigrations()
	if err != nil {
		return err
	}

	applied, err := r.GetAppliedMigrations()
	if err != nil {
		return err
	}

	// Get applied migrations in reverse order
	var toRevert []Migration
	for i := len(migrations) - 1; i >= 0; i-- {
		migration := migrations[i]
		if applied[migration.ID] {
			toRevert = append(toRevert, migration)
			if len(toRevert) >= steps {
				break
			}
		}
	}

	for _, migration := range toRevert {
		fmt.Printf("Reverting migration: %s\n", migration.ID)

		if migration.DownSQL == "" {
			return fmt.Errorf("no DOWN SQL found for migration %s - cannot revert", migration.ID)
		}

		// Execute rollback in transaction
		err := postgresql.WithTx(r.ctx, r.client, func(txCtx context.Context) error {
			if _, err := r.client.Exec(txCtx, migration.DownSQL); err != nil {
				return err
			}

			// Remove migration record
			removeSQL := fmt.Sprintf("DELETE FROM %s.%s WHERE id = $1", r.schema, r.tableName)
			_, err := r.client.Exec(txCtx, removeSQL, migration.ID)
			return err
		})

		if err != nil {
			return fmt.Errorf("failed to revert migration %s: %v", migration.ID, err)
		}

		fmt.Printf("Reverted migration: %s\n", migration.ID)
	}

	return nil
}
