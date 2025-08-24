package domain

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserLoggedInEvent(t *testing.T) {
	userID := "user-123"
	sessionID := "session-456"
	ipAddress := "192.168.1.1"
	userAgent := "Mozilla/5.0"

	event := NewUserLoggedInEvent(userID, sessionID, ipAddress, userAgent)

	assert.Equal(t, EventTypeUserLoggedIn, event.EventType())
	assert.Equal(t, userID, event.AggregateID())
	assert.Equal(t, userID, event.UserID)
	assert.Equal(t, sessionID, event.SessionID)
	assert.Equal(t, ipAddress, event.IPAddress)
	assert.Equal(t, userAgent, event.UserAgent)
	assert.WithinDuration(t, time.Now(), event.OccurredAt(), time.Second)
	assert.WithinDuration(t, time.Now(), event.LoginAt, time.Second)
	assert.Equal(t, *event, event.EventData())
}

func TestUserRegisteredEvent(t *testing.T) {
	userID := "user-123"
	email := "user@example.com"
	firstName := "John"
	lastName := "Doe"
	ipAddress := "192.168.1.1"
	userAgent := "Mozilla/5.0"

	event := NewUserRegisteredEvent(userID, email, firstName, lastName, ipAddress, userAgent)

	assert.Equal(t, EventTypeUserRegistered, event.EventType())
	assert.Equal(t, userID, event.AggregateID())
	assert.Equal(t, userID, event.UserID)
	assert.Equal(t, email, event.Email)
	assert.Equal(t, firstName, event.FirstName)
	assert.Equal(t, lastName, event.LastName)
	assert.Equal(t, ipAddress, event.IPAddress)
	assert.Equal(t, userAgent, event.UserAgent)
	assert.WithinDuration(t, time.Now(), event.OccurredAt(), time.Second)
	assert.WithinDuration(t, time.Now(), event.RegisteredAt, time.Second)
	assert.Equal(t, *event, event.EventData())
}

func TestUserLoggedOutEvent(t *testing.T) {
	userID := "user-123"
	sessionID := "session-456"
	logoutType := "manual"

	event := NewUserLoggedOutEvent(userID, sessionID, logoutType)

	assert.Equal(t, EventTypeUserLoggedOut, event.EventType())
	assert.Equal(t, userID, event.AggregateID())
	assert.Equal(t, userID, event.UserID)
	assert.Equal(t, sessionID, event.SessionID)
	assert.Equal(t, logoutType, event.LogoutType)
	assert.WithinDuration(t, time.Now(), event.OccurredAt(), time.Second)
	assert.WithinDuration(t, time.Now(), event.LogoutAt, time.Second)
	assert.Equal(t, *event, event.EventData())
}

func TestSessionExpiredEvent(t *testing.T) {
	userID := "user-123"
	sessionID := "session-456"
	lastUsedAt := time.Now().Add(-2 * time.Hour)

	event := NewSessionExpiredEvent(userID, sessionID, lastUsedAt)

	assert.Equal(t, EventTypeSessionExpired, event.EventType())
	assert.Equal(t, userID, event.AggregateID())
	assert.Equal(t, userID, event.UserID)
	assert.Equal(t, sessionID, event.SessionID)
	assert.Equal(t, lastUsedAt, event.LastUsedAt)
	assert.WithinDuration(t, time.Now(), event.OccurredAt(), time.Second)
	assert.WithinDuration(t, time.Now(), event.ExpiredAt, time.Second)
	assert.Equal(t, *event, event.EventData())
}

func TestPasswordChangedEvent(t *testing.T) {
	userID := "user-123"
	ipAddress := "192.168.1.1"
	userAgent := "Mozilla/5.0"

	event := NewPasswordChangedEvent(userID, ipAddress, userAgent)

	assert.Equal(t, EventTypePasswordChanged, event.EventType())
	assert.Equal(t, userID, event.AggregateID())
	assert.Equal(t, userID, event.UserID)
	assert.Equal(t, ipAddress, event.IPAddress)
	assert.Equal(t, userAgent, event.UserAgent)
	assert.WithinDuration(t, time.Now(), event.OccurredAt(), time.Second)
	assert.WithinDuration(t, time.Now(), event.ChangedAt, time.Second)
	assert.Equal(t, *event, event.EventData())
}

func TestUserLoggedInEvent_JSON(t *testing.T) {
	event := NewUserLoggedInEvent("user-123", "session-456", "192.168.1.1", "Mozilla/5.0")

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, EventTypeUserLoggedIn, unmarshaled["event_type"])
	assert.Equal(t, "user-123", unmarshaled["user_id"])
	assert.Equal(t, "session-456", unmarshaled["session_id"])
	assert.Equal(t, "192.168.1.1", unmarshaled["ip_address"])
	assert.Equal(t, "Mozilla/5.0", unmarshaled["user_agent"])
	assert.NotEmpty(t, unmarshaled["login_at"])
}

func TestUserRegisteredEvent_JSON(t *testing.T) {
	event := NewUserRegisteredEvent("user-123", "user@example.com", "John", "Doe", "192.168.1.1", "Mozilla/5.0")

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, EventTypeUserRegistered, unmarshaled["event_type"])
	assert.Equal(t, "user-123", unmarshaled["user_id"])
	assert.Equal(t, "user@example.com", unmarshaled["email"])
	assert.Equal(t, "John", unmarshaled["first_name"])
	assert.Equal(t, "Doe", unmarshaled["last_name"])
	assert.Equal(t, "192.168.1.1", unmarshaled["ip_address"])
	assert.Equal(t, "Mozilla/5.0", unmarshaled["user_agent"])
	assert.NotEmpty(t, unmarshaled["registered_at"])
}

func TestUserLoggedOutEvent_JSON(t *testing.T) {
	event := NewUserLoggedOutEvent("user-123", "session-456", "manual")

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, EventTypeUserLoggedOut, unmarshaled["event_type"])
	assert.Equal(t, "user-123", unmarshaled["user_id"])
	assert.Equal(t, "session-456", unmarshaled["session_id"])
	assert.Equal(t, "manual", unmarshaled["logout_type"])
	assert.NotEmpty(t, unmarshaled["logout_at"])
}

func TestSessionExpiredEvent_JSON(t *testing.T) {
	lastUsedAt := time.Now().Add(-2 * time.Hour)
	event := NewSessionExpiredEvent("user-123", "session-456", lastUsedAt)

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, EventTypeSessionExpired, unmarshaled["event_type"])
	assert.Equal(t, "user-123", unmarshaled["user_id"])
	assert.Equal(t, "session-456", unmarshaled["session_id"])
	assert.NotEmpty(t, unmarshaled["expired_at"])
	assert.NotEmpty(t, unmarshaled["last_used_at"])
}

func TestPasswordChangedEvent_JSON(t *testing.T) {
	event := NewPasswordChangedEvent("user-123", "192.168.1.1", "Mozilla/5.0")

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, EventTypePasswordChanged, unmarshaled["event_type"])
	assert.Equal(t, "user-123", unmarshaled["user_id"])
	assert.Equal(t, "192.168.1.1", unmarshaled["ip_address"])
	assert.Equal(t, "Mozilla/5.0", unmarshaled["user_agent"])
	assert.NotEmpty(t, unmarshaled["changed_at"])
}

func TestEventTypes(t *testing.T) {
	assert.Equal(t, "auth.user.logged_in", EventTypeUserLoggedIn)
	assert.Equal(t, "auth.user.registered", EventTypeUserRegistered)
	assert.Equal(t, "auth.user.logged_out", EventTypeUserLoggedOut)
	assert.Equal(t, "auth.session.expired", EventTypeSessionExpired)
	assert.Equal(t, "auth.password.changed", EventTypePasswordChanged)
}

func TestLogoutTypes(t *testing.T) {
	// Test different logout types
	logoutTypes := []string{"manual", "timeout", "forced"}

	for _, logoutType := range logoutTypes {
		event := NewUserLoggedOutEvent("user-123", "session-456", logoutType)
		assert.Equal(t, logoutType, event.LogoutType)
	}
}
