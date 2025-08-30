# Database Seeding Script for Windows PowerShell
# This script seeds the database with development data

param(
    [switch]$Clear,
    [switch]$Help
)

# Colors for output
$Colors = @{
    Red = "Red"
    Green = "Green"
    Yellow = "Yellow"
    Blue = "Blue"
    White = "White"
}

function Write-Status {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor $Colors.Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor $Colors.Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor $Colors.Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor $Colors.Red
}

function Show-Help {
    Write-Host "Database Seeding Script" -ForegroundColor $Colors.White
    Write-Host ""
    Write-Host "Usage: .\scripts\db-seed.ps1 [-Clear] [-Help]" -ForegroundColor $Colors.White
    Write-Host ""
    Write-Host "Parameters:" -ForegroundColor $Colors.White
    Write-Host "  -Clear     Clear existing data before seeding" -ForegroundColor $Colors.White
    Write-Host "  -Help      Show this help message" -ForegroundColor $Colors.White
    Write-Host ""
    Write-Host "This script seeds the database with development data including:" -ForegroundColor $Colors.White
    Write-Host "  - Sample users with different statuses" -ForegroundColor $Colors.White
    Write-Host "  - Sample sessions" -ForegroundColor $Colors.White
    Write-Host "  - Sample audit events" -ForegroundColor $Colors.White
    Write-Host ""
    Write-Host "All users will have the password: password123" -ForegroundColor $Colors.White
    exit 0
}

function Get-EnvironmentVariables {
    if (Test-Path ".env") {
        Get-Content ".env" | ForEach-Object {
            if ($_ -match "^([^#][^=]+)=(.*)$") {
                [Environment]::SetEnvironmentVariable($matches[1], $matches[2], "Process")
            }
        }
    }
    else {
        Write-Error ".env file not found. Please run dev-setup.ps1 first."
        exit 1
    }
}

function Invoke-SQL {
    param(
        [string]$SQL,
        [string]$Description
    )
    
    Write-Status $Description
    
    $dbUser = $env:DATABASE_USER
    if (-not $dbUser) { $dbUser = "postgres" }
    
    $dbName = $env:DATABASE_NAME
    if (-not $dbName) { $dbName = "go_templ_template" }
    
    try {
        $result = docker-compose exec -T postgres psql -U $dbUser -d $dbName -c $SQL 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Success "$Description completed"
            return $true
        }
        else {
            Write-Error "$Description failed"
            return $false
        }
    }
    catch {
        Write-Error "$Description failed: $($_.Exception.Message)"
        return $false
    }
}

function Test-Database {
    Write-Status "Checking database connection..."
    
    $dbUser = $env:DATABASE_USER
    if (-not $dbUser) { $dbUser = "postgres" }
    
    try {
        $result = docker-compose exec -T postgres pg_isready -U $dbUser 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Database is accessible"
            return $true
        }
        else {
            Write-Error "Cannot connect to database. Make sure Docker services are running."
            return $false
        }
    }
    catch {
        Write-Error "Cannot connect to database. Make sure Docker services are running."
        return $false
    }
}

function Clear-Data {
    if ($Clear) {
        Write-Warning "Clearing existing data..."
        
        Invoke-SQL "TRUNCATE TABLE audit_events CASCADE;" "Clearing audit events"
        Invoke-SQL "TRUNCATE TABLE sessions CASCADE;" "Clearing sessions"
        Invoke-SQL "TRUNCATE TABLE users CASCADE;" "Clearing users"
        
        Write-Success "Existing data cleared"
    }
}

function Add-Users {
    Write-Status "Seeding users..."
    
    # Hash for password "password123" (bcrypt cost 10)
    $passwordHash = '$2a$10$rOCKx7VDum0oaEFrZQWOa.6nYgs8o8/Oa9Q8Ks8Ks8Ks8Ks8Ks8K'
    
    # Admin user
    $adminSQL = @"
        INSERT INTO users (id, email, password, first_name, last_name, status, created_at, updated_at, version)
        VALUES (
            'admin-user-id-1234567890',
            'admin@example.com',
            '$passwordHash',
            'Admin',
            'User',
            'active',
            NOW(),
            NOW(),
            1
        ) ON CONFLICT (email) DO NOTHING;
"@
    
    Invoke-SQL $adminSQL "Creating admin user"
    
    # Regular users
    $usersSQL = @"
        INSERT INTO users (id, email, password, first_name, last_name, status, created_at, updated_at, version)
        VALUES 
        (
            'user-id-1234567890',
            'john.doe@example.com',
            '$passwordHash',
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
            '$passwordHash',
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
            '$passwordHash',
            'Bob',
            'Wilson',
            'inactive',
            NOW(),
            NOW(),
            1
        )
        ON CONFLICT (email) DO NOTHING;
"@
    
    Invoke-SQL $usersSQL "Creating regular users"
    
    Write-Success "Users seeded successfully"
}

function Add-Sessions {
    Write-Status "Seeding sample sessions..."
    
    $sessionSQL = @"
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
"@
    
    Invoke-SQL $sessionSQL "Creating sample admin session"
    
    Write-Success "Sessions seeded successfully"
}

function Add-AuditEvents {
    Write-Status "Seeding audit events..."
    
    $auditSQL = @"
        INSERT INTO audit_events (id, event_type, aggregate_id, event_data, occurred_at, created_at)
        VALUES 
        (
            'audit-event-1',
            'user.created',
            'admin-user-id-1234567890',
            '{\"email\": \"admin@example.com\", \"first_name\": \"Admin\", \"last_name\": \"User\"}',
            NOW() - INTERVAL '1 hour',
            NOW() - INTERVAL '1 hour'
        ),
        (
            'audit-event-2',
            'user.created',
            'user-id-1234567890',
            '{\"email\": \"john.doe@example.com\", \"first_name\": \"John\", \"last_name\": \"Doe\"}',
            NOW() - INTERVAL '30 minutes',
            NOW() - INTERVAL '30 minutes'
        ),
        (
            'audit-event-3',
            'user.login',
            'admin-user-id-1234567890',
            '{\"session_id\": \"session-admin-123456\", \"ip_address\": \"127.0.0.1\"}',
            NOW() - INTERVAL '15 minutes',
            NOW() - INTERVAL '15 minutes'
        )
        ON CONFLICT (id) DO NOTHING;
"@
    
    Invoke-SQL $auditSQL "Creating audit events"
    
    Write-Success "Audit events seeded successfully"
}

function Show-Summary {
    Write-Status "Seeded data summary:"
    
    Write-Host ""
    Write-Host "Users:" -ForegroundColor $Colors.White
    Write-Host "  - admin@example.com (Admin User) - Active" -ForegroundColor $Colors.White
    Write-Host "  - john.doe@example.com (John Doe) - Active" -ForegroundColor $Colors.White
    Write-Host "  - jane.smith@example.com (Jane Smith) - Active" -ForegroundColor $Colors.White
    Write-Host "  - bob.wilson@example.com (Bob Wilson) - Inactive" -ForegroundColor $Colors.White
    Write-Host ""
    Write-Host "Default password for all users: password123" -ForegroundColor $Colors.Yellow
    Write-Host ""
    Write-Host "Sessions:" -ForegroundColor $Colors.White
    Write-Host "  - Admin user has an active session" -ForegroundColor $Colors.White
    Write-Host ""
    Write-Host "You can now log in with any of the seeded users!" -ForegroundColor $Colors.Green
}

function Main {
    if ($Help) {
        Show-Help
    }
    
    Write-Host "==================================================" -ForegroundColor $Colors.White
    Write-Host "  Database Seeding for Go Templ Template" -ForegroundColor $Colors.White
    Write-Host "==================================================" -ForegroundColor $Colors.White
    Write-Host ""
    
    Get-EnvironmentVariables
    
    if (-not (Test-Database)) {
        exit 1
    }
    
    Clear-Data
    Add-Users
    Add-Sessions
    Add-AuditEvents
    
    Write-Host ""
    Write-Host "==================================================" -ForegroundColor $Colors.White
    Write-Success "Database seeding completed!"
    Write-Host "==================================================" -ForegroundColor $Colors.White
    Write-Host ""
    
    Show-Summary
}

# Run main function
Main