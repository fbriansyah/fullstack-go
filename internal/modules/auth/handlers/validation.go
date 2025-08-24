package handlers

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// Error implements the error interface
func (v ValidationErrors) Error() string {
	var messages []string
	for _, err := range v.Errors {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, "; ")
}

// ValidateLoginRequest validates the login request
func ValidateLoginRequest(req *LoginRequest) error {
	var errors []ValidationError

	// Validate email
	if req.Email == "" {
		errors = append(errors, ValidationError{Field: "email", Message: "email is required"})
	} else {
		if _, err := mail.ParseAddress(req.Email); err != nil {
			errors = append(errors, ValidationError{Field: "email", Message: "invalid email format"})
		}
		if len(req.Email) > 255 {
			errors = append(errors, ValidationError{Field: "email", Message: "email cannot exceed 255 characters"})
		}
	}

	// Validate password
	if req.Password == "" {
		errors = append(errors, ValidationError{Field: "password", Message: "password is required"})
	}

	if len(errors) > 0 {
		return ValidationErrors{Errors: errors}
	}

	return nil
}

// ValidateRegisterRequest validates the register request
func ValidateRegisterRequest(req *RegisterRequest) error {
	var errors []ValidationError

	// Validate email
	if req.Email == "" {
		errors = append(errors, ValidationError{Field: "email", Message: "email is required"})
	} else {
		if _, err := mail.ParseAddress(req.Email); err != nil {
			errors = append(errors, ValidationError{Field: "email", Message: "invalid email format"})
		}
		if len(req.Email) > 255 {
			errors = append(errors, ValidationError{Field: "email", Message: "email cannot exceed 255 characters"})
		}
	}

	// Validate password
	if req.Password == "" {
		errors = append(errors, ValidationError{Field: "password", Message: "password is required"})
	} else {
		if len(req.Password) < 8 {
			errors = append(errors, ValidationError{Field: "password", Message: "password must be at least 8 characters long"})
		}
		if len(req.Password) > 128 {
			errors = append(errors, ValidationError{Field: "password", Message: "password cannot exceed 128 characters"})
		}
		if !isValidPassword(req.Password) {
			errors = append(errors, ValidationError{Field: "password", Message: "password must contain at least one uppercase letter, one lowercase letter, and one digit"})
		}
	}

	// Validate first name
	if req.FirstName == "" {
		errors = append(errors, ValidationError{Field: "first_name", Message: "first name is required"})
	} else {
		if len(req.FirstName) > 100 {
			errors = append(errors, ValidationError{Field: "first_name", Message: "first name cannot exceed 100 characters"})
		}
		if !isValidName(req.FirstName) {
			errors = append(errors, ValidationError{Field: "first_name", Message: "first name contains invalid characters"})
		}
	}

	// Validate last name
	if req.LastName == "" {
		errors = append(errors, ValidationError{Field: "last_name", Message: "last name is required"})
	} else {
		if len(req.LastName) > 100 {
			errors = append(errors, ValidationError{Field: "last_name", Message: "last name cannot exceed 100 characters"})
		}
		if !isValidName(req.LastName) {
			errors = append(errors, ValidationError{Field: "last_name", Message: "last name contains invalid characters"})
		}
	}

	if len(errors) > 0 {
		return ValidationErrors{Errors: errors}
	}

	return nil
}

// ValidateChangePasswordRequest validates the change password request
func ValidateChangePasswordRequest(req *ChangePasswordRequest) error {
	var errors []ValidationError

	// Validate old password
	if req.OldPassword == "" {
		errors = append(errors, ValidationError{Field: "old_password", Message: "old password is required"})
	}

	// Validate new password
	if req.NewPassword == "" {
		errors = append(errors, ValidationError{Field: "new_password", Message: "new password is required"})
	} else {
		if len(req.NewPassword) < 8 {
			errors = append(errors, ValidationError{Field: "new_password", Message: "new password must be at least 8 characters long"})
		}
		if len(req.NewPassword) > 128 {
			errors = append(errors, ValidationError{Field: "new_password", Message: "new password cannot exceed 128 characters"})
		}
		if !isValidPassword(req.NewPassword) {
			errors = append(errors, ValidationError{Field: "new_password", Message: "new password must contain at least one uppercase letter, one lowercase letter, and one digit"})
		}
	}

	// Ensure new password is different from old password
	if req.OldPassword != "" && req.NewPassword != "" && req.OldPassword == req.NewPassword {
		errors = append(errors, ValidationError{Field: "new_password", Message: "new password must be different from the current password"})
	}

	if len(errors) > 0 {
		return ValidationErrors{Errors: errors}
	}

	return nil
}

// BindAndValidate binds the request and validates it
func BindAndValidate(c echo.Context, req interface{}) error {
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(400, "Invalid request format")
	}

	switch v := req.(type) {
	case *LoginRequest:
		return ValidateLoginRequest(v)
	case *RegisterRequest:
		return ValidateRegisterRequest(v)
	case *ChangePasswordRequest:
		return ValidateChangePasswordRequest(v)
	}

	return nil
}

// isValidPassword checks if password meets complexity requirements
func isValidPassword(password string) bool {
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	return hasUpper && hasLower && hasDigit
}

// isValidName checks if name contains only valid characters
func isValidName(name string) bool {
	nameRegex := regexp.MustCompile(`^[a-zA-Z\s\-']+$`)
	return nameRegex.MatchString(name)
}
