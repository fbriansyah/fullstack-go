# Design Document

## Overview

This template project creates a modern fullstack Go web application using Templ for type-safe HTML templating. The architecture follows clean code principles with clear separation between HTTP handlers, business logic, data access, and presentation layers. Templ provides compile-time safety for HTML templates while maintaining the simplicity and performance of Go.

## Architecture

### Modular Monolith with Event-Driven Architecture

```mermaid
graph TB
    Client[Web Browser] --> Router[Gin Router]
    Router --> Auth[Auth Middleware]
    Router --> Handlers[HTTP Handlers]
    
    subgraph "User Module"
        UH[User Handlers] --> US[User Service]
        US --> UR[User Repository]
        US --> UE[User Events]
    end
    
    subgraph "Auth Module"
        AH[Auth Handlers] --> AS[Auth Service]
        AS --> AR[Auth Repository]
        AS --> AE[Auth Events]
    end
    
    subgraph "Event Infrastructure"
        EB[Event Bus] --> RMQ[RabbitMQ]
        UE --> EB
        AE --> EB
        EB --> EH[Event Handlers]
    end
    
    UR --> DB[(Database)]
    AR --> DB
    Handlers --> Templates[Templ Components]
    Templates --> Static[Static Assets]
```

### Modular Architecture Principles

- **Module Boundaries**: Each module (User, Auth, Notification, etc.) is self-contained
- **Event-Driven Communication**: Modules communicate through domain events via RabbitMQ
- **Shared Kernel**: Common infrastructure (database, event bus, HTTP) shared across modules
- **Module Independence**: Each module can be developed, tested, and deployed independently
- **Single Database**: Shared database with module-specific schemas/tables

## Components and Interfaces

### Core Components

#### 1. Application Bootstrap (`cmd/server/main.go`)
```go
type App struct {
    Router    *gin.Engine
    DB        *sqlx.DB
    Config    *config.Config
    EventBus  *events.EventBus
    Modules   []Module
}

type Module interface {
    Name() string
    Initialize(app *App) error
    RegisterRoutes(router *gin.RouterGroup)
    RegisterEventHandlers(bus *events.EventBus)
}
```

#### 2. Event Infrastructure (`internal/events/`)
```go
type EventBus interface {
    Publish(event DomainEvent) error
    Subscribe(eventType string, handler EventHandler) error
    Start() error
    Stop() error
}

type DomainEvent interface {
    EventType() string
    AggregateID() string
    OccurredAt() time.Time
    EventData() interface{}
}

type EventHandler interface {
    Handle(event DomainEvent) error
    EventType() string
}
```

#### 3. User Module (`internal/modules/user/`)
```go
type UserService interface {
    CreateUser(cmd CreateUserCommand) (*User, error)
    GetUser(id string) (*User, error)
    UpdateUser(cmd UpdateUserCommand) error
}

type UserEvents struct {
    UserCreated   UserCreatedEvent
    UserUpdated   UserUpdatedEvent
    UserDeleted   UserDeletedEvent
}
```

#### 4. Auth Module (`internal/modules/auth/`)
```go
type AuthService interface {
    Login(cmd LoginCommand) (*Session, error)
    Register(cmd RegisterCommand) (*User, error)
    ValidateSession(sessionID string) (*User, error)
    Logout(sessionID string) error
}

type AuthEvents struct {
    UserLoggedIn    UserLoggedInEvent
    UserRegistered  UserRegisteredEvent
    UserLoggedOut   UserLoggedOutEvent
}
```

#### 5. Configuration Management (`internal/config/`)
```go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    RabbitMQ RabbitMQConfig
    Modules  ModulesConfig
}

type RabbitMQConfig struct {
    URL          string
    Exchange     string
    QueuePrefix  string
    Durable      bool
}
```

### Templ Component Structure

#### Base Layout (`web/templates/layouts/base.templ`)
```templ
templ Base(title string, content templ.Component) {
    <!DOCTYPE html>
    <html lang="en">
        <head>
            <meta charset="UTF-8"/>
            <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
            <title>{title}</title>
            <link href="/static/css/tailwind.css" rel="stylesheet"/>
        </head>
        <body class="bg-gray-50">
            @Header()
            <main class="container mx-auto px-4 py-8">
                @content
            </main>
            @Footer()
        </body>
    </html>
}
```

#### Component Hierarchy
- **Layouts**: Base layout, authenticated layout, public layout
- **Components**: Header, footer, navigation, forms, buttons, cards
- **Pages**: Home, login, register, dashboard, profile

## Data Models

### Domain Models

#### User Aggregate (`internal/modules/user/domain/`)
```go
type User struct {
    ID        string    `db:"id" json:"id"`
    Email     string    `db:"email" json:"email"`
    Password  string    `db:"password" json:"-"`
    FirstName string    `db:"first_name" json:"first_name"`
    LastName  string    `db:"last_name" json:"last_name"`
    Status    UserStatus `db:"status" json:"status"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
    UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
    Version   int       `db:"version" json:"version"` // Optimistic locking
}

type UserStatus string
const (
    UserStatusActive   UserStatus = "active"
    UserStatusInactive UserStatus = "inactive"
    UserStatusSuspended UserStatus = "suspended"
)
```

#### Session Aggregate (`internal/modules/auth/domain/`)
```go
type Session struct {
    ID        string    `db:"id" json:"id"`
    UserID    string    `db:"user_id" json:"user_id"`
    ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
    IPAddress string    `db:"ip_address" json:"ip_address"`
    UserAgent string    `db:"user_agent" json:"user_agent"`
}
```

#### Domain Events
```go
type UserCreatedEvent struct {
    UserID    string    `json:"user_id"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

type UserLoggedInEvent struct {
    UserID    string    `json:"user_id"`
    SessionID string    `json:"session_id"`
    IPAddress string    `json:"ip_address"`
    LoginAt   time.Time `json:"login_at"`
}
```

## Error Handling

### Error Types
```go
type AppError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Type    string `json:"type"`
}

const (
    ErrTypeValidation = "validation"
    ErrTypeAuth      = "authentication"
    ErrTypeNotFound  = "not_found"
    ErrTypeInternal  = "internal"
)
```

### Error Pages
- Custom Templ components for 404, 500, and authentication errors
- Consistent error styling and user-friendly messages
- Proper HTTP status codes and logging

## Testing Strategy

### Unit Testing
- Service layer tests with mocked dependencies
- Repository tests with test database
- Templ component rendering tests
- Authentication flow tests

### Integration Testing
- HTTP endpoint tests using httptest
- Database integration tests
- Template rendering integration tests
- Authentication middleware tests

### Test Structure
```
tests/
├── unit/
│   ├── services/
│   ├── repository/
│   └── auth/
├── integration/
│   ├── api/
│   └── templates/
└── fixtures/
    └── test_data.sql
```

## Development Workflow

### Hot Reload Configuration (`.air.toml`)
```toml
[build]
  cmd = "templ generate && go build -o ./tmp/main ./cmd/server"
  bin = "tmp/main"
  full_bin = "./tmp/main"
  include_ext = ["go", "templ"]
  exclude_dir = ["tmp", "vendor", "node_modules"]

[misc]
  clean_on_exit = true
```

### Build Process
1. Generate Go code from Templ files (`templ generate`)
2. Build CSS with Tailwind (`tailwindcss build`)
3. Compile Go application
4. Run with Air for development or binary for production

## Security Considerations

### Authentication
- Secure password hashing using bcrypt
- Session-based authentication with secure cookies
- CSRF protection for forms
- Session timeout and cleanup

### Input Validation
- Server-side validation for all forms
- SQL injection prevention through SQLx prepared statements
- XSS prevention through Templ's automatic escaping
- Rate limiting for authentication endpoints

### Configuration Security
- Environment-based configuration
- Secure defaults for production
- Database connection security
- HTTPS enforcement in production

## Deployment Architecture

### Docker Configuration
```dockerfile
FROM golang:1.21-alpine AS builder
RUN go install github.com/a-h/templ/cmd/templ@latest
COPY . .
RUN templ generate && go build -o app ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/app .
COPY --from=builder /app/web/static ./web/static
CMD ["./app"]
```

### Docker Compose for Development
```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - rabbitmq
    environment:
      - DATABASE_URL=postgres://user:pass@postgres:5432/appdb
      - RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: appdb
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass

  rabbitmq:
    image: rabbitmq:3-management-alpine
    ports:
      - "15672:15672"  # Management UI
    environment:
      RABBITMQ_DEFAULT_USER: guest
      RABBITMQ_DEFAULT_PASS: guest
```

### Module Structure
```
internal/
├── modules/
│   ├── user/
│   │   ├── domain/          # Domain models and events
│   │   ├── application/     # Use cases and commands
│   │   ├── infrastructure/  # Repository implementations
│   │   ├── handlers/        # HTTP handlers
│   │   └── module.go        # Module registration
│   ├── auth/
│   │   ├── domain/
│   │   ├── application/
│   │   ├── infrastructure/
│   │   ├── handlers/
│   │   └── module.go
│   └── notification/        # Future module
├── shared/
│   ├── events/             # Event bus implementation
│   ├── database/           # Database utilities
│   ├── middleware/         # Shared middleware
│   └── errors/             # Error handling
└── config/                 # Configuration management
```