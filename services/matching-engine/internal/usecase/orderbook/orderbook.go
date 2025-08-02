package orderbook

import (
	"fmt"
	"sort"
	"sync"

	orderbookv1 "github.com/muhammadchandra19/exchange/services/matching-engine/internal/domain/orderbook/v1"
	snapshotv1 "github.com/muhammadchandra19/exchange/services/matching-engine/internal/domain/snapshot/v1"
)

// Orderbook represents a simple order book
type Orderbook struct {
	mu        sync.RWMutex
	AskLimits map[float64]*orderbookv1.Limit // price -> limit
	BidLimits map[float64]*orderbookv1.Limit // price -> limit
	Orders    map[string]*orderbookv1.Order  // orderID -> order
}

// NewOrderbook creates a new orderbook
func NewOrderbook() *Orderbook {
	return &Orderbook{
		AskLimits: make(map[float64]*orderbookv1.Limit),
		BidLimits: make(map[float64]*orderbookv1.Limit),
		Orders:    make(map[string]*orderbookv1.Order),
	}
}

// PlaceLimitOrder places a limit order
func (ob *Orderbook) PlaceLimitOrder(price float64, order *orderbookv1.Order) error {
	if order == nil {
		return fmt.Errorf("order cannot be nil")
	}
	if price <= 0 {
		return fmt.Errorf("price must be positive")
	}
	if order.Size <= 0 {
		return fmt.Errorf("order size must be positive")
	}
	if order.ID == "" {
		return fmt.Errorf("order ID cannot be empty")
	}

	ob.mu.Lock()
	defer ob.mu.Unlock()

	// Check if order already exists
	if _, exists := ob.Orders[order.ID]; exists {
		return fmt.Errorf("order with ID %s already exists", order.ID)
	}

	// Find or create limit
	var limits map[float64]*orderbookv1.Limit
	if order.IsBid() {
		limits = ob.BidLimits
	} else {
		limits = ob.AskLimits
	}

	limit, exists := limits[price]
	if !exists {
		limit = orderbookv1.NewLimit(price)
		limits[price] = limit
	}

	// Add order to limit
	if err := limit.AddOrder(order); err != nil {
		return err
	}

	// Add to orders map
	ob.Orders[order.ID] = order

	return nil
}

// PlaceMarketOrder places a market order and returns matches
func (ob *Orderbook) PlaceMarketOrder(order *orderbookv1.Order) ([]orderbookv1.Match, error) {
	if order == nil {
		return nil, fmt.Errorf("order cannot be nil")
	}

	ob.mu.Lock()
	defer ob.mu.Unlock()

	var matches []orderbookv1.Match
	var limits []*orderbookv1.Limit

	// Get limits in price priority order
	if order.IsBid() {
		// Buy order: match against asks (lowest price first)
		for price := range ob.AskLimits {
			limits = append(limits, ob.AskLimits[price])
		}
		sort.Slice(limits, func(i, j int) bool {
			return limits[i].Price < limits[j].Price
		})
	} else {
		// Sell order: match against bids (highest price first)
		for price := range ob.BidLimits {
			limits = append(limits, ob.BidLimits[price])
		}
		sort.Slice(limits, func(i, j int) bool {
			return limits[i].Price > limits[j].Price
		})
	}

	// Process limits until order is filled
	for _, limit := range limits {
		if order.Size <= 0 {
			break
		}

		limitMatches := limit.Fill(order)
		matches = append(matches, limitMatches...)

		// Remove empty limits
		if limit.IsEmpty() {
			if order.IsBid() {
				delete(ob.AskLimits, limit.Price)
			} else {
				delete(ob.BidLimits, limit.Price)
			}
		}
	}

	return matches, nil
}

// CancelOrder removes an order
func (ob *Orderbook) CancelOrder(orderID string) error {
	if orderID == "" {
		return fmt.Errorf("order ID cannot be empty")
	}

	ob.mu.Lock()
	defer ob.mu.Unlock()

	order, exists := ob.Orders[orderID]
	if !exists {
		return fmt.Errorf("order with ID %s does not exist", orderID)
	}

	// Store limit reference before removing order (since RemoveOrder sets order.Limit to nil)
	limit := order.Limit

	// Remove from limit
	if limit != nil {
		if err := limit.RemoveOrder(order); err != nil {
			return err
		}

		// Remove empty limit (use stored reference since order.Limit is now nil)
		if limit.IsEmpty() {
			if order.IsBid() {
				delete(ob.BidLimits, limit.Price)
			} else {
				delete(ob.AskLimits, limit.Price)
			}
		}
	}

	// Remove from orders map
	delete(ob.Orders, orderID)

	return nil
}

// Asks returns ask limits sorted by price (ascending)
func (ob *Orderbook) Asks() []*orderbookv1.Limit {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	var limits []*orderbookv1.Limit
	for _, limit := range ob.AskLimits {
		limits = append(limits, limit)
	}
	sort.Slice(limits, func(i, j int) bool {
		return limits[i].Price < limits[j].Price
	})
	return limits
}

// Bids returns bid limits sorted by price (descending)
func (ob *Orderbook) Bids() []*orderbookv1.Limit {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	var limits []*orderbookv1.Limit
	for _, limit := range ob.BidLimits {
		limits = append(limits, limit)
	}
	sort.Slice(limits, func(i, j int) bool {
		return limits[i].Price > limits[j].Price
	})
	return limits
}

// AskTotalVolume returns total ask volume
func (ob *Orderbook) AskTotalVolume() float64 {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	total := 0.0
	for _, limit := range ob.AskLimits {
		total += limit.GetTotalVolume()
	}
	return total
}

// BidTotalVolume returns total bid volume
func (ob *Orderbook) BidTotalVolume() float64 {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	total := 0.0
	for _, limit := range ob.BidLimits {
		total += limit.GetTotalVolume()
	}
	return total
}

// CreateSnapshot creates a snapshot of the current orderbook state
func (ob *Orderbook) CreateSnapshot() *snapshotv1.Snapshot {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	var bookOrders []snapshotv1.BookOrder

	// Collect all orders from all limits
	for _, limit := range ob.AskLimits {
		orders := limit.GetOrders()
		for _, order := range orders {
			bookOrders = append(bookOrders, snapshotv1.BookOrder{
				OrderID:   order.ID,
				Size:      order.Size,
				Bid:       order.Bid,
				Price:     limit.Price,
				UserID:    order.UserID,
				Timestamp: order.Timestamp,
			})
		}
	}

	for _, limit := range ob.BidLimits {
		orders := limit.GetOrders()
		for _, order := range orders {
			bookOrders = append(bookOrders, snapshotv1.BookOrder{
				OrderID:   order.ID,
				Size:      order.Size,
				Bid:       order.Bid,
				Price:     limit.Price,
				UserID:    order.UserID,
				Timestamp: order.Timestamp,
			})
		}
	}

	return &snapshotv1.Snapshot{
		OrderOffset: 0, // This will be set by the engine
		OrderBookSnapshot: snapshotv1.OrderBookSnapshot{
			Orders:        bookOrders,
			TradeSequence: 0, // Not used in simplified version
			LogSequence:   0, // Not used in simplified version
		},
	}
}

// RestoreOrderbook restores the orderbook from a snapshot
func (ob *Orderbook) RestoreOrderbook(snapshot *snapshotv1.Snapshot) error {
	if snapshot == nil {
		return fmt.Errorf("snapshot cannot be nil")
	}

	ob.mu.Lock()
	defer ob.mu.Unlock()

	// Clear current state
	ob.AskLimits = make(map[float64]*orderbookv1.Limit)
	ob.BidLimits = make(map[float64]*orderbookv1.Limit)
	ob.Orders = make(map[string]*orderbookv1.Order)

	// Restore orders from snapshot
	for _, bookOrder := range snapshot.OrderBookSnapshot.Orders {
		// Create the order
		order := &orderbookv1.Order{
			ID:        bookOrder.OrderID,
			UserID:    bookOrder.UserID,
			Size:      bookOrder.Size,
			Bid:       bookOrder.Bid,
			Timestamp: bookOrder.Timestamp,
		}

		// Find or create the appropriate limit
		var limits map[float64]*orderbookv1.Limit
		if order.IsBid() {
			limits = ob.BidLimits
		} else {
			limits = ob.AskLimits
		}

		limit, exists := limits[bookOrder.Price]
		if !exists {
			limit = orderbookv1.NewLimit(bookOrder.Price)
			limits[bookOrder.Price] = limit
		}

		// Add order to limit
		if err := limit.AddOrder(order); err != nil {
			return fmt.Errorf("failed to restore order %s: %w", bookOrder.OrderID, err)
		}

		// Add to orders map
		ob.Orders[order.ID] = order
	}

	return nil
}
