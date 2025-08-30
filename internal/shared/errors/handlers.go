package errors

import (
	"net/http"

	"go-templ-template/web/templates/pages"

	"github.com/labstack/echo/v4"
)

// ErrorHandlers provides HTTP handlers for error pages
type ErrorHandlers struct {
	config ErrorHandlersConfig
}

// ErrorHandlersConfig holds configuration for error handlers
type ErrorHandlersConfig struct {
	// ShowDetailedErrors shows detailed error information (dev only)
	ShowDetailedErrors bool

	// LogErrors logs error page requests
	LogErrors bool
}

// NewErrorHandlers creates a new error handlers instance
func NewErrorHandlers(config ErrorHandlersConfig) *ErrorHandlers {
	return &ErrorHandlers{
		config: config,
	}
}

// Handle404 handles 404 Not Found errors
func (h *ErrorHandlers) Handle404(c echo.Context) error {
	// Log the 404 if configured
	if h.config.LogErrors {
		c.Logger().Warnf("404 Not Found: %s %s", c.Request().Method, c.Request().URL.Path)
	}

	// Set appropriate headers
	c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Response().Header().Set("Pragma", "no-cache")
	c.Response().Header().Set("Expires", "0")

	// Check if request accepts HTML
	if isHTMLRequest(c) {
		// Set content type and render HTML error page
		c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
		c.Response().WriteHeader(http.StatusNotFound)
		return pages.Error404Page().Render(c.Request().Context(), c.Response().Writer)
	}

	// Return JSON error for API requests
	return c.JSON(http.StatusNotFound, map[string]interface{}{
		"error": map[string]interface{}{
			"code":    "NOT_FOUND",
			"message": "The requested resource was not found",
			"status":  http.StatusNotFound,
		},
	})
}

// Handle500 handles 500 Internal Server Error
func (h *ErrorHandlers) Handle500(c echo.Context) error {
	// Log the 500 error
	if h.config.LogErrors {
		c.Logger().Errorf("500 Internal Server Error: %s %s", c.Request().Method, c.Request().URL.Path)
	}

	// Set appropriate headers
	c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Response().Header().Set("Pragma", "no-cache")
	c.Response().Header().Set("Expires", "0")

	// Check if request accepts HTML
	if isHTMLRequest(c) {
		// Set content type and render HTML error page
		c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
		c.Response().WriteHeader(http.StatusInternalServerError)
		return pages.Error500Page().Render(c.Request().Context(), c.Response().Writer)
	}

	// Return JSON error for API requests
	return c.JSON(http.StatusInternalServerError, map[string]interface{}{
		"error": map[string]interface{}{
			"code":    "INTERNAL_ERROR",
			"message": "An internal server error occurred",
			"status":  http.StatusInternalServerError,
		},
	})
}

// Handle401 handles 401 Unauthorized errors
func (h *ErrorHandlers) Handle401(c echo.Context) error {
	// Log the 401 if configured
	if h.config.LogErrors {
		c.Logger().Warnf("401 Unauthorized: %s %s", c.Request().Method, c.Request().URL.Path)
	}

	// Set appropriate headers
	c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Response().Header().Set("Pragma", "no-cache")
	c.Response().Header().Set("Expires", "0")

	// Check if request accepts HTML
	if isHTMLRequest(c) {
		// Set content type and render HTML error page
		c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
		c.Response().WriteHeader(http.StatusUnauthorized)
		return pages.ErrorAuthPage().Render(c.Request().Context(), c.Response().Writer)
	}

	// Return JSON error for API requests
	return c.JSON(http.StatusUnauthorized, map[string]interface{}{
		"error": map[string]interface{}{
			"code":    "UNAUTHORIZED",
			"message": "Authentication is required to access this resource",
			"status":  http.StatusUnauthorized,
		},
	})
}

// Handle403 handles 403 Forbidden errors
func (h *ErrorHandlers) Handle403(c echo.Context) error {
	// Log the 403 if configured
	if h.config.LogErrors {
		c.Logger().Warnf("403 Forbidden: %s %s", c.Request().Method, c.Request().URL.Path)
	}

	// Set appropriate headers
	c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Response().Header().Set("Pragma", "no-cache")
	c.Response().Header().Set("Expires", "0")

	// Check if request accepts HTML
	if isHTMLRequest(c) {
		// Set content type and render HTML error page
		c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
		c.Response().WriteHeader(http.StatusForbidden)
		return pages.ErrorForbiddenPage().Render(c.Request().Context(), c.Response().Writer)
	}

	// Return JSON error for API requests
	return c.JSON(http.StatusForbidden, map[string]interface{}{
		"error": map[string]interface{}{
			"code":    "FORBIDDEN",
			"message": "You don't have permission to access this resource",
			"status":  http.StatusForbidden,
		},
	})
}

// HandleGeneric handles generic errors with custom message
func (h *ErrorHandlers) HandleGeneric(c echo.Context, statusCode int, title, message, code string) error {
	// Log the error if configured
	if h.config.LogErrors {
		c.Logger().Warnf("%d %s: %s %s", statusCode, title, c.Request().Method, c.Request().URL.Path)
	}

	// Set appropriate headers
	c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Response().Header().Set("Pragma", "no-cache")
	c.Response().Header().Set("Expires", "0")

	// Check if request accepts HTML
	if isHTMLRequest(c) {
		// Set content type and render HTML error page
		c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
		c.Response().WriteHeader(statusCode)
		return pages.GenericErrorPage(title, message, code).Render(c.Request().Context(), c.Response().Writer)
	}

	// Return JSON error for API requests
	return c.JSON(statusCode, map[string]interface{}{
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
			"status":  statusCode,
		},
	})
}

// RegisterRoutes registers error page routes
func (h *ErrorHandlers) RegisterRoutes(e *echo.Echo) {
	// Register explicit error page routes
	e.GET("/error/404", h.Handle404)
	e.GET("/error/500", h.Handle500)
	e.GET("/error/401", h.Handle401)
	e.GET("/error/403", h.Handle403)
}

// CustomHTTPErrorHandler creates a custom HTTP error handler for Echo
func (h *ErrorHandlers) CustomHTTPErrorHandler(err error, c echo.Context) {
	// Don't handle if response already started
	if c.Response().Committed {
		return
	}

	var (
		code = http.StatusInternalServerError
		msg  interface{}
	)

	// Handle different error types
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		msg = he.Message
		if he.Internal != nil {
			err = he.Internal
		}
	} else if appErr, ok := err.(*AppError); ok {
		code = appErr.HTTPStatus
		msg = appErr.Message
	}

	// Handle specific status codes
	switch code {
	case http.StatusNotFound:
		h.Handle404(c)
		return
	case http.StatusUnauthorized:
		h.Handle401(c)
		return
	case http.StatusForbidden:
		h.Handle403(c)
		return
	case http.StatusInternalServerError:
		h.Handle500(c)
		return
	default:
		// Handle other status codes with generic error page
		title := http.StatusText(code)
		message := "An error occurred while processing your request"
		if msg != nil {
			if msgStr, ok := msg.(string); ok && msgStr != "" {
				message = msgStr
			}
		}

		h.HandleGeneric(c, code, title, message, http.StatusText(code))
	}
}

// isHTMLRequest checks if the request accepts HTML content
func isHTMLRequest(c echo.Context) bool {
	accept := c.Request().Header.Get("Accept")

	// Check for explicit HTML accept header
	if accept != "" {
		// Simple check for HTML content type
		return contains(accept, "text/html") || contains(accept, "application/xhtml")
	}

	// Default to HTML for browser requests (no explicit Accept header)
	userAgent := c.Request().Header.Get("User-Agent")
	if userAgent != "" {
		// Check for API tools/clients that should get JSON
		apiTools := []string{"curl", "wget", "HTTPie", "Postman", "Insomnia"}
		for _, tool := range apiTools {
			if contains(userAgent, tool) {
				return false
			}
		}

		// Check for common browser user agents
		browsers := []string{"Mozilla", "Chrome", "Safari", "Edge", "Opera"}
		for _, browser := range browsers {
			if contains(userAgent, browser) {
				return true
			}
		}
	}

	// Check if it's an API request path
	path := c.Request().URL.Path
	if contains(path, "/api/") {
		return false
	}

	// Default to HTML for non-API requests
	return true
}

// ErrorPageMiddleware creates middleware that handles errors with custom error pages
func (h *ErrorHandlers) ErrorPageMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err != nil {
				// Let the custom error handler deal with it
				h.CustomHTTPErrorHandler(err, c)
				return nil
			}
			return nil
		}
	}
}

// NotFoundHandler creates a handler for 404 errors that can be used as Echo's NotFoundHandler
func (h *ErrorHandlers) NotFoundHandler() echo.HandlerFunc {
	return h.Handle404
}

// MethodNotAllowedHandler creates a handler for 405 Method Not Allowed errors
func (h *ErrorHandlers) MethodNotAllowedHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Log the 405 if configured
		if h.config.LogErrors {
			c.Logger().Warnf("405 Method Not Allowed: %s %s", c.Request().Method, c.Request().URL.Path)
		}

		// Set appropriate headers
		c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Response().Header().Set("Pragma", "no-cache")
		c.Response().Header().Set("Expires", "0")

		// Check if request accepts HTML
		if isHTMLRequest(c) {
			// Set content type and render HTML error page
			c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
			c.Response().WriteHeader(http.StatusMethodNotAllowed)
			return pages.GenericErrorPage(
				"Method Not Allowed",
				"The requested method is not allowed for this resource",
				"405",
			).Render(c.Request().Context(), c.Response().Writer)
		}

		// Return JSON error for API requests
		return c.JSON(http.StatusMethodNotAllowed, map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "METHOD_NOT_ALLOWED",
				"message": "The requested method is not allowed for this resource",
				"status":  http.StatusMethodNotAllowed,
			},
		})
	}
}
