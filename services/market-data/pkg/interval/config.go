package interval

import (
	"fmt"
)

// Config holds interval-related configuration
type Config struct {
	EnabledIntervals  []string `env:"ENABLED_INTERVALS" envSeparator:"," envDefault:"1m,5m,15m,1h,1d"`
	MaxHistoryDays    int      `env:"MAX_HISTORY_DAYS" envDefault:"365"`
	AggregationBuffer int      `env:"AGGREGATION_BUFFER_SIZE" envDefault:"1000"`
	RealTimeIntervals []string `env:"REALTIME_INTERVALS" envSeparator:"," envDefault:"1m,5m"`
	BatchIntervals    []string `env:"BATCH_INTERVALS" envSeparator:"," envDefault:"1h,1d"`
}

// GetEnabledIntervals returns only the enabled intervals
func (c Config) GetEnabledIntervals() ([]Interval, error) {
	enabled := make([]Interval, 0, len(c.EnabledIntervals))

	for _, name := range c.EnabledIntervals {
		interval, err := GetInterval(name)
		if err != nil {
			return nil, fmt.Errorf("invalid interval in config: %s", name)
		}
		enabled = append(enabled, interval)
	}

	return enabled, nil
}

// IsRealTimeInterval checks if an interval should be processed in real-time
func (c Config) IsRealTimeInterval(intervalName string) bool {
	for _, name := range c.RealTimeIntervals {
		if name == intervalName {
			return true
		}
	}
	return false
}

// IsBatchInterval checks if an interval should be processed in batch
func (c Config) IsBatchInterval(intervalName string) bool {
	for _, name := range c.BatchIntervals {
		if name == intervalName {
			return true
		}
	}
	return false
}
