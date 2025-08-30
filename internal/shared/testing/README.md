# Mock Implementations for Unit Testing

This package provides mock implementations of key services and repositories for unit testing without external dependencies.

## Overview

The mock implementations allow you to:
- Test business logic without database connections
- Test event handling without message brokers
- Test rate limiting without external rate limiters
- Create predictable test scenarios
- Verify service interactions

## Available Mocks

### MockSessionRepository
Mock implementation of `SessionRepository` interface with in-memory storage.

**Features:**
- In-memory session storage
- Thread-safe operations
- Automatic session filtering by expiration and active status
- Full CRUD operations
- Fallback behavior when no explicit mocks are set

### MockUserRepository
Mock implementation of `UserRepository` interface with in-memory storage.

**Features:**
- In-memory user storage
- Thread-safe operations
- Full CRUD operations with filtering
- Fallback behavior when no explicit mocks are set
- Support for preloaded test data

### MockEventBus
Mock implementation of `EventBus` interface for testing event publishing.

**Features:**
- Tracks all published events
- Supports event handler registration
- Provides methods to inspect published events
- Can clear events between tests
- Auto-capture mode for integration tests

### MockRateLimiter
Mock implementation of `RateLimiter` interface for testing rate limiting.

**Features:**
- In-memory attempt tracking
- Configurable allow/deny behavior
- Attempt count management
- Reset functionality
- Predefined limits for specific keys

### MockUserService
Mock implementation of `UserService` interface for testing user operations.

**Features:**
- Full user service interface coverage
- Configurable return values and errors
- Integration with testify/mock for expectations

### MockAuditLogger
Mock implementation of `AuditLogger` interface for testing audit logging.

**Features:**
- In-memory event storage
- Thread-safe operations
- Event filtering capabilities
- Methods to inspect logged events
- Can clear events between tests

## Usage Examples

### Basic Mock Setup

```go
func TestMyService(t *testing.T) {
    // Create mock services
    mocks := NewMockServices()
    
    // Setup default behavior (all operations succeed)
    SetupDefaultMockBehavior(mocks)
    
    // Create service under test
    authService := NewMockAuthService(mocks)
    
    // Your test logic here...
}
```

### Custom Mock Behavior

```go
func TestLoginFailure(t *testing.T) {
    mocks := NewMockServices()
    
    // Setup specific behavior
    userQuery := &userApp.GetUserByEmailQuery{Email: "test@example.com"}
    mocks.UserService.On("GetUserByEmail", mock.Anything, userQuery).
        Return(nil, userApp.NewUserNotFoundError("test@example.com"))
    
    // Test the failure scenario
    authService := NewMockAuthService(mocks)
    result, err := authService.Login(ctx, loginCmd)
    
    assert.Error(t, err)
    assert.Nil(t, result)
}
```

### Verifying Interactions

```go
func TestEventPublishing(t *testing.T) {
    mocks := NewMockServices()
    SetupDefaultMockBehavior(mocks)
    
    // Execute operation that should publish events
    authService := NewMockAuthService(mocks)
    authService.Login(ctx, loginCmd)
    
    // Verify events were published
    publishedEvents := mocks.EventBus.GetPublishedEvents()
    assert.Len(t, publishedEvents, 1)
    assert.Equal(t, "auth.user.logged_in", publishedEvents[0].EventType())
    
    // Verify mock expectations
    mocks.EventBus.AssertExpectations(t)
}
```

### Using Preloaded Data (Fallback Behavior)

```go
func TestWithPreloadedData(t *testing.T) {
    // Create test data
    user1 := CreateTestUser("user-1", "user1@example.com")
    user2 := CreateTestUser("user-2", "user2@example.com")
    
    // Create repository with preloaded data
    userRepo := NewMockUserRepositoryWithData(user1, user2)
    
    // No explicit mock setup needed - uses internal storage
    retrievedUser, err := userRepo.GetByID(ctx, "user-1")
    assert.NoError(t, err)
    assert.Equal(t, user1.ID, retrievedUser.ID)
}
```

### Specialized Factory Functions

```go
func TestSpecializedMocks(t *testing.T) {
    // Event bus that captures all events without explicit setup
    eventBus := NewMockEventBusWithCapture()
    
    // Rate limiter with predefined limits
    allowedKeys := []string{"user-123"}
    blockedKeys := []string{"user-456"}
    rateLimiter := NewMockRateLimiterWithLimits(allowedKeys, blockedKeys)
    
    // Test allowed user
    allowed, err := rateLimiter.Allow(ctx, "user-123")
    assert.True(t, allowed)
    
    // Test blocked user
    blocked, err := rateLimiter.Allow(ctx, "user-456")
    assert.False(t, blocked)
}
```

### Rate Limiting Tests

```go
func TestRateLimiting(t *testing.T) {
    mocks := NewMockServices()
    
    // Setup rate limiter to deny requests
    rateLimitError := application.NewRateLimitExceededError("Too many attempts")
    mocks.RateLimiter.On("Allow", mock.Anything, "login:test@example.com").
        Return(false, rateLimitError)
    
    // Test rate limited scenario
    authService := NewMockAuthService(mocks)
    result, err := authService.Login(ctx, loginCmd)
    
    assert.Error(t, err)
    assert.True(t, application.IsRateLimitError(err))
}
```

## Test Data Factories

The package includes factory functions for creating test data:

### CreateTestUser
Creates a user with default test values:
```go
user := CreateTestUser("user-123", "test@example.com")
```

### CreateTestSession
Creates a valid session:
```go
session := CreateTestSession("user-123")
```

### CreateExpiredTestSession
Creates an expired session for testing expiration logic:
```go
expiredSession := CreateExpiredTestSession("user-123")
```

### CreateInactiveTestSession
Creates an inactive session:
```go
inactiveSession := CreateInactiveTestSession("user-123")
```

## Best Practices

### 1. Use Default Behavior Setup
Always call `SetupDefaultMockBehavior(mocks)` unless you need specific failure scenarios:

```go
mocks := NewMockServices()
SetupDefaultMockBehavior(mocks) // Sets up success scenarios
```

### 2. Clear State Between Tests
For event bus testing, clear published events between tests:

```go
func TestSomething(t *testing.T) {
    mocks := NewMockServices()
    mocks.EventBus.ClearPublishedEvents() // Start with clean state
    
    // Your test logic...
}
```

### 3. Verify Expectations
Always verify mock expectations to ensure proper service interactions:

```go
// At the end of your test
mocks.SessionRepo.AssertExpectations(t)
mocks.EventBus.AssertExpectations(t)
mocks.RateLimiter.AssertExpectations(t)
mocks.UserService.AssertExpectations(t)
```

### 4. Test Both Success and Failure Scenarios
Use mocks to test error conditions:

```go
// Test database error
mocks.SessionRepo.On("Create", mock.Anything, mock.Anything).
    Return(database.NewDatabaseError("Create", "sessions", errors.New("connection failed")))

// Test rate limiting
mocks.RateLimiter.On("Allow", mock.Anything, mock.Anything).
    Return(false, application.NewRateLimitExceededError("Too many attempts"))
```

### 5. Use Specific Matchers
Use specific argument matchers when needed:

```go
// Match specific user query
userQuery := &userApp.GetUserByEmailQuery{Email: "specific@example.com"}
mocks.UserService.On("GetUserByEmail", mock.Anything, userQuery).Return(user, nil)

// Match any argument
mocks.EventBus.On("Publish", mock.Anything, mock.Anything).Return(nil)
```

## Integration with Existing Tests

These mocks can be integrated into existing test suites by:

1. Replacing database-dependent repositories with mock repositories
2. Replacing event bus with mock event bus
3. Replacing rate limiter with mock rate limiter
4. Using factory functions for consistent test data

This allows existing tests to run without external dependencies while maintaining the same test coverage and behavior verification.