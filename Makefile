.PHONY: help build run test clean docker-up docker-down install-deps templ-generate migrate-up migrate-down migrate-version db-health

# Default target
help:
	@echo "Available commands:"
	@echo "  build          - Build the application"
	@echo "  run            - Run the application"
	@echo "  test           - Run tests"
	@echo "  clean          - Clean build artifacts"
	@echo "  docker-up      - Start Docker services"
	@echo "  docker-down    - Stop Docker services"
	@echo "  install-deps   - Install dependencies"
	@echo "  templ-generate - Generate Go code from Templ files"
	@echo "  migrate-up     - Run database migrations up"
	@echo "  migrate-down   - Run database migrations down"
	@echo "  migrate-version - Show current migration version"
	@echo "  db-health      - Check database health"

# Install dependencies
install-deps:
	go mod tidy
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/cosmtrek/air@latest

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

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf tmp/
	rm -f *_templ.go

# Start Docker services
docker-up:
	docker-compose up -d

# Stop Docker services
docker-down:
	docker-compose down

# Start development environment
dev: docker-up
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