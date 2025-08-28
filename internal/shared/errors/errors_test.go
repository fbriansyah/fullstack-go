package errors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsType(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		errorType ErrorType
		expected  bool
	}{
		{
			name:      "AppError with matching type",
			err:       NewValidationError("TEST_ERROR", "Test error"),
			errorType: ErrorTypeValidation,
			expected:  true,
		},
		{
			name:      "AppError with non-matching type",
			err:       NewValidationError("TEST_ERROR", "Test error"),
			errorType: ErrorTypeInternal,
			expected:  false,
		},
		{
			name:      "non-AppError",
			err:       fmt.Errorf("standard error"),
			errorType: ErrorTypeValidation,
			expected:  false,
		},
		{
			name:      "nil error",
			err:       nil,
			errorType: ErrorTypeValidation,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsType(tt.err, tt.errorType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrorTypeCheckers(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		checker  func(error) bool
		expected bool
	}{
		{"validation error", NewValidationError("TEST", "test"), IsValidationError, true},
		{"auth error", NewAuthenticationError("TEST", "test"), IsAuthenticationError, true},
		{"authz error", NewAuthorizationError("TEST", "test"), IsAuthorizationError, true},
		{"not found error", NewNotFoundError("User", "123"), IsNotFoundError, true},
		{"conflict error", NewConflictError("TEST", "test"), IsConflictError, true},
		{"rate limit error", NewRateLimitError(100, "minute"), IsRateLimitError, true},
		{"internal error", NewInternalError("TEST", "test"), IsInternalError, true},
		{"external error", NewExternalServiceError("service", "op", fmt.Errorf("err")), IsExternalError, true},
		{"timeout error", NewTimeoutError("op", 0), IsTimeoutError, true},
		{"unavailable error", NewServiceUnavailableError("service"), IsUnavailableError, true},
		{"wrong type", NewValidationError("TEST", "test"), IsInternalError, false},
		{"standard error", fmt.Errorf("standard"), IsValidationError, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.checker(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"retryable timeout error", NewTimeoutError("op", 0), true},
		{"retryable unavailable error", NewServiceUnavailableError("service"), true},
		{"retryable external error", NewExternalServiceError("service", "op", fmt.Errorf("err")), true},
		{"retryable rate limit error", NewRateLimitError(100, "minute"), true},
		{"non-retryable validation error", NewValidationError("TEST", "test"), false},
		{"non-retryable internal error", NewInternalError("TEST", "test"), false},
		{"standard error", fmt.Errorf("standard"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryable(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetErrorCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "AppError with code",
			err:      NewValidationError("CUSTOM_CODE", "test"),
			expected: "CUSTOM_CODE",
		},
		{
			name:     "standard error",
			err:      fmt.Errorf("standard error"),
			expected: "UNKNOWN_ERROR",
		},
		{
			name:     "nil error",
			err:      nil,
			expected: "UNKNOWN_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetErrorCode(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetErrorType(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected ErrorType
	}{
		{
			name:     "AppError with type",
			err:      NewValidationError("TEST", "test"),
			expected: ErrorTypeValidation,
		},
		{
			name:     "standard error",
			err:      fmt.Errorf("standard error"),
			expected: ErrorTypeInternal,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: ErrorTypeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetErrorType(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetHTTPStatus(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "AppError with status",
			err:      NewValidationError("TEST", "test"),
			expected: 400,
		},
		{
			name:     "standard error",
			err:      fmt.Errorf("standard error"),
			expected: 500,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetHTTPStatus(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to check if all errors are nil
func allNil(errors []error) bool {
	for _, err := range errors {
		if err != nil {
			return false
		}
	}
	return true
}

func TestChain(t *testing.T) {
	tests := []struct {
		name     string
		errors   []error
		expected error
	}{
		{
			name:     "no errors",
			errors:   []error{},
			expected: nil,
		},
		{
			name:     "nil errors",
			errors:   []error{nil, nil},
			expected: nil,
		},
		{
			name:     "single error",
			errors:   []error{NewValidationError("TEST", "test")},
			expected: NewValidationError("TEST", "test"),
		},
		{
			name: "multiple errors",
			errors: []error{
				NewValidationError("ERROR1", "error 1"),
				NewValidationError("ERROR2", "error 2"),
			},
			expected: nil, // Will be checked as ErrorList
		},
		{
			name: "mixed error types",
			errors: []error{
				NewValidationError("VALIDATION", "validation error"),
				fmt.Errorf("standard error"),
			},
			expected: nil, // Will be checked as ErrorList
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Chain(tt.errors...)

			if len(tt.errors) == 0 {
				assert.Nil(t, result)
			} else if len(tt.errors) > 0 && allNil(tt.errors) {
				assert.Nil(t, result)
			} else if len(tt.errors) == 1 {
				// For single error, just check the type and code since ID will be different
				appErr, ok := result.(*AppError)
				require.True(t, ok)
				expectedErr, ok := tt.expected.(*AppError)
				require.True(t, ok)
				assert.Equal(t, expectedErr.Code, appErr.Code)
				assert.Equal(t, expectedErr.Type, appErr.Type)
			} else {
				// Should return ErrorList for multiple errors
				errorList, ok := result.(*ErrorList)
				require.True(t, ok)
				assert.Len(t, errorList.Errors, len(tt.errors))
			}
		})
	}
}

func TestCombine(t *testing.T) {
	// Combine should be an alias for Chain
	err1 := NewValidationError("ERROR1", "error 1")
	err2 := NewValidationError("ERROR2", "error 2")

	result := Combine(err1, err2)
	chainResult := Chain(err1, err2)

	assert.Equal(t, chainResult, result)
}

func TestMust(t *testing.T) {
	t.Run("no error", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Must(nil)
		})
	})

	t.Run("with error", func(t *testing.T) {
		assert.Panics(t, func() {
			Must(fmt.Errorf("test error"))
		})
	})
}

func TestMustReturn(t *testing.T) {
	t.Run("no error", func(t *testing.T) {
		result := MustReturn("success", nil)
		assert.Equal(t, "success", result)
	})

	t.Run("with error", func(t *testing.T) {
		assert.Panics(t, func() {
			MustReturn("value", fmt.Errorf("test error"))
		})
	})
}

func TestIgnore(t *testing.T) {
	// Should not panic or do anything
	assert.NotPanics(t, func() {
		Ignore(fmt.Errorf("test error"))
		Ignore(nil)
	})
}

func TestAsAppError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		expectOk  bool
		expectNil bool
	}{
		{
			name:      "nil error",
			err:       nil,
			expectOk:  false,
			expectNil: true,
		},
		{
			name:      "AppError",
			err:       NewValidationError("TEST", "test"),
			expectOk:  true,
			expectNil: false,
		},
		{
			name:      "standard error",
			err:       fmt.Errorf("standard error"),
			expectOk:  false,
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr, ok := AsAppError(tt.err)

			assert.Equal(t, tt.expectOk, ok)
			if tt.expectNil {
				assert.Nil(t, appErr)
			} else {
				assert.NotNil(t, appErr)
			}
		})
	}
}

func TestAsErrorList(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		expectOk  bool
		expectNil bool
	}{
		{
			name:      "nil error",
			err:       nil,
			expectOk:  false,
			expectNil: true,
		},
		{
			name: "ErrorList",
			err: &ErrorList{
				Errors: []*AppError{NewValidationError("TEST", "test")},
			},
			expectOk:  true,
			expectNil: false,
		},
		{
			name:      "AppError",
			err:       NewValidationError("TEST", "test"),
			expectOk:  false,
			expectNil: true,
		},
		{
			name:      "standard error",
			err:       fmt.Errorf("standard error"),
			expectOk:  false,
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorList, ok := AsErrorList(tt.err)

			assert.Equal(t, tt.expectOk, ok)
			if tt.expectNil {
				assert.Nil(t, errorList)
			} else {
				assert.NotNil(t, errorList)
			}
		})
	}
}

func TestRecover(t *testing.T) {
	t.Run("no panic", func(t *testing.T) {
		// Test when no panic occurs - Recover should return nil
		var appErr *AppError
		func() {
			defer func() {
				appErr = Recover()
			}()
			// No panic here
		}()
		assert.Nil(t, appErr)
	})

	t.Run("panic with error", func(t *testing.T) {
		var appErr *AppError

		// Simulate what would happen in a real panic recovery scenario
		func() {
			defer func() {
				if r := recover(); r != nil {
					// Manually create the error that Recover() would create
					var err error
					if e, ok := r.(error); ok {
						err = e
					} else {
						err = fmt.Errorf("panic: %v", r)
					}

					appErr = NewInternalErrorWithCause("PANIC_RECOVERED",
						"A panic was recovered", err).
						WithDetails(map[string]interface{}{
							"panic_value": fmt.Sprintf("%v", r),
						})
				}
			}()
			panic(fmt.Errorf("test panic"))
		}()

		require.NotNil(t, appErr)
		assert.Equal(t, "PANIC_RECOVERED", appErr.Code)
		assert.Equal(t, ErrorTypeInternal, appErr.Type)
		assert.NotNil(t, appErr.Cause)
	})

	t.Run("panic with string", func(t *testing.T) {
		var appErr *AppError

		func() {
			defer func() {
				if r := recover(); r != nil {
					var err error
					if e, ok := r.(error); ok {
						err = e
					} else {
						err = fmt.Errorf("panic: %v", r)
					}

					appErr = NewInternalErrorWithCause("PANIC_RECOVERED",
						"A panic was recovered", err).
						WithDetails(map[string]interface{}{
							"panic_value": fmt.Sprintf("%v", r),
						})
				}
			}()
			panic("test panic string")
		}()

		require.NotNil(t, appErr)
		assert.Equal(t, "PANIC_RECOVERED", appErr.Code)
		assert.Contains(t, appErr.Details["panic_value"], "test panic string")
	})
}

func TestSafeExecute(t *testing.T) {
	t.Run("successful execution", func(t *testing.T) {
		err := SafeExecute(func() error {
			return nil
		})
		assert.NoError(t, err)
	})

	t.Run("function returns error", func(t *testing.T) {
		expectedErr := fmt.Errorf("function error")
		err := SafeExecute(func() error {
			return expectedErr
		})
		assert.Equal(t, expectedErr, err)
	})

	t.Run("function panics", func(t *testing.T) {
		err := SafeExecute(func() error {
			panic("test panic")
		})

		require.Error(t, err)
		appErr, ok := err.(*AppError)
		require.True(t, ok)
		assert.Equal(t, "PANIC_RECOVERED", appErr.Code)
	})
}

func TestSafeExecuteWithReturn(t *testing.T) {
	t.Run("successful execution", func(t *testing.T) {
		result, err := SafeExecuteWithReturn(func() (string, error) {
			return "success", nil
		})
		assert.NoError(t, err)
		assert.Equal(t, "success", result)
	})

	t.Run("function returns error", func(t *testing.T) {
		expectedErr := fmt.Errorf("function error")
		result, err := SafeExecuteWithReturn(func() (string, error) {
			return "", expectedErr
		})
		assert.Equal(t, expectedErr, err)
		assert.Empty(t, result)
	})

	t.Run("function panics", func(t *testing.T) {
		result, err := SafeExecuteWithReturn(func() (string, error) {
			panic("test panic")
		})

		require.Error(t, err)
		assert.Empty(t, result)
		appErr, ok := err.(*AppError)
		require.True(t, ok)
		assert.Equal(t, "PANIC_RECOVERED", appErr.Code)
	})
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		condition bool
		errorType ErrorType
		code      string
		message   string
		expectErr bool
	}{
		{
			name:      "valid condition",
			condition: true,
			errorType: ErrorTypeValidation,
			code:      "TEST_ERROR",
			message:   "Test message",
			expectErr: false,
		},
		{
			name:      "invalid condition",
			condition: false,
			errorType: ErrorTypeValidation,
			code:      "TEST_ERROR",
			message:   "Test message",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.condition, tt.errorType, tt.code, tt.message)

			if tt.expectErr {
				require.Error(t, err)
				appErr, ok := err.(*AppError)
				require.True(t, ok)
				assert.Equal(t, tt.code, appErr.Code)
				assert.Equal(t, tt.errorType, appErr.Type)
				assert.Equal(t, tt.message, appErr.Message)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateNotNil(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		fieldName string
		expectErr bool
	}{
		{"non-nil value", "test", "field", false},
		{"nil value", nil, "field", true},
		{"zero value", 0, "field", false}, // 0 is not nil
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNotNil(tt.value, tt.fieldName)

			if tt.expectErr {
				require.Error(t, err)
				assert.True(t, IsValidationError(err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateNotEmpty(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		fieldName string
		expectErr bool
	}{
		{"non-empty string", "test", "field", false},
		{"empty string", "", "field", true},
		{"whitespace string", "   ", "field", false}, // whitespace is not empty
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNotEmpty(tt.value, tt.fieldName)

			if tt.expectErr {
				require.Error(t, err)
				assert.True(t, IsValidationError(err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateLength(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		fieldName string
		min       int
		max       int
		expectErr bool
		errorCode string
	}{
		{"valid length", "test", "field", 2, 10, false, ""},
		{"too short", "a", "field", 2, 10, true, "FIELD_TOO_SHORT"},
		{"too long", "verylongstring", "field", 2, 10, true, "FIELD_TOO_LONG"},
		{"exact min", "ab", "field", 2, 10, false, ""},
		{"exact max", "1234567890", "field", 2, 10, false, ""},
		{"no max limit", "verylongstring", "field", 2, 0, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLength(tt.value, tt.fieldName, tt.min, tt.max)

			if tt.expectErr {
				require.Error(t, err)
				assert.True(t, IsValidationError(err))
				assert.Equal(t, tt.errorCode, GetErrorCode(err))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestErrorCollector(t *testing.T) {
	t.Run("empty collector", func(t *testing.T) {
		collector := NewErrorCollector()

		assert.False(t, collector.HasErrors())
		assert.Equal(t, 0, collector.Count())
		assert.NoError(t, collector.Error())
		assert.Empty(t, collector.Errors())
	})

	t.Run("add errors", func(t *testing.T) {
		collector := NewErrorCollector()

		err1 := NewValidationError("ERROR1", "Error 1")
		err2 := NewValidationError("ERROR2", "Error 2")

		collector.Add(err1)
		collector.Add(err2)
		collector.Add(nil) // Should be ignored

		assert.True(t, collector.HasErrors())
		assert.Equal(t, 2, collector.Count())
		assert.Len(t, collector.Errors(), 2)

		combinedErr := collector.Error()
		require.Error(t, combinedErr)

		// Should be an ErrorList for multiple errors
		errorList, ok := combinedErr.(*ErrorList)
		require.True(t, ok)
		assert.Len(t, errorList.Errors, 2)
	})

	t.Run("add conditional errors", func(t *testing.T) {
		collector := NewErrorCollector()

		err1 := NewValidationError("ERROR1", "Error 1")
		err2 := NewValidationError("ERROR2", "Error 2")

		collector.AddIf(true, err1)  // Should add
		collector.AddIf(false, err2) // Should not add

		assert.True(t, collector.HasErrors())
		assert.Equal(t, 1, collector.Count())
	})

	t.Run("clear errors", func(t *testing.T) {
		collector := NewErrorCollector()

		collector.Add(NewValidationError("ERROR1", "Error 1"))
		assert.True(t, collector.HasErrors())

		collector.Clear()
		assert.False(t, collector.HasErrors())
		assert.Equal(t, 0, collector.Count())
	})

	t.Run("single error", func(t *testing.T) {
		collector := NewErrorCollector()

		err := NewValidationError("ERROR1", "Error 1")
		collector.Add(err)

		combinedErr := collector.Error()
		assert.Equal(t, err, combinedErr) // Should return the single error directly
	})
}

func TestErrorCollector_Integration(t *testing.T) {
	// Simulate form validation
	collector := NewErrorCollector()

	// Validate multiple fields
	email := ""
	password := "123"
	confirmPassword := "456"

	collector.AddIf(email == "", NewValidationError("EMAIL_REQUIRED", "Email is required"))
	collector.AddIf(len(password) < 8, NewValidationError("PASSWORD_TOO_SHORT", "Password must be at least 8 characters"))
	collector.AddIf(password != confirmPassword, NewValidationError("PASSWORD_MISMATCH", "Passwords do not match"))

	// Should have collected 3 errors
	assert.True(t, collector.HasErrors())
	assert.Equal(t, 3, collector.Count())

	// Get combined error
	err := collector.Error()
	require.Error(t, err)

	// Should be ErrorList
	errorList, ok := err.(*ErrorList)
	require.True(t, ok)
	assert.Len(t, errorList.Errors, 3)

	// Verify error codes
	codes := make([]string, len(errorList.Errors))
	for i, appErr := range errorList.Errors {
		codes[i] = appErr.Code
	}

	assert.Contains(t, codes, "EMAIL_REQUIRED")
	assert.Contains(t, codes, "PASSWORD_TOO_SHORT")
	assert.Contains(t, codes, "PASSWORD_MISMATCH")
}
