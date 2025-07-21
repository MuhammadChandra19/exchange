package orderreaderv1

import (
	"context"

	orderbookv1 "github.com/muhammadchandra19/exchange/services/matching-service/internal/domain/orderbook/v1"
	"github.com/segmentio/kafka-go"
)

// OrderReader defines the interface for reading orders from a source.
//
//go:generate mockgen -source interface.go -destination=mock/interface_mock.go -package=orderreaderv1_mock
type OrderReader interface {
	// ReadMessage reads a message and returns the offset and parsed order
	ReadMessage(ctx context.Context) (kafka.Message, orderbookv1.PlaceOrderRequest, error)
	// SetOffset sets the offset for the reader
	SetOffset(offset int64) error
	// Close closes the reader
	Close() error

	// CommitMessages commits the messages to Kafka after processing
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
}
