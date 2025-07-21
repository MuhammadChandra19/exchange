package ohlc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5"
	ohlcdomain "github.com/muhammadchandra19/exchange/services/market-data-service/internal/domain/ohlc/v1"
	mockOhlc "github.com/muhammadchandra19/exchange/services/market-data-service/internal/infrastructure/questdb/mock"
	"github.com/stretchr/testify/assert"
)

func TestOhlc_Store(t *testing.T) {
	query := `INSERT INTO ohlc (timestamp, symbol, interval, open, high, low, close, volume, trade_count) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	now := time.Now()
	testCases := []struct {
		name     string
		mockFn   func(testData *ohlcdomain.OHLC, mock *mockOhlc.MockQuestDBClient)
		assertFn func(t *testing.T, err error)
		testData *ohlcdomain.OHLC
	}{
		{
			name: "success",
			mockFn: func(testData *ohlcdomain.OHLC, mock *mockOhlc.MockQuestDBClient) {
				mock.EXPECT().Exec(
					gomock.Any(),
					query,
					testData.Timestamp,
					testData.Symbol,
					testData.Interval,
					testData.Open,
					testData.High,
					testData.Low,
					testData.Close,
					testData.Volume,
					testData.TradeCount,
				).Return(nil)
			},
			assertFn: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			testData: &ohlcdomain.OHLC{
				Timestamp:  now,
				Symbol:     "BTCUSDT",
				Interval:   "1m",
				Open:       10000,
				High:       10000,
				Low:        9000,
				Close:      10500,
				Volume:     100000,
				TradeCount: 100,
			},
		},
		{
			name: "error - exec fails",
			mockFn: func(testData *ohlcdomain.OHLC, mock *mockOhlc.MockQuestDBClient) {
				mock.EXPECT().Exec(
					gomock.Any(),
					query,
					testData.Timestamp,
					testData.Symbol,
					testData.Interval,
					testData.Open,
					testData.High,
					testData.Low,
					testData.Close,
					testData.Volume,
					testData.TradeCount,
				).Return(errors.New("exec failed"))
			},
			assertFn: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
			testData: &ohlcdomain.OHLC{
				Timestamp:  now,
				Symbol:     "BTCUSDT",
				Interval:   "1m",
				Open:       10000,
				High:       10000,
				Low:        9000,
				Close:      10500,
				Volume:     100000,
				TradeCount: 100,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mockOhlc.NewMockQuestDBClient(ctrl)

			tc.mockFn(tc.testData, mockClient)

			repo := NewRepository(mockClient)
			err := repo.Store(context.Background(), tc.testData)
			tc.assertFn(t, err)
		})
	}
}

func TestOhlc_StoreBatch(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name     string
		mockFn   func(testData []*ohlcdomain.OHLC, mock *mockOhlc.MockQuestDBClient)
		assertFn func(t *testing.T, err error)
		testData []*ohlcdomain.OHLC
	}{
		{
			name: "success",
			mockFn: func(testData []*ohlcdomain.OHLC, mock *mockOhlc.MockQuestDBClient) {
				mock.EXPECT().CopyFrom(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			},
			assertFn: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			testData: []*ohlcdomain.OHLC{
				{
					Timestamp: now,
					Symbol:    "BTCUSDT",
					Interval:  "1m",
				},
			},
		},
		{
			name: "error - copy from fails",
			mockFn: func(testData []*ohlcdomain.OHLC, mock *mockOhlc.MockQuestDBClient) {
				mock.EXPECT().CopyFrom(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), errors.New("copy from failed"))
			},
			assertFn: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
			testData: []*ohlcdomain.OHLC{
				{
					Timestamp: now,
					Symbol:    "BTCUSDT",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mockOhlc.NewMockQuestDBClient(ctrl)

			tc.mockFn(tc.testData, mockClient)

			repo := NewRepository(mockClient)
			err := repo.StoreBatch(context.Background(), tc.testData)
			tc.assertFn(t, err)
		})
	}

}

func TestOhlc_GetByFilter(t *testing.T) {
	query := "SELECT timestamp, symbol, interval, open, high, low, close, volume, trade_count FROM ohlc WHERE 1=1"
	now := time.Now()
	testCases := []struct {
		name     string
		mockFn   func(mock *mockOhlc.MockQuestDBClient, mockRows *mockOhlc.MockRowsInterface)
		assertFn func(t *testing.T, err error, ticks []*ohlcdomain.OHLC)
		filter   ohlcdomain.OHLCFilter
	}{
		{
			name: "success: with all filters",
			mockFn: func(mock *mockOhlc.MockQuestDBClient, mockRows *mockOhlc.MockRowsInterface) {
				mock.EXPECT().Query(
					gomock.Any(),
					query+" AND symbol = $1 AND interval = $2 AND timestamp >= $3 AND timestamp <= $4 ORDER BY timestamp DESC LIMIT $5",
					[]interface{}{"BTCUSDT", "1m", now, now, 10},
				).Return(mockRows, nil)

				mockRows.EXPECT().Next().Return(true)
				mockRows.EXPECT().Scan(gomock.Any()).DoAndReturn(func(dest ...any) error {
					*dest[0].(*time.Time) = now
					*dest[1].(*string) = "BTCUSDT"
					*dest[2].(*string) = "1m"
					*dest[3].(*float64) = 10000
					*dest[4].(*float64) = 10000
					*dest[5].(*float64) = 9000
					*dest[6].(*float64) = 10500
					*dest[7].(*int64) = 100000
					*dest[8].(*int64) = 100
					return nil
				})
				mockRows.EXPECT().Next().Return(false)
				mockRows.EXPECT().Err().Return(nil)
				mockRows.EXPECT().Close()
			},
			filter: ohlcdomain.OHLCFilter{
				Symbol:   "BTCUSDT",
				Interval: "1m",
				From:     &now,
				To:       &now,
				Limit:    10,
			},
			assertFn: func(t *testing.T, err error, ticks []*ohlcdomain.OHLC) {
				assert.NoError(t, err)
				assert.Len(t, ticks, 1)
			},
		},
		{
			name: "success: no rows",
			mockFn: func(mock *mockOhlc.MockQuestDBClient, mockRows *mockOhlc.MockRowsInterface) {
				mock.EXPECT().Query(
					gomock.Any(),
					query+" AND symbol = $1 AND interval = $2 AND timestamp >= $3 AND timestamp <= $4 ORDER BY timestamp DESC LIMIT $5",
					[]interface{}{"BTCUSDT", "1m", now, now, 10},
				).Return(mockRows, nil)

				mockRows.EXPECT().Next().Return(false)
				mockRows.EXPECT().Err().Return(nil)
				mockRows.EXPECT().Close()
			},
			filter: ohlcdomain.OHLCFilter{
				Symbol:   "BTCUSDT",
				Interval: "1m",
				From:     &now,
				To:       &now,
				Limit:    10,
			},
			assertFn: func(t *testing.T, err error, ticks []*ohlcdomain.OHLC) {
				assert.NoError(t, err)
				assert.Len(t, ticks, 0)
			},
		},
		{
			name: "error: query fails",
			mockFn: func(mock *mockOhlc.MockQuestDBClient, mockRows *mockOhlc.MockRowsInterface) {
				mock.EXPECT().Query(
					gomock.Any(),
					query+" AND symbol = $1 AND interval = $2 AND timestamp >= $3 AND timestamp <= $4 ORDER BY timestamp DESC LIMIT $5",
					[]interface{}{"BTCUSDT", "1m", now, now, 10},
				).Return(nil, errors.New("query failed"))
			},
			filter: ohlcdomain.OHLCFilter{
				Symbol:   "BTCUSDT",
				Interval: "1m",
				From:     &now,
				To:       &now,
				Limit:    10,
			},
			assertFn: func(t *testing.T, err error, ticks []*ohlcdomain.OHLC) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "query failed")
			},
		},
		{
			name: "error: scan fails",
			mockFn: func(mock *mockOhlc.MockQuestDBClient, mockRows *mockOhlc.MockRowsInterface) {
				mock.EXPECT().Query(
					gomock.Any(),
					query+" AND symbol = $1 AND interval = $2 AND timestamp >= $3 AND timestamp <= $4 ORDER BY timestamp DESC LIMIT $5",
					[]interface{}{"BTCUSDT", "1m", now, now, 10},
				).Return(mockRows, nil)

				mockRows.EXPECT().Next().Return(true)
				mockRows.EXPECT().Scan(gomock.Any()).Return(errors.New("scan failed"))
				mockRows.EXPECT().Close()
			},
			filter: ohlcdomain.OHLCFilter{
				Symbol:   "BTCUSDT",
				Interval: "1m",
				From:     &now,
				To:       &now,
				Limit:    10,
			},
			assertFn: func(t *testing.T, err error, ticks []*ohlcdomain.OHLC) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "scan failed")
			},
		},
		{
			name: "error: rows.Err() fails",
			mockFn: func(mock *mockOhlc.MockQuestDBClient, mockRows *mockOhlc.MockRowsInterface) {
				mock.EXPECT().Query(
					gomock.Any(),
					query+" AND symbol = $1 AND interval = $2 AND timestamp >= $3 AND timestamp <= $4 ORDER BY timestamp DESC LIMIT $5",
					[]interface{}{"BTCUSDT", "1m", now, now, 10},
				).Return(mockRows, nil)
				mockRows.EXPECT().Next().Return(false) // No rows
				mockRows.EXPECT().Err().Return(errors.New("iteration error"))
				mockRows.EXPECT().Close() // Cleanup
			},
			filter: ohlcdomain.OHLCFilter{
				Symbol:   "BTCUSDT",
				Interval: "1m",
				From:     &now,
				To:       &now,
				Limit:    10,
			},
			assertFn: func(t *testing.T, err error, ticks []*ohlcdomain.OHLC) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "iteration error")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mockOhlc.NewMockQuestDBClient(ctrl)

			mockRows := mockOhlc.NewMockRowsInterface(ctrl)
			tc.mockFn(mockClient, mockRows)

			repo := NewRepository(mockClient)
			ticks, err := repo.GetByFilter(context.Background(), tc.filter)
			tc.assertFn(t, err, ticks)
		})
	}
}

func TestOhlc_GetLatest(t *testing.T) {
	query := `SELECT timestamp, symbol, interval, open, high, low, close, volume, trade_count
			  FROM ohlc 
			  WHERE symbol = $1 AND interval = $2 
			  ORDER BY timestamp DESC 
			  LIMIT 1`
	now := time.Now()
	testCases := []struct {
		name     string
		mockFn   func(mock *mockOhlc.MockQuestDBClient, mockRows *mockOhlc.MockRowsInterface)
		assertFn func(t *testing.T, err error, ohlc *ohlcdomain.OHLC)
		symbol   string
		interval string
	}{
		{
			name: "success",
			mockFn: func(mock *mockOhlc.MockQuestDBClient, mockRows *mockOhlc.MockRowsInterface) {
				mock.EXPECT().QueryRow(gomock.Any(), query, "BTCUSDT", "1m").Return(mockRows)
				mockRows.EXPECT().Scan(gomock.Any()).DoAndReturn(func(dest ...any) error {
					*dest[0].(*time.Time) = now
					*dest[1].(*string) = "BTCUSDT"
					*dest[2].(*string) = "1m"
					*dest[3].(*float64) = 10000
					*dest[4].(*float64) = 10500
					*dest[5].(*float64) = 9000
					*dest[6].(*float64) = 10200
					*dest[7].(*int64) = 100000
					*dest[8].(*int64) = 100
					return nil
				})
			},
			symbol:   "BTCUSDT",
			interval: "1m",
			assertFn: func(t *testing.T, err error, ohlc *ohlcdomain.OHLC) {
				assert.NoError(t, err)
				assert.NotNil(t, ohlc)
				assert.Equal(t, "BTCUSDT", ohlc.Symbol)
				assert.Equal(t, "1m", ohlc.Interval)
				assert.Equal(t, 10000.0, ohlc.Open)
				assert.Equal(t, 10500.0, ohlc.High)
				assert.Equal(t, 9000.0, ohlc.Low)
				assert.Equal(t, 10200.0, ohlc.Close)
				assert.Equal(t, int64(100000), ohlc.Volume)
				assert.Equal(t, int64(100), ohlc.TradeCount)
			},
		},
		{
			name: "no rows - returns nil",
			mockFn: func(mock *mockOhlc.MockQuestDBClient, mockRows *mockOhlc.MockRowsInterface) {
				mock.EXPECT().QueryRow(gomock.Any(), query, "BTCUSDT", "1m").Return(mockRows)
				mockRows.EXPECT().Scan(gomock.Any()).Return(pgx.ErrNoRows)
			},
			symbol:   "BTCUSDT",
			interval: "1m",
			assertFn: func(t *testing.T, err error, ohlc *ohlcdomain.OHLC) {
				assert.NoError(t, err)
				assert.Nil(t, ohlc)
			},
		},
		{
			name: "error - query fails",
			mockFn: func(mock *mockOhlc.MockQuestDBClient, mockRows *mockOhlc.MockRowsInterface) {
				mock.EXPECT().QueryRow(gomock.Any(), query, "BTCUSDT", "1m").Return(mockRows)
				mockRows.EXPECT().Scan(gomock.Any()).Return(errors.New("query failed"))
			},
			symbol:   "BTCUSDT",
			interval: "1m",
			assertFn: func(t *testing.T, err error, ohlc *ohlcdomain.OHLC) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to get latest OHLC")
				assert.Nil(t, ohlc)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mockOhlc.NewMockQuestDBClient(ctrl)
			mockRows := mockOhlc.NewMockRowsInterface(ctrl)

			tc.mockFn(mockClient, mockRows)

			repo := NewRepository(mockClient)
			ohlc, err := repo.GetLatest(context.Background(), tc.symbol, tc.interval)
			tc.assertFn(t, err, ohlc)
		})
	}
}

func TestOhlc_GetIntradayData(t *testing.T) {
	query := `SELECT timestamp, symbol, interval, open, high, low, close, volume, trade_count
			  FROM ohlc 
			  WHERE symbol = $1 AND interval = $2 
			  ORDER BY timestamp DESC 
			  LIMIT $3`
	now := time.Now()
	testCases := []struct {
		name     string
		mockFn   func(mock *mockOhlc.MockQuestDBClient, mockRows *mockOhlc.MockRowsInterface)
		assertFn func(t *testing.T, err error, ohlcs []*ohlcdomain.OHLC)
		symbol   string
		interval string
		limit    int
	}{
		{
			name: "success: with data",
			mockFn: func(mock *mockOhlc.MockQuestDBClient, mockRows *mockOhlc.MockRowsInterface) {
				mock.EXPECT().Query(gomock.Any(), query, "BTCUSDT", "1m", 10).Return(mockRows, nil)

				mockRows.EXPECT().Next().Return(true)
				mockRows.EXPECT().Scan(gomock.Any()).DoAndReturn(func(dest ...any) error {
					*dest[0].(*time.Time) = now
					*dest[1].(*string) = "BTCUSDT"
					*dest[2].(*string) = "1m"
					*dest[3].(*float64) = 10000
					*dest[4].(*float64) = 10500
					*dest[5].(*float64) = 9000
					*dest[6].(*float64) = 10200
					*dest[7].(*int64) = 100000
					*dest[8].(*int64) = 100
					return nil
				})
				mockRows.EXPECT().Next().Return(false)
				mockRows.EXPECT().Err().Return(nil)
				mockRows.EXPECT().Close()
			},
			symbol:   "BTCUSDT",
			interval: "1m",
			limit:    10,
			assertFn: func(t *testing.T, err error, ohlcs []*ohlcdomain.OHLC) {
				assert.NoError(t, err)
				assert.Len(t, ohlcs, 1)
				assert.Equal(t, "BTCUSDT", ohlcs[0].Symbol)
				assert.Equal(t, "1m", ohlcs[0].Interval)
			},
		},
		{
			name: "success: no rows",
			mockFn: func(mock *mockOhlc.MockQuestDBClient, mockRows *mockOhlc.MockRowsInterface) {
				mock.EXPECT().Query(gomock.Any(), query, "BTCUSDT", "1m", 10).Return(mockRows, nil)

				mockRows.EXPECT().Next().Return(false)
				mockRows.EXPECT().Err().Return(nil)
				mockRows.EXPECT().Close()
			},
			symbol:   "BTCUSDT",
			interval: "1m",
			limit:    10,
			assertFn: func(t *testing.T, err error, ohlcs []*ohlcdomain.OHLC) {
				assert.NoError(t, err)
				assert.Len(t, ohlcs, 0)
			},
		},
		{
			name: "error: query fails",
			mockFn: func(mock *mockOhlc.MockQuestDBClient, mockRows *mockOhlc.MockRowsInterface) {
				mock.EXPECT().Query(gomock.Any(), query, "BTCUSDT", "1m", 10).Return(nil, errors.New("query failed"))
			},
			symbol:   "BTCUSDT",
			interval: "1m",
			limit:    10,
			assertFn: func(t *testing.T, err error, ohlcs []*ohlcdomain.OHLC) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to query intraday OHLC")
				assert.Nil(t, ohlcs)
			},
		},
		{
			name: "error: scan fails",
			mockFn: func(mock *mockOhlc.MockQuestDBClient, mockRows *mockOhlc.MockRowsInterface) {
				mock.EXPECT().Query(gomock.Any(), query, "BTCUSDT", "1m", 10).Return(mockRows, nil)

				mockRows.EXPECT().Next().Return(true)
				mockRows.EXPECT().Scan(gomock.Any()).Return(errors.New("scan failed"))
				mockRows.EXPECT().Close()
			},
			symbol:   "BTCUSDT",
			interval: "1m",
			limit:    10,
			assertFn: func(t *testing.T, err error, ohlcs []*ohlcdomain.OHLC) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to scan OHLC")
				assert.Nil(t, ohlcs)
			},
		},
		{
			name: "error: rows.Err() fails",
			mockFn: func(mock *mockOhlc.MockQuestDBClient, mockRows *mockOhlc.MockRowsInterface) {
				mock.EXPECT().Query(gomock.Any(), query, "BTCUSDT", "1m", 10).Return(mockRows, nil)

				mockRows.EXPECT().Next().Return(false)
				mockRows.EXPECT().Err().Return(errors.New("iteration error"))
				mockRows.EXPECT().Close()
			},
			symbol:   "BTCUSDT",
			interval: "1m",
			limit:    10,
			assertFn: func(t *testing.T, err error, ohlcs []*ohlcdomain.OHLC) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "error iterating rows")
				assert.Nil(t, ohlcs)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mockOhlc.NewMockQuestDBClient(ctrl)
			mockRows := mockOhlc.NewMockRowsInterface(ctrl)

			tc.mockFn(mockClient, mockRows)

			repo := NewRepository(mockClient)
			ohlcs, err := repo.GetIntradayData(context.Background(), tc.symbol, tc.interval, tc.limit)
			tc.assertFn(t, err, ohlcs)
		})
	}
}
