package bootstrap

import (
	ohlcInfra "github.com/muhammadchandra19/exchange/services/market-data/internal/infrastructure/questdb/ohlc"
	orderInfra "github.com/muhammadchandra19/exchange/services/market-data/internal/infrastructure/questdb/order"
	tickInfra "github.com/muhammadchandra19/exchange/services/market-data/internal/infrastructure/questdb/tick"
)

// Repository is the repository for the market data service.
type Repository struct {
	OrderRepository orderInfra.OrderRepository
	TickRepository  tickInfra.TickRepository
	OhlcRepository  ohlcInfra.OHLCRepository
}

// registerRepository registers the repository.
func (b *Bootstrap) registerRepository() {
	b.Repository.OrderRepository = orderInfra.NewRepository(b.QuestDB)
	b.Repository.TickRepository = tickInfra.NewRepository(b.QuestDB)
	b.Repository.OhlcRepository = ohlcInfra.NewRepository(b.QuestDB)
}
