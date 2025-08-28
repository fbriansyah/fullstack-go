package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appError *AppError
		expected string
	}{
		{
			name: "error without cause",
			appError: &AppError{
				Code:    "TEST_ERROR",
				Type:    ErrorTypeValidation,
				Message: "Test error message",
			},
			expected: "[TEST_ERROR] validation: Test error message",
		},
		{
			name: "error with cause",
			appError: &AppError{
				Code:    "TEST_ERROR",
				Type:    ErrorTypeInternal,
				Message: "Test error message",
				Cause:   assert.AnError,
			},
			expected: "[TEST_ERROR] internal: Test error message (caused by: assert.AnError general error for testing)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.appError.Error())
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	cause := assert.AnError
	appError := &AppError{
		Code:  "TEST_ERROR",
		Type:  ErrorTypeInternal,
		Cause: cause,
	}

	assert.Equal(t, cause, appError.Unwrap())
}

func TestAppError_Is(t *testing.T) {
	appError1 := &AppError{
		Code: "TEST_ERROR",
		Type: ErrorTypeValidation,
	}

	appError2 := &AppError{
		Code: "TEST_ERROR",
		Type: ErrorTypeValidation,
	}

	appError3 := &AppError{
		Code: "OTHER_ERROR",
		Type: ErrorTypeValidation,
	}

	appError4 := &AppError{
		Code: "TEST_ERROR",
		Type: ErrorTypeInternal,
	}

	// Same code should match
	assert.True(t, appError1.Is(appError2))

	// Same type should match
	assert.True(t, appError1.Is(appError3))

	// Different type should not match
	assert.False(t, appError1.Is(appError4))

	// Non-AppError should not match
	assert.False(t, appError1.Is(assert.AnError))
}

func TestAppError_WithContext(t *testing.T) {
	appError := &AppError{
		Code: "TEST_ERROR",
		Type: ErrorTypeValidation,
	}

	ctx := ErrorContext{
		RequestID: "req123",
		UserID:    "user456",
	}

	result := appError.WithContext(ctx)

	assert.Equal(t, appError, result) // Should return same instance
	assert.Equal(t, ctx, appError.Context)
}

func TestAppError_WithDetails(t *testing.T) {
	appError := &AppError{
		Code: "TEST_ERROR",
		Type: ErrorTypeValidation,
	}

	details := map[string]interface{}{
		"field": "email",
		"value": "invalid",
	}

	result := appError.WithDetails(details)

	assert.Equal(t, appError, result) // Should return same instance
	assert.Equal(t, details, appError.Details)

	// Test adding more details
	moreDetails := map[string]interface{}{
		"reason": "format",
	}

	result = appError.WithDetails(moreDetails)

	assert.Len(t, appError.Details, 3)
	assert.Equal(t, "email", appError.Details["field"])
	assert.Equal(t, "invalid", appError.Details["value"])
	assert.Equal(t, "format", appError.Details["reason"])
}

func TestAppError_WithUserMessage(t *testing.T) {
	appError := &AppError{
		Code: "TEST_ERROR",
		Type: ErrorTypeValidation,
	}

	userMessage := "Please check your input"
	result := appError.WithUserMessage(userMessage)

	assert.Equal(t, appError, result) // Should return same instance
	assert.Equal(t, userMessage, appError.UserMessage)
}

func TestAppError_ToHTTPResponse(t *testing.T) {
	appError := &AppError{
		ID:        "error123",
		Code:      "TEST_ERROR",
		Type:      ErrorTypeValidation,
		Message:   "Test error message",
		Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		Details: map[string]interface{}{
			"field":    "email",
			"password": "secret", // Should be filtered
		},
		UserMessage: "Please check your input",
		Retryable:   true,
	}

	response := appError.ToHTTPResponse()

	// Verify structure
	require.Contains(t, response, "error")
	errorData := response["error"].(map[string]interface{})

	assert.Equal(t, "error123", errorData["id"])
	assert.Equal(t, "TEST_ERROR", errorData["code"])
	assert.Equal(t, "validation", errorData["type"])
	assert.Equal(t, "Test error message", errorData["message"])
	assert.Equal(t, "Please check your input", errorData["user_message"])
	assert.Equal(t, true, errorData["retryable"])

	// Verify sensitive data is filtered
	if details, exists := errorData["details"]; exists {
		detailsMap := details.(map[string]interface{})
		assert.Contains(t, detailsMap, "field")
		assert.NotContains(t, detailsMap, "password")
	}
}

func TestAppError_JSONSerialization(t *testing.T) {
	appError := &AppError{
		ID:        "error123",
		Code:      "TEST_ERROR",
		Type:      ErrorTypeValidation,
		Message:   "Test error message",
		Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		Details: map[string]interface{}{
			"field": "email",
		},
		UserMessage: "Please check your input",
		Retryable:   true,
	}

	// Test JSON marshaling
	data, err := json.Marshal(appError)
	require.NoError(t, err)

	// Test JSON unmarshaling
	var unmarshaled AppError
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, appError.ID, unmarshaled.ID)
	assert.Equal(t, appError.Code, unmarshaled.Code)
	assert.Equal(t, appError.Type, unmarshaled.Type)
	assert.Equal(t, appError.Message, unmarshaled.Message)
	assert.Equal(t, appError.UserMessage, unmarshaled.UserMessage)
	assert.Equal(t, appError.Retryable, unmarshaled.Retryable)
}

func TestErrorList_Error(t *testing.T) {
	tests := []struct {
		name     string
		errors   []*AppError
		expected string
	}{
		{
			name:     "empty error list",
			errors:   []*AppError{},
			expected: "no errors",
		},
		{
			name: "single error",
			errors: []*AppError{
				{Code: "TEST_ERROR", Type: ErrorTypeValidation, Message: "Test error"},
			},
			expected: "[TEST_ERROR] validation: Test error",
		},
		{
			name: "multiple errors",
			errors: []*AppError{
				{Code: "ERROR1", Type: ErrorTypeValidation, Message: "Error 1"},
				{Code: "ERROR2", Type: ErrorTypeValidation, Message: "Error 2"},
			},
			expected: "multiple errors: 2 errors occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorList := &ErrorList{Errors: tt.errors}
			assert.Equal(t, tt.expected, errorList.Error())
		})
	}
}

func TestErrorList_Add(t *testing.T) {
	errorList := &ErrorList{}

	appError := &AppError{
		Code: "TEST_ERROR",
		Type: ErrorTypeValidation,
	}

	errorList.Add(appError)

	assert.Len(t, errorList.Errors, 1)
	assert.Equal(t, appError, errorList.Errors[0])
}

func TestErrorList_HasErrors(t *testing.T) {
	errorList := &ErrorList{}
	assert.False(t, errorList.HasErrors())

	errorList.Add(&AppError{Code: "TEST_ERROR", Type: ErrorTypeValidation})
	assert.True(t, errorList.HasErrors())
}

func TestErrorList_ToHTTPResponse(t *testing.T) {
	tests := []struct {
		name     string
		errors   []*AppError
		expected map[string]interface{}
	}{
		{
			name:   "empty error list",
			errors: []*AppError{},
			expected: map[string]interface{}{
				"errors": []interface{}{},
			},
		},
		{
			name: "single error",
			errors: []*AppError{
				{
					ID:      "error1",
					Code:    "TEST_ERROR",
					Type:    ErrorTypeValidation,
					Message: "Test error",
				},
			},
		},
		{
			name: "multiple errors",
			errors: []*AppError{
				{
					ID:      "error1",
					Code:    "ERROR1",
					Type:    ErrorTypeValidation,
					Message: "Error 1",
				},
				{
					ID:      "error2",
					Code:    "ERROR2",
					Type:    ErrorTypeValidation,
					Message: "Error 2",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorList := &ErrorList{Errors: tt.errors}
			response := errorList.ToHTTPResponse()

			if len(tt.errors) == 0 {
				assert.Equal(t, tt.expected, response)
			} else if len(tt.errors) == 1 {
				// Single error should return single error format
				assert.Contains(t, response, "error")
			} else {
				// Multiple errors should return errors array
				assert.Contains(t, response, "errors")
				errors := response["errors"].([]interface{})
				assert.Len(t, errors, len(tt.errors))
			}
		})
	}
}

func TestErrorList_GetHTTPStatus(t *testing.T) {
	tests := []struct {
		name           string
		errors         []*AppError
		expectedStatus int
	}{
		{
			name:           "empty error list",
			errors:         []*AppError{},
			expectedStatus: http.StatusOK,
		},
		{
			name: "single validation error",
			errors: []*AppError{
				{HTTPStatus: http.StatusBadRequest},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "multiple client errors",
			errors: []*AppError{
				{HTTPStatus: http.StatusBadRequest},
				{HTTPStatus: http.StatusUnauthorized},
			},
			expectedStatus: http.StatusUnauthorized, // Higher status
		},
		{
			name: "mixed client and server errors",
			errors: []*AppError{
				{HTTPStatus: http.StatusBadRequest},
				{HTTPStatus: http.StatusInternalServerError},
			},
			expectedStatus: http.StatusInternalServerError, // Server error takes precedence
		},
		{
			name: "all client errors returns 400",
			errors: []*AppError{
				{HTTPStatus: http.StatusBadRequest},
				{HTTPStatus: http.StatusNotFound},
			},
			expectedStatus: http.StatusNotFound, // Higher client error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorList := &ErrorList{Errors: tt.errors}
			status := errorList.GetHTTPStatus()
			assert.Equal(t, tt.expectedStatus, status)
		})
	}
}

func TestErrorContext_Serialization(t *testing.T) {
	ctx := ErrorContext{
		Operation: "user_creation",
		Component: "user_service",
		UserID:    "user123",
		RequestID: "req456",
		SessionID: "session789",
		IPAddress: "192.168.1.1",
		UserAgent: "Mozilla/5.0",
		Metadata: map[string]interface{}{
			"method": "POST",
			"path":   "/api/users",
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(ctx)
	require.NoError(t, err)

	// Test JSON unmarshaling
	var unmarshaled ErrorContext
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, ctx.Operation, unmarshaled.Operation)
	assert.Equal(t, ctx.Component, unmarshaled.Component)
	assert.Equal(t, ctx.UserID, unmarshaled.UserID)
	assert.Equal(t, ctx.RequestID, unmarshaled.RequestID)
	assert.Equal(t, ctx.SessionID, unmarshaled.SessionID)
	assert.Equal(t, ctx.IPAddress, unmarshaled.IPAddress)
	assert.Equal(t, ctx.UserAgent, unmarshaled.UserAgent)
	assert.Equal(t, ctx.Metadata, unmarshaled.Metadata)
}

func TestIsSensitiveField(t *testing.T) {
	tests := []struct {
		field    string
		expected bool
	}{
		{"password", true},
		{"token", true},
		{"secret", true},
		{"key", true},
		{"credential", true},
		{"authorization", true},
		{"session", true},
		{"cookie", true},
		{"private", true},
		{"username", false},
		{"email", false},
		{"name", false},
		{"id", false},
		{"public", false},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			result := isSensitiveField(tt.field)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"password", "password", true},
		{"user_password", "password", true},
		{"password_hash", "password", true},
		{"my_password_field", "password", true},
		{"username", "password", false},
		{"", "password", false},
		{"password", "", true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s contains %s", tt.s, tt.substr), func(t *testing.T) {
			result := contains(tt.s, tt.substr)
			assert.Equal(t, tt.expected, result)
		})
	}
}
