package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/muhammadchandra19/exchange/services/market-data-service/internal/infrastructure/questdb"
	// tickrepo "github.com/muhammadchandra19/exchange/services/market-data-service/internal/infrastructure/questdb/tick"
	// ohlcrepo "github.com/muhammadchandra19/exchange/services/market-data-service/internal/infrastructure/questdb/ohlc"
	"github.com/muhammadchandra19/exchange/services/market-data-service/pkg/config"
)

func main() {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger, err := initLogger(cfg.App.LogLevel)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Initialize QuestDB client - returns interface
	questdbClient, err := questdb.NewClient(ctx, cfg.QuestDB)
	if err != nil {
		logger.Fatal("Failed to initialize QuestDB client", zap.Error(err))
	}
	defer questdbClient.Close()

	logger.Info("QuestDB client connected successfully")

	// Test connection
	if err := questdbClient.Ping(ctx); err != nil {
		logger.Fatal("Failed to ping QuestDB", zap.Error(err))
	}

	// Initialize repositories with interface
	// tickRepo := tickrepo.NewRepository(questdbClient)
	// ohlcRepo := ohlcrepo.NewRepository(questdbClient)

	logger.Info("Market Data Service started successfully",
		zap.String("app", cfg.App.Name),
		zap.String("environment", cfg.App.Environment),
		zap.Int("http_port", cfg.App.Port),
		zap.Int("grpc_port", cfg.App.GRPCPort),
	)

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Market Data Service...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Add any cleanup logic here
	_ = shutdownCtx

	logger.Info("Market Data Service stopped")
}

func initLogger(level string) (*zap.Logger, error) {
	var config zap.Config
	
	switch level {
	case "debug":
		config = zap.NewDevelopmentConfig()
	case "production":
		config = zap.NewProductionConfig()
	default:
		config = zap.NewDevelopmentConfig()
	}

	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return logger, nil
}
```