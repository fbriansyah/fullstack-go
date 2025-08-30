package testing

import (
	"context"
	"time"

	"go-templ-template/internal/modules/auth/application"
	"go-templ-template/internal/modules/auth/domain"
	userApp "go-templ-template/internal/modules/user/application"
	userDomain "go-templ-template/internal/modules/user/domain"
	"go-templ-template/internal/shared/database"

	"github.com/stretchr/testify/mock"
)

// MockServices contains all mock services for testing
type MockServices struct {
	SessionRepo *MockSessionRepository
	EventBus    *MockEventBus
	RateLimiter *MockRateLimiter
	UserService *MockUserService
	UserRepo    *MockUserRepository
	AuditLogger *MockAuditLogger
	DB          *database.DB
}

// NewMockServices creates a new set of mock services for testing
func NewMockServices() *MockServices {
	return &MockServices{
		SessionRepo: NewMockSessionRepository(),
		EventBus:    NewMockEventBus(),
		RateLimiter: NewMockRateLimiter(),
		UserService: NewMockUserService(),
		UserRepo:    NewMockUserRepository(),
		AuditLogger: NewMockAuditLogger(),
		DB:          nil, // Will be set to a test database when needed
	}
}

// NewMockAuthService creates a new auth service with mock dependencies
func NewMockAuthService(mocks *MockServices) application.AuthService {
	sessionConfig := domain.SessionConfig{
		DefaultDuration: time.Hour * 24,
		MaxDuration:     time.Hour * 24 * 7,
		CleanupInterval: time.Hour,
	}

	return application.NewSimpleAuthService(
		mocks.SessionRepo,
		mocks.UserService,
		mocks.EventBus,
		mocks.RateLimiter,
		sessionConfig,
	)
}

// SetupDefaultMockBehavior configures default behavior for mock services
func SetupDefaultMockBehavior(mocks *MockServices) {
	// Default EventBus behavior - always succeed
	mocks.EventBus.On("Publish", mock.Anything, mock.Anything).Return(nil)

	// Default RateLimiter behavior - always allow
	mocks.RateLimiter.On("Allow", mock.Anything, mock.Anything).Return(true, nil)
	mocks.RateLimiter.On("Reset", mock.Anything, mock.Anything).Return(nil)

	// Default SessionRepository behavior
	mocks.SessionRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	// Default UserRepository behavior
	mocks.UserRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	mocks.UserRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	// Default AuditLogger behavior - always succeed
	mocks.AuditLogger.On("LogEvent", mock.Anything, mock.Anything).Return(nil)
}

// MockUserService provides a mock implementation of UserService for testing
type MockUserService struct {
	mock.Mock
	users map[string]*userDomain.User
}

// NewMockUserService creates a new mock user service
func NewMockUserService() *MockUserService {
	return &MockUserService{
		users: make(map[string]*userDomain.User),
	}
}

// CreateUser mocks user creation
func (m *MockUserService) CreateUser(ctx context.Context, cmd *userApp.CreateUserCommand) (*userDomain.User, error) {
	args := m.Called(ctx, cmd)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDomain.User), nil
}

// GetUser mocks user retrieval by ID
func (m *MockUserService) GetUser(ctx context.Context, query *userApp.GetUserQuery) (*userDomain.User, error) {
	args := m.Called(ctx, query)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDomain.User), nil
}

// GetUserByEmail mocks user retrieval by email
func (m *MockUserService) GetUserByEmail(ctx context.Context, query *userApp.GetUserByEmailQuery) (*userDomain.User, error) {
	args := m.Called(ctx, query)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDomain.User), nil
}

// UpdateUser mocks user update
func (m *MockUserService) UpdateUser(ctx context.Context, cmd *userApp.UpdateUserCommand) (*userDomain.User, error) {
	args := m.Called(ctx, cmd)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDomain.User), nil
}

// DeleteUser mocks user deletion
func (m *MockUserService) DeleteUser(ctx context.Context, cmd *userApp.DeleteUserCommand) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

// UpdateUserEmail mocks email update
func (m *MockUserService) UpdateUserEmail(ctx context.Context, cmd *userApp.UpdateUserEmailCommand) (*userDomain.User, error) {
	args := m.Called(ctx, cmd)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDomain.User), nil
}

// ChangeUserPassword mocks password change
func (m *MockUserService) ChangeUserPassword(ctx context.Context, cmd *userApp.ChangeUserPasswordCommand) (*userDomain.User, error) {
	args := m.Called(ctx, cmd)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDomain.User), nil
}

// ChangeUserStatus mocks status change
func (m *MockUserService) ChangeUserStatus(ctx context.Context, cmd *userApp.ChangeUserStatusCommand) (*userDomain.User, error) {
	args := m.Called(ctx, cmd)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDomain.User), nil
}

// ListUsers mocks user listing
func (m *MockUserService) ListUsers(ctx context.Context, query *userApp.ListUsersQuery) ([]*userDomain.User, int64, error) {
	args := m.Called(ctx, query)
	if args.Error(2) != nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*userDomain.User), args.Get(1).(int64), nil
}

// Test Data Factories

// CreateTestUser creates a test user with default values
func CreateTestUser(id, email string) *userDomain.User {
	user, _ := userDomain.NewUser(id, email, "Password123", "Test", "User")
	return user
}

// CreateTestSession creates a test session with default values
func CreateTestSession(userID string) *domain.Session {
	config := domain.SessionConfig{
		DefaultDuration: time.Hour * 24,
		MaxDuration:     time.Hour * 24 * 7,
		CleanupInterval: time.Hour,
	}

	session, _ := domain.NewSession(userID, "127.0.0.1", "test-agent", config)
	return session
}

// CreateExpiredTestSession creates an expired test session
func CreateExpiredTestSession(userID string) *domain.Session {
	session := CreateTestSession(userID)
	session.ExpiresAt = time.Now().Add(-time.Hour) // Expired 1 hour ago
	return session
}

// CreateInactiveTestSession creates an inactive test session
func CreateInactiveTestSession(userID string) *domain.Session {
	session := CreateTestSession(userID)
	session.IsActive = false
	return session
}

// NewMockUserRepository creates a new mock user repository with optional test data
func NewMockUserRepositoryWithData(users ...*userDomain.User) *MockUserRepository {
	repo := NewMockUserRepository()
	for _, user := range users {
		repo.users[user.ID] = user
	}
	return repo
}

// NewMockSessionRepositoryWithData creates a new mock session repository with optional test data
func NewMockSessionRepositoryWithData(sessions ...*domain.Session) *MockSessionRepository {
	repo := NewMockSessionRepository()
	for _, session := range sessions {
		repo.sessions[session.ID] = session
	}
	return repo
}

// NewMockEventBusWithCapture creates a mock event bus that captures all published events
func NewMockEventBusWithCapture() *MockEventBus {
	bus := NewMockEventBus()
	// Set up to capture all events without requiring explicit mock setup
	bus.On("Publish", mock.Anything, mock.Anything).Return(nil)
	return bus
}

// NewMockRateLimiterWithLimits creates a mock rate limiter with predefined limits
func NewMockRateLimiterWithLimits(allowedKeys []string, blockedKeys []string) *MockRateLimiter {
	limiter := NewMockRateLimiter()

	// Set up allowed keys
	for _, key := range allowedKeys {
		limiter.On("Allow", mock.Anything, key).Return(true, nil)
	}

	// Set up blocked keys
	for _, key := range blockedKeys {
		limiter.On("Allow", mock.Anything, key).Return(false, nil)
	}

	return limiter
}
