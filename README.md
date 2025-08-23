# Go Templ Template

A comprehensive fullstack Go application template using Templ for type-safe HTML templating, featuring modular architecture, event-driven communication, and modern development practices.

## Features

- **Backend**: Go with Gin/Echo framework
- **Frontend**: Templ for type-safe HTML templates
- **Database**: PostgreSQL with SQLx
- **Message Broker**: RabbitMQ for event-driven architecture
- **Authentication**: Session-based auth with secure cookies
- **Styling**: Tailwind CSS
- **Development**: Hot reload with Air
- **Containerization**: Docker and Docker Compose

## Quick Start

### Prerequisites

- Go 1.21+
- Docker and Docker Compose
- Make (optional, for convenience commands)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd go-templ-template
```

2. Copy environment file:
```bash
cp .env.example .env
```

3. Install dependencies:
```bash
make install-deps
# or manually:
go mod tidy
go install github.com/a-h/templ/cmd/templ@latest
go install github.com/cosmtrek/air@latest
```

4. Start services:
```bash
make docker-up
```

5. Run the application:
```bash
make run
```

The application will be available at `http://localhost:8080`

## Development

### Hot Reload Development

Start the development environment with hot reload:

```bash
make dev
```

This will:
- Start PostgreSQL and RabbitMQ services
- Watch for file changes and automatically restart the server
- Regenerate Templ files on changes

### Available Commands

```bash
make help              # Show all available commands
make build             # Build the application
make run               # Run the application
make test              # Run tests
make clean             # Clean build artifacts
make docker-up         # Start Docker services
make docker-down       # Stop Docker services
make templ-generate    # Generate Go code from Templ files
```

### Project Structure

```
/
├── cmd/server/         # Application entry point
├── internal/           # Private application code
│   ├── config/        # Configuration management
│   ├── modules/       # Business modules (user, auth, etc.)
│   └── shared/        # Shared infrastructure
├── web/               # Frontend assets and templates
├── migrations/        # Database migrations
├── docker/           # Docker configuration files
└── tests/            # Test files
```

## Configuration

The application uses environment variables for configuration. See `.env.example` for all available options.

Key configuration areas:
- **Server**: Port, host, environment
- **Database**: PostgreSQL connection settings
- **RabbitMQ**: Message broker configuration

## Services

### PostgreSQL
- **Port**: 5432
- **Database**: go_templ_template
- **User/Password**: postgres/postgres

### RabbitMQ
- **AMQP Port**: 5672
- **Management UI**: http://localhost:15672
- **User/Password**: guest/guest

## Architecture

This template follows a modular monolith architecture with:

- **Event-Driven Communication**: Modules communicate via RabbitMQ events
- **Clean Architecture**: Clear separation of concerns
- **Domain-Driven Design**: Business logic organized in domain modules
- **Type-Safe Templates**: Templ provides compile-time safety for HTML

## Next Steps

1. Implement your business modules in `internal/modules/`
2. Create Templ components in `web/templates/`
3. Add database migrations in `migrations/`
4. Configure authentication and authorization
5. Add your business logic and API endpoints

## License

[Your License Here]