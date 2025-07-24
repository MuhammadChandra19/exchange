package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/muhammadchandra19/exchange/services/market-data-service/app/rpc"
	"github.com/muhammadchandra19/exchange/services/market-data-service/pkg/config"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
	}
	grpcServer, err := rpc.NewGrpcServer(ctx, cfg.App)
	if err != nil {
		slog.Error("Failed to create gRPC server", "error", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.App.GRPCPort))
	if err != nil {
		slog.Error("Failed to listen", "error", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func(gs *grpc.Server, lis net.Listener) {
		if err := gs.Serve(lis); err != nil {
			slog.Error("Failed to serve gRPC server", "error", err)
		}
	}(grpcServer.Server, lis)

	<-quit

	slog.Info("Shutting down gRPC server...")
	grpcServer.Stop()

	slog.Info("gRPC server stopped")
}
