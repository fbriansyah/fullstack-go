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

// ValidateCreateUserRequest validates the create user request
func ValidateCreateUserRequest(req *CreateUserRequest) error {
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

// ValidateUpdateUserRequest validates the update user request
func ValidateUpdateUserRequest(req *UpdateUserRequest) error {
	var errors []ValidationError

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

	// Validate version
	if req.Version < 1 {
		errors = append(errors, ValidationError{Field: "version", Message: "version must be greater than 0"})
	}

	if len(errors) > 0 {
		return ValidationErrors{Errors: errors}
	}

	return nil
}

// ValidateUpdateUserEmailRequest validates the update user email request
func ValidateUpdateUserEmailRequest(req *UpdateUserEmailRequest) error {
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

	// Validate version
	if req.Version < 1 {
		errors = append(errors, ValidationError{Field: "version", Message: "version must be greater than 0"})
	}

	if len(errors) > 0 {
		return ValidationErrors{Errors: errors}
	}

	return nil
}

// ValidateChangeUserPasswordRequest validates the change user password request
func ValidateChangeUserPasswordRequest(req *ChangeUserPasswordRequest) error {
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

	// Validate version
	if req.Version < 1 {
		errors = append(errors, ValidationError{Field: "version", Message: "version must be greater than 0"})
	}

	if len(errors) > 0 {
		return ValidationErrors{Errors: errors}
	}

	return nil
}

// ValidateChangeUserStatusRequest validates the change user status request
func ValidateChangeUserStatusRequest(req *ChangeUserStatusRequest) error {
	var errors []ValidationError

	// Validate status
	if !req.Status.IsValid() {
		errors = append(errors, ValidationError{Field: "status", Message: "invalid user status"})
	}

	// Validate version
	if req.Version < 1 {
		errors = append(errors, ValidationError{Field: "version", Message: "version must be greater than 0"})
	}

	if len(errors) > 0 {
		return ValidationErrors{Errors: errors}
	}

	return nil
}

// ValidateListUsersRequest validates the list users request
func ValidateListUsersRequest(req *ListUsersRequest) error {
	var errors []ValidationError

	// Validate limit
	if req.Limit <= 0 {
		req.Limit = 20 // Default limit
	}
	if req.Limit > 100 {
		errors = append(errors, ValidationError{Field: "limit", Message: "limit cannot exceed 100"})
	}

	// Validate offset
	if req.Offset < 0 {
		errors = append(errors, ValidationError{Field: "offset", Message: "offset cannot be negative"})
	}

	// Validate status if provided
	if req.Status != nil && !req.Status.IsValid() {
		errors = append(errors, ValidationError{Field: "status", Message: "invalid user status"})
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
	case *CreateUserRequest:
		return ValidateCreateUserRequest(v)
	case *UpdateUserRequest:
		return ValidateUpdateUserRequest(v)
	case *UpdateUserEmailRequest:
		return ValidateUpdateUserEmailRequest(v)
	case *ChangeUserPasswordRequest:
		return ValidateChangeUserPasswordRequest(v)
	case *ChangeUserStatusRequest:
		return ValidateChangeUserStatusRequest(v)
	case *ListUsersRequest:
		return ValidateListUsersRequest(v)
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
