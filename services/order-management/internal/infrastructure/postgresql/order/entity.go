package order

import (
	"time"

	pb "github.com/muhammadchandra19/exchange/proto/go/modules/market-data/v1/shared"
	v1 "github.com/muhammadchandra19/exchange/service/order-management/domain/order-consumer/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Status represents the status of an order.
type Status string

const (
	// OrderStatusPlaced is the status of an order that has been placed.
	OrderStatusPlaced Status = "placed"

	// OrderStatusCancelled is the status of an order that has been cancelled.
	OrderStatusCancelled Status = "cancelled"

	// OrderStatusModified is the status of an order that has been modified.
	OrderStatusModified Status = "modified"
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
	Status    Status    `json:"status"`
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
		Status:    string(o.Status),
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
		o.Status = OrderStatusPlaced
	case v1.OrderCancelled:
		o.Status = OrderStatusCancelled
	case v1.OrderModified:
		o.Status = OrderStatusModified
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

// Event represents a single order event.
type Event struct {
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

// Filter represents the filter criteria for order data.
type Filter struct {
	Symbol        string     `json:"symbol"`
	Side          string     `json:"side"`
	Status        string     `json:"status"`
	UserID        string     `json:"userID"`
	From          *time.Time `json:"from"`
	To            *time.Time `json:"to"`
	Limit         int        `json:"limit"`
	Offset        int        `json:"offset"`
	SortDirection string     `json:"sortDirection"`
}

// BookLevel represents a single order book level.
type BookLevel struct {
	Price    float64 `json:"price"`
	Quantity int64   `json:"quantity"`
	Orders   int64   `json:"orders"` // Number of orders at this level
}

// Book represents a single order book.
type Book struct {
	Symbol    string      `json:"symbol"`
	Timestamp time.Time   `json:"timestamp"`
	Bids      []BookLevel `json:"bids"`
	Asks      []BookLevel `json:"asks"`
}
