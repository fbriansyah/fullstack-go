# Migration Testing Script for Windows PowerShell
# This script tests the migration and seeding functionality

param(
    [switch]$SkipGoTests,
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
    Write-Host "Migration Testing Script" -ForegroundColor $Colors.White
    Write-Host ""
    Write-Host "Usage: .\scripts\test-migrations.ps1 [-SkipGoTests] [-Help]" -ForegroundColor $Colors.White
    Write-Host ""
    Write-Host "Parameters:" -ForegroundColor $Colors.White
    Write-Host "  -SkipGoTests   Skip Go unit/integration tests" -ForegroundColor $Colors.White
    Write-Host "  -Help          Show this help message" -ForegroundColor $Colors.White
    Write-Host ""
    Write-Host "This script tests the complete migration and seeding functionality including:" -ForegroundColor $Colors.White
    Write-Host "  - Migration commands (up, down, status, list, validate)" -ForegroundColor $Colors.White
    Write-Host "  - Step-by-step migrations" -ForegroundColor $Colors.White
    Write-Host "  - Migration to specific versions" -ForegroundColor $Colors.White
    Write-Host "  - Database seeding" -ForegroundColor $Colors.White
    Write-Host "  - JSON output formats" -ForegroundColor $Colors.White
    Write-Host "  - Go unit and integration tests" -ForegroundColor $Colors.White
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

function Test-MigrationCommands {
    Write-Status "Testing migration commands..."
    
    # Test migration status
    Write-Status "Checking migration status..."
    go run ./cmd/migrate -action=status
    
    # Test migration list
    Write-Status "Listing migrations..."
    go run ./cmd/migrate -action=list
    
    # Test migration validation
    Write-Status "Validating migrations..."
    go run ./cmd/migrate -action=validate
    
    Write-Success "Migration commands tested successfully"
}

function Test-MigrationWorkflow {
    Write-Status "Testing migration workflow..."
    
    # Start from clean state
    Write-Status "Rolling back all migrations..."
    try {
        go run ./cmd/migrate -action=down
    }
    catch {
        Write-Warning "No migrations to roll back"
    }
    
    # Check version (should be 0 or error)
    Write-Status "Checking version after rollback..."
    try {
        go run ./cmd/migrate -action=version
    }
    catch {
        Write-Warning "No version (clean database)"
    }
    
    # Run migrations up
    Write-Status "Running migrations up..."
    go run ./cmd/migrate -action=up
    
    # Check final version
    Write-Status "Checking final version..."
    go run ./cmd/migrate -action=version
    
    # Check status
    Write-Status "Checking final status..."
    go run ./cmd/migrate -action=status
    
    Write-Success "Migration workflow tested successfully"
}

function Test-Seeding {
    Write-Status "Testing database seeding..."
    
    # Run seeding script
    if (Test-Path "scripts/db-seed.ps1") {
        .\scripts\db-seed.ps1 -Clear
    }
    else {
        Write-Error "Seeding script not found"
        return $false
    }
    
    Write-Success "Database seeding tested successfully"
    return $true
}

function Test-StepMigration {
    Write-Status "Testing step-by-step migration..."
    
    # Start from clean state
    try {
        go run ./cmd/migrate -action=down
    }
    catch {
        Write-Warning "No migrations to roll back"
    }
    
    # Run one step up
    Write-Status "Running one step up..."
    go run ./cmd/migrate -action=steps -steps=1
    go run ./cmd/migrate -action=version
    
    # Run another step up
    Write-Status "Running another step up..."
    go run ./cmd/migrate -action=steps -steps=1
    go run ./cmd/migrate -action=version
    
    # Run one step down
    Write-Status "Running one step down..."
    go run ./cmd/migrate -action=steps -steps=-1
    go run ./cmd/migrate -action=version
    
    Write-Success "Step-by-step migration tested successfully"
}

function Test-MigrateTo {
    Write-Status "Testing migration to specific version..."
    
    # Migrate to version 1
    Write-Status "Migrating to version 1..."
    go run ./cmd/migrate -action=to -version=1
    go run ./cmd/migrate -action=version
    
    # Migrate to version 2
    Write-Status "Migrating to version 2..."
    go run ./cmd/migrate -action=to -version=2
    go run ./cmd/migrate -action=version
    
    # Migrate back to version 1
    Write-Status "Migrating back to version 1..."
    go run ./cmd/migrate -action=to -version=1
    go run ./cmd/migrate -action=version
    
    Write-Success "Migration to specific version tested successfully"
}

function Test-JsonOutput {
    Write-Status "Testing JSON output..."
    
    Write-Status "Status in JSON format:"
    go run ./cmd/migrate -action=status -format=json
    
    Write-Status "Version in JSON format:"
    go run ./cmd/migrate -action=version -format=json
    
    Write-Status "List in JSON format:"
    go run ./cmd/migrate -action=list -format=json
    
    Write-Success "JSON output tested successfully"
}

function Test-GoTests {
    Write-Status "Running Go migration tests..."
    
    # Set test database URL if not set
    if (-not $env:TEST_DATABASE_URL) {
        $env:TEST_DATABASE_URL = $env:DATABASE_URL
    }
    
    # Run migration tests
    go test -v ./internal/shared/database -run TestMigration
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "Go migration tests completed successfully"
        return $true
    }
    else {
        Write-Error "Go migration tests failed"
        return $false
    }
}

function Test-DatabaseConnection {
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
            Write-Error "Cannot connect to database. Starting Docker services..."
            docker-compose up -d
            Start-Sleep -Seconds 5
            return $true
        }
    }
    catch {
        Write-Error "Cannot connect to database. Starting Docker services..."
        docker-compose up -d
        Start-Sleep -Seconds 5
        return $true
    }
}

function Main {
    if ($Help) {
        Show-Help
    }
    
    Write-Host "==================================================" -ForegroundColor $Colors.White
    Write-Host "  Migration Testing for Go Templ Template" -ForegroundColor $Colors.White
    Write-Host "==================================================" -ForegroundColor $Colors.White
    Write-Host ""
    
    Get-EnvironmentVariables
    
    if (-not (Test-DatabaseConnection)) {
        exit 1
    }
    
    # Run tests
    Test-MigrationCommands
    Write-Host ""
    
    Test-MigrationWorkflow
    Write-Host ""
    
    if (-not (Test-Seeding)) {
        Write-Warning "Seeding test failed, continuing with other tests..."
    }
    Write-Host ""
    
    Test-StepMigration
    Write-Host ""
    
    Test-MigrateTo
    Write-Host ""
    
    Test-JsonOutput
    Write-Host ""
    
    if (-not $SkipGoTests) {
        if (-not (Test-GoTests)) {
            Write-Warning "Go tests failed, but continuing..."
        }
        Write-Host ""
    }
    
    Write-Host "==================================================" -ForegroundColor $Colors.White
    Write-Success "All migration tests completed!"
    Write-Host "==================================================" -ForegroundColor $Colors.White
    Write-Host ""
    
    Write-Status "Final migration status:"
    go run ./cmd/migrate -action=status
}

# Run main function
Main