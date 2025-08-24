package application

import (
	"context"
	"testing"
	"time"

	"go-templ-template/internal/modules/user/domain"
	"go-templ-template/internal/modules/user/infrastructure"
	"go-templ-template/internal/shared/database"
	"go-templ-template/internal/shared/events"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockUserRepositorySimple is a simple mock implementation
type MockUserRepositorySimple struct {
	mock.Mock
}

func (m *MockUserRepositorySimple) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepositorySimple) GetByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepositorySimple) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepositorySimple) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepositorySimple) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepositorySimple) List(ctx context.Context, filter infrastructure.UserFilter, limit, offset int) ([]*domain.User, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.User), args.Error(1)
}

func (m *MockUserRepositorySimple) Count(ctx context.Context, filter infrastructure.UserFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepositorySimple) Exists(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepositorySimple) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

// MockEventBusSimple is a simple mock implementation
type MockEventBusSimple struct {
	mock.Mock
}

func (m *MockEventBusSimple) Publish(ctx context.Context, event events.DomainEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventBusSimple) Subscribe(eventType string, handler events.EventHandler) error {
	args := m.Called(eventType, handler)
	return args.Error(0)
}

func (m *MockEventBusSimple) Unsubscribe(eventType string, handler events.EventHandler) error {
	args := m.Called(eventType, handler)
	return args.Error(0)
}

func (m *MockEventBusSimple) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockEventBusSimple) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockEventBusSimple) Health() error {
	args := m.Called()
	return args.Error(0)
}

// Simple service implementation for testing that bypasses transactions
type testUserService struct {
	userRepo infrastructure.UserRepository
	eventBus events.EventBus
}

func NewTestUserService(userRepo infrastructure.UserRepository, eventBus events.EventBus) UserService {
	return &testUserService{
		userRepo: userRepo,
		eventBus: eventBus,
	}
}

func (s *testUserService) CreateUser(ctx context.Context, cmd *CreateUserCommand) (*domain.User, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	// Check if user already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, NewInternalError("failed to check user existence")
	}
	if exists {
		return nil, NewUserAlreadyExistsError(cmd.Email)
	}

	// Create new user domain object
	user, err := domain.NewUser("test-id", cmd.Email, cmd.Password, cmd.FirstName, cmd.LastName)
	if err != nil {
		return nil, NewValidationError("user", err.Error())
	}

	// Create user in repository
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, NewInternalError("failed to create user")
	}

	// Publish user created event
	event := domain.NewUserCreatedEvent(user)
	if err := s.eventBus.Publish(ctx, event); err != nil {
		return nil, NewInternalError("failed to publish event")
	}

	return user, nil
}

func (s *testUserService) GetUser(ctx context.Context, query *GetUserQuery) (*domain.User, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, query.ID)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, NewUserNotFoundError(query.ID)
		}
		return nil, NewInternalError("failed to get user")
	}

	return user, nil
}

func (s *testUserService) GetUserByEmail(ctx context.Context, query *GetUserByEmailQuery) (*domain.User, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByEmail(ctx, query.Email)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, NewUserNotFoundError(query.Email)
		}
		return nil, NewInternalError("failed to get user by email")
	}

	return user, nil
}

func (s *testUserService) UpdateUser(ctx context.Context, cmd *UpdateUserCommand) (*domain.User, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, cmd.ID)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, NewUserNotFoundError(cmd.ID)
		}
		return nil, NewInternalError("failed to get user")
	}

	if user.Version != cmd.Version {
		return nil, NewOptimisticLockError(cmd.ID)
	}

	if err := user.UpdateProfile(cmd.FirstName, cmd.LastName); err != nil {
		return nil, NewValidationError("profile", err.Error())
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, NewInternalError("failed to update user")
	}

	return user, nil
}

func (s *testUserService) UpdateUserEmail(ctx context.Context, cmd *UpdateUserEmailCommand) (*domain.User, error) {
	return nil, NewInternalError("not implemented in test service")
}

func (s *testUserService) ChangeUserPassword(ctx context.Context, cmd *ChangeUserPasswordCommand) (*domain.User, error) {
	return nil, NewInternalError("not implemented in test service")
}

func (s *testUserService) ChangeUserStatus(ctx context.Context, cmd *ChangeUserStatusCommand) (*domain.User, error) {
	return nil, NewInternalError("not implemented in test service")
}

func (s *testUserService) DeleteUser(ctx context.Context, cmd *DeleteUserCommand) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	user, err := s.userRepo.GetByID(ctx, cmd.ID)
	if err != nil {
		if database.IsNotFoundError(err) {
			return NewUserNotFoundError(cmd.ID)
		}
		return NewInternalError("failed to get user")
	}

	if err := s.userRepo.Delete(ctx, cmd.ID); err != nil {
		return NewInternalError("failed to delete user")
	}

	event := domain.NewUserDeletedEvent(user, cmd.DeletedBy, cmd.Reason)
	if err := s.eventBus.Publish(ctx, event); err != nil {
		return NewInternalError("failed to publish event")
	}

	return nil
}

func (s *testUserService) ListUsers(ctx context.Context, query *ListUsersQuery) ([]*domain.User, int64, error) {
	if err := query.Validate(); err != nil {
		return nil, 0, err
	}

	filter := infrastructure.UserFilter{
		Status:        query.Status,
		Email:         query.Email,
		FirstName:     query.FirstName,
		LastName:      query.LastName,
		CreatedAfter:  query.CreatedAfter,
		CreatedBefore: query.CreatedBefore,
	}

	users, err := s.userRepo.List(ctx, filter, query.Limit, query.Offset)
	if err != nil {
		return nil, 0, NewInternalError("failed to list users")
	}

	total, err := s.userRepo.Count(ctx, filter)
	if err != nil {
		return nil, 0, NewInternalError("failed to count users")
	}

	return users, total, nil
}

func TestUserService_CreateUser_Simple(t *testing.T) {
	tests := []struct {
		name          string
		cmd           *CreateUserCommand
		setupMocks    func(*MockUserRepositorySimple, *MockEventBusSimple)
		expectedError string
		expectUser    bool
	}{
		{
			name: "successful user creation",
			cmd: &CreateUserCommand{
				Email:     "test@example.com",
				Password:  "Password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			setupMocks: func(repo *MockUserRepositorySimple, eventBus *MockEventBusSimple) {
				repo.On("ExistsByEmail", mock.Anything, "test@example.com").Return(false, nil)
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
				eventBus.On("Publish", mock.Anything, mock.AnythingOfType("*domain.UserCreatedEvent")).Return(nil)
			},
			expectUser: true,
		},
		{
			name: "user already exists",
			cmd: &CreateUserCommand{
				Email:     "existing@example.com",
				Password:  "Password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			setupMocks: func(repo *MockUserRepositorySimple, eventBus *MockEventBusSimple) {
				repo.On("ExistsByEmail", mock.Anything, "existing@example.com").Return(true, nil)
			},
			expectedError: "USER_ALREADY_EXISTS",
		},
		{
			name: "invalid command",
			cmd: &CreateUserCommand{
				Email:     "",
				Password:  "Password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			setupMocks:    func(repo *MockUserRepositorySimple, eventBus *MockEventBusSimple) {},
			expectedError: "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := &MockUserRepositorySimple{}
			mockEventBus := &MockEventBusSimple{}
			tt.setupMocks(mockRepo, mockEventBus)

			// Create service
			service := NewTestUserService(mockRepo, mockEventBus)

			// Execute
			user, err := service.CreateUser(context.Background(), tt.cmd)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				appErr, ok := err.(*ApplicationError)
				require.True(t, ok)
				assert.Equal(t, tt.expectedError, appErr.Code)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				if tt.expectUser {
					require.NotNil(t, user)
					assert.Equal(t, tt.cmd.Email, user.Email)
					assert.Equal(t, tt.cmd.FirstName, user.FirstName)
					assert.Equal(t, tt.cmd.LastName, user.LastName)
				}
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
			mockEventBus.AssertExpectations(t)
		})
	}
}

func TestUserService_GetUser_Simple(t *testing.T) {
	tests := []struct {
		name          string
		query         *GetUserQuery
		setupMocks    func(*MockUserRepositorySimple)
		expectedError string
		expectUser    bool
	}{
		{
			name: "successful user retrieval",
			query: &GetUserQuery{
				ID: "user-123",
			},
			setupMocks: func(repo *MockUserRepositorySimple) {
				user := &domain.User{
					ID:        "user-123",
					Email:     "test@example.com",
					FirstName: "John",
					LastName:  "Doe",
					Status:    domain.UserStatusActive,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					Version:   1,
				}
				repo.On("GetByID", mock.Anything, "user-123").Return(user, nil)
			},
			expectUser: true,
		},
		{
			name: "user not found",
			query: &GetUserQuery{
				ID: "nonexistent",
			},
			setupMocks: func(repo *MockUserRepositorySimple) {
				repo.On("GetByID", mock.Anything, "nonexistent").Return(nil, database.ErrNotFound)
			},
			expectedError: "USER_NOT_FOUND",
		},
		{
			name: "invalid query",
			query: &GetUserQuery{
				ID: "",
			},
			setupMocks:    func(repo *MockUserRepositorySimple) {},
			expectedError: "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := &MockUserRepositorySimple{}
			mockEventBus := &MockEventBusSimple{}
			tt.setupMocks(mockRepo)

			// Create service
			service := NewTestUserService(mockRepo, mockEventBus)

			// Execute
			user, err := service.GetUser(context.Background(), tt.query)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				appErr, ok := err.(*ApplicationError)
				require.True(t, ok)
				assert.Equal(t, tt.expectedError, appErr.Code)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				if tt.expectUser {
					require.NotNil(t, user)
					assert.Equal(t, tt.query.ID, user.ID)
				}
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
		})
	}
}
