package handlers

import (
	"go-templ-template/internal/modules/auth/application"
	"go-templ-template/internal/shared/middleware"

	"github.com/labstack/echo/v4"
)

// RegisterAuthRoutes registers all authentication-related routes
func RegisterAuthRoutes(e *echo.Echo, authService application.AuthService) {
	authHandler := NewAuthHandler(authService)
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Create CSRF middleware
	csrfConfig := middleware.DefaultCSRFConfig()
	csrfMiddleware := middleware.NewCSRFMiddleware(csrfConfig)

	// API v1 routes
	v1 := e.Group("/api/v1")

	// Public auth routes (no authentication required)
	auth := v1.Group("/auth")
	// Apply CSRF protection to all auth routes
	auth.Use(csrfMiddleware.Protect)
	{
		auth.POST("/login", authHandler.Login)                   // POST /api/v1/auth/login
		auth.POST("/register", authHandler.Register)             // POST /api/v1/auth/register
		auth.GET("/validate", authHandler.ValidateSession)       // GET /api/v1/auth/validate
		auth.GET("/csrf-token", csrfMiddleware.CSRFTokenHandler) // GET /api/v1/auth/csrf-token
	}

	// Protected auth routes (authentication required)
	authProtected := v1.Group("/auth")
	authProtected.Use(authMiddleware.RequireAuth)
	authProtected.Use(csrfMiddleware.Protect) // Apply CSRF protection to protected routes too
	{
		authProtected.POST("/logout", authHandler.Logout)          // POST /api/v1/auth/logout
		authProtected.GET("/me", authHandler.Me)                   // GET /api/v1/auth/me
		authProtected.POST("/refresh", authHandler.RefreshSession) // POST /api/v1/auth/refresh
		authProtected.PUT("/password", authHandler.ChangePassword) // PUT /api/v1/auth/password
	}
}

// RegisterAuthRoutesWithMiddleware registers auth routes with custom middleware
func RegisterAuthRoutesWithMiddleware(e *echo.Echo, authService application.AuthService, middlewares ...echo.MiddlewareFunc) {
	authHandler := NewAuthHandler(authService)
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Create CSRF middleware
	csrfConfig := middleware.DefaultCSRFConfig()
	csrfMiddleware := middleware.NewCSRFMiddleware(csrfConfig)

	// API v1 routes
	v1 := e.Group("/api/v1")

	// Apply custom middleware to the group
	for _, mw := range middlewares {
		v1.Use(mw)
	}

	// Public auth routes (no authentication required)
	auth := v1.Group("/auth")
	auth.Use(csrfMiddleware.Protect) // Apply CSRF protection
	{
		auth.POST("/login", authHandler.Login)                   // POST /api/v1/auth/login
		auth.POST("/register", authHandler.Register)             // POST /api/v1/auth/register
		auth.GET("/validate", authHandler.ValidateSession)       // GET /api/v1/auth/validate
		auth.GET("/csrf-token", csrfMiddleware.CSRFTokenHandler) // GET /api/v1/auth/csrf-token
	}

	// Protected auth routes (authentication required)
	authProtected := v1.Group("/auth")
	authProtected.Use(authMiddleware.RequireAuth)
	authProtected.Use(csrfMiddleware.Protect) // Apply CSRF protection
	{
		authProtected.POST("/logout", authHandler.Logout)          // POST /api/v1/auth/logout
		authProtected.GET("/me", authHandler.Me)                   // GET /api/v1/auth/me
		authProtected.POST("/refresh", authHandler.RefreshSession) // POST /api/v1/auth/refresh
		authProtected.PUT("/password", authHandler.ChangePassword) // PUT /api/v1/auth/password
	}
}

// GetAuthMiddleware returns the auth middleware for use in other modules
func GetAuthMiddleware(authService application.AuthService) *middleware.AuthMiddleware {
	return middleware.NewAuthMiddleware(authService)
}
