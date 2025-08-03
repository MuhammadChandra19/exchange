package order

import (
	"context"

	orderInfra "github.com/muhammadchandra19/exchange/service/order-management/internal/infrastructure/postgresql/order"
)

//go:generate mockgen -source=interface.go -destination=mock/repository_mock.go -package=mock

// Usecase is the repository for the order.
type Usecase interface {
	DeleteOrder(ctx context.Context, orderID string) error
	GetOrder(ctx context.Context, orderID string) (*orderInfra.Order, error)
	GetOrderList(ctx context.Context, filter orderInfra.Filter) ([]*orderInfra.Order, error)
	StoreOrder(ctx context.Context, order *orderInfra.Order) error
	StoreOrders(ctx context.Context, orders []*orderInfra.Order) error
	UpdateOrder(ctx context.Context, orderID string, payload *orderInfra.Order) error
}
