package ohlc

import (
	"context"

	"github.com/muhammadchandra19/exchange/pkg/errors"
	"github.com/muhammadchandra19/exchange/pkg/logger"
	"github.com/muhammadchandra19/exchange/services/market-data-service/internal/infrastructure/questdb/ohlc"
)

// Usecase is the usecase for the OHLC.
type Usecase struct {
	ohlcRepository ohlc.OHLCRepository
	logger         logger.Logger
}

// NewUsecase creates a new OHLC usecase.
func NewUsecase(ohlcRepository ohlc.OHLCRepository, logger logger.Logger) *Usecase {
	return &Usecase{ohlcRepository: ohlcRepository, logger: logger}
}

// GetOHLC gets the OHLC for a given symbol and interval.
func (u *Usecase) GetOHLC(ctx context.Context, symbol string, interval string) (*ohlc.OHLC, error) {
	ohlc, err := u.ohlcRepository.GetLatest(ctx, symbol, interval)
	if err != nil {
		return nil, errors.TracerFromError(err)
	}
	return ohlc, nil
}

// GetOHLCByFilter gets the OHLC for a given filter.
func (u *Usecase) GetOHLCByFilter(ctx context.Context, filter ohlc.OHLCFilter) ([]*ohlc.OHLC, error) {
	ohlcs, err := u.ohlcRepository.GetByFilter(ctx, filter)
	if err != nil {
		return nil, errors.TracerFromError(err)
	}
	return ohlcs, nil
}

// GetIntradayData gets the OHLC for a given symbol and interval.
func (u *Usecase) GetIntradayData(ctx context.Context, symbol string, interval string, limit int) ([]*ohlc.OHLC, error) {
	ohlcs, err := u.ohlcRepository.GetIntradayData(ctx, symbol, interval, limit)
	if err != nil {
		return nil, errors.TracerFromError(err)
	}
	return ohlcs, nil
}

// StoreOHLC stores a OHLC.
func (u *Usecase) StoreOHLC(ctx context.Context, ohlc *ohlc.OHLC) error {
	err := u.ohlcRepository.Store(ctx, ohlc)
	if err != nil {
		return errors.TracerFromError(err)
	}
	return nil
}

// StoreOHLCs stores a batch of OHLCs.
func (u *Usecase) StoreOHLCs(ctx context.Context, ohlcs []*ohlc.OHLC) error {
	err := u.ohlcRepository.StoreBatch(ctx, ohlcs)
	if err != nil {
		return errors.TracerFromError(err)
	}
	return nil
}
