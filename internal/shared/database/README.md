# Database Repository Patterns and Utilities

This package provides a comprehensive set of database utilities, repository patterns, and transaction management for the Go Templ Template project.

## Features

- **Generic Repository Pattern**: Type-safe CRUD operations with SQLx
- **Transaction Management**: Robust transaction handling with proper rollback
- **Error Handling**: Comprehensive error types and utilities
- **Query Builder**: Fluent interface for building SQL queries
- **Test Utilities**: Complete test setup and utilities for database testing
- **Connection Management**: Database connection pooling and health checks

## Components

### Base Repository

The `BaseRepository` provides a generic foundation for all repositories with type-safe CRUD operations:

```go
type Repository[T any, ID comparable] interface {
    Create(ctx context.Context, entity *T) error
    GetByID(ctx context.Context, id ID) (*T, error)
    Update(ctx context.Context, entity *T) error
    Delete(ctx context.Context, id ID) error
    List(ctx context.Context, limit, offset int) ([]*T, error)
    Count(ctx context.Context) (int64, error)
    Exists(ctx context.Context, id ID) (bool, error)
}
```

### Transaction Management

The transaction utilities provide robust transaction handling:

```go
// Execute operations within a transaction
err := ExecuteInTransaction(ctx, db, func(txCtx context.Context) error {
    // All operations within this function are transactional
    if err := repo.Create(txCtx, entity1); err != nil {
        return err // Will rollback
    }
    
    if err := repo.Create(txCtx, entity2); err != nil {
        return err // Will rollback both operations
    }
    
    return nil // Will commit both operations
})
```

### Error Handling

Comprehensive error types for database operations:

```go
var (
    ErrNotFound            = errors.New("entity not found")
    ErrDuplicateKey        = errors.New("duplicate key violation")
    ErrForeignKeyViolation = errors.New("foreign key constraint violation")
    ErrOptimisticLock      = errors.New("optimistic locking conflict")
)

// Check error types
if IsNotFoundError(err) {
    // Handle not found
}
```

## Usage Examples

### Creating a Repository

```go
// Define your entity
type User struct {
    ID        string    `db:"id" json:"id"`
    Email     string    `db:"email" json:"email"`
    FirstName string    `db:"first_name" json:"first_name"`
    LastName  string    `db:"last_name" json:"last_name"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
    UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
    Version   int       `db:"version" json:"version"`
}

// Implement repository
type UserRepository struct {
    *BaseRepository[User, string]
}

func NewUserRepository(db *DB) *UserRepository {
    return &UserRepository{
        BaseRepository: NewBaseRepository[User, string](db, "users", "id"),
    }
}

// Implement CRUD operations
func (r *UserRepository) Create(ctx context.Context, user *User) error {
    query := `
        INSERT INTO users (id, email, first_name, last_name, created_at, updated_at, version)
        VALUES (:id, :email, :first_name, :last_name, NOW(), NOW(), 1)
        RETURNING created_at, updated_at`
    
    return r.BaseRepository.Create(ctx, user, query)
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*User, error) {
    query := `SELECT * FROM users WHERE id = $1`
    return r.BaseRepository.GetByID(ctx, id, query)
}
```

### Using Transactions

```go
// Service with transaction support
type UserService struct {
    userRepo *UserRepository
    tm       *TransactionManager
}

func (s *UserService) CreateUserWithProfile(ctx context.Context, user *User, profile *Profile) error {
    return s.tm.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
        // Create user
        if err := s.userRepo.Create(txCtx, user); err != nil {
            return err
        }
        
        // Create profile (linked to user)
        profile.UserID = user.ID
        if err := s.profileRepo.Create(txCtx, profile); err != nil {
            return err // Will rollback user creation too
        }
        
        return nil // Commits both operations
    })
}
```

### Query Builder

```go
// Build complex queries fluently
qb := NewQueryBuilder()
query, args := qb.
    Select("id, email, first_name, last_name").
    From("users").
    Where("status = ?", "active").
    And("created_at > ?", time.Now().AddDate(0, -1, 0)).
    OrderBy("created_at", "DESC").
    Limit(10).
    Offset(0).
    Build()

var users []*User
err := db.SelectContext(ctx, &users, query, args...)
```

### Testing

```go
func TestUserRepository(t *testing.T) {
    // Skip if no test database available within 5 seconds
    SkipIfNoDatabaseWithTimeout(t, 5*time.Second)
    
    // Set up test database with custom timeout
    testDB := NewTestDatabaseWithTimeout(t, 10*time.Second)
    defer testDB.Close()
    
    // Create test table
    testDB.CreateTable(`
        CREATE TABLE users (
            id VARCHAR PRIMARY KEY,
            email VARCHAR UNIQUE NOT NULL,
            first_name VARCHAR NOT NULL,
            last_name VARCHAR NOT NULL,
            created_at TIMESTAMP DEFAULT NOW(),
            updated_at TIMESTAMP DEFAULT NOW(),
            version INTEGER DEFAULT 1
        )
    `)
    
    // Test repository operations
    repo := NewUserRepository(testDB.DB)
    
    user := &User{
        ID:        "test-id",
        Email:     "test@example.com",
        FirstName: "John",
        LastName:  "Doe",
    }
    
    err := repo.Create(context.Background(), user)
    assert.NoError(t, err)
    
    retrieved, err := repo.GetByID(context.Background(), user.ID)
    assert.NoError(t, err)
    assert.Equal(t, user.Email, retrieved.Email)
}

// Example of conditional testing based on database availability
func TestUserRepositoryIntegration(t *testing.T) {
    // Check if database is available before expensive setup
    if !IsTestDatabaseAvailable(2*time.Second) {
        t.Skip("Database not available for integration test")
    }
    
    // Use test suite with timeout
    suite := NewTestSuiteWithTimeout(t, 15*time.Second)
    defer suite.Teardown()
    
    suite.Setup()
    suite.CreateTestEntitiesTable()
    
    // Run integration tests...
}
```

## Configuration

### Test Database Setup

Set environment variables for test database configuration:

```bash
export TEST_DB_HOST=localhost
export TEST_DB_PORT=5432
export TEST_DB_USER=postgres
export TEST_DB_PASSWORD=postgres
export TEST_DB_NAME=test_db
export TEST_DB_SSLMODE=disable
export TEST_DB_TIMEOUT=30s
```

### Timeout Configuration

The test utilities now support configurable timeouts for database connections:

```go
// Use default timeout (30s or TEST_DB_TIMEOUT env var)
testDB := NewTestDatabase(t)

// Use custom timeout
testDB := NewTestDatabaseWithTimeout(t, 10*time.Second)

// Skip test if database not available within timeout
SkipIfNoDatabaseWithTimeout(t, 5*time.Second)

// Check database availability conditionally
if IsTestDatabaseAvailable(2*time.Second) {
    // Run integration tests
} else {
    // Use mocks or skip
}

// Try to connect without failing the test
testDB := TryConnectWithTimeout(3*time.Second)
if testDB != nil {
    defer testDB.Close()
    // Use real database
} else {
    // Use mock implementations
}
```

### Connection Options

Configure database connection pooling:

```go
opts := ConnectionOptions{
    MaxOpenConns:    25,
    MaxIdleConns:    5,
    ConnMaxLifetime: 5 * time.Minute,
    ConnMaxIdleTime: 5 * time.Minute,
}

db, err := NewConnection(cfg, opts)
```

## Best Practices

### Repository Implementation

1. **Use Prepared Statements**: Always use parameterized queries to prevent SQL injection
2. **Handle Transactions**: Use the transaction context when available
3. **Optimistic Locking**: Include version fields for concurrent update safety
4. **Error Handling**: Return appropriate error types for different scenarios

### Transaction Management

1. **Keep Transactions Short**: Minimize transaction duration to reduce lock contention
2. **Handle Rollbacks**: Always handle rollback scenarios properly
3. **Nested Transactions**: The system handles nested transactions automatically
4. **Context Propagation**: Always pass transaction context through the call chain

### Testing

1. **Skip When No DB**: Use `SkipIfNoDatabase(t)` or `SkipIfNoDatabaseWithTimeout(t, timeout)` for integration tests
2. **Configure Timeouts**: Use appropriate timeouts for different test scenarios (unit tests: short, integration tests: longer)
3. **Clean Up**: Always clean up test data between tests
4. **Test Transactions**: Test both success and rollback scenarios
5. **Use Test Utilities**: Leverage the provided test utilities for consistency
6. **Conditional Testing**: Use `IsTestDatabaseAvailable()` to conditionally run tests or switch to mocks

## Error Handling

The package provides comprehensive error handling:

```go
// Check specific error types
if IsNotFoundError(err) {
    return http.StatusNotFound, "User not found"
}

if IsDuplicateKeyError(err) {
    return http.StatusConflict, "Email already exists"
}

if IsOptimisticLockError(err) {
    return http.StatusConflict, "User was modified by another process"
}

// Wrap database errors with context
if err != nil {
    return NewDatabaseError("create_user", "users", err)
}
```

## Integration with Modules

This database infrastructure integrates seamlessly with the modular architecture:

```go
// In a module's repository
type UserModule struct {
    userRepo *UserRepository
    tm       *TransactionManager
}

func (m *UserModule) Initialize(db *DB) error {
    m.tm = NewTransactionManager(db)
    m.userRepo = NewUserRepository(db)
    return nil
}
```

The repository patterns and transaction utilities provide a solid foundation for building robust, maintainable database operations in the Go Templ Template project.