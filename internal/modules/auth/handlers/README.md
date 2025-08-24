# Auth Handlers

This package provides HTTP handlers for authentication operations in the Go Templ template application.

## Overview

The auth handlers implement RESTful endpoints for user authentication, session management, and security features including CSRF protection. All handlers follow clean architecture principles with proper error handling, validation, and testing.

## Endpoints

### Public Endpoints (No Authentication Required)

#### POST /api/v1/auth/login
Authenticates a user with email and password.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response (200 OK):**
```json
{
  "user": {
    "id": "user-123",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "full_name": "John Doe",
    "status": "active",
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z"
  },
  "session": {
    "id": "session-123",
    "expires_at": "2023-01-02T00:00:00Z",
    "created_at": "2023-01-01T00:00:00Z"
  },
  "message": "Login successful"
}
```

**Error Responses:**
- `400 Bad Request` - Validation errors
- `401 Unauthorized` - Invalid credentials
- `403 Forbidden` - Account suspended
- `429 Too Many Requests` - Rate limit exceeded

#### POST /api/v1/auth/register
Creates a new user account and logs them in.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "Password123",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Response (201 Created):**
```json
{
  "user": { /* same as login */ },
  "session": { /* same as login */ },
  "message": "Registration successful"
}
```

**Error Responses:**
- `400 Bad Request` - Validation errors
- `409 Conflict` - User already exists
- `429 Too Many Requests` - Rate limit exceeded

#### GET /api/v1/auth/validate
Validates a session token.

**Query Parameters:**
- `session_id` (optional) - Session ID to validate. If not provided, checks cookie or Authorization header.

**Response (200 OK):**
```json
{
  "valid": true,
  "user": { /* user object */ },
  "session": { /* session object */ }
}
```

**Error Responses:**
- `400 Bad Request` - Missing session ID
- `401 Unauthorized` - Invalid or expired session

#### GET /api/v1/auth/csrf-token
Generates and returns a CSRF token for form submissions.

**Response (200 OK):**
```json
{
  "csrf_token": "base64-encoded-token"
}
```

### Protected Endpoints (Authentication Required)

#### POST /api/v1/auth/logout
Terminates the current user session.

**Response (200 OK):**
```json
{
  "message": "Logout successful"
}
```

**Error Responses:**
- `401 Unauthorized` - No active session

#### GET /api/v1/auth/me
Returns information about the currently authenticated user.

**Response (200 OK):**
```json
{
  "id": "user-123",
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "full_name": "John Doe",
  "status": "active",
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2023-01-01T00:00:00Z"
}
```

#### POST /api/v1/auth/refresh
Extends the current session's expiration time.

**Response (200 OK):**
```json
{
  "message": "Session refreshed successfully",
  "data": {
    "id": "session-123",
    "expires_at": "2023-01-02T00:00:00Z",
    "created_at": "2023-01-01T00:00:00Z"
  }
}
```

#### PUT /api/v1/auth/password
Changes the current user's password.

**Request Body:**
```json
{
  "old_password": "currentPassword",
  "new_password": "NewPassword123"
}
```

**Response (200 OK):**
```json
{
  "message": "Password changed successfully. Please log in again."
}
```

**Error Responses:**
- `400 Bad Request` - Validation errors
- `401 Unauthorized` - Invalid old password

## Authentication

### Session-Based Authentication

The application uses session-based authentication with secure HTTP-only cookies:

- **Cookie Name:** `session_id`
- **Duration:** 24 hours (configurable)
- **Security:** HTTP-only, Secure (HTTPS), SameSite=Strict
- **Storage:** Database-backed sessions with cleanup

### Alternative Authentication Methods

Sessions can also be provided via:
- **Authorization Header:** `Authorization: Bearer <session_id>`
- **Direct Session ID:** For API clients

## Security Features

### CSRF Protection

Cross-Site Request Forgery protection is implemented with:
- **Token Generation:** Cryptographically secure random tokens
- **Token Validation:** Constant-time comparison to prevent timing attacks
- **Cookie Storage:** Separate CSRF cookie for validation
- **Header/Form Support:** Accepts tokens in `X-CSRF-Token` header or `_csrf_token` form field

### Rate Limiting

Rate limiting is applied to prevent abuse:
- **Login Attempts:** 5 attempts per 15 minutes per email
- **Registration:** 3 attempts per hour per IP
- **Password Changes:** 5 attempts per 15 minutes per user

### Input Validation

Comprehensive validation includes:
- **Email Format:** RFC-compliant email validation
- **Password Strength:** Minimum 8 characters, uppercase, lowercase, digit
- **Name Validation:** Letters, spaces, hyphens, apostrophes only
- **SQL Injection Prevention:** Parameterized queries
- **XSS Prevention:** Automatic escaping in templates

## Error Handling

### Error Response Format

All errors follow a consistent format:
```json
{
  "error": "ERROR_CODE",
  "message": "Human-readable error message",
  "field": "field_name" // Optional, for validation errors
}
```

### Error Codes

- `VALIDATION_ERROR` - Input validation failed
- `INVALID_CREDENTIALS` - Authentication failed
- `USER_ALREADY_EXISTS` - Registration with existing email
- `SESSION_NOT_FOUND` - Session doesn't exist
- `SESSION_EXPIRED` - Session has expired
- `SESSION_INVALID` - Session is invalid
- `ACCOUNT_SUSPENDED` - User account is suspended
- `RATE_LIMIT_EXCEEDED` - Too many requests
- `CSRF_TOKEN_MISSING` - CSRF token required
- `CSRF_TOKEN_INVALID` - CSRF token validation failed
- `INTERNAL_ERROR` - Server error

## Middleware

### AuthMiddleware

Provides authentication and authorization middleware:

#### RequireAuth
Requires valid authentication for protected routes.

```go
authMiddleware := middleware.NewAuthMiddleware(authService)
protectedGroup.Use(authMiddleware.RequireAuth)
```

#### OptionalAuth
Optionally authenticates users, allowing both authenticated and anonymous access.

```go
optionalGroup.Use(authMiddleware.OptionalAuth)
```

#### RequireRole
Requires specific user status/role.

```go
adminGroup.Use(authMiddleware.RequireRole("active"))
```

#### CSRF
Provides CSRF protection.

```go
formGroup.Use(authMiddleware.CSRF)
```

### Context Helpers

Retrieve user and session information from request context:

```go
user := middleware.GetUserFromContext(c)
session := middleware.GetSessionFromContext(c)
```

## Testing

### Unit Tests

Individual handler methods are tested with mocked dependencies:
- Input validation
- Success scenarios
- Error conditions
- Edge cases

### Integration Tests

Full request/response cycle testing:
- Route registration
- Middleware integration
- End-to-end flows
- Error handling

### Test Coverage

- **Handlers:** 100% line coverage
- **Middleware:** 100% line coverage
- **Validation:** All validation rules tested
- **Error Scenarios:** All error paths covered

## Usage Examples

### Basic Setup

```go
// Create auth service
authService := application.NewAuthService(/* dependencies */)

// Register routes
handlers.RegisterAuthRoutes(e, authService)

// Use middleware
authMiddleware := middleware.NewAuthMiddleware(authService)
protectedGroup := e.Group("/protected")
protectedGroup.Use(authMiddleware.RequireAuth)
```

### Custom CSRF Configuration

```go
csrfConfig := middleware.CSRFConfig{
    TokenLength:    32,
    CookieName:     "csrf_token",
    HeaderName:     "X-CSRF-Token",
    CookieMaxAge:   3600,
    CookieSecure:   true, // Enable in production
}

csrfMiddleware := middleware.NewCSRFMiddleware(csrfConfig)
formGroup.Use(csrfMiddleware.Protect)
```

### Error Handling

```go
func (h *AuthHandler) handleApplicationError(c echo.Context, err error) error {
    if appErr, ok := err.(*application.AuthError); ok {
        statusCode := h.getStatusCodeForError(appErr.Code)
        return c.JSON(statusCode, ErrorResponse{
            Error:   appErr.Code,
            Message: appErr.Message,
            Field:   appErr.Field,
        })
    }
    
    return c.JSON(http.StatusInternalServerError, ErrorResponse{
        Error:   "INTERNAL_ERROR",
        Message: "An unexpected error occurred",
    })
}
```

## Configuration

### Session Configuration

```go
sessionConfig := domain.SessionConfig{
    DefaultDuration: time.Hour * 24,     // 24 hours
    MaxDuration:     time.Hour * 24 * 7, // 7 days max
    CleanupInterval: time.Hour,          // cleanup every hour
}
```

### Rate Limiter Configuration

```go
rateLimiterConfig := application.RateLimiterConfig{
    MaxAttempts: 5,                // 5 attempts
    Window:      time.Minute * 15, // within 15 minutes
    LockoutTime: time.Minute * 30, // locked out for 30 minutes
}
```

## Best Practices

1. **Always validate input** - Use the validation functions for all user input
2. **Handle errors gracefully** - Provide meaningful error messages without exposing internals
3. **Use HTTPS in production** - Enable secure cookies and CSRF protection
4. **Implement rate limiting** - Protect against brute force attacks
5. **Log security events** - Monitor authentication failures and suspicious activity
6. **Regular session cleanup** - Remove expired sessions to prevent database bloat
7. **Test thoroughly** - Maintain high test coverage for security-critical code

## Dependencies

- **Echo Framework** - HTTP router and middleware
- **Application Layer** - Business logic and commands
- **Domain Layer** - User and session entities
- **Middleware Package** - Authentication and security middleware
- **Testify** - Testing framework and assertions