package ohlc

import (
	"sync"
	"time"

	"github.com/muhammadchandra19/exchange/proto/go/modules/market-data/v1/shared"
	"github.com/muhammadchandra19/exchange/services/market-data/pkg/interval"
)

// Buffer holds temporary tick data for aggregation
type Buffer struct {
	Symbol     string
	Interval   shared.Interval
	BucketTime time.Time
	Ticks      []interval.TickData
	LastUpdate time.Time
	Mutex      sync.Mutex
}
