package redis

import (
	"context"
	"math"
	"math/rand/v2"
	"time"

	"github.com/muhammadchandra19/exchange/pkg/errors"
	"github.com/muhammadchandra19/exchange/pkg/logger"
	"github.com/redis/go-redis/v9"
)

type client struct {
	logger  *logger.Logger
	config  *Config
	cmdable redis.Cmdable
}

// NewClient creates a new Redis client with the provided logger and configuration.
func NewClient(logger *logger.Logger, config *Config) Client {
	return &client{
		logger: logger,
		config: config,
	}
}

func (c *client) Connect(ctx context.Context) error {
	var cmdable redis.Cmdable
	if c.config == nil {
		return errors.NewErrorDetails("Redis config is nil", string(errors.RedisConfigError), "connect")
	}

	if len(c.config.Addrs) == 0 {
		return errors.NewErrorDetails("Redis addresses are empty", string(errors.RedisConfigError), "connect")
	}

	if c.config.Mode != Standalone && c.config.Mode != Cluster {
		return errors.NewErrorDetails("Invalid Redis mode", string(errors.RedisConfigError), "connect")
	}

	if c.config.ConnectTimeout <= 0 {
		return errors.NewErrorDetails("Invalid Redis connect timeout", string(errors.RedisConfigError), "connect")
	}
	if c.config.PoolSize <= 0 {
		return errors.NewErrorDetails("Invalid Redis pool size", string(errors.RedisConfigError), "connect")
	}

	if c.config.MaxIdleConns < 0 {
		return errors.NewErrorDetails("Invalid Redis max idle connections", string(errors.RedisConfigError), "connect")
	}

	if c.config.MaxIdleConns < 0 {
		return errors.NewErrorDetails("Invalid Redis max idle connections", string(errors.RedisConfigError), "connect")
	}
	if c.config.ConnMaxLifetime <= 0 {
		return errors.NewErrorDetails("Invalid Redis connection max lifetime", string(errors.RedisConfigError), "connect")
	}

	if c.config.ConnMaxIdleTime <= 0 {
		return errors.NewErrorDetails("Invalid Redis connection max idle time", string(errors.RedisConfigError), "connect")
	}

	if c.config.PoolTimeout <= 0 {
		return errors.NewErrorDetails("Invalid Redis pool timeout", string(errors.RedisConfigError), "connect")
	}

	if c.config.MaxRetries < 0 {
		return errors.NewErrorDetails("Invalid Redis max retries", string(errors.RedisConfigError), "connect")
	}

	if c.config.MinRetryBackoff < 0 {
		return errors.NewErrorDetails("Invalid Redis minimum retry backoff", string(errors.RedisConfigError), "connect")
	}

	if c.config.MaxRetryBackoff < 0 {
		return errors.NewErrorDetails("Invalid Redis maximum retry backoff", string(errors.RedisConfigError), "connect")
	}

	switch c.config.Mode {
	case Standalone:
		cmdable = redis.NewClient(&redis.Options{
			Addr:            c.config.Addrs[0],
			Username:        c.config.Username,
			Password:        c.config.Password,
			DB:              c.config.DB,
			MaxRetries:      c.config.MaxRetries,
			MinRetryBackoff: c.config.MinRetryBackoff,
			MaxRetryBackoff: c.config.MaxRetryBackoff,
			DialTimeout:     c.config.ConnectTimeout,
			ReadTimeout:     c.config.ConnectTimeout,
			WriteTimeout:    c.config.ConnectTimeout,
			PoolSize:        c.config.PoolSize,
			MinIdleConns:    c.config.MinIdleConns,
			MaxIdleConns:    c.config.MaxIdleConns,
			ConnMaxLifetime: c.config.ConnMaxLifetime,
			ConnMaxIdleTime: c.config.ConnMaxIdleTime,
			PoolTimeout:     c.config.PoolTimeout,
		})
	case Cluster:
		cmdable = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:           c.config.Addrs,
			Username:        c.config.Username,
			Password:        c.config.Password,
			MaxRetries:      c.config.MaxRetries,
			MinRetryBackoff: c.config.MinRetryBackoff,
			MaxRetryBackoff: c.config.MaxRetryBackoff,
			DialTimeout:     c.config.ConnectTimeout,
			ReadTimeout:     c.config.ConnectTimeout,
			WriteTimeout:    c.config.ConnectTimeout,
			PoolSize:        c.config.PoolSize,
			MinIdleConns:    c.config.MinIdleConns,
			MaxIdleConns:    c.config.MaxIdleConns,
			ConnMaxLifetime: c.config.ConnMaxLifetime,
			ConnMaxIdleTime: c.config.ConnMaxIdleTime,
			PoolTimeout:     c.config.PoolTimeout,
		})
	default:
		return errors.NewErrorDetails("Unsupported Redis mode", string(errors.RedisConnectionError), "connect")

	}

	c.cmdable = cmdable

	return c.cmdable.Ping(ctx).Err()

}

func (c *client) Reconnect(ctx context.Context) bool {
	baseDelay := c.config.MinRetryBackoff
	maxDelay := c.config.MaxRetryBackoff

	for i := range c.config.ReconnectMaxRetries {
		backoff := min(baseDelay*time.Duration(math.Pow(2, float64(i))), maxDelay)

		jitter := time.Duration(rand.IntN(1000)) * time.Millisecond
		totalDelay := backoff + jitter

		c.logger.Info("Reconnecting to Redis", logger.Field{
			Key:   "attempt",
			Value: i + 1,
		}, logger.Field{
			Key:   "delay",
			Value: totalDelay,
		})

		select {
		case <-ctx.Done():
			c.logger.Info("Reconnect cancelled", logger.Field{
				Key:   "reason",
				Value: ctx.Err(),
			})
			return false
		case <-time.After(totalDelay):
			connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			err := c.Connect(connectCtx)
			cancel()
			if err == nil {
				c.logger.Info("Reconnected to Redis successfully", logger.Field{
					Key:   "attempt",
					Value: i + 1,
				})
				return true
			}
			c.logger.Error(errors.TracerFromError(err), logger.Field{
				Key:   "attempt",
				Value: i + 1,
			}, logger.Field{
				Key:   "error",
				Value: err.Error(),
			})
		}
	}

	return true
}

func (c *client) Disconnect(ctx context.Context) error {
	switch c.config.Mode {
	case Standalone:
		return c.cmdable.(*redis.Client).Close()
	case Cluster:
		return c.cmdable.(*redis.ClusterClient).Close()
	default:
		return errors.NewErrorDetails("Unsupported Redis mode for disconnect", string(errors.RedisDisconnectionError), "disconnect")
	}
}

func (c *client) Ping(ctx context.Context) error {
	if err := c.cmdable.Ping(ctx).Err(); err != nil {
		return errors.NewErrorDetails("Failed to ping Redis", string(errors.RedisPingError), "ping")
	}
	return nil
}

func (c *client) Get(ctx context.Context, key string) (string, error) {
	val, err := c.cmdable.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", errors.NewErrorDetails("Failed to get value from Redis", string(errors.RedisGetError), "get")
	}
	return val, nil
}

func (c *client) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	if err := c.cmdable.Set(ctx, key, value, expiration).Err(); err != nil {
		return errors.NewErrorDetails("Failed to set value in Redis", string(errors.RedisSetError), "set")
	}
	return nil
}

func (c *client) SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
	ok, err := c.cmdable.SetNX(ctx, key, value, expiration).Result()
	if err != nil {
		return false, errors.NewErrorDetails("Failed to set value with NX in Redis", string(errors.RedisSetNXError), "setnx")
	}
	return ok, nil
}

func (c *client) Del(ctx context.Context, keys ...string) (int64, error) {
	deleted, err := c.cmdable.Del(ctx, keys...).Result()
	if err != nil {
		return 0, errors.NewErrorDetails("Failed to delete keys from Redis", string(errors.RedisDelError), "del")
	}
	return deleted, nil
}

func (c *client) HGet(ctx context.Context, key, field string) (string, error) {
	val, err := c.cmdable.HGet(ctx, key, field).Result()
	if err == redis.Nil {
		return "", nil // Field does not exist
	}
	if err != nil {
		return "", errors.NewErrorDetails("Failed to get field from hash in Redis", string(errors.RedisHGetError), "hget")
	}
	return val, nil
}

func (c *client) HSet(ctx context.Context, key string, values map[string]any) (int64, error) {
	affected, err := c.cmdable.HSet(ctx, key, values).Result()
	if err != nil {
		return 0, errors.NewErrorDetails("Failed to set fields in hash in Redis", string(errors.RedisHSetError), "hset")
	}
	return affected, nil
}
func (c *client) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	deleted, err := c.cmdable.HDel(ctx, key, fields...).Result()
	if err != nil {
		return 0, errors.NewErrorDetails("Failed to delete fields from hash in Redis", string(errors.RedisHDelError), "hdel")
	}
	return deleted, nil
}

func (c *client) ZAdd(ctx context.Context, key string, members ...redis.Z) (int64, error) {
	added, err := c.cmdable.ZAdd(ctx, key, members...).Result()
	if err != nil {
		return 0, errors.NewErrorDetails("Failed to add members to sorted set in Redis", string(errors.RedisZAddError), "zadd")
	}
	return added, nil
}

func (c *client) Subscribe(ctx context.Context, channels ...string) (*redis.PubSub, error) {
	var pubSub *redis.PubSub

	switch c.config.Mode {
	case Standalone:
		pubSub = c.cmdable.(*redis.Client).Subscribe(ctx, channels...)
	case Cluster:
		pubSub = c.cmdable.(*redis.ClusterClient).Subscribe(ctx, channels...)
	default:
		return nil, errors.NewErrorDetails("Unsupported Redis mode for subscribe", string(errors.RedisSubscribeError), "subscribe")
	}

	_, err := pubSub.Receive(ctx)
	if err != nil {
		return nil, errors.NewErrorDetails("Failed to subscribe to channels in Redis", string(errors.RedisSubscribeError), "subscribe")
	}
	return pubSub, nil
}

func (c *client) Publish(ctx context.Context, channel string, message any) (int64, error) {
	var published int64

	switch c.config.Mode {
	case Standalone:
		published, _ = c.cmdable.(*redis.Client).Publish(ctx, channel, message).Result()
	case Cluster:
		published, _ = c.cmdable.(*redis.ClusterClient).Publish(ctx, channel, message).Result()
	default:
		return 0, errors.NewErrorDetails("Unsupported Redis mode for publish", string(errors.RedisPublishError), "publish")
	}

	if published == 0 {
		return 0, errors.NewErrorDetails("No subscribers for the channel", string(errors.RedisPublishError), "publish")
	}
	return published, nil
}

func (c *client) XAdd(ctx context.Context, args *redis.XAddArgs) (string, error) {
	var streamID string

	switch c.config.Mode {
	case Standalone:
		streamID, _ = c.cmdable.(*redis.Client).XAdd(ctx, args).Result()
	case Cluster:
		streamID, _ = c.cmdable.(*redis.ClusterClient).XAdd(ctx, args).Result()
	default:
		return "", errors.NewErrorDetails("Unsupported Redis mode for XAdd", string(errors.RedisXAddError), "xadd")
	}

	if streamID == "" {
		return "", errors.NewErrorDetails("Failed to add entry to stream", string(errors.RedisXAddError), "xadd")
	}
	return streamID, nil
}

func (c *client) XLen(ctx context.Context, stream string) (int64, error) {
	var length int64

	switch c.config.Mode {
	case Standalone:
		length, _ = c.cmdable.(*redis.Client).XLen(ctx, stream).Result()
	case Cluster:
		length, _ = c.cmdable.(*redis.ClusterClient).XLen(ctx, stream).Result()
	default:
		return 0, errors.NewErrorDetails("Unsupported Redis mode for XLen", string(errors.RedisXLenError), "xlen")
	}

	if length < 0 {
		return 0, errors.NewErrorDetails("Failed to get stream length", string(errors.RedisXLenError), "xlen")
	}
	return length, nil
}

func (c *client) XRead(ctx context.Context, args *redis.XReadArgs) ([]redis.XStream, error) {
	var streams []redis.XStream

	switch c.config.Mode {
	case Standalone:
		streams, _ = c.cmdable.(*redis.Client).XRead(ctx, args).Result()
	case Cluster:
		streams, _ = c.cmdable.(*redis.ClusterClient).XRead(ctx, args).Result()
	default:
		return nil, errors.NewErrorDetails("Unsupported Redis mode for XRead", string(errors.RedisXReadError), "xread")
	}

	if len(streams) == 0 {
		return nil, errors.NewErrorDetails("No streams found", string(errors.RedisXReadError), "xread")
	}
	return streams, nil
}

func (c *client) XReadGroup(ctx context.Context, args *redis.XReadGroupArgs) ([]redis.XStream, error) {
	var streams []redis.XStream

	switch c.config.Mode {
	case Standalone:
		streams, _ = c.cmdable.(*redis.Client).XReadGroup(ctx, args).Result()
	case Cluster:
		streams, _ = c.cmdable.(*redis.ClusterClient).XReadGroup(ctx, args).Result()
	default:
		return nil, errors.NewErrorDetails("Unsupported Redis mode for XReadGroup", string(errors.RedisXReadGroupError), "xreadgroup")
	}

	if len(streams) == 0 {
		return nil, errors.NewErrorDetails("No streams found in group", string(errors.RedisXReadGroupError), "xreadgroup")
	}
	return streams, nil
}
