package plugin

import (
	"errors"
	"strings"
)

// Error codes that can be set on PluginError.Code
const (
	// ErrorCodeUnsupported indicates the backend does not implement this operation
	ErrorCodeUnsupported = "UNSUPPORTED"

	// ErrorCodeNotFound indicates the referenced entity does not exist
	ErrorCodeNotFound = "NOT_FOUND"
)

// ErrUnsupportedOperation is a sentinel error for operations a backend
// does not implement. Errors returned by backends can be tested with
// errors.Is(err, plugin.ErrUnsupportedOperation).
var ErrUnsupportedOperation = errors.New("operation not supported by this plugin")

// Unsupported builds a *PluginError signaling that the named plugin does
// not implement the given operation. The optional hint is appended to the
// message to tell the caller what to do instead.
func Unsupported(pluginName, operation, hint string) *PluginError {
	message := "backend does not support this operation"
	if hint != "" {
		message += ": " + hint
	}

	return &PluginError{
		Plugin:      pluginName,
		Operation:   operation,
		Message:     message,
		Code:        ErrorCodeUnsupported,
		Recoverable: false,
		Cause:       ErrUnsupportedOperation,
	}
}

// IsUnsupported reports whether err (or anything it wraps) is an
// unsupported-operation error from a plugin.
func IsUnsupported(err error) bool {
	if err == nil {
		return false
	}

	var pe *PluginError
	if errors.As(err, &pe) {
		return pe.Code == ErrorCodeUnsupported || errors.Is(err, ErrUnsupportedOperation)
	}

	return errors.Is(err, ErrUnsupportedOperation)
}

// NotFound builds a *PluginError signaling that the referenced entity
// does not exist (e.g. "task <id>").
func NotFound(pluginName, operation, entity string) *PluginError {
	var sb strings.Builder
	sb.WriteString("not found: ")
	sb.WriteString(entity)

	return &PluginError{
		Plugin:      pluginName,
		Operation:   operation,
		Message:     sb.String(),
		Code:        ErrorCodeNotFound,
		Recoverable: false,
	}
}

// IsNotFound reports whether err (or anything it wraps) is a
// not-found error from a plugin.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}

	var pe *PluginError
	if errors.As(err, &pe) {
		return pe.Code == ErrorCodeNotFound
	}

	return false
}
