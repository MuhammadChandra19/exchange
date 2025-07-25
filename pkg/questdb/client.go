package questdb

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Client is the QuestDB client.
type Client struct {
	pool   *pgxpool.Pool
	config Config
}

// Config is the QuestDB client configuration.
type Config struct {
	Host     string `env:"HOST" envDefault:"localhost"`
	Port     int    `env:"PORT" envDefault:"8812"`
	Database string `env:"DATABASE" envDefault:"qdb"`
	Username string `env:"USERNAME" envDefault:"admin"`
	Password string `env:"PASSWORD" envDefault:"quest"`

	// Connection pool settings
	MaxConns        int32         `env:"MAX_CONNS" envDefault:"25"`
	MinConns        int32         `env:"MIN_CONNS" envDefault:"5"`
	MaxConnLifetime time.Duration `env:"MAX_CONN_LIFETIME" envDefault:"1h"`
	MaxConnIdleTime time.Duration `env:"MAX_CONN_IDLE_TIME" envDefault:"30m"`

	// Connection timeout settings
	ConnectTimeout time.Duration `env:"CONNECT_TIMEOUT" envDefault:"10s"`
}

// Ensure Client implements QuestDBClient interface
var _ QuestDBClient = (*Client)(nil)

// NewClient creates a new QuestDB client.
func NewClient(ctx context.Context, config Config) (QuestDBClient, error) {
	// Build connection string
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	// Parse config
	pgxConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse questdb config: %w", err)
	}

	// Set pool configuration
	pgxConfig.MaxConns = config.MaxConns
	pgxConfig.MinConns = config.MinConns
	pgxConfig.MaxConnLifetime = config.MaxConnLifetime
	pgxConfig.MaxConnIdleTime = config.MaxConnIdleTime
	pgxConfig.ConnConfig.ConnectTimeout = config.ConnectTimeout

	// Create connection pool
	pool, err := pgxpool.New(ctx, pgxConfig.ConnString())
	if err != nil {
		return nil, fmt.Errorf("failed to create questdb pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping questdb: %w", err)
	}

	return &Client{
		pool:   pool,
		config: config,
	}, nil
}

// Pool returns the connection pool.
func (c *Client) Pool() *pgxpool.Pool {
	return c.pool
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

// Exec executes a query without returning any rows
func (c *Client) Exec(ctx context.Context, sql string, args ...any) error {
	if tx, ok := GetTx(ctx); ok {
		_, err := tx.Exec(ctx, sql, args...)
		return err
	}
	_, err := c.pool.Exec(ctx, sql, args...)
	return err
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

// CopyFrom wraps the pool's CopyFrom method for batch operations
func (c *Client) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	if tx, ok := GetTx(ctx); ok {
		return tx.CopyFrom(ctx, tableName, columnNames, rowSrc)
	}
	return c.pool.CopyFrom(ctx, tableName, columnNames, rowSrc)
}
