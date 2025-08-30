#!/bin/bash

# Development Environment Setup Script
# This script sets up the complete development environment

set -e

echo "🚀 Setting up Go Templ Template development environment..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if required tools are installed
check_dependencies() {
    print_status "Checking dependencies..."
    
    # Check Go
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.21+ first."
        exit 1
    fi
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        print_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    print_success "All dependencies are available"
}

# Install Go dependencies
install_go_deps() {
    print_status "Installing Go dependencies..."
    
    go mod tidy
    
    # Install Templ CLI
    print_status "Installing Templ CLI..."
    go install github.com/a-h/templ/cmd/templ@latest
    
    # Install Air for hot reload
    print_status "Installing Air for hot reload..."
    go install github.com/cosmtrek/air@latest
    
    print_success "Go dependencies installed"
}

# Setup environment file
setup_env() {
    print_status "Setting up environment configuration..."
    
    if [ ! -f .env ]; then
        if [ -f .env.example ]; then
            cp .env.example .env
            print_success "Created .env file from .env.example"
        else
            print_warning ".env.example not found, creating basic .env file"
            cat > .env << EOF
# Server Configuration
SERVER_PORT=8080
SERVER_HOST=localhost
ENVIRONMENT=development

# Database Configuration
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=go_templ_template
DATABASE_USER=postgres
DATABASE_PASSWORD=postgres
DATABASE_SSL_MODE=disable

# RabbitMQ Configuration
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
RABBITMQ_EXCHANGE=go_templ_events
RABBITMQ_QUEUE_PREFIX=go_templ

# Session Configuration
SESSION_SECRET=your-secret-key-change-in-production
SESSION_MAX_AGE=86400

# Development
DEBUG=true
LOG_LEVEL=debug
EOF
        fi
    else
        print_warning ".env file already exists, skipping creation"
    fi
}

# Start Docker services
start_services() {
    print_status "Starting Docker services..."
    
    # Stop any existing services
    docker-compose down 2>/dev/null || true
    
    # Start services
    docker-compose up -d
    
    print_status "Waiting for services to be ready..."
    sleep 5
    
    # Wait for PostgreSQL to be ready
    print_status "Waiting for PostgreSQL to be ready..."
    timeout=30
    while ! docker-compose exec -T postgres pg_isready -U postgres &>/dev/null; do
        timeout=$((timeout - 1))
        if [ $timeout -eq 0 ]; then
            print_error "PostgreSQL failed to start within 30 seconds"
            exit 1
        fi
        sleep 1
    done
    
    print_success "Docker services are running"
}

# Run database migrations
run_migrations() {
    print_status "Running database migrations..."
    
    # Wait a bit more for database to be fully ready
    sleep 2
    
    # Run migrations
    if make migrate-up; then
        print_success "Database migrations completed"
    else
        print_error "Database migrations failed"
        exit 1
    fi
}

# Generate Templ files
generate_templ() {
    print_status "Generating Templ files..."
    
    if make templ-generate; then
        print_success "Templ files generated"
    else
        print_error "Templ generation failed"
        exit 1
    fi
}

# Build the application
build_app() {
    print_status "Building the application..."
    
    if make build; then
        print_success "Application built successfully"
    else
        print_error "Application build failed"
        exit 1
    fi
}

# Main setup process
main() {
    echo "=================================================="
    echo "  Go Templ Template Development Setup"
    echo "=================================================="
    echo
    
    check_dependencies
    install_go_deps
    setup_env
    start_services
    run_migrations
    generate_templ
    build_app
    
    echo
    echo "=================================================="
    print_success "Development environment setup complete!"
    echo "=================================================="
    echo
    echo "Next steps:"
    echo "  1. Start development server: make dev"
    echo "  2. Open your browser: http://localhost:8080"
    echo "  3. Start coding!"
    echo
    echo "Useful commands:"
    echo "  make help          - Show all available commands"
    echo "  make dev           - Start development with hot reload"
    echo "  make test          - Run tests"
    echo "  make docker-down   - Stop Docker services"
    echo
}

# Run main function
main "$@"