package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// MustLoad loads the configuration from environment variables and .env file.
func MustLoad[T any](cfg T) {
	_ = godotenv.Load() // Load environment variables from .env file

	env.Must(cfg, env.Parse(cfg))
}

// Load loads the configuration from environment variables and .env file.
func Load[T any](cfg T) error {
	if err := godotenv.Load(); err != nil {
		return err // Return error if .env file loading fails
	}

	if err := env.Parse(cfg); err != nil {
		return err // Return error if environment variable parsing fails
	}

	return nil // Return nil if everything is successful
}

// Config holds the configuration for the application
type Config struct {
	Pair        string               `env:"PAIR,required"` // Trading pair, e.g., BTC/USD
	KafkaConfig `envPrefix:"KAFKA_"` // Kafka configuration
	RedisConfig `envPrefix:"REDIS_"` // Redis configuration

}

// KafkaConfig holds the configuration for Kafka consumer and producer.
type KafkaConfig struct {
	Topic   string   `env:"TOPIC,required"`
	GroupID string   `env:"GROUP_ID" envDefault:"default_group"`
	Brokers []string `env:"BROKER,required"`
}

// RedisConfig holds the configuration for Redis client.
type RedisConfig struct {
	Addrs          string `env:"ADDRESS,required"` // Comma-separated list of Redis addresses
	Password       string `env:"PASSWORD" envDefault:""`
	Username       string `env:"USERNAME" envDefault:""`
	DB             int    `env:"DB" envDefault:"0"`
	DefaultChannel string `env:"DEFAULT_CHANNEL" envDefault:"exchange"`
}
