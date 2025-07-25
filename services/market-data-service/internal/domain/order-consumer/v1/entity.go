package v1

import (
	"time"
)

// OrderEventType represents the type of order event.
type OrderEventType string

const (
	// OrderPlaced is the event type for a new order being placed.
	OrderPlaced OrderEventType = "order_placed"

	// OrderCancelled is the event type for an order being cancelled.
	OrderCancelled OrderEventType = "order_cancelled"

	// OrderModified is the event type for an order being modified.
	OrderModified OrderEventType = "order_modified"

	// OrderFilled is the event type for an order being filled.
)

// RawOrderEvent represents a raw order event from the matching service.
type RawOrderEvent struct {
	EventID   string         `json:"event_id"`
	Timestamp time.Time      `json:"timestamp"`
	EventType OrderEventType `json:"event_type"` // "order_placed", "order_cancelled", "order_modified", "order_filled"
	OrderID   string         `json:"order_id"`
	Symbol    string         `json:"symbol"`
	Side      string         `json:"side"`
	Price     float64        `json:"price"`
	Quantity  int64          `json:"quantity"`
	OrderType string         `json:"order_type"`
	UserID    string         `json:"user_id"`

	// For modifications
	NewPrice    *float64 `json:"new_price,omitempty"`
	NewQuantity *int64   `json:"new_quantity,omitempty"`

	// For partial fills
	FilledQuantity *int64 `json:"filled_quantity,omitempty"`
	RemainingQty   *int64 `json:"remaining_quantity,omitempty"`
}
