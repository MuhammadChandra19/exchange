package interval

import (
	"time"
)

// TickData represents basic tick information for aggregation
type TickData struct {
	Timestamp time.Time
	Price     float64
	Volume    int64
}

// OHLCData represents aggregated OHLC data
type OHLCData struct {
	Timestamp  time.Time
	Open       float64
	High       float64
	Low        float64
	Close      float64
	Volume     int64
	TradeCount int64
}

// AggregateOHLC aggregates tick data into OHLC for a specific interval
func (i Interval) AggregateOHLC(ticks []TickData, bucketTime time.Time) OHLCData {
	if len(ticks) == 0 {
		return OHLCData{Timestamp: bucketTime}
	}

	// Sort ticks by timestamp to ensure correct open/close
	// (assuming ticks are already sorted)

	ohlc := OHLCData{
		Timestamp:  bucketTime,
		Open:       ticks[0].Price,
		High:       ticks[0].Price,
		Low:        ticks[0].Price,
		Close:      ticks[len(ticks)-1].Price,
		Volume:     0,
		TradeCount: int64(len(ticks)),
	}

	// Calculate high, low, and total volume
	for _, tick := range ticks {
		if tick.Price > ohlc.High {
			ohlc.High = tick.Price
		}
		if tick.Price < ohlc.Low {
			ohlc.Low = tick.Price
		}
		ohlc.Volume += tick.Volume
	}

	return ohlc
}

// ShouldAggregate determines if it's time to aggregate based on current time and last aggregation
func (i Interval) ShouldAggregate(lastAggregation, currentTime time.Time) bool {
	lastBucket := i.CalculateBucketTime(lastAggregation)
	currentBucket := i.CalculateBucketTime(currentTime)

	// Should aggregate if we've moved to a new bucket
	return !lastBucket.Equal(currentBucket)
}
