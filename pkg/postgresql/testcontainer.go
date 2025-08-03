package postgresql

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestContainer wraps a PostgreSQL testcontainer with utilities
type TestContainer struct {
	Container testcontainers.Container
	Client    PostgreSQLClient
	ConnStr   string
	ctx       context.Context
}

// TestContainerConfig holds configuration for the test container
type TestContainerConfig struct {
	Image             string
	Database          string
	Username          string
	Password          string
	MigrationsPath    string // Path to migration files
	StartupTimeout    time.Duration
	InitScripts       []string          // SQL scripts to run on startup
	SkipMigrations    bool              // Skip running migrations
	ContainerName     string            // Optional container name
	ExtraEnvVars      map[string]string // Additional environment variables
	MigrationPattern  string            // Pattern to match migration files (default: "*.up.sql")
	SchemaSearchPaths []string          // Additional paths to search for migrations
}

// DefaultTestContainerConfig returns a default configuration
func DefaultTestContainerConfig() *TestContainerConfig {
	return &TestContainerConfig{
		Image:             "postgres:15-alpine",
		Database:          "test_db",
		Username:          "test_user",
		Password:          "test_pass",
		StartupTimeout:    5 * time.Minute,
		InitScripts:       []string{},
		SkipMigrations:    false,
		ExtraEnvVars:      make(map[string]string),
		MigrationPattern:  "*.up.sql",
		SchemaSearchPaths: []string{},
	}
}

// NewTestContainer creates and starts a new PostgreSQL test container
func NewTestContainer(ctx context.Context, config *TestContainerConfig) (*TestContainer, error) {
	if config == nil {
		config = DefaultTestContainerConfig()
	}

	// Build container request
	req := []testcontainers.ContainerCustomizer{
		testcontainers.WithImage(config.Image),
		postgres.WithDatabase(config.Database),
		postgres.WithUsername(config.Username),
		postgres.WithPassword(config.Password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(config.StartupTimeout),
		),
	}

	// Add container name if specified
	if config.ContainerName != "" {
		req = append(req, testcontainers.WithName(config.ContainerName))
	}

	// Add extra environment variables
	for key, value := range config.ExtraEnvVars {
		req = append(req, testcontainers.WithEnv(map[string]string{key: value}))
	}

	// Add init scripts if provided
	if len(config.InitScripts) > 0 {
		req = append(req, postgres.WithInitScripts(config.InitScripts...))
	}

	// Start the container
	container, err := postgres.RunContainer(ctx, req...)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	// Get connection string
	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	// Create connection pool
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	tc := &TestContainer{
		Container: container,
		Client: &Client{
			pool: pool,
		},
		ConnStr: connStr,
		ctx:     ctx,
	}

	// Run migrations if path is provided and not skipped
	if config.MigrationsPath != "" && !config.SkipMigrations {
		if err := tc.RunMigrations(config.MigrationsPath, config.MigrationPattern); err != nil {
			tc.Close()
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	return tc, nil
}

// Close closes the connection and terminates the container
func (tc *TestContainer) Close() error {
	var errors []string

	if tc.Client != nil {
		tc.Client.Close()
	}

	if tc.Container != nil {
		if err := tc.Container.Terminate(tc.ctx); err != nil {
			errors = append(errors, fmt.Sprintf("failed to terminate container: %v", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// RunMigrations runs SQL migration files from the specified directory
func (tc *TestContainer) RunMigrations(migrationsPath, pattern string) error {
	if tc.Client == nil {
		return fmt.Errorf("database client is not initialized")
	}

	if pattern == "" {
		pattern = "*.up.sql"
	}

	// Check if migrations path exists
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		return fmt.Errorf("migrations path does not exist: %s", migrationsPath)
	}

	// Get all migration files matching the pattern
	migrationFiles, err := tc.getMigrationFiles(migrationsPath, pattern)
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	if len(migrationFiles) == 0 {
		return fmt.Errorf("no migration files found in %s with pattern %s", migrationsPath, pattern)
	}

	// Run each migration file
	for _, file := range migrationFiles {
		filePath := filepath.Join(migrationsPath, file)
		if err := tc.runMigrationFile(filePath); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", file, err)
		}
	}

	return nil
}

// RunMigrationFile runs a single migration file
func (tc *TestContainer) runMigrationFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read migration file %s: %w", filePath, err)
	}

	sqlContent := string(content)

	// Remove comments and empty lines for cleaner processing
	lines := strings.Split(sqlContent, "\n")
	var cleanLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comment-only lines
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}
		cleanLines = append(cleanLines, line)
	}

	if len(cleanLines) == 0 {
		return nil // Empty migration file
	}

	cleanSQL := strings.Join(cleanLines, "\n")

	// Handle multiple statements - split by semicolon but be careful with function definitions
	statements := tc.splitSQLStatements(cleanSQL)

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		_, err := tc.Client.Exec(tc.ctx, stmt)
		if err != nil {
			return fmt.Errorf("failed to execute statement in %s: %w\nStatement: %s",
				filepath.Base(filePath), err, stmt)
		}
	}

	return nil
}

// splitSQLStatements splits SQL content by semicolons, handling function definitions properly
func (tc *TestContainer) splitSQLStatements(sql string) []string {
	var statements []string
	var current strings.Builder
	var inFunction bool
	var functionDepth int

	lines := strings.Split(sql, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Track function definitions to avoid splitting on semicolons inside functions
		upperLine := strings.ToUpper(line)

		if strings.Contains(upperLine, "CREATE FUNCTION") ||
			strings.Contains(upperLine, "CREATE OR REPLACE FUNCTION") ||
			strings.Contains(upperLine, "$$") {
			inFunction = true
			functionDepth++
		}

		if strings.Contains(upperLine, "$$") && inFunction && functionDepth > 0 {
			functionDepth--
			if functionDepth == 0 {
				inFunction = false
			}
		}

		current.WriteString(line)
		current.WriteString("\n")

		// Split on semicolon only if we're not inside a function
		if strings.HasSuffix(line, ";") && !inFunction {
			stmt := strings.TrimSpace(current.String())
			if stmt != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
		}
	}

	// Add any remaining content
	if current.Len() > 0 {
		stmt := strings.TrimSpace(current.String())
		if stmt != "" {
			statements = append(statements, stmt)
		}
	}

	return statements
}

// getMigrationFiles returns sorted list of migration files matching the pattern
func (tc *TestContainer) getMigrationFiles(migrationsPath, pattern string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(migrationsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Get relative path from migrations directory
		relPath, err := filepath.Rel(migrationsPath, path)
		if err != nil {
			return err
		}

		// Check if file matches pattern
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err != nil {
			return err
		}

		if matched {
			files = append(files, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort by filename (assuming timestamp-based naming like 20250101120000_name.up.sql)
	sort.Strings(files)

	return files, nil
}

// RunMigrationsFromMultiplePaths runs migrations from multiple directories
func (tc *TestContainer) RunMigrationsFromMultiplePaths(paths []string, pattern string) error {
	for _, path := range paths {
		if err := tc.RunMigrations(path, pattern); err != nil {
			return fmt.Errorf("failed to run migrations from %s: %w", path, err)
		}
	}
	return nil
}

// TruncateAllTables truncates all tables in the database (useful for test cleanup)
func (tc *TestContainer) TruncateAllTables() error {
	query := `
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public' 
		AND tablename NOT LIKE 'pg_%'
		AND tablename NOT LIKE 'sql_%'
	`

	rows, err := tc.Client.Query(tc.ctx, query)
	if err != nil {
		return fmt.Errorf("failed to get table names: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	if len(tables) == 0 {
		return nil // No tables to truncate
	}

	// Disable foreign key checks temporarily
	_, err = tc.Client.Exec(tc.ctx, "SET session_replication_role = replica")
	if err != nil {
		return fmt.Errorf("failed to disable foreign key checks: %w", err)
	}

	// Truncate all tables
	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)
		_, err = tc.Client.Exec(tc.ctx, query)
		if err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}

	// Re-enable foreign key checks
	_, err = tc.Client.Exec(tc.ctx, "SET session_replication_role = DEFAULT")
	if err != nil {
		return fmt.Errorf("failed to re-enable foreign key checks: %w", err)
	}

	return nil
}

// LoadSQLFile loads and executes a single SQL file
func (tc *TestContainer) LoadSQLFile(filePath string) error {
	return tc.runMigrationFile(filePath)
}

// CreateDatabase creates a new database in the test container
func (tc *TestContainer) CreateDatabase(dbName string) error {
	query := fmt.Sprintf("CREATE DATABASE %s", dbName)
	_, err := tc.Client.Exec(tc.ctx, query)
	return err
}

// DropDatabase drops a database in the test container
func (tc *TestContainer) DropDatabase(dbName string) error {
	query := fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName)
	_, err := tc.Client.Exec(tc.ctx, query)
	return err
}

// CreateUser creates a new user in the test container
func (tc *TestContainer) CreateUser(username, password string) error {
	query := fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s'", username, password)
	_, err := tc.Client.Exec(tc.ctx, query)
	return err
}

// GrantPrivileges grants all privileges on a database to a user
func (tc *TestContainer) GrantPrivileges(database, username string) error {
	query := fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE %s TO %s", database, username)
	_, err := tc.Client.Exec(tc.ctx, query)
	return err
}

// ExecuteSQL executes arbitrary SQL (useful for test setup)
func (tc *TestContainer) ExecuteSQL(sql string) error {
	_, err := tc.Client.Exec(tc.ctx, sql)
	return err
}

// GetConnectionString returns the connection string for the test database
func (tc *TestContainer) GetConnectionString() string {
	return tc.ConnStr
}

// WaitForReady waits for the database to be ready (useful for custom wait strategies)
func (tc *TestContainer) WaitForReady(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(tc.ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for database to be ready")
		case <-ticker.C:
			if err := tc.Client.Ping(ctx); err == nil {
				return nil
			}
		}
	}
}

// GetContainerInfo returns useful information about the container
func (tc *TestContainer) GetContainerInfo() (ContainerInfo, error) {
	host, err := tc.Container.Host(tc.ctx)
	if err != nil {
		return ContainerInfo{}, err
	}

	port, err := tc.Container.MappedPort(tc.ctx, "5432")
	if err != nil {
		return ContainerInfo{}, err
	}

	return ContainerInfo{
		Host:          host,
		Port:          port.Port(),
		ContainerID:   tc.Container.GetContainerID(),
		ConnectionStr: tc.ConnStr,
	}, nil
}

// ContainerInfo holds information about the test container
type ContainerInfo struct {
	Host          string
	Port          string
	ContainerID   string
	ConnectionStr string
}

// String returns a string representation of container info
func (ci ContainerInfo) String() string {
	return fmt.Sprintf("Host: %s, Port: %s, ID: %s", ci.Host, ci.Port, ci.ContainerID)
}
