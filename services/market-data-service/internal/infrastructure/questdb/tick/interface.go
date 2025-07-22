package tick

import (
	"context"
	"time"
)

// TickRepository is the interface for the tick repository.
//
//go:generate mockgen -source=repository.go -destination=mock/repository_mock.go -package=mock
type TickRepository interface {
	GetByFilter(ctx context.Context, filter Filter) ([]*Tick, error)
	GetLatestBySymbol(ctx context.Context, symbol string) (*Tick, error)
	GetVolumeBySymbol(ctx context.Context, symbol string, from time.Time, to time.Time) (int64, error)
	Store(ctx context.Context, tick *Tick) error
	StoreBatch(ctx context.Context, ticks []*Tick) error
}
