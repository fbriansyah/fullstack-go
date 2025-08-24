package application

import (
	"go-templ-template/internal/modules/user/domain"
)

// CreateUserCommand represents a command to create a new user
type CreateUserCommand struct {
	Email     string `json:"email" validate:"required,email,max=255"`
	Password  string `json:"password" validate:"required,min=8,max=128"`
	FirstName string `json:"first_name" validate:"required,max=100"`
	LastName  string `json:"last_name" validate:"required,max=100"`
}

// Validate performs validation on the CreateUserCommand
func (c *CreateUserCommand) Validate() error {
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
	return nil
}

// UpdateUserCommand represents a command to update an existing user
type UpdateUserCommand struct {
	ID        string `json:"id" validate:"required"`
	FirstName string `json:"first_name" validate:"required,max=100"`
	LastName  string `json:"last_name" validate:"required,max=100"`
	Version   int    `json:"version" validate:"required,min=1"`
}

// Validate performs validation on the UpdateUserCommand
func (c *UpdateUserCommand) Validate() error {
	if c.ID == "" {
		return NewValidationError("id", "user ID is required")
	}
	if c.FirstName == "" {
		return NewValidationError("first_name", "first name is required")
	}
	if c.LastName == "" {
		return NewValidationError("last_name", "last name is required")
	}
	if c.Version < 1 {
		return NewValidationError("version", "version must be greater than 0")
	}
	return nil
}

// UpdateUserEmailCommand represents a command to update a user's email
type UpdateUserEmailCommand struct {
	ID      string `json:"id" validate:"required"`
	Email   string `json:"email" validate:"required,email,max=255"`
	Version int    `json:"version" validate:"required,min=1"`
}

// Validate performs validation on the UpdateUserEmailCommand
func (c *UpdateUserEmailCommand) Validate() error {
	if c.ID == "" {
		return NewValidationError("id", "user ID is required")
	}
	if c.Email == "" {
		return NewValidationError("email", "email is required")
	}
	if c.Version < 1 {
		return NewValidationError("version", "version must be greater than 0")
	}
	return nil
}

// ChangeUserPasswordCommand represents a command to change a user's password
type ChangeUserPasswordCommand struct {
	ID          string `json:"id" validate:"required"`
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=128"`
	Version     int    `json:"version" validate:"required,min=1"`
}

// Validate performs validation on the ChangeUserPasswordCommand
func (c *ChangeUserPasswordCommand) Validate() error {
	if c.ID == "" {
		return NewValidationError("id", "user ID is required")
	}
	if c.OldPassword == "" {
		return NewValidationError("old_password", "old password is required")
	}
	if c.NewPassword == "" {
		return NewValidationError("new_password", "new password is required")
	}
	if c.Version < 1 {
		return NewValidationError("version", "version must be greater than 0")
	}
	return nil
}

// ChangeUserStatusCommand represents a command to change a user's status
type ChangeUserStatusCommand struct {
	ID        string            `json:"id" validate:"required"`
	Status    domain.UserStatus `json:"status" validate:"required"`
	ChangedBy string            `json:"changed_by,omitempty"`
	Reason    string            `json:"reason,omitempty"`
	Version   int               `json:"version" validate:"required,min=1"`
}

// Validate performs validation on the ChangeUserStatusCommand
func (c *ChangeUserStatusCommand) Validate() error {
	if c.ID == "" {
		return NewValidationError("id", "user ID is required")
	}
	if !c.Status.IsValid() {
		return NewValidationError("status", "invalid user status")
	}
	if c.Version < 1 {
		return NewValidationError("version", "version must be greater than 0")
	}
	return nil
}

// DeleteUserCommand represents a command to delete a user
type DeleteUserCommand struct {
	ID        string `json:"id" validate:"required"`
	DeletedBy string `json:"deleted_by,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

// Validate performs validation on the DeleteUserCommand
func (c *DeleteUserCommand) Validate() error {
	if c.ID == "" {
		return NewValidationError("id", "user ID is required")
	}
	return nil
}

// GetUserQuery represents a query to get a user by ID
type GetUserQuery struct {
	ID string `json:"id" validate:"required"`
}

// Validate performs validation on the GetUserQuery
func (q *GetUserQuery) Validate() error {
	if q.ID == "" {
		return NewValidationError("id", "user ID is required")
	}
	return nil
}

// GetUserByEmailQuery represents a query to get a user by email
type GetUserByEmailQuery struct {
	Email string `json:"email" validate:"required,email"`
}

// Validate performs validation on the GetUserByEmailQuery
func (q *GetUserByEmailQuery) Validate() error {
	if q.Email == "" {
		return NewValidationError("email", "email is required")
	}
	return nil
}

// ListUsersQuery represents a query to list users with filtering and pagination
type ListUsersQuery struct {
	Status        *domain.UserStatus `json:"status,omitempty"`
	Email         *string            `json:"email,omitempty"`
	FirstName     *string            `json:"first_name,omitempty"`
	LastName      *string            `json:"last_name,omitempty"`
	CreatedAfter  *string            `json:"created_after,omitempty"`
	CreatedBefore *string            `json:"created_before,omitempty"`
	Limit         int                `json:"limit" validate:"min=1,max=100"`
	Offset        int                `json:"offset" validate:"min=0"`
}

// Validate performs validation on the ListUsersQuery
func (q *ListUsersQuery) Validate() error {
	if q.Limit <= 0 {
		q.Limit = 20 // Default limit
	}
	if q.Limit > 100 {
		return NewValidationError("limit", "limit cannot exceed 100")
	}
	if q.Offset < 0 {
		return NewValidationError("offset", "offset cannot be negative")
	}
	return nil
}
