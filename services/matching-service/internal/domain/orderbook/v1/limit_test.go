package orderbookv1

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test order
func createTestOrder(userID string, size float64, bid bool) *Order {
	order := NewOrder(userID, size, bid)
	order.Timestamp = time.Now().UnixNano()
	order.Sequence = 1
	return order
}

// Helper function to create an order with specific timestamp and sequence
func createOrderWithTimestamp(userID string, size float64, bid bool, timestamp int64, sequence int64) *Order {
	order := &Order{
		ID:        "test-id",
		UserID:    userID,
		Size:      size,
		Bid:       bid,
		Timestamp: timestamp,
		Sequence:  sequence,
	}
	return order
}

func TestNewLimit(t *testing.T) {
	limit := NewLimit(100.0)

	assert.NotNil(t, limit)
	assert.Equal(t, 100.0, limit.Price)
	assert.Equal(t, 0.0, limit.TotalVolume)
	assert.Empty(t, limit.Orders)
	assert.True(t, limit.IsEmpty())
}

func TestLimit_AddOrder(t *testing.T) {
	limit := NewLimit(100.0)

	t.Run("Add valid order", func(t *testing.T) {
		order := createTestOrder("user1", 10.0, true)
		err := limit.AddOrder(order)

		require.NoError(t, err)
		assert.Equal(t, 1, len(limit.Orders))
		assert.Equal(t, 10.0, limit.TotalVolume)
		assert.Equal(t, limit, order.Limit)
		assert.False(t, limit.IsEmpty())
	})

	t.Run("Add nil order", func(t *testing.T) {
		err := limit.AddOrder(nil)
		assert.ErrorIs(t, err, ErrNilOrder)
	})

	t.Run("Add order with zero size", func(t *testing.T) {
		order := createTestOrder("user1", 0.0, true)
		err := limit.AddOrder(order)
		assert.ErrorIs(t, err, ErrInvalidSize)
	})

	t.Run("Add multiple orders", func(t *testing.T) {
		limit := NewLimit(100.0)
		order1 := createTestOrder("user1", 10.0, true)
		order2 := createTestOrder("user2", 20.0, false)

		err1 := limit.AddOrder(order1)
		err2 := limit.AddOrder(order2)

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.Equal(t, 2, len(limit.Orders))
		assert.Equal(t, 30.0, limit.TotalVolume)
	})
}

func TestLimit_RemoveOrder(t *testing.T) {
	limit := NewLimit(100.0)
	order := createTestOrder("user1", 10.0, true)

	// Add order first
	require.NoError(t, limit.AddOrder(order))

	t.Run("Remove existing order", func(t *testing.T) {
		err := limit.RemoveOrder(order)

		require.NoError(t, err)
		assert.Equal(t, 0, len(limit.Orders))
		assert.Equal(t, 0.0, limit.TotalVolume)
		assert.Nil(t, order.Limit)
		assert.True(t, limit.IsEmpty())
	})

	t.Run("Remove nil order", func(t *testing.T) {
		err := limit.RemoveOrder(nil)
		assert.ErrorIs(t, err, ErrNilOrder)
	})
}

func TestLimit_Fill_Simple(t *testing.T) {
	t.Run("Simple partial fill", func(t *testing.T) {
		limit := NewLimit(100.0)

		// Add a sell order
		sellOrder := createTestOrder("seller", 10.0, false)
		err := limit.AddOrder(sellOrder)
		require.NoError(t, err)

		// Create incoming buy order (smaller)
		buyOrder := createTestOrder("buyer", 5.0, true)

		// Fill
		matches := limit.Fill(buyOrder)

		require.Equal(t, 1, len(matches))

		match := matches[0]
		assert.Equal(t, 5.0, match.SizeFilled)
		assert.Equal(t, 100.0, match.Price)
		assert.Equal(t, buyOrder, match.Bid)
		assert.Equal(t, sellOrder, match.Ask)

		// Check remaining sizes
		assert.Equal(t, 0.0, buyOrder.Size)  // Fully filled
		assert.Equal(t, 5.0, sellOrder.Size) // Partially filled

		// Check limit state
		assert.Equal(t, 1, len(limit.Orders))   // Sell order still there
		assert.Equal(t, 5.0, limit.TotalVolume) // Volume updated
		assert.False(t, limit.IsEmpty())
	})

	t.Run("Exact match", func(t *testing.T) {
		limit := NewLimit(100.0)

		// Add a sell order
		sellOrder := createTestOrder("seller", 10.0, false)
		err := limit.AddOrder(sellOrder)
		require.NoError(t, err)

		// Create incoming buy order of same size
		buyOrder := createTestOrder("buyer", 10.0, true)

		// Fill
		matches := limit.Fill(buyOrder)

		require.Equal(t, 1, len(matches))

		match := matches[0]
		assert.Equal(t, 10.0, match.SizeFilled)

		// Both orders should be fully filled
		assert.Equal(t, 0.0, buyOrder.Size)
		assert.Equal(t, 0.0, sellOrder.Size)

		// Limit should be empty
		assert.True(t, limit.IsEmpty())
		assert.Equal(t, 0.0, limit.TotalVolume)
	})
}

func TestLimit_Fill_FIFO(t *testing.T) {
	t.Run("FIFO ordering by timestamp", func(t *testing.T) {
		limit := NewLimit(100.0)

		// Create orders with different timestamps
		order1 := createOrderWithTimestamp("user1", 10.0, false, 1000, 1) // Ask, earliest
		order2 := createOrderWithTimestamp("user2", 15.0, false, 2000, 2) // Ask, latest
		order3 := createOrderWithTimestamp("user3", 8.0, false, 1500, 3)  // Ask, middle

		require.NoError(t, limit.AddOrder(order1))
		require.NoError(t, limit.AddOrder(order2))
		require.NoError(t, limit.AddOrder(order3))

		// Create incoming bid order
		incomingOrder := createTestOrder("buyer", 25.0, true)

		matches := limit.Fill(incomingOrder)

		// Should match in timestamp order: order1 (1000), order3 (1500), order2 (2000)
		require.Equal(t, 3, len(matches))

		// First match: order1 (timestamp 1000)
		assert.Equal(t, order1, matches[0].Ask)
		assert.Equal(t, 10.0, matches[0].SizeFilled)

		// Second match: order3 (timestamp 1500)
		assert.Equal(t, order3, matches[1].Ask)
		assert.Equal(t, 8.0, matches[1].SizeFilled)

		// Third match: order2 (timestamp 2000)
		assert.Equal(t, order2, matches[2].Ask)
		assert.Equal(t, 7.0, matches[2].SizeFilled) // Remaining 7 from 25 - 10 - 8

		// Check final state
		assert.Equal(t, 1, len(limit.Orders))   // Only order2 remains (partially filled)
		assert.Equal(t, 8.0, limit.TotalVolume) // 15 - 7 = 8 remaining
		assert.True(t, incomingOrder.IsFilled())
	})
}
