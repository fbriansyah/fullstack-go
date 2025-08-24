package application

import (
	"fmt"
)

// ApplicationError represents an error that occurred in the application layer
type ApplicationError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

// Error implements the error interface
func (e *ApplicationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s (field: %s)", e.Code, e.Message, e.Field)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Error codes
const (
	ErrCodeValidation        = "VALIDATION_ERROR"
	ErrCodeUserNotFound      = "USER_NOT_FOUND"
	ErrCodeUserAlreadyExists = "USER_ALREADY_EXISTS"
	ErrCodeInvalidPassword   = "INVALID_PASSWORD"
	ErrCodeOptimisticLock    = "OPTIMISTIC_LOCK_ERROR"
	ErrCodeBusinessRule      = "BUSINESS_RULE_VIOLATION"
	ErrCodeInternal          = "INTERNAL_ERROR"
)

// NewValidationError creates a new validation error
func NewValidationError(field, message string) *ApplicationError {
	return &ApplicationError{
		Code:    ErrCodeValidation,
		Message: message,
		Field:   field,
	}
}

// NewUserNotFoundError creates a new user not found error
func NewUserNotFoundError(id string) *ApplicationError {
	return &ApplicationError{
		Code:    ErrCodeUserNotFound,
		Message: fmt.Sprintf("user with ID '%s' not found", id),
	}
}

// NewUserAlreadyExistsError creates a new user already exists error
func NewUserAlreadyExistsError(email string) *ApplicationError {
	return &ApplicationError{
		Code:    ErrCodeUserAlreadyExists,
		Message: fmt.Sprintf("user with email '%s' already exists", email),
	}
}

// NewInvalidPasswordError creates a new invalid password error
func NewInvalidPasswordError() *ApplicationError {
	return &ApplicationError{
		Code:    ErrCodeInvalidPassword,
		Message: "invalid password provided",
	}
}

// NewOptimisticLockError creates a new optimistic lock error
func NewOptimisticLockError(id string) *ApplicationError {
	return &ApplicationError{
		Code:    ErrCodeOptimisticLock,
		Message: fmt.Sprintf("user with ID '%s' has been modified by another process", id),
	}
}

// NewBusinessRuleError creates a new business rule violation error
func NewBusinessRuleError(message string) *ApplicationError {
	return &ApplicationError{
		Code:    ErrCodeBusinessRule,
		Message: message,
	}
}

// NewInternalError creates a new internal error
func NewInternalError(message string) *ApplicationError {
	return &ApplicationError{
		Code:    ErrCodeInternal,
		Message: message,
	}
}

// IsValidationError checks if the error is a validation error
func IsValidationError(err error) bool {
	if appErr, ok := err.(*ApplicationError); ok {
		return appErr.Code == ErrCodeValidation
	}
	return false
}

// IsUserNotFoundError checks if the error is a user not found error
func IsUserNotFoundError(err error) bool {
	if appErr, ok := err.(*ApplicationError); ok {
		return appErr.Code == ErrCodeUserNotFound
	}
	return false
}

// IsUserAlreadyExistsError checks if the error is a user already exists error
func IsUserAlreadyExistsError(err error) bool {
	if appErr, ok := err.(*ApplicationError); ok {
		return appErr.Code == ErrCodeUserAlreadyExists
	}
	return false
}

// IsOptimisticLockError checks if the error is an optimistic lock error
func IsOptimisticLockError(err error) bool {
	if appErr, ok := err.(*ApplicationError); ok {
		return appErr.Code == ErrCodeOptimisticLock
	}
	return false
}
