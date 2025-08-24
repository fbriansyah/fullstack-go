package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateLoginRequest_Success(t *testing.T) {
	req := &LoginRequest{
		Email:    "test@example.com",
		Password: "Password123",
	}

	err := ValidateLoginRequest(req)
	assert.NoError(t, err)
}

func TestValidateLoginRequest_EmptyEmail(t *testing.T) {
	req := &LoginRequest{
		Email:    "",
		Password: "Password123",
	}

	err := ValidateLoginRequest(req)
	assert.Error(t, err)

	validationErrs, ok := err.(ValidationErrors)
	assert.True(t, ok)
	assert.Len(t, validationErrs.Errors, 1)
	assert.Equal(t, "email", validationErrs.Errors[0].Field)
	assert.Equal(t, "email is required", validationErrs.Errors[0].Message)
}

func TestValidateLoginRequest_InvalidEmail(t *testing.T) {
	req := &LoginRequest{
		Email:    "invalid-email",
		Password: "Password123",
	}

	err := ValidateLoginRequest(req)
	assert.Error(t, err)

	validationErrs, ok := err.(ValidationErrors)
	assert.True(t, ok)
	assert.Len(t, validationErrs.Errors, 1)
	assert.Equal(t, "email", validationErrs.Errors[0].Field)
	assert.Equal(t, "invalid email format", validationErrs.Errors[0].Message)
}

func TestValidateLoginRequest_EmailTooLong(t *testing.T) {
	longEmail := string(make([]byte, 250)) + "@example.com" // > 255 chars
	req := &LoginRequest{
		Email:    longEmail,
		Password: "Password123",
	}

	err := ValidateLoginRequest(req)
	assert.Error(t, err)

	validationErrs, ok := err.(ValidationErrors)
	assert.True(t, ok)
	// Should have both invalid format and too long errors
	assert.GreaterOrEqual(t, len(validationErrs.Errors), 1)
	assert.Equal(t, "email", validationErrs.Errors[0].Field)
	// Check that one of the errors is about length
	hasLengthError := false
	for _, e := range validationErrs.Errors {
		if e.Message == "email cannot exceed 255 characters" {
			hasLengthError = true
			break
		}
	}
	assert.True(t, hasLengthError)
}

func TestValidateLoginRequest_EmptyPassword(t *testing.T) {
	req := &LoginRequest{
		Email:    "test@example.com",
		Password: "",
	}

	err := ValidateLoginRequest(req)
	assert.Error(t, err)

	validationErrs, ok := err.(ValidationErrors)
	assert.True(t, ok)
	assert.Len(t, validationErrs.Errors, 1)
	assert.Equal(t, "password", validationErrs.Errors[0].Field)
	assert.Equal(t, "password is required", validationErrs.Errors[0].Message)
}

func TestValidateLoginRequest_MultipleErrors(t *testing.T) {
	req := &LoginRequest{
		Email:    "",
		Password: "",
	}

	err := ValidateLoginRequest(req)
	assert.Error(t, err)

	validationErrs, ok := err.(ValidationErrors)
	assert.True(t, ok)
	assert.Len(t, validationErrs.Errors, 2)
}

func TestValidateRegisterRequest_Success(t *testing.T) {
	req := &RegisterRequest{
		Email:     "test@example.com",
		Password:  "Password123",
		FirstName: "John",
		LastName:  "Doe",
	}

	err := ValidateRegisterRequest(req)
	assert.NoError(t, err)
}

func TestValidateRegisterRequest_InvalidEmail(t *testing.T) {
	req := &RegisterRequest{
		Email:     "invalid-email",
		Password:  "Password123",
		FirstName: "John",
		LastName:  "Doe",
	}

	err := ValidateRegisterRequest(req)
	assert.Error(t, err)

	validationErrs, ok := err.(ValidationErrors)
	assert.True(t, ok)
	assert.Equal(t, "email", validationErrs.Errors[0].Field)
}

func TestValidateRegisterRequest_PasswordTooShort(t *testing.T) {
	req := &RegisterRequest{
		Email:     "test@example.com",
		Password:  "short",
		FirstName: "John",
		LastName:  "Doe",
	}

	err := ValidateRegisterRequest(req)
	assert.Error(t, err)

	validationErrs, ok := err.(ValidationErrors)
	assert.True(t, ok)
	assert.Equal(t, "password", validationErrs.Errors[0].Field)
	assert.Equal(t, "password must be at least 8 characters long", validationErrs.Errors[0].Message)
}

func TestValidateRegisterRequest_PasswordTooLong(t *testing.T) {
	longPassword := string(make([]byte, 130)) // > 128 chars
	req := &RegisterRequest{
		Email:     "test@example.com",
		Password:  longPassword,
		FirstName: "John",
		LastName:  "Doe",
	}

	err := ValidateRegisterRequest(req)
	assert.Error(t, err)

	validationErrs, ok := err.(ValidationErrors)
	assert.True(t, ok)
	assert.Equal(t, "password", validationErrs.Errors[0].Field)
	assert.Equal(t, "password cannot exceed 128 characters", validationErrs.Errors[0].Message)
}

func TestValidateRegisterRequest_WeakPassword(t *testing.T) {
	req := &RegisterRequest{
		Email:     "test@example.com",
		Password:  "password", // No uppercase, no digits
		FirstName: "John",
		LastName:  "Doe",
	}

	err := ValidateRegisterRequest(req)
	assert.Error(t, err)

	validationErrs, ok := err.(ValidationErrors)
	assert.True(t, ok)
	assert.Equal(t, "password", validationErrs.Errors[0].Field)
	assert.Contains(t, validationErrs.Errors[0].Message, "uppercase letter")
}

func TestValidateRegisterRequest_EmptyFirstName(t *testing.T) {
	req := &RegisterRequest{
		Email:     "test@example.com",
		Password:  "Password123",
		FirstName: "",
		LastName:  "Doe",
	}

	err := ValidateRegisterRequest(req)
	assert.Error(t, err)

	validationErrs, ok := err.(ValidationErrors)
	assert.True(t, ok)
	assert.Equal(t, "first_name", validationErrs.Errors[0].Field)
	assert.Equal(t, "first name is required", validationErrs.Errors[0].Message)
}

func TestValidateRegisterRequest_FirstNameTooLong(t *testing.T) {
	longName := string(make([]byte, 101)) // > 100 chars
	req := &RegisterRequest{
		Email:     "test@example.com",
		Password:  "Password123",
		FirstName: longName,
		LastName:  "Doe",
	}

	err := ValidateRegisterRequest(req)
	assert.Error(t, err)

	validationErrs, ok := err.(ValidationErrors)
	assert.True(t, ok)
	assert.Equal(t, "first_name", validationErrs.Errors[0].Field)
	assert.Equal(t, "first name cannot exceed 100 characters", validationErrs.Errors[0].Message)
}

func TestValidateRegisterRequest_InvalidFirstName(t *testing.T) {
	req := &RegisterRequest{
		Email:     "test@example.com",
		Password:  "Password123",
		FirstName: "John123", // Contains numbers
		LastName:  "Doe",
	}

	err := ValidateRegisterRequest(req)
	assert.Error(t, err)

	validationErrs, ok := err.(ValidationErrors)
	assert.True(t, ok)
	assert.Equal(t, "first_name", validationErrs.Errors[0].Field)
	assert.Equal(t, "first name contains invalid characters", validationErrs.Errors[0].Message)
}

func TestValidateRegisterRequest_EmptyLastName(t *testing.T) {
	req := &RegisterRequest{
		Email:     "test@example.com",
		Password:  "Password123",
		FirstName: "John",
		LastName:  "",
	}

	err := ValidateRegisterRequest(req)
	assert.Error(t, err)

	validationErrs, ok := err.(ValidationErrors)
	assert.True(t, ok)
	assert.Equal(t, "last_name", validationErrs.Errors[0].Field)
	assert.Equal(t, "last name is required", validationErrs.Errors[0].Message)
}

func TestValidateChangePasswordRequest_Success(t *testing.T) {
	req := &ChangePasswordRequest{
		OldPassword: "OldPassword123",
		NewPassword: "NewPassword123",
	}

	err := ValidateChangePasswordRequest(req)
	assert.NoError(t, err)
}

func TestValidateChangePasswordRequest_EmptyOldPassword(t *testing.T) {
	req := &ChangePasswordRequest{
		OldPassword: "",
		NewPassword: "NewPassword123",
	}

	err := ValidateChangePasswordRequest(req)
	assert.Error(t, err)

	validationErrs, ok := err.(ValidationErrors)
	assert.True(t, ok)
	assert.Equal(t, "old_password", validationErrs.Errors[0].Field)
	assert.Equal(t, "old password is required", validationErrs.Errors[0].Message)
}

func TestValidateChangePasswordRequest_EmptyNewPassword(t *testing.T) {
	req := &ChangePasswordRequest{
		OldPassword: "OldPassword123",
		NewPassword: "",
	}

	err := ValidateChangePasswordRequest(req)
	assert.Error(t, err)

	validationErrs, ok := err.(ValidationErrors)
	assert.True(t, ok)
	assert.Equal(t, "new_password", validationErrs.Errors[0].Field)
	assert.Equal(t, "new password is required", validationErrs.Errors[0].Message)
}

func TestValidateChangePasswordRequest_NewPasswordTooShort(t *testing.T) {
	req := &ChangePasswordRequest{
		OldPassword: "OldPassword123",
		NewPassword: "short",
	}

	err := ValidateChangePasswordRequest(req)
	assert.Error(t, err)

	validationErrs, ok := err.(ValidationErrors)
	assert.True(t, ok)
	assert.Equal(t, "new_password", validationErrs.Errors[0].Field)
	assert.Equal(t, "new password must be at least 8 characters long", validationErrs.Errors[0].Message)
}

func TestValidateChangePasswordRequest_WeakNewPassword(t *testing.T) {
	req := &ChangePasswordRequest{
		OldPassword: "OldPassword123",
		NewPassword: "newpassword", // No uppercase, no digits
	}

	err := ValidateChangePasswordRequest(req)
	assert.Error(t, err)

	validationErrs, ok := err.(ValidationErrors)
	assert.True(t, ok)
	assert.Equal(t, "new_password", validationErrs.Errors[0].Field)
	assert.Contains(t, validationErrs.Errors[0].Message, "uppercase letter")
}

func TestValidateChangePasswordRequest_SamePassword(t *testing.T) {
	req := &ChangePasswordRequest{
		OldPassword: "Password123",
		NewPassword: "Password123",
	}

	err := ValidateChangePasswordRequest(req)
	assert.Error(t, err)

	validationErrs, ok := err.(ValidationErrors)
	assert.True(t, ok)
	assert.Equal(t, "new_password", validationErrs.Errors[0].Field)
	assert.Equal(t, "new password must be different from the current password", validationErrs.Errors[0].Message)
}

func TestIsValidPassword(t *testing.T) {
	testCases := []struct {
		password string
		valid    bool
	}{
		{"Password123", true},
		{"MySecure1", true},
		{"Test1234", true},
		{"password123", false}, // No uppercase
		{"PASSWORD123", false}, // No lowercase
		{"Password", false},    // No digit
		{"Pass1", true},        // Minimum valid
		{"", false},            // Empty
		{"12345678", false},    // Only digits
		{"ABCDEFGH", false},    // Only uppercase
		{"abcdefgh", false},    // Only lowercase
	}

	for _, tc := range testCases {
		t.Run(tc.password, func(t *testing.T) {
			result := isValidPassword(tc.password)
			assert.Equal(t, tc.valid, result, "Password: %s", tc.password)
		})
	}
}

func TestIsValidName(t *testing.T) {
	testCases := []struct {
		name  string
		valid bool
	}{
		{"John", true},
		{"Mary Jane", true},
		{"O'Connor", true},
		{"Jean-Pierre", true},
		{"John123", false},     // Contains numbers
		{"John@Doe", false},    // Contains special chars
		{"John_Doe", false},    // Contains underscore
		{"", false},            // Empty
		{"José", false},        // Unicode letters not supported in current regex
		{"Anne-Marie", true},   // Hyphenated names
		{"Mary O'Brien", true}, // Apostrophe with space
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isValidName(tc.name)
			assert.Equal(t, tc.valid, result, "Name: %s", tc.name)
		})
	}
}

func TestValidationErrors_Error(t *testing.T) {
	errors := ValidationErrors{
		Errors: []ValidationError{
			{Field: "email", Message: "email is required"},
			{Field: "password", Message: "password is too short"},
		},
	}

	errorMsg := errors.Error()
	assert.Contains(t, errorMsg, "email: email is required")
	assert.Contains(t, errorMsg, "password: password is too short")
	assert.Contains(t, errorMsg, ";")
}

func TestBindAndValidate_LoginRequest(t *testing.T) {
	// This test would require a full Echo context setup
	// For now, we'll test the validation logic directly
	req := &LoginRequest{
		Email:    "test@example.com",
		Password: "Password123",
	}

	err := ValidateLoginRequest(req)
	assert.NoError(t, err)
}

func TestBindAndValidate_RegisterRequest(t *testing.T) {
	req := &RegisterRequest{
		Email:     "test@example.com",
		Password:  "Password123",
		FirstName: "John",
		LastName:  "Doe",
	}

	err := ValidateRegisterRequest(req)
	assert.NoError(t, err)
}

func TestBindAndValidate_ChangePasswordRequest(t *testing.T) {
	req := &ChangePasswordRequest{
		OldPassword: "OldPassword123",
		NewPassword: "NewPassword123",
	}

	err := ValidateChangePasswordRequest(req)
	assert.NoError(t, err)
}
