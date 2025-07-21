package redis

import (
	"context"
	"time"

	v9 "github.com/redis/go-redis/v9"
)

// Client defines the interface for a Redis client.
//
//go:generate mockgen -source interface.go -destination=mock/interface_mock.go -package=redis_mock
type Client interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Ping(ctx context.Context) error
	Reconnect(ctx context.Context) bool

	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error)
	Del(ctx context.Context, keys ...string) (int64, error)

	HGet(ctx context.Context, key, field string) (string, error)
	HSet(ctx context.Context, key string, values map[string]any) (int64, error)
	HDel(ctx context.Context, key string, fields ...string) (int64, error)

	ZAdd(ctx context.Context, key string, members ...v9.Z) (int64, error)

	Subscribe(ctx context.Context, channels ...string) (*v9.PubSub, error)
	Publish(ctx context.Context, channel string, message any) (int64, error)

	XAdd(ctx context.Context, args *v9.XAddArgs) (string, error)
	XLen(ctx context.Context, stream string) (int64, error)
	XRead(ctx context.Context, args *v9.XReadArgs) ([]v9.XStream, error)
	XReadGroup(ctx context.Context, args *v9.XReadGroupArgs) ([]v9.XStream, error)
}
