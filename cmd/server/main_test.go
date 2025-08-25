package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"go-templ-template/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewApp(t *testing.T) {
	// Skip if no test database available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create test configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: "8080",
			Env:  "test",
		},
		Database: config.DatabaseConfig{
			URL:      "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "postgres",
			Name:     "postgres",
			SSLMode:  "disable",
		},
		RabbitMQ: config.RabbitMQConfig{
			URL:         "amqp://guest:guest@localhost:5672/",
			Host:        "localhost",
			Port:        "5672",
			User:        "guest",
			Password:    "guest",
			Exchange:    "test_events",
			QueuePrefix: "test_app",
			Durable:     false, // Use non-durable for tests
		},
	}

	// Create application
	app, err := NewApp(cfg)
	require.NoError(t, err)
	require.NotNil(t, app)

	// Verify app components are initialized
	assert.NotNil(t, app.config)
	assert.NotNil(t, app.router)
	assert.NotNil(t, app.server)
	assert.NotNil(t, app.dbManager)
	assert.NotNil(t, app.eventBus)
	assert.NotNil(t, app.moduleRegistry)

	// Verify configuration
	assert.Equal(t, cfg, app.config)
	assert.Equal(t, "localhost:8080", app.server.Addr)
}

func TestApp_HealthEndpoints(t *testing.T) {
	// Skip if no test database available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create test configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: "8081", // Use different port for test
			Env:  "test",
		},
		Database: config.DatabaseConfig{
			URL:      "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "postgres",
			Name:     "postgres",
			SSLMode:  "disable",
		},
		RabbitMQ: config.RabbitMQConfig{
			URL:         "amqp://guest:guest@localhost:5672/",
			Host:        "localhost",
			Port:        "5672",
			User:        "guest",
			Password:    "guest",
			Exchange:    "test_events",
			QueuePrefix: "test_app",
			Durable:     false,
		},
	}

	// Create and start application
	app, err := NewApp(cfg)
	require.NoError(t, err)

	// Start application (this will fail if dependencies aren't available)
	// We'll test the health endpoints registration even if start fails
	app.registerHealthEndpoints()

	// Test that health endpoints are registered
	// We can't easily test the actual HTTP responses without starting the server
	// and having the dependencies available, but we can verify the routes exist
	routes := app.router.Routes()

	healthRoutes := []string{"/health", "/health/detailed", "/ready", "/live"}
	foundRoutes := make(map[string]bool)

	for _, route := range routes {
		for _, expectedRoute := range healthRoutes {
			if route.Path == expectedRoute && route.Method == "GET" {
				foundRoutes[expectedRoute] = true
			}
		}
	}

	// Verify all health endpoints are registered
	for _, expectedRoute := range healthRoutes {
		assert.True(t, foundRoutes[expectedRoute], "Health endpoint %s should be registered", expectedRoute)
	}
}

func TestApp_GracefulShutdown(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: "8082", // Use different port for test
			Env:  "test",
		},
		Database: config.DatabaseConfig{
			URL:      "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "postgres",
			Name:     "postgres",
			SSLMode:  "disable",
		},
		RabbitMQ: config.RabbitMQConfig{
			URL:         "amqp://guest:guest@localhost:5672/",
			Host:        "localhost",
			Port:        "5672",
			User:        "guest",
			Password:    "guest",
			Exchange:    "test_events",
			QueuePrefix: "test_app",
			Durable:     false,
		},
	}

	// Create application
	app, err := NewApp(cfg)
	if err != nil {
		// If we can't create the app due to missing dependencies, skip the test
		t.Skipf("Skipping test due to missing dependencies: %v", err)
	}

	// Test shutdown without starting (should not panic)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = app.Shutdown(ctx)
	// Shutdown might return errors if components weren't started, but shouldn't panic
	// We're mainly testing that the shutdown process completes
	assert.NotPanics(t, func() {
		app.Shutdown(ctx)
	})
}

func TestApp_ConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		expectError bool
	}{
		{
			name: "valid config",
			config: &config.Config{
				Server: config.ServerConfig{
					Host: "localhost",
					Port: "8080",
					Env:  "test",
				},
				Database: config.DatabaseConfig{
					URL:      "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
					Host:     "localhost",
					Port:     "5432",
					User:     "postgres",
					Password: "postgres",
					Name:     "postgres",
					SSLMode:  "disable",
				},
				RabbitMQ: config.RabbitMQConfig{
					URL:         "amqp://guest:guest@localhost:5672/",
					Host:        "localhost",
					Port:        "5672",
					User:        "guest",
					Password:    "guest",
					Exchange:    "test_events",
					QueuePrefix: "test_app",
					Durable:     false,
				},
			},
			expectError: false,
		},
		{
			name: "invalid database config",
			config: &config.Config{
				Server: config.ServerConfig{
					Host: "localhost",
					Port: "8080",
					Env:  "test",
				},
				Database: config.DatabaseConfig{
					URL:      "invalid-url",
					Host:     "",
					Port:     "",
					User:     "",
					Password: "",
					Name:     "",
					SSLMode:  "",
				},
				RabbitMQ: config.RabbitMQConfig{
					URL:         "amqp://guest:guest@localhost:5672/",
					Host:        "localhost",
					Port:        "5672",
					User:        "guest",
					Password:    "guest",
					Exchange:    "test_events",
					QueuePrefix: "test_app",
					Durable:     false,
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, err := NewApp(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, app)
			} else {
				if err != nil {
					// If we get an error due to missing dependencies (like DB not running),
					// skip the test instead of failing
					t.Skipf("Skipping test due to missing dependencies: %v", err)
				}
				assert.NotNil(t, app)
			}
		})
	}
}

// TestHealthHandlerResponse tests the health handler response format
func TestHealthHandlerResponse(t *testing.T) {
	// Create test configuration
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: "8083",
			Env:  "test",
		},
		Database: config.DatabaseConfig{
			URL:      "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "postgres",
			Name:     "postgres",
			SSLMode:  "disable",
		},
		RabbitMQ: config.RabbitMQConfig{
			URL:         "amqp://guest:guest@localhost:5672/",
			Host:        "localhost",
			Port:        "5672",
			User:        "guest",
			Password:    "guest",
			Exchange:    "test_events",
			QueuePrefix: "test_app",
			Durable:     false,
		},
	}

	// Create application
	app, err := NewApp(cfg)
	if err != nil {
		// If we can't create the app due to missing dependencies, skip the test
		t.Skipf("Skipping test due to missing dependencies: %v", err)
	}

	// Register health endpoints
	app.registerHealthEndpoints()

	// Verify that the health endpoints are properly registered
	// This tests the registration logic without requiring actual dependencies
	routes := app.router.Routes()

	expectedEndpoints := map[string]bool{
		"/health":          false,
		"/health/detailed": false,
		"/ready":           false,
		"/live":            false,
	}

	for _, route := range routes {
		if route.Method == http.MethodGet {
			if _, exists := expectedEndpoints[route.Path]; exists {
				expectedEndpoints[route.Path] = true
			}
		}
	}

	// Verify all endpoints are registered
	for endpoint, found := range expectedEndpoints {
		assert.True(t, found, "Endpoint %s should be registered", endpoint)
	}
}

// TestApp_HealthEndpointsUnit tests health endpoint registration without dependencies
func TestApp_HealthEndpointsUnit(t *testing.T) {
	// This test doesn't require actual database/rabbitmq connections
	// We'll create a minimal app just to test endpoint registration

	// Create test configuration with dummy values
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "localhost",
			Port: "8084",
			Env:  "test",
		},
		Database: config.DatabaseConfig{
			URL:      "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "postgres",
			Name:     "postgres",
			SSLMode:  "disable",
		},
		RabbitMQ: config.RabbitMQConfig{
			URL:         "amqp://guest:guest@localhost:5672/",
			Host:        "localhost",
			Port:        "5672",
			User:        "guest",
			Password:    "guest",
			Exchange:    "test_events",
			QueuePrefix: "test_app",
			Durable:     false,
		},
	}

	// Try to create the app, but if it fails due to missing dependencies,
	// we'll just test the route registration logic
	app, err := NewApp(cfg)
	if err != nil {
		t.Skipf("Skipping test due to missing dependencies: %v", err)
	}

	// Test health endpoint registration
	app.registerHealthEndpoints()

	// Verify routes are registered
	routes := app.router.Routes()
	healthEndpoints := []string{"/health", "/health/detailed", "/ready", "/live"}

	registeredEndpoints := make(map[string]bool)
	for _, route := range routes {
		if route.Method == "GET" {
			for _, endpoint := range healthEndpoints {
				if route.Path == endpoint {
					registeredEndpoints[endpoint] = true
				}
			}
		}
	}

	// Assert all health endpoints are registered
	for _, endpoint := range healthEndpoints {
		assert.True(t, registeredEndpoints[endpoint], "Health endpoint %s should be registered", endpoint)
	}
}
