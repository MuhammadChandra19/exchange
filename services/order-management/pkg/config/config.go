package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/muhammadchandra19/exchange/pkg/questdb"
)

// Config represents the application configuration.
type Config struct {
	App        AppConfig        `envPrefix:"APP_"`
	QuestDB    questdb.Config   `envPrefix:"QUESTDB_"`
	OrderKafka OrderKafkaConfig `envPrefix:"ORDER_KAFKA_"`
}

// AppConfig represents the application configuration.
type AppConfig struct {
	Name        string `env:"NAME" envDefault:"order-management"`
	Environment string `env:"ENVIRONMENT" envDefault:"development"`
	GRPCPort    int    `env:"GRPC_PORT" envDefault:"7777"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
}

// OrderKafkaConfig represents the Kafka configuration.
type OrderKafkaConfig struct {
	Brokers         []string `env:"BROKERS" envSeparator:"," envDefault:"localhost:9092"`
	Topic           string   `env:"TOPIC" envDefault:"orders"`
	ConsumerGroup   string   `env:"CONSUMER_GROUP" envDefault:"market-data"`
	ConsumerTimeout int      `env:"CONSUMER_TIMEOUT" envDefault:"5"`
	MaxRetries      int      `env:"MAX_RETRIES" envDefault:"3"`
}

// Load loads the configuration from the environment.
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return cfg, nil
}
