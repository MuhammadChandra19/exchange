package consumer

import (
	"context"

	"github.com/muhammadchandra19/exchange/pkg/logger"
	"github.com/muhammadchandra19/exchange/pkg/questdb"
	"github.com/muhammadchandra19/exchange/services/market-data-service/internal/bootstrap"
	"github.com/muhammadchandra19/exchange/services/market-data-service/internal/consumer"
	v1 "github.com/muhammadchandra19/exchange/services/market-data-service/internal/domain/order-consumer/v1"
	"github.com/muhammadchandra19/exchange/services/market-data-service/pkg/config"

	orderInfra "github.com/muhammadchandra19/exchange/services/market-data-service/internal/infrastructure/questdb/order"
	tickInfra "github.com/muhammadchandra19/exchange/services/market-data-service/internal/infrastructure/questdb/tick"

	orderUc "github.com/muhammadchandra19/exchange/services/market-data-service/internal/usecase/order"
	tickUc "github.com/muhammadchandra19/exchange/services/market-data-service/internal/usecase/tick"
)

// OrderConsumer is the consumer for the order topic.
type OrderConsumer struct {
	Consumer   v1.OrderConsumer
	logger     logger.Interface
	Config     config.Config
	usecase    bootstrap.Usecase
	repository bootstrap.Repository
	db         questdb.QuestDBClient
}

// InitOrderConsumer creates a new OrderConsumer.
func InitOrderConsumer(ctx context.Context, config config.Config) (*OrderConsumer, error) {
	logger, err := logger.NewLogger()
	if err != nil {
		return nil, err
	}

	orderConsumer := &OrderConsumer{
		logger:     logger,
		Config:     config,
		usecase:    bootstrap.Usecase{},
		repository: bootstrap.Repository{},
	}

	orderConsumer.initDB(ctx)
	orderConsumer.registerRepository()
	orderConsumer.registerUsecase()

	dbTx := questdb.NewTransaction(orderConsumer.db)

	orderConsumer.Consumer = consumer.NewOrderConsumer(
		config.OrderKafka,
		logger,
		orderConsumer.usecase.OrderUsecase,
		dbTx,
	)

	return orderConsumer, nil
}

func (s *OrderConsumer) initDB(ctx context.Context) {
	questdbClient, err := questdb.NewClient(ctx, s.Config.QuestDB)
	if err != nil {
		s.logger.ErrorContext(ctx, err, logger.Field{
			Key:   "action",
			Value: "init_db",
		})
		return
	}

	s.db = questdbClient
}

func (s *OrderConsumer) registerRepository() {
	s.repository.OrderRepository = orderInfra.NewRepository(s.db)
	s.repository.TickRepository = tickInfra.NewRepository(s.db)
}

func (s *OrderConsumer) registerUsecase() {
	s.usecase.OrderUsecase = orderUc.NewUsecase(s.repository.OrderRepository, s.logger)
	s.usecase.TickUsecase = tickUc.NewUsecase(s.repository.TickRepository, s.logger)
}
