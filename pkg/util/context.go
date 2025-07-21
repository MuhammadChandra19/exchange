package util

import (
	"context"
)

type FieldsFromContext struct{}

type key string

const (
	clientIPKey       = key("x-forwarded-for")
	clientIDKey       = key("client-id")
	actorIDKey        = key("actor-id")
	deviceIDKey       = key("x-device-id")
	eventIDKey        = key("event-id")
	requestHeadersKey = key("request-headers")
)

// Fields returns a map of the key-value pairs that this library has set into `context`.
func (f *FieldsFromContext) Fields(ctx context.Context) map[string]interface{} {
	mapFields := make(map[string]interface{})
	mapFields["request_id"] = GetRequestID(ctx)
	mapFields["client_ip"] = GetClientIP(ctx)
	mapFields["device_id"] = GetDeviceID(ctx)

	return mapFields
}

// WithClientIP returns a context with a client ip
func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, clientIPKey, ip)
}

// WithClientID returns a context with a client id
func WithClientID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, clientIDKey, id)
}

// WithActorID returns a context with a client id
func WithActorID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, actorIDKey, id)
}

// WithDeviceID returns a context with a client id
func WithDeviceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, deviceIDKey, id)
}

// WithRequestID returns a context with request id
// still using old implementation until deprecated
func WithRequestID(ctx context.Context, id string) context.Context {
	return ContextWithRequestID(ctx, id)
}

// WithEventID returns a context with event id
func WithEventID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, eventIDKey, id)
}

// WithRequestHeaders returns a context with request headers
func WithRequestHeaders(ctx context.Context, headers map[string][]string) context.Context {
	return context.WithValue(ctx, requestHeadersKey, headers)
}

// GetClientIP returns client ip from context
// will return nil if not present
func GetClientIP(ctx context.Context) string {
	id, _ := ctx.Value(clientIPKey).(string)
	return id
}

// GetClientID returns client ip from context
// will return nil if not present
func GetClientID(ctx context.Context) string {
	id, _ := ctx.Value(clientIDKey).(string)
	return id
}

// GetActorID returns client ip from context
// will return nil if not present
func GetActorID(ctx context.Context) string {
	id, _ := ctx.Value(actorIDKey).(string)
	return id
}

// GetDeviceID returns client ip from context
// will return nil if not present
func GetDeviceID(ctx context.Context) string {
	id, _ := ctx.Value(deviceIDKey).(string)
	return id
}

// GetRequestID returns request id from context
// will generate request id if not present
// still using old implementation until deprecated
func GetRequestID(ctx context.Context) string {
	return FromContext(ctx)
}

// GetEventID returns event id from context
// will return nil if not present
func GetEventID(ctx context.Context) string {
	id, _ := ctx.Value(eventIDKey).(string)
	return id
}

// GetRequestHeaders returns request headers from context
// will return nil valued map[string][]string if not present
func GetRequestHeaders(ctx context.Context) map[string][]string {
	headers, _ := ctx.Value(requestHeadersKey).(map[string][]string)
	return headers
}
