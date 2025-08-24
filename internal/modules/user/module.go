package user

import (
	"context"

	"go-templ-template/internal/config"
	"go-templ-template/internal/modules/user/application"
	"go-templ-template/internal/modules/user/handlers"
	"go-templ-template/internal/modules/user/infrastructure"
	"go-templ-template/internal/shared"
	"go-templ-template/internal/shared/database"
	"go-templ-template/internal/shared/events"

	"github.com/labstack/echo/v4"
)

// UserModule implements the Module interface for user management functionality
type UserModule struct {
	name        string
	userService application.UserService
	userHandler *handlers.UserHandler
	eventBus    events.EventBus
	db          *database.DB
	config      *config.Config
}

// NewUserModule creates a new user module instance
func NewUserModule() *UserModule {
	return &UserModule{
		name: "user",
	}
}

// Name returns the unique name of the module
func (m *UserModule) Name() string {
	return m.name
}

// Initialize sets up the module with its dependencies
func (m *UserModule) Initialize(ctx context.Context, container *shared.ModuleContainer) error {
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

	// Initialize repository
	userRepo := infrastructure.NewUserRepository(db)

	// Initialize service
	m.userService = application.NewUserService(userRepo, m.eventBus, db)

	// Initialize handlers
	m.userHandler = handlers.NewUserHandler(m.userService)

	return nil
}

// RegisterRoutes registers the module's HTTP routes with the router
func (m *UserModule) RegisterRoutes(router *echo.Group) {
	// Use the existing handler function that works with groups
	handlers.RegisterUserRoutesOnGroup(router, m.userService)
}

// RegisterEventHandlers registers the module's event handlers with the event bus
func (m *UserModule) RegisterEventHandlers(eventBus events.EventBus) error {
	// User module primarily publishes events, but could subscribe to others
	// For example, it might listen to auth events to update user last login time

	// Register user lifecycle event handlers if needed
	// This is where we would register handlers for events from other modules
	// For now, the user module is primarily an event publisher

	return nil
}

// Health returns the health status of the module
func (m *UserModule) Health(ctx context.Context) error {
	// Check if the user service is healthy
	// This could include checking database connectivity, etc.
	if m.userService == nil {
		return shared.NewModuleError(m.name, "user service not initialized")
	}

	// Test database connectivity by attempting a simple query
	// We can use the repository to check if we can connect to the database
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
func (m *UserModule) Shutdown(ctx context.Context) error {
	// Perform any cleanup needed for the user module
	// This might include closing connections, flushing caches, etc.

	// For now, we don't have any specific cleanup to do
	// The database and event bus will be cleaned up by the main application

	return nil
}

// GetUserService returns the user service for inter-module communication
func (m *UserModule) GetUserService() application.UserService {
	return m.userService
}

// GetUserHandler returns the user handler for testing purposes
func (m *UserModule) GetUserHandler() *handlers.UserHandler {
	return m.userHandler
}
