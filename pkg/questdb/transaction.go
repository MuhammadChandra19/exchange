package questdb

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type contextKey string

const txKey contextKey = "questdb_transaction"

// Begin starts a transaction and returns context with embedded transaction
func Begin(ctx context.Context, client QuestDBClient) (context.Context, error) {
	tx, err := client.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return context.WithValue(ctx, txKey, tx), nil
}

// Commit commits the transaction from context
func Commit(ctx context.Context) error {
	tx, ok := ctx.Value(txKey).(pgx.Tx)
	if !ok {
		return fmt.Errorf("no transaction found in context")
	}
	return tx.Commit(ctx)
}

// Rollback rolls back the transaction from context
func Rollback(ctx context.Context) error {
	tx, ok := ctx.Value(txKey).(pgx.Tx)
	if !ok {
		return fmt.Errorf("no transaction found in context")
	}
	return tx.Rollback(ctx)
}

// GetTx extracts transaction from context (helper function)
func GetTx(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey).(pgx.Tx)
	return tx, ok
}
