#!/bin/bash

# Development Environment Status Check
# This script checks the status of all development services and tools

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

check_command() {
    local cmd="$1"
    local name="$2"
    
    if command -v "$cmd" &> /dev/null; then
        local version=$($cmd --version 2>/dev/null | head -n1 || echo "unknown")
        print_success "$name is installed ($version)"
        return 0
    else
        print_error "$name is not installed"
        return 1
    fi
}

check_docker_service() {
    local service="$1"
    
    if docker-compose ps "$service" 2>/dev/null | grep -q "Up"; then
        print_success "Docker service '$service' is running"
        return 0
    else
        print_error "Docker service '$service' is not running"
        return 1
    fi
}

check_port() {
    local port="$1"
    local service="$2"
    
    if nc -z localhost "$port" 2>/dev/null; then
        print_success "$service is accessible on port $port"
        return 0
    else
        print_error "$service is not accessible on port $port"
        return 1
    fi
}

main() {
    echo "=================================================="
    echo "  Development Environment Status Check"
    echo "=================================================="
    echo
    
    print_status "Checking required tools..."
    echo
    
    # Check required tools
    local tools_ok=true
    check_command "go" "Go" || tools_ok=false
    check_command "docker" "Docker" || tools_ok=false
    check_command "docker-compose" "Docker Compose" || tools_ok=false
    
    echo
    print_status "Checking optional tools..."
    echo
    
    # Check optional tools
    check_command "make" "Make" || print_warning "Make is recommended for easier development"
    check_command "templ" "Templ CLI" || print_warning "Run 'go install github.com/a-h/templ/cmd/templ@latest'"
    check_command "air" "Air (hot reload)" || print_warning "Run 'go install github.com/cosmtrek/air@latest'"
    check_command "golangci-lint" "golangci-lint" || print_warning "Install for better code linting"
    
    echo
    print_status "Checking project files..."
    echo
    
    # Check important files
    [ -f ".env" ] && print_success ".env file exists" || print_warning ".env file missing (copy from .env.example)"
    [ -f "go.mod" ] && print_success "go.mod exists" || print_error "go.mod missing"
    [ -f ".air.toml" ] && print_success "Air configuration exists" || print_error "Air configuration missing"
    [ -f "docker-compose.yml" ] && print_success "Docker Compose configuration exists" || print_error "Docker Compose configuration missing"
    
    echo
    print_status "Checking Docker services..."
    echo
    
    # Check if Docker is running
    if ! docker info &> /dev/null; then
        print_error "Docker daemon is not running"
        echo
        echo "Please start Docker and run this script again."
        exit 1
    fi
    
    print_success "Docker daemon is running"
    
    # Check Docker services
    local services_ok=true
    check_docker_service "postgres" || services_ok=false
    check_docker_service "rabbitmq" || services_ok=false
    
    if [ "$services_ok" = false ]; then
        echo
        print_warning "Some Docker services are not running. Start them with: make docker-up"
    fi
    
    echo
    print_status "Checking service connectivity..."
    echo
    
    # Check service ports
    check_port "5432" "PostgreSQL"
    check_port "5672" "RabbitMQ"
    check_port "15672" "RabbitMQ Management"
    
    # Check if application is running
    if check_port "8080" "Application Server"; then
        print_success "Application is running at http://localhost:8080"
    else
        print_warning "Application is not running. Start with: make dev"
    fi
    
    echo
    print_status "Checking database connectivity..."
    echo
    
    # Check database health
    if make db-health &> /dev/null; then
        print_success "Database is healthy and accessible"
    else
        print_error "Database health check failed"
    fi
    
    echo
    print_status "Checking Go modules..."
    echo
    
    # Check Go modules
    if go mod verify &> /dev/null; then
        print_success "Go modules are valid"
    else
        print_warning "Go modules need attention. Run: go mod tidy"
    fi
    
    # Check if Templ files are generated
    if find . -name "*_templ.go" | grep -q .; then
        print_success "Templ files are generated"
    else
        print_warning "Templ files not generated. Run: make templ-generate"
    fi
    
    echo
    echo "=================================================="
    
    if [ "$tools_ok" = true ] && [ "$services_ok" = true ]; then
        print_success "Development environment is ready!"
        echo
        echo "Next steps:"
        echo "  • Start development: make dev"
        echo "  • Seed database: make db-seed"
        echo "  • Run tests: make test"
    else
        print_warning "Development environment needs attention"
        echo
        echo "To fix issues:"
        echo "  • Install missing tools"
        echo "  • Start Docker services: make docker-up"
        echo "  • Run setup: make setup"
    fi
    
    echo "=================================================="
}

main "$@"