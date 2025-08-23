package events

import (
	"time"

	"github.com/google/uuid"
)

// BaseEvent provides a common implementation for domain events
type BaseEvent struct {
	eventID       string
	eventType     string
	aggregateID   string
	aggregateType string
	occurredAt    time.Time
	version       int
	metadata      EventMetadata
	data          interface{}
}

// NewBaseEvent creates a new base event with required fields
func NewBaseEvent(eventType, aggregateID, aggregateType string, data interface{}) *BaseEvent {
	return &BaseEvent{
		eventID:       uuid.New().String(),
		eventType:     eventType,
		aggregateID:   aggregateID,
		aggregateType: aggregateType,
		occurredAt:    time.Now().UTC(),
		version:       1,
		data:          data,
		metadata: EventMetadata{
			CorrelationID: uuid.New().String(),
			Source:        "go-templ-template",
			Custom:        make(map[string]interface{}),
		},
	}
}

// EventID returns the unique identifier for this event instance
func (e *BaseEvent) EventID() string {
	return e.eventID
}

// EventType returns the type identifier for this event
func (e *BaseEvent) EventType() string {
	return e.eventType
}

// AggregateID returns the ID of the aggregate that generated this event
func (e *BaseEvent) AggregateID() string {
	return e.aggregateID
}

// AggregateType returns the type of the aggregate that generated this event
func (e *BaseEvent) AggregateType() string {
	return e.aggregateType
}

// OccurredAt returns when the event occurred
func (e *BaseEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// EventData returns the event-specific data
func (e *BaseEvent) EventData() interface{} {
	return e.data
}

// Version returns the version of the event schema
func (e *BaseEvent) Version() int {
	return e.version
}

// Metadata returns additional metadata about the event
func (e *BaseEvent) Metadata() EventMetadata {
	return e.metadata
}

// SetMetadata updates the event metadata
func (e *BaseEvent) SetMetadata(metadata EventMetadata) {
	e.metadata = metadata
}

// SetCorrelationID sets the correlation ID for event tracing
func (e *BaseEvent) SetCorrelationID(correlationID string) {
	e.metadata.CorrelationID = correlationID
}

// SetCausationID sets the causation ID to link this event to its cause
func (e *BaseEvent) SetCausationID(causationID string) {
	e.metadata.CausationID = causationID
}

// SetUserID sets the user ID who triggered this event
func (e *BaseEvent) SetUserID(userID string) {
	e.metadata.UserID = userID
}

// SetTraceID sets the trace ID for distributed tracing
func (e *BaseEvent) SetTraceID(traceID string) {
	e.metadata.TraceID = traceID
}

// AddCustomMetadata adds custom metadata to the event
func (e *BaseEvent) AddCustomMetadata(key string, value interface{}) {
	if e.metadata.Custom == nil {
		e.metadata.Custom = make(map[string]interface{})
	}
	e.metadata.Custom[key] = value
}
