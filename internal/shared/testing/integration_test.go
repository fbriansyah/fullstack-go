package testing

import (
	"context"
	"testing"
	"time"

	"go-templ-template/internal/modules/auth/application"
	"go-templ-template/internal/modules/auth/domain"
	userApp "go-templ-template/internal/modules/user/application"
	"go-templ-template/internal/shared/audit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestFullAuthFlow_WithAllMocks demonstrates a complete authentication flow using all mock implementations
func TestFullAuthFlow_WithAllMocks(t *testing.T) {
	// Setup all mocks
	mocks := NewMockServices()
	SetupDefaultMockBehavior(mocks)

	// Create test user
	testUser := CreateTestUser("user-123", "test@example.com")

	// Setup UserService mock behavior for login
	mocks.UserService.On("GetUserByEmail", mock.Anything, &userApp.GetUserByEmailQuery{
		Email: "test@example.com",
	}).Return(testUser, nil)

	// Setup SessionRepository mock behavior for Create
	mocks.SessionRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Session")).Return(nil)

	// Create auth service with mocks
	authService := NewMockAuthService(mocks)

	// Test login
	ctx := context.Background()
	loginCmd := &application.LoginCommand{
		Email:     "test@example.com",
		Password:  "Password123",
		IPAddress: "127.0.0.1",
		UserAgent: "test-agent",
	}

	result, err := authService.Login(ctx, loginCmd)

	// Verify login succeeded
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Session)
	assert.NotEmpty(t, result.Session.ID)
	assert.Equal(t, testUser.ID, result.User.ID)

	// Verify all mocks were called as expected
	mocks.UserService.AssertExpectations(t)
	mocks.SessionRepo.AssertExpectations(t)
	mocks.EventBus.AssertExpectations(t)
	mocks.RateLimiter.AssertExpectations(t)

	// Verify events were published
	publishedEvents := mocks.EventBus.GetPublishedEvents()
	assert.Len(t, publishedEvents, 1)
	assert.Equal(t, "auth.user.logged_in", publishedEvents[0].EventType())
}

// TestUserRepository_CRUD_WithMocks demonstrates CRUD operations using MockUserRepository
func TestUserRepository_CRUD_WithMocks(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()

	// Test Create
	user := CreateTestUser("user-123", "test@example.com")
	repo.On("Create", ctx, user).Return(nil)

	err := repo.Create(ctx, user)
	assert.NoError(t, err)

	// Test GetByID
	repo.On("GetByID", ctx, user.ID).Return(user, nil)

	retrievedUser, err := repo.GetByID(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, retrievedUser.ID)

	// Test Update
	user.FirstName = "Updated"
	repo.On("Update", ctx, user).Return(nil)

	err = repo.Update(ctx, user)
	assert.NoError(t, err)

	// Test Delete
	repo.On("Delete", ctx, user.ID).Return(nil)

	err = repo.Delete(ctx, user.ID)
	assert.NoError(t, err)

	// Verify all expectations
	repo.AssertExpectations(t)
}

// TestAuditLogger_WithMocks demonstrates audit logging using MockAuditLogger
func TestAuditLogger_WithMocks(t *testing.T) {
	logger := NewMockAuditLogger()
	ctx := context.Background()

	// Create test audit event
	event := &audit.AuditEvent{
		EventID:       "event-123",
		EventType:     "user.login",
		AggregateID:   "user-123",
		AggregateType: "user",
		UserID:        "user-123",
		Action:        "login",
		Resource:      "user",
		ResourceID:    "user-123",
		Details:       map[string]interface{}{"ip": "127.0.0.1"},
		OccurredAt:    time.Now(),
	}

	// Setup mock behavior
	logger.On("LogEvent", ctx, event).Return(nil)

	// Execute
	err := logger.LogEvent(ctx, event)

	// Verify
	assert.NoError(t, err)
	logger.AssertExpectations(t)

	// Verify event was captured
	loggedEvents := logger.GetLoggedEvents()
	assert.Len(t, loggedEvents, 1)
	assert.Equal(t, event.EventID, loggedEvents[0].EventID)
	assert.Equal(t, event.EventType, loggedEvents[0].EventType)
}

// TestRateLimiter_Scenarios_WithMocks demonstrates different rate limiting scenarios
func TestRateLimiter_Scenarios_WithMocks(t *testing.T) {
	// Test with predefined limits
	allowedKeys := []string{"user-123"}
	blockedKeys := []string{"user-456"}
	limiter := NewMockRateLimiterWithLimits(allowedKeys, blockedKeys)

	ctx := context.Background()

	// Test allowed user
	allowed, err := limiter.Allow(ctx, "user-123")
	assert.NoError(t, err)
	assert.True(t, allowed)

	// Test blocked user
	blocked, err := limiter.Allow(ctx, "user-456")
	assert.NoError(t, err)
	assert.False(t, blocked)

	limiter.AssertExpectations(t)
}

// TestEventBus_PublishAndCapture_WithMocks demonstrates event publishing and capture
func TestEventBus_PublishAndCapture_WithMocks(t *testing.T) {
	bus := NewMockEventBusWithCapture()
	ctx := context.Background()

	// Publish multiple events
	event1 := domain.NewUserLoggedInEvent("user-123", "session-456", "127.0.0.1", "test-agent")
	event2 := domain.NewUserLoggedOutEvent("user-123", "session-456", "manual")

	err1 := bus.Publish(ctx, event1)
	err2 := bus.Publish(ctx, event2)

	// Verify no errors
	assert.NoError(t, err1)
	assert.NoError(t, err2)

	// Verify events were captured
	publishedEvents := bus.GetPublishedEvents()
	assert.Len(t, publishedEvents, 2)
	assert.Equal(t, "auth.user.logged_in", publishedEvents[0].EventType())
	assert.Equal(t, "auth.user.logged_out", publishedEvents[1].EventType())

	// Clear and verify
	bus.ClearPublishedEvents()
	assert.Len(t, bus.GetPublishedEvents(), 0)
}

// TestMockServices_CompleteSetup demonstrates setting up a complete mock environment
func TestMockServices_CompleteSetup(t *testing.T) {
	// Create all mocks
	mocks := NewMockServices()

	// Verify all services are initialized
	assert.NotNil(t, mocks.SessionRepo)
	assert.NotNil(t, mocks.EventBus)
	assert.NotNil(t, mocks.RateLimiter)
	assert.NotNil(t, mocks.UserService)
	assert.NotNil(t, mocks.UserRepo)
	assert.NotNil(t, mocks.AuditLogger)

	// Setup default behavior
	SetupDefaultMockBehavior(mocks)

	// Test that default behaviors work
	ctx := context.Background()

	// Test EventBus default behavior
	event := domain.NewUserLoggedInEvent("user-123", "session-456", "127.0.0.1", "test-agent")
	err := mocks.EventBus.Publish(ctx, event)
	assert.NoError(t, err)

	// Test RateLimiter default behavior
	allowed, err := mocks.RateLimiter.Allow(ctx, "test-key")
	assert.NoError(t, err)
	assert.True(t, allowed)

	// Test SessionRepo default behavior
	session := CreateTestSession("user-123")
	err = mocks.SessionRepo.Create(ctx, session)
	assert.NoError(t, err)

	// Test UserRepo default behavior
	user := CreateTestUser("user-123", "test@example.com")
	err = mocks.UserRepo.Create(ctx, user)
	assert.NoError(t, err)

	// Test AuditLogger default behavior
	auditEvent := &audit.AuditEvent{
		EventID:   "event-123",
		EventType: "test.event",
		UserID:    "user-123",
		Action:    "test",
		Resource:  "test",
	}
	err = mocks.AuditLogger.LogEvent(ctx, auditEvent)
	assert.NoError(t, err)
}

// TestMockFactories_WithPreloadedData demonstrates using factory functions with preloaded data
func TestMockFactories_WithPreloadedData(t *testing.T) {
	// Create test data
	user1 := CreateTestUser("user-1", "user1@example.com")
	user2 := CreateTestUser("user-2", "user2@example.com")
	session1 := CreateTestSession("user-1")
	session2 := CreateTestSession("user-2")

	// Create repositories with preloaded data
	userRepo := NewMockUserRepositoryWithData(user1, user2)
	sessionRepo := NewMockSessionRepositoryWithData(session1, session2)

	// Verify data was preloaded
	assert.Len(t, userRepo.users, 2)
	assert.Len(t, sessionRepo.sessions, 2)

	// Test fallback behavior (when no explicit mock is set)
	ctx := context.Background()

	// This should use fallback to internal storage
	retrievedUser, err := userRepo.GetByID(ctx, user1.ID)
	assert.NoError(t, err)
	assert.Equal(t, user1.ID, retrievedUser.ID)

	retrievedSession, err := sessionRepo.GetByID(ctx, session1.ID)
	assert.NoError(t, err)
	assert.Equal(t, session1.ID, retrievedSession.ID)
}
