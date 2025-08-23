# Events Package

This package provides event-driven architecture support for the Go Templ Template application using RabbitMQ as the message broker.

## Features

- **Event Bus Interface**: Clean abstraction for event publishing and subscription
- **RabbitMQ Implementation**: Production-ready event bus using RabbitMQ
- **Serializable Events**: JSON serialization support for event persistence and transport
- **Retry Logic**: Automatic retry mechanism for failed event processing
- **Type Safety**: Strongly typed events and handlers
- **Configuration**: Environment-based and programmatic configuration
- **Testing Support**: Comprehensive test suite with mocks and benchmarks

## Quick Start

### 1. Basic Usage

```go
package main

import (
    "context"
    "log"
    
    "go-templ-template/internal/shared/events"
)

func main() {
    // Create configuration
    config := events.DefaultRabbitMQConfig()
    
    // Create event bus
    factory := events.NewEventBusFactory()
    bus := factory.CreateRabbitMQEventBus(config)
    
    ctx := context.Background()
    
    // Start the event bus
    if err := bus.Start(ctx); err != nil {
        log.Fatal(err)
    }
    defer bus.Stop(ctx)
    
    // Create event handler
    handler := events.NewExampleEmailNotificationHandler()
    
    // Subscribe to events
    bus.Subscribe(events.UserCreatedEventType, handler)
    
    // Publish an event
    event := events.NewExampleUserCreatedEvent("user-123", "user@example.com")
    bus.Publish(ctx, event)
}
```

### 2. Environment Configuration

Set environment variables:

```bash
export RABBITMQ_URL="amqp://guest:guest@localhost:5672/"
export RABBITMQ_EXCHANGE="events"
export RABBITMQ_EXCHANGE_TYPE="topic"
export RABBITMQ_QUEUE_PREFIX="my-app"
export RABBITMQ_DURABLE="true"
```

Then use:

```go
factory := events.NewEventBusFactory()
bus := factory.CreateRabbitMQEventBusFromEnv()
```

## Creating Custom Events

### 1. Define Event Data Structure

```go
type UserRegisteredEvent struct {
    *events.BaseEvent
    UserID    string `json:"user_id"`
    Email     string `json:"email"`
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
}

func NewUserRegisteredEvent(userID, email, firstName, lastName string) *UserRegisteredEvent {
    return &UserRegisteredEvent{
        BaseEvent: events.NewBaseEvent(
            "user.registered",
            userID,
            "User",
            map[string]interface{}{
                "user_id":    userID,
                "email":      email,
                "first_name": firstName,
                "last_name":  lastName,
            },
        ),
        UserID:    userID,
        Email:     email,
        FirstName: firstName,
        LastName:  lastName,
    }
}
```

### 2. Create Event Handler

```go
type WelcomeEmailHandler struct {
    emailService EmailService
}

func NewWelcomeEmailHandler(emailService EmailService) *WelcomeEmailHandler {
    return &WelcomeEmailHandler{
        emailService: emailService,
    }
}

func (h *WelcomeEmailHandler) Handle(ctx context.Context, event events.DomainEvent) error {
    // Type assertion to get specific event data
    if userEvent, ok := event.(*UserRegisteredEvent); ok {
        return h.emailService.SendWelcomeEmail(ctx, userEvent.Email, userEvent.FirstName)
    }
    return nil
}

func (h *WelcomeEmailHandler) EventType() string {
    return "user.registered"
}

func (h *WelcomeEmailHandler) HandlerName() string {
    return "welcome-email-handler"
}
```

## Configuration Options

### RabbitMQ Configuration

```go
type RabbitMQConfig struct {
    URL          string // RabbitMQ connection URL
    Exchange     string // Exchange name
    ExchangeType string // Exchange type (topic, direct, fanout, headers)
    QueuePrefix  string // Prefix for queue names
    Durable      bool   // Whether exchanges and queues are durable
    AutoDelete   bool   // Whether to auto-delete when unused
    Exclusive    bool   // Whether queues are exclusive
    NoWait       bool   // Whether to wait for server confirmation
}
```

### Default Configuration

```go
config := events.DefaultRabbitMQConfig()
// URL: "amqp://guest:guest@localhost:5672/"
// Exchange: "events"
// ExchangeType: "topic"
// QueuePrefix: "go-templ-template"
// Durable: true
// AutoDelete: false
// Exclusive: false
// NoWait: false
```

## Event Types

The package includes predefined event types:

```go
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
```

## Error Handling and Retry Logic

The event bus automatically handles retries for failed event processing:

- **Max Retries**: 3 attempts by default
- **Retry Strategy**: Failed messages are requeued for retry
- **Dead Letter**: Messages exceeding max retries are discarded
- **Logging**: All errors are logged with context

## Testing

### Unit Tests

```bash
go test ./internal/shared/events/... -v
```

### Integration Tests (requires RabbitMQ)

```bash
go test ./internal/shared/events/... -v -tags=integration
```

### Benchmarks

```bash
go test ./internal/shared/events/... -bench=. -benchmem
```

## Production Considerations

### 1. Connection Management

- The event bus automatically reconnects on connection failures
- Use health checks to monitor connection status
- Implement graceful shutdown in your application

### 2. Performance

- Use connection pooling for high-throughput scenarios
- Monitor queue depths and consumer lag
- Consider message batching for bulk operations

### 3. Monitoring

- Monitor RabbitMQ metrics (queue depth, message rates, etc.)
- Log event processing times and error rates
- Set up alerts for connection failures

### 4. Security

- Use TLS for production connections
- Implement proper authentication and authorization
- Rotate credentials regularly

## Docker Compose Example

```yaml
version: '3.8'
services:
  rabbitmq:
    image: rabbitmq:3-management-alpine
    ports:
      - "5672:5672"
      - "15672:15672"
    environment:
      RABBITMQ_DEFAULT_USER: guest
      RABBITMQ_DEFAULT_PASS: guest
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq

volumes:
  rabbitmq_data:
```

## Contributing

When adding new event types or handlers:

1. Define event constants in `event_types.go`
2. Create concrete event structs that embed `BaseEvent`
3. Implement the `EventHandler` interface for handlers
4. Add comprehensive tests
5. Update this documentation