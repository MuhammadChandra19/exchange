package order

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/muhammadchandra19/exchange/pkg/questdb"
)

// Repository is the repository for the order.
type Repository struct {
	client questdb.QuestDBClient
}

// NewRepository creates a new repository.
func NewRepository(client questdb.QuestDBClient) *Repository {
	return &Repository{
		client: client,
	}
}

// Store stores an order.
func (r *Repository) Store(ctx context.Context, order *Order) error {
	query := `INSERT INTO orders (order_id, timestamp, symbol, side, price, quantity, order_type, status, user_id) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	err := r.client.Exec(ctx, query,
		order.ID, order.Timestamp, order.Symbol, order.Side, order.Price,
		order.Quantity, order.Type, order.Status, order.UserID)

	if err != nil {
		return fmt.Errorf("failed to store order: %w", err)
	}

	return nil
}

// StoreBatch stores a batch of orders.
func (r *Repository) StoreBatch(ctx context.Context, orders []*Order) error {
	if len(orders) == 0 {
		return nil
	}

	copyCount, err := r.client.CopyFrom(
		ctx,
		pgx.Identifier{"orders"},
		[]string{"order_id", "timestamp", "symbol", "side", "price", "quantity", "order_type", "status", "user_id"},
		pgx.CopyFromSlice(len(orders), func(i int) ([]any, error) {
			order := orders[i]
			return []any{
				order.ID, order.Timestamp, order.Symbol, order.Side, order.Price,
				order.Quantity, order.Type, order.Status, order.UserID,
			}, nil
		}),
	)

	if err != nil {
		return fmt.Errorf("failed to copy orders batch: %w", err)
	}

	fmt.Printf("Inserted %d orders\n", copyCount)
	return nil
}

// Update updates an order.
func (r *Repository) Update(ctx context.Context, orderID string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	argIndex := 1

	for column, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", column, argIndex))
		args = append(args, value)
		argIndex++
	}

	query := fmt.Sprintf("UPDATE orders SET %s WHERE order_id = $%d",
		strings.Join(setParts, ", "), argIndex)
	args = append(args, orderID)

	err := r.client.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update order %s: %w", orderID, err)
	}

	return nil
}

// GetByID gets an order by ID.
func (r *Repository) GetByID(ctx context.Context, orderID string) (*Order, error) {
	query := `SELECT order_id, timestamp, symbol, side, price, quantity, order_type, status, user_id
			  FROM orders WHERE order_id = $1`

	order := &Order{}
	err := r.client.QueryRow(ctx, query, orderID).Scan(
		&order.ID, &order.Timestamp, &order.Symbol, &order.Side, &order.Price,
		&order.Quantity, &order.Type, &order.Status, &order.UserID)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return order, nil
}

// Delete deletes an order.
func (r *Repository) Delete(ctx context.Context, orderID string) error {
	query := `DELETE FROM orders WHERE order_id = $1`

	err := r.client.Exec(ctx, query, orderID)
	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	return nil
}

// StoreEvent stores an order event.
func (r *Repository) StoreEvent(ctx context.Context, event *OrderEvent) error {
	query := `INSERT INTO order_events (event_id, timestamp, order_id, event_type, symbol, side, price, quantity, user_id, new_price, new_quantity) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	err := r.client.Exec(ctx, query,
		event.ID, event.Timestamp, event.OrderID, event.EventType, event.Symbol,
		event.Side, event.Price, event.Quantity, event.UserID,
		event.NewPrice, event.NewQuantity)

	if err != nil {
		return fmt.Errorf("failed to store order event: %w", err)
	}

	return nil
}

// GetActiveOrdersBySymbol gets active orders by symbol and side.
func (r *Repository) GetActiveOrdersBySymbol(ctx context.Context, symbol string, side string, limit int, offset int) ([]*Order, error) {
	query := `SELECT order_id, timestamp, symbol, side, price, quantity, order_type, status, user_id
			  FROM orders WHERE status = 'active'`

	args := []interface{}{}
	argIndex := 1

	if symbol != "" {
		query += fmt.Sprintf(" AND symbol = $%d", argIndex)
		args = append(args, symbol)
		argIndex++
	}

	if side != "" {
		query += fmt.Sprintf(" AND side = $%d", argIndex)
		args = append(args, side)
		argIndex++
	}

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
		argIndex++
	}

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, offset)
		argIndex++
	}

	rows, err := r.client.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query active orders: %w", err)
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		order := &Order{}
		err := rows.Scan(&order.ID, &order.Timestamp, &order.Symbol, &order.Side,
			&order.Price, &order.Quantity, &order.Type, &order.Status, &order.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// GetOrderBookSnapshot gets an order book snapshot.
func (r *Repository) GetOrderBookSnapshot(ctx context.Context, symbol string, depth int) (*OrderBook, error) {
	// Get aggregated bid levels
	bidQuery := `SELECT price, SUM(quantity) as total_quantity, COUNT(*) as order_count
				 FROM orders 
				 WHERE symbol = $1 AND side = 'buy' AND status = 'active'
				 GROUP BY price
				 ORDER BY price DESC
				 LIMIT $2`

	bidRows, err := r.client.Query(ctx, bidQuery, symbol, depth)
	if err != nil {
		return nil, fmt.Errorf("failed to query bid levels: %w", err)
	}
	defer bidRows.Close()

	var bids []OrderBookLevel
	for bidRows.Next() {
		var level OrderBookLevel
		err := bidRows.Scan(&level.Price, &level.Quantity, &level.Orders)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bid level: %w", err)
		}
		bids = append(bids, level)
	}

	// Get aggregated ask levels
	askQuery := `SELECT price, SUM(quantity) as total_quantity, COUNT(*) as order_count
				 FROM orders 
				 WHERE symbol = $1 AND side = 'sell' AND status = 'active'
				 GROUP BY price
				 ORDER BY price ASC
				 LIMIT $2`

	askRows, err := r.client.Query(ctx, askQuery, symbol, depth)
	if err != nil {
		return nil, fmt.Errorf("failed to query ask levels: %w", err)
	}
	defer askRows.Close()

	var asks []OrderBookLevel
	for askRows.Next() {
		var level OrderBookLevel
		err := askRows.Scan(&level.Price, &level.Quantity, &level.Orders)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ask level: %w", err)
		}
		asks = append(asks, level)
	}

	return &OrderBook{
		Symbol:    symbol,
		Timestamp: time.Now(),
		Bids:      bids,
		Asks:      asks,
	}, nil
}

// Helper methods for the rest of the interface

// GetByFilter gets orders by filter.
func (r *Repository) GetByFilter(ctx context.Context, filter OrderFilter) ([]*Order, error) {
	// Implementation similar to tick repository GetByFilter
	// ... (implement based on your filtering needs)
	return nil, nil
}

// StoreEventBatch stores a batch of order events.
func (r *Repository) StoreEventBatch(ctx context.Context, events []*OrderEvent) error {
	// Implementation using CopyFrom for batch event storage
	// ... (implement similar to other batch operations)
	return nil
}

// GetEventsByOrderID gets events by order ID.
func (r *Repository) GetEventsByOrderID(ctx context.Context, orderID string) ([]*OrderEvent, error) {
	// Implementation to get all events for a specific order
	// ... (implement based on your needs)
	return nil, nil
}
