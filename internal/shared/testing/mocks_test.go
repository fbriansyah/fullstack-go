package testing

import (
	"context"
	"testing"
	"time"

	"go-templ-template/internal/modules/auth/domain"
	userDomain "go-templ-template/internal/modules/user/domain"
	"go-templ-template/internal/shared/audit"
	"go-templ-template/internal/shared/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMockSessionRepository_Create(t *testing.T) {
	mockRepo := NewMockSessionRepository()
	ctx := context.Background()

	session := CreateTestSession("user-123")

	// Setup mock expectation
	mockRepo.On("Create", ctx, session).Return(nil)

	// Execute
	err := mockRepo.Create(ctx, session)

	// Verify
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)

	// Verify session was stored in mock
	storedSession, exists := mockRepo.sessions[session.ID]
	assert.True(t, exists)
	assert.Equal(t, session.ID, storedSession.ID)
	assert.Equal(t, session.UserID, storedSession.UserID)
}

func TestMockSessionRepository_GetByID(t *testing.T) {
	mockRepo := NewMockSessionRepository()
	ctx := context.Background()

	session := CreateTestSession("user-123")

	// Store session in mock
	mockRepo.sessions[session.ID] = session

	// Setup mock expectation
	mockRepo.On("GetByID", ctx, session.ID).Return(session, nil)

	// Execute
	retrievedSession, err := mockRepo.GetByID(ctx, session.ID)

	// Verify
	assert.NoError(t, err)
	assert.Equal(t, session.ID, retrievedSession.ID)
	assert.Equal(t, session.UserID, retrievedSession.UserID)
	mockRepo.AssertExpectations(t)
}

func TestMockSessionRepository_GetByID_NotFound(t *testing.T) {
	mockRepo := NewMockSessionRepository()
	ctx := context.Background()

	// Setup mock expectation for non-existent session
	mockRepo.On("GetByID", ctx, "non-existent").Return((*domain.Session)(nil), database.ErrNotFound)

	// Execute
	session, err := mockRepo.GetByID(ctx, "non-existent")

	// Verify
	assert.Nil(t, session)
	assert.Equal(t, database.ErrNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestMockSessionRepository_GetByUserID(t *testing.T) {
	mockRepo := NewMockSessionRepository()
	ctx := context.Background()
	userID := "user-123"

	// Create test sessions
	session1 := CreateTestSession(userID)
	session2 := CreateTestSession(userID)
	session3 := CreateTestSession("other-user")

	// Store sessions in mock
	mockRepo.sessions[session1.ID] = session1
	mockRepo.sessions[session2.ID] = session2
	mockRepo.sessions[session3.ID] = session3

	// Setup mock expectation
	expectedSessions := []*domain.Session{session1, session2}
	mockRepo.On("GetByUserID", ctx, userID).Return(expectedSessions, nil)

	// Execute
	sessions, err := mockRepo.GetByUserID(ctx, userID)

	// Verify
	assert.NoError(t, err)
	assert.Len(t, sessions, 2)
	mockRepo.AssertExpectations(t)
}

func TestMockSessionRepository_Update(t *testing.T) {
	mockRepo := NewMockSessionRepository()
	ctx := context.Background()

	session := CreateTestSession("user-123")
	mockRepo.sessions[session.ID] = session

	// Modify session
	session.ExpiresAt = time.Now().Add(time.Hour * 48)

	// Setup mock expectation
	mockRepo.On("Update", ctx, session).Return(nil)

	// Execute
	err := mockRepo.Update(ctx, session)

	// Verify
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)

	// Verify session was updated in mock
	updatedSession := mockRepo.sessions[session.ID]
	assert.Equal(t, session.ExpiresAt, updatedSession.ExpiresAt)
}

func TestMockSessionRepository_Delete(t *testing.T) {
	mockRepo := NewMockSessionRepository()
	ctx := context.Background()

	session := CreateTestSession("user-123")
	mockRepo.sessions[session.ID] = session

	// Setup mock expectation
	mockRepo.On("Delete", ctx, session.ID).Return(nil)

	// Execute
	err := mockRepo.Delete(ctx, session.ID)

	// Verify
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)

	// Verify session was deleted from mock
	_, exists := mockRepo.sessions[session.ID]
	assert.False(t, exists)
}

func TestMockSessionRepository_DeleteByUserID(t *testing.T) {
	mockRepo := NewMockSessionRepository()
	ctx := context.Background()
	userID := "user-123"

	// Create test sessions
	session1 := CreateTestSession(userID)
	session2 := CreateTestSession(userID)
	session3 := CreateTestSession("other-user")

	// Store sessions in mock
	mockRepo.sessions[session1.ID] = session1
	mockRepo.sessions[session2.ID] = session2
	mockRepo.sessions[session3.ID] = session3

	// Setup mock expectation
	mockRepo.On("DeleteByUserID", ctx, userID).Return(nil)

	// Execute
	err := mockRepo.DeleteByUserID(ctx, userID)

	// Verify
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)

	// Verify only user sessions were deleted
	_, exists1 := mockRepo.sessions[session1.ID]
	_, exists2 := mockRepo.sessions[session2.ID]
	_, exists3 := mockRepo.sessions[session3.ID]

	assert.False(t, exists1)
	assert.False(t, exists2)
	assert.True(t, exists3) // Other user's session should remain
}

func TestMockEventBus_Publish(t *testing.T) {
	mockBus := NewMockEventBus()
	ctx := context.Background()

	// Create a test event
	event := domain.NewUserLoggedInEvent("user-123", "session-456", "127.0.0.1", "test-agent")

	// Setup mock expectation
	mockBus.On("Publish", ctx, event).Return(nil)

	// Execute
	err := mockBus.Publish(ctx, event)

	// Verify
	assert.NoError(t, err)
	mockBus.AssertExpectations(t)

	// Verify event was stored in mock
	publishedEvents := mockBus.GetPublishedEvents()
	assert.Len(t, publishedEvents, 1)
	assert.Equal(t, event.EventType(), publishedEvents[0].EventType())
}

func TestMockEventBus_ClearPublishedEvents(t *testing.T) {
	mockBus := NewMockEventBus()
	ctx := context.Background()

	// Publish some events
	event1 := domain.NewUserLoggedInEvent("user-123", "session-456", "127.0.0.1", "test-agent")
	event2 := domain.NewUserLoggedOutEvent("user-123", "session-456", "manual")

	mockBus.On("Publish", ctx, mock.Anything).Return(nil)

	mockBus.Publish(ctx, event1)
	mockBus.Publish(ctx, event2)

	// Verify events were published
	assert.Len(t, mockBus.GetPublishedEvents(), 2)

	// Clear events
	mockBus.ClearPublishedEvents()

	// Verify events were cleared
	assert.Len(t, mockBus.GetPublishedEvents(), 0)
}

func TestMockRateLimiter_Allow(t *testing.T) {
	mockLimiter := NewMockRateLimiter()
	ctx := context.Background()
	key := "test-key"

	// Setup mock expectation
	mockLimiter.On("Allow", ctx, key).Return(true, nil)

	// Execute
	allowed, err := mockLimiter.Allow(ctx, key)

	// Verify
	assert.NoError(t, err)
	assert.True(t, allowed)
	mockLimiter.AssertExpectations(t)

	// Verify attempt was recorded in mock
	assert.Equal(t, 1, mockLimiter.attempts[key])
}

func TestMockRateLimiter_Reset(t *testing.T) {
	mockLimiter := NewMockRateLimiter()
	ctx := context.Background()
	key := "test-key"

	// Set some attempts
	mockLimiter.SetAttempts(key, 3)
	assert.Equal(t, 3, mockLimiter.attempts[key])

	// Setup mock expectation
	mockLimiter.On("Reset", ctx, key).Return(nil)

	// Execute
	err := mockLimiter.Reset(ctx, key)

	// Verify
	assert.NoError(t, err)
	mockLimiter.AssertExpectations(t)

	// Verify attempts were cleared
	_, exists := mockLimiter.attempts[key]
	assert.False(t, exists)
}

func TestMockRateLimiter_GetAttempts(t *testing.T) {
	mockLimiter := NewMockRateLimiter()
	ctx := context.Background()
	key := "test-key"

	// Set some attempts
	mockLimiter.SetAttempts(key, 5)

	// Setup mock expectation
	mockLimiter.On("GetAttempts", ctx, key).Return(5, nil)

	// Execute
	attempts, err := mockLimiter.GetAttempts(ctx, key)

	// Verify
	assert.NoError(t, err)
	assert.Equal(t, 5, attempts)
	mockLimiter.AssertExpectations(t)
}

func TestCreateTestUser(t *testing.T) {
	user := CreateTestUser("user-123", "test@example.com")

	assert.Equal(t, "user-123", user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "Test", user.FirstName)
	assert.Equal(t, "User", user.LastName)
	assert.True(t, user.IsActive())
}

func TestCreateTestSession(t *testing.T) {
	userID := "user-123"
	session := CreateTestSession(userID)

	assert.Equal(t, userID, session.UserID)
	assert.Equal(t, "127.0.0.1", session.IPAddress)
	assert.Equal(t, "test-agent", session.UserAgent)
	assert.True(t, session.IsActive)
	assert.False(t, session.IsExpired())
	assert.True(t, session.IsValid())
}

func TestCreateExpiredTestSession(t *testing.T) {
	userID := "user-123"
	session := CreateExpiredTestSession(userID)

	assert.Equal(t, userID, session.UserID)
	assert.True(t, session.IsActive)
	assert.True(t, session.IsExpired())
	assert.False(t, session.IsValid())
}

func TestCreateInactiveTestSession(t *testing.T) {
	userID := "user-123"
	session := CreateInactiveTestSession(userID)

	assert.Equal(t, userID, session.UserID)
	assert.False(t, session.IsActive)
	assert.False(t, session.IsExpired())
	assert.False(t, session.IsValid())
}

func TestMockUserRepository_Create(t *testing.T) {
	mockRepo := NewMockUserRepository()
	ctx := context.Background()

	user := CreateTestUser("user-123", "test@example.com")

	// Setup mock expectation
	mockRepo.On("Create", ctx, user).Return(nil)

	// Execute
	err := mockRepo.Create(ctx, user)

	// Verify
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)

	// Verify user was stored in mock
	storedUser, exists := mockRepo.users[user.ID]
	assert.True(t, exists)
	assert.Equal(t, user.ID, storedUser.ID)
	assert.Equal(t, user.Email, storedUser.Email)
}

func TestMockUserRepository_GetByID(t *testing.T) {
	mockRepo := NewMockUserRepository()
	ctx := context.Background()

	user := CreateTestUser("user-123", "test@example.com")

	// Store user in mock
	mockRepo.users[user.ID] = user

	// Setup mock expectation
	mockRepo.On("GetByID", ctx, user.ID).Return(user, nil)

	// Execute
	retrievedUser, err := mockRepo.GetByID(ctx, user.ID)

	// Verify
	assert.NoError(t, err)
	assert.Equal(t, user.ID, retrievedUser.ID)
	assert.Equal(t, user.Email, retrievedUser.Email)
	mockRepo.AssertExpectations(t)
}

func TestMockUserRepository_GetByID_NotFound(t *testing.T) {
	mockRepo := NewMockUserRepository()
	ctx := context.Background()

	// Setup mock expectation for non-existent user
	mockRepo.On("GetByID", ctx, "non-existent").Return((*userDomain.User)(nil), database.ErrNotFound)

	// Execute
	user, err := mockRepo.GetByID(ctx, "non-existent")

	// Verify
	assert.Nil(t, user)
	assert.Equal(t, database.ErrNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestMockUserRepository_GetByEmail(t *testing.T) {
	mockRepo := NewMockUserRepository()
	ctx := context.Background()

	user := CreateTestUser("user-123", "test@example.com")

	// Store user in mock
	mockRepo.users[user.ID] = user

	// Setup mock expectation
	mockRepo.On("GetByEmail", ctx, user.Email).Return(user, nil)

	// Execute
	retrievedUser, err := mockRepo.GetByEmail(ctx, user.Email)

	// Verify
	assert.NoError(t, err)
	assert.Equal(t, user.ID, retrievedUser.ID)
	assert.Equal(t, user.Email, retrievedUser.Email)
	mockRepo.AssertExpectations(t)
}

func TestMockUserRepository_Update(t *testing.T) {
	mockRepo := NewMockUserRepository()
	ctx := context.Background()

	user := CreateTestUser("user-123", "test@example.com")
	mockRepo.users[user.ID] = user

	// Modify user
	user.FirstName = "Updated"

	// Setup mock expectation
	mockRepo.On("Update", ctx, user).Return(nil)

	// Execute
	err := mockRepo.Update(ctx, user)

	// Verify
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)

	// Verify user was updated in mock
	updatedUser := mockRepo.users[user.ID]
	assert.Equal(t, "Updated", updatedUser.FirstName)
}

func TestMockUserRepository_Delete(t *testing.T) {
	mockRepo := NewMockUserRepository()
	ctx := context.Background()

	user := CreateTestUser("user-123", "test@example.com")
	mockRepo.users[user.ID] = user

	// Setup mock expectation
	mockRepo.On("Delete", ctx, user.ID).Return(nil)

	// Execute
	err := mockRepo.Delete(ctx, user.ID)

	// Verify
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)

	// Verify user was deleted from mock
	_, exists := mockRepo.users[user.ID]
	assert.False(t, exists)
}

func TestMockUserRepository_ExistsByEmail(t *testing.T) {
	mockRepo := NewMockUserRepository()
	ctx := context.Background()

	user := CreateTestUser("user-123", "test@example.com")
	mockRepo.users[user.ID] = user

	// Setup mock expectation
	mockRepo.On("ExistsByEmail", ctx, user.Email).Return(true, nil)

	// Execute
	exists, err := mockRepo.ExistsByEmail(ctx, user.Email)

	// Verify
	assert.NoError(t, err)
	assert.True(t, exists)
	mockRepo.AssertExpectations(t)
}

func TestMockAuditLogger_LogEvent(t *testing.T) {
	mockLogger := NewMockAuditLogger()
	ctx := context.Background()

	event := &audit.AuditEvent{
		EventID:   "event-123",
		EventType: "user.login",
		UserID:    "user-123",
		Action:    "login",
		Resource:  "user",
	}

	// Setup mock expectation
	mockLogger.On("LogEvent", ctx, event).Return(nil)

	// Execute
	err := mockLogger.LogEvent(ctx, event)

	// Verify
	assert.NoError(t, err)
	mockLogger.AssertExpectations(t)

	// Verify event was stored in mock
	loggedEvents := mockLogger.GetLoggedEvents()
	assert.Len(t, loggedEvents, 1)
	assert.Equal(t, event.EventID, loggedEvents[0].EventID)
	assert.Equal(t, event.EventType, loggedEvents[0].EventType)
}

func TestMockAuditLogger_GetEvents(t *testing.T) {
	mockLogger := NewMockAuditLogger()
	ctx := context.Background()

	// Add some events to mock
	event1 := audit.AuditEvent{EventID: "event-1", EventType: "user.login"}
	event2 := audit.AuditEvent{EventID: "event-2", EventType: "user.logout"}
	mockLogger.events = []audit.AuditEvent{event1, event2}

	filter := &audit.AuditFilter{}
	expectedEvents := []*audit.AuditEvent{&event1, &event2}

	// Setup mock expectation
	mockLogger.On("GetEvents", ctx, filter).Return(expectedEvents, nil)

	// Execute
	events, err := mockLogger.GetEvents(ctx, filter)

	// Verify
	assert.NoError(t, err)
	assert.Len(t, events, 2)
	assert.Equal(t, "event-1", events[0].EventID)
	assert.Equal(t, "event-2", events[1].EventID)
	mockLogger.AssertExpectations(t)
}

func TestMockAuditLogger_ClearLoggedEvents(t *testing.T) {
	mockLogger := NewMockAuditLogger()

	// Add some events
	event1 := audit.AuditEvent{EventID: "event-1", EventType: "user.login"}
	event2 := audit.AuditEvent{EventID: "event-2", EventType: "user.logout"}
	mockLogger.events = []audit.AuditEvent{event1, event2}

	// Verify events were added
	assert.Len(t, mockLogger.GetLoggedEvents(), 2)

	// Clear events
	mockLogger.ClearLoggedEvents()

	// Verify events were cleared
	assert.Len(t, mockLogger.GetLoggedEvents(), 0)
}

func TestNewMockUserRepositoryWithData(t *testing.T) {
	user1 := CreateTestUser("user-1", "user1@example.com")
	user2 := CreateTestUser("user-2", "user2@example.com")

	repo := NewMockUserRepositoryWithData(user1, user2)

	// Verify users were stored
	assert.Len(t, repo.users, 2)
	assert.Equal(t, user1, repo.users[user1.ID])
	assert.Equal(t, user2, repo.users[user2.ID])
}

func TestNewMockSessionRepositoryWithData(t *testing.T) {
	session1 := CreateTestSession("user-1")
	session2 := CreateTestSession("user-2")

	repo := NewMockSessionRepositoryWithData(session1, session2)

	// Verify sessions were stored
	assert.Len(t, repo.sessions, 2)
	assert.Equal(t, session1, repo.sessions[session1.ID])
	assert.Equal(t, session2, repo.sessions[session2.ID])
}

func TestNewMockEventBusWithCapture(t *testing.T) {
	bus := NewMockEventBusWithCapture()
	ctx := context.Background()

	event := domain.NewUserLoggedInEvent("user-123", "session-456", "127.0.0.1", "test-agent")

	// Execute - should work without explicit mock setup
	err := bus.Publish(ctx, event)

	// Verify
	assert.NoError(t, err)

	// Verify event was captured
	publishedEvents := bus.GetPublishedEvents()
	assert.Len(t, publishedEvents, 1)
	assert.Equal(t, event.EventType(), publishedEvents[0].EventType())
}

func TestNewMockRateLimiterWithLimits(t *testing.T) {
	allowedKeys := []string{"allowed-key"}
	blockedKeys := []string{"blocked-key"}

	limiter := NewMockRateLimiterWithLimits(allowedKeys, blockedKeys)
	ctx := context.Background()

	// Test allowed key
	allowed, err := limiter.Allow(ctx, "allowed-key")
	assert.NoError(t, err)
	assert.True(t, allowed)

	// Test blocked key
	blocked, err := limiter.Allow(ctx, "blocked-key")
	assert.NoError(t, err)
	assert.False(t, blocked)
}
