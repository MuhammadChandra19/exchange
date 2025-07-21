package errors

import "github.com/pkg/errors"

// ErrorTracer is a custom error type that includes a message and an underlying error.
type ErrorTracer struct {
	Message string
	Err     error
}

// NewTracer creates a new ErrorTracer with the provided message.
func NewTracer(message string) *ErrorTracer {
	return &ErrorTracer{
		Message: message,
	}
}

// TracerFromError creates a new ErrorTracer from an existing error, preserving the stack trace.
func TracerFromError(err error) *ErrorTracer {
	tracer := NewTracer(err.Error())
	tracer.Err = err
	_, ok := err.(StackTracer)
	if !ok {
		tracer.Err = errors.WithStack(err)
	}
	return tracer
}

// StackTracer is an interface that requires a StackTrace method.
type StackTracer interface {
	StackTrace() errors.StackTrace
}

func (e *ErrorTracer) Error() string {
	return e.Message
}

func (e *ErrorTracer) Unwrap() error {
	return e.Err
}

// Wrap wraps an existing error into the ErrorTracer, preserving the stack trace.
func (e *ErrorTracer) Wrap(err error) *ErrorTracer {
	e.Err = err
	_, ok := err.(StackTracer)
	if !ok {
		e.Err = errors.WithStack(err)
	}

	return e
}

// StackTrace returns the stack trace of the underlying error if it implements StackTracer.
func (e *ErrorTracer) StackTrace() errors.StackTrace {
	err := e.Unwrap()
	errWithStack, ok := err.(StackTracer)
	if ok {
		return errWithStack.StackTrace()
	}
	return nil
}
