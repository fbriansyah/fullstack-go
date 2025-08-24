# Router Implementation - Before and After

## Problem: Duplicate Route Definitions

Previously, we had two ways to register routes, leading to code duplication:

### 1. Handler-level routing (Original)
```go
// internal/modules/user/handlers/routes.go
func RegisterUserRoutes(e *echo.Echo, userService application.UserService) {
    // Creates its own /api/v1 group
    v1 := e.Group("/api/v1")
    users := v1.Group("/users")
    // ... route definitions
}
```

### 2. Module-level routing (Duplicated)
```go
// internal/modules/user/module.go - BEFORE (BAD)
func (m *UserModule) RegisterRoutes(router *echo.Group) {
    // Duplicated the same route logic!
    userHandler := handlers.NewUserHandler(m.userService)
    users := router.Group("/users")
    // ... same route definitions again
}
```

## Solution: Single Source of Truth

Now we have a clean separation:

### 1. Handler-level functions (Multiple variants)
```go
// For standalone usage (creates own /api/v1 group)
func RegisterUserRoutes(e *echo.Echo, userService application.UserService)

// For module system usage (uses provided group)
func RegisterUserRoutesOnGroup(group *echo.Group, userService application.UserService)

// With custom middleware
func RegisterUserRoutesWithMiddleware(e *echo.Echo, userService application.UserService, middlewares ...echo.MiddlewareFunc)
```

### 2. Module-level routing (Delegates to handlers)
```go
// internal/modules/user/module.go - AFTER (GOOD)
func (m *UserModule) RegisterRoutes(router *echo.Group) {
    // Delegates to the handler - no duplication!
    handlers.RegisterUserRoutesOnGroup(router, m.userService)
}
```

## Usage Examples

### Standalone Application (without module system)
```go
func main() {
    e := echo.New()
    userService := setupUserService()
    
    // Use the Echo-based function
    handlers.RegisterUserRoutes(e, userService)
    
    e.Start(":8080")
}
```

### Module-based Application
```go
func main() {
    e := echo.New()
    registry := shared.NewModuleRegistry(eventBus, db, config, e)
    
    // Register modules
    registry.Register(user.NewUserModule())
    registry.Register(auth.NewAuthModule())
    
    // Start modules (automatically registers routes using the group-based functions)
    registry.Start(ctx)
    
    e.Start(":8080")
}
```

## Benefits

✅ **Single Source of Truth**: Route definitions exist in one place  
✅ **No Code Duplication**: Module system reuses handler functions  
✅ **Flexibility**: Can use either approach depending on needs  
✅ **Maintainability**: Changes only need to be made in one place  
✅ **Consistency**: Both approaches produce the same routes  

## Route Structure

Both approaches produce the same final routes:
- `POST /api/v1/users`
- `GET /api/v1/users`
- `GET /api/v1/users/:id`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/register`
- etc.