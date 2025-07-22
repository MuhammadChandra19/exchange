package order

import (
	"context"

	"github.com/muhammadchandra19/exchange/pkg/errors"
	"github.com/muhammadchandra19/exchange/pkg/logger"
	"github.com/muhammadchandra19/exchange/services/market-data-service/internal/infrastructure/questdb/order"
)

// Usecase is the usecase for the order.
type Usecase struct {
	orderRepository order.OrderRepository
	logger          logger.Logger
}

// NewUsecase creates a new order usecase.
func NewUsecase(orderRepository order.OrderRepository, logger logger.Logger) *Usecase {
	return &Usecase{orderRepository: orderRepository, logger: logger}
}

// GetOrder gets the order for a given order ID.
func (u *Usecase) GetOrder(ctx context.Context, orderID string) (*order.Order, error) {
	order, err := u.orderRepository.GetByID(ctx, orderID)
	if err != nil {
		return nil, errors.TracerFromError(err)
	}
	return order, nil
}

// GetOrderByFilter gets the order for a given filter.
func (u *Usecase) GetOrderByFilter(ctx context.Context, filter order.OrderFilter) ([]*order.Order, error) {
	orders, err := u.orderRepository.GetByFilter(ctx, filter)
	if err != nil {
		return nil, errors.TracerFromError(err)
	}
	return orders, nil
}

// GetActiveOrdersBySymbol gets the active orders for a given symbol and side.
func (u *Usecase) GetActiveOrdersBySymbol(ctx context.Context, symbol string, side string) ([]*order.Order, error) {
	orders, err := u.orderRepository.GetActiveOrdersBySymbol(ctx, symbol, side)
	if err != nil {
		return nil, errors.TracerFromError(err)
	}
	return orders, nil
}

// GetOrderBookSnapshot gets the order book snapshot for a given symbol and depth.
func (u *Usecase) GetOrderBookSnapshot(ctx context.Context, symbol string, depth int) (*order.OrderBook, error) {
	orderBook, err := u.orderRepository.GetOrderBookSnapshot(ctx, symbol, depth)
	if err != nil {
		return nil, errors.TracerFromError(err)
	}
	return orderBook, nil
}

// StoreOrder stores an order.
func (u *Usecase) StoreOrder(ctx context.Context, order *order.Order) error {
	err := u.orderRepository.Store(ctx, order)
	if err != nil {
		return errors.TracerFromError(err)
	}
	return nil
}

// StoreOrders stores a batch of orders.
func (u *Usecase) StoreOrders(ctx context.Context, orders []*order.Order) error {
	err := u.orderRepository.StoreBatch(ctx, orders)
	if err != nil {
		return errors.TracerFromError(err)
	}
	return nil
}

// UpdateOrder updates an order.
func (u *Usecase) UpdateOrder(ctx context.Context, orderID string, updates map[string]interface{}) error {
	err := u.orderRepository.Update(ctx, orderID, updates)
	if err != nil {
		return errors.TracerFromError(err)
	}
	return nil
}

// StoreOrderEvent stores an order event.
func (u *Usecase) StoreOrderEvent(ctx context.Context, event *order.OrderEvent) error {
	err := u.orderRepository.StoreEvent(ctx, event)
	if err != nil {
		return errors.TracerFromError(err)
	}
	return nil
}

// StoreOrderEvents stores a batch of order events.
func (u *Usecase) StoreOrderEvents(ctx context.Context, events []*order.OrderEvent) error {
	err := u.orderRepository.StoreEventBatch(ctx, events)
	if err != nil {
		return errors.TracerFromError(err)
	}
	return nil
}

// GetEventsByOrderID gets the events for a given order ID.
func (u *Usecase) GetEventsByOrderID(ctx context.Context, orderID string) ([]*order.OrderEvent, error) {
	events, err := u.orderRepository.GetEventsByOrderID(ctx, orderID)
	if err != nil {
		return nil, errors.TracerFromError(err)
	}
	return events, nil
}
