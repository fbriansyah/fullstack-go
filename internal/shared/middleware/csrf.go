package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

const (
	// CSRFTokenLength is the length of CSRF tokens in bytes
	CSRFTokenLength = 32

	// CSRFCookieName is the name of the CSRF cookie
	CSRFCookieName = "csrf_token"

	// CSRFHeaderName is the name of the CSRF header
	CSRFHeaderName = "X-CSRF-Token"

	// CSRFFormFieldName is the name of the CSRF form field
	CSRFFormFieldName = "_csrf_token"
)

// CSRFConfig holds configuration for CSRF protection
type CSRFConfig struct {
	// TokenLength is the length of the CSRF token in bytes
	TokenLength int

	// CookieName is the name of the CSRF cookie
	CookieName string

	// HeaderName is the name of the CSRF header
	HeaderName string

	// FormFieldName is the name of the CSRF form field
	FormFieldName string

	// CookieMaxAge is the max age of the CSRF cookie in seconds
	CookieMaxAge int

	// CookiePath is the path of the CSRF cookie
	CookiePath string

	// CookieDomain is the domain of the CSRF cookie
	CookieDomain string

	// CookieSecure indicates if the CSRF cookie should be secure
	CookieSecure bool

	// CookieHTTPOnly indicates if the CSRF cookie should be HTTP only
	CookieHTTPOnly bool

	// CookieSameSite is the SameSite attribute of the CSRF cookie
	CookieSameSite http.SameSite
}

// DefaultCSRFConfig returns a default CSRF configuration
func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{
		TokenLength:    CSRFTokenLength,
		CookieName:     CSRFCookieName,
		HeaderName:     CSRFHeaderName,
		FormFieldName:  CSRFFormFieldName,
		CookieMaxAge:   3600, // 1 hour
		CookiePath:     "/",
		CookieSecure:   false, // Set to true in production with HTTPS
		CookieHTTPOnly: false, // Must be false so JavaScript can read it
		CookieSameSite: http.SameSiteStrictMode,
	}
}

// CSRFMiddleware provides CSRF protection middleware
type CSRFMiddleware struct {
	config CSRFConfig
}

// NewCSRFMiddleware creates a new CSRF middleware instance
func NewCSRFMiddleware(config CSRFConfig) *CSRFMiddleware {
	return &CSRFMiddleware{
		config: config,
	}
}

// Protect returns a middleware function that provides CSRF protection
func (m *CSRFMiddleware) Protect(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Skip CSRF check for safe methods
		method := c.Request().Method
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			// Generate and set CSRF token for safe methods
			token, err := m.generateToken()
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]interface{}{
					"error":   "INTERNAL_ERROR",
					"message": "Failed to generate CSRF token",
				})
			}

			m.setCSRFCookie(c, token)
			c.Set("csrf_token", token)
			return next(c)
		}

		// For unsafe methods, validate CSRF token
		token := m.getTokenFromRequest(c)
		if token == "" {
			return c.JSON(http.StatusForbidden, map[string]interface{}{
				"error":   "CSRF_TOKEN_MISSING",
				"message": "CSRF token is required",
			})
		}

		cookieToken := m.getTokenFromCookie(c)
		if cookieToken == "" {
			return c.JSON(http.StatusForbidden, map[string]interface{}{
				"error":   "CSRF_TOKEN_INVALID",
				"message": "CSRF token is invalid",
			})
		}

		// Validate token using constant-time comparison
		if !m.validateToken(token, cookieToken) {
			return c.JSON(http.StatusForbidden, map[string]interface{}{
				"error":   "CSRF_TOKEN_INVALID",
				"message": "CSRF token is invalid",
			})
		}

		return next(c)
	}
}

// GenerateToken generates a new CSRF token and sets it in the response
func (m *CSRFMiddleware) GenerateToken(c echo.Context) (string, error) {
	token, err := m.generateToken()
	if err != nil {
		return "", err
	}

	m.setCSRFCookie(c, token)
	return token, nil
}

// generateToken generates a new random CSRF token
func (m *CSRFMiddleware) generateToken() (string, error) {
	bytes := make([]byte, m.config.TokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// getTokenFromRequest extracts CSRF token from request header or form
func (m *CSRFMiddleware) getTokenFromRequest(c echo.Context) string {
	// Try header first
	token := c.Request().Header.Get(m.config.HeaderName)
	if token != "" {
		return token
	}

	// Try form field
	token = c.FormValue(m.config.FormFieldName)
	if token != "" {
		return token
	}

	return ""
}

// getTokenFromCookie extracts CSRF token from cookie
func (m *CSRFMiddleware) getTokenFromCookie(c echo.Context) string {
	cookie, err := c.Cookie(m.config.CookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// validateToken validates the CSRF token using constant-time comparison
func (m *CSRFMiddleware) validateToken(requestToken, cookieToken string) bool {
	// Decode tokens
	requestBytes, err := base64.URLEncoding.DecodeString(requestToken)
	if err != nil {
		return false
	}

	cookieBytes, err := base64.URLEncoding.DecodeString(cookieToken)
	if err != nil {
		return false
	}

	// Use constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare(requestBytes, cookieBytes) == 1
}

// setCSRFCookie sets the CSRF token cookie
func (m *CSRFMiddleware) setCSRFCookie(c echo.Context, token string) {
	cookie := &http.Cookie{
		Name:     m.config.CookieName,
		Value:    token,
		Path:     m.config.CookiePath,
		Domain:   m.config.CookieDomain,
		MaxAge:   m.config.CookieMaxAge,
		Secure:   m.config.CookieSecure,
		HttpOnly: m.config.CookieHTTPOnly,
		SameSite: m.config.CookieSameSite,
	}

	// Set secure flag based on request scheme if not explicitly configured
	if c.Request().TLS != nil {
		cookie.Secure = true
	}

	c.SetCookie(cookie)
}

// GetCSRFToken returns the CSRF token from context or generates a new one
func GetCSRFToken(c echo.Context) string {
	if token, ok := c.Get("csrf_token").(string); ok {
		return token
	}
	return ""
}

// CSRFTokenHandler returns a handler that provides CSRF token to clients
func (m *CSRFMiddleware) CSRFTokenHandler(c echo.Context) error {
	token, err := m.GenerateToken(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "INTERNAL_ERROR",
			"message": "Failed to generate CSRF token",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"csrf_token": token,
	})
}

// Enhanced CSRF middleware for the auth middleware
func (m *AuthMiddleware) CSRFEnhanced(config CSRFConfig) echo.MiddlewareFunc {
	csrfMiddleware := NewCSRFMiddleware(config)
	return csrfMiddleware.Protect
}
