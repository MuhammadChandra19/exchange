package rpc

import (
	"context"
	"time"

	"github.com/muhammadchandra19/exchange/pkg/logger"
	pb "github.com/muhammadchandra19/exchange/proto/go/modules/market-data-service/v1/public"
	"github.com/muhammadchandra19/exchange/services/market-data-service/internal/domain/ohlc"
	ohlcInfra "github.com/muhammadchandra19/exchange/services/market-data-service/internal/infrastructure/questdb/ohlc"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// OHLCRPC is the RPC server for the OHLC service.
type OHLCRPC struct {
	pb.UnimplementedOHLCServiceServer

	usecase ohlc.Usecase
	logger  logger.Interface
}

// NewOHLCRPC creates a new OHLC RPC.
func NewOHLCRPC(usecase ohlc.Usecase, logger logger.Interface) *OHLCRPC {
	return &OHLCRPC{
		usecase: usecase,
		logger:  logger,
	}
}

// GetOHLC gets the OHLC for a given symbol and interval.
func (r *OHLCRPC) GetOHLC(ctx context.Context, req *pb.GetOHLCRequest) (*pb.GetOHLCResponse, error) {
	ohlc, err := r.usecase.GetOHLC(ctx, req.Symbol, req.Interval)
	if err != nil {
		r.logger.Error(err, logger.Field{
			Key:   "symbol",
			Value: req.Symbol,
		})
		return nil, err
	}
	return &pb.GetOHLCResponse{
		Status:    "success",
		Message:   "success",
		Timestamp: timestamppb.New(time.Now()),
		Code:      codes.OK.String(),
		Data:      ohlc.ToProto(),
	}, nil
}

// GetOHLCByFilter gets the OHLC for a given filter.
func (r *OHLCRPC) GetOHLCByFilter(ctx context.Context, req *pb.GetOHLCByFilterRequest) (*pb.GetOHLCByFilterResponse, error) {
	var from *time.Time
	var to *time.Time
	if req.From != nil {
		fromTime := req.From.AsTime()
		from = &fromTime
	}
	if req.To != nil {
		toTime := req.To.AsTime()
		to = &toTime
	}

	ohlcs, err := r.usecase.GetOHLCByFilter(ctx, ohlcInfra.OHLCFilter{
		Symbol:   req.Symbol,
		Interval: req.Interval,
		From:     from,
		To:       to,
		Limit:    req.Limit,
		Offset:   req.Offset,
	})

	ohlcList := ohlcInfra.List(ohlcs)
	if err != nil {
		r.logger.Error(err, logger.Field{
			Key:   "symbol",
			Value: req.Symbol,
		})
		return nil, err
	}
	return &pb.GetOHLCByFilterResponse{
		Status:    "success",
		Message:   "success",
		Timestamp: timestamppb.New(time.Now()),
		Code:      codes.OK.String(),
		Data:      ohlcList.ToProtoList(),
	}, nil
}

// GetIntradayData gets the OHLC for a given symbol and interval.
func (r *OHLCRPC) GetIntradayData(ctx context.Context, req *pb.GetIntradayDataRequest) (*pb.GetIntradayDataResponse, error) {
	ohlcs, err := r.usecase.GetIntradayData(ctx, req.Symbol, req.Interval, int(req.Limit))
	if err != nil {
		r.logger.Error(err, logger.Field{
			Key:   "symbol",
			Value: req.Symbol,
		})
		return nil, err
	}

	ohlcList := ohlcInfra.List(ohlcs)
	return &pb.GetIntradayDataResponse{
		Status:    "success",
		Message:   "success",
		Timestamp: timestamppb.New(time.Now()),
		Code:      codes.OK.String(),
		Data:      ohlcList.ToProtoList(),
	}, nil
}
