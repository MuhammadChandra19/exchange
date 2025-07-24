package rpc

import (
	"context"
	"time"

	"github.com/muhammadchandra19/exchange/pkg/logger"
	pb "github.com/muhammadchandra19/exchange/proto/go/modules/market-data-service/v1/public"
	"github.com/muhammadchandra19/exchange/proto/go/modules/market-data-service/v1/shared"
	"github.com/muhammadchandra19/exchange/services/market-data-service/internal/domain/tick"
	tickInfra "github.com/muhammadchandra19/exchange/services/market-data-service/internal/infrastructure/questdb/tick"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TickRPC is the service for the tick API.
type TickRPC struct {
	pb.UnimplementedTickServiceServer

	usecase tick.Usecase
	logger  logger.Interface
}

// NewTickRPC creates a new TickRPC.
func NewTickRPC(usecase tick.Usecase, logger logger.Interface) *TickRPC {
	return &TickRPC{
		usecase: usecase,
		logger:  logger,
	}
}

// GetLatestTick gets the latest tick for a given symbol.
func (s *TickRPC) GetLatestTick(ctx context.Context, req *pb.GetLatestTickRequest) (*pb.GetLatestTickResponse, error) {
	res, err := s.usecase.GetLatestTick(ctx, req.Symbol)
	if err != nil {
		s.logger.Error(err, logger.Field{
			Key:   "symbol",
			Value: req.Symbol,
		})
		return &pb.GetLatestTickResponse{
			Status:    "error",
			Message:   "failed to get latest tick",
			Data:      nil,
			Timestamp: timestamppb.New(time.Now()),
			Code:      codes.Internal.String(),
		}, err
	}

	return &pb.GetLatestTickResponse{
		Status:    "success",
		Message:   "success",
		Data:      res.ToProto(),
		Timestamp: timestamppb.New(time.Now()),
		Code:      codes.OK.String(),
	}, nil
}

// GetTickVolume gets the volume for a given symbol.
func (s *TickRPC) GetTickVolume(ctx context.Context, req *pb.GetTickVolumeRequest) (*pb.GetTickVolumeResponse, error) {
	from := req.From.AsTime()
	to := req.To.AsTime()
	volume, err := s.usecase.GetTickVolume(ctx, req.Symbol, from, to)
	if err != nil {
		s.logger.Error(err, logger.Field{
			Key:   "symbol",
			Value: req.Symbol,
		})
		return &pb.GetTickVolumeResponse{
			Status:    "error",
			Message:   "failed to get tick volume",
			Data:      0,
			Timestamp: timestamppb.New(time.Now()),
			Code:      codes.Internal.String(),
		}, err
	}

	return &pb.GetTickVolumeResponse{
		Status:    "success",
		Message:   "success",
		Data:      volume,
		Timestamp: timestamppb.New(time.Now()),
		Code:      codes.OK.String(),
	}, nil
}

// GetTicks gets the ticks for a given symbol.
func (s *TickRPC) GetTicks(ctx context.Context, req *pb.GetTicksRequest) (*pb.GetTicksResponse, error) {
	from := req.From.AsTime()
	to := req.To.AsTime()
	ticks, err := s.usecase.GetTicks(ctx, tickInfra.Filter{
		Symbol: req.Symbol,
		From:   &from,
		To:     &to,
		Limit:  int(req.Limit),
		Offset: int(req.Offset),
	})

	if err != nil {
		s.logger.Error(err, logger.Field{
			Key:   "symbol",
			Value: req.Symbol,
		})
		return &pb.GetTicksResponse{
			Status:    "error",
			Message:   "failed to get ticks",
			Data:      nil,
			Timestamp: timestamppb.New(time.Now()),
			Code:      codes.Internal.String(),
		}, err
	}

	protoTicks := make([]*shared.Tick, len(ticks))

	for i, tick := range ticks {
		protoTicks[i] = tick.ToProto()
	}

	return &pb.GetTicksResponse{
		Status:    "success",
		Message:   "success",
		Data:      protoTicks,
		Timestamp: timestamppb.New(time.Now()),
		Code:      codes.OK.String(),
	}, nil
}
