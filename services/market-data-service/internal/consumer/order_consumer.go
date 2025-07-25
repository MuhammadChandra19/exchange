package consumer

import (
	"context"
	"encoding/json"

	"github.com/muhammadchandra19/exchange/pkg/logger"
	"github.com/muhammadchandra19/exchange/pkg/questdb"
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
	dbTx         questdb.TX
	msgChan      chan kafka.Message
}

// NewOrderConsumer creates a new OrderConsumer.
func NewOrderConsumer(config config.OrderKafkaConfig, logger logger.Interface, orderUsecase order.Usecase, dbTx questdb.TX) *OrderConsumer {
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
		dbTx:         dbTx,
		msgChan:      make(chan kafka.Message),
	}
}

// Start starts the OrderConsumer.
func (c *OrderConsumer) Start(ctx context.Context) {
	c.logger.InfoContext(ctx, "starting order consumer", logger.Field{
		Key:   "action",
		Value: "order_consumer_start",
	})

	for {
		select {
		case <-ctx.Done():
			c.logger.InfoContext(ctx, "context done", logger.Field{
				Key:   "action",
				Value: "order_consumer_stop",
			})
			return
		default:
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
}

// Stop stops the OrderConsumer.
func (c *OrderConsumer) Stop() error {
	c.logger.InfoContext(context.Background(), "stopping order consumer", logger.Field{
		Key:   "action",
		Value: "order_consumer_stop",
	})
	return c.kafkaReader.Close()
}

// Subscribe subscribes to the OrderConsumer.
func (c *OrderConsumer) Subscribe(ctx context.Context) {
	c.logger.InfoContext(ctx, "subscribing to order consumer", logger.Field{
		Key:   "action",
		Value: "order_consumer_subscribe",
	})

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
	txCtx, err := c.dbTx.Begin(ctx)
	if err != nil {
		return err
	}

	defer c.dbTx.Rollback(txCtx)

	order := &orderInfra.Order{}

	order.FromOrderEvent(orderEvent)
	err = c.orderUsecase.StoreOrder(txCtx, order)
	if err != nil {
		c.logger.ErrorContext(ctx, err, logger.Field{
			Key:   "action",
			Value: "store_order",
		})
		return err
	}

	err = c.orderUsecase.StoreOrderEvent(txCtx, &orderInfra.OrderEvent{
		ID:          orderEvent.OrderID,
		Timestamp:   orderEvent.Timestamp,
		OrderID:     orderEvent.OrderID,
		EventType:   orderEvent.EventType,
		Symbol:      orderEvent.Symbol,
		Side:        orderEvent.Side,
		Price:       orderEvent.Price,
		Quantity:    orderEvent.Quantity,
		UserID:      orderEvent.UserID,
		NewPrice:    orderEvent.NewPrice,
		NewQuantity: orderEvent.NewQuantity,
	})
	if err != nil {
		c.logger.ErrorContext(ctx, err, logger.Field{
			Key:   "action",
			Value: "store_order_event",
		})
		return err
	}

	if err := c.dbTx.Commit(txCtx); err != nil {
		c.logger.ErrorContext(ctx, err, logger.Field{
			Key:   "action",
			Value: "commit_transaction",
		})

		return err
	}

	return nil
}
