package tick

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5"
	mock "github.com/muhammadchandra19/exchange/pkg/questdb/mock"
	"github.com/stretchr/testify/assert"
)

func TestTickRepository_Store(t *testing.T) {
	query := `INSERT INTO ticks (timestamp, symbol, price, volume, side) 
			  VALUES ($1, $2, $3, $4, $5)`
	testCases := []struct {
		name     string
		mockFn   func(tickData *Tick, mock *mock.MockQuestDBClient)
		assertFn func(t *testing.T, err error)
		tick     *Tick
	}{
		{
			name: "success",
			mockFn: func(tickData *Tick, mock *mock.MockQuestDBClient) {
				mock.EXPECT().Exec(gomock.Any(), query, tickData.Timestamp, tickData.Symbol, tickData.Price, tickData.Volume, tickData.Side).Return(nil)
			},
			tick: &Tick{
				Timestamp: time.Now(),
				Symbol:    "BTCUSDT",
				Price:     10000,
				Volume:    100,
				Side:      "buy",
			},
			assertFn: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "error",
			mockFn: func(tickData *Tick, mock *mock.MockQuestDBClient) {
				mock.EXPECT().Exec(gomock.Any(), query, tickData.Timestamp, tickData.Symbol, tickData.Price, tickData.Volume, tickData.Side).Return(errors.New("error"))
			},
			tick: &Tick{
				Timestamp: time.Now(),
				Symbol:    "BTCUSDT",
				Price:     10000,
				Volume:    100,
				Side:      "buy",
			},
			assertFn: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mock := mock.NewMockQuestDBClient(ctrl)
			tc.mockFn(tc.tick, mock)

			repo := NewRepository(mock)
			err := repo.Store(context.Background(), tc.tick)
			tc.assertFn(t, err)
		})
	}
}

func TestTickRepository_StoreBatch(t *testing.T) {
	testCases := []struct {
		name     string
		mockFn   func(ticks []*Tick, mock *mock.MockQuestDBClient)
		assertFn func(t *testing.T, err error)
		ticks    []*Tick
	}{
		{
			name: "success",
			mockFn: func(ticks []*Tick, mock *mock.MockQuestDBClient) {
				mock.EXPECT().CopyFrom(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(1), nil)
			},
			ticks: []*Tick{
				{
					Timestamp: time.Now(),
					Symbol:    "BTCUSDT",
					Price:     10000,
					Volume:    100,
					Side:      "buy",
				},
			},
			assertFn: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "error",
			mockFn: func(ticks []*Tick, mock *mock.MockQuestDBClient) {
				mock.EXPECT().CopyFrom(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(int64(0), errors.New("error"))
			},
			ticks: []*Tick{
				{
					Timestamp: time.Now(),
					Symbol:    "BTCUSDT",
					Price:     10000,
					Volume:    100,
					Side:      "buy",
				},
			},
			assertFn: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mock := mock.NewMockQuestDBClient(ctrl)
			tc.mockFn(tc.ticks, mock)

			repo := NewRepository(mock)
			err := repo.StoreBatch(context.Background(), tc.ticks)
			tc.assertFn(t, err)
		})
	}
}

func TestTickRepository_GetByFilter(t *testing.T) {
	now := time.Now()
	query := "SELECT timestamp, symbol, price, volume, side FROM ticks WHERE 1=1"
	testCases := []struct {
		name     string
		mockFn   func(mock *mock.MockQuestDBClient, mockRows *mock.MockRowsInterface)
		assertFn func(t *testing.T, err error, ticks []*Tick)
		filter   Filter
	}{
		{
			name: "success: with all filters",
			mockFn: func(mock *mock.MockQuestDBClient, mockRows *mock.MockRowsInterface) {
				mock.EXPECT().Query(
					gomock.Any(),
					query+" AND symbol = $1 AND timestamp >= $2 AND timestamp <= $3 ORDER BY timestamp DESC LIMIT $4 OFFSET $5",
					[]interface{}{"BTCUSDT", now, now, 10, 1},
				).Return(mockRows, nil)

				mockRows.EXPECT().Next().Return(true)
				mockRows.EXPECT().Scan(gomock.Any()).DoAndReturn(func(dest ...any) error {
					*dest[0].(*time.Time) = now
					*dest[1].(*string) = "BTCUSDT"
					*dest[2].(*float64) = 50000.0
					*dest[3].(*int64) = 100
					*dest[4].(*string) = "buy"
					return nil
				})
				mockRows.EXPECT().Next().Return(false)
				mockRows.EXPECT().Err().Return(nil)
				mockRows.EXPECT().Close()
			},
			filter: Filter{Symbol: "BTCUSDT", From: &now, To: &now, Limit: 10, Offset: 1},
			assertFn: func(t *testing.T, err error, ticks []*Tick) {
				assert.NoError(t, err)
				assert.Len(t, ticks, 1)
			},
		},
		{
			name: "success - single row",
			mockFn: func(mock *mock.MockQuestDBClient, mockRows *mock.MockRowsInterface) {
				mock.EXPECT().Query(gomock.Any(), query+" AND symbol = $1 ORDER BY timestamp DESC", gomock.Any()).Return(mockRows, nil)

				// Pattern: One row of data
				mockRows.EXPECT().Next().Return(true) // First call: has data
				mockRows.EXPECT().Scan(gomock.Any()).DoAndReturn(func(dest ...any) error {
					// Fill mock data
					*dest[0].(*time.Time) = time.Now()
					*dest[1].(*string) = "BTCUSDT"
					*dest[2].(*float64) = 50000.0
					*dest[3].(*int64) = 100
					*dest[4].(*string) = "buy"
					return nil
				})
				mockRows.EXPECT().Next().Return(false) // Second call: no more data
				mockRows.EXPECT().Err().Return(nil)    // Check for errors
				mockRows.EXPECT().Close()              // Cleanup
			},
			filter: Filter{Symbol: "BTCUSDT"},
			assertFn: func(t *testing.T, err error, ticks []*Tick) {
				assert.NoError(t, err)
				assert.Len(t, ticks, 1)
			},
		},
		{
			name: "success - no rows",
			mockFn: func(mock *mock.MockQuestDBClient, mockRows *mock.MockRowsInterface) {
				mock.EXPECT().Query(gomock.Any(), query+" AND symbol = $1 ORDER BY timestamp DESC", gomock.Any()).Return(mockRows, nil)

				// Pattern: No rows
				mockRows.EXPECT().Next().Return(false) // First call: no data
				mockRows.EXPECT().Err().Return(nil)    // Check for errors
				mockRows.EXPECT().Close()              // Cleanup
			},
			filter: Filter{Symbol: "NONE"},
			assertFn: func(t *testing.T, err error, ticks []*Tick) {
				assert.NoError(t, err)
				assert.Len(t, ticks, 0)
			},
		},
		{
			name: "error - query fails",
			mockFn: func(mock *mock.MockQuestDBClient, mockRows *mock.MockRowsInterface) {
				// Pattern: Query fails immediately
				mock.EXPECT().Query(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("query failed"))
				// No rows expectations needed - query failed
			},
			filter: Filter{Symbol: "BTCUSDT"},
			assertFn: func(t *testing.T, err error, ticks []*Tick) {
				assert.Error(t, err)
				assert.Nil(t, ticks)
			},
		},
		{
			name: "error - scan fails",
			mockFn: func(mock *mock.MockQuestDBClient, mockRows *mock.MockRowsInterface) {
				mock.EXPECT().Query(gomock.Any(), query+" AND symbol = $1 ORDER BY timestamp DESC", gomock.Any()).Return(mockRows, nil)

				// Pattern: Row exists but scan fails
				mockRows.EXPECT().Next().Return(true) // Has data
				mockRows.EXPECT().Scan(gomock.Any()).Return(errors.New("scan failed"))
				mockRows.EXPECT().Close() // Cleanup (defer always runs)
			},
			filter: Filter{Symbol: "BTCUSDT"},
			assertFn: func(t *testing.T, err error, ticks []*Tick) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "scan failed")
			},
		},
		{
			name: "error - rows.Err() fails",
			mockFn: func(mock *mock.MockQuestDBClient, mockRows *mock.MockRowsInterface) {
				mock.EXPECT().Query(gomock.Any(), query+" AND symbol = $1 ORDER BY timestamp DESC", gomock.Any()).Return(mockRows, nil)

				// Pattern: Iteration completes but has error
				mockRows.EXPECT().Next().Return(false) // No rows
				mockRows.EXPECT().Err().Return(errors.New("iteration error"))
				mockRows.EXPECT().Close() // Cleanup
			},
			filter: Filter{Symbol: "BTCUSDT"},
			assertFn: func(t *testing.T, err error, ticks []*Tick) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "iteration error")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mock.NewMockQuestDBClient(ctrl)
			mockRows := mock.NewMockRowsInterface(ctrl)

			tc.mockFn(mockClient, mockRows)

			repo := NewRepository(mockClient)
			ticks, err := repo.GetByFilter(context.Background(), tc.filter)
			tc.assertFn(t, err, ticks)
		})
	}
}

func TestTickRepository_GetLatestBySymbol(t *testing.T) {
	query := `SELECT timestamp, symbol, price, volume, side 
			  FROM ticks 
			  WHERE symbol = $1 
			  ORDER BY timestamp DESC 
			  LIMIT 1`
	testCases := []struct {
		name     string
		mockFn   func(mock *mock.MockQuestDBClient, mockRows *mock.MockRowsInterface)
		assertFn func(t *testing.T, err error, tick *Tick)
		symbol   string
	}{
		{
			name: "success",
			mockFn: func(mock *mock.MockQuestDBClient, mockRows *mock.MockRowsInterface) {
				mock.EXPECT().QueryRow(gomock.Any(), query, "BTCUSDT").Return(mockRows)
				mockRows.EXPECT().Scan(gomock.Any()).DoAndReturn(func(dest ...any) error {
					*dest[0].(*time.Time) = time.Now()
					*dest[1].(*string) = "BTCUSDT"
					*dest[2].(*float64) = 50000.0
					*dest[3].(*int64) = 100
					*dest[4].(*string) = "buy"
					return nil
				})
			},
			symbol: "BTCUSDT",
			assertFn: func(t *testing.T, err error, tick *Tick) {
				assert.NoError(t, err)
				assert.Equal(t, "BTCUSDT", tick.Symbol)
				assert.Equal(t, 50000.0, tick.Price)
				assert.Equal(t, int64(100), tick.Volume)
				assert.Equal(t, "buy", tick.Side)
			},
		},
		{
			name: "error - no rows",
			mockFn: func(mock *mock.MockQuestDBClient, mockRows *mock.MockRowsInterface) {
				mock.EXPECT().QueryRow(gomock.Any(), query, "BTCUSDT").Return(mockRows)
				mockRows.EXPECT().Scan(gomock.Any()).DoAndReturn(func(dest ...any) error {
					return pgx.ErrNoRows
				})
			},
			symbol: "BTCUSDT",
			assertFn: func(t *testing.T, err error, tick *Tick) {
				assert.NoError(t, err)
				assert.Nil(t, tick)
			},
		},
		{
			name: "error - query fails",
			mockFn: func(mock *mock.MockQuestDBClient, mockRows *mock.MockRowsInterface) {
				mock.EXPECT().QueryRow(gomock.Any(), query, "BTCUSDT").Return(mockRows)
				mockRows.EXPECT().Scan(gomock.Any()).DoAndReturn(func(dest ...any) error {
					return errors.New("query failed")
				})
			},
			symbol: "BTCUSDT",
			assertFn: func(t *testing.T, err error, tick *Tick) {
				assert.Error(t, err)
				assert.Nil(t, tick)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mock.NewMockQuestDBClient(ctrl)
			mockRows := mock.NewMockRowsInterface(ctrl)

			tc.mockFn(mockClient, mockRows)

			repo := NewRepository(mockClient)
			tick, err := repo.GetLatestBySymbol(context.Background(), tc.symbol)
			tc.assertFn(t, err, tick)
		})
	}
}

func TestTickRepository_GetVolumeBySymbol(t *testing.T) {
	now := time.Now()
	query := `SELECT COALESCE(SUM(volume), 0) FROM ticks 
			  WHERE symbol = $1 AND timestamp >= $2 AND timestamp <= $3`
	testCases := []struct {
		name     string
		mockFn   func(mock *mock.MockQuestDBClient, mockRows *mock.MockRowsInterface)
		assertFn func(t *testing.T, err error, volume int64)
		symbol   string
		from     time.Time
		to       time.Time
	}{
		{
			name: "success",
			mockFn: func(mock *mock.MockQuestDBClient, mockRows *mock.MockRowsInterface) {
				mock.EXPECT().QueryRow(gomock.Any(), query, "BTCUSDT", now, now).Return(mockRows)
				mockRows.EXPECT().Scan(gomock.Any()).DoAndReturn(func(dest ...any) error {
					*dest[0].(*int64) = 100
					return nil
				})
			},
			symbol: "BTCUSDT",
			from:   now,
			to:     now,
			assertFn: func(t *testing.T, err error, volume int64) {
				assert.NoError(t, err)
				assert.Equal(t, int64(100), volume)
			},
		},
		{
			name: "error - query fails",
			mockFn: func(mock *mock.MockQuestDBClient, mockRows *mock.MockRowsInterface) {
				mock.EXPECT().QueryRow(gomock.Any(), query, "BTCUSDT", now, now).Return(mockRows)
				mockRows.EXPECT().Scan(gomock.Any()).DoAndReturn(func(dest ...any) error {
					return errors.New("query failed")
				})
			},
			symbol: "BTCUSDT",
			from:   now,
			to:     now,
			assertFn: func(t *testing.T, err error, volume int64) {
				assert.Error(t, err)
				assert.Equal(t, int64(0), volume)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mock.NewMockQuestDBClient(ctrl)
			mockRows := mock.NewMockRowsInterface(ctrl)

			tc.mockFn(mockClient, mockRows)

			repo := NewRepository(mockClient)
			volume, err := repo.GetVolumeBySymbol(context.Background(), tc.symbol, tc.from, tc.to)
			tc.assertFn(t, err, volume)
		})
	}
}
