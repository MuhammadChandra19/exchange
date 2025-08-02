package ohlc

import (
	"context"
)

//go:generate mockgen -source=interface.go -destination=mock/repository_mock.go -package=mock

// OHLCRepository represents the repository interface for OHLC data.
type OHLCRepository interface {
	Store(ctx context.Context, ohlc *OHLC) error
	StoreBatch(ctx context.Context, ohlcs []*OHLC) error
	GetByFilter(ctx context.Context, filter OHLCFilter) ([]*OHLC, error)
	GetLatest(ctx context.Context, symbol, interval string) (*OHLC, error)
	GetIntradayData(ctx context.Context, symbol string, interval string, limit int) ([]*OHLC, error)
}
