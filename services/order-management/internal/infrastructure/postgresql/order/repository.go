package order

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/muhammadchandra19/exchange/pkg/errors"
	"github.com/muhammadchandra19/exchange/pkg/logger"
	"github.com/muhammadchandra19/exchange/pkg/postgresql"
)

// Repository is the repository for the order.
type repository struct {
	db     postgresql.PostgreSQLClient
	logger logger.Interface
}

// NewRepository creates a new repository.
func NewRepository(db postgresql.PostgreSQLClient, logger logger.Interface) *repository {
	return &repository{
		db:     db,
		logger: logger,
	}
}

// Store stores an order.
func (r *repository) Store(ctx context.Context, order *Order) error {
	query := `INSERT INTO orders (id, user_id, symbol, side, price, quantity, type, status, timestamp) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	cmd, err := r.db.Exec(ctx, query,
		order.ID,
		order.UserID,
		order.Symbol,
		order.Side,
		order.Price,
		order.Quantity,
		order.Type,
		order.Status,
		order.Timestamp,
	)
	if err != nil {
		return errors.TracerFromError(err)
	}

	r.logger.Info("Inserted order", logger.Field{
		Key:   "commandTag",
		Value: cmd.String(),
	})

	return nil
}

// StoreBatch stores a batch of orders.
func (r *repository) StoreBatch(ctx context.Context, orders []*Order) error {
	copyCount, err := r.db.CopyFrom(ctx, pgx.Identifier{"orders"}, []string{
		"id",
		"user_id",
		"symbol",
		"side",
		"price",
		"quantity",
		"type",
		"status",
		"timestamp",
	}, pgx.CopyFromSlice(len(orders), func(i int) ([]any, error) {
		order := orders[i]
		return []any{
			order.ID,
			order.UserID,
			order.Symbol,
			order.Side,
			order.Price,
			order.Quantity,
			order.Type,
			order.Status,
			order.Timestamp,
		}, nil
	}))

	if err != nil {
		return errors.TracerFromError(err)
	}

	r.logger.Info("Inserted batch of orders", logger.Field{
		Key:   "copyCount",
		Value: copyCount,
	})

	return nil
}

// Update updates an order.
func (r *repository) Update(ctx context.Context, order *Order) error {
	query := `UPDATE orders SET user_id = $1, symbol = $2, side = $3, price = $4, quantity = $5, type = $6, status = $7, timestamp = $8 WHERE id = $9`

	cmd, err := r.db.Exec(ctx, query,
		order.UserID,
		order.Symbol,
		order.Side,
		order.Price,
		order.Quantity,
		order.Type,
		order.Status,
		order.Timestamp,
		order.ID,
	)

	r.logger.Info("Updated order", logger.Field{
		Key:   "commandTag",
		Value: cmd.String(),
	})

	if err != nil {
		return errors.TracerFromError(err)
	}

	return nil
}

// GetByID gets an order by ID.
func (r *repository) GetByID(ctx context.Context, id string) (*Order, error) {
	query := `SELECT id, user_id, symbol, side, price, quantity, type, status, timestamp FROM orders WHERE id = $1`

	order := &Order{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&order.ID,
		&order.UserID,
		&order.Symbol,
		&order.Side,
		&order.Price,
		&order.Quantity,
		&order.Type,
		&order.Status,
		&order.Timestamp,
	)

	if err != nil {
		return nil, errors.TracerFromError(err)
	}

	return order, nil
}

// Delete deletes an order.
func (r *repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM orders WHERE id = $1`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return errors.TracerFromError(err)
	}

	return nil
}

// List lists orders.
func (r *repository) List(ctx context.Context, filter Filter) ([]*Order, error) {
	query := `SELECT id, user_id, symbol, side, price, quantity, type, status, timestamp FROM orders WHERE 1=1`
	args := []any{}
	argIndex := 1

	if filter.UserID != "" {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, filter.UserID)
		argIndex++
	}

	if filter.Symbol != "" {
		query += fmt.Sprintf(" AND symbol = $%d", argIndex)
		args = append(args, filter.Symbol)
		argIndex++
	}

	if filter.Side != "" {
		query += fmt.Sprintf(" AND side = $%d", argIndex)
		args = append(args, filter.Side)
		argIndex++
	}

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, filter.Status)
		argIndex++
	}

	if filter.From != nil {
		query += fmt.Sprintf(" AND timestamp >= $%d", argIndex)
		args = append(args, *filter.From)
		argIndex++
	}

	if filter.To != nil {
		query += fmt.Sprintf(" AND timestamp <= $%d", argIndex)
		args = append(args, *filter.To)
		argIndex++
	}

	if filter.SortDirection != "" {
		query += fmt.Sprintf(" ORDER BY timestamp %s", filter.SortDirection)
	} else {
		query += " ORDER BY timestamp DESC"
	}

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
		argIndex++
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.TracerFromError(err)
	}
	defer rows.Close()

	orders := []*Order{}
	for rows.Next() {
		order := &Order{}
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Symbol,
			&order.Side,
			&order.Price,
			&order.Quantity,
			&order.Type,
			&order.Status,
			&order.Timestamp,
		)
		if err != nil {
			return nil, errors.TracerFromError(err)
		}
		orders = append(orders, order)
	}

	return orders, nil
}
