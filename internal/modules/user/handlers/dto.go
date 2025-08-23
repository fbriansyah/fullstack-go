package handlers

import (
	"time"

	"go-templ-template/internal/modules/user/domain"
)

// CreateUserRequest represents the request payload for creating a user
type CreateUserRequest struct {
	Email     string `json:"email" validate:"required,email,max=255"`
	Password  string `json:"password" validate:"required,min=8,max=128"`
	FirstName string `json:"first_name" validate:"required,max=100"`
	LastName  string `json:"last_name" validate:"required,max=100"`
}

// UpdateUserRequest represents the request payload for updating a user
type UpdateUserRequest struct {
	FirstName string `json:"first_name" validate:"required,max=100"`
	LastName  string `json:"last_name" validate:"required,max=100"`
	Version   int    `json:"version" validate:"required,min=1"`
}

// UpdateUserEmailRequest represents the request payload for updating a user's email
type UpdateUserEmailRequest struct {
	Email   string `json:"email" validate:"required,email,max=255"`
	Version int    `json:"version" validate:"required,min=1"`
}

// ChangeUserPasswordRequest represents the request payload for changing a user's password
type ChangeUserPasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=128"`
	Version     int    `json:"version" validate:"required,min=1"`
}

// ChangeUserStatusRequest represents the request payload for changing a user's status
type ChangeUserStatusRequest struct {
	Status    domain.UserStatus `json:"status" validate:"required"`
	ChangedBy string            `json:"changed_by,omitempty"`
	Reason    string            `json:"reason,omitempty"`
	Version   int               `json:"version" validate:"required,min=1"`
}

// ListUsersRequest represents the request parameters for listing users
type ListUsersRequest struct {
	Status        *domain.UserStatus `query:"status"`
	Email         *string            `query:"email"`
	FirstName     *string            `query:"first_name"`
	LastName      *string            `query:"last_name"`
	CreatedAfter  *string            `query:"created_after"`
	CreatedBefore *string            `query:"created_before"`
	Limit         int                `query:"limit"`
	Offset        int                `query:"offset"`
}

// UserResponse represents the response payload for a user
type UserResponse struct {
	ID        string            `json:"id"`
	Email     string            `json:"email"`
	FirstName string            `json:"first_name"`
	LastName  string            `json:"last_name"`
	FullName  string            `json:"full_name"`
	Status    domain.UserStatus `json:"status"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Version   int               `json:"version"`
}

// ListUsersResponse represents the response payload for listing users
type ListUsersResponse struct {
	Users   []*UserResponse `json:"users"`
	Total   int64           `json:"total"`
	Limit   int             `json:"limit"`
	Offset  int             `json:"offset"`
	HasMore bool            `json:"has_more"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ToUserResponse converts a domain User to UserResponse
func ToUserResponse(user *domain.User) *UserResponse {
	return &UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		FullName:  user.FullName(),
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Version:   user.Version,
	}
}

// ToUserResponseList converts a slice of domain Users to UserResponse slice
func ToUserResponseList(users []*domain.User) []*UserResponse {
	responses := make([]*UserResponse, len(users))
	for i, user := range users {
		responses[i] = ToUserResponse(user)
	}
	return responses
}
