package v1

import (
	"fmt"
	"time"

	"github.com/muhammadchandra19/exchange/services/market-data-service/pkg/interval"
)

// OHLC represents a single OHLC (Open, High, Low, Close) data point.
type OHLC struct {
	Timestamp  time.Time
	Symbol     string
	Interval   string // Use interval.GetAllIntervalNames() for validation
	Open       float64
	High       float64
	Low        float64
	Close      float64
	Volume     int64
	TradeCount int64
	Exchange   string
}

// OHLCFilter represents the filter criteria for OHLC data.
type OHLCFilter struct {
	Symbol   string
	Interval string // Must be one of interval.GetAllIntervalNames()
	Exchange string
	From     *time.Time
	To       *time.Time
	Limit    int
}

// ValidateInterval validates the interval field
func (o *OHLC) ValidateInterval() error {
	if !interval.IsValidInterval(o.Interval) {
		return fmt.Errorf("invalid interval: %s, supported: %v",
			o.Interval, interval.GetAllIntervalNames())
	}
	return nil
}

// GetBucketTime calculates the correct bucket time for this OHLC
func (o *OHLC) GetBucketTime() (time.Time, error) {
	return interval.CalculateBucketTime(o.Timestamp, o.Interval)
}
