package order

import "context"

// OrderRepository is the repository for the order.

//go:generate mockgen -source=interface.go -destination=mock/repository_mock.go -package=mock

// OrderRepository is the repository for the order.
type OrderRepository interface {
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*Order, error)
	List(ctx context.Context, filter Filter) ([]*Order, error)
	Store(ctx context.Context, order *Order) error
	StoreBatch(ctx context.Context, orders []*Order) error
	Update(ctx context.Context, order *Order) error
}
