#!/bin/bash

# Migration Testing Script
# This script tests the migration and seeding functionality

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
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
else
    print_error ".env file not found. Please run dev-setup.sh first."
    exit 1
fi

# Test migration commands
test_migration_commands() {
    print_status "Testing migration commands..."
    
    # Test migration status
    print_status "Checking migration status..."
    go run ./cmd/migrate -action=status
    
    # Test migration list
    print_status "Listing migrations..."
    go run ./cmd/migrate -action=list
    
    # Test migration validation
    print_status "Validating migrations..."
    go run ./cmd/migrate -action=validate
    
    print_success "Migration commands tested successfully"
}

# Test migration workflow
test_migration_workflow() {
    print_status "Testing migration workflow..."
    
    # Start from clean state
    print_status "Rolling back all migrations..."
    go run ./cmd/migrate -action=down || print_warning "No migrations to roll back"
    
    # Check version (should be 0 or error)
    print_status "Checking version after rollback..."
    go run ./cmd/migrate -action=version || print_warning "No version (clean database)"
    
    # Run migrations up
    print_status "Running migrations up..."
    go run ./cmd/migrate -action=up
    
    # Check final version
    print_status "Checking final version..."
    go run ./cmd/migrate -action=version
    
    # Check status
    print_status "Checking final status..."
    go run ./cmd/migrate -action=status
    
    print_success "Migration workflow tested successfully"
}

# Test seeding
test_seeding() {
    print_status "Testing database seeding..."
    
    # Run seeding script
    if [ -f "scripts/db-seed.sh" ]; then
        chmod +x scripts/db-seed.sh
        ./scripts/db-seed.sh --clear
    else
        print_error "Seeding script not found"
        return 1
    fi
    
    print_success "Database seeding tested successfully"
}

# Test step-by-step migration
test_step_migration() {
    print_status "Testing step-by-step migration..."
    
    # Start from clean state
    go run ./cmd/migrate -action=down || print_warning "No migrations to roll back"
    
    # Run one step up
    print_status "Running one step up..."
    go run ./cmd/migrate -action=steps -steps=1
    go run ./cmd/migrate -action=version
    
    # Run another step up
    print_status "Running another step up..."
    go run ./cmd/migrate -action=steps -steps=1
    go run ./cmd/migrate -action=version
    
    # Run one step down
    print_status "Running one step down..."
    go run ./cmd/migrate -action=steps -steps=-1
    go run ./cmd/migrate -action=version
    
    print_success "Step-by-step migration tested successfully"
}

# Test migration to specific version
test_migrate_to() {
    print_status "Testing migration to specific version..."
    
    # Migrate to version 1
    print_status "Migrating to version 1..."
    go run ./cmd/migrate -action=to -version=1
    go run ./cmd/migrate -action=version
    
    # Migrate to version 2
    print_status "Migrating to version 2..."
    go run ./cmd/migrate -action=to -version=2
    go run ./cmd/migrate -action=version
    
    # Migrate back to version 1
    print_status "Migrating back to version 1..."
    go run ./cmd/migrate -action=to -version=1
    go run ./cmd/migrate -action=version
    
    print_success "Migration to specific version tested successfully"
}

# Test JSON output
test_json_output() {
    print_status "Testing JSON output..."
    
    print_status "Status in JSON format:"
    go run ./cmd/migrate -action=status -format=json
    
    print_status "Version in JSON format:"
    go run ./cmd/migrate -action=version -format=json
    
    print_status "List in JSON format:"
    go run ./cmd/migrate -action=list -format=json
    
    print_success "JSON output tested successfully"
}

# Run Go tests
test_go_tests() {
    print_status "Running Go migration tests..."
    
    # Set test database URL if not set
    if [ -z "$TEST_DATABASE_URL" ]; then
        export TEST_DATABASE_URL="$DATABASE_URL"
    fi
    
    # Run migration tests
    go test -v ./internal/shared/database -run TestMigration
    
    print_success "Go migration tests completed successfully"
}

# Main function
main() {
    echo "=================================================="
    echo "  Migration Testing for Go Templ Template"
    echo "=================================================="
    echo
    
    # Check if database is accessible
    print_status "Checking database connection..."
    if ! docker-compose exec -T postgres pg_isready -U "${DATABASE_USER:-postgres}" > /dev/null 2>&1; then
        print_error "Cannot connect to database. Make sure Docker services are running."
        print_status "Starting Docker services..."
        docker-compose up -d
        sleep 5
    fi
    
    # Run tests
    test_migration_commands
    echo
    
    test_migration_workflow
    echo
    
    test_seeding
    echo
    
    test_step_migration
    echo
    
    test_migrate_to
    echo
    
    test_json_output
    echo
    
    if [ "$1" != "--skip-go-tests" ]; then
        test_go_tests
        echo
    fi
    
    echo "=================================================="
    print_success "All migration tests completed successfully!"
    echo "=================================================="
    echo
    
    print_status "Final migration status:"
    go run ./cmd/migrate -action=status
}

# Show usage if help requested
if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    echo "Migration Testing Script"
    echo
    echo "Usage: $0 [--skip-go-tests]"
    echo
    echo "Options:"
    echo "  --skip-go-tests    Skip Go unit/integration tests"
    echo "  --help             Show this help message"
    echo
    echo "This script tests the complete migration and seeding functionality including:"
    echo "  - Migration commands (up, down, status, list, validate)"
    echo "  - Step-by-step migrations"
    echo "  - Migration to specific versions"
    echo "  - Database seeding"
    echo "  - JSON output formats"
    echo "  - Go unit and integration tests"
    exit 0
fi

# Run main function
main "$@"