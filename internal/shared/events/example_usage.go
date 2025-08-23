package events

import (
	"context"
	"fmt"
	"log"
	"time"
)

// ExampleUserCreatedEvent demonstrates a concrete domain event
type ExampleUserCreatedEvent struct {
	*BaseEvent
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

// NewExampleUserCreatedEvent creates a new user created event
func NewExampleUserCreatedEvent(userID, email string) *ExampleUserCreatedEvent {
	return &ExampleUserCreatedEvent{
		BaseEvent: NewBaseEvent(
			UserCreatedEventType,
			userID,
			UserAggregateType,
			map[string]interface{}{
				"user_id": userID,
				"email":   email,
			},
		),
		UserID: userID,
		Email:  email,
	}
}

// ExampleEmailNotificationHandler demonstrates an event handler
type ExampleEmailNotificationHandler struct {
	name string
}

// NewExampleEmailNotificationHandler creates a new email notification handler
func NewExampleEmailNotificationHandler() *ExampleEmailNotificationHandler {
	return &ExampleEmailNotificationHandler{
		name: "email-notification-handler",
	}
}

// Handle processes user created events to send welcome emails
func (h *ExampleEmailNotificationHandler) Handle(ctx context.Context, event DomainEvent) error {
	// In a real implementation, this would send an email
	log.Printf("Sending welcome email for user: %s", event.AggregateID())

	// Simulate some processing time
	time.Sleep(10 * time.Millisecond)

	return nil
}

// EventType returns the event type this handler processes
func (h *ExampleEmailNotificationHandler) EventType() string {
	return UserCreatedEventType
}

// HandlerName returns the handler name
func (h *ExampleEmailNotificationHandler) HandlerName() string {
	return h.name
}

// ExampleUsage demonstrates how to use the RabbitMQ event bus
func ExampleUsage() error {
	// Create configuration
	config := DefaultRabbitMQConfig()
	config.QueuePrefix = "example-app"

	// Validate configuration
	if err := ValidateRabbitMQConfig(config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Create event bus
	factory := NewEventBusFactory()
	bus := factory.CreateRabbitMQEventBus(config)

	ctx := context.Background()

	// Start the event bus
	if err := bus.Start(ctx); err != nil {
		return fmt.Errorf("failed to start event bus: %w", err)
	}
	defer bus.Stop(ctx)

	// Create and subscribe event handlers
	emailHandler := NewExampleEmailNotificationHandler()
	if err := bus.Subscribe(UserCreatedEventType, emailHandler); err != nil {
		return fmt.Errorf("failed to subscribe handler: %w", err)
	}

	// Give some time for consumers to be set up
	time.Sleep(100 * time.Millisecond)

	// Create and publish events
	userEvent := NewExampleUserCreatedEvent("user-123", "user@example.com")
	if err := bus.Publish(ctx, userEvent); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Println("Event published successfully")

	// In a real application, you would keep the bus running
	// and handle graceful shutdown

	return nil
}

// ExampleWithEnvironmentConfig demonstrates using environment-based configuration
func ExampleWithEnvironmentConfig() error {
	// This would read from environment variables:
	// RABBITMQ_URL, RABBITMQ_EXCHANGE, etc.
	factory := NewEventBusFactory()
	bus := factory.CreateRabbitMQEventBusFromEnv()

	ctx := context.Background()

	// Start the event bus
	if err := bus.Start(ctx); err != nil {
		// In production, you might want to retry or use a different transport
		log.Printf("Failed to start RabbitMQ event bus: %v", err)
		return err
	}
	defer bus.Stop(ctx)

	// Rest of the application logic...

	return nil
}

// ExampleEventHandlerWithError demonstrates error handling in event handlers
type ExampleAuditHandler struct {
	name        string
	shouldError bool
}

func NewExampleAuditHandler() *ExampleAuditHandler {
	return &ExampleAuditHandler{
		name: "audit-handler",
	}
}

func (h *ExampleAuditHandler) Handle(ctx context.Context, event DomainEvent) error {
	if h.shouldError {
		return fmt.Errorf("simulated audit system error")
	}

	// Log the event for audit purposes
	log.Printf("AUDIT: Event %s occurred for aggregate %s at %s",
		event.EventType(),
		event.AggregateID(),
		event.OccurredAt().Format(time.RFC3339))

	return nil
}

func (h *ExampleAuditHandler) EventType() string {
	return UserCreatedEventType
}

func (h *ExampleAuditHandler) HandlerName() string {
	return h.name
}

// SetShouldError allows testing error scenarios
func (h *ExampleAuditHandler) SetShouldError(shouldError bool) {
	h.shouldError = shouldError
}
