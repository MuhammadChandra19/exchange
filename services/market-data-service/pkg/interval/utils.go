package interval

import (
	"fmt"
	"time"
)

// Interval represents a time interval for OHLC data
type Interval struct {
	Name     string
	Duration time.Duration
	Format   string
}

// Supported intervals configuration
var (
	Interval1m  = Interval{Name: "1m", Duration: time.Minute, Format: "2006-01-02 15:04:00"}
	Interval5m  = Interval{Name: "5m", Duration: 5 * time.Minute, Format: "2006-01-02 15:04:00"}
	Interval15m = Interval{Name: "15m", Duration: 15 * time.Minute, Format: "2006-01-02 15:04:00"}
	Interval30m = Interval{Name: "30m", Duration: 30 * time.Minute, Format: "2006-01-02 15:04:00"}
	Interval1h  = Interval{Name: "1h", Duration: time.Hour, Format: "2006-01-02 15:00:00"}
	Interval4h  = Interval{Name: "4h", Duration: 4 * time.Hour, Format: "2006-01-02 15:00:00"}
	Interval1d  = Interval{Name: "1d", Duration: 24 * time.Hour, Format: "2006-01-02 00:00:00"}
	Interval1w  = Interval{Name: "1w", Duration: 7 * 24 * time.Hour, Format: "2006-01-02 00:00:00"}
)

// All supported intervals
var AllIntervals = []Interval{
	Interval1m, Interval5m, Interval15m, Interval30m,
	Interval1h, Interval4h, Interval1d, Interval1w,
}

// Common interval groups
var (
	ShortTermIntervals  = []Interval{Interval1m, Interval5m, Interval15m}
	MediumTermIntervals = []Interval{Interval30m, Interval1h, Interval4h}
	LongTermIntervals   = []Interval{Interval1d, Interval1w}
)

// Interval registry for lookup
var intervalRegistry = make(map[string]Interval)

func init() {
	for _, interval := range AllIntervals {
		intervalRegistry[interval.Name] = interval
	}
}

// GetInterval returns an interval by name
func GetInterval(name string) (Interval, error) {
	interval, exists := intervalRegistry[name]
	if !exists {
		return Interval{}, fmt.Errorf("unsupported interval: %s", name)
	}
	return interval, nil
}

// IsValidInterval checks if interval name is supported
func IsValidInterval(name string) bool {
	_, exists := intervalRegistry[name]
	return exists
}

// GetAllIntervalNames returns all supported interval names
func GetAllIntervalNames() []string {
	names := make([]string, 0, len(AllIntervals))
	for _, interval := range AllIntervals {
		names = append(names, interval.Name)
	}
	return names
}

// CalculateBucketTime calculates the bucket start time for a given timestamp and interval
func CalculateBucketTime(timestamp time.Time, intervalName string) (time.Time, error) {
	interval, err := GetInterval(intervalName)
	if err != nil {
		return time.Time{}, err
	}

	return interval.CalculateBucketTime(timestamp), nil
}

// GetAffectedIntervals returns all intervals that should be updated for a given tick
func GetAffectedIntervals(timestamp time.Time) map[string]time.Time {
	affected := make(map[string]time.Time)

	for _, interval := range AllIntervals {
		bucketTime := interval.CalculateBucketTime(timestamp)
		affected[interval.Name] = bucketTime
	}

	return affected
}

// GetIntervalsForTimeframe returns intervals suitable for a specific timeframe
func GetIntervalsForTimeframe(timeframe string) []Interval {
	switch timeframe {
	case "scalping", "short":
		return ShortTermIntervals
	case "day_trading", "medium":
		return MediumTermIntervals
	case "swing_trading", "position", "long":
		return LongTermIntervals
	default:
		return AllIntervals
	}
}
