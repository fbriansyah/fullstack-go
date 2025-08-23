# Technology Stack

## Backend
- **Language**: Go 1.21+
- **Web Framework**: Gin, Echo, or standard net/http
- **Database**: PostgreSQL with SQLx
- **Message Broker**: RabbitMQ for event-driven communication
- **Authentication**: Session-based auth with secure cookies
- **Testing**: Go's built-in testing package + testify

## Frontend
- **Template Engine**: Templ (https://github.com/a-h/templ) - Type-safe HTML templates in Go
- **Styling**: Tailwind CSS or vanilla CSS
- **JavaScript**: Minimal vanilla JS for interactivity (HTMX integration recommended)
- **Build**: Templ generates Go code from .templ files

## Development Tools
- **Package Management**: Go modules (`go.mod`)
- **Code Formatting**: `gofmt`, `goimports`
- **Linting**: `golangci-lint`
- **Hot Reload**: Air for Go backend

## Common Commands

### Backend Development
```bash
# Initialize Go module
go mod init [module-name]

# Install Templ CLI
go install github.com/a-h/templ/cmd/templ@latest

# Install dependencies
go mod tidy

# Generate Go code from .templ files
templ generate

# Run the application
go run main.go

# Build for production
templ generate && go build -o bin/app

# Run tests
go test ./...

# Format code
go fmt ./...
templ fmt .

# Install Air for hot reload with Templ
go install github.com/cosmtrek/air@latest
air
```

### Database
```bash
# Run migrations (if using migrate tool)
migrate -path ./migrations -database "postgres://..." up

# Generate models (if using sqlc)
sqlc generate
```

### Docker
```bash
# Build container
docker build -t fullstack-go .

# Run with docker-compose
docker-compose up -d
```