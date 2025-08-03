package order

import (
	"context"

	"github.com/muhammadchandra19/exchange/pkg/errors"
	"github.com/muhammadchandra19/exchange/pkg/logger"
	"github.com/muhammadchandra19/exchange/service/order-management/internal/infrastructure/postgresql/order"
)

type usecase struct {
	orderRepository order.OrderRepository
	logger          logger.Interface
}

// NewUsecase creates a new order usecase.
func NewUsecase(orderRepository order.OrderRepository, logger logger.Interface) *usecase {
	return &usecase{orderRepository: orderRepository, logger: logger}
}

// StoreOrder stores an order.
func (u *usecase) StoreOrder(ctx context.Context, order *order.Order) error {
	err := u.orderRepository.Store(ctx, order)
	if err != nil {
		return err
	}
	return nil
}

// StoreOrders stores a batch of orders.
func (u *usecase) StoreOrders(ctx context.Context, orders []*order.Order) error {
	if len(orders) == 0 {
		return errors.TracerFromError(errors.NewErrorDetails(
			string(errors.GeneralBadRequestError),
			"orders",
			"orders",
		))
	}

	err := u.orderRepository.StoreBatch(ctx, orders)
	if err != nil {
		return err
	}
	return nil
}

// GetOrder gets an order.
func (u *usecase) GetOrder(ctx context.Context, orderID string) (*order.Order, error) {
	order, err := u.orderRepository.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	return order, nil
}

// GetOrderList gets the order list.
func (u *usecase) GetOrderList(ctx context.Context, filter order.Filter) ([]*order.Order, error) {
	orders, err := u.orderRepository.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (u *usecase) UpdateOrder(ctx context.Context, orderID string, payload *order.Order) error {
	order, err := u.orderRepository.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	if order == nil {
		u.logger.Error(err, logger.Field{
			Key:   "Order not found",
			Value: err,
		})
		return errors.TracerFromError(err)
	}

	err = u.orderRepository.Update(ctx, order)
	if err != nil {
		u.logger.Error(err, logger.Field{
			Key:   "Error updating order",
			Value: err,
		})
		return err
	}

	return nil
}

// DeleteOrder deletes an order.
func (u *usecase) DeleteOrder(ctx context.Context, orderID string) error {
	order, err := u.orderRepository.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	if order == nil {
		u.logger.Error(err, logger.Field{
			Key:   "Order not found",
			Value: err,
		})
		return errors.TracerFromError(err)
	}

	err = u.orderRepository.Delete(ctx, orderID)
	if err != nil {
		u.logger.Error(err, logger.Field{
			Key:   "Error deleting order",
			Value: err,
		})
		return err
	}
	return nil
}
