package order

import (
	"context"
)

// OrderRepository is the interface for the order repository.
//
//go:generate mockgen -source=repository.go -destination=mock/repository_mock.go -package=mock
type OrderRepository interface {
	// Order management
	Store(ctx context.Context, order *Order) error
	StoreBatch(ctx context.Context, orders []*Order) error
	Update(ctx context.Context, orderID string, updates map[string]interface{}) error
	GetByID(ctx context.Context, orderID string) (*Order, error)
	GetByFilter(ctx context.Context, filter OrderFilter) ([]*Order, error)

	// Order events
	StoreEvent(ctx context.Context, event *OrderEvent) error
	StoreEventBatch(ctx context.Context, events []*OrderEvent) error
	GetEventsByOrderID(ctx context.Context, orderID string) ([]*OrderEvent, error)

	// Order book reconstruction
	GetActiveOrdersBySymbol(ctx context.Context, symbol string, side string) ([]*Order, error)
	GetOrderBookSnapshot(ctx context.Context, symbol string, depth int) (*OrderBook, error)
}
