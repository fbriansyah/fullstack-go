package errors

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// NewAppError creates a new AppError with the given parameters
func NewAppError(errorType ErrorType, code, message string, httpStatus int) *AppError {
	return &AppError{
		ID:         generateErrorID(),
		Code:       code,
		Type:       errorType,
		Severity:   getSeverityForType(errorType),
		Message:    message,
		Timestamp:  time.Now().UTC(),
		HTTPStatus: httpStatus,
		Retryable:  isRetryableByType(errorType),
	}
}

// NewAppErrorWithCause creates a new AppError with an underlying cause
func NewAppErrorWithCause(errorType ErrorType, code, message string, httpStatus int, cause error) *AppError {
	return &AppError{
		ID:         generateErrorID(),
		Code:       code,
		Type:       errorType,
		Severity:   getSeverityForType(errorType),
		Message:    message,
		Timestamp:  time.Now().UTC(),
		HTTPStatus: httpStatus,
		Cause:      cause,
		Retryable:  isRetryableByType(errorType),
	}
}

// Validation error builders
func NewValidationError(code, message string) *AppError {
	return NewAppError(ErrorTypeValidation, code, message, http.StatusBadRequest)
}

func NewValidationErrorWithDetails(code, message string, details map[string]interface{}) *AppError {
	return NewAppError(ErrorTypeValidation, code, message, http.StatusBadRequest).WithDetails(details)
}

// Authentication error builders
func NewAuthenticationError(code, message string) *AppError {
	return NewAppError(ErrorTypeAuthentication, code, message, http.StatusUnauthorized)
}

func NewInvalidCredentialsError() *AppError {
	return NewAuthenticationError("INVALID_CREDENTIALS", "Invalid username or password").
		WithUserMessage("The username or password you entered is incorrect. Please try again.")
}

func NewSessionExpiredError() *AppError {
	return NewAuthenticationError("SESSION_EXPIRED", "Session has expired").
		WithUserMessage("Your session has expired. Please log in again.")
}

func NewSessionInvalidError() *AppError {
	return NewAuthenticationError("SESSION_INVALID", "Session is invalid").
		WithUserMessage("Your session is invalid. Please log in again.")
}

// Authorization error builders
func NewAuthorizationError(code, message string) *AppError {
	return NewAppError(ErrorTypeAuthorization, code, message, http.StatusForbidden)
}

func NewInsufficientPermissionsError() *AppError {
	return NewAuthorizationError("INSUFFICIENT_PERMISSIONS", "Insufficient permissions to perform this action").
		WithUserMessage("You don't have permission to perform this action.")
}

func NewAccountSuspendedError() *AppError {
	return NewAuthorizationError("ACCOUNT_SUSPENDED", "Account has been suspended").
		WithUserMessage("Your account has been suspended. Please contact support for assistance.")
}

// Not found error builders
func NewNotFoundError(resource, id string) *AppError {
	message := fmt.Sprintf("%s not found", resource)
	if id != "" {
		message = fmt.Sprintf("%s with ID '%s' not found", resource, id)
	}

	return NewAppError(ErrorTypeNotFound, "RESOURCE_NOT_FOUND", message, http.StatusNotFound).
		WithDetails(map[string]interface{}{
			"resource": resource,
			"id":       id,
		}).
		WithUserMessage("The requested resource could not be found.")
}

func NewUserNotFoundError(userID string) *AppError {
	return NewNotFoundError("User", userID)
}

func NewSessionNotFoundError(sessionID string) *AppError {
	return NewNotFoundError("Session", sessionID)
}

// Conflict error builders
func NewConflictError(code, message string) *AppError {
	return NewAppError(ErrorTypeConflict, code, message, http.StatusConflict)
}

func NewDuplicateResourceError(resource, field, value string) *AppError {
	message := fmt.Sprintf("%s with %s '%s' already exists", resource, field, value)
	return NewConflictError("DUPLICATE_RESOURCE", message).
		WithDetails(map[string]interface{}{
			"resource": resource,
			"field":    field,
			"value":    value,
		}).
		WithUserMessage(fmt.Sprintf("A %s with this %s already exists.", resource, field))
}

func NewEmailAlreadyExistsError(email string) *AppError {
	return NewDuplicateResourceError("User", "email", email)
}

func NewOptimisticLockError(resource string) *AppError {
	return NewConflictError("OPTIMISTIC_LOCK_CONFLICT",
		fmt.Sprintf("The %s has been modified by another process", resource)).
		WithUserMessage("The resource has been modified by another user. Please refresh and try again.")
}

// Rate limit error builders
func NewRateLimitError(limit int, window string) *AppError {
	message := fmt.Sprintf("Rate limit exceeded: %d requests per %s", limit, window)
	return NewAppError(ErrorTypeRateLimit, "RATE_LIMIT_EXCEEDED", message, http.StatusTooManyRequests).
		WithDetails(map[string]interface{}{
			"limit":  limit,
			"window": window,
		}).
		WithUserMessage("You have made too many requests. Please wait a moment and try again.")
}

// Internal error builders
func NewInternalError(code, message string) *AppError {
	return NewAppError(ErrorTypeInternal, code, message, http.StatusInternalServerError).
		WithUserMessage("An internal error occurred. Please try again later.")
}

func NewInternalErrorWithCause(code, message string, cause error) *AppError {
	return NewAppErrorWithCause(ErrorTypeInternal, code, message, http.StatusInternalServerError, cause).
		WithUserMessage("An internal error occurred. Please try again later.")
}

func NewDatabaseError(operation string, cause error) *AppError {
	return NewInternalErrorWithCause("DATABASE_ERROR",
		fmt.Sprintf("Database error during %s", operation), cause).
		WithDetails(map[string]interface{}{
			"operation": operation,
		})
}

func NewEventBusError(operation string, cause error) *AppError {
	return NewInternalErrorWithCause("EVENT_BUS_ERROR",
		fmt.Sprintf("Event bus error during %s", operation), cause).
		WithDetails(map[string]interface{}{
			"operation": operation,
		})
}

// External service error builders
func NewExternalServiceError(service, operation string, cause error) *AppError {
	message := fmt.Sprintf("External service '%s' error during %s", service, operation)
	return NewAppErrorWithCause(ErrorTypeExternal, "EXTERNAL_SERVICE_ERROR", message,
		http.StatusBadGateway, cause).
		WithDetails(map[string]interface{}{
			"service":   service,
			"operation": operation,
		}).
		WithUserMessage("An external service is currently unavailable. Please try again later.")
}

// Timeout error builders
func NewTimeoutError(operation string, timeout time.Duration) *AppError {
	message := fmt.Sprintf("Operation '%s' timed out after %v", operation, timeout)
	return NewAppError(ErrorTypeTimeout, "OPERATION_TIMEOUT", message, http.StatusRequestTimeout).
		WithDetails(map[string]interface{}{
			"operation": operation,
			"timeout":   timeout.String(),
		}).
		WithUserMessage("The operation took too long to complete. Please try again.")
}

// Service unavailable error builders
func NewServiceUnavailableError(service string) *AppError {
	message := fmt.Sprintf("Service '%s' is currently unavailable", service)
	return NewAppError(ErrorTypeUnavailable, "SERVICE_UNAVAILABLE", message,
		http.StatusServiceUnavailable).
		WithDetails(map[string]interface{}{
			"service": service,
		}).
		WithUserMessage("The service is currently unavailable. Please try again later.")
}

// Helper functions

// generateErrorID generates a unique error ID
func generateErrorID() string {
	return uuid.New().String()
}

// getSeverityForType returns the default severity for an error type
func getSeverityForType(errorType ErrorType) ErrorSeverity {
	switch errorType {
	case ErrorTypeValidation, ErrorTypeAuthentication, ErrorTypeAuthorization, ErrorTypeNotFound:
		return SeverityLow
	case ErrorTypeConflict, ErrorTypeRateLimit:
		return SeverityMedium
	case ErrorTypeInternal, ErrorTypeExternal, ErrorTypeTimeout, ErrorTypeUnavailable:
		return SeverityHigh
	default:
		return SeverityMedium
	}
}

// isRetryableByType returns whether an error type is generally retryable
func isRetryableByType(errorType ErrorType) bool {
	switch errorType {
	case ErrorTypeTimeout, ErrorTypeUnavailable, ErrorTypeExternal:
		return true
	case ErrorTypeRateLimit:
		return true // After waiting
	case ErrorTypeInternal:
		return false // Usually not retryable without changes
	case ErrorTypeValidation, ErrorTypeAuthentication, ErrorTypeAuthorization,
		ErrorTypeNotFound, ErrorTypeConflict:
		return false
	default:
		return false
	}
}

// WrapError wraps an existing error as an AppError
func WrapError(err error, errorType ErrorType, code, message string) *AppError {
	if err == nil {
		return nil
	}

	// If it's already an AppError, return it as-is or wrap it
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}

	httpStatus := getHTTPStatusForType(errorType)
	return NewAppErrorWithCause(errorType, code, message, httpStatus, err)
}

// getHTTPStatusForType returns the default HTTP status for an error type
func getHTTPStatusForType(errorType ErrorType) int {
	switch errorType {
	case ErrorTypeValidation:
		return http.StatusBadRequest
	case ErrorTypeAuthentication:
		return http.StatusUnauthorized
	case ErrorTypeAuthorization:
		return http.StatusForbidden
	case ErrorTypeNotFound:
		return http.StatusNotFound
	case ErrorTypeConflict:
		return http.StatusConflict
	case ErrorTypeRateLimit:
		return http.StatusTooManyRequests
	case ErrorTypeTimeout:
		return http.StatusRequestTimeout
	case ErrorTypeUnavailable:
		return http.StatusServiceUnavailable
	case ErrorTypeExternal:
		return http.StatusBadGateway
	case ErrorTypeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
