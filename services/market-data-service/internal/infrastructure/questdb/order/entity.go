package order

import (
	"time"
)

// Order represents a single order.
type Order struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Symbol    string    `json:"symbol"`
	Side      string    `json:"side"`
	Price     float64   `json:"price"`
	Quantity  int64     `json:"quantity"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	UserID    string    `json:"userID"`
}

// OrderEvent represents a single order event.
type OrderEvent struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	OrderID   string    `json:"orderID"`
	EventType string    `json:"eventType"` // "placed", "cancelled", "modified", "filled", "partial_fill"
	Symbol    string    `json:"symbol"`
	Side      string    `json:"side"`
	Price     float64   `json:"price"`
	Quantity  int64     `json:"quantity"`
	UserID    string    `json:"userID"`
	// For modifications
	NewPrice    *float64 `json:"newPrice"`
	NewQuantity *int64   `json:"newQuantity"`
}

// OrderFilter represents the filter criteria for order data.
type OrderFilter struct {
	Symbol string     `json:"symbol"`
	Side   string     `json:"side"`
	Status string     `json:"status"`
	UserID string     `json:"userID"`
	From   *time.Time `json:"from"`
	To     *time.Time `json:"to"`
	Limit  int        `json:"limit"`
	Offset int        `json:"offset"`
}

// OrderBookLevel represents a single order book level.
type OrderBookLevel struct {
	Price    float64 `json:"price"`
	Quantity int64   `json:"quantity"`
	Orders   int64   `json:"orders"` // Number of orders at this level
}

// OrderBook represents a single order book.
type OrderBook struct {
	Symbol    string           `json:"symbol"`
	Timestamp time.Time        `json:"timestamp"`
	Bids      []OrderBookLevel `json:"bids"`
	Asks      []OrderBookLevel `json:"asks"`
}
