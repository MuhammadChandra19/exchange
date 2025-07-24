package order

import (
	"context"

	"github.com/muhammadchandra19/exchange/services/market-data-service/internal/infrastructure/questdb/order"
)

// Usecase is the interface for the order usecase.
//
//go:generate mockgen -source=interface.go -destination=mock/usecase_mock.go -package=mock
type Usecase interface {
	GetPairActiveOrders(ctx context.Context, symbol string, side string, limit int, offset int) ([]*order.Order, error)
	GetEventsByOrderID(ctx context.Context, orderID string) ([]*order.OrderEvent, error)
	GetOrder(ctx context.Context, orderID string) (*order.Order, error)
	GetOrderBookSnapshot(ctx context.Context, symbol string, depth int) (*order.OrderBook, error)
	GetOrderByFilter(ctx context.Context, filter order.OrderFilter) ([]*order.Order, error)
	StoreOrder(ctx context.Context, order *order.Order) error
	StoreOrderEvent(ctx context.Context, event *order.OrderEvent) error
	DeleteOrder(ctx context.Context, orderID string) error
	StoreOrderEvents(ctx context.Context, events []*order.OrderEvent) error
	StoreOrders(ctx context.Context, orders []*order.Order) error
	UpdateOrder(ctx context.Context, orderID string, updates map[string]interface{}) error
}
