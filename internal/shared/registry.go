package shared

import (
	"context"
	"fmt"
	"sort"

	"go-templ-template/internal/shared/events"

	"github.com/labstack/echo/v4"
)

// ModuleRegistry manages the registration and lifecycle of all application modules
type ModuleRegistry struct {
	container *ModuleContainer
	modules   []Module
	router    *echo.Echo
}

// NewModuleRegistry creates a new module registry
func NewModuleRegistry(eventBus events.EventBus, db interface{}, config interface{}, router *echo.Echo) *ModuleRegistry {
	return &ModuleRegistry{
		container: NewModuleContainer(eventBus, db, config),
		modules:   make([]Module, 0),
		router:    router,
	}
}

// Register registers a module with the registry
func (r *ModuleRegistry) Register(module Module) error {
	// Check if module is already registered
	for _, existingModule := range r.modules {
		if existingModule.Name() == module.Name() {
			return fmt.Errorf("module %s is already registered", module.Name())
		}
	}

	// Register module in container
	if err := r.container.RegisterModule(module); err != nil {
		return fmt.Errorf("failed to register module %s in container: %w", module.Name(), err)
	}

	// Add to modules list
	r.modules = append(r.modules, module)

	return nil
}

// Initialize initializes all registered modules in dependency order
func (r *ModuleRegistry) Initialize(ctx context.Context) error {
	// Sort modules by dependency order
	// For now, we'll use a simple approach where user module comes before auth module
	// In a more complex system, you might want to implement a proper dependency graph
	sortedModules := r.sortModulesByDependency()

	// Initialize each module
	for _, module := range sortedModules {
		if err := module.Initialize(ctx, r.container); err != nil {
			return fmt.Errorf("failed to initialize module %s: %w", module.Name(), err)
		}
	}

	return nil
}

// RegisterRoutes registers all module routes with the router
func (r *ModuleRegistry) RegisterRoutes() error {
	// Create API v1 group
	v1 := r.router.Group("/api/v1")

	// Register routes for each module
	for _, module := range r.modules {
		module.RegisterRoutes(v1)
	}

	return nil
}

// RegisterEventHandlers registers all module event handlers with the event bus
func (r *ModuleRegistry) RegisterEventHandlers() error {
	for _, module := range r.modules {
		if err := module.RegisterEventHandlers(r.container.EventBus); err != nil {
			return fmt.Errorf("failed to register event handlers for module %s: %w", module.Name(), err)
		}
	}

	return nil
}

// Start starts all modules (initialize, register routes, register event handlers)
func (r *ModuleRegistry) Start(ctx context.Context) error {
	// Initialize modules
	if err := r.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize modules: %w", err)
	}

	// Register routes
	if err := r.RegisterRoutes(); err != nil {
		return fmt.Errorf("failed to register routes: %w", err)
	}

	// Register event handlers
	if err := r.RegisterEventHandlers(); err != nil {
		return fmt.Errorf("failed to register event handlers: %w", err)
	}

	return nil
}

// Health checks the health of all modules
func (r *ModuleRegistry) Health(ctx context.Context) map[string]error {
	healthStatus := make(map[string]error)

	for _, module := range r.modules {
		healthStatus[module.Name()] = module.Health(ctx)
	}

	return healthStatus
}

// Shutdown gracefully shuts down all modules
func (r *ModuleRegistry) Shutdown(ctx context.Context) error {
	// Shutdown modules in reverse order
	for i := len(r.modules) - 1; i >= 0; i-- {
		module := r.modules[i]
		if err := module.Shutdown(ctx); err != nil {
			// Log error but continue shutting down other modules
			fmt.Printf("Warning: failed to shutdown module %s: %v\n", module.Name(), err)
		}
	}

	return nil
}

// GetModule retrieves a module by name
func (r *ModuleRegistry) GetModule(name string) (Module, bool) {
	return r.container.GetModule(name)
}

// GetAllModules returns all registered modules
func (r *ModuleRegistry) GetAllModules() []Module {
	// Return a copy to prevent external modification
	result := make([]Module, len(r.modules))
	copy(result, r.modules)
	return result
}

// GetContainer returns the module container
func (r *ModuleRegistry) GetContainer() *ModuleContainer {
	return r.container
}

// sortModulesByDependency sorts modules by their dependency order
// This is a simple implementation - in a more complex system you might want
// to implement a proper topological sort based on declared dependencies
func (r *ModuleRegistry) sortModulesByDependency() []Module {
	sorted := make([]Module, len(r.modules))
	copy(sorted, r.modules)

	// Simple dependency order: user module first, then auth module, then others
	sort.Slice(sorted, func(i, j int) bool {
		nameI := sorted[i].Name()
		nameJ := sorted[j].Name()

		// User module comes first
		if nameI == "user" {
			return true
		}
		if nameJ == "user" {
			return false
		}

		// Auth module comes after user but before others
		if nameI == "auth" {
			return true
		}
		if nameJ == "auth" {
			return false
		}

		// For other modules, maintain alphabetical order
		return nameI < nameJ
	})

	return sorted
}
