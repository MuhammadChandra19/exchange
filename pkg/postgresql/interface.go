package postgresql

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:generate mockgen -source=interface.go -destination=mock/interface_mock.go -package=mock

// RowsInterface wraps pgx.Rows for mocking
type RowsInterface interface {
	Next() bool
	Scan(dest ...any) error
	Close()
	Err() error
	Values() ([]any, error)
	FieldDescriptions() []pgconn.FieldDescription
}

// RowsWrapper wraps pgx.Rows to implement RowsInterface
type RowsWrapper struct {
	rows pgx.Rows
}

// NewRowsWrapper creates a new RowsWrapper.
func NewRowsWrapper(rows pgx.Rows) RowsInterface {
	return &RowsWrapper{rows: rows}
}

// Next returns true if there are more rows to read.
func (r *RowsWrapper) Next() bool {
	return r.rows.Next()
}

// Scan scans the next row into the given destination.
func (r *RowsWrapper) Scan(dest ...any) error {
	return r.rows.Scan(dest...)
}

// Close closes the RowsWrapper.
func (r *RowsWrapper) Close() {
	r.rows.Close()
}

// Err returns the error from the RowsWrapper.
func (r *RowsWrapper) Err() error {
	return r.rows.Err()
}

// Values returns the decoded row values.
func (r *RowsWrapper) Values() ([]any, error) {
	return r.rows.Values()
}

// FieldDescriptions returns the field descriptions.
func (r *RowsWrapper) FieldDescriptions() []pgconn.FieldDescription {
	return r.rows.FieldDescriptions()
}

// RowInterface wraps pgx.Row for mocking
type RowInterface interface {
	Scan(dest ...any) error
}

// RowWrapper wraps pgx.Row to implement RowInterface
type RowWrapper struct {
	row pgx.Row
}

// NewRowWrapper creates a new RowWrapper.
func NewRowWrapper(row pgx.Row) RowInterface {
	return &RowWrapper{row: row}
}

// Scan scans the next row into the given destination.
func (r *RowWrapper) Scan(dest ...any) error {
	return r.row.Scan(dest...)
}

// PostgreSQLClient defines the interface for PostgreSQL operations.
type PostgreSQLClient interface {
	// Basic query operations
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (RowsInterface, error)
	QueryRow(ctx context.Context, sql string, args ...any) RowInterface

	// Transaction operations
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)

	// Batch operations
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)

	// Prepared statements
	Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error)
	Deallocate(ctx context.Context, name string) error

	// Connection management
	Ping(ctx context.Context) error
	Close()
	Acquire(ctx context.Context) (*pgxpool.Conn, error)

	// Pool access and stats
	Pool() *pgxpool.Pool
	Stats() *pgxpool.Stat

	// Database introspection
	DatabaseName() string
	Host() string
	Port() int
}
