package questdb

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:generate mockgen -source=interface.go -destination=mock/interface_mock.go -package=mock

// QuestDBClient defines the interface for QuestDB operations
type QuestDBClient interface {
	// Basic query operations
	Exec(ctx context.Context, sql string, args ...any) error
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
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
