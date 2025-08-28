package errors

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

// ErrorMiddleware provides comprehensive error handling for HTTP requests
type ErrorMiddleware struct {
	logger Logger
	config ErrorMiddlewareConfig
}

// ErrorMiddlewareConfig holds configuration for error middleware
type ErrorMiddlewareConfig struct {
	// IncludeStackTrace includes stack trace in error responses (dev only)
	IncludeStackTrace bool

	// LogAllErrors logs all errors, including client errors
	LogAllErrors bool

	// LogRequestDetails includes request details in error logs
	LogRequestDetails bool

	// HideInternalErrors hides internal error details from clients
	HideInternalErrors bool

	// CustomErrorHandler allows custom error handling logic
	CustomErrorHandler func(c echo.Context, err error) error
}

// DefaultErrorMiddlewareConfig returns default configuration
func DefaultErrorMiddlewareConfig() ErrorMiddlewareConfig {
	return ErrorMiddlewareConfig{
		IncludeStackTrace:  false,
		LogAllErrors:       true,
		LogRequestDetails:  true,
		HideInternalErrors: true,
	}
}

// Logger interface for structured logging
type Logger interface {
	WithFields(fields map[string]interface{}) Logger
	WithError(err error) Logger
	Error(msg string)
	Warn(msg string)
	Info(msg string)
	Debug(msg string)
}

// LogrusLogger implements Logger interface using logrus
type LogrusLogger struct {
	logger *logrus.Logger
	entry  *logrus.Entry
}

// NewLogrusLogger creates a new LogrusLogger
func NewLogrusLogger(logger *logrus.Logger) *LogrusLogger {
	return &LogrusLogger{
		logger: logger,
		entry:  logrus.NewEntry(logger),
	}
}

func (l *LogrusLogger) WithFields(fields map[string]interface{}) Logger {
	return &LogrusLogger{
		logger: l.logger,
		entry:  l.entry.WithFields(fields),
	}
}

func (l *LogrusLogger) WithError(err error) Logger {
	return &LogrusLogger{
		logger: l.logger,
		entry:  l.entry.WithError(err),
	}
}

func (l *LogrusLogger) Error(msg string) { l.entry.Error(msg) }
func (l *LogrusLogger) Warn(msg string)  { l.entry.Warn(msg) }
func (l *LogrusLogger) Info(msg string)  { l.entry.Info(msg) }
func (l *LogrusLogger) Debug(msg string) { l.entry.Debug(msg) }

// NewErrorMiddleware creates a new error middleware instance
func NewErrorMiddleware(logger Logger, config ErrorMiddlewareConfig) *ErrorMiddleware {
	return &ErrorMiddleware{
		logger: logger,
		config: config,
	}
}

// Handler returns the error handling middleware function
func (m *ErrorMiddleware) Handler() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Add request ID to context if not present
			requestID := c.Response().Header().Get(echo.HeaderXRequestID)
			if requestID == "" {
				requestID = generateErrorID()
				c.Response().Header().Set(echo.HeaderXRequestID, requestID)
			}

			// Execute the handler
			err := next(c)
			if err != nil {
				return m.handleError(c, err, requestID)
			}

			return nil
		}
	}
}

// handleError processes and responds to errors
func (m *ErrorMiddleware) handleError(c echo.Context, err error, requestID string) error {
	// Create error context
	errorCtx := m.createErrorContext(c, requestID)

	// Convert to AppError if needed
	appErr := m.convertToAppError(err, errorCtx)

	// Log the error
	m.logError(appErr, errorCtx)

	// Call custom error handler if provided
	if m.config.CustomErrorHandler != nil {
		if customErr := m.config.CustomErrorHandler(c, appErr); customErr != nil {
			return customErr
		}
	}

	// Send error response
	return m.sendErrorResponse(c, appErr)
}

// createErrorContext creates error context from the request
func (m *ErrorMiddleware) createErrorContext(c echo.Context, requestID string) ErrorContext {
	ctx := ErrorContext{
		RequestID: requestID,
		IPAddress: c.RealIP(),
		UserAgent: c.Request().UserAgent(),
		Metadata: map[string]interface{}{
			"method": c.Request().Method,
			"path":   c.Request().URL.Path,
			"query":  c.Request().URL.RawQuery,
		},
	}

	// Add user context if available
	if userID := getUserIDFromContext(c); userID != "" {
		ctx.UserID = userID
	}

	// Add session context if available
	if sessionID := getSessionIDFromContext(c); sessionID != "" {
		ctx.SessionID = sessionID
	}

	return ctx
}

// convertToAppError converts any error to AppError
func (m *ErrorMiddleware) convertToAppError(err error, ctx ErrorContext) *AppError {
	// If it's already an AppError, add context and return
	if appErr, ok := err.(*AppError); ok {
		return appErr.WithContext(ctx)
	}

	// Handle Echo HTTP errors
	if httpErr, ok := err.(*echo.HTTPError); ok {
		return m.convertHTTPError(httpErr, ctx)
	}

	// Handle standard errors
	return m.convertStandardError(err, ctx)
}

// convertHTTPError converts Echo HTTP error to AppError
func (m *ErrorMiddleware) convertHTTPError(httpErr *echo.HTTPError, ctx ErrorContext) *AppError {
	var errorType ErrorType
	var code string

	switch httpErr.Code {
	case http.StatusBadRequest:
		errorType = ErrorTypeValidation
		code = "BAD_REQUEST"
	case http.StatusUnauthorized:
		errorType = ErrorTypeAuthentication
		code = "UNAUTHORIZED"
	case http.StatusForbidden:
		errorType = ErrorTypeAuthorization
		code = "FORBIDDEN"
	case http.StatusNotFound:
		errorType = ErrorTypeNotFound
		code = "NOT_FOUND"
	case http.StatusConflict:
		errorType = ErrorTypeConflict
		code = "CONFLICT"
	case http.StatusTooManyRequests:
		errorType = ErrorTypeRateLimit
		code = "RATE_LIMIT_EXCEEDED"
	case http.StatusRequestTimeout:
		errorType = ErrorTypeTimeout
		code = "REQUEST_TIMEOUT"
	case http.StatusServiceUnavailable:
		errorType = ErrorTypeUnavailable
		code = "SERVICE_UNAVAILABLE"
	default:
		errorType = ErrorTypeInternal
		code = "INTERNAL_ERROR"
	}

	message := fmt.Sprintf("%v", httpErr.Message)
	if message == "" {
		message = http.StatusText(httpErr.Code)
	}

	appErr := NewAppError(errorType, code, message, httpErr.Code).WithContext(ctx)

	// Add internal error as cause if it exists
	if httpErr.Internal != nil {
		appErr.Cause = httpErr.Internal
	}

	return appErr
}

// convertStandardError converts standard error to AppError
func (m *ErrorMiddleware) convertStandardError(err error, ctx ErrorContext) *AppError {
	// Check for context cancellation
	if err == context.Canceled {
		return NewAppError(ErrorTypeTimeout, "REQUEST_CANCELED",
			"Request was canceled", http.StatusRequestTimeout).WithContext(ctx)
	}

	if err == context.DeadlineExceeded {
		return NewAppError(ErrorTypeTimeout, "REQUEST_TIMEOUT",
			"Request timed out", http.StatusRequestTimeout).WithContext(ctx)
	}

	// Default to internal error
	return NewInternalErrorWithCause("UNHANDLED_ERROR",
		"An unexpected error occurred", err).WithContext(ctx)
}

// logError logs the error with appropriate level and context
func (m *ErrorMiddleware) logError(appErr *AppError, ctx ErrorContext) {
	// Prepare log fields
	fields := map[string]interface{}{
		"error_id":    appErr.ID,
		"error_code":  appErr.Code,
		"error_type":  appErr.Type,
		"severity":    appErr.Severity,
		"http_status": appErr.HTTPStatus,
		"timestamp":   appErr.Timestamp,
		"retryable":   appErr.Retryable,
	}

	// Add context fields
	if ctx.RequestID != "" {
		fields["request_id"] = ctx.RequestID
	}
	if ctx.UserID != "" {
		fields["user_id"] = ctx.UserID
	}
	if ctx.SessionID != "" {
		fields["session_id"] = ctx.SessionID
	}
	if ctx.IPAddress != "" {
		fields["ip_address"] = ctx.IPAddress
	}
	if ctx.Operation != "" {
		fields["operation"] = ctx.Operation
	}
	if ctx.Component != "" {
		fields["component"] = ctx.Component
	}

	// Add request details if configured
	if m.config.LogRequestDetails {
		for k, v := range ctx.Metadata {
			fields[fmt.Sprintf("request_%s", k)] = v
		}
	}

	// Add error details (filtered)
	if len(appErr.Details) > 0 {
		for k, v := range appErr.Details {
			if !isSensitiveField(k) {
				fields[fmt.Sprintf("detail_%s", k)] = v
			}
		}
	}

	// Create logger with fields
	logger := m.logger.WithFields(fields)

	// Add underlying error if present
	if appErr.Cause != nil {
		logger = logger.WithError(appErr.Cause)

		// Add stack trace for internal errors in development
		if m.config.IncludeStackTrace && appErr.Severity >= SeverityHigh {
			fields["stack_trace"] = string(debug.Stack())
		}
	}

	// Log based on severity and configuration
	message := fmt.Sprintf("Error occurred: %s", appErr.Message)

	switch appErr.Severity {
	case SeverityCritical:
		logger.Error(message)
	case SeverityHigh:
		logger.Error(message)
	case SeverityMedium:
		logger.Warn(message)
	case SeverityLow:
		if m.config.LogAllErrors {
			logger.Info(message)
		}
	default:
		logger.Error(message)
	}
}

// sendErrorResponse sends the error response to the client
func (m *ErrorMiddleware) sendErrorResponse(c echo.Context, appErr *AppError) error {
	// Prepare response
	response := appErr.ToHTTPResponse()

	// Add request ID to response
	if appErr.Context.RequestID != "" {
		response["request_id"] = appErr.Context.RequestID
	}

	// Hide internal error details in production
	if m.config.HideInternalErrors && appErr.Severity >= SeverityHigh {
		if errorMap, ok := response["error"].(map[string]interface{}); ok {
			// Keep only safe fields
			safeResponse := map[string]interface{}{
				"error": map[string]interface{}{
					"id":        errorMap["id"],
					"code":      "INTERNAL_ERROR",
					"type":      "internal",
					"message":   "An internal error occurred",
					"timestamp": errorMap["timestamp"],
				},
			}

			// Keep user message if available
			if userMsg, exists := errorMap["user_message"]; exists {
				safeResponse["error"].(map[string]interface{})["user_message"] = userMsg
			}

			// Keep retryable flag
			if retryable, exists := errorMap["retryable"]; exists {
				safeResponse["error"].(map[string]interface{})["retryable"] = retryable
			}

			response = safeResponse
		}
	}

	// Add stack trace in development for debugging
	if m.config.IncludeStackTrace && appErr.Cause != nil {
		if errorMap, ok := response["error"].(map[string]interface{}); ok {
			errorMap["stack_trace"] = string(debug.Stack())
		}
	}

	return c.JSON(appErr.HTTPStatus, response)
}

// Helper functions to extract context information

func getUserIDFromContext(c echo.Context) string {
	if user := c.Get("user"); user != nil {
		// Try to extract user ID from user object
		if userMap, ok := user.(map[string]interface{}); ok {
			if id, exists := userMap["id"]; exists {
				return fmt.Sprintf("%v", id)
			}
		}
		// Try to get ID field directly
		if userWithID, ok := user.(interface{ GetID() string }); ok {
			return userWithID.GetID()
		}
	}
	return ""
}

func getSessionIDFromContext(c echo.Context) string {
	if session := c.Get("session"); session != nil {
		// Try to extract session ID from session object
		if sessionMap, ok := session.(map[string]interface{}); ok {
			if id, exists := sessionMap["id"]; exists {
				return fmt.Sprintf("%v", id)
			}
		}
		// Try to get ID field directly
		if sessionWithID, ok := session.(interface{ GetID() string }); ok {
			return sessionWithID.GetID()
		}
	}
	return ""
}

// Recovery middleware that converts panics to AppErrors
func (m *ErrorMiddleware) RecoveryHandler() echo.MiddlewareFunc {
	return middleware.RecoverWithConfig(middleware.RecoverConfig{
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			// Create panic error
			panicErr := NewInternalError("PANIC_RECOVERED", "A panic occurred during request processing")
			panicErr.Cause = err
			panicErr.Severity = SeverityCritical

			// Add context
			requestID := c.Response().Header().Get(echo.HeaderXRequestID)
			if requestID == "" {
				requestID = generateErrorID()
				c.Response().Header().Set(echo.HeaderXRequestID, requestID)
			}

			errorCtx := m.createErrorContext(c, requestID)
			panicErr = panicErr.WithContext(errorCtx)

			// Add stack trace to details
			panicErr = panicErr.WithDetails(map[string]interface{}{
				"stack_trace": string(stack),
				"panic_value": fmt.Sprintf("%v", err),
			})

			// Log the panic
			m.logError(panicErr, errorCtx)

			// Send error response
			return m.sendErrorResponse(c, panicErr)
		},
	})
}
