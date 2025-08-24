package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go-templ-template/internal/modules/auth/application"
	"go-templ-template/internal/modules/auth/domain"
	userDomain "go-templ-template/internal/modules/user/domain"
	"go-templ-template/internal/shared/middleware"

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

func TestAuthHandler_Login_Success(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	handler := NewAuthHandler(mockService)
	e := setupEcho()

	// Mock expectations
	authResult := &application.AuthResult{
		User:    createTestUser(),
		Session: createTestSession(),
	}
	mockService.On("Login", mock.Anything, mock.MatchedBy(func(cmd *application.LoginCommand) bool {
		return cmd.Email == "test@example.com" && cmd.Password == "Password123"
	})).Return(authResult, nil)

	// Create request
	loginReq := LoginRequest{
		Email:    "test@example.com",
		Password: "Password123",
	}
	reqBody, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := handler.Login(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response AuthResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Login successful", response.Message)
	assert.Equal(t, "test@example.com", response.User.Email)
	assert.NotEmpty(t, response.Session.ID)

	// Check session cookie was set
	cookies := rec.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == middleware.SessionCookieName {
			sessionCookie = cookie
			break
		}
	}
	require.NotNil(t, sessionCookie)
	assert.Equal(t, authResult.Session.ID, sessionCookie.Value)

	mockService.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	handler := NewAuthHandler(mockService)
	e := setupEcho()

	// Mock expectations
	authError := &application.AuthError{
		Code:    "INVALID_CREDENTIALS",
		Message: "Invalid email or password",
	}
	mockService.On("Login", mock.Anything, mock.Anything).Return(nil, authError)

	// Create request
	loginReq := LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}
	reqBody, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := handler.Login(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "INVALID_CREDENTIALS", response.Error)

	mockService.AssertExpectations(t)
}

func TestAuthHandler_Login_ValidationError(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	handler := NewAuthHandler(mockService)
	e := setupEcho()

	// Create request with invalid email
	loginReq := LoginRequest{
		Email:    "invalid-email",
		Password: "Password123",
	}
	reqBody, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := handler.Login(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "VALIDATION_ERROR", response["error"])
}

func TestAuthHandler_Register_Success(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	handler := NewAuthHandler(mockService)
	e := setupEcho()

	// Mock expectations
	authResult := &application.AuthResult{
		User:    createTestUser(),
		Session: createTestSession(),
	}
	mockService.On("Register", mock.Anything, mock.MatchedBy(func(cmd *application.RegisterCommand) bool {
		return cmd.Email == "newuser@example.com" && cmd.FirstName == "Jane"
	})).Return(authResult, nil)

	// Create request
	registerReq := RegisterRequest{
		Email:     "newuser@example.com",
		Password:  "Password123",
		FirstName: "Jane",
		LastName:  "Smith",
	}
	reqBody, _ := json.Marshal(registerReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := handler.Register(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response AuthResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Registration successful", response.Message)
	assert.Equal(t, "test@example.com", response.User.Email)

	mockService.AssertExpectations(t)
}

func TestAuthHandler_Register_UserAlreadyExists(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	handler := NewAuthHandler(mockService)
	e := setupEcho()

	// Mock expectations
	authError := &application.AuthError{
		Code:    "USER_ALREADY_EXISTS",
		Message: "User with this email already exists",
	}
	mockService.On("Register", mock.Anything, mock.Anything).Return(nil, authError)

	// Create request
	registerReq := RegisterRequest{
		Email:     "existing@example.com",
		Password:  "Password123",
		FirstName: "Jane",
		LastName:  "Smith",
	}
	reqBody, _ := json.Marshal(registerReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := handler.Register(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)

	var response ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "USER_ALREADY_EXISTS", response.Error)

	mockService.AssertExpectations(t)
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	handler := NewAuthHandler(mockService)
	e := setupEcho()

	// Mock expectations
	mockService.On("Logout", mock.Anything, mock.MatchedBy(func(cmd *application.LogoutCommand) bool {
		return cmd.SessionID == "session-123"
	})).Return(nil)

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set session and user in context (simulating auth middleware)
	user := createTestUser()
	session := createTestSession()
	session.ID = "session-123"
	c.Set(middleware.UserContextKey, user)
	c.Set(middleware.SessionContextKey, session)

	// Execute
	err := handler.Logout(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response SuccessResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Logout successful", response.Message)

	// Check session cookie was cleared
	cookies := rec.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == middleware.SessionCookieName {
			sessionCookie = cookie
			break
		}
	}
	require.NotNil(t, sessionCookie)
	assert.Equal(t, "", sessionCookie.Value)
	assert.Equal(t, -1, sessionCookie.MaxAge)

	mockService.AssertExpectations(t)
}

func TestAuthHandler_Logout_NoSession(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	handler := NewAuthHandler(mockService)
	e := setupEcho()

	// Create request without session in context
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := handler.Logout(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "UNAUTHORIZED", response.Error)
}

func TestAuthHandler_Me_Success(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	handler := NewAuthHandler(mockService)
	e := setupEcho()

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set user in context (simulating auth middleware)
	user := createTestUser()
	c.Set(middleware.UserContextKey, user)

	// Execute
	err := handler.Me(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response UserResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, user.Email, response.Email)
	assert.Equal(t, user.FirstName, response.FirstName)
	assert.Equal(t, user.LastName, response.LastName)
}

func TestAuthHandler_RefreshSession_Success(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	handler := NewAuthHandler(mockService)
	e := setupEcho()

	// Mock expectations
	refreshedSession := createTestSession()
	refreshedSession.ID = "refreshed-session-123"
	mockService.On("RefreshSession", mock.Anything, mock.MatchedBy(func(cmd *application.RefreshSessionCommand) bool {
		return cmd.SessionID == "session-123"
	})).Return(refreshedSession, nil)

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set session and user in context (simulating auth middleware)
	user := createTestUser()
	session := createTestSession()
	session.ID = "session-123"
	c.Set(middleware.UserContextKey, user)
	c.Set(middleware.SessionContextKey, session)

	// Execute
	err := handler.RefreshSession(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response SuccessResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Session refreshed successfully", response.Message)

	// Check session cookie was updated
	cookies := rec.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == middleware.SessionCookieName {
			sessionCookie = cookie
			break
		}
	}
	require.NotNil(t, sessionCookie)
	assert.Equal(t, refreshedSession.ID, sessionCookie.Value)

	mockService.AssertExpectations(t)
}

func TestAuthHandler_ChangePassword_Success(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	handler := NewAuthHandler(mockService)
	e := setupEcho()

	// Mock expectations
	mockService.On("ChangePassword", mock.Anything, mock.MatchedBy(func(cmd *application.ChangePasswordCommand) bool {
		return cmd.UserID == "user-123" && cmd.OldPassword == "OldPassword123" && cmd.NewPassword == "NewPassword123"
	})).Return(nil)

	// Create request
	changePasswordReq := ChangePasswordRequest{
		OldPassword: "OldPassword123",
		NewPassword: "NewPassword123",
	}
	reqBody, _ := json.Marshal(changePasswordReq)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/auth/password", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set user in context (simulating auth middleware)
	user := createTestUser()
	c.Set(middleware.UserContextKey, user)

	// Execute
	err := handler.ChangePassword(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response SuccessResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Message, "Password changed successfully")

	// Check session cookie was cleared (forcing re-login)
	cookies := rec.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == middleware.SessionCookieName {
			sessionCookie = cookie
			break
		}
	}
	require.NotNil(t, sessionCookie)
	assert.Equal(t, "", sessionCookie.Value)
	assert.Equal(t, -1, sessionCookie.MaxAge)

	mockService.AssertExpectations(t)
}

func TestAuthHandler_ValidateSession_Success(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	handler := NewAuthHandler(mockService)
	e := setupEcho()

	// Mock expectations
	validationResult := &application.SessionValidationResult{
		User:    createTestUser(),
		Session: createTestSession(),
		Valid:   true,
	}
	mockService.On("ValidateSession", mock.Anything, mock.MatchedBy(func(query *application.ValidateSessionQuery) bool {
		return query.SessionID == "session-123"
	})).Return(validationResult, nil)

	// Create request with session ID in query param
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/validate?session_id=session-123", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := handler.ValidateSession(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["valid"].(bool))
	assert.NotNil(t, response["user"])
	assert.NotNil(t, response["session"])

	mockService.AssertExpectations(t)
}

func TestAuthHandler_ValidateSession_InvalidSession(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	handler := NewAuthHandler(mockService)
	e := setupEcho()

	// Mock expectations
	validationResult := &application.SessionValidationResult{
		Valid: false,
	}
	mockService.On("ValidateSession", mock.Anything, mock.Anything).Return(validationResult, nil)

	// Create request with session ID in cookie
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/validate", nil)
	req.AddCookie(&http.Cookie{
		Name:  middleware.SessionCookieName,
		Value: "invalid-session",
	})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := handler.ValidateSession(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "INVALID_SESSION", response.Error)

	mockService.AssertExpectations(t)
}

// Integration test with middleware - simplified without CSRF for now
func TestAuthHandler_Integration_LoginAndProtectedRoute(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	e := setupEcho()

	// Register routes without CSRF for this test
	authHandler := NewAuthHandler(mockService)
	authMiddleware := middleware.NewAuthMiddleware(mockService)

	// API v1 routes without CSRF
	v1 := e.Group("/api/v1")
	auth := v1.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.GET("/validate", authHandler.ValidateSession)
	}

	authProtected := v1.Group("/auth")
	authProtected.Use(authMiddleware.RequireAuth)
	{
		authProtected.GET("/me", authHandler.Me)
	}

	// Mock login
	authResult := &application.AuthResult{
		User:    createTestUser(),
		Session: createTestSession(),
	}
	mockService.On("Login", mock.Anything, mock.Anything).Return(authResult, nil)

	// Mock session validation for protected route
	validationResult := &application.SessionValidationResult{
		User:    createTestUser(),
		Session: createTestSession(),
		Valid:   true,
	}
	mockService.On("ValidateSession", mock.Anything, mock.Anything).Return(validationResult, nil)

	// Test login
	loginReq := LoginRequest{
		Email:    "test@example.com",
		Password: "Password123",
	}
	reqBody, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Assert login success
	assert.Equal(t, http.StatusOK, rec.Code)

	// Extract session cookie
	cookies := rec.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == middleware.SessionCookieName {
			sessionCookie = cookie
			break
		}
	}
	require.NotNil(t, sessionCookie)

	// Test protected route with session cookie
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req2.AddCookie(sessionCookie)
	rec2 := httptest.NewRecorder()

	e.ServeHTTP(rec2, req2)

	// Assert protected route access
	assert.Equal(t, http.StatusOK, rec2.Code)

	mockService.AssertExpectations(t)
}

// Test malformed JSON requests
func TestAuthHandler_MalformedJSON(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	handler := NewAuthHandler(mockService)
	e := setupEcho()

	// Create request with malformed JSON
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader("{invalid json"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := handler.Login(c)

	// Assert - Echo's bind will fail and return an HTTP error
	if err != nil {
		// This is expected - Echo will return a binding error
		assert.Contains(t, err.Error(), "code=400")
	} else {
		// If no error, check that we got a 400 response
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

// Test empty request body
func TestAuthHandler_EmptyRequestBody(t *testing.T) {
	// Setup
	mockService := new(mockAuthService)
	handler := NewAuthHandler(mockService)
	e := setupEcho()

	// Create request with empty body
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(""))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := handler.Login(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
