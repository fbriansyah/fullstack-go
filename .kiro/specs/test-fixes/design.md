# Design Document

## Overview

This design addresses the test failures and build errors in the Go fullstack application. The main issues identified are:

1. **Database Connection Timeouts**: Tests are hanging when trying to connect to PostgreSQL
2. **Missing Interface Methods**: `SimpleAuthService` doesn't implement the `GetUserSessions` method required by `AuthService` interface
3. **Build Dependencies**: Some packages have circular or missing dependencies

## Architecture

The solution follows a layered approach to fix these issues systematically:

### Test Infrastructure Layer
- **Database Test Utilities**: Enhance existing test utilities with better connection handling and timeouts
- **Mock Services**: Provide mock implementations for testing without external dependencies
- **Test Configuration**: Environment-based test configuration with fallbacks

### Interface Compliance Layer
- **Method Implementation**: Add missing interface methods to ensure all implementations are complete
- **Interface Validation**: Ensure all service implementations satisfy their contracts

### Build System Layer
- **Dependency Resolution**: Fix circular dependencies and missing imports
- **Build Validation**: Ensure all packages can be built independently

## Components and Interfaces

### Enhanced Test Database Utilities

```go
type TestDatabaseManager interface {
    // Connection management with timeouts
    ConnectWithTimeout(timeout time.Duration) (*TestDatabase, error)
    
    // Skip tests when database unavailable
    SkipIfUnavailable(t *testing.T)
    
    // Mock database for unit tests
    NewMockDatabase() *MockDatabase
}
```

### Interface Compliance

```go
// Ensure SimpleAuthService implements all AuthService methods
type AuthService interface {
    Login(ctx context.Context, cmd *LoginCommand) (*AuthResult, error)
    Register(ctx context.Context, cmd *RegisterCommand) (*AuthResult, error)
    Logout(ctx context.Context, cmd *LogoutCommand) error
    ValidateSession(ctx context.Context, query *ValidateSessionQuery) (*SessionValidationResult, error)
    RefreshSession(ctx context.Context, cmd *RefreshSessionCommand) (*domain.Session, error)
    ChangePassword(ctx context.Context, cmd *ChangePasswordCommand) error
    CleanupExpiredSessions(ctx context.Context) error
    GetUserSessions(ctx context.Context, userID string) ([]*domain.Session, error) // Missing method
}
```

## Data Models

### Test Configuration
```go
type TestConfig struct {
    DatabaseURL      string
    ConnectionTimeout time.Duration
    SkipIntegration  bool
    MockServices     bool
}
```

### Mock Implementations
```go
type MockSessionRepository struct {
    sessions map[string]*domain.Session
    mu       sync.RWMutex
}
```

## Error Handling

### Database Connection Errors
- **Timeout Handling**: Implement proper timeouts for database connections
- **Graceful Degradation**: Skip tests when database is unavailable
- **Error Classification**: Distinguish between connection errors and test failures

### Build Errors
- **Interface Validation**: Compile-time checks for interface compliance
- **Dependency Management**: Clear separation of concerns to avoid circular dependencies

## Testing Strategy

### Unit Tests
- **Mock Dependencies**: Use mocks for external services (database, event bus)
- **Fast Execution**: Unit tests should run quickly without external dependencies
- **Isolation**: Each test should be independent and not affect others

### Integration Tests
- **Database Tests**: Use real database connections with proper setup/teardown
- **Environment Detection**: Skip integration tests when dependencies unavailable
- **Test Containers**: Consider using testcontainers for isolated database testing

### Test Categories
```go
// Build tags for different test types
// +build unit
// +build integration
// +build e2e
```

## Implementation Plan

### Phase 1: Fix Interface Compliance
1. Add missing `GetUserSessions` method to `SimpleAuthService`
2. Ensure all interface implementations are complete
3. Validate build success across all packages

### Phase 2: Enhance Test Infrastructure
1. Improve database connection handling with timeouts
2. Add mock implementations for unit testing
3. Implement test skipping for unavailable dependencies

### Phase 3: Test Stabilization
1. Fix hanging database tests
2. Ensure all tests pass consistently
3. Add proper test categorization (unit vs integration)

## Configuration Management

### Environment Variables
```bash
# Test database configuration
TEST_DB_HOST=localhost
TEST_DB_PORT=5432
TEST_DB_USER=postgres
TEST_DB_PASSWORD=postgres
TEST_DB_NAME=test_db
TEST_DB_TIMEOUT=30s

# Test behavior
SKIP_INTEGRATION_TESTS=false
USE_MOCK_SERVICES=false
```

### Test Modes
- **Unit Mode**: Use mocks, no external dependencies
- **Integration Mode**: Use real services with proper setup
- **CI Mode**: Skip tests requiring manual setup