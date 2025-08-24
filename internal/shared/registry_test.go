package shared

import (
	"context"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestModuleRegistry_Register(t *testing.T) {
	// Arrange
	eventBus := &MockEventBus{}
	router := echo.New()
	registry := NewModuleRegistry(eventBus, nil, nil, router)
	module := NewMockModule("test-module")

	// Act
	err := registry.Register(module)

	// Assert
	assert.NoError(t, err)

	retrievedModule, exists := registry.GetModule("test-module")
	assert.True(t, exists)
	assert.Equal(t, module, retrievedModule)
}

func TestModuleRegistry_Register_DuplicateModule(t *testing.T) {
	// Arrange
	eventBus := &MockEventBus{}
	router := echo.New()
	registry := NewModuleRegistry(eventBus, nil, nil, router)
	module1 := NewMockModule("test-module")
	module2 := NewMockModule("test-module")

	// Act
	err1 := registry.Register(module1)
	err2 := registry.Register(module2)

	// Assert
	assert.NoError(t, err1)
	assert.Error(t, err2)
	assert.Contains(t, err2.Error(), "already registered")
}

func TestModuleRegistry_Initialize(t *testing.T) {
	// Arrange
	eventBus := &MockEventBus{}
	router := echo.New()
	registry := NewModuleRegistry(eventBus, nil, nil, router)

	module1 := NewMockModule("user")
	module2 := NewMockModule("auth")

	// Set up expectations
	module1.On("Initialize", mock.Anything, mock.Anything).Return(nil)
	module2.On("Initialize", mock.Anything, mock.Anything).Return(nil)

	registry.Register(module1)
	registry.Register(module2)

	// Act
	ctx := context.Background()
	err := registry.Initialize(ctx)

	// Assert
	assert.NoError(t, err)
	module1.AssertExpectations(t)
	module2.AssertExpectations(t)
}

func TestModuleRegistry_Initialize_ModuleFailure(t *testing.T) {
	// Arrange
	eventBus := &MockEventBus{}
	router := echo.New()
	registry := NewModuleRegistry(eventBus, nil, nil, router)

	module := NewMockModule("test-module")

	// Set up expectations - module initialization fails
	module.On("Initialize", mock.Anything, mock.Anything).Return(assert.AnError)

	registry.Register(module)

	// Act
	ctx := context.Background()
	err := registry.Initialize(ctx)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize module test-module")
	module.AssertExpectations(t)
}

func TestModuleRegistry_RegisterEventHandlers(t *testing.T) {
	// Arrange
	eventBus := &MockEventBus{}
	router := echo.New()
	registry := NewModuleRegistry(eventBus, nil, nil, router)

	module1 := NewMockModule("module1")
	module2 := NewMockModule("module2")

	// Set up expectations
	module1.On("RegisterEventHandlers", eventBus).Return(nil)
	module2.On("RegisterEventHandlers", eventBus).Return(nil)

	registry.Register(module1)
	registry.Register(module2)

	// Act
	err := registry.RegisterEventHandlers()

	// Assert
	assert.NoError(t, err)
	module1.AssertExpectations(t)
	module2.AssertExpectations(t)
}

func TestModuleRegistry_RegisterEventHandlers_Failure(t *testing.T) {
	// Arrange
	eventBus := &MockEventBus{}
	router := echo.New()
	registry := NewModuleRegistry(eventBus, nil, nil, router)

	module := NewMockModule("test-module")

	// Set up expectations - event handler registration fails
	module.On("RegisterEventHandlers", eventBus).Return(assert.AnError)

	registry.Register(module)

	// Act
	err := registry.RegisterEventHandlers()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to register event handlers for module test-module")
	module.AssertExpectations(t)
}

func TestModuleRegistry_Health(t *testing.T) {
	// Arrange
	eventBus := &MockEventBus{}
	router := echo.New()
	registry := NewModuleRegistry(eventBus, nil, nil, router)

	module1 := NewMockModule("module1")
	module2 := NewMockModule("module2")

	// Set up expectations
	module1.On("Health", mock.Anything).Return(nil)
	module2.On("Health", mock.Anything).Return(assert.AnError)

	registry.Register(module1)
	registry.Register(module2)

	// Act
	ctx := context.Background()
	healthStatus := registry.Health(ctx)

	// Assert
	assert.Len(t, healthStatus, 2)
	assert.NoError(t, healthStatus["module1"])
	assert.Error(t, healthStatus["module2"])
	module1.AssertExpectations(t)
	module2.AssertExpectations(t)
}

func TestModuleRegistry_Shutdown(t *testing.T) {
	// Arrange
	eventBus := &MockEventBus{}
	router := echo.New()
	registry := NewModuleRegistry(eventBus, nil, nil, router)

	module1 := NewMockModule("module1")
	module2 := NewMockModule("module2")

	// Set up expectations - modules should be shut down in reverse order
	module1.On("Shutdown", mock.Anything).Return(nil)
	module2.On("Shutdown", mock.Anything).Return(nil)

	registry.Register(module1)
	registry.Register(module2)

	// Act
	ctx := context.Background()
	err := registry.Shutdown(ctx)

	// Assert
	assert.NoError(t, err)
	module1.AssertExpectations(t)
	module2.AssertExpectations(t)
}

func TestModuleRegistry_GetAllModules(t *testing.T) {
	// Arrange
	eventBus := &MockEventBus{}
	router := echo.New()
	registry := NewModuleRegistry(eventBus, nil, nil, router)

	module1 := NewMockModule("module1")
	module2 := NewMockModule("module2")

	registry.Register(module1)
	registry.Register(module2)

	// Act
	allModules := registry.GetAllModules()

	// Assert
	assert.Len(t, allModules, 2)

	// Verify that modifying the returned slice doesn't affect the registry
	allModules[0] = NewMockModule("modified")
	originalModules := registry.GetAllModules()
	assert.NotEqual(t, "modified", originalModules[0].Name())
}

func TestModuleRegistry_SortModulesByDependency(t *testing.T) {
	// Arrange
	eventBus := &MockEventBus{}
	router := echo.New()
	registry := NewModuleRegistry(eventBus, nil, nil, router)

	// Register modules in random order
	authModule := NewMockModule("auth")
	userModule := NewMockModule("user")
	otherModule := NewMockModule("other")

	registry.Register(authModule)
	registry.Register(otherModule)
	registry.Register(userModule)

	// Act
	sorted := registry.sortModulesByDependency()

	// Assert
	assert.Len(t, sorted, 3)

	// User module should come first
	assert.Equal(t, "user", sorted[0].Name())

	// Auth module should come second
	assert.Equal(t, "auth", sorted[1].Name())

	// Other modules should come last
	assert.Equal(t, "other", sorted[2].Name())
}
