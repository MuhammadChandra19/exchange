package ohlc

import (
	"context"

	"github.com/muhammadchandra19/exchange/services/market-data-service/internal/infrastructure/questdb/ohlc"
)

// Usecase is the interface for the OHLC usecase.
type Usecase interface {
	GetOHLC(ctx context.Context, symbol string, interval string) (*ohlc.OHLC, error)
	GetOHLCByFilter(ctx context.Context, filter ohlc.OHLCFilter) ([]*ohlc.OHLC, error)
	GetIntradayData(ctx context.Context, symbol string, interval string, limit int) ([]*ohlc.OHLC, error)
	StoreOHLC(ctx context.Context, ohlc *ohlc.OHLC) error
	StoreOHLCs(ctx context.Context, ohlcs []*ohlc.OHLC) error
}
