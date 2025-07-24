package rpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	loggerMock "github.com/muhammadchandra19/exchange/pkg/logger/mock"
	pb "github.com/muhammadchandra19/exchange/proto/go/modules/market-data-service/v1/public"
	tickUcMock "github.com/muhammadchandra19/exchange/services/market-data-service/internal/domain/tick/mock"
	"github.com/muhammadchandra19/exchange/services/market-data-service/internal/infrastructure/questdb/tick"
	tickInfra "github.com/muhammadchandra19/exchange/services/market-data-service/internal/infrastructure/questdb/tick"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestTick_GetLatestTick(t *testing.T) {
	testCases := []struct {
		name       string
		mockFn     func(t *testing.T, testParams *pb.GetLatestTickRequest, tickUc *tickUcMock.MockUsecase, logger *loggerMock.MockInterface)
		assertFn   func(t *testing.T, res *pb.GetLatestTickResponse, err error)
		testParams *pb.GetLatestTickRequest
	}{
		{
			name: "success",
			mockFn: func(t *testing.T, testParams *pb.GetLatestTickRequest, tickUc *tickUcMock.MockUsecase, logger *loggerMock.MockInterface) {
				tickUc.EXPECT().GetLatestTick(gomock.Any(), testParams.Symbol).Return(&tick.Tick{
					Symbol: "BTCUSDT",
					Price:  10000,
					Volume: 100,
				}, nil)
			},
			assertFn: func(t *testing.T, res *pb.GetLatestTickResponse, err error) {
				assert.NoError(t, err)
				assert.Equal(t, codes.OK.String(), res.Code)
				assert.Equal(t, "BTCUSDT", res.Data.Symbol)
				assert.Equal(t, float64(10000), res.Data.Price)
				assert.Equal(t, int64(100), res.Data.Volume)
			},
			testParams: &pb.GetLatestTickRequest{
				Symbol: "BTCUSDT",
			},
		},
		{
			name: "error",
			mockFn: func(t *testing.T, testParams *pb.GetLatestTickRequest, tickUc *tickUcMock.MockUsecase, logger *loggerMock.MockInterface) {
				tickUc.EXPECT().GetLatestTick(gomock.Any(), testParams.Symbol).Return(nil, errors.New("error"))
				logger.EXPECT().Error(gomock.Any(), gomock.Any()).Times(1)
			},
			assertFn: func(t *testing.T, res *pb.GetLatestTickResponse, err error) {
				assert.Error(t, err)
				assert.Equal(t, codes.Internal.String(), res.Code)
			},
			testParams: &pb.GetLatestTickRequest{
				Symbol: "BTCUSDT",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tickUc := tickUcMock.NewMockUsecase(ctrl)
			logger := loggerMock.NewMockInterface(ctrl)

			testCase.mockFn(t, testCase.testParams, tickUc, logger)

			res, err := NewTickRpc(tickUc, logger).GetLatestTick(context.Background(), testCase.testParams)
			testCase.assertFn(t, res, err)
		})
	}
}

func TestTick_GetTickVolume(t *testing.T) {
	now := time.Now().UTC()
	nowProto := timestamppb.New(now)
	testCases := []struct {
		name       string
		mockFn     func(t *testing.T, testParams *pb.GetTickVolumeRequest, tickUc *tickUcMock.MockUsecase, logger *loggerMock.MockInterface)
		assertFn   func(t *testing.T, res *pb.GetTickVolumeResponse, err error)
		testParams *pb.GetTickVolumeRequest
	}{
		{
			name: "success",
			mockFn: func(t *testing.T, testParams *pb.GetTickVolumeRequest, tickUc *tickUcMock.MockUsecase, logger *loggerMock.MockInterface) {
				tickUc.EXPECT().GetTickVolume(gomock.Any(), testParams.Symbol, now, now).Return(int64(100), nil)
			},
			assertFn: func(t *testing.T, res *pb.GetTickVolumeResponse, err error) {
				assert.NoError(t, err)
				assert.Equal(t, codes.OK.String(), res.Code)
				assert.Equal(t, int64(100), res.Data)
			},
			testParams: &pb.GetTickVolumeRequest{
				Symbol: "BTCUSDT",
				From:   nowProto,
				To:     nowProto,
			},
		},
		{
			name: "error",
			mockFn: func(t *testing.T, testParams *pb.GetTickVolumeRequest, tickUc *tickUcMock.MockUsecase, logger *loggerMock.MockInterface) {
				tickUc.EXPECT().GetTickVolume(gomock.Any(), testParams.Symbol, now, now).Return(int64(0), errors.New("error"))
				logger.EXPECT().Error(gomock.Any(), gomock.Any()).Times(1)
			},
			assertFn: func(t *testing.T, res *pb.GetTickVolumeResponse, err error) {
				assert.Error(t, err)
				assert.Equal(t, codes.Internal.String(), res.Code)
			},
			testParams: &pb.GetTickVolumeRequest{
				Symbol: "BTCUSDT",
				From:   nowProto,
				To:     nowProto,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tickUc := tickUcMock.NewMockUsecase(ctrl)
			logger := loggerMock.NewMockInterface(ctrl)

			testCase.mockFn(t, testCase.testParams, tickUc, logger)

			res, err := NewTickRpc(tickUc, logger).GetTickVolume(context.Background(), testCase.testParams)
			testCase.assertFn(t, res, err)
		})
	}
}
func TestTick_GetTicks(t *testing.T) {
	now := time.Now().UTC()
	nowProto := timestamppb.New(now)
	testCases := []struct {
		name       string
		mockFn     func(t *testing.T, testParams *pb.GetTicksRequest, tickUc *tickUcMock.MockUsecase, logger *loggerMock.MockInterface)
		assertFn   func(t *testing.T, res *pb.GetTicksResponse, err error)
		testParams *pb.GetTicksRequest
	}{
		{
			name: "success",
			mockFn: func(t *testing.T, testParams *pb.GetTicksRequest, tickUc *tickUcMock.MockUsecase, logger *loggerMock.MockInterface) {
				tickUc.EXPECT().GetTicks(gomock.Any(), tickInfra.Filter{
					Symbol: testParams.Symbol,
					From:   &now,
					To:     &now,
					Limit:  int(testParams.Limit),
					Offset: int(testParams.Offset),
				}).Return([]*tick.Tick{
					{
						Symbol: "BTCUSDT",
						Price:  10000,
						Volume: 100,
					},
				}, nil)
			},
			assertFn: func(t *testing.T, res *pb.GetTicksResponse, err error) {
				assert.NoError(t, err)
				assert.Equal(t, codes.OK.String(), res.Code)
				assert.Equal(t, "BTCUSDT", res.Data[0].Symbol)
				assert.Equal(t, float64(10000), res.Data[0].Price)
				assert.Equal(t, int64(100), res.Data[0].Volume)
			},
			testParams: &pb.GetTicksRequest{
				Symbol: "BTCUSDT",
				From:   nowProto,
				To:     nowProto,
				Limit:  10,
				Offset: 0,
			},
		},
		{
			name: "error",
			mockFn: func(t *testing.T, testParams *pb.GetTicksRequest, tickUc *tickUcMock.MockUsecase, logger *loggerMock.MockInterface) {
				tickUc.EXPECT().GetTicks(gomock.Any(), tickInfra.Filter{
					Symbol: testParams.Symbol,
					From:   &now,
					To:     &now,
					Limit:  int(testParams.Limit),
					Offset: int(testParams.Offset),
				}).Return(nil, errors.New("error"))
				logger.EXPECT().Error(gomock.Any(), gomock.Any()).Times(1)
			},
			assertFn: func(t *testing.T, res *pb.GetTicksResponse, err error) {
				assert.Error(t, err)
				assert.Equal(t, codes.Internal.String(), res.Code)
			},
			testParams: &pb.GetTicksRequest{
				Symbol: "BTCUSDT",
				From:   nowProto,
				To:     nowProto,
				Limit:  10,
				Offset: 0,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tickUc := tickUcMock.NewMockUsecase(ctrl)
			logger := loggerMock.NewMockInterface(ctrl)

			testCase.mockFn(t, testCase.testParams, tickUc, logger)

			res, err := NewTickRpc(tickUc, logger).GetTicks(context.Background(), testCase.testParams)
			testCase.assertFn(t, res, err)
		})
	}
}
