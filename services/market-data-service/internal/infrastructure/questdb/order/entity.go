package order

import (
	"time"

	pb "github.com/muhammadchandra19/exchange/proto/go/modules/market-data-service/v1/shared"
	v1 "github.com/muhammadchandra19/exchange/services/market-data-service/internal/domain/order-consumer/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderStatus string

const (
	// OrderStatusPlaced is the status of an order that has been placed.
	OrderStatusPlaced OrderStatus = "placed"

	// OrderStatusCancelled is the status of an order that has been cancelled.
	OrderStatusCancelled OrderStatus = "cancelled"

	// OrderStatusModified is the status of an order that has been modified.
	OrderStatusModified OrderStatus = "modified"
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

// ToProto converts the order to a protobuf message.
func (o *Order) ToProto() *pb.Order {
	if o == nil {
		return nil
	}

	return &pb.Order{
		Id:        o.ID,
		Symbol:    o.Symbol,
		Side:      o.Side,
		Price:     o.Price,
		Quantity:  o.Quantity,
		Type:      o.Type,
		Status:    o.Status,
		Timestamp: timestamppb.New(o.Timestamp).AsTime().Format(time.RFC3339),
		UserID:    o.UserID,
	}
}

// FromOrderEvent converts a raw order event to an order.
func (o *Order) FromOrderEvent(orderEvent *v1.RawOrderEvent) {
	o.ID = orderEvent.OrderID
	o.Symbol = orderEvent.Symbol
	o.Side = orderEvent.Side
	o.Price = orderEvent.Price
	o.Quantity = orderEvent.Quantity
	o.Type = orderEvent.OrderType
	o.Timestamp = orderEvent.Timestamp
	o.UserID = orderEvent.UserID

	switch orderEvent.EventType {
	case v1.OrderPlaced:
		o.Status = string(OrderStatusPlaced)
	case v1.OrderCancelled:
		o.Status = string(OrderStatusCancelled)
	case v1.OrderModified:
		o.Status = string(OrderStatusModified)
	}
}

// CheckOrderDiff checks the difference between the order and the order event.
func (o *Order) CheckOrderDiff(orderEvent *v1.RawOrderEvent) map[string]interface{} {
	diff := make(map[string]interface{})
	if o.Price != orderEvent.Price {
		diff["price"] = orderEvent.Price
	}

	if o.Quantity != orderEvent.Quantity {
		diff["quantity"] = orderEvent.Quantity
	}

	return diff
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
