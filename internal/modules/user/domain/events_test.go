package domain

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseEvent_Methods(t *testing.T) {
	now := time.Now().UTC()
	event := BaseEvent{
		ID:           "event-123",
		Type:         "test.event",
		AggregateId:  "aggregate-123",
		OccurredOn:   now,
		EventVersion: 1,
	}

	assert.Equal(t, "test.event", event.EventType())
	assert.Equal(t, "aggregate-123", event.AggregateID())
	assert.Equal(t, now, event.OccurredAt())
}

func TestNewUserCreatedEvent(t *testing.T) {
	user, err := NewUser("user-123", "john.doe@example.com", "Password123", "John", "Doe")
	require.NoError(t, err)

	event := NewUserCreatedEvent(user)

	assert.NotNil(t, event)
	assert.Equal(t, "user.created", event.EventType())
	assert.Equal(t, user.ID, event.AggregateID())
	assert.Equal(t, user.ID, event.UserID)
	assert.Equal(t, user.Email, event.Email)
	assert.Equal(t, user.FirstName, event.FirstName)
	assert.Equal(t, user.LastName, event.LastName)
	assert.Equal(t, string(user.Status), event.Status)
	assert.NotEmpty(t, event.BaseEvent.ID)
	assert.Equal(t, 1, event.BaseEvent.Version)
	assert.True(t, time.Since(event.OccurredAt()) < time.Second)

	// Test EventData
	data := event.EventData()
	dataMap, ok := data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, user.ID, dataMap["user_id"])
	assert.Equal(t, user.Email, dataMap["email"])
	assert.Equal(t, user.FirstName, dataMap["first_name"])
	assert.Equal(t, user.LastName, dataMap["last_name"])
	assert.Equal(t, string(user.Status), dataMap["status"])
}

func TestNewUserUpdatedEvent(t *testing.T) {
	user, err := NewUser("user-123", "john.doe@example.com", "Password123", "John", "Doe")
	require.NoError(t, err)

	// Simulate an update
	user.Version = 2
	changes := map[string]interface{}{
		"first_name": "Jane",
		"last_name":  "Smith",
	}

	event := NewUserUpdatedEvent(user, changes)

	assert.NotNil(t, event)
	assert.Equal(t, "user.updated", event.EventType())
	assert.Equal(t, user.ID, event.AggregateID())
	assert.Equal(t, user.ID, event.UserID)
	assert.Equal(t, changes, event.Changes)
	assert.Equal(t, 1, event.PreviousVersion) // user.Version - 1
	assert.NotEmpty(t, event.BaseEvent.ID)
	assert.True(t, time.Since(event.OccurredAt()) < time.Second)

	// Test EventData
	data := event.EventData()
	dataMap, ok := data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, user.ID, dataMap["user_id"])
	assert.Equal(t, changes, dataMap["changes"])
	assert.Equal(t, 1, dataMap["previous_version"])
}

func TestNewUserDeletedEvent(t *testing.T) {
	user, err := NewUser("user-123", "john.doe@example.com", "Password123", "John", "Doe")
	require.NoError(t, err)

	deletedBy := "admin-456"
	reason := "Account violation"

	event := NewUserDeletedEvent(user, deletedBy, reason)

	assert.NotNil(t, event)
	assert.Equal(t, "user.deleted", event.EventType())
	assert.Equal(t, user.ID, event.AggregateID())
	assert.Equal(t, user.ID, event.UserID)
	assert.Equal(t, user.Email, event.Email)
	assert.Equal(t, deletedBy, event.DeletedBy)
	assert.Equal(t, reason, event.Reason)
	assert.NotEmpty(t, event.BaseEvent.ID)
	assert.True(t, time.Since(event.OccurredAt()) < time.Second)

	// Test EventData
	data := event.EventData()
	dataMap, ok := data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, user.ID, dataMap["user_id"])
	assert.Equal(t, user.Email, dataMap["email"])
	assert.Equal(t, deletedBy, dataMap["deleted_by"])
	assert.Equal(t, reason, dataMap["reason"])
}

func TestNewUserDeletedEvent_WithoutOptionalFields(t *testing.T) {
	user, err := NewUser("user-123", "john.doe@example.com", "Password123", "John", "Doe")
	require.NoError(t, err)

	event := NewUserDeletedEvent(user, "", "")

	assert.NotNil(t, event)
	assert.Equal(t, "user.deleted", event.EventType())
	assert.Equal(t, user.ID, event.UserID)
	assert.Equal(t, user.Email, event.Email)
	assert.Empty(t, event.DeletedBy)
	assert.Empty(t, event.Reason)

	// Test EventData
	data := event.EventData()
	dataMap, ok := data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "", dataMap["deleted_by"])
	assert.Equal(t, "", dataMap["reason"])
}

func TestNewUserStatusChangedEvent(t *testing.T) {
	user, err := NewUser("user-123", "john.doe@example.com", "Password123", "John", "Doe")
	require.NoError(t, err)

	// Change status from active to suspended
	previousStatus := user.Status
	user.Status = UserStatusSuspended
	changedBy := "admin-456"
	reason := "Policy violation"

	event := NewUserStatusChangedEvent(user, previousStatus, changedBy, reason)

	assert.NotNil(t, event)
	assert.Equal(t, "user.status_changed", event.EventType())
	assert.Equal(t, user.ID, event.AggregateID())
	assert.Equal(t, user.ID, event.UserID)
	assert.Equal(t, string(previousStatus), event.PreviousStatus)
	assert.Equal(t, string(user.Status), event.NewStatus)
	assert.Equal(t, changedBy, event.ChangedBy)
	assert.Equal(t, reason, event.Reason)
	assert.NotEmpty(t, event.BaseEvent.ID)
	assert.True(t, time.Since(event.OccurredAt()) < time.Second)

	// Test EventData
	data := event.EventData()
	dataMap, ok := data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, user.ID, dataMap["user_id"])
	assert.Equal(t, string(previousStatus), dataMap["previous_status"])
	assert.Equal(t, string(user.Status), dataMap["new_status"])
	assert.Equal(t, changedBy, dataMap["changed_by"])
	assert.Equal(t, reason, dataMap["reason"])
}

func TestNewUserEmailChangedEvent(t *testing.T) {
	user, err := NewUser("user-123", "john.doe@example.com", "Password123", "John", "Doe")
	require.NoError(t, err)

	previousEmail := user.Email
	user.Email = "jane.doe@example.com"

	event := NewUserEmailChangedEvent(user, previousEmail)

	assert.NotNil(t, event)
	assert.Equal(t, "user.email_changed", event.EventType())
	assert.Equal(t, user.ID, event.AggregateID())
	assert.Equal(t, user.ID, event.UserID)
	assert.Equal(t, previousEmail, event.PreviousEmail)
	assert.Equal(t, user.Email, event.NewEmail)
	assert.NotEmpty(t, event.BaseEvent.ID)
	assert.True(t, time.Since(event.OccurredAt()) < time.Second)

	// Test EventData
	data := event.EventData()
	dataMap, ok := data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, user.ID, dataMap["user_id"])
	assert.Equal(t, previousEmail, dataMap["previous_email"])
	assert.Equal(t, user.Email, dataMap["new_email"])
}

func TestToJSON(t *testing.T) {
	user, err := NewUser("user-123", "john.doe@example.com", "Password123", "John", "Doe")
	require.NoError(t, err)

	event := NewUserCreatedEvent(user)

	jsonData, err := ToJSON(event)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Verify it's valid JSON by unmarshaling
	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(t, err)

	// Check some key fields
	assert.Equal(t, "user.created", unmarshaled["type"])
	assert.Equal(t, user.ID, unmarshaled["aggregate_id"])
	assert.Equal(t, user.ID, unmarshaled["user_id"])
	assert.Equal(t, user.Email, unmarshaled["email"])
}

func TestGenerateEventID(t *testing.T) {
	id1 := generateEventID()
	id2 := generateEventID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2) // Should generate unique IDs

	// Check UUID format (should contain hyphens and be 36 characters)
	assert.Contains(t, id1, "-")
	assert.Contains(t, id2, "-")
	assert.Len(t, id1, 36) // Standard UUID length
	assert.Len(t, id2, 36) // Standard UUID length
}

func TestEventSerialization(t *testing.T) {
	user, err := NewUser("user-123", "john.doe@example.com", "Password123", "John", "Doe")
	require.NoError(t, err)

	tests := []struct {
		name  string
		event DomainEvent
	}{
		{"UserCreatedEvent", NewUserCreatedEvent(user)},
		{"UserUpdatedEvent", NewUserUpdatedEvent(user, map[string]interface{}{"first_name": "Jane"})},
		{"UserDeletedEvent", NewUserDeletedEvent(user, "admin", "test")},
		{"UserStatusChangedEvent", NewUserStatusChangedEvent(user, UserStatusInactive, "admin", "test")},
		{"UserEmailChangedEvent", NewUserEmailChangedEvent(user, "old@example.com")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON serialization
			jsonData, err := ToJSON(tt.event)
			assert.NoError(t, err)
			assert.NotEmpty(t, jsonData)

			// Verify it's valid JSON
			var unmarshaled map[string]interface{}
			err = json.Unmarshal(jsonData, &unmarshaled)
			assert.NoError(t, err)

			// Check common fields
			assert.Equal(t, tt.event.EventType(), unmarshaled["type"])
			assert.Equal(t, tt.event.AggregateID(), unmarshaled["aggregate_id"])
			assert.NotEmpty(t, unmarshaled["id"])
			assert.NotEmpty(t, unmarshaled["occurred_at"])
			assert.Equal(t, float64(1), unmarshaled["version"]) // JSON numbers are float64
		})
	}
}
