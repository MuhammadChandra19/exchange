package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/muhammadchandra19/exchange/services/market-data/app/consumer"
	"github.com/muhammadchandra19/exchange/services/market-data/pkg/config"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
	}

	orderConsumer, err := consumer.InitOrderConsumer(ctx, *cfg)
	if err != nil {
		slog.Error("Failed to create order consumer", "error", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		orderConsumer.Consumer.Start(ctx)
	}()

	go func() {
		defer wg.Done()
		orderConsumer.Consumer.Subscribe(ctx)
	}()

	<-quit

	slog.Info("Shutting down order consumer...")
	cancel()
	orderConsumer.Consumer.Stop()

	slog.Info("Order consumer stopped")
}
