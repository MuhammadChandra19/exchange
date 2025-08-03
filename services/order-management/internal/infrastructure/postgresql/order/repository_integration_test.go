package order

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/muhammadchandra19/exchange/pkg/logger"
	mockLogger "github.com/muhammadchandra19/exchange/pkg/logger/mock"
	"github.com/muhammadchandra19/exchange/pkg/postgresql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type RepositoryTestSuite struct {
	suite.Suite
	helper *postgresql.TestHelper
	repo   OrderRepository
	ctx    context.Context
}

// SetupSuite runs once before all tests
func (suite *RepositoryTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Get absolute path to migrations
	migrationsPath, err := filepath.Abs("../migrations")
	require.NoError(suite.T(), err)

	// Create test helper with actual migrations
	config := &postgresql.TestContainerConfig{
		Image:            "postgres:15-alpine",
		Database:         "order_test_db",
		Username:         "order_test_user",
		Password:         "order_test_pass",
		MigrationsPath:   migrationsPath,
		MigrationPattern: "*.up.sql", // Only run UP migrations
		StartupTimeout:   3 * time.Minute,
	}

	suite.helper = postgresql.NewTestHelperWithConfig(suite.T(), config)

	logger, err := logger.NewLogger()
	require.NoError(suite.T(), err)
	suite.repo = NewRepository(suite.helper.GetClient(), logger)
}

// SetupTest runs before each test
func (suite *RepositoryTestSuite) SetupTest() {
	// Clean up tables before each test
	suite.helper.CleanupTables()
}

// Test Store method
func (suite *RepositoryTestSuite) TestStore() {
	tests := []struct {
		name        string
		order       *Order
		expectError bool
		mockFn      func(mockLogger *mockLogger.MockInterface)
	}{
		{
			name: "valid order",
			order: &Order{
				ID:        "order-123",
				UserID:    "user-456",
				Symbol:    "BTC/USDT",
				Side:      "buy",
				Price:     50000.12345678,
				Quantity:  100000000, // 1 BTC in satoshis
				Type:      "limit",
				Status:    OrderStatusPlaced,
				Timestamp: time.Now(),
			},
			expectError: false,
		},
		{
			name: "duplicate order ID",
			order: &Order{
				ID:        "order-123", // Same ID as above
				UserID:    "user-789",
				Symbol:    "ETH/USDT",
				Side:      "sell",
				Price:     3000.0,
				Quantity:  200000000,
				Type:      "market",
				Status:    OrderStatusPlaced,
				Timestamp: time.Now(),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := suite.repo.Store(suite.ctx, tt.order)

			if tt.expectError {
				assert.Error(suite.T(), err)
			} else {
				assert.NoError(suite.T(), err)

				// Verify the order was stored
				storedOrder, err := suite.repo.GetByID(suite.ctx, tt.order.ID)
				require.NoError(suite.T(), err)
				require.NotNil(suite.T(), storedOrder)

				assert.Equal(suite.T(), tt.order.ID, storedOrder.ID)
				assert.Equal(suite.T(), tt.order.UserID, storedOrder.UserID)
				assert.Equal(suite.T(), tt.order.Symbol, storedOrder.Symbol)
				assert.Equal(suite.T(), tt.order.Side, storedOrder.Side)
				assert.Equal(suite.T(), tt.order.Price, storedOrder.Price)
				assert.Equal(suite.T(), tt.order.Quantity, storedOrder.Quantity)
				assert.Equal(suite.T(), tt.order.Type, storedOrder.Type)
				assert.Equal(suite.T(), tt.order.Status, storedOrder.Status)
			}
		})
	}
}

// Test StoreBatch method
func (suite *RepositoryTestSuite) TestStoreBatch() {
	orders := []*Order{
		{
			ID:        "batch-order-1",
			UserID:    "user-1",
			Symbol:    "BTC/USDT",
			Side:      "buy",
			Price:     50000.0,
			Quantity:  100000000,
			Type:      "limit",
			Status:    OrderStatusPlaced,
			Timestamp: time.Now(),
		},
		{
			ID:        "batch-order-2",
			UserID:    "user-2",
			Symbol:    "ETH/USDT",
			Side:      "sell",
			Price:     3000.0,
			Quantity:  200000000,
			Type:      "market",
			Status:    OrderStatusPlaced,
			Timestamp: time.Now(),
		},
		{
			ID:        "batch-order-3",
			UserID:    "user-1",
			Symbol:    "BTC/USDT",
			Side:      "sell",
			Price:     51000.0,
			Quantity:  50000000,
			Type:      "limit",
			Status:    OrderStatusPlaced,
			Timestamp: time.Now(),
		},
	}

	err := suite.repo.StoreBatch(suite.ctx, orders)
	assert.NoError(suite.T(), err)

	// Verify all orders were stored
	for _, order := range orders {
		storedOrder, err := suite.repo.GetByID(suite.ctx, order.ID)
		require.NoError(suite.T(), err)
		require.NotNil(suite.T(), storedOrder)
		assert.Equal(suite.T(), order.ID, storedOrder.ID)
	}
}

// Test GetByID method
func (suite *RepositoryTestSuite) TestGetByID() {
	// First, store an order
	order := &Order{
		ID:        "get-test-order",
		UserID:    "user-123",
		Symbol:    "BTC/USDT",
		Side:      "buy",
		Price:     50000.0,
		Quantity:  100000000,
		Type:      "limit",
		Status:    OrderStatusPlaced,
		Timestamp: time.Now(),
	}

	err := suite.repo.Store(suite.ctx, order)
	require.NoError(suite.T(), err)

	tests := []struct {
		name        string
		orderID     string
		expectOrder bool
		expectError bool
	}{
		{
			name:        "existing order",
			orderID:     "get-test-order",
			expectOrder: true,
			expectError: false,
		},
		{
			name:        "non-existing order",
			orderID:     "non-existing-order",
			expectOrder: false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result, err := suite.repo.GetByID(suite.ctx, tt.orderID)

			if tt.expectError {
				assert.Error(suite.T(), err)
			} else {
				assert.NoError(suite.T(), err)

				if tt.expectOrder {
					require.NotNil(suite.T(), result)
					assert.Equal(suite.T(), tt.orderID, result.ID)
				} else {
					assert.Nil(suite.T(), result)
				}
			}
		})
	}
}

// Test Update method
func (suite *RepositoryTestSuite) TestUpdate() {
	// First, store an order
	originalOrder := &Order{
		ID:        "update-test-order",
		UserID:    "user-123",
		Symbol:    "BTC/USDT",
		Side:      "buy",
		Price:     50000.0,
		Quantity:  100000000,
		Type:      "limit",
		Status:    OrderStatusPlaced,
		Timestamp: time.Now(),
	}

	err := suite.repo.Store(suite.ctx, originalOrder)
	require.NoError(suite.T(), err)

	// Update the order
	updatedOrder := &Order{
		ID:        "update-test-order",
		UserID:    "user-123",
		Symbol:    "BTC/USDT",
		Side:      "buy",
		Price:     51000.0,   // Changed price
		Quantity:  150000000, // Changed quantity
		Type:      "limit",
		Status:    OrderStatusModified, // Changed status
		Timestamp: time.Now(),
	}

	err = suite.repo.Update(suite.ctx, updatedOrder)
	assert.NoError(suite.T(), err)

	// Verify the update
	storedOrder, err := suite.repo.GetByID(suite.ctx, "update-test-order")
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), storedOrder)

	assert.Equal(suite.T(), updatedOrder.Price, storedOrder.Price)
	assert.Equal(suite.T(), updatedOrder.Quantity, storedOrder.Quantity)
	assert.Equal(suite.T(), updatedOrder.Status, storedOrder.Status)
}

// Test Delete method
func (suite *RepositoryTestSuite) TestDelete() {
	// First, store an order
	order := &Order{
		ID:        "delete-test-order",
		UserID:    "user-123",
		Symbol:    "BTC/USDT",
		Side:      "buy",
		Price:     50000.0,
		Quantity:  100000000,
		Type:      "limit",
		Status:    OrderStatusPlaced,
		Timestamp: time.Now(),
	}

	err := suite.repo.Store(suite.ctx, order)
	require.NoError(suite.T(), err)

	// Verify it exists
	storedOrder, err := suite.repo.GetByID(suite.ctx, "delete-test-order")
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), storedOrder)

	// Delete the order
	err = suite.repo.Delete(suite.ctx, "delete-test-order")
	assert.NoError(suite.T(), err)

	// Verify it's gone
	deletedOrder, err := suite.repo.GetByID(suite.ctx, "delete-test-order")
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), deletedOrder)

	// Test deleting non-existing order (should not error)
	err = suite.repo.Delete(suite.ctx, "non-existing-order")
	assert.NoError(suite.T(), err)
}

// Test List method
func (suite *RepositoryTestSuite) TestList() {
	// Setup test data
	testOrders := []*Order{
		{
			ID:        "list-order-1",
			UserID:    "user-1",
			Symbol:    "BTC/USDT",
			Side:      "buy",
			Price:     50000.0,
			Quantity:  100000000,
			Type:      "limit",
			Status:    OrderStatusPlaced,
			Timestamp: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:        "list-order-2",
			UserID:    "user-1",
			Symbol:    "ETH/USDT",
			Side:      "sell",
			Price:     3000.0,
			Quantity:  200000000,
			Type:      "market",
			Status:    OrderStatusCancelled,
			Timestamp: time.Now().Add(-1 * time.Hour),
		},
		{
			ID:        "list-order-3",
			UserID:    "user-2",
			Symbol:    "BTC/USDT",
			Side:      "sell",
			Price:     51000.0,
			Quantity:  50000000,
			Type:      "limit",
			Status:    OrderStatusPlaced,
			Timestamp: time.Now(),
		},
	}

	// Store test orders
	err := suite.repo.StoreBatch(suite.ctx, testOrders)
	require.NoError(suite.T(), err)

	tests := []struct {
		name          string
		filter        Filter
		expectedCount int
		expectedIDs   []string
	}{
		{
			name:          "no filter - get all",
			filter:        Filter{},
			expectedCount: 3,
			expectedIDs:   []string{"list-order-3", "list-order-2", "list-order-1"}, // DESC order
		},
		{
			name:          "filter by user",
			filter:        Filter{UserID: "user-1"},
			expectedCount: 2,
			expectedIDs:   []string{"list-order-2", "list-order-1"},
		},
		{
			name:          "filter by symbol",
			filter:        Filter{Symbol: "BTC/USDT"},
			expectedCount: 2,
			expectedIDs:   []string{"list-order-3", "list-order-1"},
		},
		{
			name:          "filter by side",
			filter:        Filter{Side: "buy"},
			expectedCount: 1,
			expectedIDs:   []string{"list-order-1"},
		},
		{
			name:          "filter by status",
			filter:        Filter{Status: "placed"},
			expectedCount: 2,
			expectedIDs:   []string{"list-order-3", "list-order-1"},
		},
		{
			name:          "combined filters",
			filter:        Filter{UserID: "user-1", Symbol: "BTC/USDT"},
			expectedCount: 1,
			expectedIDs:   []string{"list-order-1"},
		},
		{
			name:          "with limit",
			filter:        Filter{Limit: 2},
			expectedCount: 2,
			expectedIDs:   []string{"list-order-3", "list-order-2"},
		},
		{
			name:          "with offset",
			filter:        Filter{Offset: 1},
			expectedCount: 2,
			expectedIDs:   []string{"list-order-2", "list-order-1"},
		},
		{
			name:          "ascending sort",
			filter:        Filter{SortDirection: "ASC"},
			expectedCount: 3,
			expectedIDs:   []string{"list-order-1", "list-order-2", "list-order-3"},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			results, err := suite.repo.List(suite.ctx, tt.filter)
			require.NoError(suite.T(), err)

			assert.Len(suite.T(), results, tt.expectedCount)

			// Check order of results
			for i, expectedID := range tt.expectedIDs {
				if i < len(results) {
					assert.Equal(suite.T(), expectedID, results[i].ID)
				}
			}
		})
	}
}

// Test List with time filters
func (suite *RepositoryTestSuite) TestListWithTimeFilter() {
	now := time.Now()

	// Store orders with different timestamps
	testOrders := []*Order{
		{
			ID:        "time-order-1",
			UserID:    "user-1",
			Symbol:    "BTC/USDT",
			Side:      "buy",
			Price:     50000.0,
			Quantity:  100000000,
			Type:      "limit",
			Status:    OrderStatusPlaced,
			Timestamp: now.Add(-3 * time.Hour),
		},
		{
			ID:        "time-order-2",
			UserID:    "user-1",
			Symbol:    "BTC/USDT",
			Side:      "buy",
			Price:     50000.0,
			Quantity:  100000000,
			Type:      "limit",
			Status:    OrderStatusPlaced,
			Timestamp: now.Add(-1 * time.Hour),
		},
		{
			ID:        "time-order-3",
			UserID:    "user-1",
			Symbol:    "BTC/USDT",
			Side:      "buy",
			Price:     50000.0,
			Quantity:  100000000,
			Type:      "limit",
			Status:    OrderStatusPlaced,
			Timestamp: now,
		},
	}

	err := suite.repo.StoreBatch(suite.ctx, testOrders)
	require.NoError(suite.T(), err)

	// Test time range filter
	from := now.Add(-2 * time.Hour)
	to := now

	results, err := suite.repo.List(suite.ctx, Filter{
		From: &from,
		To:   &to,
	})

	require.NoError(suite.T(), err)
	assert.Len(suite.T(), results, 2) // Should get orders 2 and 3

	// Verify the correct orders
	ids := make([]string, len(results))
	for i, order := range results {
		ids[i] = order.ID
	}
	assert.Contains(suite.T(), ids, "time-order-2")
	assert.Contains(suite.T(), ids, "time-order-3")
	assert.NotContains(suite.T(), ids, "time-order-1")
}

// Test error handling
func (suite *RepositoryTestSuite) TestErrorHandling() {
	// Test with invalid context (cancelled)
	cancelledCtx, cancel := context.WithCancel(suite.ctx)
	cancel() // Cancel immediately

	order := &Order{
		ID:        "error-test-order",
		UserID:    "user-123",
		Symbol:    "BTC/USDT",
		Side:      "buy",
		Price:     50000.0,
		Quantity:  100000000,
		Type:      "limit",
		Status:    OrderStatusPlaced,
		Timestamp: time.Now(),
	}

	err := suite.repo.Store(cancelledCtx, order)
	assert.Error(suite.T(), err)
}

// Run the test suite
func TestRepositoryIntegration(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}
