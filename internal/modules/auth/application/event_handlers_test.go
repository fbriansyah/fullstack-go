package application

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"go-templ-template/internal/modules/auth/domain"
	userDomain "go-templ-template/internal/modules/user/domain"
	"go-templ-template/internal/shared/audit"
	"go-templ-template/internal/shared/database"
	"go-templ-template/internal/shared/events"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAuthService is a mock implementation of AuthService for testing
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(ctx context.Context, cmd *LoginCommand) (*AuthResult, error) {
	args := m.Called(ctx, cmd)
	return args.Get(0).(*AuthResult), args.Error(1)
}

func (m *MockAuthService) Register(ctx context.Context, cmd *RegisterCommand) (*AuthResult, error) {
	args := m.Called(ctx, cmd)
	return args.Get(0).(*AuthResult), args.Error(1)
}

func (m *MockAuthService) Logout(ctx context.Context, cmd *LogoutCommand) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *MockAuthService) ValidateSession(ctx context.Context, query *ValidateSessionQuery) (*SessionValidationResult, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(*SessionValidationResult), args.Error(1)
}

func (m *MockAuthService) RefreshSession(ctx context.Context, cmd *RefreshSessionCommand) (*domain.Session, error) {
	args := m.Called(ctx, cmd)
	return args.Get(0).(*domain.Session), args.Error(1)
}

func (m *MockAuthService) ChangePassword(ctx context.Context, cmd *ChangePasswordCommand) error {
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

// MockAuditLogger is a mock implementation of AuditLogger for testing
type MockAuditLogger struct {
	mock.Mock
}

func (m *MockAuditLogger) LogEvent(ctx context.Context, event *audit.AuditEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockAuditLogger) GetEvents(ctx context.Context, filter *audit.AuditFilter) ([]*audit.AuditEvent, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*audit.AuditEvent), args.Error(1)
}

// TestUserCreatedEventHandler tests the UserCreatedEventHandler
func TestUserCreatedEventHandler(t *testing.T) {
	// Setup
	mockAuthService := new(MockAuthService)
	mockAuditLogger := new(MockAuditLogger)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := NewUserCreatedEventHandler(mockAuthService, logger, mockAuditLogger)

	// Create test event
	userCreatedEvent := userDomain.NewUserCreatedEvent(&userDomain.User{
		ID:        "user-123",
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Status:    userDomain.UserStatusActive,
	})

	ctx := context.Background()

	// Setup expectations
	mockAuditLogger.On("LogEvent", ctx, mock.MatchedBy(func(event *audit.AuditEvent) bool {
		return event.EventType == "user.lifecycle.created" &&
			event.UserID == "user-123" &&
			event.Action == "user_created" &&
			event.Resource == "user"
	})).Return(nil)

	// Execute
	err := handler.Handle(ctx, userCreatedEvent)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "user.created", handler.EventType())
	assert.Equal(t, "auth.user_created_handler", handler.HandlerName())

	// Verify expectations
	mockAuditLogger.AssertExpectations(t)
}

// TestUserUpdatedEventHandler tests the UserUpdatedEventHandler
func TestUserUpdatedEventHandler(t *testing.T) {
	// Setup
	mockAuthService := new(MockAuthService)
	mockAuditLogger := new(MockAuditLogger)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := NewUserUpdatedEventHandler(mockAuthService, logger, mockAuditLogger)

	// Create test user and event
	user := &userDomain.User{
		ID:        "user-123",
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Status:    userDomain.UserStatusActive,
		Version:   2,
	}

	changes := map[string]interface{}{
		"first_name": map[string]interface{}{
			"old": "John",
			"new": "Johnny",
		},
	}

	userUpdatedEvent := userDomain.NewUserUpdatedEvent(user, changes)

	ctx := context.Background()

	// Setup expectations
	mockAuditLogger.On("LogEvent", ctx, mock.MatchedBy(func(event *audit.AuditEvent) bool {
		return event.EventType == "user.lifecycle.updated" &&
			event.UserID == "user-123" &&
			event.Action == "user_updated" &&
			event.Resource == "user"
	})).Return(nil)

	// Execute
	err := handler.Handle(ctx, userUpdatedEvent)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "user.updated", handler.EventType())
	assert.Equal(t, "auth.user_updated_handler", handler.HandlerName())

	// Verify expectations
	mockAuditLogger.AssertExpectations(t)
}

// TestUserDeletedEventHandler tests the UserDeletedEventHandler
func TestUserDeletedEventHandler(t *testing.T) {
	// Setup
	mockAuthService := new(MockAuthService)
	mockAuditLogger := new(MockAuditLogger)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := NewUserDeletedEventHandler(mockAuthService, logger, mockAuditLogger)

	// Create test user and event
	user := &userDomain.User{
		ID:        "user-123",
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Status:    userDomain.UserStatusActive,
	}

	userDeletedEvent := userDomain.NewUserDeletedEvent(user, "admin-456", "account_closure")

	ctx := context.Background()

	// Setup expectations
	mockAuditLogger.On("LogEvent", ctx, mock.MatchedBy(func(event *audit.AuditEvent) bool {
		return event.EventType == "user.lifecycle.deleted" &&
			event.UserID == "user-123" &&
			event.Action == "user_deleted" &&
			event.Resource == "user"
	})).Return(nil)

	// Execute
	err := handler.Handle(ctx, userDeletedEvent)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "user.deleted", handler.EventType())
	assert.Equal(t, "auth.user_deleted_handler", handler.HandlerName())

	// Verify expectations
	mockAuditLogger.AssertExpectations(t)
}

// TestEventHandlersWithInvalidEventData tests handlers with invalid event data
func TestEventHandlersWithInvalidEventData(t *testing.T) {
	// Setup
	mockAuthService := new(MockAuthService)
	mockAuditLogger := new(MockAuditLogger)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create handlers
	userCreatedHandler := NewUserCreatedEventHandler(mockAuthService, logger, mockAuditLogger)
	userUpdatedHandler := NewUserUpdatedEventHandler(mockAuthService, logger, mockAuditLogger)
	userDeletedHandler := NewUserDeletedEventHandler(mockAuthService, logger, mockAuditLogger)

	// Create invalid event (with string data instead of map)
	invalidEvent := &mockDomainEvent{
		eventType:     "user.created",
		aggregateID:   "user-123",
		aggregateType: "user",
		occurredAt:    time.Now(),
		eventData:     "invalid-data", // This should be a map
		eventID:       "event-123",
		version:       1,
		metadata:      events.EventMetadata{},
	}

	ctx := context.Background()

	// Test UserCreatedEventHandler with invalid data
	err := userCreatedHandler.Handle(ctx, invalidEvent)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid event data format")

	// Test UserUpdatedEventHandler with invalid data
	invalidEvent.eventType = "user.updated"
	err = userUpdatedHandler.Handle(ctx, invalidEvent)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid event data format")

	// Test UserDeletedEventHandler with invalid data
	invalidEvent.eventType = "user.deleted"
	err = userDeletedHandler.Handle(ctx, invalidEvent)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid event data format")
}

// mockDomainEvent is a mock implementation of DomainEvent for testing
type mockDomainEvent struct {
	eventType     string
	aggregateID   string
	aggregateType string
	occurredAt    time.Time
	eventData     interface{}
	eventID       string
	version       int
	metadata      events.EventMetadata
}

func (m *mockDomainEvent) EventType() string {
	return m.eventType
}

func (m *mockDomainEvent) AggregateID() string {
	return m.aggregateID
}

func (m *mockDomainEvent) AggregateType() string {
	return m.aggregateType
}

func (m *mockDomainEvent) OccurredAt() time.Time {
	return m.occurredAt
}

func (m *mockDomainEvent) EventData() interface{} {
	return m.eventData
}

func (m *mockDomainEvent) EventID() string {
	return m.eventID
}

func (m *mockDomainEvent) Version() int {
	return m.version
}

func (m *mockDomainEvent) Metadata() events.EventMetadata {
	return m.metadata
}

// Integration test for audit logger
func TestAuditLoggerIntegration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Skip if no database available
	database.SkipIfNoDatabase(t)

	// Setup test database
	testDB := database.NewTestDatabase(t)
	defer testDB.Close()

	// Create audit events table
	err := audit.CreateAuditEventsTable(context.Background(), testDB.DB.DB)
	require.NoError(t, err)

	// Clean up any existing audit events
	testDB.TruncateTable("audit_events")

	// Create audit logger
	auditLogger := audit.NewAuditLogger(testDB.DB)

	ctx := context.Background()

	// Create test audit event with unique identifiers
	uniqueEventID := fmt.Sprintf("event-test-%d", time.Now().UnixNano())
	uniqueUserID := fmt.Sprintf("user-test-%d", time.Now().UnixNano())

	auditEvent := &audit.AuditEvent{
		EventID:       uniqueEventID,
		EventType:     "user.lifecycle.created",
		AggregateID:   uniqueUserID,
		AggregateType: "user",
		UserID:        uniqueUserID,
		Action:        "user_created",
		Resource:      "user",
		ResourceID:    uniqueUserID,
		Details: map[string]interface{}{
			"email":      "test@example.com",
			"first_name": "John",
			"last_name":  "Doe",
		},
		OccurredAt: time.Now().UTC(),
		Metadata: events.EventMetadata{
			Source: "user-module",
		},
	}

	// Log the event
	err = auditLogger.LogEvent(ctx, auditEvent)
	require.NoError(t, err)

	// Query the events with specific filters to avoid conflicts
	filter := &audit.AuditFilter{
		EventID: uniqueEventID,
		UserID:  uniqueUserID,
		Action:  "user_created",
		Limit:   10,
	}

	events, err := auditLogger.GetEvents(ctx, filter)
	require.NoError(t, err)
	require.Len(t, events, 1)

	// Verify the retrieved event
	retrievedEvent := events[0]
	assert.Equal(t, auditEvent.EventID, retrievedEvent.EventID)
	assert.Equal(t, auditEvent.EventType, retrievedEvent.EventType)
	assert.Equal(t, auditEvent.UserID, retrievedEvent.UserID)
	assert.Equal(t, auditEvent.Action, retrievedEvent.Action)
	assert.Equal(t, auditEvent.Resource, retrievedEvent.Resource)
	assert.Equal(t, auditEvent.Details["email"], retrievedEvent.Details["email"])
}

// TestEventHandlerRegistration tests that event handlers are properly registered
func TestEventHandlerRegistration(t *testing.T) {
	// This test would typically be part of the module integration tests
	// but we can test the handler creation here

	mockAuthService := new(MockAuthService)
	mockAuditLogger := new(MockAuditLogger)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Test handler creation
	userCreatedHandler := NewUserCreatedEventHandler(mockAuthService, logger, mockAuditLogger)
	assert.NotNil(t, userCreatedHandler)
	assert.Equal(t, "user.created", userCreatedHandler.EventType())
	assert.Equal(t, "auth.user_created_handler", userCreatedHandler.HandlerName())

	userUpdatedHandler := NewUserUpdatedEventHandler(mockAuthService, logger, mockAuditLogger)
	assert.NotNil(t, userUpdatedHandler)
	assert.Equal(t, "user.updated", userUpdatedHandler.EventType())
	assert.Equal(t, "auth.user_updated_handler", userUpdatedHandler.HandlerName())

	userDeletedHandler := NewUserDeletedEventHandler(mockAuthService, logger, mockAuditLogger)
	assert.NotNil(t, userDeletedHandler)
	assert.Equal(t, "user.deleted", userDeletedHandler.EventType())
	assert.Equal(t, "auth.user_deleted_handler", userDeletedHandler.HandlerName())
}
