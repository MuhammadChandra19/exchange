package bootstrap

import (
	orderInfra "github.com/muhammadchandra19/exchange/services/market-data-service/internal/infrastructure/questdb/order"
	tickInfra "github.com/muhammadchandra19/exchange/services/market-data-service/internal/infrastructure/questdb/tick"
)

// Repository is the repository for the market data service.
type Repository struct {
	OrderRepository orderInfra.OrderRepository
	TickRepository  tickInfra.TickRepository
}

// registerRepository registers the repository.
func (b *Bootstrap) registerRepository() {
	b.Repository.OrderRepository = orderInfra.NewRepository(b.QuestDB)
	b.Repository.TickRepository = tickInfra.NewRepository(b.QuestDB)
}
