package v1

import (
	"context"
	"time"
)

//go:generate mockgen -source=repository.go -destination=mock/repository_mock.go -package=mock

// TickRepository represents the repository interface for tick data.
type TickRepository interface {
	Store(ctx context.Context, tick *Tick) error
	StoreBatch(ctx context.Context, ticks []*Tick) error
	GetByFilter(ctx context.Context, filter TickFilter) ([]*Tick, error)
	GetLatestBySymbol(ctx context.Context, symbol string) (*Tick, error)
	GetVolumeBySymbol(ctx context.Context, symbol string, from, to time.Time) (int64, error)
}
