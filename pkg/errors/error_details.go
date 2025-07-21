package errors

// ErrorDetails represents detailed information about an error.
type ErrorDetails struct {
	// Message (required) is the user-defined error message.
	// E.g. "user email has invalid format".
	Message string

	// Code (required) is the user-defined error code string that follows
	// the Buka2.0 conventions.
	// E.g. "USER_EMAIL_INVALID".
	Code string

	// Field (optional) is the related field the error occurred on, if any.
	Field string

	// Object (optional) is the related object the error occured on, if any.
	Object interface{}
}

// NewErrorDetails creates a new ErrorDetails struct with the given parameters.
func NewErrorDetails(message, code, field string) *ErrorDetails {
	return &ErrorDetails{
		Message: message,
		Code:    code,
		Field:   field,
	}
}

// NewErrorDetailsWithObject creates a new ErrorDetails struct with an associated object.
func NewErrorDetailsWithObject(message, code, field string, object interface{}) *ErrorDetails {
	return &ErrorDetails{
		Message: message,
		Code:    code,
		Field:   field,
		Object:  object,
	}
}

// Error() is used to implement the Golang `error` interface.
func (e *ErrorDetails) Error() string {
	return e.Message
}

// ErrorCodeEquals checks whether a given `error` has a specific code.
func ErrorCodeEquals(err error, code string) bool {
	errDetails, ok := err.(*ErrorDetails)
	if !ok {
		return false
	}

	return errDetails.Code == code
}
