// Package middleware provides HTTP middleware for error handling, logging, and request processing.
// This middleware integrates with the application's error handling system to provide consistent
// error responses and proper logging.
package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"go-templ-template/internal/shared/errors"
	"go-templ-template/web/templates/pages"

	"github.com/labstack/echo/v4"
)

// ErrorHandlerConfig configures the error handling middleware
type ErrorHandlerConfig struct {
	// ShowStackTrace determines if stack traces should be shown in responses
	ShowStackTrace bool

	// LogErrors determines if errors should be logged
	LogErrors bool

	// CustomErrorPages enables custom error page rendering
	CustomErrorPages bool

	// JSONAPIErrors determines if API errors should be returned as JSON
	JSONAPIErrors bool
}

// DefaultErrorHandlerConfig returns the default error handler configuration
func DefaultErrorHandlerConfig() ErrorHandlerConfig {
	return ErrorHandlerConfig{
		ShowStackTrace:   false,
		LogErrors:        true,
		CustomErrorPages: true,
		JSONAPIErrors:    true,
	}
}

// ErrorHandler creates an error handling middleware with the given configuration
func ErrorHandler(config ErrorHandlerConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err != nil {
				return handleError(c, err, config)
			}
			return nil
		}
	}
}

// CustomErrorHandler creates a custom Echo error handler
func CustomErrorHandler(config ErrorHandlerConfig) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		handleError(c, err, config)
	}
}

// handleError processes and responds to errors based on the configuration
func handleError(c echo.Context, err error, config ErrorHandlerConfig) error {
	// Log the error if configured
	if config.LogErrors {
		logError(c, err)
	}

	// Convert to AppError if possible
	var appErr *errors.AppError
	if appError, ok := errors.AsAppError(err); ok {
		appErr = appError
	} else if echoErr, ok := err.(*echo.HTTPError); ok {
		// Convert Echo HTTP error to AppError
		appErr = convertEchoError(echoErr)
	} else {
		// Create generic internal error
		appErr = errors.NewInternalError("INTERNAL_ERROR", "An internal error occurred")
		appErr.Cause = err
	}

	// Determine response format based on request
	if isAPIRequest(c) && config.JSONAPIErrors {
		return sendJSONError(c, appErr, config)
	} else if config.CustomErrorPages {
		return sendHTMLError(c, appErr)
	} else {
		return sendPlainError(c, appErr)
	}
}

// logError logs the error with context information
func logError(c echo.Context, err error) {
	req := c.Request()

	// Extract context information
	method := req.Method
	uri := req.RequestURI
	userAgent := req.UserAgent()
	remoteAddr := c.RealIP()

	// Log with context
	if appErr, ok := errors.AsAppError(err); ok {
		log.Printf("[ERROR] %s %s - %s [%s] - IP: %s, UA: %s, Details: %+v",
			method, uri, appErr.Message, appErr.Code, remoteAddr, userAgent, appErr.Details)

		if appErr.Cause != nil {
			log.Printf("[ERROR] Caused by: %v", appErr.Cause)
		}
	} else {
		log.Printf("[ERROR] %s %s - %v - IP: %s, UA: %s",
			method, uri, err, remoteAddr, userAgent)
	}
}

// isAPIRequest determines if the request is for an API endpoint
func isAPIRequest(c echo.Context) bool {
	path := c.Request().URL.Path
	accept := c.Request().Header.Get("Accept")
	contentType := c.Request().Header.Get("Content-Type")

	// Check if path starts with /api
	if strings.HasPrefix(path, "/api/") {
		return true
	}

	// Check Accept header for JSON
	if strings.Contains(accept, "application/json") {
		return true
	}

	// Check Content-Type for JSON
	if strings.Contains(contentType, "application/json") {
		return true
	}

	// Check for AJAX requests
	if c.Request().Header.Get("X-Requested-With") == "XMLHttpRequest" {
		return true
	}

	return false
}

// sendJSONError sends a JSON error response
func sendJSONError(c echo.Context, appErr *errors.AppError, config ErrorHandlerConfig) error {
	response := appErr.ToHTTPResponse()

	// Add stack trace if configured and it's an internal error
	if config.ShowStackTrace && appErr.Type == errors.ErrorTypeInternal && appErr.Cause != nil {
		response["error"].(map[string]interface{})["stack_trace"] = fmt.Sprintf("%+v", appErr.Cause)
	}

	return c.JSON(appErr.HTTPStatus, response)
}

// sendHTMLError sends an HTML error page response
func sendHTMLError(c echo.Context, appErr *errors.AppError) error {
	// Set the appropriate HTTP status
	c.Response().WriteHeader(appErr.HTTPStatus)

	// Render appropriate error page based on error type and status
	switch appErr.HTTPStatus {
	case http.StatusNotFound:
		return pages.Error404Page().Render(c.Request().Context(), c.Response().Writer)

	case http.StatusUnauthorized:
		return pages.ErrorAuthPage().Render(c.Request().Context(), c.Response().Writer)

	case http.StatusForbidden:
		return pages.ErrorForbiddenPage().Render(c.Request().Context(), c.Response().Writer)

	case http.StatusInternalServerError:
		return pages.Error500Page().Render(c.Request().Context(), c.Response().Writer)

	default:
		// Use generic error page for other status codes
		title := getErrorTitle(appErr.HTTPStatus)
		message := appErr.Message
		if appErr.UserMessage != "" {
			message = appErr.UserMessage
		}
		code := fmt.Sprintf("%d", appErr.HTTPStatus)

		return pages.GenericErrorPage(title, message, code).Render(c.Request().Context(), c.Response().Writer)
	}
}

// sendPlainError sends a plain text error response
func sendPlainError(c echo.Context, appErr *errors.AppError) error {
	message := appErr.Message
	if appErr.UserMessage != "" {
		message = appErr.UserMessage
	}

	return c.String(appErr.HTTPStatus, message)
}

// convertEchoError converts an Echo HTTP error to an AppError
func convertEchoError(echoErr *echo.HTTPError) *errors.AppError {
	code := fmt.Sprintf("HTTP_%d", echoErr.Code)
	message := fmt.Sprintf("%v", echoErr.Message)

	var errorType errors.ErrorType
	switch echoErr.Code {
	case http.StatusBadRequest:
		errorType = errors.ErrorTypeValidation
	case http.StatusUnauthorized:
		errorType = errors.ErrorTypeAuthentication
	case http.StatusForbidden:
		errorType = errors.ErrorTypeAuthorization
	case http.StatusNotFound:
		errorType = errors.ErrorTypeNotFound
	case http.StatusConflict:
		errorType = errors.ErrorTypeConflict
	case http.StatusTooManyRequests:
		errorType = errors.ErrorTypeRateLimit
	case http.StatusServiceUnavailable:
		errorType = errors.ErrorTypeUnavailable
	case http.StatusGatewayTimeout:
		errorType = errors.ErrorTypeTimeout
	default:
		if echoErr.Code >= 500 {
			errorType = errors.ErrorTypeInternal
		} else {
			errorType = errors.ErrorTypeValidation
		}
	}

	appErr := errors.NewAppError(errorType, code, message, echoErr.Code)
	if echoErr.Internal != nil {
		appErr.Cause = echoErr.Internal
	}

	return appErr
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

// NotFoundHandler creates a custom 404 handler
func NotFoundHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		appErr := errors.NewNotFoundError("ROUTE_NOT_FOUND", "The requested route was not found")

		if isAPIRequest(c) {
			return c.JSON(http.StatusNotFound, appErr.ToHTTPResponse())
		}

		c.Response().WriteHeader(http.StatusNotFound)
		return pages.Error404Page().Render(c.Request().Context(), c.Response().Writer)
	}
}

// MethodNotAllowedHandler creates a custom method not allowed handler
func MethodNotAllowedHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		appErr := errors.NewValidationError("METHOD_NOT_ALLOWED", "The HTTP method is not allowed for this route")
		appErr.HTTPStatus = http.StatusMethodNotAllowed

		if isAPIRequest(c) {
			return c.JSON(http.StatusMethodNotAllowed, appErr.ToHTTPResponse())
		}

		c.Response().WriteHeader(http.StatusMethodNotAllowed)
		return pages.GenericErrorPage("Method Not Allowed",
			"The HTTP method you used is not allowed for this resource.", "405").
			Render(c.Request().Context(), c.Response().Writer)
	}
}

// RecoveryHandler creates a panic recovery handler that integrates with error handling
func RecoveryHandler(config ErrorHandlerConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					var err error
					if e, ok := r.(error); ok {
						err = e
					} else {
						err = fmt.Errorf("panic: %v", r)
					}

					// Create panic error
					appErr := errors.NewInternalError("PANIC_RECOVERED", "A server panic was recovered")
					appErr.Cause = err
					appErr.WithDetails(map[string]interface{}{
						"panic_value": fmt.Sprintf("%v", r),
					})

					// Handle the panic as an error
					handleError(c, appErr, config)
				}
			}()

			return next(c)
		}
	}
}
