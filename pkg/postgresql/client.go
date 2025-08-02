package postgresql

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Client is the PostgreSQL client.
type Client struct {
	pool   *pgxpool.Pool
	config Config
}

// Config is the PostgreSQL client configuration.
type Config struct {
	Host     string `env:"HOST" envDefault:"localhost"`
	Port     int    `env:"PORT" envDefault:"5432"`
	Database string `env:"DATABASE" envDefault:"exchange"`
	Username string `env:"USERNAME" envDefault:"postgres"`
	Password string `env:"PASSWORD" envDefault:""`

	// SSL configuration
	SSLMode     string `env:"SSL_MODE" envDefault:"prefer"`
	SSLCert     string `env:"SSL_CERT"`
	SSLKey      string `env:"SSL_KEY"`
	SSLRootCert string `env:"SSL_ROOT_CERT"`

	// Connection pool settings optimized for OLTP
	MaxConns        int32         `env:"MAX_CONNS" envDefault:"50"`
	MinConns        int32         `env:"MIN_CONNS" envDefault:"10"`
	MaxConnLifetime time.Duration `env:"MAX_CONN_LIFETIME" envDefault:"2h"`
	MaxConnIdleTime time.Duration `env:"MAX_CONN_IDLE_TIME" envDefault:"15m"`

	// Connection timeout settings
	ConnectTimeout time.Duration `env:"CONNECT_TIMEOUT" envDefault:"5s"`
	QueryTimeout   time.Duration `env:"QUERY_TIMEOUT" envDefault:"30s"`

	// Statement cache settings for better performance
	StatementCacheCapacity int `env:"STATEMENT_CACHE_CAPACITY" envDefault:"512"`

	// Application name for connection tracking
	ApplicationName string `env:"APPLICATION_NAME" envDefault:"exchange-api"`

	// Search path
	SearchPath string `env:"SEARCH_PATH" envDefault:"public"`
}

// Ensure Client implements PostgreSQLClient interface
var _ PostgreSQLClient = (*Client)(nil)

// NewClient creates a new PostgreSQL client optimized for OLTP workloads.
func NewClient(ctx context.Context, config Config) (PostgreSQLClient, error) {
	// Build connection string
	connString := buildConnectionString(config)

	// Parse config
	pgxConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgresql config: %w", err)
	}

	// Set pool configuration optimized for OLTP
	pgxConfig.MaxConns = config.MaxConns
	pgxConfig.MinConns = config.MinConns
	pgxConfig.MaxConnLifetime = config.MaxConnLifetime
	pgxConfig.MaxConnIdleTime = config.MaxConnIdleTime
	pgxConfig.ConnConfig.ConnectTimeout = config.ConnectTimeout

	// Enable statement caching for better performance
	pgxConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	pgxConfig.ConnConfig.StatementCacheCapacity = config.StatementCacheCapacity

	// Set application name for monitoring
	if config.ApplicationName != "" {
		pgxConfig.ConnConfig.RuntimeParams["application_name"] = config.ApplicationName
	}

	// Set search path
	if config.SearchPath != "" {
		pgxConfig.ConnConfig.RuntimeParams["search_path"] = config.SearchPath
	}

	// Create connection pool
	pool, err := pgxpool.New(ctx, pgxConfig.ConnString())
	if err != nil {
		return nil, fmt.Errorf("failed to create postgresql pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping postgresql: %w", err)
	}

	return &Client{
		pool:   pool,
		config: config,
	}, nil
}

// buildConnectionString constructs the PostgreSQL connection string
func buildConnectionString(config Config) string {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	// Add SSL mode
	connString += fmt.Sprintf("?sslmode=%s", config.SSLMode)

	// Add SSL certificates if provided
	if config.SSLCert != "" {
		connString += fmt.Sprintf("&sslcert=%s", config.SSLCert)
	}
	if config.SSLKey != "" {
		connString += fmt.Sprintf("&sslkey=%s", config.SSLKey)
	}
	if config.SSLRootCert != "" {
		connString += fmt.Sprintf("&sslrootcert=%s", config.SSLRootCert)
	}

	return connString
}

// Pool returns the connection pool.
func (c *Client) Pool() *pgxpool.Pool {
	return c.pool
}

// Stats returns connection pool statistics for monitoring.
func (c *Client) Stats() *pgxpool.Stat {
	return c.pool.Stat()
}

// DatabaseName returns the database name.
func (c *Client) DatabaseName() string {
	return c.config.Database
}

// Host returns the host.
func (c *Client) Host() string {
	return c.config.Host
}

// Port returns the port.
func (c *Client) Port() int {
	return c.config.Port
}

// Close closes the connection pool.
func (c *Client) Close() {
	if c.pool != nil {
		c.pool.Close()
	}
}

// Ping pings the connection pool.
func (c *Client) Ping(ctx context.Context) error {
	return c.pool.Ping(ctx)
}

// Acquire acquires a connection from the pool for advanced operations.
func (c *Client) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	return c.pool.Acquire(ctx)
}

// Exec executes a query without returning any rows
func (c *Client) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if tx, ok := GetTx(ctx); ok {
		return tx.Exec(ctx, sql, args...)
	}
	return c.pool.Exec(ctx, sql, args...)
}

// Query executes a query that returns rows.
func (c *Client) Query(ctx context.Context, sql string, args ...any) (RowsInterface, error) {
	if tx, ok := GetTx(ctx); ok {
		rows, err := tx.Query(ctx, sql, args...)
		if err != nil {
			return nil, err
		}
		return NewRowsWrapper(rows), nil
	}

	rows, err := c.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return NewRowsWrapper(rows), nil
}

// QueryRow executes a query that is expected to return at most one row
func (c *Client) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if tx, ok := GetTx(ctx); ok {
		return tx.QueryRow(ctx, sql, args...)
	}
	return c.pool.QueryRow(ctx, sql, args...)
}

// Begin starts a transaction
func (c *Client) Begin(ctx context.Context) (pgx.Tx, error) {
	return c.pool.Begin(ctx)
}

// BeginTx starts a transaction with specific options
func (c *Client) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return c.pool.BeginTx(ctx, txOptions)
}

// SendBatch sends a batch of queries for better performance
func (c *Client) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	if tx, ok := GetTx(ctx); ok {
		return tx.SendBatch(ctx, b)
	}
	return c.pool.SendBatch(ctx, b)
}

// CopyFrom performs efficient bulk inserts using PostgreSQL COPY
func (c *Client) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	if tx, ok := GetTx(ctx); ok {
		return tx.CopyFrom(ctx, tableName, columnNames, rowSrc)
	}
	return c.pool.CopyFrom(ctx, tableName, columnNames, rowSrc)
}

// Prepare creates a prepared statement for reuse
func (c *Client) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	conn, err := c.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	return conn.Conn().Prepare(ctx, name, sql)
}

// Deallocate removes a prepared statement
func (c *Client) Deallocate(ctx context.Context, name string) error {
	conn, err := c.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	return conn.Conn().Deallocate(ctx, name)
}
