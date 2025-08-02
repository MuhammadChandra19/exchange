package orderreader

import (
	"context"
	"encoding/json"

	"github.com/muhammadchandra19/exchange/pkg/logger"
	pb "github.com/muhammadchandra19/exchange/proto/go/kafka/v1"
	"github.com/muhammadchandra19/exchange/services/matching-engine/pkg/config"
	"github.com/segmentio/kafka-go"
)

// Reader represents a Kafka Reader for consuming messages from the order topic.
type Reader struct {
	kafkaReader *kafka.Reader
	logger      logger.Logger
}

// NewReader creates a new Kafka reader for consuming messages from the order topic.
// It returns an implementation of the OrderReader interface.
func NewReader(config config.KafkaConfig, log logger.Logger) Reader {
	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: config.Brokers,
		Topic:   config.Topic,
		// GroupID:     config.GroupID,
		Partition:   0,
		MinBytes:    1,
		MaxBytes:    10e6,
		StartOffset: kafka.LastOffset,
	})

	return Reader{
		kafkaReader: kafkaReader,
		logger:      log,
	}
}

// logError is a helper method to log errors consistently
func (r Reader) logError(err error, operation string) {
	r.logger.Error(err,
		logger.Field{Key: "error", Value: err.Error()},
		logger.Field{Key: "operation", Value: operation},
	)
}

// SetOffset sets the offset for the Kafka reader.
func (r Reader) SetOffset(offset int64) error {
	if err := r.kafkaReader.SetOffset(offset); err != nil {
		r.logError(err, "SetOffset")
		return err
	}
	return nil
}

// ReadMessage reads a message from the Kafka topic and parses it as an Order.
func (r Reader) ReadMessage(ctx context.Context) (kafka.Message, *pb.PlaceOrderPayload, error) {
	msg, err := r.kafkaReader.ReadMessage(ctx)
	if err != nil {
		r.logError(err, "ReadMessage")
		return kafka.Message{Offset: 0}, nil, err
	}

	var order pb.PlaceOrderPayload
	if err := json.Unmarshal(msg.Value, &order); err != nil {
		r.logError(err, "UnmarshalOrder")
		return kafka.Message{Offset: 0}, nil, err
	}

	r.logger.Info("ReadMessage",
		logger.Field{
			Key:   "userID",
			Value: order.UserID,
		},
		logger.Field{
			Key:   "size",
			Value: order.Size,
		},
		logger.Field{
			Key:   "price",
			Value: order.Price,
		},
		logger.Field{
			Key:   "type",
			Value: order.Type,
		},
		logger.Field{
			Key:   "bid",
			Value: order.Bid,
		},
	)

	order.Offset = msg.Offset // Set the offset in the order request

	return msg, &order, nil
}

// Close properly closes the Kafka reader.
func (r Reader) Close() error {
	if err := r.kafkaReader.Close(); err != nil {
		r.logError(err, "Close")
		return err
	}
	return nil
}

// CommitMessages commits the messages to Kafka after processing.
func (r Reader) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	// if err := r.kafkaReader.CommitMessages(ctx, msgs...); err != nil {
	// 	r.logError(err, "CommitMessages")
	// 	return err
	// }
	return nil
}
