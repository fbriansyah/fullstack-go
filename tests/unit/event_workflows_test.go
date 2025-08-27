package unit

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	userApp "go-templ-template/internal/modules/user/application"
	"go-templ-template/internal/modules/user/domain"
	"go-templ-template/internal/shared/audit"
	"go-templ-template/internal/shared/events"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockEventBus is a mock implementation of EventBus for testing
type MockEventBus struct {
	mock.Mock
	publishedEvents []events.DomainEvent
}

func (m *MockEventBus) Publish(ctx context.Context, event events.DomainEvent) error {
	args := m.Called(ctx, event)
	m.publishedEvents = append(m.publishedEvents, event)
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

func (m *MockEventBus) GetPublishedEvents() []events.DomainEvent {
	return m.publishedEvents
}

func (m *MockEventBus) ClearPublishedEvents() {
	m.publishedEvents = nil
}

// MockAuditLogger is a mock implementation of AuditLogger for testing
type MockAuditLogger struct {
	mock.Mock
	loggedEvents []*audit.AuditEvent
}

func (m *MockAuditLogger) LogEvent(ctx context.Context, event *audit.AuditEvent) error {
	args := m.Called(ctx, event)
	m.loggedEvents = append(m.loggedEvents, event)
	return args.Error(0)
}

func (m *MockAuditLogger) GetEvents(ctx context.Context, filter *audit.AuditFilter) ([]*audit.AuditEvent, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*audit.AuditEvent), args.Error(1)
}

func (m *MockAuditLogger) GetLoggedEvents() []*audit.AuditEvent {
	return m.loggedEvents
}

func (m *MockAuditLogger) ClearLoggedEvents() {
	m.loggedEvents = nil
}

// MockActivationTokenRepository is a mock implementation for testing
type MockActivationTokenRepository struct {
	mock.Mock
	tokens map[string]*domain.ActivationToken
}

func NewMockActivationTokenRepository() *MockActivationTokenRepository {
	return &MockActivationTokenRepository{
		tokens: make(map[string]*domain.ActivationToken),
	}
}

func (m *MockActivationTokenRepository) Create(ctx context.Context, token *domain.ActivationToken) error {
	args := m.Called(ctx, token)
	if args.Error(0) == nil {
		m.tokens[token.Token] = token
	}
	return args.Error(0)
}

func (m *MockActivationTokenRepository) GetByToken(ctx context.Context, token string) (*domain.ActivationToken, error) {
	args := m.Called(ctx, token)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}
	if t, exists := m.tokens[token]; exists {
		return t, nil
	}
	return args.Get(0).(*domain.ActivationToken), args.Error(1)
}

func (m *MockActivationTokenRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.ActivationToken, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*domain.ActivationToken), args.Error(1)
}

func (m *MockActivationTokenRepository) Update(ctx context.Context, token *domain.ActivationToken) error {
	args := m.Called(ctx, token)
	if args.Error(0) == nil {
		m.tokens[token.Token] = token
	}
	return args.Error(0)
}

func (m *MockActivationTokenRepository) Delete(ctx context.Context, tokenID string) error {
	args := m.Called(ctx, tokenID)
	return args.Error(0)
}

func (m *MockActivationTokenRepository) DeleteByUserID(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockActivationTokenRepository) DeleteExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// TestUserActivationWorkflow tests the user activation workflow
func TestUserActivationWorkflow(t *testing.T) {
	// Test user activation requested event creation
	t.Run("UserActivationRequestedEvent", func(t *testing.T) {
		user, err := domain.NewUser("user-123", "test@example.com", "Password123!", "Test", "User")
		require.NoError(t, err)

		token := "activation-token-123"
		expiresAt := time.Now().Add(24 * time.Hour)
		requestedBy := "admin-user"

		event := domain.NewUserActivationRequestedEvent(user, token, expiresAt, requestedBy)

		assert.Equal(t, "user.activation_requested", event.EventType())
		assert.Equal(t, user.ID, event.AggregateID())
		assert.Equal(t, "user", event.AggregateType())
		assert.Equal(t, user.ID, event.UserID)
		assert.Equal(t, user.Email, event.Email)
		assert.Equal(t, token, event.ActivationToken)
		assert.Equal(t, expiresAt, event.ExpiresAt)
		assert.Equal(t, requestedBy, event.RequestedBy)

		// Test event data
		eventData := event.EventData().(map[string]interface{})
		assert.Equal(t, user.ID, eventData["user_id"])
		assert.Equal(t, user.Email, eventData["email"])
		assert.Equal(t, token, eventData["activation_token"])
		assert.Equal(t, expiresAt, eventData["expires_at"])
		assert.Equal(t, requestedBy, eventData["requested_by"])
	})

	// Test user activated event creation
	t.Run("UserActivatedEvent", func(t *testing.T) {
		user, err := domain.NewUser("user-123", "test@example.com", "Password123!", "Test", "User")
		require.NoError(t, err)

		activatedBy := "admin-user"
		method := "token"

		event := domain.NewUserActivatedEvent(user, activatedBy, method)

		assert.Equal(t, "user.activated", event.EventType())
		assert.Equal(t, user.ID, event.AggregateID())
		assert.Equal(t, "user", event.AggregateType())
		assert.Equal(t, user.ID, event.UserID)
		assert.Equal(t, user.Email, event.Email)
		assert.Equal(t, activatedBy, event.ActivatedBy)
		assert.Equal(t, method, event.Method)
		assert.WithinDuration(t, time.Now(), event.ActivatedAt, time.Second)

		// Test event data
		eventData := event.EventData().(map[string]interface{})
		assert.Equal(t, user.ID, eventData["user_id"])
		assert.Equal(t, user.Email, eventData["email"])
		assert.Equal(t, activatedBy, eventData["activated_by"])
		assert.Equal(t, method, eventData["method"])
	})

	// Test user deactivated event creation
	t.Run("UserDeactivatedEvent", func(t *testing.T) {
		user, err := domain.NewUser("user-123", "test@example.com", "Password123!", "Test", "User")
		require.NoError(t, err)

		deactivatedBy := "admin-user"
		reason := "policy violation"

		event := domain.NewUserDeactivatedEvent(user, deactivatedBy, reason)

		assert.Equal(t, "user.deactivated", event.EventType())
		assert.Equal(t, user.ID, event.AggregateID())
		assert.Equal(t, "user", event.AggregateType())
		assert.Equal(t, user.ID, event.UserID)
		assert.Equal(t, user.Email, event.Email)
		assert.Equal(t, deactivatedBy, event.DeactivatedBy)
		assert.Equal(t, reason, event.Reason)
		assert.WithinDuration(t, time.Now(), event.DeactivatedAt, time.Second)

		// Test event data
		eventData := event.EventData().(map[string]interface{})
		assert.Equal(t, user.ID, eventData["user_id"])
		assert.Equal(t, user.Email, eventData["email"])
		assert.Equal(t, deactivatedBy, eventData["deactivated_by"])
		assert.Equal(t, reason, eventData["reason"])
	})

	// Test activation token expiration event
	t.Run("UserActivationTokenExpiredEvent", func(t *testing.T) {
		userID := "user-123"
		email := "test@example.com"
		token := "expired-token-123"

		event := domain.NewUserActivationTokenExpiredEvent(userID, email, token)

		assert.Equal(t, "user.activation_token_expired", event.EventType())
		assert.Equal(t, userID, event.AggregateID())
		assert.Equal(t, "user", event.AggregateType())
		assert.Equal(t, userID, event.UserID)
		assert.Equal(t, email, event.Email)
		assert.Equal(t, token, event.ActivationToken)
		assert.WithinDuration(t, time.Now(), event.ExpiredAt, time.Second)

		// Test event data
		eventData := event.EventData().(map[string]interface{})
		assert.Equal(t, userID, eventData["user_id"])
		assert.Equal(t, email, eventData["email"])
		assert.Equal(t, token, eventData["activation_token"])
	})
}

// TestActivationTokenValidation tests activation token validation logic
func TestActivationTokenValidation(t *testing.T) {
	t.Run("ValidToken", func(t *testing.T) {
		token := &domain.ActivationToken{
			ID:        "token-123",
			UserID:    "user-123",
			Token:     "valid-token",
			ExpiresAt: time.Now().Add(1 * time.Hour),
			UsedAt:    nil,
			CreatedAt: time.Now(),
		}

		assert.False(t, token.IsExpired())
		assert.False(t, token.IsUsed())
		assert.True(t, token.IsValid())
	})

	t.Run("ExpiredToken", func(t *testing.T) {
		token := &domain.ActivationToken{
			ID:        "token-123",
			UserID:    "user-123",
			Token:     "expired-token",
			ExpiresAt: time.Now().Add(-1 * time.Hour),
			UsedAt:    nil,
			CreatedAt: time.Now().Add(-2 * time.Hour),
		}

		assert.True(t, token.IsExpired())
		assert.False(t, token.IsUsed())
		assert.False(t, token.IsValid())
	})

	t.Run("UsedToken", func(t *testing.T) {
		usedAt := time.Now().Add(-30 * time.Minute)
		token := &domain.ActivationToken{
			ID:        "token-123",
			UserID:    "user-123",
			Token:     "used-token",
			ExpiresAt: time.Now().Add(1 * time.Hour),
			UsedAt:    &usedAt,
			CreatedAt: time.Now().Add(-1 * time.Hour),
		}

		assert.False(t, token.IsExpired())
		assert.True(t, token.IsUsed())
		assert.False(t, token.IsValid())
	})
}

// TestEventAuditHandler tests the event audit handler
func TestEventAuditHandler(t *testing.T) {
	t.Run("HandleUserCreatedEvent", func(t *testing.T) {
		mockAuditLogger := &MockAuditLogger{}
		mockAuditLogger.On("LogEvent", mock.Anything, mock.Anything).Return(nil)

		// Create a no-op logger for testing
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		handler := audit.NewEventAuditHandler(mockAuditLogger, logger)

		// Create a user created event
		user, err := domain.NewUser("user-123", "test@example.com", "Password123!", "Test", "User")
		require.NoError(t, err)

		event := domain.NewUserCreatedEvent(user)

		// Handle the event
		err = handler.Handle(context.Background(), event)
		require.NoError(t, err)

		// Verify audit logger was called
		mockAuditLogger.AssertCalled(t, "LogEvent", mock.Anything, mock.MatchedBy(func(auditEvent *audit.AuditEvent) bool {
			return auditEvent.Action == "user_created" &&
				auditEvent.Resource == "user" &&
				auditEvent.AggregateID == user.ID
		}))
	})

	t.Run("HandleUserActivatedEvent", func(t *testing.T) {
		mockAuditLogger := &MockAuditLogger{}
		mockAuditLogger.On("LogEvent", mock.Anything, mock.Anything).Return(nil)

		// Create a no-op logger for testing
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		handler := audit.NewEventAuditHandler(mockAuditLogger, logger)

		// Create a user activated event
		user, err := domain.NewUser("user-123", "test@example.com", "Password123!", "Test", "User")
		require.NoError(t, err)

		event := domain.NewUserActivatedEvent(user, "admin", "token")

		// Handle the event
		err = handler.Handle(context.Background(), event)
		require.NoError(t, err)

		// Verify audit logger was called
		mockAuditLogger.AssertCalled(t, "LogEvent", mock.Anything, mock.MatchedBy(func(auditEvent *audit.AuditEvent) bool {
			return auditEvent.Action == "user_activated" &&
				auditEvent.Resource == "user" &&
				auditEvent.AggregateID == user.ID
		}))
	})
}

// TestActivationCommands tests activation command validation
func TestActivationCommands(t *testing.T) {
	t.Run("RequestActivationCommand", func(t *testing.T) {
		// Valid command
		cmd := &userApp.RequestActivationCommand{
			UserID:      "user-123",
			RequestedBy: "admin",
		}
		assert.NoError(t, cmd.Validate())

		// Invalid command - empty user ID
		cmd = &userApp.RequestActivationCommand{
			UserID:      "",
			RequestedBy: "admin",
		}
		assert.Error(t, cmd.Validate())
	})

	t.Run("ActivateUserCommand", func(t *testing.T) {
		// Valid command
		cmd := &userApp.ActivateUserCommand{
			Token:       "valid-token-123",
			ActivatedBy: "admin",
		}
		assert.NoError(t, cmd.Validate())

		// Invalid command - empty token
		cmd = &userApp.ActivateUserCommand{
			Token:       "",
			ActivatedBy: "admin",
		}
		assert.Error(t, cmd.Validate())

		// Invalid command - short token
		cmd = &userApp.ActivateUserCommand{
			Token:       "short",
			ActivatedBy: "admin",
		}
		assert.Error(t, cmd.Validate())
	})

	t.Run("DeactivateUserCommand", func(t *testing.T) {
		// Valid command
		cmd := &userApp.DeactivateUserCommand{
			UserID:        "user-123",
			Version:       1,
			DeactivatedBy: "admin",
			Reason:        "policy violation",
		}
		assert.NoError(t, cmd.Validate())

		// Invalid command - empty user ID
		cmd = &userApp.DeactivateUserCommand{
			UserID:        "",
			Version:       1,
			DeactivatedBy: "admin",
			Reason:        "policy violation",
		}
		assert.Error(t, cmd.Validate())

		// Invalid command - invalid version
		cmd = &userApp.DeactivateUserCommand{
			UserID:        "user-123",
			Version:       0,
			DeactivatedBy: "admin",
			Reason:        "policy violation",
		}
		assert.Error(t, cmd.Validate())
	})
}
