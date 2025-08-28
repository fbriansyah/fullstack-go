package errors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultLoggingConfig(t *testing.T) {
	config := DefaultLoggingConfig()

	assert.Equal(t, "info", config.Level)
	assert.Equal(t, "json", config.Format)
	assert.Equal(t, "stdout", config.Output)
	assert.False(t, config.IncludeStackTrace)
	assert.True(t, config.IncludeCaller)
	assert.Equal(t, time.RFC3339, config.TimestampFormat)
	assert.Equal(t, "go-templ-template", config.ServiceName)
	assert.Equal(t, "1.0.0", config.ServiceVersion)
	assert.Equal(t, "development", config.Environment)
}

func TestNewStructuredLogger(t *testing.T) {
	tests := []struct {
		name        string
		config      LoggingConfig
		expectError bool
	}{
		{
			name:        "valid config",
			config:      DefaultLoggingConfig(),
			expectError: false,
		},
		{
			name: "invalid log level",
			config: LoggingConfig{
				Level:  "invalid",
				Format: "json",
				Output: "stdout",
			},
			expectError: true,
		},
		{
			name: "invalid format",
			config: LoggingConfig{
				Level:  "info",
				Format: "invalid",
				Output: "stdout",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewStructuredLogger(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, logger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
			}
		})
	}
}

func TestStructuredLogger_JSONFormat(t *testing.T) {
	var buf bytes.Buffer

	config := LoggingConfig{
		Level:           "debug",
		Format:          "json",
		Output:          "stdout",
		ServiceName:     "test-service",
		ServiceVersion:  "1.0.0",
		Environment:     "test",
		TimestampFormat: time.RFC3339,
	}

	logger, err := NewStructuredLogger(config)
	require.NoError(t, err)

	// Redirect output to buffer
	logger.logger.SetOutput(&buf)

	// Log a message
	logger.Info("test message")

	// Parse JSON output
	var logEntry map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	// Verify structure
	assert.Equal(t, "info", logEntry["level"])
	assert.Equal(t, "test message", logEntry["message"])
	assert.Equal(t, "test-service", logEntry["service"])
	assert.Equal(t, "1.0.0", logEntry["version"])
	assert.Equal(t, "test", logEntry["environment"])
	assert.Contains(t, logEntry, "timestamp")
}

func TestStructuredLogger_TextFormat(t *testing.T) {
	var buf bytes.Buffer

	config := LoggingConfig{
		Level:           "info",
		Format:          "text",
		Output:          "stdout",
		ServiceName:     "test-service",
		ServiceVersion:  "1.0.0",
		Environment:     "test",
		TimestampFormat: time.RFC3339,
	}

	logger, err := NewStructuredLogger(config)
	require.NoError(t, err)

	// Redirect output to buffer
	logger.logger.SetOutput(&buf)

	// Log a message
	logger.Info("test message")

	output := buf.String()
	assert.Contains(t, output, "test message")
	assert.Contains(t, output, "info")
	assert.Contains(t, output, "test-service")
}

func TestStructuredLogger_FileOutput(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	config := LoggingConfig{
		Level:          "info",
		Format:         "json",
		Output:         logFile,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
	}

	logger, err := NewStructuredLogger(config)
	require.NoError(t, err)
	defer logger.Close() // Close the logger to release file handle

	// Log a message
	logger.Info("test message")

	// Close the logger before reading the file
	err = logger.Close()
	require.NoError(t, err)

	// Verify file was created and contains log entry
	assert.FileExists(t, logFile)

	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	var logEntry map[string]interface{}
	err = json.Unmarshal(content, &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "test message", logEntry["message"])
	assert.Equal(t, "test-service", logEntry["service"])
}

func TestStructuredLogger_WithFields(t *testing.T) {
	var buf bytes.Buffer

	config := DefaultLoggingConfig()
	config.Format = "json"

	logger, err := NewStructuredLogger(config)
	require.NoError(t, err)

	logger.logger.SetOutput(&buf)

	// Log with fields
	fields := map[string]interface{}{
		"user_id":    "123",
		"request_id": "req-456",
	}

	logger.WithFields(fields).Info("test message")

	// Parse JSON output
	var logEntry map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "123", logEntry["user_id"])
	assert.Equal(t, "req-456", logEntry["request_id"])
	assert.Equal(t, config.ServiceName, logEntry["service"])
}

func TestStructuredLogger_WithError(t *testing.T) {
	var buf bytes.Buffer

	config := DefaultLoggingConfig()
	config.Format = "json"
	config.IncludeStackTrace = true

	logger, err := NewStructuredLogger(config)
	require.NoError(t, err)

	logger.logger.SetOutput(&buf)

	// Create AppError with cause
	appErr := NewInternalErrorWithCause("TEST_ERROR", "Test error", fmt.Errorf("underlying error"))

	logger.WithError(appErr).Error("error occurred")

	// Parse JSON output
	var logEntry map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Contains(t, logEntry, "error")
	assert.Equal(t, config.ServiceName, logEntry["service"])

	// Should include stack trace for AppError with cause
	if config.IncludeStackTrace {
		assert.Contains(t, logEntry, "stack_trace")
	}
}

func TestContextLogger(t *testing.T) {
	var buf bytes.Buffer

	config := DefaultLoggingConfig()
	config.Format = "json"

	logger, err := NewStructuredLogger(config)
	require.NoError(t, err)

	logger.logger.SetOutput(&buf)

	// Create context with values
	ctx := context.Background()
	ctx = context.WithValue(ctx, "request_id", "req-123")
	ctx = context.WithValue(ctx, "user_id", "user-456")

	// Log with context
	contextLogger := logger.WithContext(ctx)
	contextLogger.Info("test message")

	// Parse JSON output
	var logEntry map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "req-123", logEntry["request_id"])
	assert.Equal(t, "user-456", logEntry["user_id"])
	assert.Equal(t, config.ServiceName, logEntry["service"])
}

func TestContextLogger_WithFields(t *testing.T) {
	var buf bytes.Buffer

	config := DefaultLoggingConfig()
	config.Format = "json"

	logger, err := NewStructuredLogger(config)
	require.NoError(t, err)

	logger.logger.SetOutput(&buf)

	// Create context logger and add fields
	ctx := context.Background()
	contextLogger := logger.WithContext(ctx)

	fields := map[string]interface{}{
		"operation": "user_creation",
		"component": "user_service",
	}

	contextLogger.WithFields(fields).Info("operation completed")

	// Parse JSON output
	var logEntry map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "user_creation", logEntry["operation"])
	assert.Equal(t, "user_service", logEntry["component"])
	assert.Equal(t, config.ServiceName, logEntry["service"])
}

func TestErrorLogger(t *testing.T) {
	var buf bytes.Buffer

	config := DefaultLoggingConfig()
	config.Format = "json"

	structuredLogger, err := NewStructuredLogger(config)
	require.NoError(t, err)

	structuredLogger.logger.SetOutput(&buf)

	logger := structuredLogger.WithFields(nil)
	errorLogger := NewErrorLogger(logger, config)

	// Create AppError with full context
	appErr := NewValidationError("INVALID_INPUT", "Invalid input provided")
	appErr = appErr.WithContext(ErrorContext{
		Operation: "user_creation",
		Component: "user_service",
		UserID:    "user123",
		RequestID: "req456",
		SessionID: "session789",
		IPAddress: "192.168.1.1",
		Metadata: map[string]interface{}{
			"method": "POST",
			"path":   "/api/users",
		},
	})
	appErr = appErr.WithDetails(map[string]interface{}{
		"field": "email",
		"value": "invalid-email",
	})

	// Log the error
	errorLogger.LogError(appErr)

	// Parse JSON output
	var logEntry map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	// Verify error fields
	assert.Equal(t, appErr.ID, logEntry["error_id"])
	assert.Equal(t, appErr.Code, logEntry["error_code"])
	assert.Equal(t, string(appErr.Type), logEntry["error_type"])
	assert.Equal(t, string(appErr.Severity), logEntry["severity"])

	// Verify context fields
	assert.Equal(t, "user_creation", logEntry["operation"])
	assert.Equal(t, "user_service", logEntry["component"])
	assert.Equal(t, "user123", logEntry["user_id"])
	assert.Equal(t, "req456", logEntry["request_id"])
	assert.Equal(t, "session789", logEntry["session_id"])
	assert.Equal(t, "192.168.1.1", logEntry["ip_address"])

	// Verify metadata fields
	assert.Equal(t, "POST", logEntry["context_method"])
	assert.Equal(t, "/api/users", logEntry["context_path"])

	// Verify detail fields
	assert.Equal(t, "email", logEntry["detail_field"])
	assert.Equal(t, "invalid-email", logEntry["detail_value"])
}

func TestErrorLogger_SeverityLevels(t *testing.T) {
	tests := []struct {
		name          string
		severity      ErrorSeverity
		expectedLevel logrus.Level
	}{
		{"critical severity", SeverityCritical, logrus.ErrorLevel},
		{"high severity", SeverityHigh, logrus.ErrorLevel},
		{"medium severity", SeverityMedium, logrus.WarnLevel},
		{"low severity", SeverityLow, logrus.InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			var capturedLevel logrus.Level

			config := DefaultLoggingConfig()
			config.Format = "json"

			structuredLogger, err := NewStructuredLogger(config)
			require.NoError(t, err)

			structuredLogger.logger.SetOutput(&buf)

			// Add hook to capture log level
			structuredLogger.logger.AddHook(&levelCaptureHook{
				callback: func(level logrus.Level) {
					capturedLevel = level
				},
			})

			logger := structuredLogger.WithFields(nil)
			errorLogger := NewErrorLogger(logger, config)

			// Create error with specific severity
			appErr := NewValidationError("TEST_ERROR", "Test error")
			appErr.Severity = tt.severity

			errorLogger.LogError(appErr)

			assert.Equal(t, tt.expectedLevel, capturedLevel)
		})
	}
}

func TestLogrusLogger_Implementation(t *testing.T) {
	logrusLogger := logrus.New()
	logger := NewLogrusLogger(logrusLogger)

	// Test that it implements Logger interface
	var _ Logger = logger

	// Test methods don't panic
	assert.NotPanics(t, func() {
		logger.Error("test error")
		logger.Warn("test warning")
		logger.Info("test info")
		logger.Debug("test debug")
	})

	// Test WithFields
	fieldsLogger := logger.WithFields(map[string]interface{}{
		"test": "value",
	})
	assert.NotNil(t, fieldsLogger)

	// Test WithError
	errorLogger := logger.WithError(fmt.Errorf("test error"))
	assert.NotNil(t, errorLogger)
}

func TestContextExtraction(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected map[string]string
	}{
		{
			name: "context with all values",
			ctx: func() context.Context {
				ctx := context.Background()
				ctx = context.WithValue(ctx, "request_id", "req-123")
				ctx = context.WithValue(ctx, "user_id", "user-456")
				ctx = context.WithValue(ctx, "session_id", "session-789")
				ctx = context.WithValue(ctx, "trace_id", "trace-abc")
				return ctx
			}(),
			expected: map[string]string{
				"request_id": "req-123",
				"user_id":    "user-456",
				"session_id": "session-789",
				"trace_id":   "trace-abc",
			},
		},
		{
			name: "context with no values",
			ctx:  context.Background(),
			expected: map[string]string{
				"request_id": "",
				"user_id":    "",
				"session_id": "",
				"trace_id":   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected["request_id"], getRequestIDFromContext(tt.ctx))
			assert.Equal(t, tt.expected["user_id"], getUserIDFromContextValue(tt.ctx))
			assert.Equal(t, tt.expected["session_id"], getSessionIDFromContextValue(tt.ctx))
			assert.Equal(t, tt.expected["trace_id"], getTraceIDFromContext(tt.ctx))
		})
	}
}

func TestGetStackTrace(t *testing.T) {
	stackTrace := getStackTrace()

	assert.NotEmpty(t, stackTrace)
	assert.Contains(t, stackTrace, "TestGetStackTrace")
	assert.Contains(t, stackTrace, "logging_test.go")
}

// Helper types for testing

type levelCaptureHook struct {
	callback func(logrus.Level)
}

func (h *levelCaptureHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *levelCaptureHook) Fire(entry *logrus.Entry) error {
	if h.callback != nil {
		h.callback(entry.Level)
	}
	return nil
}

func TestStructuredLogger_LogLevels(t *testing.T) {
	tests := []struct {
		name        string
		configLevel string
		logLevel    string
		shouldLog   bool
	}{
		{"debug config logs debug", "debug", "debug", true},
		{"debug config logs info", "debug", "info", true},
		{"debug config logs warn", "debug", "warn", true},
		{"debug config logs error", "debug", "error", true},
		{"info config skips debug", "info", "debug", false},
		{"info config logs info", "info", "info", true},
		{"info config logs warn", "info", "warn", true},
		{"info config logs error", "info", "error", true},
		{"warn config skips debug", "warn", "debug", false},
		{"warn config skips info", "warn", "info", false},
		{"warn config logs warn", "warn", "warn", true},
		{"warn config logs error", "warn", "error", true},
		{"error config skips debug", "error", "debug", false},
		{"error config skips info", "error", "info", false},
		{"error config skips warn", "error", "warn", false},
		{"error config logs error", "error", "error", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logged := false

			config := DefaultLoggingConfig()
			config.Level = tt.configLevel
			config.Format = "json"

			logger, err := NewStructuredLogger(config)
			require.NoError(t, err)

			logger.logger.SetOutput(&buf)
			logger.logger.AddHook(&levelCaptureHook{
				callback: func(level logrus.Level) {
					logged = true
				},
			})

			// Log at the specified level
			switch tt.logLevel {
			case "debug":
				logger.Debug("test message")
			case "info":
				logger.Info("test message")
			case "warn":
				logger.Warn("test message")
			case "error":
				logger.Error("test message")
			}

			assert.Equal(t, tt.shouldLog, logged)

			if tt.shouldLog {
				assert.NotEmpty(t, buf.String())
			}
		})
	}
}

func TestStructuredLogger_Integration(t *testing.T) {
	// Create temporary log file
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "integration.log")

	config := LoggingConfig{
		Level:             "debug",
		Format:            "json",
		Output:            logFile,
		IncludeStackTrace: true,
		IncludeCaller:     true,
		ServiceName:       "integration-test",
		ServiceVersion:    "1.0.0",
		Environment:       "test",
	}

	logger, err := NewStructuredLogger(config)
	require.NoError(t, err)
	defer logger.Close() // Close the logger to release file handle

	// Test various logging scenarios
	t.Run("basic logging", func(t *testing.T) {
		logger.Info("integration test started")
		logger.Debug("debug information")
		logger.Warn("warning message")
		logger.Error("error message")
	})

	t.Run("logging with fields", func(t *testing.T) {
		logger.WithFields(map[string]interface{}{
			"test_case": "with_fields",
			"iteration": 1,
		}).Info("logging with additional fields")
	})

	t.Run("logging with error", func(t *testing.T) {
		testErr := NewInternalErrorWithCause("TEST_ERROR", "Test error for logging",
			fmt.Errorf("underlying cause"))

		logger.WithError(testErr).Error("error with cause occurred")
	})

	t.Run("context logging", func(t *testing.T) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, "request_id", "integration-req-123")
		ctx = context.WithValue(ctx, "user_id", "integration-user-456")

		contextLogger := logger.WithContext(ctx)
		contextLogger.WithFields(map[string]interface{}{
			"operation": "integration_test",
		}).Info("context logging test")
	})

	// Close the logger before reading the file
	err = logger.Close()
	require.NoError(t, err)

	// Verify log file was created and contains expected content
	assert.FileExists(t, logFile)

	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	logContent := string(content)
	assert.Contains(t, logContent, "integration test started")
	assert.Contains(t, logContent, "integration-test") // service name
	assert.Contains(t, logContent, "test")             // environment
	assert.Contains(t, logContent, "with_fields")
	assert.Contains(t, logContent, "integration-req-123")
	assert.Contains(t, logContent, "TEST_ERROR")

	// Verify JSON format
	lines := strings.Split(strings.TrimSpace(logContent), "\n")
	for _, line := range lines {
		if line != "" {
			var logEntry map[string]interface{}
			err := json.Unmarshal([]byte(line), &logEntry)
			assert.NoError(t, err, "Each log line should be valid JSON")
		}
	}
}
