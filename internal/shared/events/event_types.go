package events

import "time"

// Event type constants for the application
const (
	// User module events
	UserCreatedEventType     = "user.created"
	UserUpdatedEventType     = "user.updated"
	UserDeletedEventType     = "user.deleted"
	UserActivatedEventType   = "user.activated"
	UserDeactivatedEventType = "user.deactivated"

	// Auth module events
	UserLoggedInEventType   = "auth.user_logged_in"
	UserLoggedOutEventType  = "auth.user_logged_out"
	UserRegisteredEventType = "auth.user_registered"
	SessionExpiredEventType = "auth.session_expired"
	LoginFailedEventType    = "auth.login_failed"

	// System events
	SystemStartedEventType      = "system.started"
	SystemShuttingDownEventType = "system.shutting_down"
	HealthCheckFailedEventType  = "system.health_check_failed"
)

// Aggregate type constants
const (
	UserAggregateType    = "User"
	SessionAggregateType = "Session"
	SystemAggregateType  = "System"
)

// EventEnvelope wraps a domain event with additional routing information
type EventEnvelope struct {
	Event     DomainEvent `json:"event"`
	Timestamp time.Time   `json:"timestamp"`
	Retry     int         `json:"retry"`
	MaxRetry  int         `json:"max_retry"`
}

// NewEventEnvelope creates a new event envelope
func NewEventEnvelope(event DomainEvent) *EventEnvelope {
	return &EventEnvelope{
		Event:     event,
		Timestamp: time.Now().UTC(),
		Retry:     0,
		MaxRetry:  3,
	}
}

// ShouldRetry returns true if the event should be retried
func (e *EventEnvelope) ShouldRetry() bool {
	return e.Retry < e.MaxRetry
}

// IncrementRetry increments the retry counter
func (e *EventEnvelope) IncrementRetry() {
	e.Retry++
}
