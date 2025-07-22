package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/muhammadchandra19/exchange/pkg/logger"
	"github.com/muhammadchandra19/exchange/pkg/redis"
	app "github.com/muhammadchandra19/exchange/services/matching-service/internal/app/engine"
	matchpublisher "github.com/muhammadchandra19/exchange/services/matching-service/internal/usecase/match-publisher"
	orderreader "github.com/muhammadchandra19/exchange/services/matching-service/internal/usecase/order-reader"
	orderbook "github.com/muhammadchandra19/exchange/services/matching-service/internal/usecase/orderbook"
	snapshot "github.com/muhammadchandra19/exchange/services/matching-service/internal/usecase/snapshot"
	"github.com/muhammadchandra19/exchange/services/matching-service/pkg/config"
)

var cfg *config.Config
var log *logger.Logger

func init() {
	var err error
	cfg = &config.Config{}
	err = config.Load(cfg)
	if err != nil {
		panic(err)
	}

	logger, err := logger.NewLogger()
	if err != nil {
		panic(err)
	}

	log = logger
}

func main() {
	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	redisConfig := redis.DefaultConfig()
	redisConfig.Addrs = []string{cfg.RedisConfig.Addrs}
	redisConfig.Password = cfg.RedisConfig.Password
	redisConfig.Username = cfg.RedisConfig.Username
	redisConfig.DB = cfg.RedisConfig.DB
	// Initialize Redis client
	rclient := redis.NewClient(log, redisConfig)

	if err := rclient.Connect(ctx); err != nil {
		log.Error(err, logger.Field{
			Key:   "action",
			Value: "connect_redis",
		})
		return
	}

	// Initialize components
	ob := orderbook.NewOrderbook()
	oReader := orderreader.NewReader(cfg.KafkaConfig, *log)
	snapshotStore := snapshot.NewSnapshotStore(rclient, cfg.Pair, log)
	matchPublisher := matchpublisher.NewPublisher(cfg.MatchPublisherConfig, *log)
	engine := app.NewEngine(
		ob,
		oReader,
		snapshotStore,
		matchPublisher,
		log,
		cfg,
	)

	// Start the engine
	if err := engine.Start(ctx); err != nil {
		log.Error(err, logger.Field{
			Key:   "action",
			Value: "start_engine",
		})
		return
	}

	log.Info("Matching service started successfully", logger.Field{
		Key:   "pair",
		Value: cfg.Pair,
	})

	// Wait for shutdown signal
	sig := <-sigChan
	log.Info("Received shutdown signal", logger.Field{
		Key:   "signal",
		Value: sig.String(),
	})

	// Cancel the main context to signal shutdown
	cancel()

	// Create a timeout context for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Stop the engine gracefully
	if err := engine.Stop(shutdownCtx); err != nil {
		log.Error(err, logger.Field{
			Key:   "action",
			Value: "stop_engine",
		})
	}

	// Close Redis client if it has a close method
	if closer, ok := rclient.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			log.Error(err, logger.Field{
				Key:   "action",
				Value: "close_redis_client",
			})
		}
	}

	log.Info("Matching service shutdown complete")
}
