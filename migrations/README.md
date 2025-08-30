# Database Migrations

This directory contains database migration files for the Go Templ Template project. The migration system uses [golang-migrate](https://github.com/golang-migrate/migrate) for managing database schema changes.

## Migration Files

Migration files are organized in pairs:
- `{version}_{name}.up.sql` - Contains the migration SQL
- `{version}_{name}.down.sql` - Contains the rollback SQL

### Current Migrations

1. **001_initial_schema** - Creates the basic database structure
   - Users table with authentication fields
   - Sessions table for user sessions
   - Indexes and triggers for performance and data integrity

2. **002_add_session_is_active** - Adds session activity tracking
   - Adds `is_active` column to sessions table
   - Creates indexes for active session queries

3. **003_create_audit_events** - Adds audit logging capability
   - Creates audit_events table for tracking user actions
   - Indexes for efficient audit log queries

## Migration Commands

### Basic Commands

```bash
# Run all pending migrations
make migrate-up

# Rollback all migrations
make migrate-down

# Show current migration version
make migrate-version

# Show detailed migration status
make migrate-status

# List all available migrations
make migrate-list

# Validate all migration files
make migrate-validate
```

### Advanced Commands

```bash
# Run specific number of migration steps
make migrate-steps STEPS=2    # Run 2 steps up
make migrate-steps STEPS=-1   # Run 1 step down

# Migrate to specific version
make migrate-to VERSION=2     # Migrate to version 2

# Force migration version (use with caution)
make migrate-force VERSION=1

# Create new migration
make migrate-create NAME="add user roles"
```

### Direct CLI Usage

You can also use the migration CLI directly:

```bash
# Basic usage
go run ./cmd/migrate -action=up
go run ./cmd/migrate -action=status -format=json

# Create new migration
go run ./cmd/migrate -action=create -name="add user preferences"

# Step-by-step migration
go run ./cmd/migrate -action=steps -steps=1

# Migrate to specific version
go run ./cmd/migrate -action=to -version=3
```

## Migration Best Practices

### 1. Always Create Both Up and Down Migrations

Every migration should have both an up and down file. The down migration should exactly reverse the changes made by the up migration.

```sql
-- 004_add_user_roles.up.sql
CREATE TABLE user_roles (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    role VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 004_add_user_roles.down.sql
DROP TABLE IF EXISTS user_roles;
```

### 2. Use Transactions for Complex Migrations

Wrap complex migrations in transactions to ensure atomicity:

```sql
BEGIN;

-- Multiple related changes
ALTER TABLE users ADD COLUMN department_id INTEGER;
CREATE TABLE departments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);
ALTER TABLE users ADD CONSTRAINT fk_user_department 
    FOREIGN KEY (department_id) REFERENCES departments(id);

COMMIT;
```

### 3. Handle Data Migrations Carefully

When migrating data, consider:
- Large datasets (use batching)
- Data validation
- Rollback strategy

```sql
-- Good: Batch processing for large tables
UPDATE users SET status = 'active' 
WHERE status IS NULL AND id IN (
    SELECT id FROM users WHERE status IS NULL LIMIT 1000
);
```

### 4. Test Migrations Thoroughly

Always test migrations on a copy of production data:

```bash
# Test the complete migration workflow
make test-migrations

# Test specific scenarios
go test -v ./internal/shared/database -run TestMigration
```

### 5. Use Descriptive Names

Migration names should clearly describe what they do:

```bash
# Good
make migrate-create NAME="add user email verification"
make migrate-create NAME="remove deprecated user fields"

# Bad
make migrate-create NAME="user changes"
make migrate-create NAME="fix stuff"
```

## Database Seeding

After running migrations, you can seed the database with development data:

```bash
# Seed with sample data
make db-seed

# Clear existing data and re-seed
./scripts/db-seed.sh --clear
```

The seeding scripts create:
- Sample users with different statuses
- Active sessions for testing
- Audit events for demonstration

## Troubleshooting

### Migration is Stuck (Dirty State)

If a migration fails and leaves the database in a "dirty" state:

```bash
# Check current status
make migrate-version

# Force to a known good version
make migrate-force VERSION=2

# Then continue normally
make migrate-up
```

### Migration Validation Errors

```bash
# Validate all migrations
make migrate-validate

# Common issues:
# - Missing up or down files
# - Empty migration files
# - Version number gaps
```

### Connection Issues

```bash
# Check database health
make db-health

# Ensure Docker services are running
make docker-up

# Check environment variables
cat .env
```

## Development Workflow

### Adding a New Migration

1. Create the migration files:
   ```bash
   make migrate-create NAME="add user preferences table"
   ```

2. Edit the generated files:
   - `migrations/004_add_user_preferences_table.up.sql`
   - `migrations/004_add_user_preferences_table.down.sql`

3. Test the migration:
   ```bash
   make migrate-up
   make migrate-down
   make migrate-up
   ```

4. Validate the migration:
   ```bash
   make migrate-validate
   ```

5. Test with seeded data:
   ```bash
   make db-seed
   ```

### Testing Migrations

The project includes comprehensive migration tests:

```bash
# Run all migration tests
make test-migrations

# Run specific Go tests
go test -v ./internal/shared/database -run TestMigration

# Run integration tests
go test -v ./internal/shared/database -run TestIntegration
```

## Production Considerations

### Backup Before Migration

Always backup your database before running migrations in production:

```bash
# PostgreSQL backup
pg_dump -h localhost -U postgres -d mydb > backup.sql
```

### Zero-Downtime Migrations

For production systems, consider:
- Backward-compatible schema changes
- Feature flags for new functionality
- Gradual rollouts

### Monitoring

Monitor migration performance and impact:
- Migration execution time
- Database locks and blocking
- Application error rates during migration

## Configuration

Migration configuration is handled through environment variables:

```bash
# Database connection
DATABASE_URL=postgres://user:pass@localhost:5432/dbname

# Or individual components
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=postgres
DATABASE_PASSWORD=postgres
DATABASE_NAME=go_templ_template
DATABASE_SSLMODE=disable
```

## Files Structure

```
migrations/
├── README.md                           # This file
├── 001_initial_schema.up.sql          # Initial database schema
├── 001_initial_schema.down.sql        # Rollback initial schema
├── 002_add_session_is_active.up.sql   # Add session activity tracking
├── 002_add_session_is_active.down.sql # Rollback session changes
├── 003_create_audit_events.up.sql     # Add audit logging
└── 003_create_audit_events.down.sql   # Rollback audit logging
```

## Related Documentation

- [Development Setup](../docs/development.md)
- [Database Documentation](../docs/database.md)
- [API Documentation](../README.md)

## Support

If you encounter issues with migrations:

1. Check the [troubleshooting section](#troubleshooting)
2. Run the test suite: `make test-migrations`
3. Check the logs for detailed error messages
4. Consult the [golang-migrate documentation](https://github.com/golang-migrate/migrate)