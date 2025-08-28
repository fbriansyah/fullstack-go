package errors

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

// LoggingConfig holds configuration for structured logging
type LoggingConfig struct {
	// Level is the minimum log level to output
	Level string `json:"level" yaml:"level"`

	// Format is the log format (json, text)
	Format string `json:"format" yaml:"format"`

	// Output is the output destination (stdout, stderr, file path)
	Output string `json:"output" yaml:"output"`

	// IncludeStackTrace includes stack traces for errors
	IncludeStackTrace bool `json:"include_stack_trace" yaml:"include_stack_trace"`

	// IncludeCaller includes caller information in logs
	IncludeCaller bool `json:"include_caller" yaml:"include_caller"`

	// TimestampFormat is the format for timestamps
	TimestampFormat string `json:"timestamp_format" yaml:"timestamp_format"`

	// ServiceName is the name of the service for structured logging
	ServiceName string `json:"service_name" yaml:"service_name"`

	// ServiceVersion is the version of the service
	ServiceVersion string `json:"service_version" yaml:"service_version"`

	// Environment is the deployment environment
	Environment string `json:"environment" yaml:"environment"`
}

// DefaultLoggingConfig returns default logging configuration
func DefaultLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Level:             "info",
		Format:            "json",
		Output:            "stdout",
		IncludeStackTrace: false,
		IncludeCaller:     true,
		TimestampFormat:   time.RFC3339,
		ServiceName:       "go-templ-template",
		ServiceVersion:    "1.0.0",
		Environment:       "development",
	}
}

// StructuredLogger provides structured logging with contextual information
type StructuredLogger struct {
	logger *logrus.Logger
	config LoggingConfig
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger(config LoggingConfig) (*StructuredLogger, error) {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level '%s': %w", config.Level, err)
	}
	logger.SetLevel(level)

	// Set formatter
	switch config.Format {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: config.TimestampFormat,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "caller",
			},
		})
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: config.TimestampFormat,
			FullTimestamp:   true,
		})
	default:
		return nil, fmt.Errorf("unsupported log format '%s'", config.Format)
	}

	// Set output
	switch config.Output {
	case "stdout":
		logger.SetOutput(os.Stdout)
	case "stderr":
		logger.SetOutput(os.Stderr)
	default:
		// Assume it's a file path
		if err := os.MkdirAll(filepath.Dir(config.Output), 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		file, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		logger.SetOutput(file)
	}

	// Set caller reporting
	logger.SetReportCaller(config.IncludeCaller)

	return &StructuredLogger{
		logger: logger,
		config: config,
	}, nil
}

// WithContext creates a logger with context information
func (l *StructuredLogger) WithContext(ctx context.Context) *ContextLogger {
	return &ContextLogger{
		logger: l.logger,
		config: l.config,
		ctx:    ctx,
		fields: make(map[string]interface{}),
	}
}

// WithFields creates a logger with additional fields
func (l *StructuredLogger) WithFields(fields map[string]interface{}) Logger {
	entry := l.logger.WithFields(l.addServiceFields(fields))
	return &LogrusLogger{
		logger: l.logger,
		entry:  entry,
	}
}

// WithError creates a logger with error information
func (l *StructuredLogger) WithError(err error) Logger {
	fields := l.addServiceFields(map[string]interface{}{})
	entry := l.logger.WithError(err).WithFields(fields)

	// Add stack trace if configured and it's an AppError
	if l.config.IncludeStackTrace {
		if appErr, ok := err.(*AppError); ok && appErr.Cause != nil {
			entry = entry.WithField("stack_trace", getStackTrace())
		}
	}

	return &LogrusLogger{
		logger: l.logger,
		entry:  entry,
	}
}

// Error logs an error message
func (l *StructuredLogger) Error(msg string) {
	l.logger.WithFields(l.addServiceFields(nil)).Error(msg)
}

// Warn logs a warning message
func (l *StructuredLogger) Warn(msg string) {
	l.logger.WithFields(l.addServiceFields(nil)).Warn(msg)
}

// Info logs an info message
func (l *StructuredLogger) Info(msg string) {
	l.logger.WithFields(l.addServiceFields(nil)).Info(msg)
}

// Debug logs a debug message
func (l *StructuredLogger) Debug(msg string) {
	l.logger.WithFields(l.addServiceFields(nil)).Debug(msg)
}

// Close closes the logger and any associated resources
func (l *StructuredLogger) Close() error {
	// If the logger is writing to a file, close it
	if file, ok := l.logger.Out.(*os.File); ok && file != os.Stdout && file != os.Stderr {
		return file.Close()
	}
	return nil
}

// addServiceFields adds service-level fields to log entries
func (l *StructuredLogger) addServiceFields(fields map[string]interface{}) logrus.Fields {
	if fields == nil {
		fields = make(map[string]interface{})
	}

	fields["service"] = l.config.ServiceName
	fields["version"] = l.config.ServiceVersion
	fields["environment"] = l.config.Environment

	return logrus.Fields(fields)
}

// ContextLogger provides logging with request context
type ContextLogger struct {
	logger *logrus.Logger
	config LoggingConfig
	ctx    context.Context
	fields map[string]interface{}
}

// WithFields adds fields to the context logger
func (cl *ContextLogger) WithFields(fields map[string]interface{}) Logger {
	newFields := make(map[string]interface{})

	// Copy existing fields
	for k, v := range cl.fields {
		newFields[k] = v
	}

	// Add new fields
	for k, v := range fields {
		newFields[k] = v
	}

	// Add service fields
	newFields["service"] = cl.config.ServiceName
	newFields["version"] = cl.config.ServiceVersion
	newFields["environment"] = cl.config.Environment

	// Add context fields
	cl.addContextFields(newFields)

	entry := cl.logger.WithFields(logrus.Fields(newFields))
	return &LogrusLogger{
		logger: cl.logger,
		entry:  entry,
	}
}

// WithError adds error information to the context logger
func (cl *ContextLogger) WithError(err error) Logger {
	fields := make(map[string]interface{})

	// Copy existing fields
	for k, v := range cl.fields {
		fields[k] = v
	}

	// Add service fields
	fields["service"] = cl.config.ServiceName
	fields["version"] = cl.config.ServiceVersion
	fields["environment"] = cl.config.Environment

	// Add context fields
	cl.addContextFields(fields)

	entry := cl.logger.WithError(err).WithFields(logrus.Fields(fields))

	// Add stack trace if configured and it's an AppError
	if cl.config.IncludeStackTrace {
		if appErr, ok := err.(*AppError); ok && appErr.Cause != nil {
			entry = entry.WithField("stack_trace", getStackTrace())
		}
	}

	return &LogrusLogger{
		logger: cl.logger,
		entry:  entry,
	}
}

// Error logs an error message with context
func (cl *ContextLogger) Error(msg string) {
	cl.WithFields(nil).Error(msg)
}

// Warn logs a warning message with context
func (cl *ContextLogger) Warn(msg string) {
	cl.WithFields(nil).Warn(msg)
}

// Info logs an info message with context
func (cl *ContextLogger) Info(msg string) {
	cl.WithFields(nil).Info(msg)
}

// Debug logs a debug message with context
func (cl *ContextLogger) Debug(msg string) {
	cl.WithFields(nil).Debug(msg)
}

// addContextFields extracts relevant fields from context
func (cl *ContextLogger) addContextFields(fields map[string]interface{}) {
	// Add request ID if available
	if requestID := getRequestIDFromContext(cl.ctx); requestID != "" {
		fields["request_id"] = requestID
	}

	// Add user ID if available
	if userID := getUserIDFromContextValue(cl.ctx); userID != "" {
		fields["user_id"] = userID
	}

	// Add session ID if available
	if sessionID := getSessionIDFromContextValue(cl.ctx); sessionID != "" {
		fields["session_id"] = sessionID
	}

	// Add trace ID if available (for distributed tracing)
	if traceID := getTraceIDFromContext(cl.ctx); traceID != "" {
		fields["trace_id"] = traceID
	}
}

// ErrorLogger provides specialized error logging
type ErrorLogger struct {
	logger Logger
	config LoggingConfig
}

// NewErrorLogger creates a new error logger
func NewErrorLogger(logger Logger, config LoggingConfig) *ErrorLogger {
	return &ErrorLogger{
		logger: logger,
		config: config,
	}
}

// LogError logs an AppError with full context
func (el *ErrorLogger) LogError(err *AppError) {
	fields := map[string]interface{}{
		"error_id":    err.ID,
		"error_code":  err.Code,
		"error_type":  err.Type,
		"severity":    err.Severity,
		"http_status": err.HTTPStatus,
		"retryable":   err.Retryable,
		"timestamp":   err.Timestamp,
	}

	// Add context fields
	if err.Context.RequestID != "" {
		fields["request_id"] = err.Context.RequestID
	}
	if err.Context.UserID != "" {
		fields["user_id"] = err.Context.UserID
	}
	if err.Context.SessionID != "" {
		fields["session_id"] = err.Context.SessionID
	}
	if err.Context.IPAddress != "" {
		fields["ip_address"] = err.Context.IPAddress
	}
	if err.Context.Operation != "" {
		fields["operation"] = err.Context.Operation
	}
	if err.Context.Component != "" {
		fields["component"] = err.Context.Component
	}

	// Add metadata
	for k, v := range err.Context.Metadata {
		fields[fmt.Sprintf("context_%s", k)] = v
	}

	// Add error details (filtered)
	for k, v := range err.Details {
		if !isSensitiveField(k) {
			fields[fmt.Sprintf("detail_%s", k)] = v
		}
	}

	// Create logger with fields
	logger := el.logger.WithFields(fields)

	// Add underlying error if present
	if err.Cause != nil {
		logger = logger.WithError(err.Cause)
	}

	// Log based on severity
	message := fmt.Sprintf("Error occurred: %s", err.Message)

	switch err.Severity {
	case SeverityCritical, SeverityHigh:
		logger.Error(message)
	case SeverityMedium:
		logger.Warn(message)
	case SeverityLow:
		logger.Info(message)
	default:
		logger.Error(message)
	}
}

// Helper functions for context extraction

func getRequestIDFromContext(ctx context.Context) string {
	if requestID := ctx.Value("request_id"); requestID != nil {
		return fmt.Sprintf("%v", requestID)
	}
	return ""
}

func getUserIDFromContextValue(ctx context.Context) string {
	if userID := ctx.Value("user_id"); userID != nil {
		return fmt.Sprintf("%v", userID)
	}
	return ""
}

func getSessionIDFromContextValue(ctx context.Context) string {
	if sessionID := ctx.Value("session_id"); sessionID != nil {
		return fmt.Sprintf("%v", sessionID)
	}
	return ""
}

func getTraceIDFromContext(ctx context.Context) string {
	if traceID := ctx.Value("trace_id"); traceID != nil {
		return fmt.Sprintf("%v", traceID)
	}
	return ""
}

func getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}
