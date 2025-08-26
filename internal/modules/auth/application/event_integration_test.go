package application

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	userDomain "go-templ-template/internal/modules/user/domain"
	"go-templ-template/internal/shared/audit"
	"go-templ-template/internal/shared/database"
	"go-templ-template/internal/shared/events"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUserLifecycleEventFlow tests the complete flow of user lifecycle events
func TestUserLifecycleEventFlow(t *testing.T) {
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

	// Setup components
	auditLogger := audit.NewAuditLogger(testDB.DB)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockAuthService := new(MockAuthService)

	// Create event handlers
	userCreatedHandler := NewUserCreatedEventHandler(mockAuthService, logger, auditLogger)
	userUpdatedHandler := NewUserUpdatedEventHandler(mockAuthService, logger, auditLogger)
	userDeletedHandler := NewUserDeletedEventHandler(mockAuthService, logger, auditLogger)

	ctx := context.Background()

	// Test 1: User Created Event
	t.Run("UserCreatedEvent", func(t *testing.T) {
		user := &userDomain.User{
			ID:        "user-123",
			Email:     "test@example.com",
			FirstName: "John",
			LastName:  "Doe",
			Status:    userDomain.UserStatusActive,
		}

		userCreatedEvent := userDomain.NewUserCreatedEvent(user)

		// Handle the event
		err := userCreatedHandler.Handle(ctx, userCreatedEvent)
		require.NoError(t, err)

		// Verify audit log entry was created
		filter := &audit.AuditFilter{
			UserID:    "user-123",
			Action:    "user_created",
			EventType: "user.lifecycle.created",
			Limit:     1,
		}

		auditEvents, err := auditLogger.GetEvents(ctx, filter)
		require.NoError(t, err)
		require.Len(t, auditEvents, 1)

		auditEvent := auditEvents[0]
		assert.Equal(t, "user-123", auditEvent.UserID)
		assert.Equal(t, "user_created", auditEvent.Action)
		assert.Equal(t, "user", auditEvent.Resource)
		assert.Equal(t, "test@example.com", auditEvent.Details["email"])
		assert.Equal(t, "John", auditEvent.Details["first_name"])
		assert.Equal(t, "Doe", auditEvent.Details["last_name"])
	})

	// Test 2: User Updated Event
	t.Run("UserUpdatedEvent", func(t *testing.T) {
		user := &userDomain.User{
			ID:        "user-456",
			Email:     "updated@example.com",
			FirstName: "Jane",
			LastName:  "Smith",
			Status:    userDomain.UserStatusActive,
			Version:   2,
		}

		changes := map[string]interface{}{
			"first_name": map[string]interface{}{
				"old": "Jane",
				"new": "Janet",
			},
			"email": map[string]interface{}{
				"old": "jane@example.com",
				"new": "updated@example.com",
			},
		}

		userUpdatedEvent := userDomain.NewUserUpdatedEvent(user, changes)

		// Handle the event
		err := userUpdatedHandler.Handle(ctx, userUpdatedEvent)
		require.NoError(t, err)

		// Verify audit log entry was created
		filter := &audit.AuditFilter{
			UserID:    "user-456",
			Action:    "user_updated",
			EventType: "user.lifecycle.updated",
			Limit:     1,
		}

		auditEvents, err := auditLogger.GetEvents(ctx, filter)
		require.NoError(t, err)
		require.Len(t, auditEvents, 1)

		auditEvent := auditEvents[0]
		assert.Equal(t, "user-456", auditEvent.UserID)
		assert.Equal(t, "user_updated", auditEvent.Action)
		assert.Equal(t, "user", auditEvent.Resource)
		assert.NotNil(t, auditEvent.Details["changes"])
	})

	// Test 3: User Deleted Event
	t.Run("UserDeletedEvent", func(t *testing.T) {
		user := &userDomain.User{
			ID:        "user-789",
			Email:     "deleted@example.com",
			FirstName: "Bob",
			LastName:  "Johnson",
			Status:    userDomain.UserStatusActive,
		}

		userDeletedEvent := userDomain.NewUserDeletedEvent(user, "admin-123", "account_closure")

		// Handle the event
		err := userDeletedHandler.Handle(ctx, userDeletedEvent)
		require.NoError(t, err)

		// Verify audit log entry was created
		filter := &audit.AuditFilter{
			UserID:    "user-789",
			Action:    "user_deleted",
			EventType: "user.lifecycle.deleted",
			Limit:     1,
		}

		auditEvents, err := auditLogger.GetEvents(ctx, filter)
		require.NoError(t, err)
		require.Len(t, auditEvents, 1)

		auditEvent := auditEvents[0]
		assert.Equal(t, "user-789", auditEvent.UserID)
		assert.Equal(t, "user_deleted", auditEvent.Action)
		assert.Equal(t, "user", auditEvent.Resource)
		assert.Equal(t, "deleted@example.com", auditEvent.Details["email"])
		assert.Equal(t, "admin-123", auditEvent.Details["deleted_by"])
		assert.Equal(t, "account_closure", auditEvent.Details["reason"])
	})

	// Test 4: Query audit events with various filters
	t.Run("AuditEventQueries", func(t *testing.T) {
		// Query all events
		allEventsFilter := &audit.AuditFilter{
			Limit: 10,
		}

		allEvents, err := auditLogger.GetEvents(ctx, allEventsFilter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(allEvents), 3) // At least the 3 events we created

		// Query events by action
		createdEventsFilter := &audit.AuditFilter{
			Action: "user_created",
			Limit:  10,
		}

		createdEvents, err := auditLogger.GetEvents(ctx, createdEventsFilter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(createdEvents), 1)

		// Query events by resource
		userResourceFilter := &audit.AuditFilter{
			Resource: "user",
			Limit:    10,
		}

		userEvents, err := auditLogger.GetEvents(ctx, userResourceFilter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(userEvents), 3)

		// Query events by time range
		now := time.Now()
		oneHourAgo := now.Add(-1 * time.Hour)

		timeRangeFilter := &audit.AuditFilter{
			StartTime: oneHourAgo,
			EndTime:   now,
			Limit:     10,
		}

		timeRangeEvents, err := auditLogger.GetEvents(ctx, timeRangeFilter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(timeRangeEvents), 3)
	})
}

// TestEventHandlerErrorHandling tests error handling in event handlers
func TestEventHandlerErrorHandling(t *testing.T) {
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

	// Setup components
	auditLogger := audit.NewAuditLogger(testDB.DB)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockAuthService := new(MockAuthService)

	// Create event handlers
	userCreatedHandler := NewUserCreatedEventHandler(mockAuthService, logger, auditLogger)

	ctx := context.Background()

	// Test with invalid event data
	t.Run("InvalidEventData", func(t *testing.T) {
		invalidEvent := &mockDomainEvent{
			eventType:     "user.created",
			aggregateID:   "user-123",
			aggregateType: "user",
			occurredAt:    time.Now(),
			eventData:     "invalid-string-data", // Should be a map
			eventID:       "event-123",
			version:       1,
			metadata:      events.EventMetadata{},
		}

		err := userCreatedHandler.Handle(ctx, invalidEvent)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid event data format")
	})

	// Test with nil event data
	t.Run("NilEventData", func(t *testing.T) {
		invalidEvent := &mockDomainEvent{
			eventType:     "user.created",
			aggregateID:   "user-123",
			aggregateType: "user",
			occurredAt:    time.Now(),
			eventData:     nil,
			eventID:       "event-123",
			version:       1,
			metadata:      events.EventMetadata{},
		}

		err := userCreatedHandler.Handle(ctx, invalidEvent)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid event data format")
	})
}

// TestConcurrentEventHandling tests handling multiple events concurrently
func TestConcurrentEventHandling(t *testing.T) {
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

	// Setup components
	auditLogger := audit.NewAuditLogger(testDB.DB)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockAuthService := new(MockAuthService)

	// Create event handler
	userCreatedHandler := NewUserCreatedEventHandler(mockAuthService, logger, auditLogger)

	ctx := context.Background()

	// Create multiple events concurrently
	numEvents := 10
	eventChan := make(chan error, numEvents)

	for i := 0; i < numEvents; i++ {
		go func(index int) {
			user := &userDomain.User{
				ID:        fmt.Sprintf("user-%d", index),
				Email:     fmt.Sprintf("test%d@example.com", index),
				FirstName: "John",
				LastName:  "Doe",
				Status:    userDomain.UserStatusActive,
			}

			userCreatedEvent := userDomain.NewUserCreatedEvent(user)
			err := userCreatedHandler.Handle(ctx, userCreatedEvent)
			eventChan <- err
		}(i)
	}

	// Wait for all events to complete
	for i := 0; i < numEvents; i++ {
		err := <-eventChan
		assert.NoError(t, err)
	}

	// Verify all events were logged
	filter := &audit.AuditFilter{
		Action: "user_created",
		Limit:  numEvents + 10, // Add buffer for other tests
	}

	auditEvents, err := auditLogger.GetEvents(ctx, filter)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(auditEvents), numEvents)
}
