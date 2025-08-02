package orderbookv1

import (
	"time"

	pb "github.com/muhammadchandra19/exchange/proto/go/kafka/v1"
)

// OrderType represents the type of order.
type OrderType string

const (
	// OrderTypeMarket represents a market order.
	OrderTypeMarket OrderType = "market"
	// OrderTypeLimit represents a limit order.
	OrderTypeLimit OrderType = "limit"
	// OrderTypeCancel represents a cancel order.
	OrderTypeCancel OrderType = "cancel"
)

// Order represents a single order in the order book.
type Order struct {
	ID        string  `json:"id"`
	UserID    string  `json:"userID"`
	Size      float64 `json:"size"`
	Bid       bool    `json:"bid"`
	Limit     *Limit  `json:"-"`
	Timestamp int64   `json:"timestamp"`
	Sequence  int64   `json:"sequence"` // Sequence number for the order
}

// PlaceOrderRequest represents a request to place an order in the order book.
type PlaceOrderRequest struct {
	OrderID string    `json:"orderID"`
	UserID  string    `json:"userID"`
	Type    OrderType `json:"type"`
	Bid     bool      `json:"bid"`
	Size    float64   `json:"size"`
	Price   float64   `json:"price"`
	Offset  int64     // Offset for the order in the stream
}

// FromKafkaPayload converts a Kafka payload to a PlaceOrderRequest.
func (r *PlaceOrderRequest) FromKafkaPayload(payload *pb.PlaceOrderPayload) *PlaceOrderRequest {
	return &PlaceOrderRequest{
		OrderID: payload.OrderID,
		UserID:  payload.UserID,
		Type:    OrderType(payload.Type),
		Bid:     payload.Bid,
		Size:    payload.Size,
		Price:   payload.Price,
		Offset:  payload.Offset,
	}
}

// NewOrder creates a new order with the given parameters.
func NewOrder(userID string, size float64, bid bool, id string) *Order {
	return &Order{
		ID:        id,
		UserID:    userID,
		Size:      size,
		Bid:       bid,
		Timestamp: time.Now().UnixNano(),
	}
}

// IsBid checks if the order is a bid (buy) order.
func (o *Order) IsBid() bool {
	return o.Bid
}

// IsAsk checks if the order is an ask (sell) order.
func (o *Order) IsAsk() bool {
	return !o.Bid
}

// IsFilled checks if the order is filled (size is zero).
func (o *Order) IsFilled() bool {
	return o.Size == 0.0
}

// NextSequence increments the order's sequence number and returns the new value.
func (o *Order) NextSequence() int64 {
	o.Sequence++
	return o.Sequence
}
