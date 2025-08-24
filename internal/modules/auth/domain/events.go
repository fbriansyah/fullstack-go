package domain

import (
	"encoding/json"
	"time"

	"go-templ-template/internal/shared/events"

	"github.com/google/uuid"
)

// Event types for authentication domain
const (
	EventTypeUserLoggedIn    = "auth.user.logged_in"
	EventTypeUserRegistered  = "auth.user.registered"
	EventTypeUserLoggedOut   = "auth.user.logged_out"
	EventTypeSessionExpired  = "auth.session.expired"
	EventTypePasswordChanged = "auth.password.changed"
)

// UserLoggedInEvent represents a user login event
type UserLoggedInEvent struct {
	UserID    string    `json:"user_id"`
	SessionID string    `json:"session_id"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	LoginAt   time.Time `json:"login_at"`
	eventID   string
	metadata  map[string]interface{}
}

// EventType returns the event type
func (e UserLoggedInEvent) EventType() string {
	return EventTypeUserLoggedIn
}

// AggregateID returns the aggregate ID (user ID)
func (e UserLoggedInEvent) AggregateID() string {
	return e.UserID
}

// AggregateType returns the aggregate type
func (e UserLoggedInEvent) AggregateType() string {
	return "User"
}

// OccurredAt returns when the event occurred
func (e UserLoggedInEvent) OccurredAt() time.Time {
	return e.LoginAt
}

// EventData returns the event data
func (e UserLoggedInEvent) EventData() interface{} {
	return e
}

// EventID returns the unique event ID
func (e UserLoggedInEvent) EventID() string {
	return e.eventID
}

// Version returns the event version
func (e UserLoggedInEvent) Version() int {
	return 1
}

// Metadata returns event metadata
func (e UserLoggedInEvent) Metadata() events.EventMetadata {
	return events.EventMetadata{
		CorrelationID: e.eventID,
		Source:        "auth-service",
		Custom:        e.metadata,
	}
}

// UserRegisteredEvent represents a user registration event
type UserRegisteredEvent struct {
	UserID       string    `json:"user_id"`
	Email        string    `json:"email"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	RegisteredAt time.Time `json:"registered_at"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	eventID      string
	metadata     map[string]interface{}
}

// EventType returns the event type
func (e UserRegisteredEvent) EventType() string {
	return EventTypeUserRegistered
}

// AggregateID returns the aggregate ID (user ID)
func (e UserRegisteredEvent) AggregateID() string {
	return e.UserID
}

// AggregateType returns the aggregate type
func (e UserRegisteredEvent) AggregateType() string {
	return "User"
}

// OccurredAt returns when the event occurred
func (e UserRegisteredEvent) OccurredAt() time.Time {
	return e.RegisteredAt
}

// EventData returns the event data
func (e UserRegisteredEvent) EventData() interface{} {
	return e
}

// EventID returns the unique event ID
func (e UserRegisteredEvent) EventID() string {
	return e.eventID
}

// Version returns the event version
func (e UserRegisteredEvent) Version() int {
	return 1
}

// Metadata returns event metadata
func (e UserRegisteredEvent) Metadata() events.EventMetadata {
	return events.EventMetadata{
		CorrelationID: e.eventID,
		Source:        "auth-service",
		Custom:        e.metadata,
	}
}

// UserLoggedOutEvent represents a user logout event
type UserLoggedOutEvent struct {
	UserID     string    `json:"user_id"`
	SessionID  string    `json:"session_id"`
	LogoutAt   time.Time `json:"logout_at"`
	LogoutType string    `json:"logout_type"` // "manual", "timeout", "forced"
	eventID    string
	metadata   map[string]interface{}
}

// EventType returns the event type
func (e UserLoggedOutEvent) EventType() string {
	return EventTypeUserLoggedOut
}

// AggregateID returns the aggregate ID (user ID)
func (e UserLoggedOutEvent) AggregateID() string {
	return e.UserID
}

// AggregateType returns the aggregate type
func (e UserLoggedOutEvent) AggregateType() string {
	return "User"
}

// OccurredAt returns when the event occurred
func (e UserLoggedOutEvent) OccurredAt() time.Time {
	return e.LogoutAt
}

// EventData returns the event data
func (e UserLoggedOutEvent) EventData() interface{} {
	return e
}

// EventID returns the unique event ID
func (e UserLoggedOutEvent) EventID() string {
	return e.eventID
}

// Version returns the event version
func (e UserLoggedOutEvent) Version() int {
	return 1
}

// Metadata returns event metadata
func (e UserLoggedOutEvent) Metadata() events.EventMetadata {
	return events.EventMetadata{
		CorrelationID: e.eventID,
		Source:        "auth-service",
		Custom:        e.metadata,
	}
}

// SessionExpiredEvent represents a session expiration event
type SessionExpiredEvent struct {
	UserID     string    `json:"user_id"`
	SessionID  string    `json:"session_id"`
	ExpiredAt  time.Time `json:"expired_at"`
	LastUsedAt time.Time `json:"last_used_at"`
	eventID    string
	metadata   map[string]interface{}
}

// EventType returns the event type
func (e SessionExpiredEvent) EventType() string {
	return EventTypeSessionExpired
}

// AggregateID returns the aggregate ID (user ID)
func (e SessionExpiredEvent) AggregateID() string {
	return e.UserID
}

// AggregateType returns the aggregate type
func (e SessionExpiredEvent) AggregateType() string {
	return "Session"
}

// OccurredAt returns when the event occurred
func (e SessionExpiredEvent) OccurredAt() time.Time {
	return e.ExpiredAt
}

// EventData returns the event data
func (e SessionExpiredEvent) EventData() interface{} {
	return e
}

// EventID returns the unique event ID
func (e SessionExpiredEvent) EventID() string {
	return e.eventID
}

// Version returns the event version
func (e SessionExpiredEvent) Version() int {
	return 1
}

// Metadata returns event metadata
func (e SessionExpiredEvent) Metadata() events.EventMetadata {
	return events.EventMetadata{
		CorrelationID: e.eventID,
		Source:        "auth-service",
		Custom:        e.metadata,
	}
}

// PasswordChangedEvent represents a password change event
type PasswordChangedEvent struct {
	UserID    string    `json:"user_id"`
	ChangedAt time.Time `json:"changed_at"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	eventID   string
	metadata  map[string]interface{}
}

// EventType returns the event type
func (e PasswordChangedEvent) EventType() string {
	return EventTypePasswordChanged
}

// AggregateID returns the aggregate ID (user ID)
func (e PasswordChangedEvent) AggregateID() string {
	return e.UserID
}

// AggregateType returns the aggregate type
func (e PasswordChangedEvent) AggregateType() string {
	return "User"
}

// OccurredAt returns when the event occurred
func (e PasswordChangedEvent) OccurredAt() time.Time {
	return e.ChangedAt
}

// EventData returns the event data
func (e PasswordChangedEvent) EventData() interface{} {
	return e
}

// EventID returns the unique event ID
func (e PasswordChangedEvent) EventID() string {
	return e.eventID
}

// Version returns the event version
func (e PasswordChangedEvent) Version() int {
	return 1
}

// Metadata returns event metadata
func (e PasswordChangedEvent) Metadata() events.EventMetadata {
	return events.EventMetadata{
		CorrelationID: e.eventID,
		Source:        "auth-service",
		Custom:        e.metadata,
	}
}

// NewUserLoggedInEvent creates a new user logged in event
func NewUserLoggedInEvent(userID, sessionID, ipAddress, userAgent string) *UserLoggedInEvent {
	return &UserLoggedInEvent{
		UserID:    userID,
		SessionID: sessionID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		LoginAt:   time.Now(),
		eventID:   uuid.New().String(),
		metadata:  make(map[string]interface{}),
	}
}

// NewUserRegisteredEvent creates a new user registered event
func NewUserRegisteredEvent(userID, email, firstName, lastName, ipAddress, userAgent string) *UserRegisteredEvent {
	return &UserRegisteredEvent{
		UserID:       userID,
		Email:        email,
		FirstName:    firstName,
		LastName:     lastName,
		RegisteredAt: time.Now(),
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		eventID:      uuid.New().String(),
		metadata:     make(map[string]interface{}),
	}
}

// NewUserLoggedOutEvent creates a new user logged out event
func NewUserLoggedOutEvent(userID, sessionID, logoutType string) *UserLoggedOutEvent {
	return &UserLoggedOutEvent{
		UserID:     userID,
		SessionID:  sessionID,
		LogoutAt:   time.Now(),
		LogoutType: logoutType,
		eventID:    uuid.New().String(),
		metadata:   make(map[string]interface{}),
	}
}

// NewSessionExpiredEvent creates a new session expired event
func NewSessionExpiredEvent(userID, sessionID string, lastUsedAt time.Time) *SessionExpiredEvent {
	return &SessionExpiredEvent{
		UserID:     userID,
		SessionID:  sessionID,
		ExpiredAt:  time.Now(),
		LastUsedAt: lastUsedAt,
		eventID:    uuid.New().String(),
		metadata:   make(map[string]interface{}),
	}
}

// NewPasswordChangedEvent creates a new password changed event
func NewPasswordChangedEvent(userID, ipAddress, userAgent string) *PasswordChangedEvent {
	return &PasswordChangedEvent{
		UserID:    userID,
		ChangedAt: time.Now(),
		IPAddress: ipAddress,
		UserAgent: userAgent,
		eventID:   uuid.New().String(),
		metadata:  make(map[string]interface{}),
	}
}

// MarshalJSON implements json.Marshaler for all events
func (e UserLoggedInEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		EventType string    `json:"event_type"`
		UserID    string    `json:"user_id"`
		SessionID string    `json:"session_id"`
		IPAddress string    `json:"ip_address"`
		UserAgent string    `json:"user_agent"`
		LoginAt   time.Time `json:"login_at"`
	}{
		EventType: e.EventType(),
		UserID:    e.UserID,
		SessionID: e.SessionID,
		IPAddress: e.IPAddress,
		UserAgent: e.UserAgent,
		LoginAt:   e.LoginAt,
	})
}

func (e UserRegisteredEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		EventType    string    `json:"event_type"`
		UserID       string    `json:"user_id"`
		Email        string    `json:"email"`
		FirstName    string    `json:"first_name"`
		LastName     string    `json:"last_name"`
		RegisteredAt time.Time `json:"registered_at"`
		IPAddress    string    `json:"ip_address"`
		UserAgent    string    `json:"user_agent"`
	}{
		EventType:    e.EventType(),
		UserID:       e.UserID,
		Email:        e.Email,
		FirstName:    e.FirstName,
		LastName:     e.LastName,
		RegisteredAt: e.RegisteredAt,
		IPAddress:    e.IPAddress,
		UserAgent:    e.UserAgent,
	})
}

func (e UserLoggedOutEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		EventType  string    `json:"event_type"`
		UserID     string    `json:"user_id"`
		SessionID  string    `json:"session_id"`
		LogoutAt   time.Time `json:"logout_at"`
		LogoutType string    `json:"logout_type"`
	}{
		EventType:  e.EventType(),
		UserID:     e.UserID,
		SessionID:  e.SessionID,
		LogoutAt:   e.LogoutAt,
		LogoutType: e.LogoutType,
	})
}

func (e SessionExpiredEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		EventType  string    `json:"event_type"`
		UserID     string    `json:"user_id"`
		SessionID  string    `json:"session_id"`
		ExpiredAt  time.Time `json:"expired_at"`
		LastUsedAt time.Time `json:"last_used_at"`
	}{
		EventType:  e.EventType(),
		UserID:     e.UserID,
		SessionID:  e.SessionID,
		ExpiredAt:  e.ExpiredAt,
		LastUsedAt: e.LastUsedAt,
	})
}

func (e PasswordChangedEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		EventType string    `json:"event_type"`
		UserID    string    `json:"user_id"`
		ChangedAt time.Time `json:"changed_at"`
		IPAddress string    `json:"ip_address"`
		UserAgent string    `json:"user_agent"`
	}{
		EventType: e.EventType(),
		UserID:    e.UserID,
		ChangedAt: e.ChangedAt,
		IPAddress: e.IPAddress,
		UserAgent: e.UserAgent,
	})
}
