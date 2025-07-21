package engine

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/muhammadchandra19/exchange/pkg/logger"
	orderreaderv1 "github.com/muhammadchandra19/exchange/services/matching-service/internal/domain/order-reader/v1"
	orderreadermock "github.com/muhammadchandra19/exchange/services/matching-service/internal/domain/order-reader/v1/mock"
	orderbookv1 "github.com/muhammadchandra19/exchange/services/matching-service/internal/domain/orderbook/v1"
	snapshotv1 "github.com/muhammadchandra19/exchange/services/matching-service/internal/domain/snapshot/v1"
	snapshotmock "github.com/muhammadchandra19/exchange/services/matching-service/internal/domain/snapshot/v1/mock"
	"github.com/muhammadchandra19/exchange/services/matching-service/internal/usecase/orderbook"
	"github.com/muhammadchandra19/exchange/services/matching-service/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test fixtures and helpers
type testFixture struct {
	ctrl              *gomock.Controller
	mockOrderReader   *orderreadermock.MockOrderReader
	mockSnapshotStore *snapshotmock.MockStore
	orderbook         *orderbook.Orderbook
	logger            *logger.Logger
	config            *config.Config
}

func setupTestFixture(t *testing.T) *testFixture {
	ctrl := gomock.NewController(t)

	log, err := logger.NewLogger()
	require.NoError(t, err)

	return &testFixture{
		ctrl:              ctrl,
		mockOrderReader:   orderreadermock.NewMockOrderReader(ctrl),
		mockSnapshotStore: snapshotmock.NewMockStore(ctrl),
		orderbook:         orderbook.NewOrderbook(),
		logger:            log,
		config: &config.Config{
			Pair: "BTC-USD",
			KafkaConfig: config.KafkaConfig{
				Brokers: []string{"localhost:9092"},
				Topic:   "orders",
			},
			RedisConfig: config.RedisConfig{
				Addrs:    "localhost:6379",
				Password: "",
				DB:       0,
			},
		},
	}
}

func (f *testFixture) teardown() {
	f.ctrl.Finish()
}

func createTestOrderRequest(userID string, orderType orderbookv1.OrderType, bid bool, size, price float64, offset int64) orderbookv1.PlaceOrderRequest {
	return orderbookv1.PlaceOrderRequest{
		UserID: userID,
		Type:   orderType,
		Bid:    bid,
		Size:   size,
		Price:  price,
		Offset: offset,
	}
}

// Helper function to create engine with initialized context
func createTestEngine(fixture *testFixture) *Engine {
	engine := NewEngine(
		fixture.orderbook,
		fixture.mockOrderReader,
		fixture.mockSnapshotStore,
		fixture.logger,
		fixture.config,
	)

	// Initialize context to prevent nil pointer dereference
	engine.ctx = context.Background()

	return engine
}

func TestNewEngine(t *testing.T) {
	testCases := []struct {
		name                string
		setupMocks          func(*testFixture)
		expectedOrderOffset int64
		expectedError       bool
	}{
		{
			name: "successful engine creation with nil snapshot",
			setupMocks: func(f *testFixture) {
				f.mockSnapshotStore.EXPECT().
					LoadStore(gomock.Any()).
					Return(nil, nil).
					Times(1)
			},
			expectedOrderOffset: -1,
			expectedError:       false,
		},
		{
			name: "successful engine creation with existing snapshot",
			setupMocks: func(f *testFixture) {
				snapshot := &snapshotv1.Snapshot{
					OrderOffset: 100,
					OrderBookSnapshot: snapshotv1.OrderBookSnapshot{
						Orders: []snapshotv1.BookOrder{},
					},
				}
				f.mockSnapshotStore.EXPECT().
					LoadStore(gomock.Any()).
					Return(snapshot, nil).
					Times(1)
			},
			expectedOrderOffset: 100,
			expectedError:       false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := setupTestFixture(t)
			defer fixture.teardown()

			tc.setupMocks(fixture)

			engine := NewEngine(
				fixture.orderbook,
				fixture.mockOrderReader,
				fixture.mockSnapshotStore,
				fixture.logger,
				fixture.config,
			)

			assert.NotNil(t, engine)
			assert.Equal(t, tc.expectedOrderOffset, engine.GetOrderOffset())
			assert.Equal(t, fixture.orderbook, engine.orderbook)
			assert.Equal(t, fixture.mockOrderReader, engine.orderReader)
			assert.Equal(t, fixture.mockSnapshotStore, engine.snapshotStore)
		})
	}
}

func TestNewEngineWithOptions(t *testing.T) {
	testCases := []struct {
		name                     string
		options                  *Options
		setupMocks               func(*testFixture)
		expectedSnapshotInterval time.Duration
		expectedOffsetDelta      int64
	}{
		{
			name: "engine with custom options",
			options: &Options{
				SnapshotInterval:    10 * time.Second,
				SnapshotOffsetDelta: 500,
			},
			setupMocks: func(f *testFixture) {
				f.mockSnapshotStore.EXPECT().
					LoadStore(gomock.Any()).
					Return(nil, nil).
					Times(1)
			},
			expectedSnapshotInterval: 10 * time.Second,
			expectedOffsetDelta:      500,
		},
		{
			name:    "engine with default options",
			options: DefaultEngineOptions(),
			setupMocks: func(f *testFixture) {
				f.mockSnapshotStore.EXPECT().
					LoadStore(gomock.Any()).
					Return(nil, nil).
					Times(1)
			},
			expectedSnapshotInterval: DefaultEngineOptions().SnapshotInterval,
			expectedOffsetDelta:      DefaultEngineOptions().SnapshotOffsetDelta,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := setupTestFixture(t)
			defer fixture.teardown()

			tc.setupMocks(fixture)

			engine := NewEngineWithOptions(
				fixture.orderbook,
				fixture.mockOrderReader,
				fixture.mockSnapshotStore,
				fixture.logger,
				fixture.config,
				tc.options,
			)

			assert.NotNil(t, engine)
			assert.Equal(t, tc.expectedSnapshotInterval, engine.snapshotInterval)
			assert.Equal(t, tc.expectedOffsetDelta, engine.snapshotOffsetDelta)
		})
	}
}

func TestEngine_ProcessOrder(t *testing.T) {
	testCases := []struct {
		name           string
		orderRequest   *orderbookv1.PlaceOrderRequest
		setupMocks     func(*testFixture)
		setupOrderbook func(*orderbook.Orderbook)
		expectedError  bool
		expectedOrders int
		expectedLimits int
		expectMatch    bool
	}{
		{
			name: "process valid limit order",
			orderRequest: &orderbookv1.PlaceOrderRequest{
				UserID: "user1",
				Type:   orderbookv1.OrderTypeLimit,
				Bid:    false,
				Size:   10.0,
				Price:  50000.0,
				Offset: 1,
			},
			setupMocks:     func(f *testFixture) {},
			setupOrderbook: func(ob *orderbook.Orderbook) {},
			expectedError:  false,
			expectedOrders: 1,
			expectedLimits: 1,
			expectMatch:    false,
		},
		{
			name: "process market order with existing limit order",
			orderRequest: &orderbookv1.PlaceOrderRequest{
				UserID: "buyer",
				Type:   orderbookv1.OrderTypeMarket,
				Bid:    true,
				Size:   5.0,
				Price:  0.0,
				Offset: 2,
			},
			setupMocks: func(f *testFixture) {},
			setupOrderbook: func(ob *orderbook.Orderbook) {
				// Add a sell limit order first
				sellOrder := orderbookv1.NewOrder("seller", 10.0, false)
				ob.PlaceLimitOrder(49000.0, sellOrder)
			},
			expectedError:  false,
			expectedOrders: 1, // Original sell order remains (partially filled)
			expectedLimits: 1,
			expectMatch:    true,
		},
		{
			name: "process invalid limit order - negative price",
			orderRequest: &orderbookv1.PlaceOrderRequest{
				UserID: "user1",
				Type:   orderbookv1.OrderTypeLimit,
				Bid:    false,
				Size:   10.0,
				Price:  -1.0,
				Offset: 3,
			},
			setupMocks:     func(f *testFixture) {},
			setupOrderbook: func(ob *orderbook.Orderbook) {},
			expectedError:  true,
			expectedOrders: 0,
			expectedLimits: 0,
			expectMatch:    false,
		},
		{
			name: "process invalid order - zero size",
			orderRequest: &orderbookv1.PlaceOrderRequest{
				UserID: "user1",
				Type:   orderbookv1.OrderTypeLimit,
				Bid:    false,
				Size:   0.0,
				Price:  50000.0,
				Offset: 4,
			},
			setupMocks:     func(f *testFixture) {},
			setupOrderbook: func(ob *orderbook.Orderbook) {},
			expectedError:  true,
			expectedOrders: 0,
			expectedLimits: 0,
			expectMatch:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := setupTestFixture(t)
			defer fixture.teardown()

			tc.setupMocks(fixture)

			// Setup snapshot loading
			fixture.mockSnapshotStore.EXPECT().
				LoadStore(gomock.Any()).
				Return(nil, nil).
				Times(1)

			engine := createTestEngine(fixture)

			// Setup orderbook state if needed
			tc.setupOrderbook(fixture.orderbook)

			// Get initial match count
			initialMatches := engine.GetTotalMatches()

			// Process the order
			err := engine.processOrder(tc.orderRequest)

			// Assertions
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expectedOrders, len(fixture.orderbook.Orders))

			totalLimits := len(fixture.orderbook.AskLimits) + len(fixture.orderbook.BidLimits)
			assert.Equal(t, tc.expectedLimits, totalLimits)

			// FIXED: Check match count instead of channel
			if tc.expectMatch {
				finalMatches := engine.GetTotalMatches()
				assert.Greater(t, finalMatches, initialMatches, "Expected matches to be generated")
			} else {
				finalMatches := engine.GetTotalMatches()
				assert.Equal(t, initialMatches, finalMatches, "Expected no matches to be generated")
			}
		})
	}
}

func TestEngine_SnapshotManagement(t *testing.T) {
	testCases := []struct {
		name                   string
		currentOffset          int64
		lastSnapshotOffset     int64
		snapshotOffsetDelta    int64
		setupMocks             func(*testFixture)
		expectedShouldSnapshot bool
		testCreateSnapshot     bool
		expectStoreSuccess     bool
	}{
		{
			name:                "should create snapshot when offset delta exceeded",
			currentOffset:       1000,
			lastSnapshotOffset:  0,
			snapshotOffsetDelta: 500,
			setupMocks: func(f *testFixture) {
				f.mockSnapshotStore.EXPECT().
					Store(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
			},
			expectedShouldSnapshot: true,
			testCreateSnapshot:     true,
			expectStoreSuccess:     true,
		},
		{
			name:                   "should not create snapshot when offset delta not exceeded",
			currentOffset:          100,
			lastSnapshotOffset:     50,
			snapshotOffsetDelta:    500,
			setupMocks:             func(f *testFixture) {},
			expectedShouldSnapshot: false,
			testCreateSnapshot:     false,
			expectStoreSuccess:     false,
		},
		{
			name:                   "should not create snapshot with zero current offset",
			currentOffset:          0,
			lastSnapshotOffset:     0,
			snapshotOffsetDelta:    100,
			setupMocks:             func(f *testFixture) {},
			expectedShouldSnapshot: false,
			testCreateSnapshot:     false,
			expectStoreSuccess:     false,
		},
		{
			name:                "should create snapshot and handle store error",
			currentOffset:       1000,
			lastSnapshotOffset:  0,
			snapshotOffsetDelta: 500,
			setupMocks: func(f *testFixture) {
				f.mockSnapshotStore.EXPECT().
					Store(gomock.Any(), gomock.Any()).
					Return(errors.New("store error")).
					Times(1)
			},
			expectedShouldSnapshot: true,
			testCreateSnapshot:     true,
			expectStoreSuccess:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := setupTestFixture(t)
			defer fixture.teardown()

			// Setup snapshot loading
			fixture.mockSnapshotStore.EXPECT().
				LoadStore(gomock.Any()).
				Return(nil, nil).
				Times(1)

			// Setup test-specific mocks
			tc.setupMocks(fixture)

			options := &Options{
				SnapshotInterval:    1 * time.Minute,
				SnapshotOffsetDelta: tc.snapshotOffsetDelta,
			}

			engine := NewEngineWithOptions(
				fixture.orderbook,
				fixture.mockOrderReader,
				fixture.mockSnapshotStore,
				fixture.logger,
				fixture.config,
				options,
			)

			// Initialize context for snapshot tests
			engine.ctx = context.Background()

			// Set up engine state
			engine.setOrderOffset(tc.currentOffset)
			engine.setLastSnapshotOffset(tc.lastSnapshotOffset)

			// Test shouldCreateSnapshot
			shouldSnapshot := engine.shouldCreateSnapshot()
			assert.Equal(t, tc.expectedShouldSnapshot, shouldSnapshot)

			// Test createAndStoreSnapshot if needed
			if tc.testCreateSnapshot {
				initialLastSnapshot := engine.GetLastSnapshotOffset()

				engine.createAndStoreSnapshot()

				// Check if last snapshot offset was updated based on store success
				if tc.expectStoreSuccess {
					assert.Equal(t, tc.currentOffset, engine.GetLastSnapshotOffset())
				} else {
					// If store failed, last snapshot offset should remain unchanged
					assert.Equal(t, initialLastSnapshot, engine.GetLastSnapshotOffset())
				}
			}
		})
	}
}

func TestEngine_ConcurrentAccess(t *testing.T) {
	testCases := []struct {
		name          string
		numGoroutines int
		numOperations int
		setupMocks    func(*testFixture)
		testOperation func(*Engine, int, int)
	}{
		{
			name:          "concurrent offset access",
			numGoroutines: 5,
			numOperations: 10,
			setupMocks: func(f *testFixture) {
				f.mockSnapshotStore.EXPECT().
					LoadStore(gomock.Any()).
					Return(nil, nil).
					Times(1)
			},
			testOperation: func(engine *Engine, goroutineID, operationID int) {
				// Concurrent writes
				engine.setOrderOffset(int64(goroutineID*1000 + operationID))
				engine.setLastSnapshotOffset(int64(goroutineID*500 + operationID))

				// Concurrent reads
				_ = engine.GetOrderOffset()
				_ = engine.GetLastSnapshotOffset()
				_ = engine.GetTotalMatches()
			},
		},
		{
			name:          "concurrent order processing",
			numGoroutines: 3,
			numOperations: 5,
			setupMocks: func(f *testFixture) {
				f.mockSnapshotStore.EXPECT().
					LoadStore(gomock.Any()).
					Return(nil, nil).
					Times(1)
			},
			testOperation: func(engine *Engine, goroutineID, operationID int) {
				orderRequest := createTestOrderRequest(
					"user",
					orderbookv1.OrderTypeLimit,
					goroutineID%2 == 0, // Alternate bid/ask
					10.0,
					50000.0+float64(goroutineID*100+operationID),
					int64(goroutineID*1000+operationID),
				)
				_ = engine.processOrder(&orderRequest)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := setupTestFixture(t)
			defer fixture.teardown()

			tc.setupMocks(fixture)

			engine := createTestEngine(fixture)

			// Run concurrent operations
			done := make(chan bool, tc.numGoroutines)

			for i := 0; i < tc.numGoroutines; i++ {
				go func(goroutineID int) {
					defer func() { done <- true }()
					for j := 0; j < tc.numOperations; j++ {
						tc.testOperation(engine, goroutineID, j)
					}
				}(i)
			}

			// Wait for all goroutines to complete
			for i := 0; i < tc.numGoroutines; i++ {
				select {
				case <-done:
				case <-time.After(2 * time.Second):
					t.Fatal("Test timeout - goroutines didn't complete")
				}
			}

			// Verify final state is consistent (no panics, no race conditions)
			finalOffset := engine.GetOrderOffset()
			assert.GreaterOrEqual(t, finalOffset, int64(-1))
		})
	}
}

// Test the new GetTotalMatches functionality
func TestEngine_GetTotalMatches(t *testing.T) {
	fixture := setupTestFixture(t)
	defer fixture.teardown()

	fixture.mockSnapshotStore.EXPECT().
		LoadStore(gomock.Any()).
		Return(nil, nil).
		Times(1)

	engine := createTestEngine(fixture)

	// Initially should be 0
	assert.Equal(t, int64(0), engine.GetTotalMatches())

	// Add a sell order
	sellOrder := orderbookv1.NewOrder("seller", 10.0, false)
	fixture.orderbook.PlaceLimitOrder(50000.0, sellOrder)

	// Process a market buy order that should create a match
	marketOrder := createTestOrderRequest("buyer", orderbookv1.OrderTypeMarket, true, 5.0, 0.0, 1)
	err := engine.processOrder(&marketOrder)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), engine.GetTotalMatches())
}

func TestEngine_Start(t *testing.T) {
	type fields struct {
		orderbook           orderbookv1.Orderbook
		orderReader         orderreaderv1.OrderReader
		snapshotStore       snapshotv1.Store
		logger              *logger.Logger
		config              *config.Config
		mu                  sync.RWMutex
		orderOffset         int64
		lastSnapshotOffset  int64
		ctx                 context.Context
		cancel              context.CancelFunc
		wg                  sync.WaitGroup
		snapshotInterval    time.Duration
		snapshotOffsetDelta int64
		totalMatches        int64
		matchesMutex        sync.RWMutex
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Engine{
				orderbook:           tt.fields.orderbook,
				orderReader:         tt.fields.orderReader,
				snapshotStore:       tt.fields.snapshotStore,
				logger:              tt.fields.logger,
				config:              tt.fields.config,
				mu:                  tt.fields.mu,
				orderOffset:         tt.fields.orderOffset,
				lastSnapshotOffset:  tt.fields.lastSnapshotOffset,
				ctx:                 tt.fields.ctx,
				cancel:              tt.fields.cancel,
				wg:                  tt.fields.wg,
				snapshotInterval:    tt.fields.snapshotInterval,
				snapshotOffsetDelta: tt.fields.snapshotOffsetDelta,
				totalMatches:        tt.fields.totalMatches,
				matchesMutex:        tt.fields.matchesMutex,
			}
			if err := e.Start(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Engine.Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
