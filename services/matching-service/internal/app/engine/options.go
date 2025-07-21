package engine

import "time"

// Options represents configuration options for the Engine.
type Options struct {
	SnapshotInterval    time.Duration
	SnapshotOffsetDelta int64
}

// DefaultEngineOptions returns the default engine options.
func DefaultEngineOptions() *Options {
	return &Options{
		SnapshotInterval:    30 * time.Second,
		SnapshotOffsetDelta: 1000,
	}
}
