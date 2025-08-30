# Development Environment Setup Script for Windows PowerShell
# This script sets up the complete development environment

param(
    [switch]$SkipDependencyCheck,
    [switch]$Force
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

function Test-Command {
    param([string]$Command)
    try {
        Get-Command $Command -ErrorAction Stop | Out-Null
        return $true
    }
    catch {
        return $false
    }
}

function Test-Dependencies {
    Write-Status "Checking dependencies..."
    
    $missing = @()
    
    # Check Go
    if (-not (Test-Command "go")) {
        $missing += "Go 1.21+"
    }
    
    # Check Docker
    if (-not (Test-Command "docker")) {
        $missing += "Docker"
    }
    
    # Check Docker Compose
    if (-not (Test-Command "docker-compose") -and -not (docker compose version 2>$null)) {
        $missing += "Docker Compose"
    }
    
    if ($missing.Count -gt 0) {
        Write-Error "Missing dependencies: $($missing -join ', ')"
        Write-Host "Please install the missing dependencies and run this script again." -ForegroundColor $Colors.Red
        exit 1
    }
    
    Write-Success "All dependencies are available"
}

function Install-GoDependencies {
    Write-Status "Installing Go dependencies..."
    
    # Tidy modules
    go mod tidy
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to tidy Go modules"
        exit 1
    }
    
    # Install Templ CLI
    Write-Status "Installing Templ CLI..."
    go install github.com/a-h/templ/cmd/templ@latest
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to install Templ CLI"
        exit 1
    }
    
    # Install Air for hot reload
    Write-Status "Installing Air for hot reload..."
    go install github.com/cosmtrek/air@latest
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to install Air"
        exit 1
    }
    
    Write-Success "Go dependencies installed"
}

function Setup-Environment {
    Write-Status "Setting up environment configuration..."
    
    if (-not (Test-Path ".env") -or $Force) {
        if (Test-Path ".env.example") {
            Copy-Item ".env.example" ".env"
            Write-Success "Created .env file from .env.example"
        }
        else {
            Write-Warning ".env.example not found, creating basic .env file"
            $envContent = @"
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
"@
            Set-Content -Path ".env" -Value $envContent
            Write-Success "Created basic .env file"
        }
    }
    else {
        Write-Warning ".env file already exists, skipping creation (use -Force to overwrite)"
    }
}

function Start-Services {
    Write-Status "Starting Docker services..."
    
    # Stop any existing services
    docker-compose down 2>$null
    
    # Start services
    docker-compose up -d
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to start Docker services"
        exit 1
    }
    
    Write-Status "Waiting for services to be ready..."
    Start-Sleep -Seconds 5
    
    # Wait for PostgreSQL to be ready
    Write-Status "Waiting for PostgreSQL to be ready..."
    $timeout = 30
    while ($timeout -gt 0) {
        $result = docker-compose exec -T postgres pg_isready -U postgres 2>$null
        if ($LASTEXITCODE -eq 0) {
            break
        }
        $timeout--
        if ($timeout -eq 0) {
            Write-Error "PostgreSQL failed to start within 30 seconds"
            exit 1
        }
        Start-Sleep -Seconds 1
    }
    
    Write-Success "Docker services are running"
}

function Invoke-Migrations {
    Write-Status "Running database migrations..."
    
    # Wait a bit more for database to be fully ready
    Start-Sleep -Seconds 2
    
    # Run migrations using make
    make migrate-up
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Database migrations failed"
        exit 1
    }
    
    Write-Success "Database migrations completed"
}

function Invoke-TemplGeneration {
    Write-Status "Generating Templ files..."
    
    make templ-generate
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Templ generation failed"
        exit 1
    }
    
    Write-Success "Templ files generated"
}

function Build-Application {
    Write-Status "Building the application..."
    
    make build
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Application build failed"
        exit 1
    }
    
    Write-Success "Application built successfully"
}

function Main {
    Write-Host "==================================================" -ForegroundColor $Colors.White
    Write-Host "  Go Templ Template Development Setup" -ForegroundColor $Colors.White
    Write-Host "==================================================" -ForegroundColor $Colors.White
    Write-Host ""
    
    if (-not $SkipDependencyCheck) {
        Test-Dependencies
    }
    
    Install-GoDependencies
    Setup-Environment
    Start-Services
    Invoke-Migrations
    Invoke-TemplGeneration
    Build-Application
    
    Write-Host ""
    Write-Host "==================================================" -ForegroundColor $Colors.White
    Write-Success "Development environment setup complete!"
    Write-Host "==================================================" -ForegroundColor $Colors.White
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor $Colors.White
    Write-Host "  1. Start development server: make dev" -ForegroundColor $Colors.White
    Write-Host "  2. Open your browser: http://localhost:8080" -ForegroundColor $Colors.White
    Write-Host "  3. Start coding!" -ForegroundColor $Colors.White
    Write-Host ""
    Write-Host "Useful commands:" -ForegroundColor $Colors.White
    Write-Host "  make help          - Show all available commands" -ForegroundColor $Colors.White
    Write-Host "  make dev           - Start development with hot reload" -ForegroundColor $Colors.White
    Write-Host "  make test          - Run tests" -ForegroundColor $Colors.White
    Write-Host "  make docker-down   - Stop Docker services" -ForegroundColor $Colors.White
    Write-Host ""
}

# Run main function
Main