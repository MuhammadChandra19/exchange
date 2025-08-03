package order

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/muhammadchandra19/exchange/pkg/logger"
	mockLogger "github.com/muhammadchandra19/exchange/pkg/logger/mock"
	mockPg "github.com/muhammadchandra19/exchange/pkg/postgresql/mock"
	"github.com/stretchr/testify/assert"
)

func TestOrder_Store(t *testing.T) {
	ctx := context.Background()
	query := `INSERT INTO orders (id, user_id, symbol, side, price, quantity, type, status, timestamp) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	now := time.Now()
	testCases := []struct {
		name     string
		mockFn   func(mockpg *mockPg.MockPostgreSQLClient, mockLogger *mockLogger.MockInterface, tc *Order)
		testData *Order
		assertFn func(t *testing.T, err error)
	}{
		{
			name: "success",
			mockFn: func(mockpg *mockPg.MockPostgreSQLClient, mockLogger *mockLogger.MockInterface, tc *Order) {
				mockpg.EXPECT().
					Exec(ctx, query,
						tc.ID,
						tc.UserID,
						tc.Symbol,
						tc.Side,
						tc.Price,
						tc.Quantity,
						tc.Type,
						tc.Status,
						tc.Timestamp,
					).Return(pgconn.CommandTag{}, nil)

				mockLogger.EXPECT().
					Info("Inserted order", logger.Field{
						Key:   "commandTag",
						Value: "",
					})
			},
			testData: &Order{
				ID:        "1",
				UserID:    "1",
				Symbol:    "BTCUSDT",
				Side:      "BUY",
				Price:     10000,
				Quantity:  1,
				Type:      "LIMIT",
				Status:    "PENDING",
				Timestamp: now,
			},
			assertFn: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "error",
			mockFn: func(mockpg *mockPg.MockPostgreSQLClient, mockLogger *mockLogger.MockInterface, tc *Order) {
				mockpg.EXPECT().
					Exec(ctx, query,
						tc.ID,
						tc.UserID,
						tc.Symbol,
						tc.Side,
						tc.Price,
						tc.Quantity,
						tc.Type,
						tc.Status,
						tc.Timestamp,
					).Return(pgconn.CommandTag{}, errors.New("error"))

				mockLogger.EXPECT().
					Error(errors.New("error"), logger.Field{
						Key:   "error",
						Value: "error",
					})
			},
			testData: &Order{
				ID:        "1",
				UserID:    "1",
				Symbol:    "BTCUSDT",
				Side:      "BUY",
				Price:     10000,
				Quantity:  1,
				Type:      "LIMIT",
				Status:    "PENDING",
				Timestamp: now,
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

			pg := mockPg.NewMockPostgreSQLClient(ctrl)
			log := mockLogger.NewMockInterface(ctrl)

			repo := NewRepository(pg, log)

			tc.mockFn(pg, log, tc.testData)

			err := repo.Store(ctx, tc.testData)
			tc.assertFn(t, err)
		})
	}
}

func TestOrder_StoreBatch(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	testCases := []struct {
		name     string
		mockFn   func(mockpg *mockPg.MockPostgreSQLClient, mockLogger *mockLogger.MockInterface, tc []*Order)
		testData []*Order
		assertFn func(t *testing.T, err error)
	}{
		{
			name: "success",
			mockFn: func(mockpg *mockPg.MockPostgreSQLClient, mockLogger *mockLogger.MockInterface, tc []*Order) {
				mockpg.EXPECT().
					CopyFrom(ctx,
						pgx.Identifier{"orders"},
						[]string{"id", "user_id", "symbol", "side", "price", "quantity", "type", "status", "timestamp"},
						gomock.Any(),
					).Return(int64(2), nil)

				mockLogger.EXPECT().
					Info("Inserted batch of orders", logger.Field{
						Key:   "copyCount",
						Value: int64(2),
					})
			},
			testData: []*Order{
				{
					ID:        "1",
					UserID:    "1",
					Symbol:    "BTCUSDT",
					Side:      "BUY",
					Price:     10000,
					Quantity:  1,
					Type:      "LIMIT",
					Status:    "PENDING",
					Timestamp: now,
				},
				{
					ID:        "2",
					UserID:    "1",
					Symbol:    "BTCUSDT",
					Side:      "SELL",
					Price:     10000,
					Quantity:  1,
					Type:      "LIMIT",
					Status:    "PENDING",
					Timestamp: now,
				},
			},
			assertFn: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "error",
			mockFn: func(mockpg *mockPg.MockPostgreSQLClient, mockLogger *mockLogger.MockInterface, tc []*Order) {
				mockpg.EXPECT().
					CopyFrom(ctx,
						pgx.Identifier{"orders"},
						[]string{"id", "user_id", "symbol", "side", "price", "quantity", "type", "status", "timestamp"},
						gomock.Any(),
					).Return(int64(0), errors.New("error"))

				mockLogger.EXPECT().
					Error(errors.New("error"), logger.Field{
						Key:   "error",
						Value: "error",
					})
			},
			testData: []*Order{
				{
					ID:        "1",
					UserID:    "1",
					Symbol:    "BTCUSDT",
					Side:      "BUY",
					Price:     10000,
					Quantity:  1,
					Type:      "LIMIT",
					Status:    "PENDING",
					Timestamp: now,
				},
				{
					ID:        "2",
					UserID:    "1",
					Symbol:    "BTCUSDT",
					Side:      "SELL",
					Price:     10000,
					Quantity:  1,
					Type:      "LIMIT",
					Status:    "PENDING",
					Timestamp: now,
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

			pg := mockPg.NewMockPostgreSQLClient(ctrl)
			log := mockLogger.NewMockInterface(ctrl)

			repo := NewRepository(pg, log)

			tc.mockFn(pg, log, tc.testData)

			err := repo.StoreBatch(ctx, tc.testData)
			tc.assertFn(t, err)
		})
	}
}

func TestOrder_Update(t *testing.T) {
	ctx := context.Background()
	query := `UPDATE orders SET user_id = $1, symbol = $2, side = $3, price = $4, quantity = $5, type = $6, status = $7, timestamp = $8 WHERE id = $9`
	now := time.Now()
	testCases := []struct {
		name     string
		mockFn   func(mockpg *mockPg.MockPostgreSQLClient, mockLogger *mockLogger.MockInterface, tc *Order)
		testData *Order
		assertFn func(t *testing.T, err error)
	}{
		{
			name: "success",
			mockFn: func(mockpg *mockPg.MockPostgreSQLClient, mockLogger *mockLogger.MockInterface, tc *Order) {
				mockpg.EXPECT().
					Exec(ctx, query,
						tc.UserID,
						tc.Symbol,
						tc.Side,
						tc.Price,
						tc.Quantity,
						tc.Type,
						tc.Status,
						tc.Timestamp,
						tc.ID,
					).Return(pgconn.CommandTag{}, nil)

				mockLogger.EXPECT().
					Info("Updated order", logger.Field{
						Key:   "commandTag",
						Value: "",
					})
			},
			testData: &Order{
				ID:        "1",
				UserID:    "1",
				Symbol:    "BTCUSDT",
				Side:      "BUY",
				Price:     10000,
				Quantity:  1,
				Type:      "LIMIT",
				Status:    "PENDING",
				Timestamp: now,
			},
			assertFn: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "error",
			mockFn: func(mockpg *mockPg.MockPostgreSQLClient, mockLogger *mockLogger.MockInterface, tc *Order) {
				mockpg.EXPECT().
					Exec(ctx, query,
						tc.UserID,
						tc.Symbol,
						tc.Side,
						tc.Price,
						tc.Quantity,
						tc.Type,
						tc.Status,
						tc.Timestamp,
						tc.ID,
					).Return(pgconn.CommandTag{}, errors.New("error"))

				mockLogger.EXPECT().
					Error(errors.New("error"), logger.Field{
						Key:   "error",
						Value: "error",
					})
			},
			testData: &Order{
				ID:        "1",
				UserID:    "1",
				Symbol:    "BTCUSDT",
				Side:      "BUY",
				Price:     10000,
				Quantity:  1,
				Type:      "LIMIT",
				Status:    "PENDING",
				Timestamp: now,
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

			pg := mockPg.NewMockPostgreSQLClient(ctrl)
			log := mockLogger.NewMockInterface(ctrl)

			repo := NewRepository(pg, log)

			tc.mockFn(pg, log, tc.testData)

			err := repo.Update(ctx, tc.testData)
			tc.assertFn(t, err)
		})
	}
}

func TestOrder_GetByID(t *testing.T) {
	ctx := context.Background()
	query := `SELECT id, user_id, symbol, side, price, quantity, type, status, timestamp FROM orders WHERE id = $1`
	now := time.Now()
	testCases := []struct {
		name     string
		mockFn   func(mockpg *mockPg.MockPostgreSQLClient, mockLogger *mockLogger.MockInterface, mockRow *mockPg.MockRowInterface, tc *Order)
		testData *Order
		assertFn func(t *testing.T, err error, tc *Order, order *Order)
	}{
		{
			name: "success",
			mockFn: func(mockpg *mockPg.MockPostgreSQLClient, mockLogger *mockLogger.MockInterface, mockRow *mockPg.MockRowInterface, tc *Order) {
				mockpg.EXPECT().
					QueryRow(ctx, query, tc.ID).
					Return(mockRow)

				mockRow.EXPECT().
					Scan(gomock.Any()).DoAndReturn(func(dest ...any) error {
					*dest[0].(*string) = tc.ID
					*dest[1].(*string) = tc.UserID
					*dest[2].(*string) = tc.Symbol
					*dest[3].(*string) = tc.Side
					*dest[4].(*float64) = tc.Price
					*dest[5].(*int64) = tc.Quantity
					*dest[6].(*string) = tc.Type
					*dest[7].(*Status) = Status(tc.Status)
					*dest[8].(*time.Time) = tc.Timestamp
					return nil
				})
			},
			testData: &Order{
				ID:        "1",
				UserID:    "1",
				Symbol:    "BTCUSDT",
				Side:      "BUY",
				Price:     10000,
				Quantity:  1,
				Type:      "LIMIT",
				Status:    OrderStatusPlaced,
				Timestamp: now,
			},
			assertFn: func(t *testing.T, err error, tc *Order, order *Order) {
				assert.NoError(t, err)
				assert.Equal(t, tc, order)
			},
		},
		{
			name: "error: no rows",
			mockFn: func(mockpg *mockPg.MockPostgreSQLClient, mockLogger *mockLogger.MockInterface, mockRow *mockPg.MockRowInterface, tc *Order) {
				mockpg.EXPECT().
					QueryRow(ctx, query, tc.ID).
					Return(mockRow)

				mockRow.EXPECT().
					Scan(gomock.Any()).Return(pgx.ErrNoRows)
			},
			testData: &Order{
				ID:        "1",
				UserID:    "1",
				Symbol:    "BTCUSDT",
				Side:      "BUY",
				Price:     10000,
				Quantity:  1,
				Type:      "LIMIT",
				Status:    OrderStatusPlaced,
				Timestamp: now,
			},
			assertFn: func(t *testing.T, err error, tc *Order, order *Order) {
				assert.NoError(t, err)
				assert.Nil(t, order)
			},
		},
		{
			name: "error: query fails",
			mockFn: func(mockpg *mockPg.MockPostgreSQLClient, mockLogger *mockLogger.MockInterface, mockRow *mockPg.MockRowInterface, tc *Order) {
				mockpg.EXPECT().
					QueryRow(ctx, query, tc.ID).
					Return(mockRow)

				mockRow.EXPECT().
					Scan(gomock.Any()).Return(errors.New("error"))
			},
			testData: &Order{
				ID:        "1",
				UserID:    "1",
				Symbol:    "BTCUSDT",
				Side:      "BUY",
				Price:     10000,
				Quantity:  1,
				Type:      "LIMIT",
				Status:    OrderStatusPlaced,
				Timestamp: now,
			},
			assertFn: func(t *testing.T, err error, tc *Order, order *Order) {
				assert.Error(t, err)
				assert.Nil(t, order)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			pg := mockPg.NewMockPostgreSQLClient(ctrl)
			row := mockPg.NewMockRowInterface(ctrl)
			log := mockLogger.NewMockInterface(ctrl)

			repo := NewRepository(pg, log)

			tc.mockFn(pg, log, row, tc.testData)

			order, err := repo.GetByID(ctx, tc.testData.ID)
			tc.assertFn(t, err, tc.testData, order)
		})
	}
}

func TestOrder_Delete(t *testing.T) {
	ctx := context.Background()
	query := `DELETE FROM orders WHERE id = $1`
	testCases := []struct {
		name   string
		mockFn func(mockpg *mockPg.MockPostgreSQLClient, mockLogger *mockLogger.MockInterface)

		assertFn func(t *testing.T, err error)
	}{
		{
			name: "success",
			mockFn: func(mockpg *mockPg.MockPostgreSQLClient, mockLogger *mockLogger.MockInterface) {
				mockpg.EXPECT().
					Exec(ctx, query, "1").Return(pgconn.CommandTag{}, nil)
			},
			assertFn: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "error",
			mockFn: func(mockpg *mockPg.MockPostgreSQLClient, mockLogger *mockLogger.MockInterface) {
				mockpg.EXPECT().
					Exec(ctx, query, "1").Return(pgconn.CommandTag{}, errors.New("error"))

				mockLogger.EXPECT().
					Error(errors.New("error"), logger.Field{
						Key:   "error",
						Value: "error",
					})
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

			pg := mockPg.NewMockPostgreSQLClient(ctrl)
			log := mockLogger.NewMockInterface(ctrl)

			repo := NewRepository(pg, log)

			tc.mockFn(pg, log)

			err := repo.Delete(ctx, "1")
			tc.assertFn(t, err)
		})
	}
}

func TestOrder_List(t *testing.T) {
	ctx := context.Background()
	query := "SELECT id, user_id, symbol, side, price, quantity, type, status, timestamp FROM orders WHERE 1=1"
	now := time.Now()
	testCases := []struct {
		name     string
		mockFn   func(mockpg *mockPg.MockPostgreSQLClient, mockLogger *mockLogger.MockInterface, mockRows *mockPg.MockRowsInterface, tc Filter)
		filter   Filter
		assertFn func(t *testing.T, err error, orders []*Order)
	}{
		{
			name: "success",
			filter: Filter{
				UserID:        "1",
				Symbol:        "BTCUSDT",
				Side:          "BUY",
				Status:        "modified",
				From:          &now,
				To:            &now,
				Limit:         20,
				Offset:        10,
				SortDirection: "ASC",
			},
			mockFn: func(mockpg *mockPg.MockPostgreSQLClient, mockLogger *mockLogger.MockInterface, mockRows *mockPg.MockRowsInterface, tc Filter) {
				mockpg.EXPECT().
					Query(
						ctx,
						query+" AND user_id = $1 AND symbol = $2 AND side = $3 AND status = $4 AND timestamp >= $5 AND timestamp <= $6 ORDER BY timestamp ASC LIMIT $7 OFFSET $8",
						tc.UserID,
						tc.Symbol,
						tc.Side,
						tc.Status,
						now,
						now,
						tc.Limit,
						tc.Offset,
					).Return(mockRows, nil)

				mockRows.EXPECT().Next().Return(true)
				mockRows.EXPECT().
					Scan(gomock.Any()).DoAndReturn(func(dest ...any) error {
					*dest[0].(*string) = "1"
					*dest[1].(*string) = "1"
					*dest[2].(*string) = "BTCUSDT"
					*dest[3].(*string) = "BUY"
					*dest[4].(*float64) = 10000
					*dest[5].(*int64) = 1
					*dest[6].(*string) = "LIMIT"
					*dest[7].(*Status) = OrderStatusModified
					*dest[8].(*time.Time) = now
					return nil
				})

				mockRows.EXPECT().Next().Return(false)
				mockRows.EXPECT().Close()
			},
			assertFn: func(t *testing.T, err error, orders []*Order) {
				assert.NoError(t, err)
				assert.Equal(t, 1, len(orders))
			},
		},
		{
			name: "failed to query",
			filter: Filter{
				UserID: "1",
				Symbol: "BTCUSDT",
				Side:   "BUY",
				Status: "modified",
				From:   &now,
				To:     &now,
				Limit:  20,
				Offset: 10,
			},
			mockFn: func(mockpg *mockPg.MockPostgreSQLClient, mockLogger *mockLogger.MockInterface, mockRows *mockPg.MockRowsInterface, tc Filter) {
				mockpg.EXPECT().
					Query(
						ctx,
						query+" AND user_id = $1 AND symbol = $2 AND side = $3 AND status = $4 AND timestamp >= $5 AND timestamp <= $6 ORDER BY timestamp DESC LIMIT $7 OFFSET $8",
						tc.UserID,
						tc.Symbol,
						tc.Side,
						tc.Status,
						now,
						now,
						tc.Limit,
						tc.Offset,
					).Return(mockRows, errors.New("error"))

				mockLogger.EXPECT().
					Error(errors.New("error"), logger.Field{
						Key:   "error query",
						Value: "error",
					})
			},
			assertFn: func(t *testing.T, err error, orders []*Order) {
				assert.Error(t, err)
				assert.Nil(t, orders)
			},
		},
		{
			name: "failed to scan",
			filter: Filter{
				UserID:        "1",
				Symbol:        "BTCUSDT",
				Side:          "BUY",
				Status:        "modified",
				From:          &now,
				To:            &now,
				Limit:         20,
				Offset:        10,
				SortDirection: "ASC",
			},
			mockFn: func(mockpg *mockPg.MockPostgreSQLClient, mockLogger *mockLogger.MockInterface, mockRows *mockPg.MockRowsInterface, tc Filter) {
				mockpg.EXPECT().
					Query(
						ctx,
						query+" AND user_id = $1 AND symbol = $2 AND side = $3 AND status = $4 AND timestamp >= $5 AND timestamp <= $6 ORDER BY timestamp ASC LIMIT $7 OFFSET $8",
						tc.UserID,
						tc.Symbol,
						tc.Side,
						tc.Status,
						now,
						now,
						tc.Limit,
						tc.Offset,
					).Return(mockRows, nil)

				mockRows.EXPECT().Next().Return(true)
				mockRows.EXPECT().
					Scan(gomock.Any()).Return(errors.New("error"))
				mockRows.EXPECT().Close()

				mockLogger.EXPECT().
					Error(errors.New("error"), logger.Field{
						Key:   "error scan",
						Value: "error",
					})
			},
			assertFn: func(t *testing.T, err error, orders []*Order) {
				assert.Error(t, err)
				assert.Nil(t, orders)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			pg := mockPg.NewMockPostgreSQLClient(ctrl)
			rows := mockPg.NewMockRowsInterface(ctrl)
			log := mockLogger.NewMockInterface(ctrl)

			repo := NewRepository(pg, log)

			tc.mockFn(pg, log, rows, tc.filter)

			orders, err := repo.List(ctx, tc.filter)
			tc.assertFn(t, err, orders)
		})
	}
}
