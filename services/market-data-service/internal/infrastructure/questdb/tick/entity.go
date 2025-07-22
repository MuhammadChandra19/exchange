package tick

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

// Filter represents the filter criteria for tick data.
type Filter struct {
	Symbol string
	From   *time.Time
	To     *time.Time
	Limit  int
	Offset int
}
