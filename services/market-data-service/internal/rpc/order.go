package rpc

import (
	"context"
	"time"

	"github.com/muhammadchandra19/exchange/pkg/logger"
	pb "github.com/muhammadchandra19/exchange/proto/go/modules/market-data-service/v1/public"
	"github.com/muhammadchandra19/exchange/proto/go/modules/market-data-service/v1/shared"
	"github.com/muhammadchandra19/exchange/services/market-data-service/internal/domain/order"
	orderInfra "github.com/muhammadchandra19/exchange/services/market-data-service/internal/infrastructure/questdb/order"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// OrderRPC is the RPC server for the order service.
type OrderRPC struct {
	pb.UnimplementedOrderServiceServer

	usecase order.Usecase
	logger  logger.Interface
}

// NewOrderRPC creates a new OrderRPC.
func NewOrderRPC(usecase order.Usecase, logger logger.Interface) *OrderRPC {
	return &OrderRPC{
		usecase: usecase,
		logger:  logger,
	}
}

// GetPairActiveOrders gets the active orders for a given symbol and side.
func (r *OrderRPC) GetPairActiveOrders(ctx context.Context, req *pb.GetPairActiveOrdersRequest) (*pb.GetPairActiveOrdersResponse, error) {
	orders, err := r.usecase.GetPairActiveOrders(ctx, req.Symbol, req.Side, int(req.Limit), int(req.Offset))
	if err != nil {
		r.logger.Error(err, logger.Field{
			Key:   "symbol",
			Value: req.Symbol,
		})
		return &pb.GetPairActiveOrdersResponse{
			Status:    "error",
			Message:   "failed to get pair active orders",
			Timestamp: timestamppb.New(time.Now()),
			Data:      nil,
			Code:      codes.Internal.String(),
		}, err
	}

	protoOrders := make([]*shared.Order, len(orders))
	for i, order := range orders {
		protoOrders[i] = order.ToProto()
	}

	return &pb.GetPairActiveOrdersResponse{
		Status:    "success",
		Message:   "success",
		Timestamp: timestamppb.New(time.Now()),
		Data:      protoOrders,
		Code:      codes.OK.String(),
	}, nil
}

// GetOrder gets the order for a given order ID.
func (r *OrderRPC) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	order, err := r.usecase.GetOrder(ctx, req.OrderID)
	if err != nil {
		r.logger.Error(err, logger.Field{
			Key:   "order_id",
			Value: req.OrderID,
		})
		return &pb.GetOrderResponse{
			Status:    "error",
			Message:   "failed to get order",
			Timestamp: timestamppb.New(time.Now()),
			Data:      nil,
			Code:      codes.Internal.String(),
		}, err
	}
	return &pb.GetOrderResponse{
		Status:    "success",
		Message:   "success",
		Timestamp: timestamppb.New(time.Now()),
		Data:      order.ToProto(),
		Code:      codes.OK.String(),
	}, nil
}

// GetOrders gets the orders for a given filter.
func (r *OrderRPC) GetOrders(ctx context.Context, req *pb.GetOrdersRequest) (*pb.GetOrdersResponse, error) {
	var from *time.Time
	var to *time.Time
	if req.From != nil {
		t := req.From.AsTime()
		from = &t
	}
	if req.To != nil {
		t := req.To.AsTime()
		to = &t
	}
	orders, err := r.usecase.GetOrderByFilter(ctx, orderInfra.OrderFilter{
		Symbol: req.Symbol,
		Side:   req.Side,
		Status: req.Status,
		UserID: req.UserID,
		From:   from,
		To:     to,
		Limit:  int(req.Limit),
		Offset: int(req.Offset),
	})

	if err != nil {
		r.logger.Error(err, logger.Field{
			Key:   "symbol",
			Value: req.Symbol,
		})
		return &pb.GetOrdersResponse{
			Status:    "error",
			Message:   "failed to get orders",
			Timestamp: timestamppb.New(time.Now()),
			Data:      nil,
			Code:      codes.Internal.String(),
		}, err
	}

	protoOrders := make([]*shared.Order, len(orders))
	for i, order := range orders {
		protoOrders[i] = order.ToProto()
	}

	return &pb.GetOrdersResponse{
		Status:    "success",
		Message:   "success",
		Timestamp: timestamppb.New(time.Now()),
		Data:      protoOrders,
		Code:      codes.OK.String(),
	}, nil
}
