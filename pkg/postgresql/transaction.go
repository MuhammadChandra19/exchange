package postgresql

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type contextKey string

const txKey contextKey = "postgresql_transaction"

//go:generate mockgen -source=transaction.go -destination=mock/transaction_mock.go -package=mock

// Transaction is the transaction interface with PostgreSQL-specific features.
type Transaction interface {
	Begin(ctx context.Context) (context.Context, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (context.Context, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	Savepoint(ctx context.Context, name string) error
	RollbackToSavepoint(ctx context.Context, name string) error
	ReleaseSavepoint(ctx context.Context, name string) error
}

// TX is the transaction wrapper optimized for PostgreSQL.
type TX struct {
	db PostgreSQLClient
}

// NewTransaction creates a new transaction wrapper.
func NewTransaction(db PostgreSQLClient) *TX {
	return &TX{db: db}
}

// Begin starts a transaction and returns context with embedded transaction
func (t *TX) Begin(ctx context.Context) (context.Context, error) {
	tx, err := t.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return context.WithValue(ctx, txKey, tx), nil
}

// BeginTx starts a transaction with options and returns context with embedded transaction
func (t *TX) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (context.Context, error) {
	tx, err := t.db.BeginTx(ctx, txOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction with options: %w", err)
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

// Savepoint creates a savepoint within the transaction
func (t *TX) Savepoint(ctx context.Context, name string) error {
	tx, ok := ctx.Value(txKey).(pgx.Tx)
	if !ok {
		return fmt.Errorf("no transaction found in context")
	}

	_, err := tx.Exec(ctx, fmt.Sprintf("SAVEPOINT %s", name))
	return err
}

// RollbackToSavepoint rolls back to a savepoint
func (t *TX) RollbackToSavepoint(ctx context.Context, name string) error {
	tx, ok := ctx.Value(txKey).(pgx.Tx)
	if !ok {
		return fmt.Errorf("no transaction found in context")
	}

	_, err := tx.Exec(ctx, fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", name))
	return err
}

// ReleaseSavepoint releases a savepoint
func (t *TX) ReleaseSavepoint(ctx context.Context, name string) error {
	tx, ok := ctx.Value(txKey).(pgx.Tx)
	if !ok {
		return fmt.Errorf("no transaction found in context")
	}

	_, err := tx.Exec(ctx, fmt.Sprintf("RELEASE SAVEPOINT %s", name))
	return err
}

// GetTx extracts transaction from context (helper function)
func GetTx(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey).(pgx.Tx)
	return tx, ok
}

// WithTx executes a function within a transaction with automatic rollback on error
func WithTx(ctx context.Context, db PostgreSQLClient, fn func(ctx context.Context) error) error {
	tx := NewTransaction(db)

	txCtx, err := tx.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(txCtx)
			panic(p)
		}
	}()

	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(txCtx); rbErr != nil {
			return fmt.Errorf("transaction failed: %v, rollback failed: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit(txCtx)
}

// WithTxOptions executes a function within a transaction with specific options
func WithTxOptions(ctx context.Context, db PostgreSQLClient, txOptions pgx.TxOptions, fn func(ctx context.Context) error) error {
	tx := NewTransaction(db)

	txCtx, err := tx.BeginTx(ctx, txOptions)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(txCtx)
			panic(p)
		}
	}()

	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(txCtx); rbErr != nil {
			return fmt.Errorf("transaction failed: %v, rollback failed: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit(txCtx)
}

// ReadOnlyTxOptions returns transaction options for read-only transactions
func ReadOnlyTxOptions() pgx.TxOptions {
	return pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadOnly,
	}
}

// SerializableTxOptions returns transaction options for serializable transactions
func SerializableTxOptions() pgx.TxOptions {
	return pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	}
}
