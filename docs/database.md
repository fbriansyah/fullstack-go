# Database Documentation

This document describes the database infrastructure and usage patterns for the Go Templ Template project.

## Overview

The database layer provides:
- Connection management with connection pooling using SQLx
- Database migrations using golang-migrate
- Health checking and monitoring
- Transaction support
- Connection validation

## Configuration

Database configuration is managed through environment variables:

```bash
# Primary connection string (takes precedence)
DATABASE_URL=postgres://user:password@host:port/dbname?sslmode=disable

# Individual components (used if DATABASE_URL is not set)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=go_templ_template
DB_SSLMODE=disable
```

## Usage

### Basic Usage

```go
package main

import (
    "context"
    "log"
    "time"
    
    "go-templ-template/internal/config"
    "go-templ-template/internal/shared/database"
)

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }

    // Create database manager
    manager, err := database.NewManager(&cfg.Database, "./migrations")
    if err != nil {
        log.Fatal(err)
    }
    defer manager.Close()

    // Initialize with migrations
    ctx := context.Background()
    err = manager.Initialize(ctx, true) // true = run migrations
    if err != nil {
        log.Fatal(err)
    }

    // Use the database
    db := manager.DB
    
    // Example query
    var count int
    err = db.Get(&count, "SELECT COUNT(*) FROM users")
    if err != nil {
        log.Printf("Query failed: %v", err)
    }
}
```

### Transactions

```go
// Begin transaction
tx, err := db.Beginx()
if err != nil {
    return err
}
defer tx.Rollback() // Will be ignored if Commit() is called

// Execute queries in transaction
_, err = tx.Exec("INSERT INTO users (email, password_hash) VALUES ($1, $2)", 
    "user@example.com", "hashed_password")
if err != nil {
    return err
}

// Commit transaction
return tx.Commit()
```

### Health Checking

```go
// Quick health check
healthChecker := database.NewHealthChecker(db)
err := healthChecker.QuickCheck(context.Background())
if err != nil {
    log.Printf("Database unhealthy: %v", err)
}

// Comprehensive health check
status := healthChecker.Check(context.Background())
log.Printf("Status: %s, Latency: %v", status.Status, status.Latency)
```

## Migrations

### Creating Migrations

Create migration files in the `migrations/` directory:

```
migrations/
├── 001_initial_schema.up.sql
├── 001_initial_schema.down.sql
├── 002_add_user_profiles.up.sql
└── 002_add_user_profiles.down.sql
```

Migration files follow the naming convention: `{version}_{description}.{direction}.sql`

### Running Migrations

Using the CLI tool:

```bash
# Run all pending migrations
go run ./cmd/migrate -action=up

# Rollback all migrations
go run ./cmd/migrate -action=down

# Run specific number of steps
go run ./cmd/migrate -action=steps -steps=1   # Up 1 step
go run ./cmd/migrate -action=steps -steps=-1  # Down 1 step

# Check current version
go run ./cmd/migrate -action=version

# Force version (use with caution)
go run ./cmd/migrate -action=force -version=1
```

Using Make commands:

```bash
make migrate-up
make migrate-down
make migrate-version
make migrate-steps STEPS=1
```

### Migration Best Practices

1. **Always create both up and down migrations**
2. **Test migrations on a copy of production data**
3. **Keep migrations small and focused**
4. **Use transactions for data migrations**
5. **Never modify existing migration files**

Example migration structure:

```sql
-- 002_add_user_profiles.up.sql
BEGIN;

CREATE TABLE user_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bio TEXT,
    avatar_url VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_user_profiles_user_id ON user_profiles(user_id);

COMMIT;
```

```sql
-- 002_add_user_profiles.down.sql
BEGIN;

DROP TABLE IF EXISTS user_profiles;

COMMIT;
```

## Health Monitoring

### CLI Health Check

```bash
# Basic health check
go run ./cmd/dbhealth

# JSON output
go run ./cmd/dbhealth -format=json

# Wait for database to become available
go run ./cmd/dbhealth -wait -timeout=30s
```

Using Make commands:

```bash
make db-health
make db-health-json
make db-wait
```

### Programmatic Health Monitoring

```go
// Continuous monitoring
ticker := time.NewTicker(30 * time.Second)
defer ticker.Stop()

for {
    select {
    case <-ticker.C:
        status := manager.GetHealthStatus(context.Background())
        if status.Status != "healthy" {
            log.Printf("Database unhealthy: %s", status.Message)
            // Handle unhealthy database (alerts, etc.)
        }
    }
}
```

## Connection Pool Configuration

Default connection pool settings:

```go
type ConnectionOptions struct {
    MaxOpenConns    int           // 25
    MaxIdleConns    int           // 5
    ConnMaxLifetime time.Duration // 5 minutes
    ConnMaxIdleTime time.Duration // 5 minutes
}
```

Customize for your needs:

```go
opts := database.ConnectionOptions{
    MaxOpenConns:    50,
    MaxIdleConns:    10,
    ConnMaxLifetime: 10 * time.Minute,
    ConnMaxIdleTime: 5 * time.Minute,
}

db, err := database.NewConnection(&cfg.Database, opts)
```

## Testing

### Unit Tests

Run database tests with a test database:

```bash
export TEST_DATABASE_URL="postgres://user:pass@localhost:5432/test_db"
go test ./internal/shared/database/...
```

### Integration Tests

The database package includes integration tests that require a real database connection. Set the `TEST_DATABASE_URL` environment variable to run these tests.

## Troubleshooting

### Common Issues

1. **Connection refused**: Check if PostgreSQL is running and accessible
2. **Authentication failed**: Verify username/password in configuration
3. **Database does not exist**: Create the database or check the database name
4. **Migration errors**: Check migration file syntax and database permissions

### Debug Connection Issues

```bash
# Test basic connectivity
make db-health

# Check detailed connection info
make db-health-json

# Wait for database during startup
make db-wait
```

### Connection Pool Monitoring

Monitor connection pool statistics:

```go
stats := db.Stats()
log.Printf("Open: %d, InUse: %d, Idle: %d", 
    stats.OpenConnections, stats.InUse, stats.Idle)
```

## Performance Considerations

1. **Connection Pooling**: Tune pool size based on your application's concurrency needs
2. **Query Optimization**: Use EXPLAIN ANALYZE to optimize slow queries
3. **Indexes**: Ensure proper indexes for your query patterns
4. **Prepared Statements**: SQLx automatically uses prepared statements for better performance
5. **Transaction Scope**: Keep transactions as short as possible

## Security

1. **Connection Security**: Use SSL/TLS in production (`sslmode=require`)
2. **Credentials**: Store database credentials securely (environment variables, secrets management)
3. **SQL Injection**: SQLx with named parameters prevents SQL injection
4. **Least Privilege**: Use database users with minimal required permissions
5. **Network Security**: Restrict database access to application servers only