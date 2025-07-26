package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/muhammadchandra19/exchange/services/market-data-service/app/consumer"
	"github.com/muhammadchandra19/exchange/services/market-data-service/pkg/config"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
	}

	matchConsumer, err := consumer.InitMatchConsumer(ctx, *cfg)
	if err != nil {
		slog.Error("Failed to create match consumer", "error", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		matchConsumer.Consumer.Start(ctx)
	}()

	<-quit

	slog.Info("Shutting down match consumer...")
	cancel()
	matchConsumer.Consumer.Stop(ctx)

	slog.Info("Match consumer stopped")
}
