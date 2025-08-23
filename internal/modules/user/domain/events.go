package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// DomainEvent represents a domain event interface
type DomainEvent interface {
	EventType() string
	AggregateID() string
	OccurredAt() time.Time
	EventData() interface{}
}

// BaseEvent provides common event functionality
type BaseEvent struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	AggregateId string    `json:"aggregate_id"`
	OccurredOn  time.Time `json:"occurred_at"`
	Version     int       `json:"version"`
}

// EventType returns the event type
func (e BaseEvent) EventType() string {
	return e.Type
}

// AggregateID returns the aggregate ID
func (e BaseEvent) AggregateID() string {
	return e.AggregateId
}

// OccurredAt returns when the event occurred
func (e BaseEvent) OccurredAt() time.Time {
	return e.OccurredOn
}

// UserCreatedEvent represents a user creation event
type UserCreatedEvent struct {
	BaseEvent
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Status    string `json:"status"`
}

// NewUserCreatedEvent creates a new UserCreatedEvent
func NewUserCreatedEvent(user *User) *UserCreatedEvent {
	return &UserCreatedEvent{
		BaseEvent: BaseEvent{
			ID:          generateEventID(),
			Type:        "user.created",
			AggregateId: user.ID,
			OccurredOn:  time.Now().UTC(),
			Version:     1,
		},
		UserID:    user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Status:    string(user.Status),
	}
}

// EventData returns the event data
func (e *UserCreatedEvent) EventData() interface{} {
	return map[string]interface{}{
		"user_id":    e.UserID,
		"email":      e.Email,
		"first_name": e.FirstName,
		"last_name":  e.LastName,
		"status":     e.Status,
	}
}

// UserUpdatedEvent represents a user update event
type UserUpdatedEvent struct {
	BaseEvent
	UserID          string                 `json:"user_id"`
	Changes         map[string]interface{} `json:"changes"`
	PreviousVersion int                    `json:"previous_version"`
}

// NewUserUpdatedEvent creates a new UserUpdatedEvent
func NewUserUpdatedEvent(user *User, changes map[string]interface{}) *UserUpdatedEvent {
	return &UserUpdatedEvent{
		BaseEvent: BaseEvent{
			ID:          generateEventID(),
			Type:        "user.updated",
			AggregateId: user.ID,
			OccurredOn:  time.Now().UTC(),
			Version:     1,
		},
		UserID:          user.ID,
		Changes:         changes,
		PreviousVersion: user.Version - 1,
	}
}

// EventData returns the event data
func (e *UserUpdatedEvent) EventData() interface{} {
	return map[string]interface{}{
		"user_id":          e.UserID,
		"changes":          e.Changes,
		"previous_version": e.PreviousVersion,
	}
}

// UserDeletedEvent represents a user deletion event
type UserDeletedEvent struct {
	BaseEvent
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	DeletedBy string `json:"deleted_by,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

// NewUserDeletedEvent creates a new UserDeletedEvent
func NewUserDeletedEvent(user *User, deletedBy, reason string) *UserDeletedEvent {
	return &UserDeletedEvent{
		BaseEvent: BaseEvent{
			ID:          generateEventID(),
			Type:        "user.deleted",
			AggregateId: user.ID,
			OccurredOn:  time.Now().UTC(),
			Version:     1,
		},
		UserID:    user.ID,
		Email:     user.Email,
		DeletedBy: deletedBy,
		Reason:    reason,
	}
}

// EventData returns the event data
func (e *UserDeletedEvent) EventData() interface{} {
	return map[string]interface{}{
		"user_id":    e.UserID,
		"email":      e.Email,
		"deleted_by": e.DeletedBy,
		"reason":     e.Reason,
	}
}

// UserStatusChangedEvent represents a user status change event
type UserStatusChangedEvent struct {
	BaseEvent
	UserID         string `json:"user_id"`
	PreviousStatus string `json:"previous_status"`
	NewStatus      string `json:"new_status"`
	ChangedBy      string `json:"changed_by,omitempty"`
	Reason         string `json:"reason,omitempty"`
}

// NewUserStatusChangedEvent creates a new UserStatusChangedEvent
func NewUserStatusChangedEvent(user *User, previousStatus UserStatus, changedBy, reason string) *UserStatusChangedEvent {
	return &UserStatusChangedEvent{
		BaseEvent: BaseEvent{
			ID:          generateEventID(),
			Type:        "user.status_changed",
			AggregateId: user.ID,
			OccurredOn:  time.Now().UTC(),
			Version:     1,
		},
		UserID:         user.ID,
		PreviousStatus: string(previousStatus),
		NewStatus:      string(user.Status),
		ChangedBy:      changedBy,
		Reason:         reason,
	}
}

// EventData returns the event data
func (e *UserStatusChangedEvent) EventData() interface{} {
	return map[string]interface{}{
		"user_id":         e.UserID,
		"previous_status": e.PreviousStatus,
		"new_status":      e.NewStatus,
		"changed_by":      e.ChangedBy,
		"reason":          e.Reason,
	}
}

// UserEmailChangedEvent represents a user email change event
type UserEmailChangedEvent struct {
	BaseEvent
	UserID        string `json:"user_id"`
	PreviousEmail string `json:"previous_email"`
	NewEmail      string `json:"new_email"`
}

// NewUserEmailChangedEvent creates a new UserEmailChangedEvent
func NewUserEmailChangedEvent(user *User, previousEmail string) *UserEmailChangedEvent {
	return &UserEmailChangedEvent{
		BaseEvent: BaseEvent{
			ID:          generateEventID(),
			Type:        "user.email_changed",
			AggregateId: user.ID,
			OccurredOn:  time.Now().UTC(),
			Version:     1,
		},
		UserID:        user.ID,
		PreviousEmail: previousEmail,
		NewEmail:      user.Email,
	}
}

// EventData returns the event data
func (e *UserEmailChangedEvent) EventData() interface{} {
	return map[string]interface{}{
		"user_id":        e.UserID,
		"previous_email": e.PreviousEmail,
		"new_email":      e.NewEmail,
	}
}

// ToJSON converts the event to JSON
func ToJSON(event DomainEvent) ([]byte, error) {
	return json.Marshal(event)
}

// generateEventID generates a unique event ID using UUID
func generateEventID() string {
	return uuid.New().String()
}
