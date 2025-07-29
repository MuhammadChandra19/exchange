package util

import "time"

// TimePoiner converts a time.Time to a pointer to a time.Time.
func TimePoiner(t time.Time) *time.Time {
	return &t
}
