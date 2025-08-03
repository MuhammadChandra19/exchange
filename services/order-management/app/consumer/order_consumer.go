package consumer

import (
	"context"

	"github.com/muhammadchandra19/exchange/pkg/logger"
	"github.com/muhammadchandra19/exchange/pkg/postgresql"
	"github.com/muhammadchandra19/exchange/service/order-management/app/consumer"
	"github.com/muhammadchandra19/exchange/service/order-management/domain/order"
	v1 "github.com/muhammadchandra19/exchange/service/order-management/domain/order-consumer/v1"
	orderInfra "github.com/muhammadchandra19/exchange/service/order-management/internal/infrastructure/postgresql/order"
	orderUc "github.com/muhammadchandra19/exchange/service/order-management/internal/usecase/order"
	"github.com/muhammadchandra19/exchange/service/order-management/pkg/config"
)

// Order is the consumer for the order topic.
type Order struct {
	Consumer v1.OrderConsumer
	logger   logger.Interface

	config          config.Config
	orderUsecase    order.Usecase
	orderRepository orderInfra.OrderRepository
	db              postgresql.PostgreSQLClient
}

// InitOrderConsumer initializes the order consumer.
func InitOrderConsumer(ctx context.Context, config config.Config) (*Order, error) {
	logger, err := logger.NewLogger()
	if err != nil {
		return nil, err
	}

	orderConsumer := &Order{
		logger: logger,
		config: config,
	}

	orderConsumer.initDB(ctx)
	orderConsumer.registerRepository()
	orderConsumer.registerUsecase()

	dbTx := postgresql.NewTransaction(orderConsumer.db)

	orderConsumer.Consumer = consumer.NewOrderConsumer(
		config.OrderKafka,
		logger,
		orderConsumer.orderUsecase,
		dbTx,
	)

	return orderConsumer, nil
}

func (o *Order) initDB(ctx context.Context) {
	db, err := postgresql.NewClient(ctx, o.config.PostgreSQL)
	if err != nil {
		o.logger.ErrorContext(ctx, err, logger.Field{
			Key:   "action",
			Value: "init_db",
		})
	}

	o.db = db
}

func (o *Order) registerRepository() {
	o.orderRepository = orderInfra.NewRepository(o.db, o.logger)
}

func (o *Order) registerUsecase() {
	o.orderUsecase = orderUc.NewUsecase(o.orderRepository, o.logger)
}
