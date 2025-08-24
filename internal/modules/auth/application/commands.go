package application

import (
	"go-templ-template/internal/modules/auth/domain"
)

// LoginCommand represents a command to authenticate a user
type LoginCommand struct {
	Email     string `json:"email" validate:"required,email,max=255"`
	Password  string `json:"password" validate:"required"`
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// Validate performs validation on the LoginCommand
func (c *LoginCommand) Validate() error {
	if c.Email == "" {
		return NewValidationError("email", "email is required")
	}
	if c.Password == "" {
		return NewValidationError("password", "password is required")
	}
	if err := domain.ValidateEmail(c.Email); err != nil {
		return NewValidationError("email", err.Error())
	}
	return nil
}

// RegisterCommand represents a command to register a new user
type RegisterCommand struct {
	Email     string `json:"email" validate:"required,email,max=255"`
	Password  string `json:"password" validate:"required,min=8,max=128"`
	FirstName string `json:"first_name" validate:"required,max=100"`
	LastName  string `json:"last_name" validate:"required,max=100"`
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// Validate performs validation on the RegisterCommand
func (c *RegisterCommand) Validate() error {
	if c.Email == "" {
		return NewValidationError("email", "email is required")
	}
	if c.Password == "" {
		return NewValidationError("password", "password is required")
	}
	if c.FirstName == "" {
		return NewValidationError("first_name", "first name is required")
	}
	if c.LastName == "" {
		return NewValidationError("last_name", "last name is required")
	}

	// Validate email format
	if err := domain.ValidateEmail(c.Email); err != nil {
		return NewValidationError("email", err.Error())
	}

	// Validate password strength
	validator := domain.NewPasswordValidator()
	if err := validator.Validate(c.Password); err != nil {
		return NewValidationError("password", err.Error())
	}

	// Check for common passwords
	if domain.IsCommonPassword(c.Password) {
		return NewValidationError("password", "password is too common, please choose a stronger password")
	}

	return nil
}

// LogoutCommand represents a command to logout a user
type LogoutCommand struct {
	SessionID string `json:"session_id" validate:"required"`
	UserID    string `json:"user_id,omitempty"`
}

// Validate performs validation on the LogoutCommand
func (c *LogoutCommand) Validate() error {
	if c.SessionID == "" {
		return NewValidationError("session_id", "session ID is required")
	}
	return nil
}

// ValidateSessionQuery represents a query to validate a session
type ValidateSessionQuery struct {
	SessionID string `json:"session_id" validate:"required"`
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// Validate performs validation on the ValidateSessionQuery
func (q *ValidateSessionQuery) Validate() error {
	if q.SessionID == "" {
		return NewValidationError("session_id", "session ID is required")
	}
	return nil
}

// RefreshSessionCommand represents a command to refresh/extend a session
type RefreshSessionCommand struct {
	SessionID string `json:"session_id" validate:"required"`
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// Validate performs validation on the RefreshSessionCommand
func (c *RefreshSessionCommand) Validate() error {
	if c.SessionID == "" {
		return NewValidationError("session_id", "session ID is required")
	}
	return nil
}

// ChangePasswordCommand represents a command to change a user's password
type ChangePasswordCommand struct {
	UserID      string `json:"user_id" validate:"required"`
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=128"`
	IPAddress   string `json:"ip_address,omitempty"`
	UserAgent   string `json:"user_agent,omitempty"`
}

// Validate performs validation on the ChangePasswordCommand
func (c *ChangePasswordCommand) Validate() error {
	if c.UserID == "" {
		return NewValidationError("user_id", "user ID is required")
	}
	if c.OldPassword == "" {
		return NewValidationError("old_password", "old password is required")
	}
	if c.NewPassword == "" {
		return NewValidationError("new_password", "new password is required")
	}

	// Validate new password strength
	validator := domain.NewPasswordValidator()
	if err := validator.Validate(c.NewPassword); err != nil {
		return NewValidationError("new_password", err.Error())
	}

	// Check for common passwords
	if domain.IsCommonPassword(c.NewPassword) {
		return NewValidationError("new_password", "password is too common, please choose a stronger password")
	}

	// Ensure new password is different from old password
	if c.OldPassword == c.NewPassword {
		return NewValidationError("new_password", "new password must be different from the current password")
	}

	return nil
}
