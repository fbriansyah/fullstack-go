package auth

import (
	"context"
	"fmt"

	"go-templ-template/internal/config"
	"go-templ-template/internal/modules/auth/application"
	"go-templ-template/internal/modules/auth/handlers"
	"go-templ-template/internal/modules/auth/infrastructure"
	"go-templ-template/internal/modules/user"
	userApplication "go-templ-template/internal/modules/user/application"
	"go-templ-template/internal/shared"
	"go-templ-template/internal/shared/database"
	"go-templ-template/internal/shared/events"

	"github.com/labstack/echo/v4"
)

// AuthModule implements the Module interface for authentication functionality
type AuthModule struct {
	name        string
	authService application.AuthService
	authHandler *handlers.AuthHandler
	eventBus    events.EventBus
	db          *database.DB
	config      *config.Config
	userService userApplication.UserService
}

// NewAuthModule creates a new auth module instance
func NewAuthModule() *AuthModule {
	return &AuthModule{
		name: "auth",
	}
}

// Name returns the unique name of the module
func (m *AuthModule) Name() string {
	return m.name
}

// Initialize sets up the module with its dependencies
func (m *AuthModule) Initialize(ctx context.Context, container *shared.ModuleContainer) error {
	// Type assert dependencies
	db, ok := container.DB.(*database.DB)
	if !ok {
		return shared.NewModuleError(m.name, "invalid database dependency")
	}

	config, ok := container.Config.(*config.Config)
	if !ok {
		return shared.NewModuleError(m.name, "invalid config dependency")
	}

	// Store dependencies
	m.eventBus = container.EventBus
	m.db = db
	m.config = config

	// Get user service from user module
	userModule, exists := container.GetModule("user")
	if !exists {
		return shared.NewModuleError(m.name, "user module not found - auth module depends on user module")
	}

	userMod, ok := userModule.(*user.UserModule)
	if !ok {
		return shared.NewModuleError(m.name, "invalid user module type")
	}

	m.userService = userMod.GetUserService()
	if m.userService == nil {
		return shared.NewModuleError(m.name, "user service not available from user module")
	}

	// Initialize repository
	sessionRepo := infrastructure.NewSessionRepositoryAdapter(db)

	// Initialize rate limiter
	rateLimiterConfig := application.DefaultRateLimiterConfig()
	rateLimiter := application.NewInMemoryRateLimiter(rateLimiterConfig)

	// Initialize session config
	sessionConfig := application.DefaultSessionConfig()

	// Initialize service
	m.authService = application.NewAuthService(
		sessionRepo,
		m.userService,
		m.eventBus,
		db,
		rateLimiter,
		sessionConfig,
	)

	// Initialize handlers
	m.authHandler = handlers.NewAuthHandler(m.authService)

	return nil
}

// RegisterRoutes registers the module's HTTP routes with the router
func (m *AuthModule) RegisterRoutes(router *echo.Group) {
	// Use the existing handler function that works with groups
	handlers.RegisterAuthRoutesOnGroup(router, m.authService)
}

// RegisterEventHandlers registers the module's event handlers with the event bus
func (m *AuthModule) RegisterEventHandlers(eventBus events.EventBus) error {
	// Auth module can listen to user events and react accordingly
	// For example, when a user is deleted, we might want to clean up their sessions

	// Register event handlers for user lifecycle events
	userDeletedHandler := NewUserDeletedHandler(m.authService)
	if err := eventBus.Subscribe("user.deleted", userDeletedHandler); err != nil {
		return shared.NewModuleErrorWithCause(m.name, "failed to subscribe to user.deleted event", err)
	}

	// Register handler for user status changes
	userStatusChangedHandler := NewUserStatusChangedHandler(m.authService)
	if err := eventBus.Subscribe("user.status_changed", userStatusChangedHandler); err != nil {
		return shared.NewModuleErrorWithCause(m.name, "failed to subscribe to user.status_changed event", err)
	}

	return nil
}

// Health returns the health status of the module
func (m *AuthModule) Health(ctx context.Context) error {
	// Check if the auth service is healthy
	if m.authService == nil {
		return shared.NewModuleError(m.name, "auth service not initialized")
	}

	// Check if user service dependency is healthy
	if m.userService == nil {
		return shared.NewModuleError(m.name, "user service dependency not initialized")
	}

	// Test database connectivity
	if m.db == nil {
		return shared.NewModuleError(m.name, "database not initialized")
	}

	// Check if event bus is healthy
	if m.eventBus == nil {
		return shared.NewModuleError(m.name, "event bus not initialized")
	}

	if err := m.eventBus.Health(); err != nil {
		return shared.NewModuleErrorWithCause(m.name, "event bus unhealthy", err)
	}

	return nil
}

// Shutdown gracefully shuts down the module
func (m *AuthModule) Shutdown(ctx context.Context) error {
	// Perform cleanup for auth module
	// This might include cleaning up expired sessions, etc.

	if m.authService != nil {
		// Clean up expired sessions before shutdown
		if err := m.authService.CleanupExpiredSessions(ctx); err != nil {
			// Log error but don't fail shutdown
			fmt.Printf("Warning: failed to cleanup expired sessions during shutdown: %v\n", err)
		}
	}

	return nil
}

// GetAuthService returns the auth service for inter-module communication
func (m *AuthModule) GetAuthService() application.AuthService {
	return m.authService
}

// GetAuthHandler returns the auth handler for testing purposes
func (m *AuthModule) GetAuthHandler() *handlers.AuthHandler {
	return m.authHandler
}

// UserDeletedHandler handles user deleted events
type UserDeletedHandler struct {
	authService application.AuthService
}

// NewUserDeletedHandler creates a new user deleted event handler
func NewUserDeletedHandler(authService application.AuthService) *UserDeletedHandler {
	return &UserDeletedHandler{
		authService: authService,
	}
}

// Handle processes user deleted events
func (h *UserDeletedHandler) Handle(ctx context.Context, event events.DomainEvent) error {
	// When a user is deleted, clean up all their sessions
	userID := event.AggregateID()

	// We would need to add a method to delete sessions by user ID
	// For now, we'll just log that we received the event
	fmt.Printf("Handling user deleted event for user %s\n", userID)

	return nil
}

// EventType returns the event type this handler processes
func (h *UserDeletedHandler) EventType() string {
	return "user.deleted"
}

// HandlerName returns a unique name for this handler
func (h *UserDeletedHandler) HandlerName() string {
	return "auth.user_deleted_handler"
}

// UserStatusChangedHandler handles user status changed events
type UserStatusChangedHandler struct {
	authService application.AuthService
}

// NewUserStatusChangedHandler creates a new user status changed event handler
func NewUserStatusChangedHandler(authService application.AuthService) *UserStatusChangedHandler {
	return &UserStatusChangedHandler{
		authService: authService,
	}
}

// Handle processes user status changed events
func (h *UserStatusChangedHandler) Handle(ctx context.Context, event events.DomainEvent) error {
	// When a user status changes (e.g., suspended), we might want to invalidate their sessions
	userID := event.AggregateID()

	// We would need to check the new status and potentially invalidate sessions
	// For now, we'll just log that we received the event
	fmt.Printf("Handling user status changed event for user %s\n", userID)

	return nil
}

// EventType returns the event type this handler processes
func (h *UserStatusChangedHandler) EventType() string {
	return "user.status_changed"
}

// HandlerName returns a unique name for this handler
func (h *UserStatusChangedHandler) HandlerName() string {
	return "auth.user_status_changed_handler"
}
