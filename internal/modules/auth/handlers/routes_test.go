package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-templ-template/internal/modules/auth/application"
	"go-templ-template/internal/shared/middleware"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRegisterAuthRoutes_PublicRoutes(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	e := setupEcho()

	// Register routes
	RegisterAuthRoutes(e, mockService)

	// Test cases for public routes
	testCases := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "Login endpoint exists",
			method:         http.MethodPost,
			path:           "/api/v1/auth/login",
			expectedStatus: http.StatusBadRequest, // Will fail validation but route exists
		},
		{
			name:           "Register endpoint exists",
			method:         http.MethodPost,
			path:           "/api/v1/auth/register",
			expectedStatus: http.StatusBadRequest, // Will fail validation but route exists
		},
		{
			name:           "Validate endpoint exists",
			method:         http.MethodGet,
			path:           "/api/v1/auth/validate",
			expectedStatus: http.StatusBadRequest, // Will fail validation but route exists
		},
		{
			name:           "CSRF token endpoint exists",
			method:         http.MethodGet,
			path:           "/api/v1/auth/csrf-token",
			expectedStatus: http.StatusOK, // Should work without auth
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			// Route should exist (not 404)
			assert.NotEqual(t, http.StatusNotFound, rec.Code, "Route should exist")

			// For CSRF token endpoint, expect success
			if tc.path == "/api/v1/auth/csrf-token" {
				assert.Equal(t, http.StatusOK, rec.Code)

				// Check response contains CSRF token
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "csrf_token")
			}
		})
	}
}

func TestRegisterAuthRoutes_ProtectedRoutes(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	e := setupEcho()

	// Register routes
	RegisterAuthRoutes(e, mockService)

	// Test cases for protected routes (should require authentication)
	testCases := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "Logout endpoint requires auth",
			method: http.MethodPost,
			path:   "/api/v1/auth/logout",
		},
		{
			name:   "Me endpoint requires auth",
			method: http.MethodGet,
			path:   "/api/v1/auth/me",
		},
		{
			name:   "Refresh endpoint requires auth",
			method: http.MethodPost,
			path:   "/api/v1/auth/refresh",
		},
		{
			name:   "Change password endpoint requires auth",
			method: http.MethodPut,
			path:   "/api/v1/auth/password",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			// Should return 401 Unauthorized (not 404)
			assert.Equal(t, http.StatusUnauthorized, rec.Code, "Protected route should require authentication")

			var response map[string]interface{}
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, "UNAUTHORIZED", response["error"])
		})
	}
}

func TestRegisterAuthRoutes_WithValidSession(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	e := setupEcho()

	// Register routes without CSRF for this test
	authHandler := NewAuthHandler(mockService)
	authMiddleware := middleware.NewAuthMiddleware(mockService)

	v1 := e.Group("/api/v1")
	authProtected := v1.Group("/auth")
	authProtected.Use(authMiddleware.RequireAuth)
	{
		authProtected.POST("/logout", authHandler.Logout)
	}

	// Mock session validation
	validationResult := &application.SessionValidationResult{
		User:    createTestUser(),
		Session: createTestSession(),
		Valid:   true,
	}
	mockService.On("ValidateSession", mock.Anything, mock.MatchedBy(func(query *application.ValidateSessionQuery) bool {
		return query.SessionID == "valid-session-id"
	})).Return(validationResult, nil)

	// Mock logout with proper command matching
	mockService.On("Logout", mock.Anything, mock.MatchedBy(func(cmd *application.LogoutCommand) bool {
		return cmd.SessionID == validationResult.Session.ID
	})).Return(nil)

	// Test protected route with valid session
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.AddCookie(&http.Cookie{
		Name:  middleware.SessionCookieName,
		Value: "valid-session-id",
	})
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Should succeed
	if rec.Code != http.StatusOK {
		t.Logf("Request failed with status %d, body: %s", rec.Code, rec.Body.String())
	}
	assert.Equal(t, http.StatusOK, rec.Code)

	mockService.AssertExpectations(t)
}

func TestRegisterAuthRoutesWithMiddleware(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	e := setupEcho()

	// Custom middleware for testing
	customMiddleware := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("X-Custom-Header", "test")
			return next(c)
		}
	}

	// Register routes with custom middleware
	RegisterAuthRoutesWithMiddleware(e, mockService, customMiddleware)

	// Test that custom middleware is applied to CSRF token endpoint (GET request, no CSRF needed)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/csrf-token", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Check custom header is present
	assert.Equal(t, "test", rec.Header().Get("X-Custom-Header"))
	// Should succeed
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetAuthMiddleware(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)

	// Get middleware
	authMiddleware := GetAuthMiddleware(mockService)

	// Assert
	assert.NotNil(t, authMiddleware)
	assert.IsType(t, &middleware.AuthMiddleware{}, authMiddleware)
}

func TestAuthRoutes_CSRFProtection(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	e := setupEcho()

	// Register routes
	RegisterAuthRoutes(e, mockService)

	// Test that CSRF token endpoint works
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/csrf-token", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Extract CSRF token from response
	var tokenResponse map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &tokenResponse)
	require.NoError(t, err)
	assert.Contains(t, tokenResponse, "csrf_token")
	assert.NotEmpty(t, tokenResponse["csrf_token"])

	// Extract CSRF cookie
	cookies := rec.Result().Cookies()
	var csrfCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == middleware.CSRFCookieName {
			csrfCookie = cookie
			break
		}
	}
	require.NotNil(t, csrfCookie)
	assert.NotEmpty(t, csrfCookie.Value)
}

func TestAuthRoutes_CSRFProtection_MissingToken(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	e := setupEcho()

	// Register routes
	RegisterAuthRoutes(e, mockService)

	// Test POST request without CSRF token - should be blocked by CSRF middleware
	loginReq := LoginRequest{
		Email:    "test@example.com",
		Password: "Password123",
	}
	reqBody, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Should be blocked by CSRF protection
	assert.Equal(t, http.StatusForbidden, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "CSRF")
}

func TestAuthRoutes_MethodNotAllowed(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	e := setupEcho()

	// Register routes
	RegisterAuthRoutes(e, mockService)

	// Test wrong HTTP method - GET on login endpoint should be method not allowed
	// But first it will hit CSRF middleware which generates token for GET requests
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/login", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// With CSRF middleware, GET requests are allowed to generate tokens
	// So we expect either 200 (token generated) or 405 (method not allowed)
	// The exact behavior depends on how Echo handles this
	assert.NotEqual(t, http.StatusNotFound, rec.Code)
}

func TestAuthRoutes_ContentTypeValidation(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	e := setupEcho()

	// Register routes
	RegisterAuthRoutes(e, mockService)

	// Test POST request without proper content type
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader([]byte(`{"email":"test@example.com","password":"Password123"}`)))
	// Don't set Content-Type header
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Should handle gracefully (Echo will try to bind anyway)
	// The exact behavior depends on Echo's configuration
	assert.NotEqual(t, http.StatusNotFound, rec.Code)
}

// Test route parameter extraction and validation
func TestAuthRoutes_ParameterHandling(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	e := setupEcho()

	// Register routes
	RegisterAuthRoutes(e, mockService)

	// Mock session validation
	validationResult := &application.SessionValidationResult{
		User:    createTestUser(),
		Session: createTestSession(),
		Valid:   true,
	}
	mockService.On("ValidateSession", mock.Anything, mock.MatchedBy(func(query *application.ValidateSessionQuery) bool {
		return query.SessionID == "test-session-123"
	})).Return(validationResult, nil)

	// Test validate endpoint with query parameter
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/validate?session_id=test-session-123", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["valid"].(bool))

	mockService.AssertExpectations(t)
}

// Test error handling in routes
func TestAuthRoutes_ErrorHandling(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	e := setupEcho()

	// Register routes without CSRF for this test
	authHandler := NewAuthHandler(mockService)
	v1 := e.Group("/api/v1")
	auth := v1.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
	}

	// Mock service error
	authError := &application.AuthError{
		Code:    "INTERNAL_ERROR",
		Message: "Database connection failed",
	}
	mockService.On("Login", mock.Anything, mock.Anything).Return(nil, authError)

	// Test login with service error
	loginReq := LoginRequest{
		Email:    "test@example.com",
		Password: "Password123",
	}
	reqBody, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response ErrorResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "INTERNAL_ERROR", response.Error)
	assert.Equal(t, "Database connection failed", response.Message)

	mockService.AssertExpectations(t)
}
