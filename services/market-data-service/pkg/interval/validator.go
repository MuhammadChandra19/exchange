package interval

import (
	"fmt"
	"time"
)

// ValidateTimeRange validates a time range for a specific interval
func ValidateTimeRange(from, to time.Time, intervalName string) error {
	interval, err := GetInterval(intervalName)
	if err != nil {
		return err
	}

	if from.After(to) {
		return fmt.Errorf("from time cannot be after to time")
	}

	duration := to.Sub(from)
	maxPoints := 5000 // Max data points to return

	if duration/interval.Duration > time.Duration(maxPoints) {
		return fmt.Errorf("time range too large for interval %s, max %d points allowed",
			intervalName, maxPoints)
	}

	return nil
}

// CalculateDataPoints estimates how many OHLC points will be generated
func CalculateDataPoints(from, to time.Time, intervalName string) (int, error) {
	interval, err := GetInterval(intervalName)
	if err != nil {
		return 0, err
	}

	duration := to.Sub(from)
	points := int(duration / interval.Duration)
	return points, nil
}
