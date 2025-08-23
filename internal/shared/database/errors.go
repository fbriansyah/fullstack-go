package database

import (
	"errors"
	"fmt"
)

// Common database errors
var (
	// ErrNotFound is returned when a requested entity is not found
	ErrNotFound = errors.New("entity not found")

	// ErrDuplicateKey is returned when a unique constraint is violated
	ErrDuplicateKey = errors.New("duplicate key violation")

	// ErrForeignKeyViolation is returned when a foreign key constraint is violated
	ErrForeignKeyViolation = errors.New("foreign key constraint violation")

	// ErrOptimisticLock is returned when an optimistic locking conflict occurs
	ErrOptimisticLock = errors.New("optimistic locking conflict")

	// ErrConnectionFailed is returned when database connection fails
	ErrConnectionFailed = errors.New("database connection failed")

	// ErrTransactionFailed is returned when a transaction fails
	ErrTransactionFailed = errors.New("transaction failed")

	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = errors.New("invalid input")
)

// DatabaseError wraps database-specific errors with additional context
type DatabaseError struct {
	Op      string // Operation that failed
	Table   string // Table involved in the operation
	Err     error  // Underlying error
	Code    string // Database-specific error code
	Message string // Human-readable message
}

// Error implements the error interface
func (e *DatabaseError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("database error in %s on table %s: %s", e.Op, e.Table, e.Message)
	}
	return fmt.Sprintf("database error in %s on table %s: %v", e.Op, e.Table, e.Err)
}

// Unwrap returns the underlying error
func (e *DatabaseError) Unwrap() error {
	return e.Err
}

// Is checks if the error matches the target error
func (e *DatabaseError) Is(target error) bool {
	return errors.Is(e.Err, target)
}

// NewDatabaseError creates a new database error
func NewDatabaseError(op, table string, err error) *DatabaseError {
	return &DatabaseError{
		Op:    op,
		Table: table,
		Err:   err,
	}
}

// NewDatabaseErrorWithCode creates a new database error with a specific code
func NewDatabaseErrorWithCode(op, table, code, message string, err error) *DatabaseError {
	return &DatabaseError{
		Op:      op,
		Table:   table,
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// ValidationError represents input validation errors
type ValidationError struct {
	Field   string      `json:"field"`
	Value   interface{} `json:"value"`
	Message string      `json:"message"`
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "validation errors"
	}
	if len(e) == 1 {
		return e[0].Error()
	}
	return fmt.Sprintf("validation errors: %d fields failed validation", len(e))
}

// Add adds a validation error to the collection
func (e *ValidationErrors) Add(field string, value interface{}, message string) {
	*e = append(*e, ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	})
}

// HasErrors returns true if there are validation errors
func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

// IsNotFoundError checks if an error is a "not found" error
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsDuplicateKeyError checks if an error is a duplicate key error
func IsDuplicateKeyError(err error) bool {
	return errors.Is(err, ErrDuplicateKey)
}

// IsForeignKeyViolationError checks if an error is a foreign key violation error
func IsForeignKeyViolationError(err error) bool {
	return errors.Is(err, ErrForeignKeyViolation)
}

// IsOptimisticLockError checks if an error is an optimistic locking error
func IsOptimisticLockError(err error) bool {
	return errors.Is(err, ErrOptimisticLock)
}

// IsConnectionError checks if an error is a connection error
func IsConnectionError(err error) bool {
	return errors.Is(err, ErrConnectionFailed)
}

// IsTransactionError checks if an error is a transaction error
func IsTransactionError(err error) bool {
	return errors.Is(err, ErrTransactionFailed)
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	var validationErr *ValidationError
	var validationErrs ValidationErrors
	return errors.As(err, &validationErr) || errors.As(err, &validationErrs)
}
