package tick

import (
	"context"
	"time"

	"github.com/muhammadchandra19/exchange/services/market-data-service/internal/infrastructure/questdb/tick"
)

// Usecase is the interface for the tick usecase.
type Usecase interface {
	GetLatestTick(ctx context.Context, symbol string) (*tick.Tick, error)
	GetTick(ctx context.Context, filter tick.Filter) ([]*tick.Tick, error)
	GetTickVolume(ctx context.Context, symbol string, from time.Time, to time.Time) (int64, error)
	StoreTick(ctx context.Context, tick *tick.Tick) error
	StoreTicks(ctx context.Context, ticks []*tick.Tick) error
}
