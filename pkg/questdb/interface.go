package questdb

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:generate mockgen -source=interface.go -destination=mock/interface_mock.go -package=mock

// RowsInterface wraps pgx.Rows for mocking
type RowsInterface interface {
	Next() bool
	Scan(dest ...any) error
	Close()
	Err() error
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

// QuestDBClient defines the interface for QuestDB operations.
type QuestDBClient interface {
	// Basic query operations.
	Exec(ctx context.Context, sql string, args ...any) error
	Query(ctx context.Context, sql string, args ...any) (RowsInterface, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row

	// Transaction operations
	Begin(ctx context.Context) (pgx.Tx, error)

	// Batch operations
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)

	// Connection management
	Ping(ctx context.Context) error
	Close()

	// Pool access (for advanced operations if needed)
	Pool() *pgxpool.Pool
}
