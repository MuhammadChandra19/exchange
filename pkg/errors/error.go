package errors

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
)

// ErrorCode represents a specific error code in the system.
type ErrorCode string

const (
	// GeneralInternalServerError represents a generic internal server error.
	GeneralInternalServerError ErrorCode = "general_internal_server_error"
	// GeneralBadRequestError represents a generic bad request error.
	GeneralBadRequestError ErrorCode = "general_bad_request_error"
	// GeneralNotFoundError represents a generic not found error.
	GeneralNotFoundError ErrorCode = "general_not_found_error"
	// GeneralUnauthorizedError represents a generic unauthorized error.
	GeneralUnauthorizedError ErrorCode = "general_unauthorized_error"
	// GeneralForbiddenError represents a generic forbidden error.
	GeneralForbiddenError ErrorCode = "general_forbidden_error"
	// GeneralRepositoryError represents a generic repository error.
	GeneralRepositoryError ErrorCode = "general_repository_error"

	// ErrInsufficientAskVolume represents an error when there is not enough ask volume to fill a market order.
	ErrInsufficientAskVolume ErrorCode = "insufficient_ask_volume"
	// ErrInsufficientBidVolume represents an error when there is not enough bid volume to fill a market order.
	ErrInsufficientBidVolume ErrorCode = "insufficient_bid_volume"

	// RedisConfigError represents an error when the Redis configuration is invalid or nil.
	RedisConfigError ErrorCode = "redis_config_error"
	// RedisConnectionError represents an error when connecting to Redis.
	RedisConnectionError ErrorCode = "redis_connection_error"
	// RedisDisconnectionError represents an error when disconnecting from Redis.
	RedisDisconnectionError ErrorCode = "redis_disconnection_error"
	// RedisPingError represents an error when pinging Redis.
	RedisPingError ErrorCode = "redis_pinging_error"

	// RedisGetError represents an error when getting a value from Redis.
	RedisGetError ErrorCode = "redis_get_error"
	// RedisSetError represents an error when setting a value in Redis.
	RedisSetError ErrorCode = "redis_set_error"
	// RedisDelError represents an error when deleting a value from Redis.
	RedisDelError ErrorCode = "redis_del_error"
	// RedisSetNXError represents an error when setting a value in Redis with SetNX.
	RedisSetNXError ErrorCode = "redis_setnx_error"

	// RedisHGetError represents an error when getting a field from a hash in Redis.
	RedisHGetError ErrorCode = "redis_hget_error"
	// RedisHSetError represents an error when setting fields in a hash in Redis.
	RedisHSetError ErrorCode = "redis_hset_error"
	// RedisHDelError represents an error when deleting fields from a hash in Redis.
	RedisHDelError ErrorCode = "redis_hdel_error"

	// RedisZAddError represents an error when adding members to a sorted set in Redis.
	RedisZAddError ErrorCode = "redis_zadd_error"
	// RedisSubscribeError represents an error when subscribing to channels in Redis.
	RedisSubscribeError ErrorCode = "redis_subscribe_error"
	// RedisPublishError represents an error when publishing messages to channels in Redis.
	RedisPublishError ErrorCode = "redis_publish_error"
	// RedisXAddError represents an error when adding entries to a stream in Redis.
	RedisXAddError ErrorCode = "redis_xadd_error"
	// RedisXLenError represents an error when getting the length of a stream in Redis.
	RedisXLenError ErrorCode = "redis_xlen_error"
	// RedisXReadError represents an error when reading from a stream in Redis.
	RedisXReadError ErrorCode = "redis_xread_error"
	// RedisXReadGroupError represents an error when reading from a stream group in Redis.
	RedisXReadGroupError ErrorCode = "redis_xreadgroup_error"
)

// Severity represents the severity level of an error.
type Severity string

const (
	// SeverityCritical indicates a critical error that requires immediate attention.
	SeverityCritical Severity = "critical"
	// SeverityHigh indicates a high severity error that should be addressed promptly.
	SeverityHigh Severity = "high"
	// SeverityMedium indicates a medium severity error that should be addressed in due course.
	SeverityMedium Severity = "medium"
	// SeverityLow indicates a low severity error that can be addressed at a later time.
	SeverityLow Severity = "low"
)

// Category represents the category of an error.
type Category string

const (
	// CategoryDatabase indicates an error related to database operations.
	CategoryDatabase Category = "database"
	// CategoryNetwork indicates an error related to network operations.
	CategoryNetwork Category = "network"
	// CategoryValidation indicates an error related to validation of input data.
	CategoryValidation Category = "validation"
	// CategoryBusinessLogic indicates an error related to business logic processing.
	CategoryBusinessLogic Category = "business_logic"
	// CategoryUnknown indicates an unknown error category.
	CategoryUnknown Category = "unknown"
	// CategoryExternal indicates an error related to external services or APIs.
	CategoryExternal Category = "external"
)

// BaseError is an `error` type containing an array of ErrorDetails.
// This error provides basic functions for performing transformations
// on a list of ErrorDetails.
type BaseError struct {
	details []*ErrorDetails
}

// NewBaseError create BaseError with ErrorDetails
func NewBaseError(details ...*ErrorDetails) *BaseError {
	return &BaseError{details: details}
}

// AddErrorDetails add more ErrorDetails to BaseError
func (b *BaseError) AddErrorDetails(errors ...*ErrorDetails) {
	b.details = append(b.details, errors...)
}

// GetDetails get array ErrorDetails on BaseError
func (b *BaseError) GetDetails() []*ErrorDetails {
	return b.details
}

// Error implement error interface
func (b *BaseError) Error() string {
	buff := bytes.NewBufferString("")

	buff.WriteString("Error on\n")
	for _, err := range b.details {
		buff.WriteString("code: ")
		buff.WriteString(err.Code)
		buff.WriteString("; error: ")
		buff.WriteString(err.Error())
		buff.WriteString("; field: ")
		buff.WriteString(err.Field)
		buff.WriteString("; object: ")
		if err.Object != nil {
			buff.WriteString(reflect.TypeOf(err.Object).String())
		}
		buff.WriteString("\n")
	}

	return strings.TrimSpace(buff.String())
}

// ReplaceAllObjects set all object on ErrorDetails with given object
func (b *BaseError) ReplaceAllObjects(object interface{}) {
	for _, d := range b.GetDetails() {
		d.Object = object
	}
}

// ReplaceObjects replace object on ErrorDetails from given mapping.
// usage: usecase have a single struct user as params, but inside usecase we split that struct
// into multiple struct before send it to repository. We need to change error object
// from repository into user struct as return value of usecase
// mapping example:
//
//	map[interface{}]interface{}{
//		address: user,
//		userDetail: user,
//	}
func (b *BaseError) ReplaceObjects(mapping map[interface{}]interface{}) {
	for _, d := range b.GetDetails() {
		val, ok := mapping[d.Object]
		if !ok {
			continue
		}

		d.Object = val
	}
}

// RenameFields rename field on ErrorDetails from given mapping
func (b *BaseError) RenameFields(mapping map[string]string) {
	for _, d := range b.GetDetails() {
		val, ok := mapping[d.Field]
		if !ok {
			continue
		}

		d.Field = val
	}
}

// RenameFieldsWithFunction rename field on ErrorDetails from given function mapping
func (b *BaseError) RenameFieldsWithFunction(mappFunc func(string) string) {
	for _, d := range b.GetDetails() {
		d.Field = mappFunc(d.Field)
	}
}

// PrependFields prepend all field on ErrorDetails with given prefix. Will skip ErrorDetail without field
func (b *BaseError) PrependFields(prefix string) {
	for _, d := range b.GetDetails() {
		if d.Field == "" {
			continue
		}
		d.Field = fmt.Sprintf("%s%s", prefix, d.Field)
	}
}

// PrependFieldsByObject prepend all field on ErrorDetails with given object mapping. Will skip ErrorDetail without field
func (b *BaseError) PrependFieldsByObject(prefixes map[interface{}]string) {
	for _, d := range b.GetDetails() {
		if d.Field == "" {
			continue
		}

		prefix := prefixes[d.Object]

		if prefix == "" {
			continue
		}

		d.Field = fmt.Sprintf("%s%s", prefix, d.Field)
	}
}

// UpdateCode update all code on ErrorDetails with given code
func (b *BaseError) UpdateCode(code string) {
	for _, d := range b.GetDetails() {
		d.Code = code
	}
}

// ReplaceCode update domain and resource code by given mapping
func (b *BaseError) ReplaceCode(mapping map[string]string) {
	for _, d := range b.GetDetails() {
		val, ok := mapping[d.Code]
		if ok {
			d.Code = val
		}
	}
}

// IsAllExpectedCode check if all ErrorDetails code is expected from given codes
func (b *BaseError) IsAllExpectedCode(codes ...string) bool {
	if len(b.details) == 0 {
		return false
	}

	expectedCodes := map[string]bool{}
	for _, code := range codes {
		expectedCodes[code] = true
	}

	for _, d := range b.GetDetails() {
		if !expectedCodes[d.Code] {
			return false
		}
	}
	return true
}

// IsAllCodeEqual check if all ErrorDetails code is equal with given code
func (b *BaseError) IsAllCodeEqual(code string) bool {
	if len(b.details) == 0 {
		return false
	}

	for _, d := range b.GetDetails() {
		if d.Code != code {
			return false
		}
	}
	return true
}

// IsAnyCodeEqual check if any ErrorDetails code is equal with given code
func (b *BaseError) IsAnyCodeEqual(code string) bool {
	for _, d := range b.GetDetails() {
		if d.Code == code {
			return true
		}
	}
	return false
}

// GetObjectErrorDetailsMap group ErrorDetails that has object by field
func (b *BaseError) GetObjectErrorDetailsMap(obj interface{}) map[string][]*ErrorDetails {
	errMap := make(map[string][]*ErrorDetails)

	for _, detail := range b.details {
		if detail.Object == nil || !reflect.DeepEqual(detail.Object, obj) {
			continue
		}

		errMap[detail.Field] = append(errMap[detail.Field], detail)
	}

	return errMap
}

// GetNonObjectErrorDetailsMap group ErrorDetails that doesn't have object by field
func (b *BaseError) GetNonObjectErrorDetailsMap() map[string][]*ErrorDetails {
	errMap := make(map[string][]*ErrorDetails)

	for _, detail := range b.details {
		if detail.Object != nil {
			continue
		}

		errMap[detail.Field] = append(errMap[detail.Field], detail)
	}

	return errMap
}
