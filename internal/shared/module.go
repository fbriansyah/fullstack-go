package shared

import (
	"context"

	"go-templ-template/internal/shared/events"

	"github.com/labstack/echo/v4"
)

// Module defines the interface that all application modules must implement
type Module interface {
	// Name returns the unique name of the module
	Name() string

	// Initialize sets up the module with its dependencies
	Initialize(ctx context.Context, container *ModuleContainer) error

	// RegisterRoutes registers the module's HTTP routes with the router
	RegisterRoutes(router *echo.Group)

	// RegisterEventHandlers registers the module's event handlers with the event bus
	RegisterEventHandlers(eventBus events.EventBus) error

	// Health returns the health status of the module
	Health(ctx context.Context) error

	// Shutdown gracefully shuts down the module
	Shutdown(ctx context.Context) error
}

// ModuleContainer provides dependency injection for modules
type ModuleContainer struct {
	// Core dependencies
	EventBus events.EventBus
	DB       interface{} // Using interface{} to avoid circular imports

	// Configuration
	Config interface{}

	// Module registry for inter-module communication
	modules map[string]Module
}

// NewModuleContainer creates a new module container
func NewModuleContainer(eventBus events.EventBus, db interface{}, config interface{}) *ModuleContainer {
	return &ModuleContainer{
		EventBus: eventBus,
		DB:       db,
		Config:   config,
		modules:  make(map[string]Module),
	}
}

// RegisterModule registers a module in the container
func (c *ModuleContainer) RegisterModule(module Module) error {
	if _, exists := c.modules[module.Name()]; exists {
		return NewModuleError(module.Name(), "module already registered")
	}
	c.modules[module.Name()] = module
	return nil
}

// GetModule retrieves a registered module by name
func (c *ModuleContainer) GetModule(name string) (Module, bool) {
	module, exists := c.modules[name]
	return module, exists
}

// GetAllModules returns all registered modules
func (c *ModuleContainer) GetAllModules() map[string]Module {
	// Return a copy to prevent external modification
	result := make(map[string]Module)
	for name, module := range c.modules {
		result[name] = module
	}
	return result
}

// ModuleError represents an error that occurred during module operations
type ModuleError struct {
	ModuleName string
	Message    string
	Cause      error
}

// NewModuleError creates a new module error
func NewModuleError(moduleName, message string) *ModuleError {
	return &ModuleError{
		ModuleName: moduleName,
		Message:    message,
	}
}

// NewModuleErrorWithCause creates a new module error with a cause
func NewModuleErrorWithCause(moduleName, message string, cause error) *ModuleError {
	return &ModuleError{
		ModuleName: moduleName,
		Message:    message,
		Cause:      cause,
	}
}

// Error implements the error interface
func (e *ModuleError) Error() string {
	if e.Cause != nil {
		return e.ModuleName + ": " + e.Message + ": " + e.Cause.Error()
	}
	return e.ModuleName + ": " + e.Message
}

// Unwrap returns the underlying cause
func (e *ModuleError) Unwrap() error {
	return e.Cause
}
