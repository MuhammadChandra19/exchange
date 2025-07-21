package v1

import (
	"time"
)

// Tick represents a single tick (price and volume) data point.
type Tick struct {
	Timestamp time.Time
	Symbol    string
	Price     float64
	Volume    int64
	Side      string // "buy" or "sell"
}

// TickFilter represents the filter criteria for tick data.
type TickFilter struct {
	Symbol string
	From   *time.Time
	To     *time.Time
	Limit  int
	Offset int
}
