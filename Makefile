.PHONY: help build run test clean docker-up docker-down install-deps templ-generate migrate-up migrate-down migrate-version db-health setup dev dev-setup db-seed lint fmt check-deps watch-css dev-status

# Default target
help:
	@echo "Available commands:"
	@echo ""
	@echo "Development:"
	@echo "  setup          - Complete development environment setup"
	@echo "  dev            - Start development server with hot reload"
	@echo "  dev-setup      - Run development setup script"
	@echo "  dev-status     - Check development environment status"
	@echo "  db-seed        - Seed database with development data"
	@echo "  watch-css      - Watch and build CSS files"
	@echo ""
	@echo "Building:"
	@echo "  build          - Build the application"
	@echo "  run            - Run the application"
	@echo "  templ-generate - Generate Go code from Templ files"
	@echo ""
	@echo "Testing & Quality:"
	@echo "  test           - Run tests"
	@echo "  test-verbose   - Run tests with verbose output"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  lint           - Run linter"
	@echo "  fmt            - Format code"
	@echo ""
	@echo "Database:"
	@echo "  migrate-up     - Run database migrations up"
	@echo "  migrate-down   - Run database migrations down"
	@echo "  migrate-version - Show current migration version"
	@echo "  db-health      - Check database health"
	@echo ""
	@echo "Docker:"
	@echo "  docker-up      - Start Docker services"
	@echo "  docker-down    - Stop Docker services"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run with Docker"
	@echo ""
	@echo "Utilities:"
	@echo "  install-deps   - Install dependencies"
	@echo "  check-deps     - Check if dependencies are installed"
	@echo "  clean          - Clean build artifacts"

# Check if dependencies are installed
check-deps:
	@echo "Checking dependencies..."
	@command -v go >/dev/null 2>&1 || { echo "Go is not installed. Please install Go 1.21+"; exit 1; }
	@command -v docker >/dev/null 2>&1 || { echo "Docker is not installed. Please install Docker"; exit 1; }
	@command -v docker-compose >/dev/null 2>&1 || docker compose version >/dev/null 2>&1 || { echo "Docker Compose is not installed"; exit 1; }
	@echo "All dependencies are available"

# Install dependencies
install-deps: check-deps
	@echo "Installing Go dependencies..."
	go mod tidy
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/cosmtrek/air@latest
	@echo "Dependencies installed successfully"

# Generate Templ files
templ-generate:
	templ generate

# Build the application
build: templ-generate
	go build -o bin/server ./cmd/server

# Run the application
run: templ-generate
	go run ./cmd/server

# Run tests
test:
	go test ./...

# Run tests with verbose output
test-verbose:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -cover ./...
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format code
fmt:
	go fmt ./...
	@if command -v templ >/dev/null 2>&1; then \
		templ fmt .; \
	fi

# Run linter
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		go vet ./...; \
	fi

# Check development environment status
dev-status:
	@if [ -f "scripts/dev-status.sh" ]; then \
		chmod +x scripts/dev-status.sh && ./scripts/dev-status.sh; \
	else \
		echo "Development status script not found"; \
	fi

# Watch and build CSS files (if using Tailwind)
watch-css:
	@if [ -f "tailwind.config.js" ] && command -v npx >/dev/null 2>&1; then \
		echo "Watching CSS files..."; \
		npx tailwindcss -i ./web/static/css/input.css -o ./web/static/css/output.css --watch; \
	else \
		echo "Tailwind CSS not configured or npx not available"; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf tmp/
	rm -f *_templ.go
	rm -f coverage.out coverage.html
	@echo "Clean completed"

# Start Docker services
docker-up:
	docker-compose up -d

# Stop Docker services
docker-down:
	docker-compose down

# Complete development environment setup
setup: dev-setup

# Run development setup script
dev-setup:
	@if [ -f "scripts/dev-setup.sh" ]; then \
		chmod +x scripts/dev-setup.sh && ./scripts/dev-setup.sh; \
	elif [ -f "scripts/dev-setup.ps1" ]; then \
		powershell -ExecutionPolicy Bypass -File scripts/dev-setup.ps1; \
	else \
		echo "No setup script found. Running basic setup..."; \
		$(MAKE) install-deps && $(MAKE) docker-up && $(MAKE) migrate-up && $(MAKE) templ-generate; \
	fi

# Seed database with development data
db-seed:
	@if [ -f "scripts/db-seed.sh" ]; then \
		chmod +x scripts/db-seed.sh && ./scripts/db-seed.sh; \
	elif [ -f "scripts/db-seed.ps1" ]; then \
		powershell -ExecutionPolicy Bypass -File scripts/db-seed.ps1; \
	else \
		echo "No database seeding script found"; \
	fi

# Start development environment with hot reload
dev: docker-up
	@echo "Starting development server with hot reload..."
	@echo "The application will be available at http://localhost:8080"
	@echo "Press Ctrl+C to stop"
	air

# Build and run with Docker
docker-build:
	docker-compose build

docker-run: docker-build
	docker-compose up

# Database migration commands
migrate-up:
	go run ./cmd/migrate -action=up

migrate-down:
	go run ./cmd/migrate -action=down

migrate-version:
	go run ./cmd/migrate -action=version

migrate-steps:
	go run ./cmd/migrate -action=steps -steps=$(STEPS)

# Database health check
db-health:
	go run ./cmd/dbhealth

db-health-json:
	go run ./cmd/dbhealth -format=json

db-wait:
	go run ./cmd/dbhealth -wait -timeout=30s