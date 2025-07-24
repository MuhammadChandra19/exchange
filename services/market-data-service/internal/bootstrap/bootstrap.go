package bootstrap

import (
	"github.com/muhammadchandra19/exchange/pkg/logger"
	"github.com/muhammadchandra19/exchange/services/market-data-service/internal/infrastructure/questdb"
)

// Bootstrap is the bootstrap for the market data service.
type Bootstrap struct {
	Usecase    Usecase
	Logger     logger.Interface
	RPC        RPC
	Repository Repository

	QuestDB questdb.QuestDBClient
}

// BoostrapConfig is the config for the bootstrap.
type BoostrapConfig struct {
	QuestDB questdb.QuestDBClient
	Logger  logger.Interface
}

// Init initializes the bootstrap.
func (b *Bootstrap) Init(config BoostrapConfig) Bootstrap {
	b.QuestDB = config.QuestDB
	b.Logger = config.Logger

	b.registerRepository()
	b.registerUsecase()
	b.registerRPC()

	return *b
}
