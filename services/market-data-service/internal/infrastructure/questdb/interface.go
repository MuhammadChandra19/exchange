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

func NewRowsWrapper(rows pgx.Rows) RowsInterface {
	return &RowsWrapper{rows: rows}
}

func (r *RowsWrapper) Next() bool {
	return r.rows.Next()
}

func (r *RowsWrapper) Scan(dest ...any) error {
	return r.rows.Scan(dest...)
}

func (r *RowsWrapper) Close() {
	r.rows.Close()
}

func (r *RowsWrapper) Err() error {
	return r.rows.Err()
}

// QuestDBClient defines the interface for QuestDB operations
type QuestDBClient interface {
	// Basic query operations
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
