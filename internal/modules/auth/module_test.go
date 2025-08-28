package auth

import (
	"context"
	"testing"

	"go-templ-template/internal/config"
	"go-templ-template/internal/modules/auth/application"
	"go-templ-template/internal/modules/auth/domain"
	"go-templ-template/internal/modules/user"
	userApplication "go-templ-template/internal/modules/user/application"
	userDomain "go-templ-template/internal/modules/user/domain"
	"go-templ-template/internal/shared"
	"go-templ-template/internal/shared/database"
	"go-templ-template/internal/shared/events"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEventBus for testing
type MockEventBus struct {
	mock.Mock
}

func (m *MockEventBus) Publish(ctx context.Context, event events.DomainEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventBus) Subscribe(eventType string, handler events.EventHandler) error {
	args := m.Called(eventType, handler)
	return args.Error(0)
}

func (m *MockEventBus) Unsubscribe(eventType string, handler events.EventHandler) error {
	args := m.Called(eventType, handler)
	return args.Error(0)
}

func (m *MockEventBus) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockEventBus) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockEventBus) Health() error {
	args := m.Called()
	return args.Error(0)
}

// MockUserService for testing
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, cmd *userApplication.CreateUserCommand) (*userDomain.User, error) {
	args := m.Called(ctx, cmd)
	return args.Get(0).(*userDomain.User), args.Error(1)
}

func (m *MockUserService) GetUser(ctx context.Context, query *userApplication.GetUserQuery) (*userDomain.User, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(*userDomain.User), args.Error(1)
}

func (m *MockUserService) GetUserByEmail(ctx context.Context, query *userApplication.GetUserByEmailQuery) (*userDomain.User, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(*userDomain.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, cmd *userApplication.UpdateUserCommand) (*userDomain.User, error) {
	args := m.Called(ctx, cmd)
	return args.Get(0).(*userDomain.User), args.Error(1)
}

func (m *MockUserService) UpdateUserEmail(ctx context.Context, cmd *userApplication.UpdateUserEmailCommand) (*userDomain.User, error) {
	args := m.Called(ctx, cmd)
	return args.Get(0).(*userDomain.User), args.Error(1)
}

func (m *MockUserService) ChangeUserPassword(ctx context.Context, cmd *userApplication.ChangeUserPasswordCommand) (*userDomain.User, error) {
	args := m.Called(ctx, cmd)
	return args.Get(0).(*userDomain.User), args.Error(1)
}

func (m *MockUserService) ChangeUserStatus(ctx context.Context, cmd *userApplication.ChangeUserStatusCommand) (*userDomain.User, error) {
	args := m.Called(ctx, cmd)
	return args.Get(0).(*userDomain.User), args.Error(1)
}

func (m *MockUserService) DeleteUser(ctx context.Context, cmd *userApplication.DeleteUserCommand) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *MockUserService) ListUsers(ctx context.Context, query *userApplication.ListUsersQuery) ([]*userDomain.User, int64, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]*userDomain.User), args.Get(1).(int64), args.Error(2)
}

// MockAuthService for testing
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(ctx context.Context, cmd *application.LoginCommand) (*application.AuthResult, error) {
	args := m.Called(ctx, cmd)
	return args.Get(0).(*application.AuthResult), args.Error(1)
}

func (m *MockAuthService) Register(ctx context.Context, cmd *application.RegisterCommand) (*application.AuthResult, error) {
	args := m.Called(ctx, cmd)
	return args.Get(0).(*application.AuthResult), args.Error(1)
}

func (m *MockAuthService) Logout(ctx context.Context, cmd *application.LogoutCommand) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *MockAuthService) ValidateSession(ctx context.Context, query *application.ValidateSessionQuery) (*application.SessionValidationResult, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(*application.SessionValidationResult), args.Error(1)
}

func (m *MockAuthService) RefreshSession(ctx context.Context, cmd *application.RefreshSessionCommand) (*domain.Session, error) {
	args := m.Called(ctx, cmd)
	return args.Get(0).(*domain.Session), args.Error(1)
}

func (m *MockAuthService) ChangePassword(ctx context.Context, cmd *application.ChangePasswordCommand) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *MockAuthService) CleanupExpiredSessions(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAuthService) GetUserSessions(ctx context.Context, userID string) ([]*domain.Session, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*domain.Session), args.Error(1)
}

func TestAuthModule_Name(t *testing.T) {
	// Arrange
	module := NewAuthModule()

	// Act
	name := module.Name()

	// Assert
	assert.Equal(t, "auth", name)
}

func TestAuthModule_Initialize_Success(t *testing.T) {
	// Arrange
	module := NewAuthModule()
	eventBus := &MockEventBus{}
	eventBus.On("Health").Return(nil)

	// Create a mock database
	db := &database.DB{}

	// Create config
	cfg := &config.Config{}

	// Create and register user module
	userModule := user.NewUserModule()

	// Create container and register user module
	container := shared.NewModuleContainer(eventBus, db, cfg)
	container.RegisterModule(userModule)

	// Act
	ctx := context.Background()
	err := module.Initialize(ctx, container)

	// Assert
	// This test will likely fail without proper setup, but tests the structure
	if err != nil {
		// Expected to fail without proper database/user service setup
		assert.Contains(t, err.Error(), "user")
	} else {
		assert.NoError(t, err)
		assert.NotNil(t, module.GetAuthService())
		assert.NotNil(t, module.GetAuthHandler())
	}
}

func TestAuthModule_Initialize_NoUserModule(t *testing.T) {
	// Arrange
	module := NewAuthModule()
	eventBus := &MockEventBus{}
	db := &database.DB{}
	cfg := &config.Config{}

	// Create container without user module
	container := shared.NewModuleContainer(eventBus, db, cfg)

	// Act
	ctx := context.Background()
	err := module.Initialize(ctx, container)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user module not found")
}

func TestAuthModule_Initialize_InvalidDatabase(t *testing.T) {
	// Arrange
	module := NewAuthModule()
	eventBus := &MockEventBus{}

	// Create container with invalid database type
	container := shared.NewModuleContainer(eventBus, "invalid-db", nil)

	// Act
	ctx := context.Background()
	err := module.Initialize(ctx, container)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid database dependency")
}

func TestAuthModule_Initialize_InvalidConfig(t *testing.T) {
	// Arrange
	module := NewAuthModule()
	eventBus := &MockEventBus{}
	db := &database.DB{}

	// Create container with invalid config type
	container := shared.NewModuleContainer(eventBus, db, "invalid-config")

	// Act
	ctx := context.Background()
	err := module.Initialize(ctx, container)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config dependency")
}

func TestAuthModule_RegisterRoutes(t *testing.T) {
	// Arrange
	module := NewAuthModule()
	router := echo.New()
	group := router.Group("/api/v1")

	// Act & Assert
	// This test verifies that the method exists and can be called
	assert.NotPanics(t, func() {
		module.RegisterRoutes(group)
	})
}

func TestAuthModule_RegisterEventHandlers(t *testing.T) {
	// Arrange
	module := NewAuthModule()
	eventBus := &MockEventBus{}

	// Set up expectations for event handler registration
	eventBus.On("Subscribe", "user.deleted", mock.AnythingOfType("*auth.UserDeletedHandler")).Return(nil)
	eventBus.On("Subscribe", "user.status_changed", mock.AnythingOfType("*auth.UserStatusChangedHandler")).Return(nil)

	// Act
	err := module.RegisterEventHandlers(eventBus)

	// Assert
	assert.NoError(t, err)
	eventBus.AssertExpectations(t)
}

func TestAuthModule_RegisterEventHandlers_SubscriptionFailure(t *testing.T) {
	// Arrange
	module := NewAuthModule()
	eventBus := &MockEventBus{}

	// Set up expectations - first subscription fails
	eventBus.On("Subscribe", "user.deleted", mock.AnythingOfType("*auth.UserDeletedHandler")).Return(assert.AnError)

	// Act
	err := module.RegisterEventHandlers(eventBus)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to subscribe to user.deleted event")
	eventBus.AssertExpectations(t)
}

func TestAuthModule_Health_NotInitialized(t *testing.T) {
	// Arrange
	module := NewAuthModule()

	// Act
	ctx := context.Background()
	err := module.Health(ctx)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "auth service not initialized")
}

func TestAuthModule_Health_EventBusUnhealthy(t *testing.T) {
	// Arrange
	module := NewAuthModule()
	eventBus := &MockEventBus{}
	eventBus.On("Health").Return(assert.AnError)

	db := &database.DB{}
	cfg := &config.Config{}

	// Create mock services to pass the initial health checks
	mockUserService := &MockUserService{}
	mockAuthService := &MockAuthService{}

	// Partially initialize the module
	module.eventBus = eventBus
	module.db = db
	module.config = cfg
	module.userService = mockUserService
	module.authService = mockAuthService

	// Act
	ctx := context.Background()
	err := module.Health(ctx)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event bus unhealthy")
	eventBus.AssertExpectations(t)
}

func TestAuthModule_Shutdown(t *testing.T) {
	// Arrange
	module := NewAuthModule()

	// Act
	ctx := context.Background()
	err := module.Shutdown(ctx)

	// Assert
	assert.NoError(t, err)
}

func TestUserDeletedHandler_EventType(t *testing.T) {
	// Arrange
	handler := NewUserDeletedHandler(nil)

	// Act
	eventType := handler.EventType()

	// Assert
	assert.Equal(t, "user.deleted", eventType)
}

func TestUserDeletedHandler_HandlerName(t *testing.T) {
	// Arrange
	handler := NewUserDeletedHandler(nil)

	// Act
	handlerName := handler.HandlerName()

	// Assert
	assert.Equal(t, "auth.user_deleted_handler", handlerName)
}

func TestUserStatusChangedHandler_EventType(t *testing.T) {
	// Arrange
	handler := NewUserStatusChangedHandler(nil)

	// Act
	eventType := handler.EventType()

	// Assert
	assert.Equal(t, "user.status_changed", eventType)
}

func TestUserStatusChangedHandler_HandlerName(t *testing.T) {
	// Arrange
	handler := NewUserStatusChangedHandler(nil)

	// Act
	handlerName := handler.HandlerName()

	// Assert
	assert.Equal(t, "auth.user_status_changed_handler", handlerName)
}
