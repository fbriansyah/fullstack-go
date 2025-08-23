package events

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

// MockEventHandler implements EventHandler for testing
type MockEventHandler struct {
	name          string
	eventType     string
	handledEvents []DomainEvent
	shouldError   bool
}

func NewMockEventHandler(name, eventType string) *MockEventHandler {
	return &MockEventHandler{
		name:          name,
		eventType:     eventType,
		handledEvents: make([]DomainEvent, 0),
	}
}

func (m *MockEventHandler) Handle(ctx context.Context, event DomainEvent) error {
	if m.shouldError {
		return fmt.Errorf("mock error")
	}
	m.handledEvents = append(m.handledEvents, event)
	return nil
}

func (m *MockEventHandler) EventType() string {
	return m.eventType
}

func (m *MockEventHandler) HandlerName() string {
	return m.name
}

func (m *MockEventHandler) SetShouldError(shouldError bool) {
	m.shouldError = shouldError
}

func (m *MockEventHandler) GetHandledEvents() []DomainEvent {
	return m.handledEvents
}

// TestEvent implements DomainEvent for testing
type TestEvent struct {
	*BaseEvent
	TestData string `json:"test_data"`
}

func NewTestEvent(aggregateID, testData string) *TestEvent {
	return &TestEvent{
		BaseEvent: NewBaseEvent("test.event", aggregateID, "TestAggregate", map[string]interface{}{
			"test_data": testData,
		}),
		TestData: testData,
	}
}

func TestRabbitMQEventBus_DefaultConfig(t *testing.T) {
	config := DefaultRabbitMQConfig()

	if config.URL != "amqp://guest:guest@localhost:5672/" {
		t.Errorf("Expected default URL, got %s", config.URL)
	}

	if config.Exchange != "events" {
		t.Errorf("Expected default exchange 'events', got %s", config.Exchange)
	}

	if config.ExchangeType != "topic" {
		t.Errorf("Expected exchange type 'topic', got %s", config.ExchangeType)
	}

	if !config.Durable {
		t.Error("Expected durable to be true")
	}
}

func TestRabbitMQEventBus_NewRabbitMQEventBus(t *testing.T) {
	config := DefaultRabbitMQConfig()
	bus := NewRabbitMQEventBus(config)

	if bus == nil {
		t.Fatal("Expected non-nil event bus")
	}

	if bus.config.URL != config.URL {
		t.Errorf("Expected config URL %s, got %s", config.URL, bus.config.URL)
	}

	if bus.handlers == nil {
		t.Error("Expected handlers map to be initialized")
	}

	if bus.consumers == nil {
		t.Error("Expected consumers map to be initialized")
	}
}

func TestRabbitMQEventBus_Subscribe(t *testing.T) {
	config := DefaultRabbitMQConfig()
	bus := NewRabbitMQEventBus(config)

	handler := NewMockEventHandler("test-handler", "test.event")

	err := bus.Subscribe("test.event", handler)
	if err != nil {
		t.Fatalf("Expected no error subscribing, got %v", err)
	}

	// Check that handler was added
	bus.handlersMux.RLock()
	handlers, exists := bus.handlers["test.event"]
	bus.handlersMux.RUnlock()

	if !exists {
		t.Fatal("Expected handlers to exist for test.event")
	}

	if len(handlers) != 1 {
		t.Fatalf("Expected 1 handler, got %d", len(handlers))
	}

	if handlers[0].HandlerName() != "test-handler" {
		t.Errorf("Expected handler name 'test-handler', got %s", handlers[0].HandlerName())
	}
}

func TestRabbitMQEventBus_Unsubscribe(t *testing.T) {
	config := DefaultRabbitMQConfig()
	bus := NewRabbitMQEventBus(config)

	handler1 := NewMockEventHandler("test-handler-1", "test.event")
	handler2 := NewMockEventHandler("test-handler-2", "test.event")

	// Subscribe both handlers
	bus.Subscribe("test.event", handler1)
	bus.Subscribe("test.event", handler2)

	// Verify both are subscribed
	bus.handlersMux.RLock()
	handlers := bus.handlers["test.event"]
	bus.handlersMux.RUnlock()

	if len(handlers) != 2 {
		t.Fatalf("Expected 2 handlers, got %d", len(handlers))
	}

	// Unsubscribe one handler
	err := bus.Unsubscribe("test.event", handler1)
	if err != nil {
		t.Fatalf("Expected no error unsubscribing, got %v", err)
	}

	// Verify only one handler remains
	bus.handlersMux.RLock()
	handlers = bus.handlers["test.event"]
	bus.handlersMux.RUnlock()

	if len(handlers) != 1 {
		t.Fatalf("Expected 1 handler after unsubscribe, got %d", len(handlers))
	}

	if handlers[0].HandlerName() != "test-handler-2" {
		t.Errorf("Expected remaining handler to be 'test-handler-2', got %s", handlers[0].HandlerName())
	}
}

func TestRabbitMQEventBus_Health_NotStarted(t *testing.T) {
	config := DefaultRabbitMQConfig()
	bus := NewRabbitMQEventBus(config)

	err := bus.Health()
	if err == nil {
		t.Error("Expected health check to fail when not started")
	}
}

func TestSerializableEvent_Creation(t *testing.T) {
	testEvent := NewTestEvent("test-id", "test-data")

	serializableEvent, err := NewSerializableEvent(testEvent)
	if err != nil {
		t.Fatalf("Expected no error creating serializable event, got %v", err)
	}

	if serializableEvent.EventID() != testEvent.EventID() {
		t.Errorf("Expected event ID %s, got %s", testEvent.EventID(), serializableEvent.EventID())
	}

	if serializableEvent.EventType() != testEvent.EventType() {
		t.Errorf("Expected event type %s, got %s", testEvent.EventType(), serializableEvent.EventType())
	}

	if serializableEvent.AggregateID() != testEvent.AggregateID() {
		t.Errorf("Expected aggregate ID %s, got %s", testEvent.AggregateID(), serializableEvent.AggregateID())
	}
}

func TestSerializableEvent_JSONSerialization(t *testing.T) {
	testEvent := NewTestEvent("test-id", "test-data")

	serializableEvent, err := NewSerializableEvent(testEvent)
	if err != nil {
		t.Fatalf("Expected no error creating serializable event, got %v", err)
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(serializableEvent)
	if err != nil {
		t.Fatalf("Expected no error marshaling to JSON, got %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled SerializableEvent
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Expected no error unmarshaling from JSON, got %v", err)
	}

	if unmarshaled.EventID() != serializableEvent.EventID() {
		t.Errorf("Expected event ID %s after unmarshal, got %s",
			serializableEvent.EventID(), unmarshaled.EventID())
	}

	if unmarshaled.EventType() != serializableEvent.EventType() {
		t.Errorf("Expected event type %s after unmarshal, got %s",
			serializableEvent.EventType(), unmarshaled.EventType())
	}
}

func TestSerializableEventEnvelope_Creation(t *testing.T) {
	testEvent := NewTestEvent("test-id", "test-data")

	envelope, err := NewSerializableEventEnvelope(testEvent)
	if err != nil {
		t.Fatalf("Expected no error creating envelope, got %v", err)
	}

	if envelope.Event == nil {
		t.Fatal("Expected event to be set in envelope")
	}

	if envelope.Event.EventID() != testEvent.EventID() {
		t.Errorf("Expected event ID %s in envelope, got %s",
			testEvent.EventID(), envelope.Event.EventID())
	}

	if envelope.MaxRetry != 3 {
		t.Errorf("Expected max retry 3, got %d", envelope.MaxRetry)
	}

	if envelope.Retry != 0 {
		t.Errorf("Expected retry count 0, got %d", envelope.Retry)
	}
}

func TestSerializableEventEnvelope_RetryLogic(t *testing.T) {
	testEvent := NewTestEvent("test-id", "test-data")

	envelope, err := NewSerializableEventEnvelope(testEvent)
	if err != nil {
		t.Fatalf("Expected no error creating envelope, got %v", err)
	}

	// Should be able to retry initially
	if !envelope.ShouldRetry() {
		t.Error("Expected envelope to allow retry initially")
	}

	// Increment retry count
	envelope.IncrementRetry()
	if envelope.Retry != 1 {
		t.Errorf("Expected retry count 1 after increment, got %d", envelope.Retry)
	}

	// Should still be able to retry
	if !envelope.ShouldRetry() {
		t.Error("Expected envelope to allow retry after first increment")
	}

	// Increment to max retries
	envelope.IncrementRetry()
	envelope.IncrementRetry()

	if envelope.Retry != 3 {
		t.Errorf("Expected retry count 3, got %d", envelope.Retry)
	}

	// Should not be able to retry anymore
	if envelope.ShouldRetry() {
		t.Error("Expected envelope to not allow retry after max retries")
	}
}

func TestSerializableEventEnvelope_JSONSerialization(t *testing.T) {
	testEvent := NewTestEvent("test-id", "test-data")

	envelope, err := NewSerializableEventEnvelope(testEvent)
	if err != nil {
		t.Fatalf("Expected no error creating envelope, got %v", err)
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("Expected no error marshaling envelope to JSON, got %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled SerializableEventEnvelope
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Expected no error unmarshaling envelope from JSON, got %v", err)
	}

	if unmarshaled.Event.EventID() != envelope.Event.EventID() {
		t.Errorf("Expected event ID %s after unmarshal, got %s",
			envelope.Event.EventID(), unmarshaled.Event.EventID())
	}

	if unmarshaled.MaxRetry != envelope.MaxRetry {
		t.Errorf("Expected max retry %d after unmarshal, got %d",
			envelope.MaxRetry, unmarshaled.MaxRetry)
	}
}

// Benchmark tests
func BenchmarkSerializableEvent_Creation(b *testing.B) {
	testEvent := NewTestEvent("test-id", "test-data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewSerializableEvent(testEvent)
		if err != nil {
			b.Fatalf("Error creating serializable event: %v", err)
		}
	}
}

func BenchmarkSerializableEvent_JSONMarshal(b *testing.B) {
	testEvent := NewTestEvent("test-id", "test-data")
	serializableEvent, _ := NewSerializableEvent(testEvent)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(serializableEvent)
		if err != nil {
			b.Fatalf("Error marshaling event: %v", err)
		}
	}
}

func BenchmarkEventEnvelope_Creation(b *testing.B) {
	testEvent := NewTestEvent("test-id", "test-data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewSerializableEventEnvelope(testEvent)
		if err != nil {
			b.Fatalf("Error creating envelope: %v", err)
		}
	}
}
