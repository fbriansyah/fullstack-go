package events

import (
	"context"
	"time"
)

// EventBus defines the interface for publishing and subscribing to domain events
type EventBus interface {
	// Publish sends an event to all registered handlers
	Publish(ctx context.Context, event DomainEvent) error

	// Subscribe registers an event handler for a specific event type
	Subscribe(eventType string, handler EventHandler) error

	// Unsubscribe removes an event handler for a specific event type
	Unsubscribe(eventType string, handler EventHandler) error

	// Start initializes the event bus and begins processing events
	Start(ctx context.Context) error

	// Stop gracefully shuts down the event bus
	Stop(ctx context.Context) error

	// Health returns the health status of the event bus
	Health() error
}

// DomainEvent represents a domain event that occurred in the system
type DomainEvent interface {
	// EventType returns the type identifier for this event
	EventType() string

	// AggregateID returns the ID of the aggregate that generated this event
	AggregateID() string

	// AggregateType returns the type of the aggregate that generated this event
	AggregateType() string

	// OccurredAt returns when the event occurred
	OccurredAt() time.Time

	// EventData returns the event-specific data
	EventData() interface{}

	// EventID returns a unique identifier for this event instance
	EventID() string

	// Version returns the version of the event schema
	Version() int

	// Metadata returns additional metadata about the event
	Metadata() EventMetadata
}

// EventHandler defines the interface for handling domain events
type EventHandler interface {
	// Handle processes the given domain event
	Handle(ctx context.Context, event DomainEvent) error

	// EventType returns the type of event this handler processes
	EventType() string

	// HandlerName returns a unique name for this handler
	HandlerName() string
}

// EventMetadata contains additional information about an event
type EventMetadata struct {
	// CorrelationID links related events together
	CorrelationID string `json:"correlation_id"`

	// CausationID identifies the event that caused this event
	CausationID string `json:"causation_id"`

	// UserID identifies the user who triggered the event (if applicable)
	UserID string `json:"user_id,omitempty"`

	// Source identifies the service/module that generated the event
	Source string `json:"source"`

	// TraceID for distributed tracing
	TraceID string `json:"trace_id,omitempty"`

	// Additional custom metadata
	Custom map[string]interface{} `json:"custom,omitempty"`
}
