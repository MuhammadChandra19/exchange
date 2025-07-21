package interval

import (
	"time"
)

// CalculateBucketTime calculates the start time of the interval bucket
func (i Interval) CalculateBucketTime(timestamp time.Time) time.Time {
	switch i.Name {
	case "1m":
		return timestamp.Truncate(time.Minute)
	case "5m":
		return timestamp.Truncate(5 * time.Minute)
	case "15m":
		return timestamp.Truncate(15 * time.Minute)
	case "30m":
		return timestamp.Truncate(30 * time.Minute)
	case "1h":
		return timestamp.Truncate(time.Hour)
	case "4h":
		return timestamp.Truncate(4 * time.Hour)
	case "1d":
		// Truncate to start of day in UTC
		return time.Date(timestamp.Year(), timestamp.Month(), timestamp.Day(), 0, 0, 0, 0, timestamp.Location())
	case "1w":
		// Truncate to start of week (Monday)
		days := int(timestamp.Weekday())
		if days == 0 { // Sunday
			days = 7
		}
		return timestamp.AddDate(0, 0, 1-days).Truncate(24 * time.Hour)
	default:
		return timestamp.Truncate(i.Duration)
	}
}

// GetBucketRange returns the start and end time of the interval bucket
func (i Interval) GetBucketRange(timestamp time.Time) (start, end time.Time) {
	start = i.CalculateBucketTime(timestamp)
	end = start.Add(i.Duration)
	return start, end
}

// IsInBucket checks if a timestamp falls within the same bucket as another timestamp
func (i Interval) IsInBucket(timestamp1, timestamp2 time.Time) bool {
	bucket1 := i.CalculateBucketTime(timestamp1)
	bucket2 := i.CalculateBucketTime(timestamp2)
	return bucket1.Equal(bucket2)
}
