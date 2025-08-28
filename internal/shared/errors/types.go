package errors

import (
	"fmt"
	"net/http"
	"time"
)

// ErrorType represents the category of error
type ErrorType string

const (
	// ErrorTypeValidation represents validation errors
	ErrorTypeValidation ErrorType = "validation"

	// ErrorTypeAuthentication represents authentication errors
	ErrorTypeAuthentication ErrorType = "authentication"

	// ErrorTypeAuthorization represents authorization errors
	ErrorTypeAuthorization ErrorType = "authorization"

	// ErrorTypeNotFound represents resource not found errors
	ErrorTypeNotFound ErrorType = "not_found"

	// ErrorTypeConflict represents resource conflict errors
	ErrorTypeConflict ErrorType = "conflict"

	// ErrorTypeRateLimit represents rate limiting errors
	ErrorTypeRateLimit ErrorType = "rate_limit"

	// ErrorTypeInternal represents internal server errors
	ErrorTypeInternal ErrorType = "internal"

	// ErrorTypeExternal represents external service errors
	ErrorTypeExternal ErrorType = "external"

	// ErrorTypeTimeout represents timeout errors
	ErrorTypeTimeout ErrorType = "timeout"

	// ErrorTypeUnavailable represents service unavailable errors
	ErrorTypeUnavailable ErrorType = "unavailable"
)

// ErrorSeverity represents the severity level of an error
type ErrorSeverity string

const (
	// SeverityLow represents low severity errors (user errors, validation)
	SeverityLow ErrorSeverity = "low"

	// SeverityMedium represents medium severity errors (business logic errors)
	SeverityMedium ErrorSeverity = "medium"

	// SeverityHigh represents high severity errors (system errors)
	SeverityHigh ErrorSeverity = "high"

	// SeverityCritical represents critical errors (data corruption, security)
	SeverityCritical ErrorSeverity = "critical"
)

// AppError represents a structured application error
type AppError struct {
	// ID is a unique identifier for this error instance
	ID string `json:"id"`

	// Code is a machine-readable error code
	Code string `json:"code"`

	// Type categorizes the error
	Type ErrorType `json:"type"`

	// Severity indicates the severity level
	Severity ErrorSeverity `json:"severity"`

	// Message is a human-readable error message
	Message string `json:"message"`

	// Details provides additional error context
	Details map[string]interface{} `json:"details,omitempty"`

	// Timestamp when the error occurred
	Timestamp time.Time `json:"timestamp"`

	// HTTPStatus is the HTTP status code to return
	HTTPStatus int `json:"http_status"`

	// Cause is the underlying error that caused this error
	Cause error `json:"-"`

	// Context provides additional context about where the error occurred
	Context ErrorContext `json:"context,omitempty"`

	// Retryable indicates if the operation can be retried
	Retryable bool `json:"retryable"`

	// UserMessage is a user-friendly message (optional)
	UserMessage string `json:"user_message,omitempty"`
}

// ErrorContext provides context about where an error occurred
type ErrorContext struct {
	// Operation is the operation that was being performed
	Operation string `json:"operation,omitempty"`

	// Component is the component where the error occurred
	Component string `json:"component,omitempty"`

	// UserID is the ID of the user associated with the error
	UserID string `json:"user_id,omitempty"`

	// RequestID is the ID of the request that caused the error
	RequestID string `json:"request_id,omitempty"`

	// SessionID is the ID of the session associated with the error
	SessionID string `json:"session_id,omitempty"`

	// IPAddress is the IP address of the client
	IPAddress string `json:"ip_address,omitempty"`

	// UserAgent is the user agent of the client
	UserAgent string `json:"user_agent,omitempty"`

	// Additional metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %s (caused by: %v)", e.Code, e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches the target error
func (e *AppError) Is(target error) bool {
	if appErr, ok := target.(*AppError); ok {
		// Errors match if they have the same type, or if they have the same code and type
		return e.Type == appErr.Type
	}
	return false
}

// WithContext adds context to the error
func (e *AppError) WithContext(ctx ErrorContext) *AppError {
	e.Context = ctx
	return e
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// WithUserMessage sets a user-friendly message
func (e *AppError) WithUserMessage(message string) *AppError {
	e.UserMessage = message
	return e
}

// ToHTTPResponse converts the error to an HTTP response format
func (e *AppError) ToHTTPResponse() map[string]interface{} {
	response := map[string]interface{}{
		"error": map[string]interface{}{
			"id":        e.ID,
			"code":      e.Code,
			"type":      string(e.Type),
			"message":   e.Message,
			"timestamp": e.Timestamp,
		},
	}

	// Add user message if available
	if e.UserMessage != "" {
		response["error"].(map[string]interface{})["user_message"] = e.UserMessage
	}

	// Add details if available (but filter sensitive information)
	if len(e.Details) > 0 {
		filteredDetails := make(map[string]interface{})
		for k, v := range e.Details {
			// Filter out sensitive information
			if !isSensitiveField(k) {
				filteredDetails[k] = v
			}
		}
		if len(filteredDetails) > 0 {
			response["error"].(map[string]interface{})["details"] = filteredDetails
		}
	}

	// Add retry information for retryable errors
	if e.Retryable {
		response["error"].(map[string]interface{})["retryable"] = true
	}

	return response
}

// isSensitiveField checks if a field contains sensitive information
func isSensitiveField(field string) bool {
	sensitiveFields := []string{
		"password", "token", "secret", "key", "credential",
		"authorization", "session", "cookie", "private",
	}

	fieldLower := fmt.Sprintf("%v", field)
	for _, sensitive := range sensitiveFields {
		if contains(fieldLower, sensitive) {
			return true
		}
	}
	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr))))
}

// containsSubstring checks if string contains substring
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ErrorList represents a collection of errors
type ErrorList struct {
	Errors []*AppError `json:"errors"`
}

// Error implements the error interface
func (el *ErrorList) Error() string {
	if len(el.Errors) == 0 {
		return "no errors"
	}
	if len(el.Errors) == 1 {
		return el.Errors[0].Error()
	}
	return fmt.Sprintf("multiple errors: %d errors occurred", len(el.Errors))
}

// Add adds an error to the list
func (el *ErrorList) Add(err *AppError) {
	el.Errors = append(el.Errors, err)
}

// HasErrors returns true if there are errors in the list
func (el *ErrorList) HasErrors() bool {
	return len(el.Errors) > 0
}

// ToHTTPResponse converts the error list to an HTTP response format
func (el *ErrorList) ToHTTPResponse() map[string]interface{} {
	if len(el.Errors) == 0 {
		return map[string]interface{}{
			"errors": []interface{}{},
		}
	}

	if len(el.Errors) == 1 {
		return el.Errors[0].ToHTTPResponse()
	}

	errors := make([]interface{}, len(el.Errors))
	for i, err := range el.Errors {
		errors[i] = err.ToHTTPResponse()["error"]
	}

	return map[string]interface{}{
		"errors": errors,
	}
}

// GetHTTPStatus returns the appropriate HTTP status code for the error list
func (el *ErrorList) GetHTTPStatus() int {
	if len(el.Errors) == 0 {
		return http.StatusOK
	}

	if len(el.Errors) == 1 {
		return el.Errors[0].HTTPStatus
	}

	// For multiple errors, return the highest status code
	maxStatus := http.StatusOK
	hasServerError := false

	for _, err := range el.Errors {
		if err.HTTPStatus >= 500 {
			hasServerError = true
		}
		if err.HTTPStatus > maxStatus {
			maxStatus = err.HTTPStatus
		}
	}

	// If we have server errors, return the highest server error
	if hasServerError {
		return maxStatus
	}

	// For client errors only, return the highest client error status
	return maxStatus
}
