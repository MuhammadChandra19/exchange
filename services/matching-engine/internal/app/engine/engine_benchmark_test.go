package engine

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/muhammadchandra19/exchange/pkg/logger"
	matchpublisherv1_mock "github.com/muhammadchandra19/exchange/services/matching-engine/internal/domain/match-publisher/v1/mock"
	orderreadermock "github.com/muhammadchandra19/exchange/services/matching-engine/internal/domain/order-reader/v1/mock"
	orderbookv1 "github.com/muhammadchandra19/exchange/services/matching-engine/internal/domain/orderbook/v1"
	snapshotmock "github.com/muhammadchandra19/exchange/services/matching-engine/internal/domain/snapshot/v1/mock"
	"github.com/muhammadchandra19/exchange/services/matching-engine/internal/usecase/orderbook"
	"github.com/muhammadchandra19/exchange/services/matching-engine/pkg/config"
)

// Benchmark test cases structure
type benchmarkTestCase struct {
	name        string
	setupEngine func(*testing.B) *Engine
	setupData   func(*Engine, *testing.B)
	operation   func(*Engine, int)
	cleanup     func(*Engine)
}

func setupBenchmarkEngine(b *testing.B) *Engine {
	ctrl := gomock.NewController(b)

	mockOrderReader := orderreadermock.NewMockOrderReader(ctrl)
	mockSnapshotStore := snapshotmock.NewMockStore(ctrl)
	mockMatchPublisher := matchpublisherv1_mock.NewMockMatchPublisher(ctrl)

	ob := orderbook.NewOrderbook()
	log, err := logger.NewLogger()
	if err != nil {
		b.Fatal(err)
	}

	cfg := &config.Config{
		Pair: "BTC-USD",
	}

	// Setup basic expectations
	mockSnapshotStore.EXPECT().
		LoadStore(gomock.Any()).
		Return(nil, nil).
		Times(1)

	// Setup match publisher expectations for when matches occur
	mockMatchPublisher.EXPECT().
		PublishMatchEvent(gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()

	engine := NewEngine(ob, mockOrderReader, mockSnapshotStore, mockMatchPublisher, log, cfg)

	// Initialize context to avoid nil pointer dereference
	engine.ctx = context.Background()

	return engine
}

func BenchmarkEngine_ProcessLimitOrder(b *testing.B) {
	testCases := []benchmarkTestCase{
		{
			name:        "single_threaded_limit_orders",
			setupEngine: setupBenchmarkEngine,
			setupData:   func(e *Engine, b *testing.B) {},
			operation: func(e *Engine, i int) {
				orderRequest := createTestOrderRequest(
					"user",
					orderbookv1.OrderTypeLimit,
					i%2 == 0, // Alternate between bid and ask
					10.0,
					50000.0+float64(i%100), // Vary price slightly
					int64(i),
				)
				_ = e.processOrder(&orderRequest)
			},
			cleanup: func(e *Engine) {},
		},
		{
			name:        "parallel_limit_orders",
			setupEngine: setupBenchmarkEngine,
			setupData:   func(e *Engine, b *testing.B) {},
			operation: func(e *Engine, i int) {
				orderRequest := createTestOrderRequest(
					"user",
					orderbookv1.OrderTypeLimit,
					i%2 == 0,
					10.0,
					50000.0+float64(i%100),
					int64(i),
				)
				_ = e.processOrder(&orderRequest)
			},
			cleanup: func(e *Engine) {},
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			engine := tc.setupEngine(b)
			tc.setupData(engine, b)

			b.ResetTimer()

			if tc.name == "parallel_limit_orders" {
				b.RunParallel(func(pb *testing.PB) {
					i := 0
					for pb.Next() {
						tc.operation(engine, i)
						i++
					}
				})
			} else {
				for i := 0; i < b.N; i++ {
					tc.operation(engine, i)
				}
			}

			b.StopTimer()
			tc.cleanup(engine)
		})
	}
}

func BenchmarkEngine_ProcessMarketOrder(b *testing.B) {
	testCases := []benchmarkTestCase{
		{
			name:        "market_orders_with_liquidity",
			setupEngine: setupBenchmarkEngine,
			setupData: func(e *Engine, b *testing.B) {
				// Pre-populate orderbook with limit orders for liquidity
				for i := 0; i < 1000; i++ {
					sellOrder := createTestOrderRequest(
						"seller",
						orderbookv1.OrderTypeLimit,
						false,
						10.0,
						50000.0+float64(i),
						int64(i),
					)
					_ = e.processOrder(&sellOrder)

					buyOrder := createTestOrderRequest(
						"buyer",
						orderbookv1.OrderTypeLimit,
						true,
						10.0,
						49000.0-float64(i),
						int64(i+1000),
					)
					_ = e.processOrder(&buyOrder)
				}
			},
			operation: func(e *Engine, i int) {
				orderRequest := createTestOrderRequest(
					"market_user",
					orderbookv1.OrderTypeMarket,
					i%2 == 0, // Alternate between market buy and sell
					5.0,
					0.0, // Market orders don't have price
					int64(i+2000),
				)
				_ = e.processOrder(&orderRequest)
			},
			cleanup: func(e *Engine) {},
		},
		{
			name:        "parallel_market_orders",
			setupEngine: setupBenchmarkEngine,
			setupData: func(e *Engine, b *testing.B) {
				// Pre-populate with fewer orders for parallel test
				for i := 0; i < 100; i++ {
					sellOrder := createTestOrderRequest(
						"seller",
						orderbookv1.OrderTypeLimit,
						false,
						10.0,
						50000.0+float64(i*10),
						int64(i),
					)
					_ = e.processOrder(&sellOrder)

					buyOrder := createTestOrderRequest(
						"buyer",
						orderbookv1.OrderTypeLimit,
						true,
						10.0,
						49000.0-float64(i*10),
						int64(i+100),
					)
					_ = e.processOrder(&buyOrder)
				}
			},
			operation: func(e *Engine, i int) {
				orderRequest := createTestOrderRequest(
					"market_user",
					orderbookv1.OrderTypeMarket,
					i%2 == 0,
					2.0, // Smaller size for parallel test
					0.0,
					int64(i+200),
				)
				_ = e.processOrder(&orderRequest)
			},
			cleanup: func(e *Engine) {},
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			engine := tc.setupEngine(b)
			tc.setupData(engine, b)

			b.ResetTimer()

			if tc.name == "parallel_market_orders" {
				b.RunParallel(func(pb *testing.PB) {
					i := 0
					for pb.Next() {
						tc.operation(engine, i)
						i++
					}
				})
			} else {
				for i := 0; i < b.N; i++ {
					tc.operation(engine, i)
				}
			}

			b.StopTimer()
			tc.cleanup(engine)
		})
	}
}

func BenchmarkEngine_SnapshotCreation(b *testing.B) {
	testCases := []benchmarkTestCase{
		{
			name: "snapshot_small_orderbook",
			setupEngine: func(b *testing.B) *Engine {
				ctrl := gomock.NewController(b)
				mockOrderReader := orderreadermock.NewMockOrderReader(ctrl)
				mockSnapshotStore := snapshotmock.NewMockStore(ctrl)
				mockMatchPublisher := matchpublisherv1_mock.NewMockMatchPublisher(ctrl)

				ob := orderbook.NewOrderbook()
				log, _ := logger.NewLogger()
				cfg := &config.Config{Pair: "BTC-USD"}

				mockSnapshotStore.EXPECT().
					LoadStore(gomock.Any()).
					Return(nil, nil).
					Times(1)

				// Expect multiple snapshot stores for benchmark
				mockSnapshotStore.EXPECT().
					Store(gomock.Any(), gomock.Any()).
					Return(nil).
					AnyTimes()

				mockMatchPublisher.EXPECT().
					PublishMatchEvent(gomock.Any(), gomock.Any()).
					Return(nil).
					AnyTimes()

				engine := NewEngine(ob, mockOrderReader, mockSnapshotStore, mockMatchPublisher, log, cfg)
				engine.ctx = context.Background()
				return engine
			},
			setupData: func(e *Engine, b *testing.B) {
				// Small orderbook - 100 orders
				for i := 0; i < 100; i++ {
					orderRequest := createTestOrderRequest(
						"user",
						orderbookv1.OrderTypeLimit,
						i%2 == 0,
						10.0,
						50000.0+float64(i),
						int64(i),
					)
					_ = e.processOrder(&orderRequest)
				}
				e.setOrderOffset(100)
			},
			operation: func(e *Engine, i int) {
				e.createAndStoreSnapshot()
			},
			cleanup: func(e *Engine) {},
		},
		{
			name: "snapshot_large_orderbook",
			setupEngine: func(b *testing.B) *Engine {
				ctrl := gomock.NewController(b)
				mockOrderReader := orderreadermock.NewMockOrderReader(ctrl)
				mockSnapshotStore := snapshotmock.NewMockStore(ctrl)
				mockMatchPublisher := matchpublisherv1_mock.NewMockMatchPublisher(ctrl)

				ob := orderbook.NewOrderbook()
				log, _ := logger.NewLogger()
				cfg := &config.Config{Pair: "BTC-USD"}

				mockSnapshotStore.EXPECT().
					LoadStore(gomock.Any()).
					Return(nil, nil).
					Times(1)

				mockSnapshotStore.EXPECT().
					Store(gomock.Any(), gomock.Any()).
					Return(nil).
					AnyTimes()

				mockMatchPublisher.EXPECT().
					PublishMatchEvent(gomock.Any(), gomock.Any()).
					Return(nil).
					AnyTimes()

				engine := NewEngine(ob, mockOrderReader, mockSnapshotStore, mockMatchPublisher, log, cfg)
				engine.ctx = context.Background()
				return engine
			},
			setupData: func(e *Engine, b *testing.B) {
				// Large orderbook - 1,000 orders (reduced from 10,000 for faster benchmarks)
				for i := 0; i < 1000; i++ {
					orderRequest := createTestOrderRequest(
						"user",
						orderbookv1.OrderTypeLimit,
						i%2 == 0,
						10.0,
						50000.0+float64(i),
						int64(i),
					)
					_ = e.processOrder(&orderRequest)
				}
				e.setOrderOffset(1000)
			},
			operation: func(e *Engine, i int) {
				e.createAndStoreSnapshot()
			},
			cleanup: func(e *Engine) {},
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			engine := tc.setupEngine(b)
			tc.setupData(engine, b)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tc.operation(engine, i)
			}
			b.StopTimer()

			tc.cleanup(engine)
		})
	}
}

func BenchmarkEngine_MixedOperations(b *testing.B) {
	testCases := []benchmarkTestCase{
		{
			name:        "mixed_orders_realistic_workload",
			setupEngine: setupBenchmarkEngine,
			setupData: func(e *Engine, b *testing.B) {
				// Pre-populate with some initial liquidity
				for i := 0; i < 50; i++ {
					sellOrder := createTestOrderRequest(
						"initial_seller",
						orderbookv1.OrderTypeLimit,
						false,
						10.0,
						50000.0+float64(i*50),
						int64(i),
					)
					_ = e.processOrder(&sellOrder)

					buyOrder := createTestOrderRequest(
						"initial_buyer",
						orderbookv1.OrderTypeLimit,
						true,
						10.0,
						49000.0-float64(i*50),
						int64(i+50),
					)
					_ = e.processOrder(&buyOrder)
				}
			},
			operation: func(e *Engine, i int) {
				var orderRequest orderbookv1.PlaceOrderRequest

				switch i % 10 {
				case 0, 1: // 20% market orders
					orderRequest = createTestOrderRequest(
						"market_user",
						orderbookv1.OrderTypeMarket,
						i%2 == 0,
						5.0,
						0.0,
						int64(i),
					)
					_ = e.processOrder(&orderRequest)
				default: // 80% limit orders
					orderRequest = createTestOrderRequest(
						"limit_user",
						orderbookv1.OrderTypeLimit,
						i%2 == 0,
						10.0,
						50000.0+float64((i%1000)-500),
						int64(i),
					)
					_ = e.processOrder(&orderRequest)
				}

				// Occasionally check stats (simulates monitoring)
				if i%100 == 0 {
					_ = e.GetOrderOffset()
					_ = e.GetLastSnapshotOffset()
					_ = e.GetTotalMatches()
				}
			},
			cleanup: func(e *Engine) {},
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			engine := tc.setupEngine(b)
			tc.setupData(engine, b)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tc.operation(engine, i)
			}
			b.StopTimer()

			tc.cleanup(engine)
		})
	}
}

func BenchmarkEngine_StateAccess(b *testing.B) {
	testCases := []benchmarkTestCase{
		{
			name:        "concurrent_offset_access",
			setupEngine: setupBenchmarkEngine,
			setupData:   func(e *Engine, b *testing.B) {},
			operation: func(e *Engine, i int) {
				// Mix of reads and writes
				if i%3 == 0 {
					e.setOrderOffset(int64(i))
				} else if i%3 == 1 {
					e.setLastSnapshotOffset(int64(i))
				} else {
					_ = e.GetOrderOffset()
					_ = e.GetLastSnapshotOffset()
				}
			},
			cleanup: func(e *Engine) {},
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			engine := tc.setupEngine(b)
			tc.setupData(engine, b)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tc.operation(engine, i)
			}
			b.StopTimer()

			tc.cleanup(engine)
		})
	}
}

// Memory allocation benchmarks
func BenchmarkEngine_MemoryAllocation(b *testing.B) {
	engine := setupBenchmarkEngine(b)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		orderRequest := createTestOrderRequest(
			"user",
			orderbookv1.OrderTypeLimit,
			i%2 == 0,
			10.0,
			50000.0+float64(i%100),
			int64(i),
		)
		_ = engine.processOrder(&orderRequest)
	}
}

// Helper function to create test order requests (uses the same function from engine_test.go)
// func createTestOrderRequest(userID string, orderType orderbookv1.OrderType, bid bool, size, price float64, offset int64) orderbookv1.PlaceOrderRequest {
// 	return orderbookv1.PlaceOrderRequest{
// 		UserID: userID,
// 		Type:   orderType,
// 		Bid:    bid,
// 		Size:   size,
// 		Price:  price,
// 		Offset: offset,
// 	}
// }
