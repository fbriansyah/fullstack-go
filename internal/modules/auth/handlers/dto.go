package handlers

import (
	"time"

	"go-templ-template/internal/modules/auth/domain"
	userDomain "go-templ-template/internal/modules/user/domain"
)

// LoginRequest represents the request payload for user login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required"`
}

// RegisterRequest represents the request payload for user registration
type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email,max=255"`
	Password  string `json:"password" validate:"required,min=8,max=128"`
	FirstName string `json:"first_name" validate:"required,max=100"`
	LastName  string `json:"last_name" validate:"required,max=100"`
}

// ChangePasswordRequest represents the request payload for changing password
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=128"`
}

// AuthResponse represents the response payload for successful authentication
type AuthResponse struct {
	User    *UserResponse    `json:"user"`
	Session *SessionResponse `json:"session"`
	Message string           `json:"message"`
}

// UserResponse represents user information in auth responses
type UserResponse struct {
	ID        string                `json:"id"`
	Email     string                `json:"email"`
	FirstName string                `json:"first_name"`
	LastName  string                `json:"last_name"`
	FullName  string                `json:"full_name"`
	Status    userDomain.UserStatus `json:"status"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
}

// SessionResponse represents session information in auth responses
type SessionResponse struct {
	ID        string    `json:"id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
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
func ToUserResponse(user *userDomain.User) *UserResponse {
	return &UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		FullName:  user.FullName(),
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// ToSessionResponse converts a domain Session to SessionResponse
func ToSessionResponse(session *domain.Session) *SessionResponse {
	return &SessionResponse{
		ID:        session.ID,
		ExpiresAt: session.ExpiresAt,
		CreatedAt: session.CreatedAt,
	}
}

// ToAuthResponse converts auth result to AuthResponse
func ToAuthResponse(user *userDomain.User, session *domain.Session, message string) *AuthResponse {
	return &AuthResponse{
		User:    ToUserResponse(user),
		Session: ToSessionResponse(session),
		Message: message,
	}
}
