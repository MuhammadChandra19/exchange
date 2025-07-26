package consumer

import (
	"context"

	"github.com/muhammadchandra19/exchange/pkg/logger"
	"github.com/muhammadchandra19/exchange/pkg/questdb"
	"github.com/muhammadchandra19/exchange/services/market-data-service/internal/bootstrap"
	"github.com/muhammadchandra19/exchange/services/market-data-service/internal/consumer"
	v1 "github.com/muhammadchandra19/exchange/services/market-data-service/internal/domain/match-consumer/v1"
	ohlcUc "github.com/muhammadchandra19/exchange/services/market-data-service/internal/usecase/ohlc"
	tickUc "github.com/muhammadchandra19/exchange/services/market-data-service/internal/usecase/tick"
	"github.com/muhammadchandra19/exchange/services/market-data-service/pkg/config"

	ohlcInfra "github.com/muhammadchandra19/exchange/services/market-data-service/internal/infrastructure/questdb/ohlc"
	tickInfra "github.com/muhammadchandra19/exchange/services/market-data-service/internal/infrastructure/questdb/tick"
)

// MatchConsumer is the consumer for the match topic.
type MatchConsumer struct {
	Consumer   v1.MatchConsumer
	logger     logger.Interface
	Config     config.Config
	usecase    bootstrap.Usecase
	repository bootstrap.Repository
	db         questdb.QuestDBClient
	dbTx       questdb.Transaction
}

// InitMatchConsumer creates a new MatchConsumer.
func InitMatchConsumer(ctx context.Context, config config.Config) (*MatchConsumer, error) {
	logger, err := logger.NewLogger()
	if err != nil {
		return nil, err
	}

	matchConsumer := &MatchConsumer{
		logger:     logger,
		Config:     config,
		usecase:    bootstrap.Usecase{},
		repository: bootstrap.Repository{},
	}

	dbTx := questdb.NewTransaction(matchConsumer.db)

	matchConsumer.Consumer = consumer.NewMatchConsumer(
		config.MatchKafka,
		logger,
		matchConsumer.usecase.TickUsecase,
		matchConsumer.usecase.OhlcUsecase,
		dbTx,
	)

	matchConsumer.initDB(ctx)
	matchConsumer.registerRepository()
	matchConsumer.registerUsecase()

	return matchConsumer, nil
}

func (s *MatchConsumer) initDB(ctx context.Context) {
	questdbClient, err := questdb.NewClient(ctx, s.Config.QuestDB)
	if err != nil {
		s.logger.ErrorContext(ctx, err, logger.Field{
			Key:   "action",
			Value: "init_db",
		})
	}
	s.db = questdbClient
}

func (s *MatchConsumer) registerRepository() {
	s.repository.OhlcRepository = ohlcInfra.NewRepository(s.db)
	s.repository.TickRepository = tickInfra.NewRepository(s.db)
}

func (s *MatchConsumer) registerUsecase() {
	s.usecase.OhlcUsecase = ohlcUc.NewUsecase(s.repository.OhlcRepository, s.logger)
	s.usecase.TickUsecase = tickUc.NewUsecase(s.repository.TickRepository, s.logger)
}
