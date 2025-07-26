package ohlc

import (
	"sync"
	"time"

	"github.com/muhammadchandra19/exchange/services/market-data-service/pkg/interval"
)

// Buffer holds temporary tick data for aggregation
type Buffer struct {
	Symbol     string
	Interval   string
	BucketTime time.Time
	Ticks      []interval.TickData
	LastUpdate time.Time
	Mutex      sync.Mutex
}
