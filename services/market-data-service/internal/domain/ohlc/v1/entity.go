package v1

import (
	"time"
)

// OHLC represents a single OHLC (Open, High, Low, Close) data point.
type OHLC struct {
	Timestamp  time.Time
	Symbol     string
	Interval   string // "1m", "5m", "15m", "1h", "4h", "1d"
	Open       float64
	High       float64
	Low        float64
	Close      float64
	Volume     int64 // Volume is part of OHLC
	TradeCount int64
}

// OHLCFilter represents the filter criteria for OHLC data.
type OHLCFilter struct {
	Symbol   string
	Interval string
	From     *time.Time
	To       *time.Time
	Limit    int
}
