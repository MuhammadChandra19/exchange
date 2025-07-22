package tick

import (
	"context"
	"time"

	"github.com/muhammadchandra19/exchange/pkg/errors"
	"github.com/muhammadchandra19/exchange/pkg/logger"
	"github.com/muhammadchandra19/exchange/services/market-data-service/internal/infrastructure/questdb/tick"
)

// Usecase is the usecase for the tick.
type Usecase struct {
	tickRepository tick.TickRepository
	logger         logger.Logger
}

// NewUsecase creates a new tick usecase.
func NewUsecase(tickRepository tick.TickRepository, logger logger.Logger) *Usecase {
	return &Usecase{tickRepository: tickRepository, logger: logger}
}

// GetLatestTick gets the latest tick for a given symbol.
func (u *Usecase) GetLatestTick(ctx context.Context, symbol string) (*tick.Tick, error) {
	tick, err := u.tickRepository.GetLatestBySymbol(ctx, symbol)
	if err != nil {
		return nil, errors.TracerFromError(err)
	}
	return tick, nil
}

// GetTicks gets the tick for a given filter.
func (u *Usecase) GetTicks(ctx context.Context, filter tick.Filter) ([]*tick.Tick, error) {
	tick, err := u.tickRepository.GetByFilter(ctx, filter)
	if err != nil {
		return nil, errors.TracerFromError(err)
	}
	return tick, nil
}

// GetTickVolume gets the volume for a given symbol and time range.
func (u *Usecase) GetTickVolume(ctx context.Context, symbol string, from time.Time, to time.Time) (int64, error) {
	volume, err := u.tickRepository.GetVolumeBySymbol(ctx, symbol, from, to)
	if err != nil {
		return 0, errors.TracerFromError(err)
	}
	return volume, nil
}

// StoreTick stores a tick.
func (u *Usecase) StoreTick(ctx context.Context, tick *tick.Tick) error {
	err := u.tickRepository.Store(ctx, tick)
	if err != nil {
		return errors.TracerFromError(err)
	}
	return nil
}

// StoreTicks stores a batch of ticks.
func (u *Usecase) StoreTicks(ctx context.Context, ticks []*tick.Tick) error {
	err := u.tickRepository.StoreBatch(ctx, ticks)
	if err != nil {
		return errors.TracerFromError(err)
	}
	return nil
}
