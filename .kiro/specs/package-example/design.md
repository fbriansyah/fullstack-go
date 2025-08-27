# Design Document

## Overview

This design addresses the compilation errors in the `examples/event_driven_workflows.go` file and ensures the package example works correctly to demonstrate event-driven workflows. The main issues are interface compatibility problems between infrastructure and application layers, missing method implementations, and incorrect service constructor arguments.

## Architecture

The package example follows a layered architecture with clear separation between:

- **Application Layer**: Defines service interfaces and business logic
- **Infrastructure Layer**: Implements data access and external integrations  
- **Domain Layer**: Contains business entities and domain logic
- **Examples Package**: Demonstrates system functionality

The current architecture has interface mismatches that prevent compilation and execution.

## Components and Interfaces

### SessionRepository Interface Alignment

**Problem**: The infrastructure `SessionRepository` interface doesn't match the application layer expectations.

**Infrastructure Interface** (`internal/modules/auth/infrastructure/repository.go`):
- Has `CleanupExpired(ctx context.Context) (int64, error)` method
- Missing `DeleteExpired(ctx context.Context) error` method expected by application layer

**Application Interface** (`internal/modules/auth/application/service.go`):
- Expects `DeleteExpired(ctx context.Context) error` method
- Does not expect `CleanupExpired` method

**Solution**: 
1. Add `DeleteExpired` method to infrastructure interface and implementation
2. Implement `DeleteExpired` as a wrapper around existing `CleanupExpired` functionality

### RateLimiter Configuration

**Problem**: `NewInMemoryRateLimiter()` called without required `RateLimiterConfig` parameter.

**Solution**: 
1. Create default rate limiter configuration
2. Pass configuration to constructor
3. Use reasonable defaults for demonstration purposes
### Service Constructor Compatibility

**Problem**: Service constructors expect specific interface types that don't match infrastructure implementations.

**Solution**:
1. Ensure infrastructure implementations satisfy application interface contracts
2. Add missing methods to infrastructure implementations
3. Maintain backward compatibility with existing functionality

## Data Models

No changes to existing data models are required. The issue is purely interface compatibility, not data structure problems.

### Session Model
- Remains unchanged
- Used by both infrastructure and application layers
- No modifications needed

### User Model  
- Remains unchanged
- Properly integrated across layers
- No interface issues identified

## Error Handling

### Compilation Error Resolution
1. **Interface Mismatch Errors**: Resolve by adding missing methods to infrastructure implementations
2. **Constructor Argument Errors**: Resolve by providing required configuration parameters
3. **Type Compatibility Errors**: Resolve by ensuring interface contracts are properly implemented

### Runtime Error Handling
- Maintain existing error handling patterns
- Ensure new methods follow established error handling conventions
- Preserve error context and logging functionality

## Testing Strategy

### Unit Testing
- Test new `DeleteExpired` method implementation
- Test rate limiter configuration creation
- Verify interface compatibility

### Integration Testing  
- Test complete workflow examples end-to-end
- Verify event-driven workflows function correctly
- Test audit trail functionality

### Example Validation
- Ensure example compiles without errors
- Verify all demonstrated workflows execute successfully
- Validate proper resource cleanup and connection management