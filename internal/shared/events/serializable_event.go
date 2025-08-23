package events

import (
	"encoding/json"
	"time"
)

// SerializableEvent represents a domain event that can be serialized to JSON
type SerializableEvent struct {
	ID        string                 `json:"event_id"`
	Type      string                 `json:"event_type"`
	AggID     string                 `json:"aggregate_id"`
	AggType   string                 `json:"aggregate_type"`
	Timestamp time.Time              `json:"occurred_at"`
	Ver       int                    `json:"version"`
	Meta      EventMetadata          `json:"metadata"`
	Data      map[string]interface{} `json:"data"`
}

// NewSerializableEvent creates a serializable event from a domain event
func NewSerializableEvent(event DomainEvent) (*SerializableEvent, error) {
	// Convert event data to map for JSON serialization
	var dataMap map[string]interface{}

	if event.EventData() != nil {
		// Marshal and unmarshal to convert to map
		dataBytes, err := json.Marshal(event.EventData())
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(dataBytes, &dataMap); err != nil {
			return nil, err
		}
	}

	return &SerializableEvent{
		ID:        event.EventID(),
		Type:      event.EventType(),
		AggID:     event.AggregateID(),
		AggType:   event.AggregateType(),
		Timestamp: event.OccurredAt(),
		Ver:       event.Version(),
		Meta:      event.Metadata(),
		Data:      dataMap,
	}, nil
}

// Implement DomainEvent interface
func (s *SerializableEvent) EventID() string {
	return s.ID
}

func (s *SerializableEvent) EventType() string {
	return s.Type
}

func (s *SerializableEvent) AggregateID() string {
	return s.AggID
}

func (s *SerializableEvent) AggregateType() string {
	return s.AggType
}

func (s *SerializableEvent) OccurredAt() time.Time {
	return s.Timestamp
}

func (s *SerializableEvent) Version() int {
	return s.Ver
}

func (s *SerializableEvent) Metadata() EventMetadata {
	return s.Meta
}

func (s *SerializableEvent) EventData() interface{} {
	return s.Data
}

// SerializableEventEnvelope wraps a serializable event with additional routing information
type SerializableEventEnvelope struct {
	Event     *SerializableEvent `json:"event"`
	Timestamp time.Time          `json:"timestamp"`
	Retry     int                `json:"retry"`
	MaxRetry  int                `json:"max_retry"`
}

// NewSerializableEventEnvelope creates a new serializable event envelope
func NewSerializableEventEnvelope(event DomainEvent) (*SerializableEventEnvelope, error) {
	serializableEvent, err := NewSerializableEvent(event)
	if err != nil {
		return nil, err
	}

	return &SerializableEventEnvelope{
		Event:     serializableEvent,
		Timestamp: time.Now().UTC(),
		Retry:     0,
		MaxRetry:  3,
	}, nil
}

// ShouldRetry returns true if the event should be retried
func (e *SerializableEventEnvelope) ShouldRetry() bool {
	return e.Retry < e.MaxRetry
}

// IncrementRetry increments the retry counter
func (e *SerializableEventEnvelope) IncrementRetry() {
	e.Retry++
}
