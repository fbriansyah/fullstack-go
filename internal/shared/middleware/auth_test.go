package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-templ-template/internal/modules/auth/application"
	"go-templ-template/internal/modules/auth/domain"
	userDomain "go-templ-template/internal/modules/user/domain"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock AuthService for testing
type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) Login(ctx context.Context, cmd *application.LoginCommand) (*application.AuthResult, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*application.AuthResult), args.Error(1)
}

func (m *mockAuthService) Register(ctx context.Context, cmd *application.RegisterCommand) (*application.AuthResult, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*application.AuthResult), args.Error(1)
}

func (m *mockAuthService) Logout(ctx context.Context, cmd *application.LogoutCommand) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *mockAuthService) ValidateSession(ctx context.Context, query *application.ValidateSessionQuery) (*application.SessionValidationResult, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*application.SessionValidationResult), args.Error(1)
}

func (m *mockAuthService) RefreshSession(ctx context.Context, cmd *application.RefreshSessionCommand) (*domain.Session, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Session), args.Error(1)
}

func (m *mockAuthService) ChangePassword(ctx context.Context, cmd *application.ChangePasswordCommand) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *mockAuthService) CleanupExpiredSessions(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockAuthService) GetUserSessions(ctx context.Context, userID string) ([]*domain.Session, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Session), args.Error(1)
}

// Test helper functions
func createTestUser() *userDomain.User {
	user, _ := userDomain.NewUser("user-123", "test@example.com", "Password123", "John", "Doe")
	return user
}

func createTestSession() *domain.Session {
	config := domain.SessionConfig{
		DefaultDuration: time.Hour * 24,
		MaxDuration:     time.Hour * 24 * 7,
		CleanupInterval: time.Hour,
	}
	session, _ := domain.NewSession("user-123", "192.168.1.1", "test-agent", config)
	return session
}

func setupEcho() *echo.Echo {
	e := echo.New()
	return e
}

func TestAuthMiddleware_RequireAuth_Success(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	middleware := NewAuthMiddleware(mockService)
	e := setupEcho()

	// Mock expectations
	validationResult := &application.SessionValidationResult{
		User:    createTestUser(),
		Session: createTestSession(),
		Valid:   true,
	}
	mockService.On("ValidateSession", mock.Anything, mock.MatchedBy(func(query *application.ValidateSessionQuery) bool {
		return query.SessionID == "valid-session-id"
	})).Return(validationResult, nil)

	// Create test handler
	testHandler := func(c echo.Context) error {
		user := GetUserFromContext(c)
		session := GetSessionFromContext(c)
		assert.NotNil(t, user)
		assert.NotNil(t, session)
		return c.JSON(http.StatusOK, map[string]string{"message": "success"})
	}

	// Create request with session cookie
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{
		Name:  SessionCookieName,
		Value: "valid-session-id",
	})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute middleware
	handler := middleware.RequireAuth(testHandler)
	err := handler(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockService.AssertExpectations(t)
}

func TestAuthMiddleware_RequireAuth_NoSession(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	middleware := NewAuthMiddleware(mockService)
	e := setupEcho()

	// Create test handler
	testHandler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "success"})
	}

	// Create request without session cookie
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute middleware
	handler := middleware.RequireAuth(testHandler)
	err := handler(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_RequireAuth_InvalidSession(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	middleware := NewAuthMiddleware(mockService)
	e := setupEcho()

	// Mock expectations
	validationResult := &application.SessionValidationResult{
		Valid: false,
	}
	mockService.On("ValidateSession", mock.Anything, mock.Anything).Return(validationResult, nil)

	// Create test handler
	testHandler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "success"})
	}

	// Create request with invalid session cookie
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{
		Name:  SessionCookieName,
		Value: "invalid-session-id",
	})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute middleware
	handler := middleware.RequireAuth(testHandler)
	err := handler(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	mockService.AssertExpectations(t)
}

func TestAuthMiddleware_RequireAuth_AuthorizationHeader(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	middleware := NewAuthMiddleware(mockService)
	e := setupEcho()

	// Mock expectations
	validationResult := &application.SessionValidationResult{
		User:    createTestUser(),
		Session: createTestSession(),
		Valid:   true,
	}
	mockService.On("ValidateSession", mock.Anything, mock.MatchedBy(func(query *application.ValidateSessionQuery) bool {
		return query.SessionID == "bearer-session-id"
	})).Return(validationResult, nil)

	// Create test handler
	testHandler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "success"})
	}

	// Create request with Authorization header
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer bearer-session-id")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute middleware
	handler := middleware.RequireAuth(testHandler)
	err := handler(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockService.AssertExpectations(t)
}

func TestAuthMiddleware_OptionalAuth_WithValidSession(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	middleware := NewAuthMiddleware(mockService)
	e := setupEcho()

	// Mock expectations
	validationResult := &application.SessionValidationResult{
		User:    createTestUser(),
		Session: createTestSession(),
		Valid:   true,
	}
	mockService.On("ValidateSession", mock.Anything, mock.Anything).Return(validationResult, nil)

	// Create test handler
	testHandler := func(c echo.Context) error {
		user := GetUserFromContext(c)
		session := GetSessionFromContext(c)
		assert.NotNil(t, user)
		assert.NotNil(t, session)
		return c.JSON(http.StatusOK, map[string]string{"message": "authenticated"})
	}

	// Create request with session cookie
	req := httptest.NewRequest(http.MethodGet, "/optional", nil)
	req.AddCookie(&http.Cookie{
		Name:  SessionCookieName,
		Value: "valid-session-id",
	})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute middleware
	handler := middleware.OptionalAuth(testHandler)
	err := handler(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockService.AssertExpectations(t)
}

func TestAuthMiddleware_OptionalAuth_WithoutSession(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	middleware := NewAuthMiddleware(mockService)
	e := setupEcho()

	// Create test handler
	testHandler := func(c echo.Context) error {
		user := GetUserFromContext(c)
		session := GetSessionFromContext(c)
		assert.Nil(t, user)
		assert.Nil(t, session)
		return c.JSON(http.StatusOK, map[string]string{"message": "anonymous"})
	}

	// Create request without session cookie
	req := httptest.NewRequest(http.MethodGet, "/optional", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute middleware
	handler := middleware.OptionalAuth(testHandler)
	err := handler(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthMiddleware_RequireRole_Success(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	middleware := NewAuthMiddleware(mockService)
	e := setupEcho()

	// Create test handler
	testHandler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "authorized"})
	}

	// Create request with user in context
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set user in context
	user := createTestUser()
	c.Set(UserContextKey, user)

	// Execute middleware
	handler := middleware.RequireRole("active")(testHandler)
	err := handler(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthMiddleware_RequireRole_InsufficientPermissions(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	middleware := NewAuthMiddleware(mockService)
	e := setupEcho()

	// Create test handler
	testHandler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "authorized"})
	}

	// Create request with user in context
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set user in context with different status
	user := createTestUser()
	user.Status = userDomain.UserStatusSuspended
	c.Set(UserContextKey, user)

	// Execute middleware
	handler := middleware.RequireRole("active")(testHandler)
	err := handler(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestAuthMiddleware_RequireRole_NoUser(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	middleware := NewAuthMiddleware(mockService)
	e := setupEcho()

	// Create test handler
	testHandler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "authorized"})
	}

	// Create request without user in context
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute middleware
	handler := middleware.RequireRole("active")(testHandler)
	err := handler(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_CSRF_GetRequest(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	middleware := NewAuthMiddleware(mockService)
	e := setupEcho()

	// Create test handler
	testHandler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "success"})
	}

	// Create GET request (should skip CSRF check)
	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute middleware
	handler := middleware.CSRF(testHandler)
	err := handler(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthMiddleware_CSRF_PostWithToken(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	middleware := NewAuthMiddleware(mockService)
	e := setupEcho()

	// Generate a valid CSRF token
	config := DefaultCSRFConfig()
	csrfMiddleware := NewCSRFMiddleware(config)
	token, err := csrfMiddleware.generateToken()
	require.NoError(t, err)

	// Create test handler
	testHandler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "success"})
	}

	// Create POST request with CSRF token in header and cookie
	req := httptest.NewRequest(http.MethodPost, "/api/data", nil)
	req.Header.Set("X-CSRF-Token", token)
	req.AddCookie(&http.Cookie{
		Name:  CSRFCookieName,
		Value: token,
	})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute middleware
	handler := middleware.CSRF(testHandler)
	err = handler(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthMiddleware_CSRF_PostWithoutToken(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	middleware := NewAuthMiddleware(mockService)
	e := setupEcho()

	// Create test handler
	testHandler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "success"})
	}

	// Create POST request without CSRF token
	req := httptest.NewRequest(http.MethodPost, "/api/data", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute middleware
	handler := middleware.CSRF(testHandler)
	err := handler(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestSetSessionCookie(t *testing.T) {
	// Setup
	e := setupEcho()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	SetSessionCookie(c, "test-session-id", 3600)

	// Assert
	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)

	cookie := cookies[0]
	assert.Equal(t, SessionCookieName, cookie.Name)
	assert.Equal(t, "test-session-id", cookie.Value)
	assert.Equal(t, 3600, cookie.MaxAge)
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
}

func TestGetUserFromContext(t *testing.T) {
	// Setup
	e := setupEcho()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Test with no user in context
	user := GetUserFromContext(c)
	assert.Nil(t, user)

	// Test with user in context
	testUser := createTestUser()
	c.Set(UserContextKey, testUser)
	user = GetUserFromContext(c)
	assert.Equal(t, testUser, user)
}

func TestGetSessionFromContext(t *testing.T) {
	// Setup
	e := setupEcho()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Test with no session in context
	session := GetSessionFromContext(c)
	assert.Nil(t, session)

	// Test with session in context
	testSession := createTestSession()
	c.Set(SessionContextKey, testSession)
	session = GetSessionFromContext(c)
	assert.Equal(t, testSession, session)
}
