package redis

import "time"

// Mode represents the mode of the Redis client.
type Mode string

const (
	// Standalone Mode is for a single Redis instance.
	Standalone Mode = "standalone"
	// Cluster Mode is for a Redis cluster setup.
	Cluster Mode = "cluster"
)

// Config holds the configuration for the Redis client.
type Config struct {
	Mode     Mode   `env:"MODE" envDefault:"standalone"`
	Username string `env:"USERNAME"`
	Password string `env:"PASSWORD"`
	DB       int    `env:"DB" envDefault:"0"`

	Addrs []string `env:"ADDRS" envDefault:"localhost:6379"`

	ConnectTimeout  time.Duration `env:"CONNECT_TIMEOUT" envDefault:"5s"`
	MaxRetries      int           `env:"MAX_RETRIES" envDefault:"3"`
	MinRetryBackoff time.Duration `env:"MIN_RETRY_BACKOFF" envDefault:"100ms"`
	MaxRetryBackoff time.Duration `env:"MAX_RETRY_BACKOFF" envDefault:"2s"`
	PoolSize        int           `env:"POOL_SIZE" envDefault:"10"`
	MinIdleConns    int           `env:"MIN_IDLE_CONNS" envDefault:"2"`
	MaxIdleConns    int           `env:"MAX_IDLE_CONNS" envDefault:"10"`
	ConnMaxLifetime time.Duration `env:"CONN_MAX_LIFETIME" envDefault:"30m"`
	ConnMaxIdleTime time.Duration `env:"CONN_MAX_IDLE_TIME" envDefault:"10m"`
	PoolTimeout     time.Duration `env:"POOL_TIMEOUT" envDefault:"4s"`
	PrefixKey       string        `env:"PREFIX_KEY" envDefault:"exchange:"`
	DefaultTTL      time.Duration `env:"DEFAULT_TTL" envDefault:"5m"`

	LiveViewersTTLSeconds           int `env:"LIVE_VIEWERS_TTL_SECONDS" envDefault:"60"`
	LiveViewersLastHeartbeatSeconds int `env:"LIVE_VIEWERS_LAST_HEARTBEAT_SECONDS" envDefault:"30"`
	ReconnectMaxRetries             int `env:"RECONNECT_MAX_RETRIES" envDefault:"3"`
}

// DefaultConfig returns a default configuration for the Redis client.
func DefaultConfig() *Config {
	return &Config{
		Mode:            Standalone,
		ConnectTimeout:  5 * time.Second,
		MaxRetries:      3,
		MinRetryBackoff: 100 * time.Millisecond,
		MaxRetryBackoff: 2 * time.Second,
		PoolSize:        10,
		MinIdleConns:    2,
		MaxIdleConns:    10,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 10 * time.Minute,
		PoolTimeout:     4 * time.Second,
	}
}
