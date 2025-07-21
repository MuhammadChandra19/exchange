package v1

import (
	"context"
)

//go:generate mockgen -source=consumer.go -destination=mock/consumer_mock.go -package=mock

type OrderConsumer interface {
	Start(ctx context.Context) error
	Stop() error
	Subscribe(handler func(ctx context.Context, event *RawOrderEvent) error) error
}
