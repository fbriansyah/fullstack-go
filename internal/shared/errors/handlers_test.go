package errors

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorHandlers_Handle404(t *testing.T) {
	tests := []struct {
		name           string
		acceptHeader   string
		userAgent      string
		expectedStatus int
		expectedType   string // "html" or "json"
		expectedBody   []string
	}{
		{
			name:           "HTML request with Accept header",
			acceptHeader:   "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			userAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			expectedStatus: http.StatusNotFound,
			expectedType:   "html",
			expectedBody:   []string{"404", "Page Not Found", "Go Home"},
		},
		{
			name:           "JSON API request",
			acceptHeader:   "application/json",
			userAgent:      "curl/7.68.0",
			expectedStatus: http.StatusNotFound,
			expectedType:   "json",
			expectedBody:   []string{"NOT_FOUND", "The requested resource was not found"},
		},
		{
			name:           "Browser request without Accept header",
			acceptHeader:   "",
			userAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			expectedStatus: http.StatusNotFound,
			expectedType:   "html",
			expectedBody:   []string{"404", "Page Not Found"},
		},
		{
			name:           "API path request",
			acceptHeader:   "",
			userAgent:      "curl/7.68.0",
			expectedStatus: http.StatusNotFound,
			expectedType:   "json",
			expectedBody:   []string{"NOT_FOUND"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Echo instance and error handlers
			e := echo.New()
			handlers := NewErrorHandlers(ErrorHandlersConfig{
				LogErrors: false, // Disable logging for tests
			})

			// Create request
			path := "/nonexistent"
			if tt.expectedType == "json" && tt.acceptHeader == "" {
				path = "/api/nonexistent"
			}

			req := httptest.NewRequest(http.MethodGet, path, nil)
			if tt.acceptHeader != "" {
				req.Header.Set("Accept", tt.acceptHeader)
			}
			if tt.userAgent != "" {
				req.Header.Set("User-Agent", tt.userAgent)
			}

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Execute handler
			err := handlers.Handle404(c)
			require.NoError(t, err)

			// Check status code
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Check content type and body
			body := rec.Body.String()
			if tt.expectedType == "html" {
				assert.Contains(t, rec.Header().Get("Content-Type"), "text/html")
				for _, expected := range tt.expectedBody {
					assert.Contains(t, body, expected)
				}
			} else {
				assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")
				for _, expected := range tt.expectedBody {
					assert.Contains(t, body, expected)
				}
			}

			// Check cache headers
			assert.Equal(t, "no-cache, no-store, must-revalidate", rec.Header().Get("Cache-Control"))
			assert.Equal(t, "no-cache", rec.Header().Get("Pragma"))
			assert.Equal(t, "0", rec.Header().Get("Expires"))
		})
	}
}

func TestErrorHandlers_Handle500(t *testing.T) {
	tests := []struct {
		name           string
		acceptHeader   string
		expectedStatus int
		expectedType   string
		expectedBody   []string
	}{
		{
			name:           "HTML request",
			acceptHeader:   "text/html",
			expectedStatus: http.StatusInternalServerError,
			expectedType:   "html",
			expectedBody:   []string{"500", "Server Error", "Try Again"},
		},
		{
			name:           "JSON request",
			acceptHeader:   "application/json",
			expectedStatus: http.StatusInternalServerError,
			expectedType:   "json",
			expectedBody:   []string{"INTERNAL_ERROR", "An internal server error occurred"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			handlers := NewErrorHandlers(ErrorHandlersConfig{
				LogErrors: false,
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Accept", tt.acceptHeader)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handlers.Handle500(c)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			body := rec.Body.String()
			for _, expected := range tt.expectedBody {
				assert.Contains(t, body, expected)
			}
		})
	}
}

func TestErrorHandlers_Handle401(t *testing.T) {
	tests := []struct {
		name           string
		acceptHeader   string
		expectedStatus int
		expectedType   string
		expectedBody   []string
	}{
		{
			name:           "HTML request",
			acceptHeader:   "text/html",
			expectedStatus: http.StatusUnauthorized,
			expectedType:   "html",
			expectedBody:   []string{"401", "Authentication Required", "Sign In"},
		},
		{
			name:           "JSON request",
			acceptHeader:   "application/json",
			expectedStatus: http.StatusUnauthorized,
			expectedType:   "json",
			expectedBody:   []string{"UNAUTHORIZED", "Authentication is required"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			handlers := NewErrorHandlers(ErrorHandlersConfig{
				LogErrors: false,
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Accept", tt.acceptHeader)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handlers.Handle401(c)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			body := rec.Body.String()
			for _, expected := range tt.expectedBody {
				assert.Contains(t, body, expected)
			}
		})
	}
}

func TestErrorHandlers_Handle403(t *testing.T) {
	tests := []struct {
		name           string
		acceptHeader   string
		expectedStatus int
		expectedBody   []string
	}{
		{
			name:           "HTML request",
			acceptHeader:   "text/html",
			expectedStatus: http.StatusForbidden,
			expectedBody:   []string{"403", "Access Forbidden", "Go Home"},
		},
		{
			name:           "JSON request",
			acceptHeader:   "application/json",
			expectedStatus: http.StatusForbidden,
			expectedBody:   []string{"FORBIDDEN", "You don't have permission"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			handlers := NewErrorHandlers(ErrorHandlersConfig{
				LogErrors: false,
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Accept", tt.acceptHeader)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handlers.Handle403(c)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			body := rec.Body.String()
			for _, expected := range tt.expectedBody {
				assert.Contains(t, body, expected)
			}
		})
	}
}

func TestErrorHandlers_HandleGeneric(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		title          string
		message        string
		code           string
		acceptHeader   string
		expectedStatus int
		expectedBody   []string
	}{
		{
			name:           "HTML generic error",
			statusCode:     http.StatusBadRequest,
			title:          "Bad Request",
			message:        "The request was invalid",
			code:           "BAD_REQUEST",
			acceptHeader:   "text/html",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   []string{"Bad Request", "The request was invalid", "BAD_REQUEST"},
		},
		{
			name:           "JSON generic error",
			statusCode:     http.StatusBadRequest,
			title:          "Bad Request",
			message:        "The request was invalid",
			code:           "BAD_REQUEST",
			acceptHeader:   "application/json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   []string{"BAD_REQUEST", "The request was invalid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			handlers := NewErrorHandlers(ErrorHandlersConfig{
				LogErrors: false,
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Accept", tt.acceptHeader)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handlers.HandleGeneric(c, tt.statusCode, tt.title, tt.message, tt.code)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			body := rec.Body.String()
			for _, expected := range tt.expectedBody {
				assert.Contains(t, body, expected)
			}
		})
	}
}

func TestErrorHandlers_CustomHTTPErrorHandler(t *testing.T) {
	tests := []struct {
		name           string
		error          error
		acceptHeader   string
		expectedStatus int
		expectedBody   []string
	}{
		{
			name:           "Echo HTTP 404 error",
			error:          echo.NewHTTPError(http.StatusNotFound, "Not found"),
			acceptHeader:   "text/html",
			expectedStatus: http.StatusNotFound,
			expectedBody:   []string{"404", "Page Not Found"},
		},
		{
			name:           "Echo HTTP 500 error",
			error:          echo.NewHTTPError(http.StatusInternalServerError, "Internal error"),
			acceptHeader:   "text/html",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   []string{"500", "Server Error"},
		},
		{
			name:           "AppError 401",
			error:          NewAuthenticationError("AUTH_REQUIRED", "Authentication required"),
			acceptHeader:   "text/html",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   []string{"401", "Authentication Required"},
		},
		{
			name:           "AppError 403",
			error:          NewAuthorizationError("FORBIDDEN", "Access denied"),
			acceptHeader:   "text/html",
			expectedStatus: http.StatusForbidden,
			expectedBody:   []string{"403", "Access Forbidden"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			handlers := NewErrorHandlers(ErrorHandlersConfig{
				LogErrors: false,
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Accept", tt.acceptHeader)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Call the custom error handler
			handlers.CustomHTTPErrorHandler(tt.error, c)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			body := rec.Body.String()
			for _, expected := range tt.expectedBody {
				assert.Contains(t, body, expected)
			}
		})
	}
}

func TestErrorHandlers_RegisterRoutes(t *testing.T) {
	e := echo.New()
	handlers := NewErrorHandlers(ErrorHandlersConfig{
		LogErrors: false,
	})

	// Register routes
	handlers.RegisterRoutes(e)

	// Test that routes are registered
	routes := e.Routes()
	expectedRoutes := []string{
		"/error/404",
		"/error/500",
		"/error/401",
		"/error/403",
	}

	registeredPaths := make(map[string]bool)
	for _, route := range routes {
		registeredPaths[route.Path] = true
	}

	for _, expectedPath := range expectedRoutes {
		assert.True(t, registeredPaths[expectedPath], "Route %s should be registered", expectedPath)
	}
}

func TestErrorHandlers_NotFoundHandler(t *testing.T) {
	e := echo.New()
	handlers := NewErrorHandlers(ErrorHandlersConfig{
		LogErrors: false,
	})

	// Set custom 404 handler
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if he, ok := err.(*echo.HTTPError); ok && he.Code == http.StatusNotFound {
			handlers.Handle404(c)
		}
	}

	// Test 404 handler
	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	req.Header.Set("Accept", "text/html")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "404")
	assert.Contains(t, rec.Body.String(), "Page Not Found")
}

func TestErrorHandlers_MethodNotAllowedHandler(t *testing.T) {
	tests := []struct {
		name           string
		acceptHeader   string
		expectedStatus int
		expectedBody   []string
	}{
		{
			name:           "HTML request",
			acceptHeader:   "text/html",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   []string{"405", "Method Not Allowed"},
		},
		{
			name:           "JSON request",
			acceptHeader:   "application/json",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   []string{"METHOD_NOT_ALLOWED"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			handlers := NewErrorHandlers(ErrorHandlersConfig{
				LogErrors: false,
			})

			req := httptest.NewRequest(http.MethodPost, "/", nil)
			req.Header.Set("Accept", tt.acceptHeader)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handlers.MethodNotAllowedHandler()(c)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			body := rec.Body.String()
			for _, expected := range tt.expectedBody {
				assert.Contains(t, body, expected)
			}
		})
	}
}

func TestIsHTMLRequest(t *testing.T) {
	tests := []struct {
		name      string
		accept    string
		userAgent string
		path      string
		expected  bool
	}{
		{
			name:     "HTML Accept header",
			accept:   "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			expected: true,
		},
		{
			name:     "JSON Accept header",
			accept:   "application/json",
			expected: false,
		},
		{
			name:      "Browser User-Agent",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			expected:  true,
		},
		{
			name:      "Chrome User-Agent",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			expected:  true,
		},
		{
			name:      "API tool User-Agent",
			userAgent: "curl/7.68.0",
			expected:  false,
		},
		{
			name:     "API path",
			path:     "/api/users",
			expected: false,
		},
		{
			name:     "Regular path",
			path:     "/users",
			expected: true,
		},
		{
			name:     "No headers, default to HTML",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			path := tt.path
			if path == "" {
				path = "/"
			}

			req := httptest.NewRequest(http.MethodGet, path, nil)
			if tt.accept != "" {
				req.Header.Set("Accept", tt.accept)
			}
			if tt.userAgent != "" {
				req.Header.Set("User-Agent", tt.userAgent)
			}

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			result := isHTMLRequest(c)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrorHandlers_ErrorPageMiddleware(t *testing.T) {
	e := echo.New()
	handlers := NewErrorHandlers(ErrorHandlersConfig{
		LogErrors: false,
	})

	// Add error page middleware
	e.Use(handlers.ErrorPageMiddleware())

	// Add a route that returns an error
	e.GET("/test-error", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusNotFound, "Test error")
	})

	req := httptest.NewRequest(http.MethodGet, "/test-error", nil)
	req.Header.Set("Accept", "text/html")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// The middleware should handle the error and return a 404 page
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "404")
}

// Benchmark tests
func BenchmarkErrorHandlers_Handle404_HTML(b *testing.B) {
	e := echo.New()
	handlers := NewErrorHandlers(ErrorHandlersConfig{
		LogErrors: false,
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "text/html")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		handlers.Handle404(c)
	}
}

func BenchmarkErrorHandlers_Handle404_JSON(b *testing.B) {
	e := echo.New()
	handlers := NewErrorHandlers(ErrorHandlersConfig{
		LogErrors: false,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Accept", "application/json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		handlers.Handle404(c)
	}
}
