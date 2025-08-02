package orderbookv1

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

var (
	ErrNilOrder      = errors.New("order cannot be nil")
	ErrInvalidPrice  = errors.New("price must be positive")
	ErrInvalidSize   = errors.New("size must be positive")
	ErrOrderNotFound = errors.New("order not found in limit")
)

// Limit represents a price level in the order book with associated orders.
type Limit struct {
	Price       float64  `json:"price"`
	Orders      []*Order `json:"orders"`
	TotalVolume float64  `json:"totalVolume"`
	mu          sync.RWMutex
}

// NewLimit creates a new Limit with the specified price.
func NewLimit(price float64) *Limit {
	return &Limit{
		Price:       price,
		Orders:      make([]*Order, 0),
		TotalVolume: 0.0,
	}
}

// AddOrder adds an order to the limit and updates the total volume.
func (l *Limit) AddOrder(order *Order) error {
	if order == nil {
		return ErrNilOrder
	}
	if order.Size <= 0 {
		return fmt.Errorf("%w: got %f", ErrInvalidSize, order.Size)
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	order.Limit = l
	l.Orders = append(l.Orders, order)
	l.TotalVolume += order.Size

	return nil
}

// RemoveOrder removes an order from the limit and updates the total volume.
func (l *Limit) RemoveOrder(order *Order) error {
	if order == nil {
		return ErrNilOrder
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	for i, o := range l.Orders {
		if o == order {
			// Remove order from slice
			l.Orders = append(l.Orders[:i], l.Orders[i+1:]...)
			l.TotalVolume -= order.Size
			order.Limit = nil
			return nil
		}
	}

	return ErrOrderNotFound
}

// Fill matches the limit with an incoming order and returns matches.
func (l *Limit) Fill(incomingOrder *Order) []Match {
	if incomingOrder == nil {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	var matches []Match

	// Sort orders by timestamp (FIFO), then by sequence for ties
	ordersToProcess := make([]*Order, len(l.Orders))
	copy(ordersToProcess, l.Orders)
	sort.Slice(ordersToProcess, func(i, j int) bool {
		if ordersToProcess[i].Timestamp == ordersToProcess[j].Timestamp {
			return ordersToProcess[i].Sequence < ordersToProcess[j].Sequence
		}
		return ordersToProcess[i].Timestamp < ordersToProcess[j].Timestamp
	})

	var ordersToRemove []*Order

	// Process orders in FIFO order
	for _, existingOrder := range ordersToProcess {
		if incomingOrder.Size <= 0 {
			break
		}

		// Create a match
		match := l.createMatch(incomingOrder, existingOrder)
		matches = append(matches, match)

		// Update total volume
		l.TotalVolume -= match.SizeFilled

		// Mark filled orders for removal
		if existingOrder.Size <= 0 {
			ordersToRemove = append(ordersToRemove, existingOrder)
		}
	}

	// Remove filled orders
	for _, orderToRemove := range ordersToRemove {
		l.removeOrderUnsafe(orderToRemove)
	}

	return matches
}

// createMatch creates a match between incoming and existing order
func (l *Limit) createMatch(incomingOrder, existingOrder *Order) Match {
	var bid, ask *Order
	var sizeFilled float64

	// Determine which is bid and which is ask
	if incomingOrder.IsBid() {
		bid = incomingOrder
		ask = existingOrder
	} else {
		bid = existingOrder
		ask = incomingOrder
	}

	// Calculate size filled
	if incomingOrder.Size >= existingOrder.Size {
		sizeFilled = existingOrder.Size
		incomingOrder.Size -= existingOrder.Size
		existingOrder.Size = 0
	} else {
		sizeFilled = incomingOrder.Size
		existingOrder.Size -= incomingOrder.Size
		incomingOrder.Size = 0
	}

	return Match{
		Ask:        ask,
		Bid:        bid,
		SizeFilled: sizeFilled,
		Price:      l.Price,
	}
}

// removeOrderUnsafe removes order without locking (internal use)
func (l *Limit) removeOrderUnsafe(order *Order) {
	for i, o := range l.Orders {
		if o == order {
			l.Orders = append(l.Orders[:i], l.Orders[i+1:]...)
			order.Limit = nil
			break
		}
	}
}

// IsEmpty checks if the limit has no orders
func (l *Limit) IsEmpty() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.Orders) == 0
}

// OrderCount returns the number of orders at this limit
func (l *Limit) OrderCount() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.Orders)
}

// GetPrice returns the price of this limit
func (l *Limit) GetPrice() float64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.Price
}

// GetTotalVolume returns the total volume at this limit
func (l *Limit) GetTotalVolume() float64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.TotalVolume
}

// GetOrders returns a copy of the orders slice
func (l *Limit) GetOrders() []*Order {
	l.mu.RLock()
	defer l.mu.RUnlock()

	orders := make([]*Order, len(l.Orders))
	copy(orders, l.Orders)
	return orders
}

// GetOrdersByPriority returns orders sorted by timestamp then sequence
func (l *Limit) GetOrdersByPriority() []*Order {
	l.mu.RLock()
	defer l.mu.RUnlock()

	orders := make([]*Order, len(l.Orders))
	copy(orders, l.Orders)

	sort.Slice(orders, func(i, j int) bool {
		if orders[i].Timestamp == orders[j].Timestamp {
			return orders[i].Sequence < orders[j].Sequence
		}
		return orders[i].Timestamp < orders[j].Timestamp
	})

	return orders
}

// Validate performs basic validation of the limit's state
func (l *Limit) Validate() error {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.Price <= 0 {
		return fmt.Errorf("%w: limit price %f", ErrInvalidPrice, l.Price)
	}

	calculatedVolume := 0.0
	for _, order := range l.Orders {
		if order == nil {
			return fmt.Errorf("nil order found in limit")
		}
		if order.Size < 0 {
			return fmt.Errorf("%w: order has size %f", ErrInvalidSize, order.Size)
		}
		calculatedVolume += order.Size
	}

	// Check volume consistency (with small tolerance for floating point)
	const epsilon = 1e-9
	if abs(calculatedVolume-l.TotalVolume) > epsilon {
		return fmt.Errorf("volume mismatch: calculated %f, stored %f", calculatedVolume, l.TotalVolume)
	}

	return nil
}

// Helper function for absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
