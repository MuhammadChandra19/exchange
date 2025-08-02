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

// PostgreSQLClient defines the interface for PostgreSQL operations.
type PostgreSQLClient interface {
	// Basic query operations
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (RowsInterface, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row

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

// QueryBuilder provides a fluent interface for building queries
type QueryBuilder interface {
	Select(columns ...string) QueryBuilder
	From(table string) QueryBuilder
	Where(condition string, args ...any) QueryBuilder
	Join(table, condition string) QueryBuilder
	LeftJoin(table, condition string) QueryBuilder
	RightJoin(table, condition string) QueryBuilder
	GroupBy(columns ...string) QueryBuilder
	Having(condition string, args ...any) QueryBuilder
	OrderBy(column string, desc ...bool) QueryBuilder
	Limit(limit int) QueryBuilder
	Offset(offset int) QueryBuilder
	Build() (string, []any)
	Reset() QueryBuilder
}

// InsertBuilder provides a fluent interface for building INSERT queries
type InsertBuilder interface {
	Into(table string) InsertBuilder
	Columns(columns ...string) InsertBuilder
	Values(values ...any) InsertBuilder
	ValuesMap(valueMap map[string]any) InsertBuilder
	OnConflict(constraint string) InsertBuilder
	OnConflictDoNothing() InsertBuilder
	OnConflictDoUpdate(setClause string, args ...any) InsertBuilder
	Returning(columns ...string) InsertBuilder
	Build() (string, []any)
	Reset() InsertBuilder
}

// UpdateBuilder provides a fluent interface for building UPDATE queries
type UpdateBuilder interface {
	Table(table string) UpdateBuilder
	Set(column string, value any) UpdateBuilder
	SetMap(updates map[string]any) UpdateBuilder
	Where(condition string, args ...any) UpdateBuilder
	Returning(columns ...string) UpdateBuilder
	Build() (string, []any)
	Reset() UpdateBuilder
}

// DeleteBuilder provides a fluent interface for building DELETE queries
type DeleteBuilder interface {
	From(table string) DeleteBuilder
	Where(condition string, args ...any) DeleteBuilder
	Returning(columns ...string) DeleteBuilder
	Build() (string, []any)
	Reset() DeleteBuilder
}
