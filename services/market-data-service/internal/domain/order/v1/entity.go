package v1

import (
	"time"
)

type Order struct {
	ID        string
	Timestamp time.Time
	Symbol    string
	Side      string // "buy" or "sell"
	Price     float64
	Quantity  int64
	Type      string // "limit", "market", "stop"
	Status    string // "active", "filled", "cancelled", "partial"
	Exchange  string
	UserID    string
}

type OrderEvent struct {
	ID        string
	Timestamp time.Time
	OrderID   string
	EventType string // "placed", "cancelled", "modified", "filled", "partial_fill"
	Symbol    string
	Side      string
	Price     float64
	Quantity  int64
	Exchange  string
	UserID    string
	// For modifications
	NewPrice    *float64
	NewQuantity *int64
}

type OrderFilter struct {
	Symbol   string
	Side     string
	Status   string
	Exchange string
	UserID   string
	From     *time.Time
	To       *time.Time
	Limit    int
	Offset   int
}

// Order book level data
type OrderBookLevel struct {
	Price    float64
	Quantity int64
	Orders   int64 // Number of orders at this level
}

type OrderBook struct {
	Symbol    string
	Timestamp time.Time
	Exchange  string
	Bids      []OrderBookLevel
	Asks      []OrderBookLevel
}
