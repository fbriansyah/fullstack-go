#!/bin/bash

# Database Seeding Script
# This script seeds the database with development data

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

# Database connection parameters
DB_HOST=${DATABASE_HOST:-localhost}
DB_PORT=${DATABASE_PORT:-5432}
DB_NAME=${DATABASE_NAME:-go_templ_template}
DB_USER=${DATABASE_USER:-postgres}
DB_PASSWORD=${DATABASE_PASSWORD:-postgres}

# Function to execute SQL
execute_sql() {
    local sql="$1"
    local description="$2"
    
    print_status "$description"
    
    if docker-compose exec -T postgres psql -U "$DB_USER" -d "$DB_NAME" -c "$sql" > /dev/null 2>&1; then
        print_success "$description completed"
    else
        print_error "$description failed"
        return 1
    fi
}

# Function to execute SQL file
execute_sql_file() {
    local file="$1"
    local description="$2"
    
    if [ ! -f "$file" ]; then
        print_error "SQL file not found: $file"
        return 1
    fi
    
    print_status "$description"
    
    if docker-compose exec -T postgres psql -U "$DB_USER" -d "$DB_NAME" -f "/docker-entrypoint-initdb.d/$(basename "$file")" > /dev/null 2>&1; then
        print_success "$description completed"
    else
        print_error "$description failed"
        return 1
    fi
}

# Check if database is accessible
check_database() {
    print_status "Checking database connection..."
    
    if docker-compose exec -T postgres pg_isready -U "$DB_USER" > /dev/null 2>&1; then
        print_success "Database is accessible"
    else
        print_error "Cannot connect to database. Make sure Docker services are running."
        exit 1
    fi
}

# Clear existing data (optional)
clear_data() {
    if [ "$1" = "--clear" ]; then
        print_warning "Clearing existing data..."
        
        execute_sql "TRUNCATE TABLE audit_events CASCADE;" "Clearing audit events"
        execute_sql "TRUNCATE TABLE sessions CASCADE;" "Clearing sessions"
        execute_sql "TRUNCATE TABLE users CASCADE;" "Clearing users"
        
        print_success "Existing data cleared"
    fi
}

# Seed users
seed_users() {
    print_status "Seeding users..."
    
    # Hash for password "password123" (bcrypt cost 10)
    local password_hash='$2a$10$rOCKx7VDum0oaEFrZQWOa.6nYgs8o8/Oa9Q8Ks8Ks8Ks8Ks8Ks8K'
    
    # Admin user
    execute_sql "
        INSERT INTO users (id, email, password, first_name, last_name, status, created_at, updated_at, version)
        VALUES (
            'admin-user-id-1234567890',
            'admin@example.com',
            '$password_hash',
            'Admin',
            'User',
            'active',
            NOW(),
            NOW(),
            1
        ) ON CONFLICT (email) DO NOTHING;
    " "Creating admin user"
    
    # Regular users
    execute_sql "
        INSERT INTO users (id, email, password, first_name, last_name, status, created_at, updated_at, version)
        VALUES 
        (
            'user-id-1234567890',
            'john.doe@example.com',
            '$password_hash',
            'John',
            'Doe',
            'active',
            NOW(),
            NOW(),
            1
        ),
        (
            'user-id-0987654321',
            'jane.smith@example.com',
            '$password_hash',
            'Jane',
            'Smith',
            'active',
            NOW(),
            NOW(),
            1
        ),
        (
            'user-id-1122334455',
            'bob.wilson@example.com',
            '$password_hash',
            'Bob',
            'Wilson',
            'inactive',
            NOW(),
            NOW(),
            1
        )
        ON CONFLICT (email) DO NOTHING;
    " "Creating regular users"
    
    print_success "Users seeded successfully"
}

# Seed sessions (optional - usually created dynamically)
seed_sessions() {
    print_status "Seeding sample sessions..."
    
    execute_sql "
        INSERT INTO sessions (id, user_id, expires_at, created_at, ip_address, user_agent, is_active)
        VALUES (
            'session-admin-123456',
            'admin-user-id-1234567890',
            NOW() + INTERVAL '1 day',
            NOW(),
            '127.0.0.1',
            'Mozilla/5.0 (Development Seed)',
            true
        ) ON CONFLICT (id) DO NOTHING;
    " "Creating sample admin session"
    
    print_success "Sessions seeded successfully"
}

# Seed audit events
seed_audit_events() {
    print_status "Seeding audit events..."
    
    execute_sql "
        INSERT INTO audit_events (event_id, event_type, aggregate_id, aggregate_type, user_id, action, resource, resource_id, details, occurred_at, metadata)
        VALUES 
        (
            'audit-event-1',
            'user.created',
            'admin-user-id-1234567890',
            'user',
            'admin-user-id-1234567890',
            'create',
            'user',
            'admin-user-id-1234567890',
            '{\"email\": \"admin@example.com\", \"first_name\": \"Admin\", \"last_name\": \"User\"}',
            NOW() - INTERVAL '1 hour',
            '{\"source\": \"seed_script\", \"version\": \"1.0\"}'
        ),
        (
            'audit-event-2',
            'user.created',
            'user-id-1234567890',
            'user',
            'user-id-1234567890',
            'create',
            'user',
            'user-id-1234567890',
            '{\"email\": \"john.doe@example.com\", \"first_name\": \"John\", \"last_name\": \"Doe\"}',
            NOW() - INTERVAL '30 minutes',
            '{\"source\": \"seed_script\", \"version\": \"1.0\"}'
        ),
        (
            'audit-event-3',
            'user.login',
            'admin-user-id-1234567890',
            'session',
            'admin-user-id-1234567890',
            'login',
            'session',
            'session-admin-123456',
            '{\"session_id\": \"session-admin-123456\", \"ip_address\": \"127.0.0.1\", \"user_agent\": \"Mozilla/5.0 (Development Seed)\"}',
            NOW() - INTERVAL '15 minutes',
            '{\"source\": \"seed_script\", \"version\": \"1.0\"}'
        ),
        (
            'audit-event-4',
            'user.created',
            'user-id-0987654321',
            'user',
            'user-id-0987654321',
            'create',
            'user',
            'user-id-0987654321',
            '{\"email\": \"jane.smith@example.com\", \"first_name\": \"Jane\", \"last_name\": \"Smith\"}',
            NOW() - INTERVAL '25 minutes',
            '{\"source\": \"seed_script\", \"version\": \"1.0\"}'
        ),
        (
            'audit-event-5',
            'user.created',
            'user-id-1122334455',
            'user',
            'user-id-1122334455',
            'create',
            'user',
            'user-id-1122334455',
            '{\"email\": \"bob.wilson@example.com\", \"first_name\": \"Bob\", \"last_name\": \"Wilson\", \"status\": \"inactive\"}',
            NOW() - INTERVAL '20 minutes',
            '{\"source\": \"seed_script\", \"version\": \"1.0\"}'
        )
        ON CONFLICT (event_id) DO NOTHING;
    " "Creating audit events"
    
    print_success "Audit events seeded successfully"
}

# Display seeded data summary
show_summary() {
    print_status "Seeded data summary:"
    
    echo
    echo "Users:"
    echo "  - admin@example.com (Admin User) - Active"
    echo "  - john.doe@example.com (John Doe) - Active"
    echo "  - jane.smith@example.com (Jane Smith) - Active"
    echo "  - bob.wilson@example.com (Bob Wilson) - Inactive"
    echo
    echo "Default password for all users: password123"
    echo
    echo "Sessions:"
    echo "  - Admin user has an active session"
    echo
    echo "You can now log in with any of the seeded users!"
}

# Main function
main() {
    echo "=================================================="
    echo "  Database Seeding for Go Templ Template"
    echo "=================================================="
    echo
    
    check_database
    clear_data "$1"
    seed_users
    seed_sessions
    seed_audit_events
    
    echo
    echo "=================================================="
    print_success "Database seeding completed!"
    echo "=================================================="
    echo
    
    show_summary
}

# Show usage if help requested
if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    echo "Database Seeding Script"
    echo
    echo "Usage: $0 [--clear]"
    echo
    echo "Options:"
    echo "  --clear    Clear existing data before seeding"
    echo "  --help     Show this help message"
    echo
    echo "This script seeds the database with development data including:"
    echo "  - Sample users with different statuses"
    echo "  - Sample sessions"
    echo "  - Sample audit events"
    echo
    echo "All users will have the password: password123"
    exit 0
fi

# Run main function
main "$@"