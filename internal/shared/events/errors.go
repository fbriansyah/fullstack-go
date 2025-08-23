package events

import (
	"errors"
	"fmt"
)

// Common event system errors
var (
	ErrEventBusNotStarted     = errors.New("event bus not started")
	ErrEventBusAlreadyStarted = errors.New("event bus already started")
	ErrHandlerNotFound        = errors.New("handler not found")
	ErrHandlerAlreadyExists   = errors.New("handler already exists")
	ErrInvalidEvent           = errors.New("invalid event")
	ErrEventPublishFailed     = errors.New("failed to publish event")
	ErrEventHandlingFailed    = errors.New("failed to handle event")
)

// EventError represents an error that occurred during event processing
type EventError struct {
	EventID   string
	EventType string
	Handler   string
	Err       error
}

// Error implements the error interface
func (e *EventError) Error() string {
	if e.Handler != "" {
		return fmt.Sprintf("event error [%s:%s] in handler %s: %v",
			e.EventType, e.EventID, e.Handler, e.Err)
	}
	return fmt.Sprintf("event error [%s:%s]: %v",
		e.EventType, e.EventID, e.Err)
}

// Unwrap returns the underlying error
func (e *EventError) Unwrap() error {
	return e.Err
}

// NewEventError creates a new event error
func NewEventError(eventID, eventType, handler string, err error) *EventError {
	return &EventError{
		EventID:   eventID,
		EventType: eventType,
		Handler:   handler,
		Err:       err,
	}
}

// PublishError represents an error that occurred during event publishing
type PublishError struct {
	Event DomainEvent
	Err   error
}

// Error implements the error interface
func (e *PublishError) Error() string {
	return fmt.Sprintf("failed to publish event %s [%s]: %v",
		e.Event.EventType(), e.Event.EventID(), e.Err)
}

// Unwrap returns the underlying error
func (e *PublishError) Unwrap() error {
	return e.Err
}

// NewPublishError creates a new publish error
func NewPublishError(event DomainEvent, err error) *PublishError {
	return &PublishError{
		Event: event,
		Err:   err,
	}
}
