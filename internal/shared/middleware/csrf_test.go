package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCSRFMiddleware_GetRequest_GeneratesToken(t *testing.T) {
	// Setup
	config := DefaultCSRFConfig()
	middleware := NewCSRFMiddleware(config)
	e := setupEcho()

	// Create test handler
	testHandler := func(c echo.Context) error {
		token := GetCSRFToken(c)
		assert.NotEmpty(t, token)
		return c.JSON(http.StatusOK, map[string]string{"message": "success"})
	}

	// Create GET request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute middleware
	handler := middleware.Protect(testHandler)
	err := handler(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Check that CSRF cookie was set
	cookies := rec.Result().Cookies()
	var csrfCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == CSRFCookieName {
			csrfCookie = cookie
			break
		}
	}
	require.NotNil(t, csrfCookie)
	assert.NotEmpty(t, csrfCookie.Value)
}

func TestCSRFMiddleware_PostRequest_ValidToken(t *testing.T) {
	// Setup
	config := DefaultCSRFConfig()
	middleware := NewCSRFMiddleware(config)
	e := setupEcho()

	// Generate a token first
	token, err := middleware.generateToken()
	require.NoError(t, err)

	// Create test handler
	testHandler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "success"})
	}

	// Create POST request with CSRF token in header
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(CSRFHeaderName, token)
	req.AddCookie(&http.Cookie{
		Name:  CSRFCookieName,
		Value: token,
	})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute middleware
	handler := middleware.Protect(testHandler)
	err = handler(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCSRFMiddleware_PostRequest_ValidTokenInForm(t *testing.T) {
	// Setup
	config := DefaultCSRFConfig()
	middleware := NewCSRFMiddleware(config)
	e := setupEcho()

	// Generate a token first
	token, err := middleware.generateToken()
	require.NoError(t, err)

	// Create test handler
	testHandler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "success"})
	}

	// Create POST request with CSRF token in form
	form := url.Values{}
	form.Add(CSRFFormFieldName, token)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	req.AddCookie(&http.Cookie{
		Name:  CSRFCookieName,
		Value: token,
	})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute middleware
	handler := middleware.Protect(testHandler)
	err = handler(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCSRFMiddleware_PostRequest_MissingToken(t *testing.T) {
	// Setup
	config := DefaultCSRFConfig()
	middleware := NewCSRFMiddleware(config)
	e := setupEcho()

	// Create test handler
	testHandler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "success"})
	}

	// Create POST request without CSRF token
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute middleware
	handler := middleware.Protect(testHandler)
	err := handler(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestCSRFMiddleware_PostRequest_InvalidToken(t *testing.T) {
	// Setup
	config := DefaultCSRFConfig()
	middleware := NewCSRFMiddleware(config)
	e := setupEcho()

	// Generate tokens
	validToken, err := middleware.generateToken()
	require.NoError(t, err)
	invalidToken, err := middleware.generateToken()
	require.NoError(t, err)

	// Create test handler
	testHandler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "success"})
	}

	// Create POST request with mismatched tokens
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(CSRFHeaderName, invalidToken)
	req.AddCookie(&http.Cookie{
		Name:  CSRFCookieName,
		Value: validToken,
	})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute middleware
	handler := middleware.Protect(testHandler)
	err = handler(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestCSRFMiddleware_PostRequest_MissingCookie(t *testing.T) {
	// Setup
	config := DefaultCSRFConfig()
	middleware := NewCSRFMiddleware(config)
	e := setupEcho()

	// Generate a token
	token, err := middleware.generateToken()
	require.NoError(t, err)

	// Create test handler
	testHandler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "success"})
	}

	// Create POST request with token in header but no cookie
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(CSRFHeaderName, token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute middleware
	handler := middleware.Protect(testHandler)
	err = handler(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestCSRFMiddleware_TokenHandler(t *testing.T) {
	// Setup
	config := DefaultCSRFConfig()
	middleware := NewCSRFMiddleware(config)
	e := setupEcho()

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/csrf-token", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute handler
	err := middleware.CSRFTokenHandler(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Check response contains token
	assert.Contains(t, rec.Body.String(), "csrf_token")

	// Check that CSRF cookie was set
	cookies := rec.Result().Cookies()
	var csrfCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == CSRFCookieName {
			csrfCookie = cookie
			break
		}
	}
	require.NotNil(t, csrfCookie)
	assert.NotEmpty(t, csrfCookie.Value)
}

func TestCSRFMiddleware_GenerateToken(t *testing.T) {
	// Setup
	config := DefaultCSRFConfig()
	middleware := NewCSRFMiddleware(config)
	e := setupEcho()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	token, err := middleware.GenerateToken(c)

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Check that cookie was set
	cookies := rec.Result().Cookies()
	var csrfCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == CSRFCookieName {
			csrfCookie = cookie
			break
		}
	}
	require.NotNil(t, csrfCookie)
	assert.Equal(t, token, csrfCookie.Value)
}

func TestCSRFMiddleware_ValidateToken(t *testing.T) {
	// Setup
	config := DefaultCSRFConfig()
	middleware := NewCSRFMiddleware(config)

	// Generate a token
	token, err := middleware.generateToken()
	require.NoError(t, err)

	// Test valid token
	assert.True(t, middleware.validateToken(token, token))

	// Test invalid token
	invalidToken, err := middleware.generateToken()
	require.NoError(t, err)
	assert.False(t, middleware.validateToken(token, invalidToken))

	// Test malformed token
	assert.False(t, middleware.validateToken("invalid", token))
	assert.False(t, middleware.validateToken(token, "invalid"))
}

func TestDefaultCSRFConfig(t *testing.T) {
	config := DefaultCSRFConfig()

	assert.Equal(t, CSRFTokenLength, config.TokenLength)
	assert.Equal(t, CSRFCookieName, config.CookieName)
	assert.Equal(t, CSRFHeaderName, config.HeaderName)
	assert.Equal(t, CSRFFormFieldName, config.FormFieldName)
	assert.Equal(t, 3600, config.CookieMaxAge)
	assert.Equal(t, "/", config.CookiePath)
	assert.False(t, config.CookieSecure)
	assert.False(t, config.CookieHTTPOnly)
	assert.Equal(t, http.SameSiteStrictMode, config.CookieSameSite)
}
