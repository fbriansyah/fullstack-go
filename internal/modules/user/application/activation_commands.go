package application

import (
	"fmt"
	"strings"
)

// RequestActivationCommand represents a command to request user activation
type RequestActivationCommand struct {
	UserID      string `json:"user_id"`
	RequestedBy string `json:"requested_by,omitempty"`
}

// Validate validates the request activation command
func (c *RequestActivationCommand) Validate() error {
	if strings.TrimSpace(c.UserID) == "" {
		return NewValidationError("user_id", "user ID is required")
	}

	return nil
}

// ActivateUserCommand represents a command to activate a user
type ActivateUserCommand struct {
	Token       string `json:"token"`
	ActivatedBy string `json:"activated_by,omitempty"`
}

// Validate validates the activate user command
func (c *ActivateUserCommand) Validate() error {
	if strings.TrimSpace(c.Token) == "" {
		return NewValidationError("token", "activation token is required")
	}

	if len(c.Token) < 10 {
		return NewValidationError("token", "activation token is invalid")
	}

	return nil
}

// DeactivateUserCommand represents a command to deactivate a user
type DeactivateUserCommand struct {
	UserID        string `json:"user_id"`
	Version       int    `json:"version"`
	DeactivatedBy string `json:"deactivated_by,omitempty"`
	Reason        string `json:"reason,omitempty"`
}

// Validate validates the deactivate user command
func (c *DeactivateUserCommand) Validate() error {
	if strings.TrimSpace(c.UserID) == "" {
		return NewValidationError("user_id", "user ID is required")
	}

	if c.Version <= 0 {
		return NewValidationError("version", "version must be greater than 0")
	}

	return nil
}

// AdminActivateUserCommand represents a command for admin to activate a user
type AdminActivateUserCommand struct {
	UserID      string `json:"user_id"`
	Version     int    `json:"version"`
	ActivatedBy string `json:"activated_by"`
	Reason      string `json:"reason,omitempty"`
}

// Validate validates the admin activate user command
func (c *AdminActivateUserCommand) Validate() error {
	if strings.TrimSpace(c.UserID) == "" {
		return NewValidationError("user_id", "user ID is required")
	}

	if c.Version <= 0 {
		return NewValidationError("version", "version must be greater than 0")
	}

	if strings.TrimSpace(c.ActivatedBy) == "" {
		return NewValidationError("activated_by", "activated by is required for admin activation")
	}

	return nil
}

// BulkActivationCommand represents a command to activate multiple users
type BulkActivationCommand struct {
	UserIDs     []string `json:"user_ids"`
	ActivatedBy string   `json:"activated_by"`
	Reason      string   `json:"reason,omitempty"`
}

// Validate validates the bulk activation command
func (c *BulkActivationCommand) Validate() error {
	if len(c.UserIDs) == 0 {
		return NewValidationError("user_ids", "at least one user ID is required")
	}

	if len(c.UserIDs) > 100 {
		return NewValidationError("user_ids", "cannot activate more than 100 users at once")
	}

	for i, userID := range c.UserIDs {
		if strings.TrimSpace(userID) == "" {
			return NewValidationError("user_ids", fmt.Sprintf("user ID at index %d is empty", i))
		}
	}

	if strings.TrimSpace(c.ActivatedBy) == "" {
		return NewValidationError("activated_by", "activated by is required")
	}

	return nil
}

// ResendActivationCommand represents a command to resend activation token
type ResendActivationCommand struct {
	UserID      string `json:"user_id"`
	RequestedBy string `json:"requested_by,omitempty"`
}

// Validate validates the resend activation command
func (c *ResendActivationCommand) Validate() error {
	if strings.TrimSpace(c.UserID) == "" {
		return NewValidationError("user_id", "user ID is required")
	}

	return nil
}
