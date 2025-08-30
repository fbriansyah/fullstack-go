package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-templ-template/internal/shared/errors"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// TestErrorHandler tests the error handling middleware
func TestErrorHandler(t *testing.T) {
	tests := []struct {
		name           string
		config         ErrorHandlerConfig
		handler        echo.HandlerFunc
		expectedStatus int
		expectedBody   string
		contentType    string
	}{
		{
			name:   "handles AppError with JSON response",
			config: ErrorHandlerConfig{JSONAPIErrors: true, LogErrors: false},
			handler: func(c echo.Context) error {
				return errors.NewValidationError("TEST_ERROR", "Test validation error")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "TEST_ERROR",
			contentType:    "application/json",
		},
		{
			name:   "handles generic error with JSON response",
			config: ErrorHandlerConfig{JSONAPIErrors: true, LogErrors: false},
			handler: func(c echo.Context) error {
				return fmt.Errorf("generic error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "INTERNAL_ERROR",
			contentType:    "application/json",
		},
		{
			name:   "handles Echo HTTP error",
			config: ErrorHandlerConfig{JSONAPIErrors: true, LogErrors: false},
			handler: func(c echo.Context) error {
				return echo.NewHTTPError(http.StatusNotFound, "Not found")
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "HTTP_404",
			contentType:    "application/json",
		},
		{
			name:   "handles HTML error pages",
			config: ErrorHandlerConfig{CustomErrorPages: true, JSONAPIErrors: false, LogErrors: false},
			handler: func(c echo.Context) error {
				return errors.NewNotFoundError("PAGE_NOT_FOUND", "Page not found")
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404",
			contentType:    "text/html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.contentType == "application/json" {
				req.Header.Set("Accept", "application/json")
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			middleware := ErrorHandler(tt.config)
			handler := middleware(tt.handler)

			err := handler(c)

			// The middleware should handle the error, so no error should be returned
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedBody)
		})
	}
}

// TestCustomErrorHandler tests the custom Echo error handler
func TestCustomErrorHandler(t *testing.T) {
	tests := []struct {
		name           string
		config         ErrorHandlerConfig
		error          error
		expectedStatus int
		expectedBody   string
		accept         string
	}{
		{
			name:           "handles AppError with JSON",
			config:         ErrorHandlerConfig{JSONAPIErrors: true, LogErrors: false},
			error:          errors.NewValidationError("VALIDATION_FAILED", "Validation failed"),
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "VALIDATION_FAILED",
			accept:         "application/json",
		},
		{
			name:           "handles Echo HTTP error with JSON",
			config:         ErrorHandlerConfig{JSONAPIErrors: true, LogErrors: false},
			error:          echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized"),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "HTTP_401",
			accept:         "application/json",
		},
		{
			name:           "handles generic error with HTML",
			config:         ErrorHandlerConfig{CustomErrorPages: true, JSONAPIErrors: false, LogErrors: false},
			error:          fmt.Errorf("generic error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "500",
			accept:         "text/html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Accept", tt.accept)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := CustomErrorHandler(tt.config)
			handler(tt.error, c)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedBody)
		})
	}
}

// TestIsAPIRequest tests the API request detection
func TestIsAPIRequest(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		accept      string
		contentType string
		xRequested  string
		expected    bool
	}{
		{
			name:     "API path",
			path:     "/api/users",
			expected: true,
		},
		{
			name:     "JSON Accept header",
			path:     "/users",
			accept:   "application/json",
			expected: true,
		},
		{
			name:        "JSON Content-Type",
			path:        "/users",
			contentType: "application/json",
			expected:    true,
		},
		{
			name:       "AJAX request",
			path:       "/users",
			xRequested: "XMLHttpRequest",
			expected:   true,
		},
		{
			name:     "Regular HTML request",
			path:     "/users",
			accept:   "text/html",
			expected: false,
		},
		{
			name:     "No headers",
			path:     "/users",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.accept != "" {
				req.Header.Set("Accept", tt.accept)
			}
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			if tt.xRequested != "" {
				req.Header.Set("X-Requested-With", tt.xRequested)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			result := isAPIRequest(c)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestConvertEchoError tests Echo error conversion
func TestConvertEchoError(t *testing.T) {
	tests := []struct {
		name         string
		echoError    *echo.HTTPError
		expectedType errors.ErrorType
		expectedCode string
	}{
		{
			name:         "400 Bad Request",
			echoError:    echo.NewHTTPError(http.StatusBadRequest, "Bad request"),
			expectedType: errors.ErrorTypeValidation,
			expectedCode: "HTTP_400",
		},
		{
			name:         "401 Unauthorized",
			echoError:    echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized"),
			expectedType: errors.ErrorTypeAuthentication,
			expectedCode: "HTTP_401",
		},
		{
			name:         "403 Forbidden",
			echoError:    echo.NewHTTPError(http.StatusForbidden, "Forbidden"),
			expectedType: errors.ErrorTypeAuthorization,
			expectedCode: "HTTP_403",
		},
		{
			name:         "404 Not Found",
			echoError:    echo.NewHTTPError(http.StatusNotFound, "Not found"),
			expectedType: errors.ErrorTypeNotFound,
			expectedCode: "HTTP_404",
		},
		{
			name:         "409 Conflict",
			echoError:    echo.NewHTTPError(http.StatusConflict, "Conflict"),
			expectedType: errors.ErrorTypeConflict,
			expectedCode: "HTTP_409",
		},
		{
			name:         "429 Too Many Requests",
			echoError:    echo.NewHTTPError(http.StatusTooManyRequests, "Rate limited"),
			expectedType: errors.ErrorTypeRateLimit,
			expectedCode: "HTTP_429",
		},
		{
			name:         "500 Internal Server Error",
			echoError:    echo.NewHTTPError(http.StatusInternalServerError, "Internal error"),
			expectedType: errors.ErrorTypeInternal,
			expectedCode: "HTTP_500",
		},
		{
			name:         "503 Service Unavailable",
			echoError:    echo.NewHTTPError(http.StatusServiceUnavailable, "Unavailable"),
			expectedType: errors.ErrorTypeUnavailable,
			expectedCode: "HTTP_503",
		},
		{
			name:         "504 Gateway Timeout",
			echoError:    echo.NewHTTPError(http.StatusGatewayTimeout, "Timeout"),
			expectedType: errors.ErrorTypeTimeout,
			expectedCode: "HTTP_504",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := convertEchoError(tt.echoError)

			assert.Equal(t, tt.expectedType, appErr.Type)
			assert.Equal(t, tt.expectedCode, appErr.Code)
			assert.Equal(t, tt.echoError.Code, appErr.HTTPStatus)
			assert.Contains(t, appErr.Message, fmt.Sprintf("%v", tt.echoError.Message))
		})
	}
}

// TestGetErrorTitle tests error title generation
func TestGetErrorTitle(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   string
	}{
		{http.StatusBadRequest, "Bad Request"},
		{http.StatusUnauthorized, "Authentication Required"},
		{http.StatusForbidden, "Access Forbidden"},
		{http.StatusNotFound, "Page Not Found"},
		{http.StatusMethodNotAllowed, "Method Not Allowed"},
		{http.StatusConflict, "Conflict"},
		{http.StatusUnprocessableEntity, "Validation Error"},
		{http.StatusTooManyRequests, "Too Many Requests"},
		{http.StatusInternalServerError, "Server Error"},
		{http.StatusBadGateway, "Bad Gateway"},
		{http.StatusServiceUnavailable, "Service Unavailable"},
		{http.StatusGatewayTimeout, "Gateway Timeout"},
		{999, "Error"}, // Unknown status code
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("status_%d", tt.statusCode), func(t *testing.T) {
			result := getErrorTitle(tt.statusCode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestNotFoundHandler tests the 404 handler
func TestNotFoundHandler(t *testing.T) {
	tests := []struct {
		name           string
		accept         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "JSON request",
			accept:         "application/json",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "ROUTE_NOT_FOUND",
		},
		{
			name:           "HTML request",
			accept:         "text/html",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
			req.Header.Set("Accept", tt.accept)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := NotFoundHandler()
			err := handler(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedBody)
		})
	}
}

// TestMethodNotAllowedHandler tests the method not allowed handler
func TestMethodNotAllowedHandler(t *testing.T) {
	tests := []struct {
		name           string
		accept         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "JSON request",
			accept:         "application/json",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "METHOD_NOT_ALLOWED",
		},
		{
			name:           "HTML request",
			accept:         "text/html",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "405",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/get-only-endpoint", nil)
			req.Header.Set("Accept", tt.accept)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := MethodNotAllowedHandler()
			err := handler(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedBody)
		})
	}
}

// TestRecoveryHandler tests the panic recovery handler
func TestRecoveryHandler(t *testing.T) {
	tests := []struct {
		name           string
		handler        echo.HandlerFunc
		config         ErrorHandlerConfig
		expectedStatus int
		expectedBody   string
		accept         string
	}{
		{
			name: "recovers from string panic with JSON",
			handler: func(c echo.Context) error {
				panic("test panic")
			},
			config:         ErrorHandlerConfig{JSONAPIErrors: true, LogErrors: false},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "PANIC_RECOVERED",
			accept:         "application/json",
		},
		{
			name: "recovers from error panic with JSON",
			handler: func(c echo.Context) error {
				panic(fmt.Errorf("test error panic"))
			},
			config:         ErrorHandlerConfig{JSONAPIErrors: true, LogErrors: false},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "PANIC_RECOVERED",
			accept:         "application/json",
		},
		{
			name: "recovers from panic with HTML",
			handler: func(c echo.Context) error {
				panic("test panic")
			},
			config:         ErrorHandlerConfig{CustomErrorPages: true, JSONAPIErrors: false, LogErrors: false},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "500",
			accept:         "text/html",
		},
		{
			name: "normal execution without panic",
			handler: func(c echo.Context) error {
				return c.String(http.StatusOK, "success")
			},
			config:         ErrorHandlerConfig{LogErrors: false},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
			accept:         "text/html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Accept", tt.accept)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			middleware := RecoveryHandler(tt.config)
			handler := middleware(tt.handler)

			err := handler(c)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedBody)
		})
	}
}

// TestDefaultErrorHandlerConfig tests the default configuration
func TestDefaultErrorHandlerConfig(t *testing.T) {
	config := DefaultErrorHandlerConfig()

	assert.False(t, config.ShowStackTrace)
	assert.True(t, config.LogErrors)
	assert.True(t, config.CustomErrorPages)
	assert.True(t, config.JSONAPIErrors)
}

// TestErrorHandlerIntegration tests the error handler in an integrated scenario
func TestErrorHandlerIntegration(t *testing.T) {
	e := echo.New()

	// Configure error handling
	config := ErrorHandlerConfig{
		ShowStackTrace:   false,
		LogErrors:        false, // Disable for testing
		CustomErrorPages: true,
		JSONAPIErrors:    true,
	}

	e.Use(ErrorHandler(config))
	e.HTTPErrorHandler = CustomErrorHandler(config)

	// Add test routes
	e.GET("/api/error", func(c echo.Context) error {
		return errors.NewValidationError("API_ERROR", "API validation error")
	})

	e.GET("/html/error", func(c echo.Context) error {
		return errors.NewNotFoundError("PAGE_ERROR", "Page not found error")
	})

	e.GET("/panic", func(c echo.Context) error {
		panic("test panic")
	})

	tests := []struct {
		name           string
		path           string
		accept         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "API error returns JSON",
			path:           "/api/error",
			accept:         "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "API_ERROR",
		},
		{
			name:           "HTML error returns error page",
			path:           "/html/error",
			accept:         "text/html",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404",
		},
		{
			name:           "Nonexistent route returns 404",
			path:           "/nonexistent",
			accept:         "text/html",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			req.Header.Set("Accept", tt.accept)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedBody)
		})
	}
}

// BenchmarkErrorHandler benchmarks the error handler middleware
func BenchmarkErrorHandler(b *testing.B) {
	e := echo.New()
	config := ErrorHandlerConfig{
		ShowStackTrace:   false,
		LogErrors:        false,
		CustomErrorPages: false,
		JSONAPIErrors:    true,
	}

	middleware := ErrorHandler(config)
	handler := middleware(func(c echo.Context) error {
		return errors.NewValidationError("BENCH_ERROR", "Benchmark error")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept", "application/json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		handler(c)
	}
}

// BenchmarkIsAPIRequest benchmarks the API request detection
func BenchmarkIsAPIRequest(b *testing.B) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isAPIRequest(c)
	}
}
