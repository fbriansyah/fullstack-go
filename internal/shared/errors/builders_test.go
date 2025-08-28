package errors

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAppError(t *testing.T) {
	errorType := ErrorTypeValidation
	code := "TEST_ERROR"
	message := "Test error message"
	httpStatus := http.StatusBadRequest

	appError := NewAppError(errorType, code, message, httpStatus)

	assert.NotEmpty(t, appError.ID)
	assert.Equal(t, code, appError.Code)
	assert.Equal(t, errorType, appError.Type)
	assert.Equal(t, message, appError.Message)
	assert.Equal(t, httpStatus, appError.HTTPStatus)
	assert.Equal(t, getSeverityForType(errorType), appError.Severity)
	assert.Equal(t, isRetryableByType(errorType), appError.Retryable)
	assert.WithinDuration(t, time.Now(), appError.Timestamp, time.Second)
}

func TestNewAppErrorWithCause(t *testing.T) {
	errorType := ErrorTypeInternal
	code := "TEST_ERROR"
	message := "Test error message"
	httpStatus := http.StatusInternalServerError
	cause := fmt.Errorf("underlying error")

	appError := NewAppErrorWithCause(errorType, code, message, httpStatus, cause)

	assert.NotEmpty(t, appError.ID)
	assert.Equal(t, code, appError.Code)
	assert.Equal(t, errorType, appError.Type)
	assert.Equal(t, message, appError.Message)
	assert.Equal(t, httpStatus, appError.HTTPStatus)
	assert.Equal(t, cause, appError.Cause)
}

func TestValidationErrorBuilders(t *testing.T) {
	t.Run("NewValidationError", func(t *testing.T) {
		err := NewValidationError("INVALID_INPUT", "Invalid input provided")

		assert.Equal(t, ErrorTypeValidation, err.Type)
		assert.Equal(t, "INVALID_INPUT", err.Code)
		assert.Equal(t, "Invalid input provided", err.Message)
		assert.Equal(t, http.StatusBadRequest, err.HTTPStatus)
		assert.Equal(t, SeverityLow, err.Severity)
		assert.False(t, err.Retryable)
	})

	t.Run("NewValidationErrorWithDetails", func(t *testing.T) {
		details := map[string]interface{}{
			"field": "email",
			"value": "invalid-email",
		}

		err := NewValidationErrorWithDetails("INVALID_EMAIL", "Invalid email format", details)

		assert.Equal(t, ErrorTypeValidation, err.Type)
		assert.Equal(t, "INVALID_EMAIL", err.Code)
		assert.Equal(t, details, err.Details)
	})
}

func TestAuthenticationErrorBuilders(t *testing.T) {
	t.Run("NewAuthenticationError", func(t *testing.T) {
		err := NewAuthenticationError("AUTH_FAILED", "Authentication failed")

		assert.Equal(t, ErrorTypeAuthentication, err.Type)
		assert.Equal(t, "AUTH_FAILED", err.Code)
		assert.Equal(t, http.StatusUnauthorized, err.HTTPStatus)
		assert.Equal(t, SeverityLow, err.Severity)
	})

	t.Run("NewInvalidCredentialsError", func(t *testing.T) {
		err := NewInvalidCredentialsError()

		assert.Equal(t, ErrorTypeAuthentication, err.Type)
		assert.Equal(t, "INVALID_CREDENTIALS", err.Code)
		assert.NotEmpty(t, err.UserMessage)
	})

	t.Run("NewSessionExpiredError", func(t *testing.T) {
		err := NewSessionExpiredError()

		assert.Equal(t, ErrorTypeAuthentication, err.Type)
		assert.Equal(t, "SESSION_EXPIRED", err.Code)
		assert.NotEmpty(t, err.UserMessage)
	})

	t.Run("NewSessionInvalidError", func(t *testing.T) {
		err := NewSessionInvalidError()

		assert.Equal(t, ErrorTypeAuthentication, err.Type)
		assert.Equal(t, "SESSION_INVALID", err.Code)
		assert.NotEmpty(t, err.UserMessage)
	})
}

func TestAuthorizationErrorBuilders(t *testing.T) {
	t.Run("NewAuthorizationError", func(t *testing.T) {
		err := NewAuthorizationError("ACCESS_DENIED", "Access denied")

		assert.Equal(t, ErrorTypeAuthorization, err.Type)
		assert.Equal(t, "ACCESS_DENIED", err.Code)
		assert.Equal(t, http.StatusForbidden, err.HTTPStatus)
	})

	t.Run("NewInsufficientPermissionsError", func(t *testing.T) {
		err := NewInsufficientPermissionsError()

		assert.Equal(t, ErrorTypeAuthorization, err.Type)
		assert.Equal(t, "INSUFFICIENT_PERMISSIONS", err.Code)
		assert.NotEmpty(t, err.UserMessage)
	})

	t.Run("NewAccountSuspendedError", func(t *testing.T) {
		err := NewAccountSuspendedError()

		assert.Equal(t, ErrorTypeAuthorization, err.Type)
		assert.Equal(t, "ACCOUNT_SUSPENDED", err.Code)
		assert.NotEmpty(t, err.UserMessage)
	})
}

func TestNotFoundErrorBuilders(t *testing.T) {
	t.Run("NewNotFoundError", func(t *testing.T) {
		err := NewNotFoundError("User", "123")

		assert.Equal(t, ErrorTypeNotFound, err.Type)
		assert.Equal(t, "RESOURCE_NOT_FOUND", err.Code)
		assert.Equal(t, http.StatusNotFound, err.HTTPStatus)
		assert.Contains(t, err.Message, "User")
		assert.Contains(t, err.Message, "123")
		assert.Equal(t, "User", err.Details["resource"])
		assert.Equal(t, "123", err.Details["id"])
		assert.NotEmpty(t, err.UserMessage)
	})

	t.Run("NewUserNotFoundError", func(t *testing.T) {
		err := NewUserNotFoundError("user123")

		assert.Equal(t, ErrorTypeNotFound, err.Type)
		assert.Equal(t, "RESOURCE_NOT_FOUND", err.Code)
		assert.Contains(t, err.Message, "User")
		assert.Contains(t, err.Message, "user123")
	})

	t.Run("NewSessionNotFoundError", func(t *testing.T) {
		err := NewSessionNotFoundError("session456")

		assert.Equal(t, ErrorTypeNotFound, err.Type)
		assert.Equal(t, "RESOURCE_NOT_FOUND", err.Code)
		assert.Contains(t, err.Message, "Session")
		assert.Contains(t, err.Message, "session456")
	})
}

func TestConflictErrorBuilders(t *testing.T) {
	t.Run("NewConflictError", func(t *testing.T) {
		err := NewConflictError("RESOURCE_CONFLICT", "Resource conflict occurred")

		assert.Equal(t, ErrorTypeConflict, err.Type)
		assert.Equal(t, "RESOURCE_CONFLICT", err.Code)
		assert.Equal(t, http.StatusConflict, err.HTTPStatus)
	})

	t.Run("NewDuplicateResourceError", func(t *testing.T) {
		err := NewDuplicateResourceError("User", "email", "test@example.com")

		assert.Equal(t, ErrorTypeConflict, err.Type)
		assert.Equal(t, "DUPLICATE_RESOURCE", err.Code)
		assert.Contains(t, err.Message, "User")
		assert.Contains(t, err.Message, "email")
		assert.Contains(t, err.Message, "test@example.com")
		assert.NotEmpty(t, err.UserMessage)
	})

	t.Run("NewEmailAlreadyExistsError", func(t *testing.T) {
		err := NewEmailAlreadyExistsError("test@example.com")

		assert.Equal(t, ErrorTypeConflict, err.Type)
		assert.Equal(t, "DUPLICATE_RESOURCE", err.Code)
		assert.Contains(t, err.Message, "test@example.com")
	})

	t.Run("NewOptimisticLockError", func(t *testing.T) {
		err := NewOptimisticLockError("User")

		assert.Equal(t, ErrorTypeConflict, err.Type)
		assert.Equal(t, "OPTIMISTIC_LOCK_CONFLICT", err.Code)
		assert.Contains(t, err.Message, "User")
		assert.NotEmpty(t, err.UserMessage)
	})
}

func TestRateLimitErrorBuilders(t *testing.T) {
	t.Run("NewRateLimitError", func(t *testing.T) {
		err := NewRateLimitError(100, "minute")

		assert.Equal(t, ErrorTypeRateLimit, err.Type)
		assert.Equal(t, "RATE_LIMIT_EXCEEDED", err.Code)
		assert.Equal(t, http.StatusTooManyRequests, err.HTTPStatus)
		assert.Contains(t, err.Message, "100")
		assert.Contains(t, err.Message, "minute")
		assert.Equal(t, 100, err.Details["limit"])
		assert.Equal(t, "minute", err.Details["window"])
		assert.NotEmpty(t, err.UserMessage)
		assert.True(t, err.Retryable)
	})
}

func TestInternalErrorBuilders(t *testing.T) {
	t.Run("NewInternalError", func(t *testing.T) {
		err := NewInternalError("SYSTEM_ERROR", "System error occurred")

		assert.Equal(t, ErrorTypeInternal, err.Type)
		assert.Equal(t, "SYSTEM_ERROR", err.Code)
		assert.Equal(t, http.StatusInternalServerError, err.HTTPStatus)
		assert.Equal(t, SeverityHigh, err.Severity)
		assert.NotEmpty(t, err.UserMessage)
	})

	t.Run("NewInternalErrorWithCause", func(t *testing.T) {
		cause := fmt.Errorf("underlying error")
		err := NewInternalErrorWithCause("SYSTEM_ERROR", "System error occurred", cause)

		assert.Equal(t, ErrorTypeInternal, err.Type)
		assert.Equal(t, "SYSTEM_ERROR", err.Code)
		assert.Equal(t, cause, err.Cause)
		assert.NotEmpty(t, err.UserMessage)
	})

	t.Run("NewDatabaseError", func(t *testing.T) {
		cause := fmt.Errorf("connection failed")
		err := NewDatabaseError("user_lookup", cause)

		assert.Equal(t, ErrorTypeInternal, err.Type)
		assert.Equal(t, "DATABASE_ERROR", err.Code)
		assert.Contains(t, err.Message, "user_lookup")
		assert.Equal(t, cause, err.Cause)
		assert.Equal(t, "user_lookup", err.Details["operation"])
	})

	t.Run("NewEventBusError", func(t *testing.T) {
		cause := fmt.Errorf("publish failed")
		err := NewEventBusError("event_publish", cause)

		assert.Equal(t, ErrorTypeInternal, err.Type)
		assert.Equal(t, "EVENT_BUS_ERROR", err.Code)
		assert.Contains(t, err.Message, "event_publish")
		assert.Equal(t, cause, err.Cause)
		assert.Equal(t, "event_publish", err.Details["operation"])
	})
}

func TestExternalServiceErrorBuilders(t *testing.T) {
	t.Run("NewExternalServiceError", func(t *testing.T) {
		cause := fmt.Errorf("service unavailable")
		err := NewExternalServiceError("payment_service", "process_payment", cause)

		assert.Equal(t, ErrorTypeExternal, err.Type)
		assert.Equal(t, "EXTERNAL_SERVICE_ERROR", err.Code)
		assert.Equal(t, http.StatusBadGateway, err.HTTPStatus)
		assert.Contains(t, err.Message, "payment_service")
		assert.Contains(t, err.Message, "process_payment")
		assert.Equal(t, cause, err.Cause)
		assert.Equal(t, "payment_service", err.Details["service"])
		assert.Equal(t, "process_payment", err.Details["operation"])
		assert.NotEmpty(t, err.UserMessage)
		assert.True(t, err.Retryable)
	})
}

func TestTimeoutErrorBuilders(t *testing.T) {
	t.Run("NewTimeoutError", func(t *testing.T) {
		timeout := 30 * time.Second
		err := NewTimeoutError("database_query", timeout)

		assert.Equal(t, ErrorTypeTimeout, err.Type)
		assert.Equal(t, "OPERATION_TIMEOUT", err.Code)
		assert.Equal(t, http.StatusRequestTimeout, err.HTTPStatus)
		assert.Contains(t, err.Message, "database_query")
		assert.Contains(t, err.Message, "30s")
		assert.Equal(t, "database_query", err.Details["operation"])
		assert.Equal(t, timeout.String(), err.Details["timeout"])
		assert.NotEmpty(t, err.UserMessage)
		assert.True(t, err.Retryable)
	})
}

func TestServiceUnavailableErrorBuilders(t *testing.T) {
	t.Run("NewServiceUnavailableError", func(t *testing.T) {
		err := NewServiceUnavailableError("user_service")

		assert.Equal(t, ErrorTypeUnavailable, err.Type)
		assert.Equal(t, "SERVICE_UNAVAILABLE", err.Code)
		assert.Equal(t, http.StatusServiceUnavailable, err.HTTPStatus)
		assert.Contains(t, err.Message, "user_service")
		assert.Equal(t, "user_service", err.Details["service"])
		assert.NotEmpty(t, err.UserMessage)
		assert.True(t, err.Retryable)
	})
}

func TestWrapError(t *testing.T) {
	t.Run("wrap nil error", func(t *testing.T) {
		result := WrapError(nil, ErrorTypeInternal, "TEST_ERROR", "Test message")
		assert.Nil(t, result)
	})

	t.Run("wrap AppError returns as-is", func(t *testing.T) {
		original := NewValidationError("ORIGINAL_ERROR", "Original message")
		result := WrapError(original, ErrorTypeInternal, "WRAPPED_ERROR", "Wrapped message")

		assert.Equal(t, original, result)
	})

	t.Run("wrap standard error", func(t *testing.T) {
		original := fmt.Errorf("standard error")
		result := WrapError(original, ErrorTypeInternal, "WRAPPED_ERROR", "Wrapped message")

		assert.Equal(t, ErrorTypeInternal, result.Type)
		assert.Equal(t, "WRAPPED_ERROR", result.Code)
		assert.Equal(t, "Wrapped message", result.Message)
		assert.Equal(t, original, result.Cause)
		assert.Equal(t, http.StatusInternalServerError, result.HTTPStatus)
	})
}

func TestGetSeverityForType(t *testing.T) {
	tests := []struct {
		errorType        ErrorType
		expectedSeverity ErrorSeverity
	}{
		{ErrorTypeValidation, SeverityLow},
		{ErrorTypeAuthentication, SeverityLow},
		{ErrorTypeAuthorization, SeverityLow},
		{ErrorTypeNotFound, SeverityLow},
		{ErrorTypeConflict, SeverityMedium},
		{ErrorTypeRateLimit, SeverityMedium},
		{ErrorTypeInternal, SeverityHigh},
		{ErrorTypeExternal, SeverityHigh},
		{ErrorTypeTimeout, SeverityHigh},
		{ErrorTypeUnavailable, SeverityHigh},
	}

	for _, tt := range tests {
		t.Run(string(tt.errorType), func(t *testing.T) {
			severity := getSeverityForType(tt.errorType)
			assert.Equal(t, tt.expectedSeverity, severity)
		})
	}
}

func TestIsRetryableByType(t *testing.T) {
	tests := []struct {
		errorType         ErrorType
		expectedRetryable bool
	}{
		{ErrorTypeValidation, false},
		{ErrorTypeAuthentication, false},
		{ErrorTypeAuthorization, false},
		{ErrorTypeNotFound, false},
		{ErrorTypeConflict, false},
		{ErrorTypeRateLimit, true},
		{ErrorTypeInternal, false},
		{ErrorTypeExternal, true},
		{ErrorTypeTimeout, true},
		{ErrorTypeUnavailable, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.errorType), func(t *testing.T) {
			retryable := isRetryableByType(tt.errorType)
			assert.Equal(t, tt.expectedRetryable, retryable)
		})
	}
}

func TestGetHTTPStatusForType(t *testing.T) {
	tests := []struct {
		errorType      ErrorType
		expectedStatus int
	}{
		{ErrorTypeValidation, http.StatusBadRequest},
		{ErrorTypeAuthentication, http.StatusUnauthorized},
		{ErrorTypeAuthorization, http.StatusForbidden},
		{ErrorTypeNotFound, http.StatusNotFound},
		{ErrorTypeConflict, http.StatusConflict},
		{ErrorTypeRateLimit, http.StatusTooManyRequests},
		{ErrorTypeTimeout, http.StatusRequestTimeout},
		{ErrorTypeUnavailable, http.StatusServiceUnavailable},
		{ErrorTypeExternal, http.StatusBadGateway},
		{ErrorTypeInternal, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(string(tt.errorType), func(t *testing.T) {
			status := getHTTPStatusForType(tt.errorType)
			assert.Equal(t, tt.expectedStatus, status)
		})
	}
}

func TestGenerateErrorID(t *testing.T) {
	id1 := generateErrorID()
	id2 := generateErrorID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2) // Should be unique

	// Should be valid UUID format
	assert.Len(t, id1, 36) // UUID length with hyphens
	assert.Contains(t, id1, "-")
}
