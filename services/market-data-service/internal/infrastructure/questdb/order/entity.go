package order

import (
	"time"
)

// Order represents a single order.
type Order struct {
	ID        string
	Timestamp time.Time
	Symbol    string
	Side      string // "buy" or "sell"
	Price     float64
	Quantity  int64
	Type      string // "limit", "market", "stop"
	Status    string // "active", "filled", "cancelled", "partial"
	UserID    string
}

// OrderEvent represents a single order event.
type OrderEvent struct {
	ID        string
	Timestamp time.Time
	OrderID   string
	EventType string // "placed", "cancelled", "modified", "filled", "partial_fill"
	Symbol    string
	Side      string
	Price     float64
	Quantity  int64
	UserID    string
	// For modifications
	NewPrice    *float64
	NewQuantity *int64
}

// OrderFilter represents the filter criteria for order data.
type OrderFilter struct {
	Symbol string
	Side   string
	Status string
	UserID string
	From   *time.Time
	To     *time.Time
	Limit  int
	Offset int
}

// OrderBookLevel represents a single order book level.
type OrderBookLevel struct {
	Price    float64
	Quantity int64
	Orders   int64 // Number of orders at this level
}

// OrderBook represents a single order book.
type OrderBook struct {
	Symbol    string
	Timestamp time.Time
	Bids      []OrderBookLevel
	Asks      []OrderBookLevel
}
