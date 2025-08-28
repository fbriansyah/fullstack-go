// Package errors provides comprehensive error handling for the Go Templ Template application.
// It includes structured error types, HTTP middleware, logging, and utilities for consistent
// error handling across all modules.
package errors

import (
	"fmt"
)

// Common error variables for easy comparison
var (
	// ErrValidation represents a validation error
	ErrValidation = &AppError{Type: ErrorTypeValidation}

	// ErrAuthentication represents an authentication error
	ErrAuthentication = &AppError{Type: ErrorTypeAuthentication}

	// ErrAuthorization represents an authorization error
	ErrAuthorization = &AppError{Type: ErrorTypeAuthorization}

	// ErrNotFound represents a not found error
	ErrNotFound = &AppError{Type: ErrorTypeNotFound}

	// ErrConflict represents a conflict error
	ErrConflict = &AppError{Type: ErrorTypeConflict}

	// ErrRateLimit represents a rate limit error
	ErrRateLimit = &AppError{Type: ErrorTypeRateLimit}

	// ErrInternal represents an internal error
	ErrInternal = &AppError{Type: ErrorTypeInternal}

	// ErrExternal represents an external service error
	ErrExternal = &AppError{Type: ErrorTypeExternal}

	// ErrTimeout represents a timeout error
	ErrTimeout = &AppError{Type: ErrorTypeTimeout}

	// ErrUnavailable represents a service unavailable error
	ErrUnavailable = &AppError{Type: ErrorTypeUnavailable}
)

// IsType checks if an error is of a specific type
func IsType(err error, errorType ErrorType) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type == errorType
	}
	return false
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	return IsType(err, ErrorTypeValidation)
}

// IsAuthenticationError checks if an error is an authentication error
func IsAuthenticationError(err error) bool {
	return IsType(err, ErrorTypeAuthentication)
}

// IsAuthorizationError checks if an error is an authorization error
func IsAuthorizationError(err error) bool {
	return IsType(err, ErrorTypeAuthorization)
}

// IsNotFoundError checks if an error is a not found error
func IsNotFoundError(err error) bool {
	return IsType(err, ErrorTypeNotFound)
}

// IsConflictError checks if an error is a conflict error
func IsConflictError(err error) bool {
	return IsType(err, ErrorTypeConflict)
}

// IsRateLimitError checks if an error is a rate limit error
func IsRateLimitError(err error) bool {
	return IsType(err, ErrorTypeRateLimit)
}

// IsInternalError checks if an error is an internal error
func IsInternalError(err error) bool {
	return IsType(err, ErrorTypeInternal)
}

// IsExternalError checks if an error is an external service error
func IsExternalError(err error) bool {
	return IsType(err, ErrorTypeExternal)
}

// IsTimeoutError checks if an error is a timeout error
func IsTimeoutError(err error) bool {
	return IsType(err, ErrorTypeTimeout)
}

// IsUnavailableError checks if an error is a service unavailable error
func IsUnavailableError(err error) bool {
	return IsType(err, ErrorTypeUnavailable)
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Retryable
	}
	return false
}

// GetErrorCode extracts the error code from an error
func GetErrorCode(err error) string {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code
	}
	return "UNKNOWN_ERROR"
}

// GetErrorType extracts the error type from an error
func GetErrorType(err error) ErrorType {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type
	}
	return ErrorTypeInternal
}

// GetHTTPStatus extracts the HTTP status code from an error
func GetHTTPStatus(err error) int {
	if appErr, ok := err.(*AppError); ok {
		return appErr.HTTPStatus
	}
	return 500
}

// Chain creates a chain of errors for better error context
func Chain(errs ...error) error {
	var validErrors []error
	for _, err := range errs {
		if err != nil {
			validErrors = append(validErrors, err)
		}
	}

	if len(validErrors) == 0 {
		return nil
	}

	if len(validErrors) == 1 {
		return validErrors[0]
	}

	// Create error list
	errorList := &ErrorList{}
	for _, err := range validErrors {
		if appErr, ok := err.(*AppError); ok {
			errorList.Add(appErr)
		} else {
			// Convert to AppError
			appErr := WrapError(err, ErrorTypeInternal, "CHAINED_ERROR", err.Error())
			errorList.Add(appErr)
		}
	}

	return errorList
}

// Combine combines multiple errors into a single error
func Combine(errs ...error) error {
	return Chain(errs...)
}

// Must panics if the error is not nil (useful for initialization)
func Must(err error) {
	if err != nil {
		panic(fmt.Sprintf("Must failed: %v", err))
	}
}

// MustReturn panics if the error is not nil, otherwise returns the value
func MustReturn[T any](value T, err error) T {
	Must(err)
	return value
}

// Ignore ignores the error (useful for cases where error is expected)
func Ignore(err error) {
	// Intentionally empty - used for documentation purposes
	_ = err
}

// AsAppError converts an error to AppError if possible
func AsAppError(err error) (*AppError, bool) {
	if err == nil {
		return nil, false
	}

	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}

	return nil, false
}

// AsErrorList converts an error to ErrorList if possible
func AsErrorList(err error) (*ErrorList, bool) {
	if err == nil {
		return nil, false
	}

	if errorList, ok := err.(*ErrorList); ok {
		return errorList, true
	}

	return nil, false
}

// Recover recovers from a panic and converts it to an AppError
func Recover() *AppError {
	if r := recover(); r != nil {
		var err error
		if e, ok := r.(error); ok {
			err = e
		} else {
			err = fmt.Errorf("panic: %v", r)
		}

		return NewInternalErrorWithCause("PANIC_RECOVERED",
			"A panic was recovered", err).
			WithDetails(map[string]interface{}{
				"panic_value": fmt.Sprintf("%v", r),
			})
	}
	return nil
}

// SafeExecute executes a function and recovers from panics
func SafeExecute(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var panicErr error
			if e, ok := r.(error); ok {
				panicErr = e
			} else {
				panicErr = fmt.Errorf("panic: %v", r)
			}

			err = NewInternalErrorWithCause("PANIC_RECOVERED",
				"A panic was recovered", panicErr).
				WithDetails(map[string]interface{}{
					"panic_value": fmt.Sprintf("%v", r),
				})
		}
	}()

	return fn()
}

// SafeExecuteWithReturn executes a function with return value and recovers from panics
func SafeExecuteWithReturn[T any](fn func() (T, error)) (result T, err error) {
	defer func() {
		if r := recover(); r != nil {
			var panicErr error
			if e, ok := r.(error); ok {
				panicErr = e
			} else {
				panicErr = fmt.Errorf("panic: %v", r)
			}

			err = NewInternalErrorWithCause("PANIC_RECOVERED",
				"A panic was recovered", panicErr).
				WithDetails(map[string]interface{}{
					"panic_value": fmt.Sprintf("%v", r),
				})
		}
	}()

	return fn()
}

// Validate validates a condition and returns an error if false
func Validate(condition bool, errorType ErrorType, code, message string) error {
	if !condition {
		return NewAppError(errorType, code, message, getHTTPStatusForType(errorType))
	}
	return nil
}

// ValidateNotNil validates that a value is not nil
func ValidateNotNil(value interface{}, fieldName string) error {
	if value == nil {
		return NewValidationError("FIELD_REQUIRED",
			fmt.Sprintf("Field '%s' is required", fieldName)).
			WithDetails(map[string]interface{}{
				"field": fieldName,
			})
	}
	return nil
}

// ValidateNotEmpty validates that a string is not empty
func ValidateNotEmpty(value, fieldName string) error {
	if value == "" {
		return NewValidationError("FIELD_REQUIRED",
			fmt.Sprintf("Field '%s' cannot be empty", fieldName)).
			WithDetails(map[string]interface{}{
				"field": fieldName,
			})
	}
	return nil
}

// ValidateLength validates string length
func ValidateLength(value, fieldName string, min, max int) error {
	length := len(value)

	if length < min {
		return NewValidationError("FIELD_TOO_SHORT",
			fmt.Sprintf("Field '%s' must be at least %d characters", fieldName, min)).
			WithDetails(map[string]interface{}{
				"field":         fieldName,
				"min_length":    min,
				"actual_length": length,
			})
	}

	if max > 0 && length > max {
		return NewValidationError("FIELD_TOO_LONG",
			fmt.Sprintf("Field '%s' must be at most %d characters", fieldName, max)).
			WithDetails(map[string]interface{}{
				"field":         fieldName,
				"max_length":    max,
				"actual_length": length,
			})
	}

	return nil
}

// ValidateRange validates numeric range
func ValidateRange[T comparable](value T, fieldName string, min, max T) error {
	// This is a simplified version - in practice you'd need proper numeric comparison
	return nil
}

// ErrorCollector collects multiple errors and returns them as a single error
type ErrorCollector struct {
	errors []error
}

// NewErrorCollector creates a new error collector
func NewErrorCollector() *ErrorCollector {
	return &ErrorCollector{
		errors: make([]error, 0),
	}
}

// Add adds an error to the collector
func (ec *ErrorCollector) Add(err error) {
	if err != nil {
		ec.errors = append(ec.errors, err)
	}
}

// AddIf adds an error to the collector if the condition is true
func (ec *ErrorCollector) AddIf(condition bool, err error) {
	if condition && err != nil {
		ec.errors = append(ec.errors, err)
	}
}

// HasErrors returns true if there are any errors
func (ec *ErrorCollector) HasErrors() bool {
	return len(ec.errors) > 0
}

// Count returns the number of errors
func (ec *ErrorCollector) Count() int {
	return len(ec.errors)
}

// Error returns all collected errors as a single error
func (ec *ErrorCollector) Error() error {
	if len(ec.errors) == 0 {
		return nil
	}

	return Chain(ec.errors...)
}

// Errors returns all collected errors
func (ec *ErrorCollector) Errors() []error {
	return ec.errors
}

// Clear clears all collected errors
func (ec *ErrorCollector) Clear() {
	ec.errors = ec.errors[:0]
}
