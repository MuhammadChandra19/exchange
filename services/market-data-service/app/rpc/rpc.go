package rpc

import (
	"context"

	"github.com/muhammadchandra19/exchange/pkg/grpclib/health"
	"github.com/muhammadchandra19/exchange/pkg/logger"
	"github.com/muhammadchandra19/exchange/pkg/questdb"
	orderPublic "github.com/muhammadchandra19/exchange/proto/go/modules/market-data-service/v1/public"
	tickPublic "github.com/muhammadchandra19/exchange/proto/go/modules/market-data-service/v1/public"
	"github.com/muhammadchandra19/exchange/services/market-data-service/internal/bootstrap"
	orderInfra "github.com/muhammadchandra19/exchange/services/market-data-service/internal/infrastructure/questdb/order"
	tickInfra "github.com/muhammadchandra19/exchange/services/market-data-service/internal/infrastructure/questdb/tick"
	"github.com/muhammadchandra19/exchange/services/market-data-service/internal/rpc"
	orderUc "github.com/muhammadchandra19/exchange/services/market-data-service/internal/usecase/order"
	tickUc "github.com/muhammadchandra19/exchange/services/market-data-service/internal/usecase/tick"
	"github.com/muhammadchandra19/exchange/services/market-data-service/pkg/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GrpcServer is the gRPC server.
type GrpcServer struct {
	Server     *grpc.Server
	logger     logger.Interface
	Config     config.Config
	usecase    bootstrap.Usecase
	repository bootstrap.Repository
	rpc        bootstrap.RPC
	db         questdb.QuestDBClient
}

// Config is the RPC config.
type Config struct {
	Server *grpc.Server
}

// NewGrpcServer creates a new gRPC server.
func NewGrpcServer(ctx context.Context, config config.AppConfig) (*GrpcServer, error) {
	logger, err := logger.NewLogger()
	if err != nil {
		return nil, err
	}

	server := &GrpcServer{
		Server:     grpc.NewServer(),
		logger:     logger,
		usecase:    bootstrap.Usecase{},
		repository: bootstrap.Repository{},
		rpc:        bootstrap.RPC{},
	}

	// Register health service
	healthService := health.NewServer()
	healthService.Register(server.Server)

	server.initDB(ctx)

	server.registerRepository()
	server.registerUsecase()
	server.registerPublicRPC()

	server.registerGrpcServer()

	if config.Environment == "development" {
		reflection.Register(server.Server)
	}

	return server, nil
}

// Stop stops the gRPC server.
func (s *GrpcServer) Stop() {
	s.Server.GracefulStop()
	s.db.Close()
}

func (s *GrpcServer) initDB(ctx context.Context) {
	questdbClient, err := questdb.NewClient(ctx, s.Config.QuestDB)
	if err != nil {
		s.logger.GetZap().DPanic("Failed to initialize QuestDB client", zap.Error(err))
	}

	s.db = questdbClient
}

func (s *GrpcServer) registerRepository() {
	s.repository.OrderRepository = orderInfra.NewRepository(s.db)
	s.repository.TickRepository = tickInfra.NewRepository(s.db)
}

func (s *GrpcServer) registerUsecase() {
	s.usecase.OrderUsecase = orderUc.NewUsecase(s.repository.OrderRepository, s.logger)
	s.usecase.TickUsecase = tickUc.NewUsecase(s.repository.TickRepository, s.logger)
}

func (s *GrpcServer) registerPublicRPC() {
	s.rpc.OrderRPC = rpc.NewOrderRPC(s.usecase.OrderUsecase, s.logger)
	s.rpc.TickRPC = rpc.NewTickRPC(s.usecase.TickUsecase, s.logger)
}

func (s *GrpcServer) registerGrpcServer() {
	orderPublic.RegisterOrderServiceServer(s.Server, s.rpc.OrderRPC)
	tickPublic.RegisterTickServiceServer(s.Server, s.rpc.TickRPC)
}
