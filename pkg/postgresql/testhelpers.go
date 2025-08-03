package postgresql

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestHelper provides common testing utilities
type TestHelper struct {
	Container *TestContainer
	T         *testing.T
}

// NewTestHelper creates a new test helper with default configuration
func NewTestHelper(t *testing.T) *TestHelper {
	return NewTestHelperWithConfig(t, nil)
}

// NewTestHelperWithConfig creates a new test helper with custom configuration
func NewTestHelperWithConfig(t *testing.T, config *TestContainerConfig) *TestHelper {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	if config == nil {
		config = DefaultTestContainerConfig()
	}

	container, err := NewTestContainer(ctx, config)
	require.NoError(t, err)

	// Cleanup on test completion
	t.Cleanup(func() {
		if err := container.Close(); err != nil {
			t.Logf("Failed to close test container: %v", err)
		}
	})

	return &TestHelper{
		Container: container,
		T:         t,
	}
}

// NewTestHelperWithMigrations creates a test helper and runs migrations from the specified path
func NewTestHelperWithMigrations(t *testing.T, migrationsPath string) *TestHelper {
	config := DefaultTestContainerConfig()
	config.MigrationsPath = migrationsPath
	return NewTestHelperWithConfig(t, config)
}

// NewTestHelperWithMigrationsAndPattern creates a test helper with custom migration pattern
func NewTestHelperWithMigrationsAndPattern(t *testing.T, migrationsPath, pattern string) *TestHelper {
	config := DefaultTestContainerConfig()
	config.MigrationsPath = migrationsPath
	config.MigrationPattern = pattern
	return NewTestHelperWithConfig(t, config)
}

// CleanupTables truncates all tables between tests
func (h *TestHelper) CleanupTables() {
	err := h.Container.TruncateAllTables()
	require.NoError(h.T, err)
}

// RequireNoError fails the test if error is not nil
func (h *TestHelper) RequireNoError(err error) {
	require.NoError(h.T, err)
}

// RequireEqual fails the test if values are not equal
func (h *TestHelper) RequireEqual(expected, actual interface{}) {
	require.Equal(h.T, expected, actual)
}

// ExecuteSQL executes SQL and fails test on error
func (h *TestHelper) ExecuteSQL(sql string) {
	err := h.Container.ExecuteSQL(sql)
	require.NoError(h.T, err)
}

// LoadSQLFile loads and executes a SQL file
func (h *TestHelper) LoadSQLFile(filePath string) {
	err := h.Container.LoadSQLFile(filePath)
	require.NoError(h.T, err)
}

// RunAdditionalMigrations runs additional migrations from a path
func (h *TestHelper) RunAdditionalMigrations(migrationsPath string) {
	err := h.Container.RunMigrations(migrationsPath, "*.up.sql")
	require.NoError(h.T, err)
}

// RunMigrationsWithPattern runs migrations with a custom pattern
func (h *TestHelper) RunMigrationsWithPattern(migrationsPath, pattern string) {
	err := h.Container.RunMigrations(migrationsPath, pattern)
	require.NoError(h.T, err)
}

// WaitForReady waits for database to be ready with a default timeout
func (h *TestHelper) WaitForReady() {
	err := h.Container.WaitForReady(30 * time.Second)
	require.NoError(h.T, err)
}

// GetClient returns the PostgreSQL client
func (h *TestHelper) GetClient() PostgreSQLClient {
	return h.Container.Client
}

// GetConnectionString returns the connection string
func (h *TestHelper) GetConnectionString() string {
	return h.Container.GetConnectionString()
}

// CreateTestUser creates a test user with the given credentials
func (h *TestHelper) CreateTestUser(username, password string) {
	err := h.Container.CreateUser(username, password)
	require.NoError(h.T, err)
}

// CreateTestDatabase creates a test database
func (h *TestHelper) CreateTestDatabase(dbName string) {
	err := h.Container.CreateDatabase(dbName)
	require.NoError(h.T, err)
}

// GrantTestPrivileges grants privileges to a user on a database
func (h *TestHelper) GrantTestPrivileges(database, username string) {
	err := h.Container.GrantPrivileges(database, username)
	require.NoError(h.T, err)
}
