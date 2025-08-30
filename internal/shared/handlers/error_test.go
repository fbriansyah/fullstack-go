package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go-templ-template/internal/shared/errors"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestErrorHandlers_RegisterRoutes tests route registration
func TestErrorHandlers_RegisterRoutes(t *testing.T) {
	e := echo.New()
	handlers := NewErrorHandlers()

	handlers.RegisterRoutes(e)

	// Test that routes are registered by making requests
	tests := []struct {
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{"/error/404", http.StatusNotFound, "404"},
		{"/error/500", http.StatusInternalServerError, "500"},
		{"/error/401", http.StatusUnauthorized, "401"},
		{"/error/403", http.StatusForbidden, "403"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedBody)
		})
	}
}

// TestErrorHandlers_Handle404 tests the 404 handler
func TestErrorHandlers_Handle404(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlers := NewErrorHandlers()
	err := handlers.Handle404(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "404")
	assert.Contains(t, rec.Body.String(), "Page Not Found")
}

// TestErrorHandlers_Handle500 tests the 500 handler
func TestErrorHandlers_Handle500(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlers := NewErrorHandlers()
	err := handlers.Handle500(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "500")
	assert.Contains(t, rec.Body.String(), "Server Error")
}

// TestErrorHandlers_HandleAuth tests the authentication handler
func TestErrorHandlers_HandleAuth(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlers := NewErrorHandlers()
	err := handlers.HandleAuth(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "401")
	assert.Contains(t, rec.Body.String(), "Authentication Required")
}

// TestErrorHandlers_HandleForbidden tests the forbidden handler
func TestErrorHandlers_HandleForbidden(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlers := NewErrorHandlers()
	err := handlers.HandleForbidden(c)

	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Contains(t, rec.Body.String(), "403")
	assert.Contains(t, rec.Body.String(), "Access Forbidden")
}

// TestErrorHandlers_HandleGeneric tests the generic error handler
func TestErrorHandlers_HandleGeneric(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "400 Bad Request",
			code:           "400",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Bad Request",
		},
		{
			name:           "401 redirects to auth handler",
			code:           "401",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Authentication Required",
		},
		{
			name:           "403 redirects to forbidden handler",
			code:           "403",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Access Forbidden",
		},
		{
			name:           "404 redirects to 404 handler",
			code:           "404",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Page Not Found",
		},
		{
			name:           "405 Method Not Allowed",
			code:           "405",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method Not Allowed",
		},
		{
			name:           "408 Request Timeout",
			code:           "408",
			expectedStatus: http.StatusRequestTimeout,
			expectedBody:   "Request Timeout",
		},
		{
			name:           "409 Conflict",
			code:           "409",
			expectedStatus: http.StatusConflict,
			expectedBody:   "Conflict",
		},
		{
			name:           "410 Gone",
			code:           "410",
			expectedStatus: http.StatusGone,
			expectedBody:   "Gone",
		},
		{
			name:           "422 Unprocessable Entity",
			code:           "422",
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   "Unprocessable Entity",
		},
		{
			name:           "429 Too Many Requests",
			code:           "429",
			expectedStatus: http.StatusTooManyRequests,
			expectedBody:   "Too Many Requests",
		},
		{
			name:           "500 redirects to 500 handler",
			code:           "500",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Server Error",
		},
		{
			name:           "502 Bad Gateway",
			code:           "502",
			expectedStatus: http.StatusBadGateway,
			expectedBody:   "Bad Gateway",
		},
		{
			name:           "503 Service Unavailable",
			code:           "503",
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   "Service Unavailable",
		},
		{
			name:           "504 Gateway Timeout",
			code:           "504",
			expectedStatus: http.StatusGatewayTimeout,
			expectedBody:   "Gateway Timeout",
		},
		{
			name:           "Unknown error code",
			code:           "999",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("code")
			c.SetParamValues(tt.code)

			handlers := NewErrorHandlers()
			err := handlers.HandleGeneric(c)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedBody)
		})
	}
}

// TestFallbackHandler_Handle404Fallback tests the 404 fallback handler
func TestFallbackHandler_Handle404Fallback(t *testing.T) {
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

			handler := NewFallbackHandler()
			err := handler.Handle404Fallback(c)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedBody)
		})
	}
}

// TestFallbackHandler_HandleMethodNotAllowed tests the method not allowed handler
func TestFallbackHandler_HandleMethodNotAllowed(t *testing.T) {
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
			req := httptest.NewRequest(http.MethodPost, "/get-only", nil)
			req.Header.Set("Accept", tt.accept)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := NewFallbackHandler()
			err := handler.HandleMethodNotAllowed(c)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedBody)
		})
	}
}

// TestIsAPIRequest tests the API request detection function
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

// TestContains tests the contains helper function
func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{
			name:     "contains substring",
			s:        "application/json",
			substr:   "json",
			expected: true,
		},
		{
			name:     "does not contain substring",
			s:        "text/html",
			substr:   "json",
			expected: false,
		},
		{
			name:     "empty substring",
			s:        "test",
			substr:   "",
			expected: true,
		},
		{
			name:     "empty string",
			s:        "",
			substr:   "test",
			expected: false,
		},
		{
			name:     "exact match",
			s:        "test",
			substr:   "test",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.s, tt.substr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIndexOf tests the indexOf helper function
func TestIndexOf(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected int
	}{
		{
			name:     "found at beginning",
			s:        "application/json",
			substr:   "app",
			expected: 0,
		},
		{
			name:     "found in middle",
			s:        "application/json",
			substr:   "json",
			expected: 12,
		},
		{
			name:     "not found",
			s:        "text/html",
			substr:   "json",
			expected: -1,
		},
		{
			name:     "empty substring",
			s:        "test",
			substr:   "",
			expected: 0,
		},
		{
			name:     "empty string",
			s:        "",
			substr:   "test",
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := indexOf(tt.s, tt.substr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestErrorPageRouter_RegisterRoutes tests the error page router
func TestErrorPageRouter_RegisterRoutes(t *testing.T) {
	e := echo.New()
	router := NewErrorPageRouter()

	router.RegisterRoutes(e)

	// Test that error routes are registered
	req := httptest.NewRequest(http.MethodGet, "/error/404", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "404")
}

// TestErrorPageRouter_handleAppError tests AppError handling
func TestErrorPageRouter_handleAppError(t *testing.T) {
	tests := []struct {
		name           string
		appErr         *errors.AppError
		accept         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "404 error with HTML",
			appErr:         errors.NewNotFoundError("NOT_FOUND", "Resource not found"),
			accept:         "text/html",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404",
		},
		{
			name:           "401 error with HTML",
			appErr:         errors.NewAuthenticationError("AUTH_REQUIRED", "Authentication required"),
			accept:         "text/html",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "401",
		},
		{
			name:           "403 error with HTML",
			appErr:         errors.NewAuthorizationError("FORBIDDEN", "Access forbidden"),
			accept:         "text/html",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "403",
		},
		{
			name:           "500 error with HTML",
			appErr:         errors.NewInternalError("INTERNAL", "Internal error"),
			accept:         "text/html",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "500",
		},
		{
			name:           "validation error with JSON",
			appErr:         errors.NewValidationError("VALIDATION", "Validation failed"),
			accept:         "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "VALIDATION",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Accept", tt.accept)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			router := NewErrorPageRouter()
			router.handleAppError(c, tt.appErr)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedBody)
		})
	}
}

// TestErrorPageRouter_handleGenericError tests generic error handling
func TestErrorPageRouter_handleGenericError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		accept         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "generic error with JSON",
			err:            assert.AnError,
			accept:         "application/json",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "INTERNAL_ERROR",
		},
		{
			name:           "generic error with HTML",
			err:            assert.AnError,
			accept:         "text/html",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Accept", tt.accept)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			router := NewErrorPageRouter()
			router.handleGenericError(c, tt.err)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedBody)
		})
	}
}

// TestGetErrorTitle tests the error title function
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
		t.Run(string(rune(tt.statusCode)), func(t *testing.T) {
			result := getErrorTitle(tt.statusCode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// BenchmarkErrorHandlers_Handle404 benchmarks the 404 handler
func BenchmarkErrorHandlers_Handle404(b *testing.B) {
	e := echo.New()
	handlers := NewErrorHandlers()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		handlers.Handle404(c)
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
