package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorMiddleware_Handler(t *testing.T) {
	tests := []struct {
		name           string
		handler        echo.HandlerFunc
		config         ErrorMiddlewareConfig
		expectedStatus int
		expectedCode   string
		expectedType   string
	}{
		{
			name: "handles AppError correctly",
			handler: func(c echo.Context) error {
				return NewValidationError("INVALID_INPUT", "Invalid input provided")
			},
			config: ErrorMiddlewareConfig{
				IncludeStackTrace:  false,
				LogAllErrors:       true,
				LogRequestDetails:  true,
				HideInternalErrors: false, // Don't hide internal errors for testing
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_INPUT",
			expectedType:   "validation",
		},
		{
			name: "handles Echo HTTP error",
			handler: func(c echo.Context) error {
				return echo.NewHTTPError(http.StatusNotFound, "Resource not found")
			},
			config: ErrorMiddlewareConfig{
				IncludeStackTrace:  false,
				LogAllErrors:       true,
				LogRequestDetails:  true,
				HideInternalErrors: false, // Don't hide internal errors for testing
			},
			expectedStatus: http.StatusNotFound,
			expectedCode:   "NOT_FOUND",
			expectedType:   "not_found",
		},
		{
			name: "handles standard error",
			handler: func(c echo.Context) error {
				return fmt.Errorf("standard error")
			},
			config: ErrorMiddlewareConfig{
				IncludeStackTrace:  false,
				LogAllErrors:       true,
				LogRequestDetails:  true,
				HideInternalErrors: false, // Don't hide internal errors for testing
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "UNHANDLED_ERROR",
			expectedType:   "internal",
		},
		{
			name: "handles context cancellation",
			handler: func(c echo.Context) error {
				return context.Canceled
			},
			config: ErrorMiddlewareConfig{
				IncludeStackTrace:  false,
				LogAllErrors:       true,
				LogRequestDetails:  true,
				HideInternalErrors: false, // Don't hide internal errors for testing
			},
			expectedStatus: http.StatusRequestTimeout,
			expectedCode:   "REQUEST_CANCELED",
			expectedType:   "timeout",
		},
		{
			name: "handles context deadline exceeded",
			handler: func(c echo.Context) error {
				return context.DeadlineExceeded
			},
			config: ErrorMiddlewareConfig{
				IncludeStackTrace:  false,
				LogAllErrors:       true,
				LogRequestDetails:  true,
				HideInternalErrors: false, // Don't hide internal errors for testing
			},
			expectedStatus: http.StatusRequestTimeout,
			expectedCode:   "REQUEST_TIMEOUT",
			expectedType:   "timeout",
		},
		{
			name: "hides internal errors in production",
			handler: func(c echo.Context) error {
				return NewInternalError("DATABASE_CONNECTION_FAILED", "Failed to connect to database")
			},
			config: ErrorMiddlewareConfig{
				HideInternalErrors: true,
				LogAllErrors:       true,
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "INTERNAL_ERROR",
			expectedType:   "internal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test logger
			logger := logrus.New()
			logger.SetLevel(logrus.DebugLevel)
			structuredLogger := NewLogrusLogger(logger)

			// Create middleware
			middleware := NewErrorMiddleware(structuredLogger, tt.config)

			// Create Echo instance
			e := echo.New()
			e.Use(middleware.Handler())

			// Create test request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Execute handler with middleware
			middlewareHandler := middleware.Handler()(tt.handler)
			err := middlewareHandler(c)

			// Verify response
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Parse response body
			var response map[string]interface{}
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			// Verify error structure
			errorData, exists := response["error"]
			require.True(t, exists, "Response should contain error field")

			errorMap, ok := errorData.(map[string]interface{})
			require.True(t, ok, "Error should be a map")

			assert.Equal(t, tt.expectedCode, errorMap["code"])
			assert.Equal(t, tt.expectedType, errorMap["type"])
			assert.NotEmpty(t, errorMap["id"])
			assert.NotEmpty(t, errorMap["timestamp"])
		})
	}
}

func TestErrorMiddleware_RecoveryHandler(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	structuredLogger := NewLogrusLogger(logger)

	// Create middleware
	config := DefaultErrorMiddlewareConfig()
	config.IncludeStackTrace = true
	middleware := NewErrorMiddleware(structuredLogger, config)

	// Create Echo instance
	e := echo.New()
	e.Use(middleware.RecoveryHandler())

	// Create handler that panics
	handler := func(c echo.Context) error {
		panic("test panic")
	}

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute handler with recovery middleware (should recover from panic)
	recoveryHandler := middleware.RecoveryHandler()(handler)
	err := recoveryHandler(c)
	assert.NoError(t, err) // Recovery middleware should handle the panic

	// Verify response
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	// Parse response body
	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify error structure
	errorData, exists := response["error"]
	require.True(t, exists)

	errorMap, ok := errorData.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "PANIC_RECOVERED", errorMap["code"])
	assert.Equal(t, "internal", errorMap["type"])
}

func TestErrorMiddleware_WithUserContext(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	structuredLogger := NewLogrusLogger(logger)

	// Create middleware with config that doesn't hide internal errors
	config := ErrorMiddlewareConfig{
		IncludeStackTrace:  false,
		LogAllErrors:       true,
		LogRequestDetails:  true,
		HideInternalErrors: false,
	}
	middleware := NewErrorMiddleware(structuredLogger, config)

	// Create Echo instance
	e := echo.New()
	e.Use(middleware.Handler())

	// Create handler that returns error
	handler := func(c echo.Context) error {
		// Set user context
		c.Set("user", map[string]interface{}{
			"id": "user123",
		})
		c.Set("session", map[string]interface{}{
			"id": "session456",
		})

		return NewValidationError("TEST_ERROR", "Test error with context")
	}

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute handler
	err := middleware.Handler()(handler)(c)
	assert.NoError(t, err)

	// Verify response contains request ID in header
	requestID := rec.Header().Get(echo.HeaderXRequestID)
	assert.NotEmpty(t, requestID)

	// Parse response body
	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify request ID in response body
	if requestIDInResponse, exists := response["request_id"]; exists {
		assert.Equal(t, requestID, requestIDInResponse)
	} else {
		// If not in response body, it should at least be in header
		assert.NotEmpty(t, requestID)
	}
}

func TestErrorMiddleware_CustomErrorHandler(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	structuredLogger := NewLogrusLogger(logger)

	// Create custom error handler
	customHandlerCalled := false
	config := DefaultErrorMiddlewareConfig()
	config.CustomErrorHandler = func(c echo.Context, err error) error {
		customHandlerCalled = true
		return nil
	}

	// Create middleware
	middleware := NewErrorMiddleware(structuredLogger, config)

	// Create Echo instance
	e := echo.New()
	e.Use(middleware.Handler())

	// Create handler that returns error
	handler := func(c echo.Context) error {
		return NewValidationError("TEST_ERROR", "Test error")
	}

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute handler
	err := middleware.Handler()(handler)(c)
	assert.NoError(t, err)

	// Verify custom handler was called
	assert.True(t, customHandlerCalled)
}

func TestErrorMiddleware_LoggingLevels(t *testing.T) {
	tests := []struct {
		name     string
		error    *AppError
		config   ErrorMiddlewareConfig
		logLevel string
	}{
		{
			name:     "logs critical errors as error",
			error:    NewInternalError("CRITICAL_ERROR", "Critical error").WithDetails(map[string]interface{}{"severity": SeverityCritical}),
			config:   DefaultErrorMiddlewareConfig(),
			logLevel: "error",
		},
		{
			name:     "logs high severity errors as error",
			error:    NewInternalError("HIGH_ERROR", "High severity error"),
			config:   DefaultErrorMiddlewareConfig(),
			logLevel: "error",
		},
		{
			name:     "logs medium severity errors as warning",
			error:    NewConflictError("MEDIUM_ERROR", "Medium severity error"),
			config:   DefaultErrorMiddlewareConfig(),
			logLevel: "warning",
		},
		{
			name:     "logs low severity errors as info when configured",
			error:    NewValidationError("LOW_ERROR", "Low severity error"),
			config:   ErrorMiddlewareConfig{LogAllErrors: true},
			logLevel: "info",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test logger with custom hook to capture log entries
			logger := logrus.New()
			logger.SetLevel(logrus.DebugLevel)

			var capturedEntry *logrus.Entry
			logger.AddHook(&testHook{
				callback: func(entry *logrus.Entry) {
					capturedEntry = entry
				},
			})

			structuredLogger := NewLogrusLogger(logger)

			// Create middleware
			middleware := NewErrorMiddleware(structuredLogger, tt.config)

			// Create Echo instance
			e := echo.New()
			e.Use(middleware.Handler())

			// Create handler that returns the test error
			handler := func(c echo.Context) error {
				return tt.error
			}

			// Create test request
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Execute handler
			err := middleware.Handler()(handler)(c)
			assert.NoError(t, err)

			// Verify log entry was captured
			require.NotNil(t, capturedEntry, "Expected log entry to be captured")

			// Verify log level matches expected
			switch tt.logLevel {
			case "error":
				assert.Equal(t, logrus.ErrorLevel, capturedEntry.Level)
			case "warning":
				assert.Equal(t, logrus.WarnLevel, capturedEntry.Level)
			case "info":
				assert.Equal(t, logrus.InfoLevel, capturedEntry.Level)
			}

			// Verify error fields are present
			assert.Equal(t, tt.error.ID, capturedEntry.Data["error_id"])
			assert.Equal(t, tt.error.Code, capturedEntry.Data["error_code"])
			assert.Equal(t, tt.error.Type, capturedEntry.Data["error_type"])
		})
	}
}

// testHook is a logrus hook for testing
type testHook struct {
	callback func(*logrus.Entry)
}

func (h *testHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *testHook) Fire(entry *logrus.Entry) error {
	if h.callback != nil {
		h.callback(entry)
	}
	return nil
}

func TestErrorMiddleware_SensitiveDataFiltering(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	structuredLogger := NewLogrusLogger(logger)

	// Create middleware
	middleware := NewErrorMiddleware(structuredLogger, DefaultErrorMiddlewareConfig())

	// Create Echo instance
	e := echo.New()
	e.Use(middleware.Handler())

	// Create handler that returns error with sensitive data
	handler := func(c echo.Context) error {
		return NewValidationError("TEST_ERROR", "Test error").WithDetails(map[string]interface{}{
			"username": "testuser",
			"password": "secret123", // This should be filtered
			"token":    "abc123",    // This should be filtered
			"email":    "test@example.com",
		})
	}

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute handler
	err := middleware.Handler()(handler)(c)
	assert.NoError(t, err)

	// Parse response body
	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify sensitive data is filtered from response
	errorData := response["error"].(map[string]interface{})
	if details, exists := errorData["details"]; exists {
		detailsMap := details.(map[string]interface{})
		assert.Contains(t, detailsMap, "username")
		assert.Contains(t, detailsMap, "email")
		assert.NotContains(t, detailsMap, "password")
		assert.NotContains(t, detailsMap, "token")
	}
}

func TestErrorList_HTTPResponse(t *testing.T) {
	// Create error list
	errorList := &ErrorList{}
	errorList.Add(NewValidationError("FIELD1_INVALID", "Field 1 is invalid"))
	errorList.Add(NewValidationError("FIELD2_INVALID", "Field 2 is invalid"))

	// Test HTTP response
	response := errorList.ToHTTPResponse()
	assert.Contains(t, response, "errors")

	errors := response["errors"].([]interface{})
	assert.Len(t, errors, 2)

	// Verify each error has proper structure
	for _, errorData := range errors {
		errorMap := errorData.(map[string]interface{})
		assert.Contains(t, errorMap, "id")
		assert.Contains(t, errorMap, "code")
		assert.Contains(t, errorMap, "type")
		assert.Contains(t, errorMap, "message")
		assert.Contains(t, errorMap, "timestamp")
	}

	// Test HTTP status
	status := errorList.GetHTTPStatus()
	assert.Equal(t, http.StatusBadRequest, status)
}

func TestErrorMiddleware_Integration(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Capture log output
	var logOutput strings.Builder
	logger.SetOutput(&logOutput)
	logger.SetFormatter(&logrus.JSONFormatter{})

	structuredLogger := NewLogrusLogger(logger)

	// Create middleware with full configuration
	config := ErrorMiddlewareConfig{
		IncludeStackTrace:  true,
		LogAllErrors:       true,
		LogRequestDetails:  true,
		HideInternalErrors: false,
	}
	middleware := NewErrorMiddleware(structuredLogger, config)

	// Create Echo instance with middleware
	e := echo.New()
	e.Use(middleware.Handler())
	e.Use(middleware.RecoveryHandler())

	// Add test routes
	e.GET("/validation-error", func(c echo.Context) error {
		return NewValidationErrorWithDetails("INVALID_DATA", "Invalid data provided", map[string]interface{}{
			"field": "email",
			"value": "invalid-email",
		})
	})

	e.GET("/auth-error", func(c echo.Context) error {
		return NewSessionExpiredError()
	})

	e.GET("/internal-error", func(c echo.Context) error {
		return NewDatabaseError("user_lookup", fmt.Errorf("connection failed"))
	})

	e.GET("/panic", func(c echo.Context) error {
		panic("test panic for recovery")
	})

	// Test validation error
	t.Run("validation error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/validation-error", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		errorData := response["error"].(map[string]interface{})
		assert.Equal(t, "INVALID_DATA", errorData["code"])
		assert.Equal(t, "validation", errorData["type"])
		assert.Contains(t, errorData, "details")
	})

	// Test authentication error
	t.Run("authentication error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth-error", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		errorData := response["error"].(map[string]interface{})
		assert.Equal(t, "SESSION_EXPIRED", errorData["code"])
		assert.Equal(t, "authentication", errorData["type"])
		assert.Contains(t, errorData, "user_message")
	})

	// Test internal error
	t.Run("internal error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/internal-error", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		errorData := response["error"].(map[string]interface{})
		assert.Equal(t, "DATABASE_ERROR", errorData["code"])
		assert.Equal(t, "internal", errorData["type"])
	})

	// Test panic recovery
	t.Run("panic recovery", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/panic", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		errorData := response["error"].(map[string]interface{})
		assert.Equal(t, "PANIC_RECOVERED", errorData["code"])
		assert.Equal(t, "internal", errorData["type"])
	})

	// Verify logging occurred
	logContent := logOutput.String()
	assert.Contains(t, logContent, "INVALID_DATA")
	assert.Contains(t, logContent, "SESSION_EXPIRED")
	assert.Contains(t, logContent, "DATABASE_ERROR")
	assert.Contains(t, logContent, "PANIC_RECOVERED")
}
