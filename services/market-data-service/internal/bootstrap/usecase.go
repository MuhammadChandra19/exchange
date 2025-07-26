package bootstrap

import (
	orderUc "github.com/muhammadchandra19/exchange/services/market-data-service/internal/usecase/order"
	tickUc "github.com/muhammadchandra19/exchange/services/market-data-service/internal/usecase/tick"

	ohlcDomain "github.com/muhammadchandra19/exchange/services/market-data-service/internal/domain/ohlc"
	orderDomain "github.com/muhammadchandra19/exchange/services/market-data-service/internal/domain/order"
	tickDomain "github.com/muhammadchandra19/exchange/services/market-data-service/internal/domain/tick"
)

// Usecase is the usecase for the market data service.
type Usecase struct {
	OrderUsecase orderDomain.Usecase
	TickUsecase  tickDomain.Usecase
	OhlcUsecase  ohlcDomain.Usecase
}

// registerUsecase registers the usecase.
func (b *Bootstrap) registerUsecase() {
	b.Usecase.OrderUsecase = orderUc.NewUsecase(b.Repository.OrderRepository, b.Logger)
	b.Usecase.TickUsecase = tickUc.NewUsecase(b.Repository.TickRepository, b.Logger)
}
