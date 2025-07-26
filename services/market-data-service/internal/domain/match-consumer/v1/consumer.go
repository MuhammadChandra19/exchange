package v1

import (
	"context"
)

//go:generate mockgen -source=consumer.go -destination=mock/consumer_mock.go -package=mock

// MatchConsumer represents a consumer that processes match events.
type MatchConsumer interface {
	Start(ctx context.Context)
	Stop(ctx context.Context) error
}
