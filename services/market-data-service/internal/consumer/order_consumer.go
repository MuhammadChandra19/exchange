package consumer

import (
	"context"
	"encoding/json"

	"github.com/muhammadchandra19/exchange/pkg/logger"
	"github.com/muhammadchandra19/exchange/services/market-data-service/internal/domain/order"
	v1 "github.com/muhammadchandra19/exchange/services/market-data-service/internal/domain/order-consumer/v1"
	orderInfra "github.com/muhammadchandra19/exchange/services/market-data-service/internal/infrastructure/questdb/order"
	"github.com/muhammadchandra19/exchange/services/market-data-service/pkg/config"
	"github.com/segmentio/kafka-go"
)

// OrderConsumer is the consumer for the order topic.
type OrderConsumer struct {
	kafkaReader *kafka.Reader
	logger      logger.Interface

	orderUsecase order.Usecase
	msgChan      chan kafka.Message
}

// NewOrderConsumer creates a new OrderConsumer.
func NewOrderConsumer(config config.OrderKafkaConfig, logger logger.Interface, orderUsecase order.Usecase) *OrderConsumer {
	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     config.Brokers,
		Topic:       config.Topic,
		GroupID:     config.ConsumerGroup,
		MinBytes:    1,
		MaxBytes:    10e6,
		StartOffset: kafka.LastOffset,
	})
	return &OrderConsumer{
		kafkaReader:  kafkaReader,
		logger:       logger,
		orderUsecase: orderUsecase,
		msgChan:      make(chan kafka.Message),
	}
}

// Start starts the OrderConsumer.
func (c *OrderConsumer) Start(ctx context.Context) {
	for {
		msg, err := c.kafkaReader.ReadMessage(ctx)
		if err != nil {
			c.logger.ErrorContext(ctx, err, logger.Field{
				Key:   "action",
				Value: "read_message",
			})
			continue
		}

		c.msgChan <- msg
	}
}

// Stop stops the OrderConsumer.
func (c *OrderConsumer) Stop() error {
	return c.kafkaReader.Close()
}

// Subscribe subscribes to the OrderConsumer.
func (c *OrderConsumer) Subscribe(ctx context.Context) {
	for msg := range c.msgChan {
		var order v1.RawOrderEvent
		if err := json.Unmarshal(msg.Value, &order); err != nil {
			c.logger.ErrorContext(ctx, err, logger.Field{
				Key:   "action",
				Value: "unmarshal_order",
			})
		}

		if err := c.handleOrder(ctx, &order); err != nil {
			c.logger.ErrorContext(ctx, err, logger.Field{
				Key:   "action",
				Value: "handle_order",
			})
		}

		if err := c.kafkaReader.CommitMessages(ctx, msg); err != nil {
			c.logger.ErrorContext(ctx, err, logger.Field{
				Key:   "action",
				Value: "commit_message",
			})
		}
	}
}

func (c *OrderConsumer) handleOrder(ctx context.Context, orderEvent *v1.RawOrderEvent) error {
	var order *orderInfra.Order

	order.FromOrderEvent(orderEvent)
	err := c.orderUsecase.StoreOrder(ctx, order)
	if err != nil {
		c.logger.ErrorContext(ctx, err, logger.Field{
			Key:   "action",
			Value: "store_order",
		})
		return err
	}

	return nil
}
