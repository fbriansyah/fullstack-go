package shared

import (
	"context"
	"testing"

	"go-templ-template/internal/shared/events"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockModule is a mock implementation of the Module interface for testing
type MockModule struct {
	mock.Mock
	name string
}

func NewMockModule(name string) *MockModule {
	return &MockModule{name: name}
}

func (m *MockModule) Name() string {
	return m.name
}

func (m *MockModule) Initialize(ctx context.Context, container *ModuleContainer) error {
	args := m.Called(ctx, container)
	return args.Error(0)
}

func (m *MockModule) RegisterRoutes(router *echo.Group) {
	m.Called(router)
}

func (m *MockModule) RegisterEventHandlers(eventBus events.EventBus) error {
	args := m.Called(eventBus)
	return args.Error(0)
}

func (m *MockModule) Health(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockModule) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockEventBus is a mock implementation of the EventBus interface for testing
type MockEventBus struct {
	mock.Mock
}

func (m *MockEventBus) Publish(ctx context.Context, event events.DomainEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventBus) Subscribe(eventType string, handler events.EventHandler) error {
	args := m.Called(eventType, handler)
	return args.Error(0)
}

func (m *MockEventBus) Unsubscribe(eventType string, handler events.EventHandler) error {
	args := m.Called(eventType, handler)
	return args.Error(0)
}

func (m *MockEventBus) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockEventBus) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockEventBus) Health() error {
	args := m.Called()
	return args.Error(0)
}

func TestModuleContainer_RegisterModule(t *testing.T) {
	// Arrange
	eventBus := &MockEventBus{}
	container := NewModuleContainer(eventBus, nil, nil)
	module := NewMockModule("test-module")

	// Act
	err := container.RegisterModule(module)

	// Assert
	assert.NoError(t, err)

	retrievedModule, exists := container.GetModule("test-module")
	assert.True(t, exists)
	assert.Equal(t, module, retrievedModule)
}

func TestModuleContainer_RegisterModule_DuplicateName(t *testing.T) {
	// Arrange
	eventBus := &MockEventBus{}
	container := NewModuleContainer(eventBus, nil, nil)
	module1 := NewMockModule("test-module")
	module2 := NewMockModule("test-module")

	// Act
	err1 := container.RegisterModule(module1)
	err2 := container.RegisterModule(module2)

	// Assert
	assert.NoError(t, err1)
	assert.Error(t, err2)
	assert.Contains(t, err2.Error(), "module already registered")
}

func TestModuleContainer_GetModule_NotFound(t *testing.T) {
	// Arrange
	eventBus := &MockEventBus{}
	container := NewModuleContainer(eventBus, nil, nil)

	// Act
	module, exists := container.GetModule("non-existent")

	// Assert
	assert.False(t, exists)
	assert.Nil(t, module)
}

func TestModuleContainer_GetAllModules(t *testing.T) {
	// Arrange
	eventBus := &MockEventBus{}
	container := NewModuleContainer(eventBus, nil, nil)
	module1 := NewMockModule("module1")
	module2 := NewMockModule("module2")

	container.RegisterModule(module1)
	container.RegisterModule(module2)

	// Act
	allModules := container.GetAllModules()

	// Assert
	assert.Len(t, allModules, 2)
	assert.Contains(t, allModules, "module1")
	assert.Contains(t, allModules, "module2")
}

func TestModuleError_Error(t *testing.T) {
	// Test without cause
	err1 := NewModuleError("test-module", "test message")
	assert.Equal(t, "test-module: test message", err1.Error())

	// Test with cause
	cause := assert.AnError
	err2 := NewModuleErrorWithCause("test-module", "test message", cause)
	assert.Equal(t, "test-module: test message: "+cause.Error(), err2.Error())
}

func TestModuleError_Unwrap(t *testing.T) {
	// Test without cause
	err1 := NewModuleError("test-module", "test message")
	assert.Nil(t, err1.Unwrap())

	// Test with cause
	cause := assert.AnError
	err2 := NewModuleErrorWithCause("test-module", "test message", cause)
	assert.Equal(t, cause, err2.Unwrap())
}
