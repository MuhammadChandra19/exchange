package orderbook

import (
	"fmt"
	"sync"
	"testing"

	orderbookv1 "github.com/muhammadchandra19/exchange/services/matching-engine/internal/domain/orderbook/v1"
	snapshotv1 "github.com/muhammadchandra19/exchange/services/matching-engine/internal/domain/snapshot/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create test order with specific ID
func createTestOrder(userID, orderID string, size float64, bid bool) *orderbookv1.Order {
	order := orderbookv1.NewOrder(userID, size, bid)
	order.ID = orderID
	return order
}

// Test 1: Basic constructor
func TestNewOrderbook(t *testing.T) {
	ob := NewOrderbook()

	assert.NotNil(t, ob)
	assert.NotNil(t, ob.Orders)
	assert.NotNil(t, ob.AskLimits)
	assert.NotNil(t, ob.BidLimits)
	assert.Equal(t, 0, len(ob.Orders))
	assert.Equal(t, 0, len(ob.AskLimits))
	assert.Equal(t, 0, len(ob.BidLimits))
}

// Test 2: Place a single limit order
func TestOrderbook_PlaceLimitOrder_Basic(t *testing.T) {
	ob := NewOrderbook()

	order := createTestOrder("user1", "order1", 10.0, false) // Ask order
	err := ob.PlaceLimitOrder(10_000, order)

	require.NoError(t, err)
	assert.Equal(t, 1, len(ob.Orders))
	assert.Equal(t, 1, len(ob.AskLimits))
	assert.Equal(t, 0, len(ob.BidLimits))

	// Check the limit was created correctly
	limit, exists := ob.AskLimits[10_000]
	assert.True(t, exists)
	assert.Equal(t, 10_000.0, limit.Price)
	assert.Equal(t, 1, limit.OrderCount())
	assert.Equal(t, 10.0, limit.GetTotalVolume())
}

// Test 3: Place multiple orders at same price
func TestOrderbook_SamePriceLevel(t *testing.T) {
	ob := NewOrderbook()

	order1 := createTestOrder("user1", "order1", 10.0, false)
	order2 := createTestOrder("user2", "order2", 5.0, false)

	err1 := ob.PlaceLimitOrder(10_000, order1)
	err2 := ob.PlaceLimitOrder(10_000, order2)

	require.NoError(t, err1)
	require.NoError(t, err2)

	assert.Equal(t, 2, len(ob.Orders))
	assert.Equal(t, 1, len(ob.AskLimits)) // Same price level

	limit := ob.AskLimits[10_000]
	assert.Equal(t, 2, limit.OrderCount())
	assert.Equal(t, 15.0, limit.GetTotalVolume())
}

// Test 4: Buy market order (bid) against ask orders
func TestOrderbook_BuyMarketOrder_Basic(t *testing.T) {
	ob := NewOrderbook()

	// Place a sell order first
	sellOrder := createTestOrder("seller", "sell1", 10.0, false)
	err := ob.PlaceLimitOrder(10_000, sellOrder)
	require.NoError(t, err)

	// Place a buy market order
	buyOrder := createTestOrder("buyer", "buy1", 5.0, true)
	matches, err := ob.PlaceMarketOrder(buyOrder)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(matches))
	assert.Equal(t, 5.0, matches[0].SizeFilled)
	assert.Equal(t, 10_000.0, matches[0].Price)
	assert.Equal(t, buyOrder, matches[0].Bid)
	assert.Equal(t, sellOrder, matches[0].Ask)

	// Check remaining sizes
	assert.Equal(t, 0.0, buyOrder.Size)  // Fully filled
	assert.Equal(t, 5.0, sellOrder.Size) // Partially filled
}

// Test 5: Buy market order across multiple ask levels
func TestOrderbook_BuyMarketOrderMultipleFills(t *testing.T) {
	ob := NewOrderbook()

	// Add sell orders at different prices
	sellOrder1 := createTestOrder("seller1", "sell1", 5.0, false)
	sellOrder2 := createTestOrder("seller2", "sell2", 3.0, false)
	sellOrder3 := createTestOrder("seller3", "sell3", 7.0, false)

	ob.PlaceLimitOrder(10_000, sellOrder1) // Best ask
	ob.PlaceLimitOrder(10_100, sellOrder2)
	ob.PlaceLimitOrder(10_200, sellOrder3)

	// Large buy market order
	buyOrder := createTestOrder("buyer", "buy1", 12.0, true)
	matches, err := ob.PlaceMarketOrder(buyOrder)

	assert.NoError(t, err)
	assert.Equal(t, 3, len(matches))

	// Should match in price priority order
	assert.Equal(t, 5.0, matches[0].SizeFilled) // First limit fully filled
	assert.Equal(t, 3.0, matches[1].SizeFilled) // Second limit fully filled
	assert.Equal(t, 4.0, matches[2].SizeFilled) // Third limit partially filled

	assert.Equal(t, 10_000.0, matches[0].Price)
	assert.Equal(t, 10_100.0, matches[1].Price)
	assert.Equal(t, 10_200.0, matches[2].Price)

	// Check remaining sizes
	assert.Equal(t, 0.0, sellOrder1.Size) // Fully filled
	assert.Equal(t, 0.0, sellOrder2.Size) // Fully filled
	assert.Equal(t, 3.0, sellOrder3.Size) // Partially filled (7-4=3)
	assert.Equal(t, 0.0, buyOrder.Size)   // Fully filled
}

// Test 6: Cancel order
func TestOrderbook_CancelOrder(t *testing.T) {
	ob := NewOrderbook()

	order := createTestOrder("user1", "order1", 10.0, false)
	err := ob.PlaceLimitOrder(10_000, order)
	require.NoError(t, err)

	err = ob.CancelOrder("order1")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(ob.Orders))
	assert.Equal(t, 0, len(ob.AskLimits)) // Limit removed when empty
}

// Test 7: Error cases
func TestOrderbook_ErrorCases(t *testing.T) {
	ob := NewOrderbook()

	t.Run("Nil order", func(t *testing.T) {
		err := ob.PlaceLimitOrder(100.0, nil)
		assert.Error(t, err)
	})

	t.Run("Invalid price", func(t *testing.T) {
		order := createTestOrder("user1", "order1", 10.0, false)
		err := ob.PlaceLimitOrder(0, order)
		assert.Error(t, err)
	})

	t.Run("Cancel non-existent order", func(t *testing.T) {
		err := ob.CancelOrder("nonexistent")
		assert.Error(t, err)
	})
}

// Test 8: Snapshot and restore functionality
func TestOrderbook_SnapshotAndRestore(t *testing.T) {
	// Create original orderbook with some orders
	ob1 := NewOrderbook()

	// Add various orders
	sellOrder1 := createTestOrder("seller1", "sell1", 10.0, false)
	sellOrder2 := createTestOrder("seller2", "sell2", 5.0, false)
	buyOrder1 := createTestOrder("buyer1", "buy1", 8.0, true)
	buyOrder2 := createTestOrder("buyer2", "buy2", 3.0, true)

	ob1.PlaceLimitOrder(10_000, sellOrder1)
	ob1.PlaceLimitOrder(10_100, sellOrder2)
	ob1.PlaceLimitOrder(9_900, buyOrder1)
	ob1.PlaceLimitOrder(9_800, buyOrder2)

	// Create snapshot
	snapshot := ob1.CreateSnapshot()

	// Verify snapshot contains all orders
	assert.Equal(t, 4, len(snapshot.OrderBookSnapshot.Orders))

	// Create new orderbook and restore from snapshot
	ob2 := NewOrderbook()
	err := ob2.RestoreOrderbook(snapshot)
	require.NoError(t, err)

	// Verify restored state matches original
	assert.Equal(t, len(ob1.Orders), len(ob2.Orders))
	assert.Equal(t, len(ob1.AskLimits), len(ob2.AskLimits))
	assert.Equal(t, len(ob1.BidLimits), len(ob2.BidLimits))

	// Verify specific orders exist
	assert.Contains(t, ob2.Orders, "sell1")
	assert.Contains(t, ob2.Orders, "sell2")
	assert.Contains(t, ob2.Orders, "buy1")
	assert.Contains(t, ob2.Orders, "buy2")

	// Verify limits exist at correct prices
	assert.Contains(t, ob2.AskLimits, 10_000.0)
	assert.Contains(t, ob2.AskLimits, 10_100.0)
	assert.Contains(t, ob2.BidLimits, 9_900.0)
	assert.Contains(t, ob2.BidLimits, 9_800.0)

	// Verify order details are preserved
	restoredSell1 := ob2.Orders["sell1"]
	assert.Equal(t, sellOrder1.UserID, restoredSell1.UserID)
	assert.Equal(t, sellOrder1.Size, restoredSell1.Size)
	assert.Equal(t, sellOrder1.Bid, restoredSell1.Bid)
	assert.Equal(t, sellOrder1.Timestamp, restoredSell1.Timestamp)

	// Verify total volumes match
	assert.Equal(t, ob1.AskTotalVolume(), ob2.AskTotalVolume())
	assert.Equal(t, ob1.BidTotalVolume(), ob2.BidTotalVolume())
}

// Test 9: Restore with empty snapshot
func TestOrderbook_RestoreEmpty(t *testing.T) {
	ob := NewOrderbook()

	// Add some orders first
	order := createTestOrder("user1", "order1", 10.0, false)
	ob.PlaceLimitOrder(10_000, order)

	// Create empty snapshot
	emptySnapshot := &snapshotv1.Snapshot{
		OrderBookSnapshot: snapshotv1.OrderBookSnapshot{
			Orders: []snapshotv1.BookOrder{},
		},
	}

	// Restore should clear existing orders
	err := ob.RestoreOrderbook(emptySnapshot)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(ob.Orders))
	assert.Equal(t, 0, len(ob.AskLimits))
	assert.Equal(t, 0, len(ob.BidLimits))
}

// Test 10: Restore with nil snapshot
func TestOrderbook_RestoreNil(t *testing.T) {
	ob := NewOrderbook()

	err := ob.RestoreOrderbook(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "snapshot cannot be nil")
}

// Test 11: Functional test - restored orderbook should work the same
func TestOrderbook_RestoredFunctionality(t *testing.T) {
	// Create and populate original orderbook
	ob1 := NewOrderbook()
	sellOrder := createTestOrder("seller", "sell1", 10.0, false)
	ob1.PlaceLimitOrder(10_000, sellOrder)

	// Create snapshot and restore to new orderbook
	snapshot := ob1.CreateSnapshot()
	ob2 := NewOrderbook()
	err := ob2.RestoreOrderbook(snapshot)
	require.NoError(t, err)

	// Test that restored orderbook functions correctly
	buyOrder := createTestOrder("buyer", "buy1", 5.0, true)
	matches, err := ob2.PlaceMarketOrder(buyOrder)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(matches))
	assert.Equal(t, 5.0, matches[0].SizeFilled)
	assert.Equal(t, 10_000.0, matches[0].Price)

	// Verify the restored sell order was partially filled
	restoredSellOrder := ob2.Orders["sell1"]
	assert.Equal(t, 5.0, restoredSellOrder.Size) // Should have 5.0 remaining
}

// Test 6: Sell market order (ask) against bid orders
func TestOrderbook_SellMarketOrder(t *testing.T) {
	ob := NewOrderbook()

	// Place buy orders first (bids)
	buyOrder1 := createTestOrder("buyer1", "buy1", 10.0, true)
	buyOrder2 := createTestOrder("buyer2", "buy2", 5.0, true)

	ob.PlaceLimitOrder(9_900, buyOrder1) // Best bid
	ob.PlaceLimitOrder(9_800, buyOrder2) // Lower bid

	// Place a sell market order
	sellOrder := createTestOrder("seller", "sell1", 8.0, false)
	matches, err := ob.PlaceMarketOrder(sellOrder)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(matches))
	assert.Equal(t, 8.0, matches[0].SizeFilled)
	assert.Equal(t, 9_900.0, matches[0].Price) // Should match at best bid price
	assert.Equal(t, buyOrder1, matches[0].Bid)
	assert.Equal(t, sellOrder, matches[0].Ask)

	// Check remaining sizes
	assert.Equal(t, 0.0, sellOrder.Size) // Fully filled
	assert.Equal(t, 2.0, buyOrder1.Size) // Partially filled (10-8=2)
	assert.Equal(t, 5.0, buyOrder2.Size) // Untouched
}

// Test 7: Sell market order across multiple bid levels
func TestOrderbook_SellMarketOrderMultipleFills(t *testing.T) {
	ob := NewOrderbook()

	// Add buy orders at different prices (bids)
	buyOrder1 := createTestOrder("buyer1", "buy1", 5.0, true)
	buyOrder2 := createTestOrder("buyer2", "buy2", 3.0, true)
	buyOrder3 := createTestOrder("buyer3", "buy3", 7.0, true)

	ob.PlaceLimitOrder(9_900, buyOrder1) // Best bid
	ob.PlaceLimitOrder(9_800, buyOrder2) // Middle bid
	ob.PlaceLimitOrder(9_700, buyOrder3) // Lowest bid

	// Large sell market order
	sellOrder := createTestOrder("seller", "sell1", 12.0, false)
	matches, err := ob.PlaceMarketOrder(sellOrder)

	assert.NoError(t, err)
	assert.Equal(t, 3, len(matches))

	// Should match in price priority order (highest bid first)
	assert.Equal(t, 5.0, matches[0].SizeFilled) // First limit fully filled
	assert.Equal(t, 3.0, matches[1].SizeFilled) // Second limit fully filled
	assert.Equal(t, 4.0, matches[2].SizeFilled) // Third limit partially filled

	assert.Equal(t, 9_900.0, matches[0].Price)
	assert.Equal(t, 9_800.0, matches[1].Price)
	assert.Equal(t, 9_700.0, matches[2].Price)

	// Check remaining sizes
	assert.Equal(t, 0.0, buyOrder1.Size) // Fully filled
	assert.Equal(t, 0.0, buyOrder2.Size) // Fully filled
	assert.Equal(t, 3.0, buyOrder3.Size) // Partially filled (7-4=3)
	assert.Equal(t, 0.0, sellOrder.Size) // Fully filled
}

// Test 8: Mixed scenario - both buy and sell market orders
func TestOrderbook_MixedMarketOrders(t *testing.T) {
	ob := NewOrderbook()

	// Set up initial book with both asks and bids
	sellOrder1 := createTestOrder("seller1", "sell1", 10.0, false)
	sellOrder2 := createTestOrder("seller2", "sell2", 5.0, false)
	buyOrder1 := createTestOrder("buyer1", "buy1", 8.0, true)
	buyOrder2 := createTestOrder("buyer2", "buy2", 3.0, true)

	ob.PlaceLimitOrder(10_100, sellOrder1) // Ask
	ob.PlaceLimitOrder(10_200, sellOrder2) // Higher ask
	ob.PlaceLimitOrder(9_900, buyOrder1)   // Bid
	ob.PlaceLimitOrder(9_800, buyOrder2)   // Lower bid

	// Test buy market order
	buyMarketOrder := createTestOrder("market_buyer", "mbuy1", 7.0, true)
	buyMatches, err := ob.PlaceMarketOrder(buyMarketOrder)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(buyMatches))
	assert.Equal(t, 7.0, buyMatches[0].SizeFilled)
	assert.Equal(t, 10_100.0, buyMatches[0].Price) // Best ask price
	assert.Equal(t, 3.0, sellOrder1.Size)          // Partially filled

	// Test sell market order
	sellMarketOrder := createTestOrder("market_seller", "msell1", 6.0, false)
	sellMatches, err := ob.PlaceMarketOrder(sellMarketOrder)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(sellMatches))
	assert.Equal(t, 6.0, sellMatches[0].SizeFilled)
	assert.Equal(t, 9_900.0, sellMatches[0].Price) // Best bid price
	assert.Equal(t, 2.0, buyOrder1.Size)           // Partially filled
}

// Test 9: Market order with no matching orders
func TestOrderbook_MarketOrderNoMatches(t *testing.T) {
	ob := NewOrderbook()

	t.Run("Buy market order with no asks", func(t *testing.T) {
		// Only add bid orders
		buyOrder1 := createTestOrder("buyer1", "buy1", 10.0, true)
		ob.PlaceLimitOrder(9_900, buyOrder1)

		// Try to place buy market order (should find no asks to match)
		buyMarketOrder := createTestOrder("market_buyer", "mbuy1", 5.0, true)
		matches, err := ob.PlaceMarketOrder(buyMarketOrder)

		assert.NoError(t, err)
		assert.Equal(t, 0, len(matches))
		assert.Equal(t, 5.0, buyMarketOrder.Size) // Unchanged
	})

	t.Run("Sell market order with no bids", func(t *testing.T) {
		// Clear and add only ask orders
		ob2 := NewOrderbook()
		sellOrder1 := createTestOrder("seller1", "sell1", 10.0, false)
		ob2.PlaceLimitOrder(10_100, sellOrder1)

		// Try to place sell market order (should find no bids to match)
		sellMarketOrder := createTestOrder("market_seller", "msell1", 5.0, false)
		matches, err := ob2.PlaceMarketOrder(sellMarketOrder)

		assert.NoError(t, err)
		assert.Equal(t, 0, len(matches))
		assert.Equal(t, 5.0, sellMarketOrder.Size) // Unchanged
	})
}

// Test 10: Market order exactly fills available liquidity
func TestOrderbook_MarketOrderExactFill(t *testing.T) {
	t.Run("Buy market order exact fill", func(t *testing.T) {
		ob := NewOrderbook()

		sellOrder1 := createTestOrder("seller1", "sell1", 5.0, false)
		sellOrder2 := createTestOrder("seller2", "sell2", 3.0, false)
		ob.PlaceLimitOrder(10_100, sellOrder1)
		ob.PlaceLimitOrder(10_200, sellOrder2)

		// Buy market order that exactly matches total ask volume
		buyMarketOrder := createTestOrder("buyer", "buy1", 8.0, true)
		matches, err := ob.PlaceMarketOrder(buyMarketOrder)

		assert.NoError(t, err)
		assert.Equal(t, 2, len(matches))
		assert.Equal(t, 0.0, buyMarketOrder.Size) // Fully filled
		assert.Equal(t, 0, len(ob.AskLimits))     // All ask limits consumed
	})

	t.Run("Sell market order exact fill", func(t *testing.T) {
		ob := NewOrderbook()

		buyOrder1 := createTestOrder("buyer1", "buy1", 5.0, true)
		buyOrder2 := createTestOrder("buyer2", "buy2", 3.0, true)
		ob.PlaceLimitOrder(9_900, buyOrder1)
		ob.PlaceLimitOrder(9_800, buyOrder2)

		// Sell market order that exactly matches total bid volume
		sellMarketOrder := createTestOrder("seller", "sell1", 8.0, false)
		matches, err := ob.PlaceMarketOrder(sellMarketOrder)

		assert.NoError(t, err)
		assert.Equal(t, 2, len(matches))
		assert.Equal(t, 0.0, sellMarketOrder.Size) // Fully filled
		assert.Equal(t, 0, len(ob.BidLimits))      // All bid limits consumed
	})
}

// Test 11: Asks() method - returns ask limits sorted by price (ascending)
func TestOrderbook_Asks(t *testing.T) {
	ob := NewOrderbook()

	t.Run("Empty orderbook", func(t *testing.T) {
		asks := ob.Asks()
		assert.Equal(t, 0, len(asks))
	})

	t.Run("Single ask limit", func(t *testing.T) {
		sellOrder := createTestOrder("seller", "sell1", 10.0, false)
		ob.PlaceLimitOrder(10_000, sellOrder)

		asks := ob.Asks()
		assert.Equal(t, 1, len(asks))
		assert.Equal(t, 10_000.0, asks[0].Price)
		assert.Equal(t, 10.0, asks[0].GetTotalVolume())
	})

	t.Run("Multiple ask limits sorted by price", func(t *testing.T) {
		ob := NewOrderbook() // Fresh orderbook

		// Add asks in random price order
		sellOrder1 := createTestOrder("seller1", "sell1", 5.0, false)
		sellOrder2 := createTestOrder("seller2", "sell2", 3.0, false)
		sellOrder3 := createTestOrder("seller3", "sell3", 7.0, false)

		ob.PlaceLimitOrder(10_200, sellOrder1) // Highest price
		ob.PlaceLimitOrder(10_000, sellOrder2) // Lowest price (best ask)
		ob.PlaceLimitOrder(10_100, sellOrder3) // Middle price

		asks := ob.Asks()
		assert.Equal(t, 3, len(asks))

		// Should be sorted by price ascending (best ask first)
		assert.Equal(t, 10_000.0, asks[0].Price) // Best ask (lowest price)
		assert.Equal(t, 10_100.0, asks[1].Price) // Middle
		assert.Equal(t, 10_200.0, asks[2].Price) // Worst ask (highest price)

		// Verify volumes
		assert.Equal(t, 3.0, asks[0].GetTotalVolume())
		assert.Equal(t, 7.0, asks[1].GetTotalVolume())
		assert.Equal(t, 5.0, asks[2].GetTotalVolume())
	})

	t.Run("Multiple orders at same price level", func(t *testing.T) {
		ob := NewOrderbook()

		sellOrder1 := createTestOrder("seller1", "sell1", 5.0, false)
		sellOrder2 := createTestOrder("seller2", "sell2", 3.0, false)

		ob.PlaceLimitOrder(10_000, sellOrder1)
		ob.PlaceLimitOrder(10_000, sellOrder2) // Same price level

		asks := ob.Asks()
		assert.Equal(t, 1, len(asks)) // Only one limit at this price
		assert.Equal(t, 10_000.0, asks[0].Price)
		assert.Equal(t, 8.0, asks[0].GetTotalVolume()) // Combined volume
		assert.Equal(t, 2, asks[0].OrderCount())       // Two orders
	})

	t.Run("Thread safety - concurrent access", func(t *testing.T) {
		ob := NewOrderbook()

		// Add some initial orders
		for i := 0; i < 5; i++ {
			order := createTestOrder(fmt.Sprintf("seller%d", i), fmt.Sprintf("sell%d", i), 10.0, false)
			ob.PlaceLimitOrder(float64(10_000+i*100), order)
		}

		// Test concurrent reads
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				asks := ob.Asks()
				assert.Equal(t, 5, len(asks))
				// Verify sorting
				for j := 1; j < len(asks); j++ {
					assert.True(t, asks[j-1].Price < asks[j].Price)
				}
			}()
		}
		wg.Wait()
	})
}

// Test 12: Bids() method - returns bid limits sorted by price (descending)
func TestOrderbook_Bids(t *testing.T) {
	ob := NewOrderbook()

	t.Run("Empty orderbook", func(t *testing.T) {
		bids := ob.Bids()
		assert.Equal(t, 0, len(bids))
	})

	t.Run("Single bid limit", func(t *testing.T) {
		buyOrder := createTestOrder("buyer", "buy1", 8.0, true)
		ob.PlaceLimitOrder(9_900, buyOrder)

		bids := ob.Bids()
		assert.Equal(t, 1, len(bids))
		assert.Equal(t, 9_900.0, bids[0].Price)
		assert.Equal(t, 8.0, bids[0].GetTotalVolume())
	})

	t.Run("Multiple bid limits sorted by price", func(t *testing.T) {
		ob := NewOrderbook() // Fresh orderbook

		// Add bids in random price order
		buyOrder1 := createTestOrder("buyer1", "buy1", 5.0, true)
		buyOrder2 := createTestOrder("buyer2", "buy2", 3.0, true)
		buyOrder3 := createTestOrder("buyer3", "buy3", 7.0, true)

		ob.PlaceLimitOrder(9_700, buyOrder1) // Lowest price
		ob.PlaceLimitOrder(9_900, buyOrder2) // Highest price (best bid)
		ob.PlaceLimitOrder(9_800, buyOrder3) // Middle price

		bids := ob.Bids()
		assert.Equal(t, 3, len(bids))

		// Should be sorted by price descending (best bid first)
		assert.Equal(t, 9_900.0, bids[0].Price) // Best bid (highest price)
		assert.Equal(t, 9_800.0, bids[1].Price) // Middle
		assert.Equal(t, 9_700.0, bids[2].Price) // Worst bid (lowest price)

		// Verify volumes
		assert.Equal(t, 3.0, bids[0].GetTotalVolume())
		assert.Equal(t, 7.0, bids[1].GetTotalVolume())
		assert.Equal(t, 5.0, bids[2].GetTotalVolume())
	})

	t.Run("Multiple orders at same price level", func(t *testing.T) {
		ob := NewOrderbook()

		buyOrder1 := createTestOrder("buyer1", "buy1", 5.0, true)
		buyOrder2 := createTestOrder("buyer2", "buy2", 3.0, true)

		ob.PlaceLimitOrder(9_900, buyOrder1)
		ob.PlaceLimitOrder(9_900, buyOrder2) // Same price level

		bids := ob.Bids()
		assert.Equal(t, 1, len(bids)) // Only one limit at this price
		assert.Equal(t, 9_900.0, bids[0].Price)
		assert.Equal(t, 8.0, bids[0].GetTotalVolume()) // Combined volume
		assert.Equal(t, 2, bids[0].OrderCount())       // Two orders
	})
}

// Test 13: Mixed orderbook - both asks and bids
func TestOrderbook_AsksAndBids(t *testing.T) {
	ob := NewOrderbook()

	// Add mixed orders
	sellOrder1 := createTestOrder("seller1", "sell1", 10.0, false)
	sellOrder2 := createTestOrder("seller2", "sell2", 5.0, false)
	buyOrder1 := createTestOrder("buyer1", "buy1", 8.0, true)
	buyOrder2 := createTestOrder("buyer2", "buy2", 3.0, true)

	ob.PlaceLimitOrder(10_100, sellOrder1) // Ask
	ob.PlaceLimitOrder(10_200, sellOrder2) // Higher ask
	ob.PlaceLimitOrder(9_900, buyOrder1)   // Bid
	ob.PlaceLimitOrder(9_800, buyOrder2)   // Lower bid

	// Test asks
	asks := ob.Asks()
	assert.Equal(t, 2, len(asks))
	assert.Equal(t, 10_100.0, asks[0].Price) // Best ask (lowest)
	assert.Equal(t, 10_200.0, asks[1].Price) // Higher ask

	// Test bids
	bids := ob.Bids()
	assert.Equal(t, 2, len(bids))
	assert.Equal(t, 9_900.0, bids[0].Price) // Best bid (highest)
	assert.Equal(t, 9_800.0, bids[1].Price) // Lower bid

	// Verify spread
	bestAsk := asks[0].Price
	bestBid := bids[0].Price
	spread := bestAsk - bestBid
	assert.Equal(t, 200.0, spread) // 10,100 - 9,900 = 200
}

// Test 14: Dynamic updates - asks/bids after market orders
func TestOrderbook_AsksAndBidsAfterMarketOrder(t *testing.T) {
	ob := NewOrderbook()

	// Set up initial book
	sellOrder1 := createTestOrder("seller1", "sell1", 10.0, false)
	sellOrder2 := createTestOrder("seller2", "sell2", 5.0, false)
	ob.PlaceLimitOrder(10_100, sellOrder1)
	ob.PlaceLimitOrder(10_200, sellOrder2)

	// Initial state
	asks := ob.Asks()
	assert.Equal(t, 2, len(asks))

	// Place market order that fully consumes best ask
	buyMarketOrder := createTestOrder("buyer", "buy1", 10.0, true)
	matches, err := ob.PlaceMarketOrder(buyMarketOrder)
	require.NoError(t, err)
	assert.Equal(t, 1, len(matches))

	// Check updated asks
	updatedAsks := ob.Asks()
	assert.Equal(t, 1, len(updatedAsks))            // One limit consumed
	assert.Equal(t, 10_200.0, updatedAsks[0].Price) // Only second ask remains
	assert.Equal(t, 5.0, updatedAsks[0].GetTotalVolume())
}

// Test 15: Volume calculations
func TestOrderbook_AskAndBidTotalVolume(t *testing.T) {
	ob := NewOrderbook()

	// Add various orders
	sellOrder1 := createTestOrder("seller1", "sell1", 10.0, false)
	sellOrder2 := createTestOrder("seller2", "sell2", 5.0, false)
	sellOrder3 := createTestOrder("seller3", "sell3", 3.0, false)
	buyOrder1 := createTestOrder("buyer1", "buy1", 8.0, true)
	buyOrder2 := createTestOrder("buyer2", "buy2", 7.0, true)

	ob.PlaceLimitOrder(10_100, sellOrder1)
	ob.PlaceLimitOrder(10_100, sellOrder2) // Same price level
	ob.PlaceLimitOrder(10_200, sellOrder3)
	ob.PlaceLimitOrder(9_900, buyOrder1)
	ob.PlaceLimitOrder(9_800, buyOrder2)

	// Test total volumes
	assert.Equal(t, 18.0, ob.AskTotalVolume()) // 10 + 5 + 3 = 18
	assert.Equal(t, 15.0, ob.BidTotalVolume()) // 8 + 7 = 15

	// Verify individual limit volumes
	asks := ob.Asks()
	assert.Equal(t, 15.0, asks[0].GetTotalVolume()) // 10 + 5 = 15 at price 10_100
	assert.Equal(t, 3.0, asks[1].GetTotalVolume())  // 3 at price 10_200

	bids := ob.Bids()
	assert.Equal(t, 8.0, bids[0].GetTotalVolume()) // 8 at price 9_900
	assert.Equal(t, 7.0, bids[1].GetTotalVolume()) // 7 at price 9_800
}

// Table-driven test for validation errors
func TestOrderbook_PlaceLimitOrderValidation(t *testing.T) {
	tests := []struct {
		name      string
		price     float64
		order     *orderbookv1.Order
		wantError bool
		errorMsg  string
	}{
		{
			name:      "nil order",
			price:     100.0,
			order:     nil,
			wantError: true,
			errorMsg:  "order cannot be nil",
		},
		{
			name:      "zero price",
			price:     0,
			order:     createTestOrder("user1", "order1", 10.0, false),
			wantError: true,
			errorMsg:  "price must be positive",
		},
		{
			name:      "negative price",
			price:     -100.0,
			order:     createTestOrder("user1", "order1", 10.0, false),
			wantError: true,
			errorMsg:  "price must be positive",
		},
		{
			name:      "zero size order",
			price:     100.0,
			order:     &orderbookv1.Order{ID: "order1", UserID: "user1", Size: 0, Bid: false},
			wantError: true,
			errorMsg:  "order size must be positive",
		},
		{
			name:      "empty order ID",
			price:     100.0,
			order:     &orderbookv1.Order{ID: "", UserID: "user1", Size: 10.0, Bid: false},
			wantError: true,
			errorMsg:  "order ID cannot be empty",
		},
		{
			name:      "valid order",
			price:     100.0,
			order:     createTestOrder("user1", "order1", 10.0, false),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ob := NewOrderbook()
			err := ob.PlaceLimitOrder(tt.price, tt.order)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Table-driven test for market order scenarios
func TestOrderbook_MarketOrderScenarios(t *testing.T) {
	tests := []struct {
		name          string
		setupOrders   func(*Orderbook)
		marketOrder   *orderbookv1.Order
		wantMatches   int
		wantFilled    float64
		wantRemaining float64
	}{
		{
			name: "buy against single ask",
			setupOrders: func(ob *Orderbook) {
				sell := createTestOrder("seller", "sell1", 10.0, false)
				ob.PlaceLimitOrder(10_000, sell)
			},
			marketOrder:   createTestOrder("buyer", "buy1", 5.0, true),
			wantMatches:   1,
			wantFilled:    5.0,
			wantRemaining: 0.0,
		},
		{
			name: "buy across multiple asks",
			setupOrders: func(ob *Orderbook) {
				sell1 := createTestOrder("seller1", "sell1", 5.0, false)
				sell2 := createTestOrder("seller2", "sell2", 3.0, false)
				ob.PlaceLimitOrder(10_000, sell1)
				ob.PlaceLimitOrder(10_100, sell2)
			},
			marketOrder:   createTestOrder("buyer", "buy1", 7.0, true),
			wantMatches:   2,
			wantFilled:    7.0,
			wantRemaining: 0.0,
		},
		{
			name: "sell against single bid",
			setupOrders: func(ob *Orderbook) {
				buy := createTestOrder("buyer", "buy1", 10.0, true)
				ob.PlaceLimitOrder(9_900, buy)
			},
			marketOrder:   createTestOrder("seller", "sell1", 5.0, false),
			wantMatches:   1,
			wantFilled:    5.0,
			wantRemaining: 0.0,
		},
		{
			name: "no matching orders",
			setupOrders: func(ob *Orderbook) {
				// Only add bid, try to place buy market order
				buy := createTestOrder("buyer", "buy1", 10.0, true)
				ob.PlaceLimitOrder(9_900, buy)
			},
			marketOrder:   createTestOrder("buyer2", "buy2", 5.0, true),
			wantMatches:   0,
			wantFilled:    0.0,
			wantRemaining: 5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ob := NewOrderbook()
			tt.setupOrders(ob)

			matches, err := ob.PlaceMarketOrder(tt.marketOrder)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantMatches, len(matches))
			assert.Equal(t, tt.wantRemaining, tt.marketOrder.Size)

			// Calculate total filled from matches
			totalFilled := 0.0
			for _, match := range matches {
				totalFilled += match.SizeFilled
			}
			assert.Equal(t, tt.wantFilled, totalFilled)
		})
	}
}
