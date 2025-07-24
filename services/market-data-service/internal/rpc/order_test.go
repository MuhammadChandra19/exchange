package rpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	loggerMock "github.com/muhammadchandra19/exchange/pkg/logger/mock"
	pb "github.com/muhammadchandra19/exchange/proto/go/modules/market-data-service/v1/public"
	orderUcMock "github.com/muhammadchandra19/exchange/services/market-data-service/internal/domain/order/mock"
	"github.com/muhammadchandra19/exchange/services/market-data-service/internal/infrastructure/questdb/order"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestOrder_GetActiveOrders(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name       string
		mockFn     func(t *testing.T, testParams *pb.GetPairActiveOrdersRequest, orderUc *orderUcMock.MockUsecase, logger *loggerMock.MockInterface)
		assertFn   func(t *testing.T, res *pb.GetPairActiveOrdersResponse, err error)
		testParams *pb.GetPairActiveOrdersRequest
	}{
		{
			name: "success",
			mockFn: func(t *testing.T, testParams *pb.GetPairActiveOrdersRequest, orderUc *orderUcMock.MockUsecase, logger *loggerMock.MockInterface) {
				orderUc.EXPECT().
					GetPairActiveOrders(gomock.Any(), testParams.Symbol, testParams.Side, int(testParams.Limit), int(testParams.Offset)).
					Return([]*order.Order{
						{
							ID:        "123",
							Symbol:    "BTCUSDT",
							Side:      "BUY",
							Price:     float64(10000),
							Quantity:  int64(1),
							Status:    "OPEN",
							Timestamp: now,
						},
					}, nil)
			},
			assertFn: func(t *testing.T, res *pb.GetPairActiveOrdersResponse, err error) {
				assert.NoError(t, err)
				assert.Equal(t, codes.OK.String(), res.Code)
				assert.Equal(t, 1, len(res.Data))
				assert.Equal(t, "123", res.Data[0].Id)
				assert.Equal(t, "BTCUSDT", res.Data[0].Symbol)
				assert.Equal(t, "BUY", res.Data[0].Side)
				assert.Equal(t, float64(10000), res.Data[0].Price)
				assert.Equal(t, int64(1), res.Data[0].Quantity)
			},
			testParams: &pb.GetPairActiveOrdersRequest{
				Symbol: "BTCUSDT",
				Side:   "BUY",
				Limit:  10,
				Offset: 0,
			},
		},
		{
			name: "error",
			mockFn: func(t *testing.T, testParams *pb.GetPairActiveOrdersRequest, orderUc *orderUcMock.MockUsecase, logger *loggerMock.MockInterface) {
				orderUc.EXPECT().
					GetPairActiveOrders(gomock.Any(), testParams.Symbol, testParams.Side, int(testParams.Limit), int(testParams.Offset)).
					Return(nil, errors.New("error"))
				logger.EXPECT().Error(gomock.Any(), gomock.Any()).Times(1)
			},
			assertFn: func(t *testing.T, res *pb.GetPairActiveOrdersResponse, err error) {
				assert.Error(t, err)
				assert.Equal(t, codes.Internal.String(), res.Code)
			},
			testParams: &pb.GetPairActiveOrdersRequest{
				Symbol: "BTCUSDT",
				Side:   "BUY",
				Limit:  10,
				Offset: 0,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			orderUc := orderUcMock.NewMockUsecase(ctrl)
			mockLogger := loggerMock.NewMockInterface(ctrl)
			tc.mockFn(t, tc.testParams, orderUc, mockLogger)

			orderService := NewOrderRPC(orderUc, mockLogger)
			res, err := orderService.GetPairActiveOrders(context.Background(), tc.testParams)
			tc.assertFn(t, res, err)
		})
	}
}

func TestOrder_GetOrder(t *testing.T) {
	testCases := []struct {
		name       string
		mockFn     func(t *testing.T, testParams *pb.GetOrderRequest, orderUc *orderUcMock.MockUsecase, logger *loggerMock.MockInterface)
		assertFn   func(t *testing.T, res *pb.GetOrderResponse, err error)
		testParams *pb.GetOrderRequest
	}{
		{
			name: "success",
			mockFn: func(t *testing.T, testParams *pb.GetOrderRequest, orderUc *orderUcMock.MockUsecase, logger *loggerMock.MockInterface) {
				orderUc.EXPECT().
					GetOrder(gomock.Any(), testParams.OrderID).
					Return(&order.Order{
						ID: "123",
					}, nil)
			},
			assertFn: func(t *testing.T, res *pb.GetOrderResponse, err error) {
				assert.NoError(t, err)
				assert.Equal(t, codes.OK.String(), res.Code)
				assert.Equal(t, "123", res.Data.Id)
			},
			testParams: &pb.GetOrderRequest{
				OrderID: "123",
			},
		},
		{
			name: "error",
			mockFn: func(t *testing.T, testParams *pb.GetOrderRequest, orderUc *orderUcMock.MockUsecase, logger *loggerMock.MockInterface) {
				orderUc.EXPECT().
					GetOrder(gomock.Any(), testParams.OrderID).
					Return(nil, errors.New("error"))
				logger.EXPECT().Error(gomock.Any(), gomock.Any()).Times(1)
			},
			assertFn: func(t *testing.T, res *pb.GetOrderResponse, err error) {
				assert.Error(t, err)
				assert.Equal(t, codes.Internal.String(), res.Code)
			},
			testParams: &pb.GetOrderRequest{
				OrderID: "123",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			orderUc := orderUcMock.NewMockUsecase(ctrl)
			mockLogger := loggerMock.NewMockInterface(ctrl)
			tc.mockFn(t, tc.testParams, orderUc, mockLogger)

			orderService := NewOrderRPC(orderUc, mockLogger)
			res, err := orderService.GetOrder(context.Background(), tc.testParams)
			tc.assertFn(t, res, err)
		})
	}
}

func TestOrder_GetOrders(t *testing.T) {
	now := time.Now().UTC()
	nowProto := timestamppb.New(now)
	testCases := []struct {
		name       string
		mockFn     func(t *testing.T, testParams *pb.GetOrdersRequest, orderUc *orderUcMock.MockUsecase, logger *loggerMock.MockInterface)
		assertFn   func(t *testing.T, res *pb.GetOrdersResponse, err error)
		testParams *pb.GetOrdersRequest
	}{
		{
			name: "success",
			mockFn: func(t *testing.T, testParams *pb.GetOrdersRequest, orderUc *orderUcMock.MockUsecase, logger *loggerMock.MockInterface) {
				orderUc.EXPECT().
					GetOrderByFilter(gomock.Any(), order.OrderFilter{
						Symbol: testParams.Symbol,
						Side:   testParams.Side,
						Status: testParams.Status,
						UserID: testParams.UserID,
						Limit:  int(testParams.Limit),
						Offset: int(testParams.Offset),
						From:   &now,
						To:     &now,
					}).
					Return([]*order.Order{
						{
							ID:     "123",
							Symbol: "BTCUSDT",
						},
					}, nil)
			},
			assertFn: func(t *testing.T, res *pb.GetOrdersResponse, err error) {
				assert.NoError(t, err)
				assert.Equal(t, codes.OK.String(), res.Code)
				assert.Equal(t, 1, len(res.Data))
				assert.Equal(t, "123", res.Data[0].Id)
				assert.Equal(t, "BTCUSDT", res.Data[0].Symbol)
			},
			testParams: &pb.GetOrdersRequest{
				Symbol: "BTCUSDT",
				Side:   "BUY",
				Status: "OPEN",
				UserID: "123",
				From:   nowProto,
				To:     nowProto,
				Limit:  10,
				Offset: 0,
			},
		},
		{
			name: "error",
			mockFn: func(t *testing.T, testParams *pb.GetOrdersRequest, orderUc *orderUcMock.MockUsecase, logger *loggerMock.MockInterface) {
				orderUc.EXPECT().
					GetOrderByFilter(gomock.Any(), order.OrderFilter{
						Symbol: testParams.Symbol,
						Side:   testParams.Side,
						Status: testParams.Status,
						UserID: testParams.UserID,
						Limit:  int(testParams.Limit),
						Offset: int(testParams.Offset),
						From:   &now,
						To:     &now,
					}).
					Return(nil, errors.New("error"))
				logger.EXPECT().Error(gomock.Any(), gomock.Any()).Times(1)
			},
			assertFn: func(t *testing.T, res *pb.GetOrdersResponse, err error) {
				assert.Error(t, err)
				assert.Equal(t, codes.Internal.String(), res.Code)
			},
			testParams: &pb.GetOrdersRequest{
				Symbol: "BTCUSDT",
				Side:   "BUY",
				Status: "OPEN",
				UserID: "123",
				From:   nowProto,
				To:     nowProto,
				Limit:  10,
				Offset: 0,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			orderUc := orderUcMock.NewMockUsecase(ctrl)
			mockLogger := loggerMock.NewMockInterface(ctrl)
			tc.mockFn(t, tc.testParams, orderUc, mockLogger)

			orderService := NewOrderRPC(orderUc, mockLogger)
			res, err := orderService.GetOrders(context.Background(), tc.testParams)
			tc.assertFn(t, res, err)
		})
	}
}
