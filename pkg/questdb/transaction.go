package questdb

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type contextKey string

const txKey contextKey = "questdb_transaction"

//go:generate mockgen -source=transaction.go -destination=mock/transaction_mock.go -package=mock

// Transaction is the transaction interface.
type Transaction interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// TX is the transaction wrapper.
type TX struct {
	db QuestDBClient
}

// NewTransaction creates a new transaction wrapper.
func NewTransaction(db QuestDBClient) TX {
	return TX{db: db}
}

// Begin starts a transaction and returns context with embedded transaction
func (t *TX) Begin(ctx context.Context) (context.Context, error) {
	tx, err := t.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return context.WithValue(ctx, txKey, tx), nil
}

// Commit commits the transaction from context
func (t *TX) Commit(ctx context.Context) error {
	tx, ok := ctx.Value(txKey).(pgx.Tx)
	if !ok {
		return fmt.Errorf("no transaction found in context")
	}
	return tx.Commit(ctx)
}

// Rollback rolls back the transaction from context
func (t *TX) Rollback(ctx context.Context) error {
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
