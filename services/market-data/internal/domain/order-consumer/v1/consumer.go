package v1

import (
	"context"
)

//go:generate mockgen -source=consumer.go -destination=mock/consumer_mock.go -package=mock

// OrderConsumer is the consumer for the order topic.
type OrderConsumer interface {
	Start(ctx context.Context)
	Stop() error
	Subscribe(ctx context.Context)
}
