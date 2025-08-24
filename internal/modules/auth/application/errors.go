package application

import (
	"fmt"
	"net/http"
)

// AuthError represents an authentication-related error
type AuthError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
	Field   string `json:"field,omitempty"`
}

// Error implements the error interface
func (e *AuthError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return e.Message
}

// HTTPStatusCode returns the appropriate HTTP status code for the error
func (e *AuthError) HTTPStatusCode() int {
	switch e.Type {
	case "validation":
		return http.StatusBadRequest
	case "authentication":
		return http.StatusUnauthorized
	case "authorization":
		return http.StatusForbidden
	case "not_found":
		return http.StatusNotFound
	case "rate_limit":
		return http.StatusTooManyRequests
	case "internal":
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// Error type constants
const (
	ErrorTypeValidation     = "validation"
	ErrorTypeAuthentication = "authentication"
	ErrorTypeAuthorization  = "authorization"
	ErrorTypeNotFound       = "not_found"
	ErrorTypeRateLimit      = "rate_limit"
	ErrorTypeInternal       = "internal"
)

// Error code constants
const (
	ErrorCodeInvalidCredentials = "INVALID_CREDENTIALS"
	ErrorCodeUserNotFound       = "USER_NOT_FOUND"
	ErrorCodeSessionNotFound    = "SESSION_NOT_FOUND"
	ErrorCodeSessionExpired     = "SESSION_EXPIRED"
	ErrorCodeSessionInvalid     = "SESSION_INVALID"
	ErrorCodeUserAlreadyExists  = "USER_ALREADY_EXISTS"
	ErrorCodePasswordTooWeak    = "PASSWORD_TOO_WEAK"
	ErrorCodeRateLimitExceeded  = "RATE_LIMIT_EXCEEDED"
	ErrorCodeAccountLocked      = "ACCOUNT_LOCKED"
	ErrorCodeAccountSuspended   = "ACCOUNT_SUSPENDED"
	ErrorCodeValidationFailed   = "VALIDATION_FAILED"
	ErrorCodeInternalError      = "INTERNAL_ERROR"
)

// NewValidationError creates a new validation error
func NewValidationError(field, message string) *AuthError {
	return &AuthError{
		Code:    ErrorCodeValidationFailed,
		Message: message,
		Type:    ErrorTypeValidation,
		Field:   field,
	}
}

// NewInvalidCredentialsError creates a new invalid credentials error
func NewInvalidCredentialsError() *AuthError {
	return &AuthError{
		Code:    ErrorCodeInvalidCredentials,
		Message: "Invalid email or password",
		Type:    ErrorTypeAuthentication,
	}
}

// NewUserNotFoundError creates a new user not found error
func NewUserNotFoundError(identifier string) *AuthError {
	return &AuthError{
		Code:    ErrorCodeUserNotFound,
		Message: fmt.Sprintf("User not found: %s", identifier),
		Type:    ErrorTypeNotFound,
	}
}

// NewSessionNotFoundError creates a new session not found error
func NewSessionNotFoundError(sessionID string) *AuthError {
	return &AuthError{
		Code:    ErrorCodeSessionNotFound,
		Message: "Session not found or invalid",
		Type:    ErrorTypeAuthentication,
	}
}

// NewSessionExpiredError creates a new session expired error
func NewSessionExpiredError() *AuthError {
	return &AuthError{
		Code:    ErrorCodeSessionExpired,
		Message: "Session has expired, please login again",
		Type:    ErrorTypeAuthentication,
	}
}

// NewSessionInvalidError creates a new session invalid error
func NewSessionInvalidError() *AuthError {
	return &AuthError{
		Code:    ErrorCodeSessionInvalid,
		Message: "Session is invalid",
		Type:    ErrorTypeAuthentication,
	}
}

// NewUserAlreadyExistsError creates a new user already exists error
func NewUserAlreadyExistsError(email string) *AuthError {
	return &AuthError{
		Code:    ErrorCodeUserAlreadyExists,
		Message: "A user with this email already exists",
		Type:    ErrorTypeValidation,
		Field:   "email",
	}
}

// NewPasswordTooWeakError creates a new password too weak error
func NewPasswordTooWeakError(message string) *AuthError {
	return &AuthError{
		Code:    ErrorCodePasswordTooWeak,
		Message: message,
		Type:    ErrorTypeValidation,
		Field:   "password",
	}
}

// NewRateLimitExceededError creates a new rate limit exceeded error
func NewRateLimitExceededError(message string) *AuthError {
	return &AuthError{
		Code:    ErrorCodeRateLimitExceeded,
		Message: message,
		Type:    ErrorTypeRateLimit,
	}
}

// NewAccountLockedError creates a new account locked error
func NewAccountLockedError() *AuthError {
	return &AuthError{
		Code:    ErrorCodeAccountLocked,
		Message: "Account is temporarily locked due to too many failed login attempts",
		Type:    ErrorTypeAuthentication,
	}
}

// NewAccountSuspendedError creates a new account suspended error
func NewAccountSuspendedError() *AuthError {
	return &AuthError{
		Code:    ErrorCodeAccountSuspended,
		Message: "Account has been suspended",
		Type:    ErrorTypeAuthorization,
	}
}

// NewInternalError creates a new internal error
func NewInternalError(message string) *AuthError {
	return &AuthError{
		Code:    ErrorCodeInternalError,
		Message: message,
		Type:    ErrorTypeInternal,
	}
}

// IsValidationError checks if the error is a validation error
func IsValidationError(err error) bool {
	if authErr, ok := err.(*AuthError); ok {
		return authErr.Type == ErrorTypeValidation
	}
	return false
}

// IsAuthenticationError checks if the error is an authentication error
func IsAuthenticationError(err error) bool {
	if authErr, ok := err.(*AuthError); ok {
		return authErr.Type == ErrorTypeAuthentication
	}
	return false
}

// IsRateLimitError checks if the error is a rate limit error
func IsRateLimitError(err error) bool {
	if authErr, ok := err.(*AuthError); ok {
		return authErr.Type == ErrorTypeRateLimit
	}
	return false
}
