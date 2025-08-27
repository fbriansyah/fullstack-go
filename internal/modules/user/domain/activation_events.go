package domain

import (
	"encoding/json"
	"time"
)

// UserActivationRequestedEvent represents a user activation request event
type UserActivationRequestedEvent struct {
	BaseEvent
	UserID          string    `json:"user_id"`
	Email           string    `json:"email"`
	ActivationToken string    `json:"activation_token"`
	ExpiresAt       time.Time `json:"expires_at"`
	RequestedBy     string    `json:"requested_by,omitempty"`
}

// NewUserActivationRequestedEvent creates a new UserActivationRequestedEvent
func NewUserActivationRequestedEvent(user *User, activationToken string, expiresAt time.Time, requestedBy string) *UserActivationRequestedEvent {
	return &UserActivationRequestedEvent{
		BaseEvent: BaseEvent{
			ID:           generateEventID(),
			Type:         "user.activation_requested",
			AggregateId:  user.ID,
			AggregateTyp: "user",
			OccurredOn:   time.Now().UTC(),
			EventVersion: 1,
		},
		UserID:          user.ID,
		Email:           user.Email,
		ActivationToken: activationToken,
		ExpiresAt:       expiresAt,
		RequestedBy:     requestedBy,
	}
}

// EventData returns the event data
func (e *UserActivationRequestedEvent) EventData() interface{} {
	return map[string]interface{}{
		"user_id":          e.UserID,
		"email":            e.Email,
		"activation_token": e.ActivationToken,
		"expires_at":       e.ExpiresAt,
		"requested_by":     e.RequestedBy,
	}
}

// UserActivatedEvent represents a user activation completion event
type UserActivatedEvent struct {
	BaseEvent
	UserID      string    `json:"user_id"`
	Email       string    `json:"email"`
	ActivatedAt time.Time `json:"activated_at"`
	ActivatedBy string    `json:"activated_by,omitempty"`
	Method      string    `json:"method"` // "token", "admin", "auto"
}

// NewUserActivatedEvent creates a new UserActivatedEvent
func NewUserActivatedEvent(user *User, activatedBy, method string) *UserActivatedEvent {
	return &UserActivatedEvent{
		BaseEvent: BaseEvent{
			ID:           generateEventID(),
			Type:         "user.activated",
			AggregateId:  user.ID,
			AggregateTyp: "user",
			OccurredOn:   time.Now().UTC(),
			EventVersion: 1,
		},
		UserID:      user.ID,
		Email:       user.Email,
		ActivatedAt: time.Now().UTC(),
		ActivatedBy: activatedBy,
		Method:      method,
	}
}

// EventData returns the event data
func (e *UserActivatedEvent) EventData() interface{} {
	return map[string]interface{}{
		"user_id":      e.UserID,
		"email":        e.Email,
		"activated_at": e.ActivatedAt,
		"activated_by": e.ActivatedBy,
		"method":       e.Method,
	}
}

// UserDeactivatedEvent represents a user deactivation event
type UserDeactivatedEvent struct {
	BaseEvent
	UserID        string    `json:"user_id"`
	Email         string    `json:"email"`
	DeactivatedAt time.Time `json:"deactivated_at"`
	DeactivatedBy string    `json:"deactivated_by,omitempty"`
	Reason        string    `json:"reason,omitempty"`
}

// NewUserDeactivatedEvent creates a new UserDeactivatedEvent
func NewUserDeactivatedEvent(user *User, deactivatedBy, reason string) *UserDeactivatedEvent {
	return &UserDeactivatedEvent{
		BaseEvent: BaseEvent{
			ID:           generateEventID(),
			Type:         "user.deactivated",
			AggregateId:  user.ID,
			AggregateTyp: "user",
			OccurredOn:   time.Now().UTC(),
			EventVersion: 1,
		},
		UserID:        user.ID,
		Email:         user.Email,
		DeactivatedAt: time.Now().UTC(),
		DeactivatedBy: deactivatedBy,
		Reason:        reason,
	}
}

// EventData returns the event data
func (e *UserDeactivatedEvent) EventData() interface{} {
	return map[string]interface{}{
		"user_id":        e.UserID,
		"email":          e.Email,
		"deactivated_at": e.DeactivatedAt,
		"deactivated_by": e.DeactivatedBy,
		"reason":         e.Reason,
	}
}

// UserActivationTokenExpiredEvent represents an activation token expiration event
type UserActivationTokenExpiredEvent struct {
	BaseEvent
	UserID          string    `json:"user_id"`
	Email           string    `json:"email"`
	ActivationToken string    `json:"activation_token"`
	ExpiredAt       time.Time `json:"expired_at"`
}

// NewUserActivationTokenExpiredEvent creates a new UserActivationTokenExpiredEvent
func NewUserActivationTokenExpiredEvent(userID, email, activationToken string) *UserActivationTokenExpiredEvent {
	return &UserActivationTokenExpiredEvent{
		BaseEvent: BaseEvent{
			ID:           generateEventID(),
			Type:         "user.activation_token_expired",
			AggregateId:  userID,
			AggregateTyp: "user",
			OccurredOn:   time.Now().UTC(),
			EventVersion: 1,
		},
		UserID:          userID,
		Email:           email,
		ActivationToken: activationToken,
		ExpiredAt:       time.Now().UTC(),
	}
}

// EventData returns the event data
func (e *UserActivationTokenExpiredEvent) EventData() interface{} {
	return map[string]interface{}{
		"user_id":          e.UserID,
		"email":            e.Email,
		"activation_token": e.ActivationToken,
		"expired_at":       e.ExpiredAt,
	}
}

// ToJSON converts the activation events to JSON
func (e *UserActivationRequestedEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func (e *UserActivatedEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func (e *UserDeactivatedEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func (e *UserActivationTokenExpiredEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}
