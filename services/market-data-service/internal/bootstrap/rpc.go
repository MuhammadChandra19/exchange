package bootstrap

import "github.com/muhammadchandra19/exchange/services/market-data-service/internal/rpc"

// RPC is the RPC server for the market data service.
type RPC struct {
	OrderRPC *rpc.OrderRPC
	TickRPC  *rpc.TickRPC
}

// registerRPC registers the RPC server.
func (b *Bootstrap) registerRPC() {
	b.RPC.TickRPC = rpc.NewTickRPC(b.Usecase.TickUsecase, b.Logger)
	b.RPC.OrderRPC = rpc.NewOrderRPC(b.Usecase.OrderUsecase, b.Logger)
}
