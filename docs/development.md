# Development Environment Setup

This guide will help you set up a complete development environment for the Go Templ Template project.

## Prerequisites

Before you begin, ensure you have the following installed on your system:

### Required Tools

- **Go 1.21+**: [Download Go](https://golang.org/dl/)
- **Docker**: [Install Docker](https://docs.docker.com/get-docker/)
- **Docker Compose**: Usually included with Docker Desktop
- **Git**: [Install Git](https://git-scm.com/downloads)

### Optional Tools (Recommended)

- **Make**: For running Makefile commands
  - **Windows**: Install via [Chocolatey](https://chocolatey.org/) (`choco install make`) or use PowerShell scripts
  - **macOS**: Install via Homebrew (`brew install make`)
  - **Linux**: Usually pre-installed or available via package manager
- **golangci-lint**: For code linting (`go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`)
- **Node.js & npm**: For Tailwind CSS development (optional)

## Quick Setup

### Automated Setup (Recommended)

The easiest way to get started is using our automated setup scripts:

#### Linux/macOS
```bash
# Make the script executable and run it
chmod +x scripts/dev-setup.sh
./scripts/dev-setup.sh

# Or use Make
make setup
```

#### Windows (PowerShell)
```powershell
# Run the PowerShell setup script
.\scripts\dev-setup.ps1

# Or use Make (if available)
make setup
```

The automated setup will:
1. Check all dependencies
2. Install Go tools (Templ, Air)
3. Create `.env` file from template
4. Start Docker services
5. Run database migrations
6. Generate Templ files
7. Build the application

### Manual Setup

If you prefer to set up manually or the automated script fails:

#### 1. Clone and Navigate
```bash
git clone <repository-url>
cd go-templ-template
```

#### 2. Install Dependencies
```bash
# Install Go dependencies
make install-deps

# Or manually:
go mod tidy
go install github.com/a-h/templ/cmd/templ@latest
go install github.com/cosmtrek/air@latest
```

#### 3. Environment Configuration
```bash
# Copy environment template
cp .env.example .env

# Edit .env file with your preferred settings
```

#### 4. Start Services
```bash
# Start PostgreSQL and RabbitMQ
make docker-up
```

#### 5. Database Setup
```bash
# Run migrations
make migrate-up

# Seed with development data (optional)
make db-seed
```

#### 6. Generate Templates
```bash
# Generate Go code from Templ files
make templ-generate
```

#### 7. Build Application
```bash
# Build the application
make build
```

## Development Workflow

### Starting Development

Once setup is complete, start the development server:

```bash
# Start development server with hot reload
make dev
```

This will:
- Start Docker services (PostgreSQL, RabbitMQ)
- Start the Go application with Air hot reload
- Watch for file changes and automatically restart

The application will be available at: `http://localhost:8080`

### Development Commands

| Command | Description |
|---------|-------------|
| `make dev` | Start development server with hot reload |
| `make build` | Build the application |
| `make test` | Run all tests |
| `make test-coverage` | Run tests with coverage report |
| `make lint` | Run code linter |
| `make fmt` | Format code |
| `make clean` | Clean build artifacts |

### Database Commands

| Command | Description |
|---------|-------------|
| `make migrate-up` | Run database migrations |
| `make migrate-down` | Rollback database migrations |
| `make db-health` | Check database connectivity |
| `make db-seed` | Seed database with development data |

### Docker Commands

| Command | Description |
|---------|-------------|
| `make docker-up` | Start Docker services |
| `make docker-down` | Stop Docker services |
| `make docker-build` | Build Docker image |

## Hot Reload Configuration

The project uses [Air](https://github.com/cosmtrek/air) for hot reload during development. The configuration is in `.air.toml`:

### What Gets Watched
- Go files in `cmd/`, `internal/`
- Templ files in `web/templates/`
- CSS and JS files in `web/static/`

### What Gets Excluded
- Test files (`*_test.go`)
- Generated files (`*_templ.go`)
- Build artifacts (`tmp/`, `bin/`)
- Dependencies (`vendor/`, `node_modules/`)

### Customizing Hot Reload

Edit `.air.toml` to customize the hot reload behavior:

```toml
[build]
  # Command to build the application
  cmd = "templ generate && go build -o ./tmp/main ./cmd/server"
  
  # Files to watch
  include_ext = ["go", "templ", "html", "css", "js"]
  
  # Directories to exclude
  exclude_dir = ["tmp", "vendor", "node_modules", ".git"]
```

## Database Development

### Connection Details

Default development database settings:

- **Host**: localhost
- **Port**: 5432
- **Database**: go_templ_template
- **Username**: postgres
- **Password**: postgres

### Migrations

Database migrations are located in the `migrations/` directory:

```
migrations/
├── 001_initial_schema.up.sql
├── 001_initial_schema.down.sql
├── 002_add_session_is_active.up.sql
└── 002_add_session_is_active.down.sql
```

#### Creating New Migrations

1. Create new migration files with sequential numbering:
   ```
   003_your_migration_name.up.sql
   003_your_migration_name.down.sql
   ```

2. Write the migration SQL in the `.up.sql` file
3. Write the rollback SQL in the `.down.sql` file
4. Run the migration: `make migrate-up`

### Development Data

Use the seeding script to populate your database with test data:

```bash
# Seed database with sample data
make db-seed

# Clear existing data and reseed
./scripts/db-seed.sh --clear
```

The seeding script creates:
- Sample users with different statuses
- Sample sessions
- Sample audit events

Default login credentials:
- **Email**: admin@example.com
- **Password**: password123

## Frontend Development

### Templ Templates

Templ files are located in `web/templates/`:

```
web/templates/
├── layouts/
│   ├── base.templ
│   └── auth.templ
├── components/
│   ├── header.templ
│   └── footer.templ
└── pages/
    ├── home.templ
    └── login.templ
```

#### Working with Templ

1. Edit `.templ` files
2. Air will automatically run `templ generate`
3. Generated Go files (`*_templ.go`) are created
4. Application rebuilds and restarts

### CSS Development

The project uses Tailwind CSS for styling:

#### Watching CSS Changes

If you have Node.js installed:

```bash
# Watch and build CSS files
make watch-css

# Or manually
npx tailwindcss -i ./web/static/css/input.css -o ./web/static/css/output.css --watch
```

#### CSS File Structure

```
web/static/
├── css/
│   ├── input.css      # Source CSS with Tailwind directives
│   └── output.css     # Generated CSS (don't edit)
├── js/
└── images/
```

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with verbose output
make test-verbose

# Run tests with coverage
make test-coverage
```

### Test Structure

```
tests/
├── unit/           # Unit tests
├── integration/    # Integration tests
└── fixtures/       # Test data and fixtures
```

### Writing Tests

- Place test files alongside source files with `_test.go` suffix
- Use Go's built-in testing package
- Consider using testify for assertions

## Debugging

### Application Logs

Development logs are written to stdout with debug level enabled.

### Database Debugging

Connect to the development database:

```bash
# Using Docker
docker-compose exec postgres psql -U postgres -d go_templ_template

# Or using psql directly
psql -h localhost -p 5432 -U postgres -d go_templ_template
```

### RabbitMQ Management

Access RabbitMQ management interface:
- **URL**: http://localhost:15672
- **Username**: guest
- **Password**: guest

## Troubleshooting

### Common Issues

#### Port Already in Use
```bash
# Check what's using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>
```

#### Docker Services Won't Start
```bash
# Check Docker status
docker-compose ps

# View logs
docker-compose logs

# Restart services
make docker-down
make docker-up
```

#### Database Connection Issues
```bash
# Check database health
make db-health

# Check if PostgreSQL is running
docker-compose exec postgres pg_isready -U postgres
```

#### Templ Generation Fails
```bash
# Manually generate Templ files
templ generate

# Check for syntax errors in .templ files
templ fmt .
```

### Getting Help

1. Check the logs for error messages
2. Ensure all dependencies are installed: `make check-deps`
3. Try restarting services: `make docker-down && make docker-up`
4. Clean and rebuild: `make clean && make build`

## IDE Configuration

### VS Code

Recommended extensions:
- Go extension
- Templ extension
- Tailwind CSS IntelliSense

### GoLand/IntelliJ

- Enable Go modules support
- Configure file watchers for Templ files

## Production Considerations

While this guide focuses on development, keep in mind:

- Use environment variables for configuration
- Never commit secrets to version control
- Use proper logging levels in production
- Configure proper database connection pooling
- Set up proper monitoring and health checks

## Next Steps

Once your development environment is set up:

1. Explore the codebase structure
2. Run the existing tests to understand the functionality
3. Try making small changes and see hot reload in action
4. Read the API documentation
5. Start implementing your features!

For more information, see:
- [Project Structure](../README.md#project-structure)
- [API Documentation](./api.md)
- [Database Schema](./database.md)