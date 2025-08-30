// Package handlers provides HTTP handlers for error pages and fallback handling.
// This package integrates with the error handling middleware to provide consistent
// error page routing and user-friendly error responses.
package handlers

import (
	"net/http"

	"go-templ-template/internal/shared/errors"
	"go-templ-template/web/templates/pages"

	"github.com/labstack/echo/v4"
)

// ErrorHandlers provides HTTP handlers for various error scenarios
type ErrorHandlers struct{}

// NewErrorHandlers creates a new ErrorHandlers instance
func NewErrorHandlers() *ErrorHandlers {
	return &ErrorHandlers{}
}

// RegisterRoutes registers error page routes with the Echo router
func (h *ErrorHandlers) RegisterRoutes(e *echo.Echo) {
	// Error page routes for direct access (useful for testing)
	errorGroup := e.Group("/error")
	errorGroup.GET("/404", h.Handle404)
	errorGroup.GET("/500", h.Handle500)
	errorGroup.GET("/401", h.HandleAuth)
	errorGroup.GET("/403", h.HandleForbidden)

	// Generic error handler with parameters
	errorGroup.GET("/generic/:code", h.HandleGeneric)
}

// Handle404 renders the 404 Not Found error page
func (h *ErrorHandlers) Handle404(c echo.Context) error {
	c.Response().WriteHeader(http.StatusNotFound)
	return pages.Error404Page().Render(c.Request().Context(), c.Response().Writer)
}

// Handle500 renders the 500 Internal Server Error page
func (h *ErrorHandlers) Handle500(c echo.Context) error {
	c.Response().WriteHeader(http.StatusInternalServerError)
	return pages.Error500Page().Render(c.Request().Context(), c.Response().Writer)
}

// HandleAuth renders the authentication required error page
func (h *ErrorHandlers) HandleAuth(c echo.Context) error {
	c.Response().WriteHeader(http.StatusUnauthorized)
	return pages.ErrorAuthPage().Render(c.Request().Context(), c.Response().Writer)
}

// HandleForbidden renders the access forbidden error page
func (h *ErrorHandlers) HandleForbidden(c echo.Context) error {
	c.Response().WriteHeader(http.StatusForbidden)
	return pages.ErrorForbiddenPage().Render(c.Request().Context(), c.Response().Writer)
}

// HandleGeneric renders a generic error page with custom status code
func (h *ErrorHandlers) HandleGeneric(c echo.Context) error {
	code := c.Param("code")

	// Parse status code
	var statusCode int
	var title, message string

	switch code {
	case "400":
		statusCode = http.StatusBadRequest
		title = "Bad Request"
		message = "The request could not be understood by the server due to malformed syntax."
	case "401":
		return h.HandleAuth(c)
	case "403":
		return h.HandleForbidden(c)
	case "404":
		return h.Handle404(c)
	case "405":
		statusCode = http.StatusMethodNotAllowed
		title = "Method Not Allowed"
		message = "The HTTP method used is not allowed for this resource."
	case "408":
		statusCode = http.StatusRequestTimeout
		title = "Request Timeout"
		message = "The server timed out waiting for the request."
	case "409":
		statusCode = http.StatusConflict
		title = "Conflict"
		message = "The request could not be completed due to a conflict with the current state of the resource."
	case "410":
		statusCode = http.StatusGone
		title = "Gone"
		message = "The requested resource is no longer available and will not be available again."
	case "422":
		statusCode = http.StatusUnprocessableEntity
		title = "Unprocessable Entity"
		message = "The request was well-formed but was unable to be followed due to semantic errors."
	case "429":
		statusCode = http.StatusTooManyRequests
		title = "Too Many Requests"
		message = "You have sent too many requests in a given amount of time. Please try again later."
	case "500":
		return h.Handle500(c)
	case "502":
		statusCode = http.StatusBadGateway
		title = "Bad Gateway"
		message = "The server received an invalid response from an upstream server."
	case "503":
		statusCode = http.StatusServiceUnavailable
		title = "Service Unavailable"
		message = "The server is currently unable to handle the request due to temporary overloading or maintenance."
	case "504":
		statusCode = http.StatusGatewayTimeout
		title = "Gateway Timeout"
		message = "The server did not receive a timely response from an upstream server."
	default:
		statusCode = http.StatusInternalServerError
		title = "Error"
		message = "An error occurred while processing your request."
	}

	c.Response().WriteHeader(statusCode)
	return pages.GenericErrorPage(title, message, code).Render(c.Request().Context(), c.Response().Writer)
}

// FallbackHandler provides a fallback error handler for unhandled routes
type FallbackHandler struct{}

// NewFallbackHandler creates a new FallbackHandler instance
func NewFallbackHandler() *FallbackHandler {
	return &FallbackHandler{}
}

// Handle404Fallback handles 404 errors as a fallback
func (h *FallbackHandler) Handle404Fallback(c echo.Context) error {
	// Check if this is an API request
	if isAPIRequest(c) {
		appErr := errors.NewNotFoundError("ROUTE_NOT_FOUND", "The requested API endpoint was not found")
		return c.JSON(http.StatusNotFound, appErr.ToHTTPResponse())
	}

	// Render HTML error page
	c.Response().WriteHeader(http.StatusNotFound)
	return pages.Error404Page().Render(c.Request().Context(), c.Response().Writer)
}

// HandleMethodNotAllowed handles method not allowed errors
func (h *FallbackHandler) HandleMethodNotAllowed(c echo.Context) error {
	// Check if this is an API request
	if isAPIRequest(c) {
		appErr := errors.NewValidationError("METHOD_NOT_ALLOWED", "The HTTP method is not allowed for this endpoint")
		appErr.HTTPStatus = http.StatusMethodNotAllowed
		return c.JSON(http.StatusMethodNotAllowed, appErr.ToHTTPResponse())
	}

	// Render HTML error page
	c.Response().WriteHeader(http.StatusMethodNotAllowed)
	return pages.GenericErrorPage("Method Not Allowed",
		"The HTTP method you used is not allowed for this resource.", "405").
		Render(c.Request().Context(), c.Response().Writer)
}

// isAPIRequest determines if the request is for an API endpoint
func isAPIRequest(c echo.Context) bool {
	path := c.Request().URL.Path
	accept := c.Request().Header.Get("Accept")
	contentType := c.Request().Header.Get("Content-Type")

	// Check if path starts with /api
	if len(path) >= 5 && path[:5] == "/api/" {
		return true
	}

	// Check Accept header for JSON
	if len(accept) > 0 && contains(accept, "application/json") {
		return true
	}

	// Check Content-Type for JSON
	if len(contentType) > 0 && contains(contentType, "application/json") {
		return true
	}

	// Check for AJAX requests
	if c.Request().Header.Get("X-Requested-With") == "XMLHttpRequest" {
		return true
	}

	return false
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

// indexOf returns the index of the first occurrence of substr in s, or -1 if not found
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// ErrorPageRouter provides routing configuration for error pages
type ErrorPageRouter struct {
	errorHandlers   *ErrorHandlers
	fallbackHandler *FallbackHandler
}

// NewErrorPageRouter creates a new ErrorPageRouter instance
func NewErrorPageRouter() *ErrorPageRouter {
	return &ErrorPageRouter{
		errorHandlers:   NewErrorHandlers(),
		fallbackHandler: NewFallbackHandler(),
	}
}

// RegisterRoutes registers all error-related routes
func (r *ErrorPageRouter) RegisterRoutes(e *echo.Echo) {
	// Register error page routes
	r.errorHandlers.RegisterRoutes(e)

	// Set custom error handlers
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		// Handle specific Echo HTTP errors
		if he, ok := err.(*echo.HTTPError); ok {
			switch he.Code {
			case http.StatusNotFound:
				r.fallbackHandler.Handle404Fallback(c)
				return
			case http.StatusMethodNotAllowed:
				r.fallbackHandler.HandleMethodNotAllowed(c)
				return
			}
		}

		// Handle AppErrors
		if appErr, ok := errors.AsAppError(err); ok {
			r.handleAppError(c, appErr)
			return
		}

		// Default error handling
		r.handleGenericError(c, err)
	}
}

// handleAppError handles AppError instances
func (r *ErrorPageRouter) handleAppError(c echo.Context, appErr *errors.AppError) {
	if isAPIRequest(c) {
		c.JSON(appErr.HTTPStatus, appErr.ToHTTPResponse())
		return
	}

	c.Response().WriteHeader(appErr.HTTPStatus)

	switch appErr.HTTPStatus {
	case http.StatusNotFound:
		pages.Error404Page().Render(c.Request().Context(), c.Response().Writer)
	case http.StatusUnauthorized:
		pages.ErrorAuthPage().Render(c.Request().Context(), c.Response().Writer)
	case http.StatusForbidden:
		pages.ErrorForbiddenPage().Render(c.Request().Context(), c.Response().Writer)
	case http.StatusInternalServerError:
		pages.Error500Page().Render(c.Request().Context(), c.Response().Writer)
	default:
		title := getErrorTitle(appErr.HTTPStatus)
		message := appErr.Message
		if appErr.UserMessage != "" {
			message = appErr.UserMessage
		}
		code := ""
		if appErr.HTTPStatus > 0 {
			code = http.StatusText(appErr.HTTPStatus)
		}
		pages.GenericErrorPage(title, message, code).Render(c.Request().Context(), c.Response().Writer)
	}
}

// handleGenericError handles generic errors
func (r *ErrorPageRouter) handleGenericError(c echo.Context, err error) {
	if isAPIRequest(c) {
		appErr := errors.NewInternalError("INTERNAL_ERROR", "An internal error occurred")
		appErr.Cause = err
		c.JSON(http.StatusInternalServerError, appErr.ToHTTPResponse())
		return
	}

	c.Response().WriteHeader(http.StatusInternalServerError)
	pages.Error500Page().Render(c.Request().Context(), c.Response().Writer)
}

// getErrorTitle returns a user-friendly title for HTTP status codes
func getErrorTitle(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return "Bad Request"
	case http.StatusUnauthorized:
		return "Authentication Required"
	case http.StatusForbidden:
		return "Access Forbidden"
	case http.StatusNotFound:
		return "Page Not Found"
	case http.StatusMethodNotAllowed:
		return "Method Not Allowed"
	case http.StatusConflict:
		return "Conflict"
	case http.StatusUnprocessableEntity:
		return "Validation Error"
	case http.StatusTooManyRequests:
		return "Too Many Requests"
	case http.StatusInternalServerError:
		return "Server Error"
	case http.StatusBadGateway:
		return "Bad Gateway"
	case http.StatusServiceUnavailable:
		return "Service Unavailable"
	case http.StatusGatewayTimeout:
		return "Gateway Timeout"
	default:
		return "Error"
	}
}
