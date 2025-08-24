package middleware

import (
	"net/http"
	"strings"

	"go-templ-template/internal/modules/auth/application"
	userDomain "go-templ-template/internal/modules/user/domain"

	"github.com/labstack/echo/v4"
)

const (
	// SessionCookieName is the name of the session cookie
	SessionCookieName = "session_id"

	// UserContextKey is the key used to store user in context
	UserContextKey = "user"

	// SessionContextKey is the key used to store session in context
	SessionContextKey = "session"
)

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	authService application.AuthService
}

// NewAuthMiddleware creates a new auth middleware instance
func NewAuthMiddleware(authService application.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// RequireAuth middleware that requires authentication
func (m *AuthMiddleware) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sessionID := m.getSessionID(c)
		if sessionID == "" {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"error":   "UNAUTHORIZED",
				"message": "Authentication required",
			})
		}

		// Validate session
		query := &application.ValidateSessionQuery{
			SessionID: sessionID,
			IPAddress: c.RealIP(),
			UserAgent: c.Request().UserAgent(),
		}

		result, err := m.authService.ValidateSession(c.Request().Context(), query)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error":   "INTERNAL_ERROR",
				"message": "Failed to validate session",
			})
		}

		if !result.Valid {
			// Clear invalid session cookie
			m.clearSessionCookie(c)
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"error":   "UNAUTHORIZED",
				"message": "Invalid or expired session",
			})
		}

		// Store user and session in context
		c.Set(UserContextKey, result.User)
		c.Set(SessionContextKey, result.Session)

		return next(c)
	}
}

// OptionalAuth middleware that optionally authenticates users
func (m *AuthMiddleware) OptionalAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sessionID := m.getSessionID(c)
		if sessionID != "" {
			// Validate session
			query := &application.ValidateSessionQuery{
				SessionID: sessionID,
				IPAddress: c.RealIP(),
				UserAgent: c.Request().UserAgent(),
			}

			result, err := m.authService.ValidateSession(c.Request().Context(), query)
			if err == nil && result.Valid {
				// Store user and session in context
				c.Set(UserContextKey, result.User)
				c.Set(SessionContextKey, result.Session)
			} else {
				// Clear invalid session cookie
				m.clearSessionCookie(c)
			}
		}

		return next(c)
	}
}

// RequireRole middleware that requires specific user role/status
func (m *AuthMiddleware) RequireRole(allowedStatuses ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userData := GetUserFromContext(c)
			if userData == nil {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error":   "UNAUTHORIZED",
					"message": "Authentication required",
				})
			}

			// Type assert to get user
			user, ok := userData.(*userDomain.User)
			if !ok {
				return c.JSON(http.StatusInternalServerError, map[string]interface{}{
					"error":   "INTERNAL_ERROR",
					"message": "Invalid user data",
				})
			}

			// Check if user status is allowed
			userStatus := string(user.Status)
			for _, status := range allowedStatuses {
				if userStatus == status {
					return next(c)
				}
			}

			return c.JSON(http.StatusForbidden, map[string]interface{}{
				"error":   "FORBIDDEN",
				"message": "Insufficient permissions",
			})
		}
	}
}

// CSRF middleware for CSRF protection (deprecated - use CSRFEnhanced instead)
func (m *AuthMiddleware) CSRF(next echo.HandlerFunc) echo.HandlerFunc {
	config := DefaultCSRFConfig()
	csrfMiddleware := NewCSRFMiddleware(config)
	return csrfMiddleware.Protect(next)
}

// getSessionID extracts session ID from cookie or Authorization header
func (m *AuthMiddleware) getSessionID(c echo.Context) string {
	// Try to get session ID from cookie first
	cookie, err := c.Cookie(SessionCookieName)
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// Try to get session ID from Authorization header
	auth := c.Request().Header.Get("Authorization")
	if auth != "" {
		// Support "Bearer <session_id>" format
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer ")
		}
		// Support direct session ID
		return auth
	}

	return ""
}

// clearSessionCookie clears the session cookie
func (m *AuthMiddleware) clearSessionCookie(c echo.Context) {
	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   c.Request().TLS != nil, // Only secure in HTTPS
		SameSite: http.SameSiteStrictMode,
	}
	c.SetCookie(cookie)
}

// SetSessionCookie sets the session cookie
func SetSessionCookie(c echo.Context, sessionID string, maxAge int) {
	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    sessionID,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   c.Request().TLS != nil, // Only secure in HTTPS
		SameSite: http.SameSiteStrictMode,
	}
	c.SetCookie(cookie)
}

// GetUserFromContext retrieves the authenticated user from context
func GetUserFromContext(c echo.Context) interface{} {
	return c.Get(UserContextKey)
}

// GetSessionFromContext retrieves the session from context
func GetSessionFromContext(c echo.Context) interface{} {
	return c.Get(SessionContextKey)
}
