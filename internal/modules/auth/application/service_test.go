package application

import (
	"context"
	"testing"
	"time"

	"go-templ-template/internal/modules/auth/domain"
	"go-templ-template/internal/modules/user/application"
	userDomain "go-templ-template/internal/modules/user/domain"
	"go-templ-template/internal/shared/events"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations
type mockSessionRepository struct {
	mock.Mock
}

func (m *mockSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *mockSessionRepository) GetByID(ctx context.Context, sessionID string) (*domain.Session, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Session), args.Error(1)
}

func (m *mockSessionRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Session, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Session), args.Error(1)
}

func (m *mockSessionRepository) Update(ctx context.Context, session *domain.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *mockSessionRepository) Delete(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *mockSessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *mockSessionRepository) DeleteExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockSessionRepository) ExistsByID(ctx context.Context, sessionID string) (bool, error) {
	args := m.Called(ctx, sessionID)
	return args.Bool(0), args.Error(1)
}

type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) CreateUser(ctx context.Context, cmd *application.CreateUserCommand) (*userDomain.User, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDomain.User), args.Error(1)
}

func (m *mockUserService) GetUser(ctx context.Context, query *application.GetUserQuery) (*userDomain.User, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDomain.User), args.Error(1)
}

func (m *mockUserService) GetUserByEmail(ctx context.Context, query *application.GetUserByEmailQuery) (*userDomain.User, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDomain.User), args.Error(1)
}

func (m *mockUserService) UpdateUser(ctx context.Context, cmd *application.UpdateUserCommand) (*userDomain.User, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDomain.User), args.Error(1)
}

func (m *mockUserService) UpdateUserEmail(ctx context.Context, cmd *application.UpdateUserEmailCommand) (*userDomain.User, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDomain.User), args.Error(1)
}

func (m *mockUserService) ChangeUserPassword(ctx context.Context, cmd *application.ChangeUserPasswordCommand) (*userDomain.User, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDomain.User), args.Error(1)
}

func (m *mockUserService) ChangeUserStatus(ctx context.Context, cmd *application.ChangeUserStatusCommand) (*userDomain.User, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDomain.User), args.Error(1)
}

func (m *mockUserService) DeleteUser(ctx context.Context, cmd *application.DeleteUserCommand) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *mockUserService) ListUsers(ctx context.Context, query *application.ListUsersQuery) ([]*userDomain.User, int64, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*userDomain.User), args.Get(1).(int64), args.Error(2)
}

type mockEventBus struct {
	mock.Mock
}

func (m *mockEventBus) Publish(ctx context.Context, event events.DomainEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *mockEventBus) Subscribe(eventType string, handler events.EventHandler) error {
	args := m.Called(eventType, handler)
	return args.Error(0)
}

func (m *mockEventBus) Unsubscribe(eventType string, handler events.EventHandler) error {
	args := m.Called(eventType, handler)
	return args.Error(0)
}

func (m *mockEventBus) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockEventBus) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockEventBus) Health() error {
	args := m.Called()
	return args.Error(0)
}

type mockRateLimiter struct {
	mock.Mock
}

func (m *mockRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *mockRateLimiter) Reset(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *mockRateLimiter) GetAttempts(ctx context.Context, key string) (int, error) {
	args := m.Called(ctx, key)
	return args.Int(0), args.Error(1)
}

// Helper functions
func createTestUser() *userDomain.User {
	user, _ := userDomain.NewUser("user-123", "test@example.com", "Password123!", "John", "Doe")
	return user
}

func createTestSession(userID string) *domain.Session {
	config := DefaultSessionConfig()
	session, _ := domain.NewSession(userID, "192.168.1.1", "test-agent", config)
	return session
}

// Test validation errors
func TestLoginCommand_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cmd     *LoginCommand
		wantErr bool
	}{
		{
			name: "valid command",
			cmd: &LoginCommand{
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "empty email",
			cmd: &LoginCommand{
				Email:    "",
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "empty password",
			cmd: &LoginCommand{
				Email:    "test@example.com",
				Password: "",
			},
			wantErr: true,
		},
		{
			name: "invalid email format",
			cmd: &LoginCommand{
				Email:    "invalid-email",
				Password: "password123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cmd.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegisterCommand_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cmd     *RegisterCommand
		wantErr bool
	}{
		{
			name: "valid command",
			cmd: &RegisterCommand{
				Email:     "test@example.com",
				Password:  "Password123!",
				FirstName: "John",
				LastName:  "Doe",
			},
			wantErr: false,
		},
		{
			name: "weak password",
			cmd: &RegisterCommand{
				Email:     "test@example.com",
				Password:  "weak",
				FirstName: "John",
				LastName:  "Doe",
			},
			wantErr: true,
		},
		{
			name: "common password",
			cmd: &RegisterCommand{
				Email:     "test@example.com",
				Password:  "password",
				FirstName: "John",
				LastName:  "Doe",
			},
			wantErr: true,
		},
		{
			name: "empty first name",
			cmd: &RegisterCommand{
				Email:     "test@example.com",
				Password:  "Password123!",
				FirstName: "",
				LastName:  "Doe",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cmd.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLogoutCommand_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cmd     *LogoutCommand
		wantErr bool
	}{
		{
			name: "valid command",
			cmd: &LogoutCommand{
				SessionID: "session-123",
			},
			wantErr: false,
		},
		{
			name: "empty session ID",
			cmd: &LogoutCommand{
				SessionID: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cmd.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateSessionQuery_Validate(t *testing.T) {
	tests := []struct {
		name    string
		query   *ValidateSessionQuery
		wantErr bool
	}{
		{
			name: "valid query",
			query: &ValidateSessionQuery{
				SessionID: "session-123",
			},
			wantErr: false,
		},
		{
			name: "empty session ID",
			query: &ValidateSessionQuery{
				SessionID: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.query.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestChangePasswordCommand_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cmd     *ChangePasswordCommand
		wantErr bool
	}{
		{
			name: "valid command",
			cmd: &ChangePasswordCommand{
				UserID:      "user-123",
				OldPassword: "OldPassword123!",
				NewPassword: "NewPassword456!",
			},
			wantErr: false,
		},
		{
			name: "empty user ID",
			cmd: &ChangePasswordCommand{
				UserID:      "",
				OldPassword: "OldPassword123!",
				NewPassword: "NewPassword456!",
			},
			wantErr: true,
		},
		{
			name: "empty old password",
			cmd: &ChangePasswordCommand{
				UserID:      "user-123",
				OldPassword: "",
				NewPassword: "NewPassword456!",
			},
			wantErr: true,
		},
		{
			name: "weak new password",
			cmd: &ChangePasswordCommand{
				UserID:      "user-123",
				OldPassword: "OldPassword123!",
				NewPassword: "weak",
			},
			wantErr: true,
		},
		{
			name: "same old and new password",
			cmd: &ChangePasswordCommand{
				UserID:      "user-123",
				OldPassword: "Password123!",
				NewPassword: "Password123!",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cmd.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultSessionConfig(t *testing.T) {
	config := DefaultSessionConfig()

	assert.Equal(t, time.Hour*24, config.DefaultDuration)
	assert.Equal(t, time.Hour*24*7, config.MaxDuration)
	assert.Equal(t, time.Hour, config.CleanupInterval)
}

func TestAuthError_Error(t *testing.T) {
	tests := []struct {
		name     string
		authErr  *AuthError
		expected string
	}{
		{
			name: "error with field",
			authErr: &AuthError{
				Code:    "VALIDATION_FAILED",
				Message: "Email is required",
				Field:   "email",
			},
			expected: "email: Email is required",
		},
		{
			name: "error without field",
			authErr: &AuthError{
				Code:    "INVALID_CREDENTIALS",
				Message: "Invalid email or password",
			},
			expected: "Invalid email or password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.authErr.Error())
		})
	}
}

func TestAuthError_HTTPStatusCode(t *testing.T) {
	tests := []struct {
		name     string
		authErr  *AuthError
		expected int
	}{
		{
			name:     "validation error",
			authErr:  &AuthError{Type: ErrorTypeValidation},
			expected: 400,
		},
		{
			name:     "authentication error",
			authErr:  &AuthError{Type: ErrorTypeAuthentication},
			expected: 401,
		},
		{
			name:     "authorization error",
			authErr:  &AuthError{Type: ErrorTypeAuthorization},
			expected: 403,
		},
		{
			name:     "not found error",
			authErr:  &AuthError{Type: ErrorTypeNotFound},
			expected: 404,
		},
		{
			name:     "rate limit error",
			authErr:  &AuthError{Type: ErrorTypeRateLimit},
			expected: 429,
		},
		{
			name:     "internal error",
			authErr:  &AuthError{Type: ErrorTypeInternal},
			expected: 500,
		},
		{
			name:     "unknown error type",
			authErr:  &AuthError{Type: "unknown"},
			expected: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.authErr.HTTPStatusCode())
		})
	}
}

func TestErrorConstructors(t *testing.T) {
	t.Run("NewValidationError", func(t *testing.T) {
		err := NewValidationError("email", "Email is required")
		assert.Equal(t, ErrorCodeValidationFailed, err.Code)
		assert.Equal(t, "Email is required", err.Message)
		assert.Equal(t, ErrorTypeValidation, err.Type)
		assert.Equal(t, "email", err.Field)
	})

	t.Run("NewInvalidCredentialsError", func(t *testing.T) {
		err := NewInvalidCredentialsError()
		assert.Equal(t, ErrorCodeInvalidCredentials, err.Code)
		assert.Equal(t, ErrorTypeAuthentication, err.Type)
	})

	t.Run("NewUserNotFoundError", func(t *testing.T) {
		err := NewUserNotFoundError("user-123")
		assert.Equal(t, ErrorCodeUserNotFound, err.Code)
		assert.Equal(t, ErrorTypeNotFound, err.Type)
		assert.Contains(t, err.Message, "user-123")
	})

	t.Run("NewSessionExpiredError", func(t *testing.T) {
		err := NewSessionExpiredError()
		assert.Equal(t, ErrorCodeSessionExpired, err.Code)
		assert.Equal(t, ErrorTypeAuthentication, err.Type)
	})

	t.Run("NewRateLimitExceededError", func(t *testing.T) {
		err := NewRateLimitExceededError("Too many attempts")
		assert.Equal(t, ErrorCodeRateLimitExceeded, err.Code)
		assert.Equal(t, ErrorTypeRateLimit, err.Type)
		assert.Equal(t, "Too many attempts", err.Message)
	})
}

func TestErrorTypeCheckers(t *testing.T) {
	t.Run("IsValidationError", func(t *testing.T) {
		validationErr := NewValidationError("field", "message")
		authErr := NewInvalidCredentialsError()
		otherErr := assert.AnError

		assert.True(t, IsValidationError(validationErr))
		assert.False(t, IsValidationError(authErr))
		assert.False(t, IsValidationError(otherErr))
	})

	t.Run("IsAuthenticationError", func(t *testing.T) {
		authErr := NewInvalidCredentialsError()
		validationErr := NewValidationError("field", "message")
		otherErr := assert.AnError

		assert.True(t, IsAuthenticationError(authErr))
		assert.False(t, IsAuthenticationError(validationErr))
		assert.False(t, IsAuthenticationError(otherErr))
	})

	t.Run("IsRateLimitError", func(t *testing.T) {
		rateLimitErr := NewRateLimitExceededError("message")
		authErr := NewInvalidCredentialsError()
		otherErr := assert.AnError

		assert.True(t, IsRateLimitError(rateLimitErr))
		assert.False(t, IsRateLimitError(authErr))
		assert.False(t, IsRateLimitError(otherErr))
	})
}
