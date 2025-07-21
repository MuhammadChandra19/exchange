package util

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey string

const (
	contextKey = key("x-request-id")
)

// ContextWithRequestID returns a context with a request id
// It will generate new request id if the provided id is empty
// Deprecated: Replaced by root-level context.WithRequestID()
// and should not be used.
func ContextWithRequestID(ctx context.Context, id string) context.Context {
	if id == "" {
		return context.WithValue(ctx, contextKey, generate())
	}

	return context.WithValue(ctx, contextKey, id)
}

// generate returns a uuid-v4 string to use as request id
func generate() string {
	return uuid.NewString()
}

// FromContext returns a request id from ctx if available
// Deprecated: Replaced by root-level context.WithRequestID()
// and should not be used.
func FromContext(ctx context.Context) string {
	id, _ := ctx.Value(contextKey).(string)

	return id
}
