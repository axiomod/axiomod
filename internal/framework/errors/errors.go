package errors

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// Common error types
var (
	ErrNotFound         = errors.New("not found")
	ErrInvalidInput     = errors.New("invalid input")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrForbidden        = errors.New("forbidden")
	ErrInternal         = errors.New("internal error")
	ErrUnavailable      = errors.New("service unavailable")
	ErrTimeout          = errors.New("timeout")
	ErrAlreadyExists    = errors.New("already exists")
	ErrConflict         = errors.New("conflict")
	ErrNotImplemented   = errors.New("not implemented")
	ErrValidation       = errors.New("validation failed")
	ErrDeadlineExceeded = errors.New("deadline exceeded")
	ErrCanceled         = errors.New("operation canceled")
)

const (
	CodeNotFound         = "NOT_FOUND"
	CodeInvalidInput     = "INVALID_INPUT"
	CodeUnauthorized     = "UNAUTHORIZED"
	CodeForbidden        = "FORBIDDEN"
	CodeInternal         = "INTERNAL_ERROR"
	CodeUnavailable      = "SERVICE_UNAVAILABLE"
	CodeTimeout          = "TIMEOUT"
	CodeAlreadyExists    = "ALREADY_EXISTS"
	CodeConflict         = "CONFLICT"
	CodeNotImplemented   = "NOT_IMPLEMENTED"
	CodeValidation       = "VALIDATION_ERROR"
	CodeDeadlineExceeded = "DEADLINE_EXCEEDED"
	CodeCanceled         = "CANCELED"
)

// Error represents an application error with stack trace and metadata
type Error struct {
	// Original is the original error
	Original error

	// Message is the error message
	Message string

	// Code is the error code
	Code string

	// Stack is the stack trace
	Stack string

	// Metadata contains additional information about the error
	Metadata map[string]interface{}
}

// New creates a new Error
func New(message string) error {
	return &Error{
		Original: errors.New(message),
		Message:  message,
		Stack:    captureStack(),
		Metadata: make(map[string]interface{}),
	}
}

// NewInternal creates a new internal error
func NewInternal(err error, message string) error {
	return WithCode(Wrap(err, message), CodeInternal)
}

// NewNotFound creates a new not found error
func NewNotFound(err error, message string) error {
	return WithCode(Wrap(err, message), CodeNotFound)
}

// NewInvalidInput creates a new invalid input error
func NewInvalidInput(err error, message string) error {
	return WithCode(Wrap(err, message), CodeInvalidInput)
}

// NewUnauthorized creates a new unauthorized error
func NewUnauthorized(err error, message string) error {
	return WithCode(Wrap(err, message), CodeUnauthorized)
}

// NewForbidden creates a new forbidden error
func NewForbidden(err error, message string) error {
	return WithCode(Wrap(err, message), CodeForbidden)
}

// NewConflict creates a new conflict error
func NewConflict(err error, message string) error {
	return WithCode(Wrap(err, message), CodeConflict)
}

// Wrap wraps an error with a message
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}

	// If the error is already an Error, just update the message
	if e, ok := err.(*Error); ok {
		return &Error{
			Original: e.Original,
			Message:  message + ": " + e.Message,
			Code:     e.Code,
			Stack:    e.Stack,
			Metadata: e.Metadata,
		}
	}

	// Create a new Error
	return &Error{
		Original: err,
		Message:  message + ": " + err.Error(),
		Stack:    captureStack(),
		Metadata: make(map[string]interface{}),
	}
}

// WithCode adds a code to an error
func WithCode(err error, code string) error {
	if err == nil {
		return nil
	}

	// If the error is already an Error, just update the code
	if e, ok := err.(*Error); ok {
		e.Code = code
		return e
	}

	// Create a new Error
	return &Error{
		Original: err,
		Message:  err.Error(),
		Code:     code,
		Stack:    captureStack(),
		Metadata: make(map[string]interface{}),
	}
}

// WithMetadata adds metadata to an error
func WithMetadata(err error, key string, value interface{}) error {
	if err == nil {
		return nil
	}

	// If the error is already an Error, just update the metadata
	if e, ok := err.(*Error); ok {
		e.Metadata[key] = value
		return e
	}

	// Create a new Error
	e := &Error{
		Original: err,
		Message:  err.Error(),
		Stack:    captureStack(),
		Metadata: make(map[string]interface{}),
	}
	e.Metadata[key] = value
	return e
}

// Error returns the error message
func (e *Error) Error() string {
	return e.Message
}

// Unwrap returns the original error
func (e *Error) Unwrap() error {
	return e.Original
}

// GetCode returns the error code
func GetCode(err error) string {
	if err == nil {
		return ""
	}

	// If the error is an Error, return its code
	if e, ok := err.(*Error); ok {
		return e.Code
	}

	return ""
}

// GetMetadata returns the error metadata
func GetMetadata(err error) map[string]interface{} {
	if err == nil {
		return nil
	}

	// If the error is an Error, return its metadata
	if e, ok := err.(*Error); ok {
		return e.Metadata
	}

	return nil
}

// GetStack returns the error stack trace
func GetStack(err error) string {
	if err == nil {
		return ""
	}

	// If the error is an Error, return its stack
	if e, ok := err.(*Error); ok {
		return e.Stack
	}

	return ""
}

// ToHTTPCode maps an error code to an HTTP status code
func ToHTTPCode(err error) int {
	code := GetCode(err)
	switch code {
	case CodeNotFound:
		return 404
	case CodeInvalidInput, CodeValidation:
		return 400
	case CodeUnauthorized:
		return 401
	case CodeForbidden:
		return 403
	case CodeConflict, CodeAlreadyExists:
		return 409
	case CodeTimeout, CodeDeadlineExceeded:
		return 408
	case CodeUnavailable:
		return 503
	case CodeNotImplemented:
		return 501
	default:
		return 500
	}
}

// ToGRPCCode maps an error code to a gRPC status code
func ToGRPCCode(err error) uint32 {
	code := GetCode(err)
	switch code {
	case CodeNotFound:
		return 5 // NotFound
	case CodeInvalidInput, CodeValidation:
		return 3 // InvalidArgument
	case CodeUnauthorized:
		return 16 // Unauthenticated
	case CodeForbidden:
		return 7 // PermissionDenied
	case CodeConflict, CodeAlreadyExists:
		return 6 // AlreadyExists
	case CodeTimeout, CodeDeadlineExceeded:
		return 4 // DeadlineExceeded
	case CodeUnavailable:
		return 14 // Unavailable
	case CodeNotImplemented:
		return 12 // Unimplemented
	case CodeCanceled:
		return 1 // Canceled
	default:
		return 13 // Internal
	}
}

// Is reports whether any error in err's chain matches target
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target, and if so, sets
// target to that error value and returns true. Otherwise, it returns false.
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// captureStack captures the current stack trace
func captureStack() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var builder strings.Builder
	for {
		frame, more := frames.Next()
		if !more {
			break
		}

		// Skip runtime and standard library frames
		if strings.Contains(frame.File, "runtime/") {
			continue
		}

		fmt.Fprintf(&builder, "%s:%d %s\n", frame.File, frame.Line, frame.Function)

		if !more {
			break
		}
	}

	return builder.String()
}
