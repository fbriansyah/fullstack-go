package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-templ-template/internal/config"
	"go-templ-template/internal/modules/auth"
	"go-templ-template/internal/modules/user"
	"go-templ-template/internal/shared"
	"go-templ-template/internal/shared/database"
	"go-templ-template/internal/shared/errors"
	"go-templ-template/internal/shared/events"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Create application
	app, err := NewApp(cfg)
	if err != nil {
		log.Fatal("Failed to create application:", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start application
	if err := app.Start(ctx); err != nil {
		log.Fatal("Failed to start application:", err)
	}

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("Server started on %s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Println("Press Ctrl+C to shutdown")

	<-quit
	log.Println("Shutting down server...")

	// Cancel context to signal shutdown
	cancel()

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Graceful shutdown
	if err := app.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during shutdown: %v", err)
		os.Exit(1)
	}

	log.Println("Server shutdown complete")
}

// App represents the main application with all its dependencies
type App struct {
	config         *config.Config
	router         *echo.Echo
	server         *http.Server
	dbManager      *database.Manager
	eventBus       events.EventBus
	moduleRegistry *shared.ModuleRegistry
	errorHandlers  *errors.ErrorHandlers
}

// NewApp creates a new application instance with all dependencies
func NewApp(cfg *config.Config) (*App, error) {
	// Create Echo router
	router := echo.New()

	// Configure Echo
	router.HideBanner = true
	router.HidePort = true

	// Create error handlers
	errorHandlersConfig := errors.ErrorHandlersConfig{
		ShowDetailedErrors: cfg.Server.Env != "production",
		LogErrors:          true,
	}
	errorHandlers := errors.NewErrorHandlers(errorHandlersConfig)

	// Set custom error handler
	router.HTTPErrorHandler = errorHandlers.CustomHTTPErrorHandler

	// Add middleware
	router.Use(middleware.Logger())
	router.Use(middleware.Recover())
	router.Use(middleware.CORS())
	router.Use(errorHandlers.ErrorPageMiddleware())

	// Create database manager
	migrationsPath := "migrations"
	// If running from cmd/server directory, adjust path
	if _, err := os.Stat("migrations"); os.IsNotExist(err) {
		migrationsPath = "../../migrations"
	}

	dbManager, err := database.NewManager(&cfg.Database, migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create database manager: %w", err)
	}

	// Create event bus
	eventBusConfig := events.RabbitMQConfig{
		URL:          cfg.RabbitMQ.URL,
		Exchange:     cfg.RabbitMQ.Exchange,
		ExchangeType: "topic",
		QueuePrefix:  cfg.RabbitMQ.QueuePrefix,
		Durable:      cfg.RabbitMQ.Durable,
		AutoDelete:   false,
		Exclusive:    false,
		NoWait:       false,
	}

	// Use default URL if not provided
	if eventBusConfig.URL == "" {
		eventBusConfig.URL = fmt.Sprintf("amqp://%s:%s@%s:%s/",
			cfg.RabbitMQ.User, cfg.RabbitMQ.Password, cfg.RabbitMQ.Host, cfg.RabbitMQ.Port)
	}

	eventBus := events.NewRabbitMQEventBus(eventBusConfig)

	// Create module registry
	moduleRegistry := shared.NewModuleRegistry(eventBus, dbManager.DB, cfg, router)

	// Register modules
	userModule := user.NewUserModule()
	if err := moduleRegistry.Register(userModule); err != nil {
		return nil, fmt.Errorf("failed to register user module: %w", err)
	}

	authModule := auth.NewAuthModule()
	if err := moduleRegistry.Register(authModule); err != nil {
		return nil, fmt.Errorf("failed to register auth module: %w", err)
	}

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &App{
		config:         cfg,
		router:         router,
		server:         server,
		dbManager:      dbManager,
		eventBus:       eventBus,
		moduleRegistry: moduleRegistry,
		errorHandlers:  errorHandlers,
	}, nil
}

// Start initializes and starts the application
func (a *App) Start(ctx context.Context) error {
	log.Println("Starting application...")

	// Initialize database
	runMigrations := a.config.Server.Env != "production" // Run migrations in dev/test
	if err := a.dbManager.Initialize(ctx, runMigrations); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Start event bus
	if err := a.eventBus.Start(ctx); err != nil {
		return fmt.Errorf("failed to start event bus: %w", err)
	}

	// Start modules (initialize, register routes, register event handlers)
	if err := a.moduleRegistry.Start(ctx); err != nil {
		return fmt.Errorf("failed to start modules: %w", err)
	}

	// Register health check endpoints
	a.registerHealthEndpoints()

	// Register error page routes
	a.errorHandlers.RegisterRoutes(a.router)

	// Start HTTP server in a goroutine
	go func() {
		log.Printf("HTTP server listening on %s", a.server.Addr)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	log.Println("Application started successfully")
	return nil
}

// Shutdown gracefully shuts down the application
func (a *App) Shutdown(ctx context.Context) error {
	log.Println("Shutting down application...")

	var lastErr error

	// Shutdown HTTP server
	log.Println("Shutting down HTTP server...")
	if err := a.server.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down HTTP server: %v", err)
		lastErr = err
	}

	// Shutdown modules
	log.Println("Shutting down modules...")
	if err := a.moduleRegistry.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down modules: %v", err)
		lastErr = err
	}

	// Stop event bus
	log.Println("Stopping event bus...")
	if err := a.eventBus.Stop(ctx); err != nil {
		log.Printf("Error stopping event bus: %v", err)
		lastErr = err
	}

	// Close database
	log.Println("Closing database...")
	if err := a.dbManager.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
		lastErr = err
	}

	if lastErr == nil {
		log.Println("Application shutdown completed successfully")
	} else {
		log.Printf("Application shutdown completed with errors: %v", lastErr)
	}

	return lastErr
}

// registerHealthEndpoints registers health check and monitoring endpoints
func (a *App) registerHealthEndpoints() {
	// Health check endpoint
	a.router.GET("/health", a.healthHandler)

	// Detailed health check endpoint
	a.router.GET("/health/detailed", a.detailedHealthHandler)

	// Readiness probe endpoint
	a.router.GET("/ready", a.readinessHandler)

	// Liveness probe endpoint
	a.router.GET("/live", a.livenessHandler)

	log.Println("Health check endpoints registered:")
	log.Println("  GET /health - Basic health check")
	log.Println("  GET /health/detailed - Detailed health status")
	log.Println("  GET /ready - Readiness probe")
	log.Println("  GET /live - Liveness probe")
}

// healthHandler provides a basic health check endpoint
func (a *App) healthHandler(c echo.Context) error {
	ctx := c.Request().Context()

	// Check database health
	dbHealth := a.dbManager.GetHealthStatus(ctx)

	// Check event bus health
	eventBusHealthy := a.eventBus.Health() == nil

	// Determine overall health
	healthy := dbHealth.Status == "healthy" && eventBusHealthy

	status := "healthy"
	if !healthy {
		status = "unhealthy"
	}

	response := map[string]interface{}{
		"status":    status,
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0", // You might want to get this from build info
	}

	if healthy {
		return c.JSON(http.StatusOK, response)
	} else {
		return c.JSON(http.StatusServiceUnavailable, response)
	}
}

// detailedHealthHandler provides detailed health information
func (a *App) detailedHealthHandler(c echo.Context) error {
	ctx := c.Request().Context()

	// Check database health
	dbHealth := a.dbManager.GetHealthStatus(ctx)

	// Check event bus health
	eventBusErr := a.eventBus.Health()
	eventBusHealth := map[string]interface{}{
		"status": "healthy",
	}
	if eventBusErr != nil {
		eventBusHealth["status"] = "unhealthy"
		eventBusHealth["error"] = eventBusErr.Error()
	}

	// Check module health
	moduleHealth := a.moduleRegistry.Health(ctx)

	// Determine overall health
	overallHealthy := dbHealth.Status == "healthy" && eventBusErr == nil
	for _, err := range moduleHealth {
		if err != nil {
			overallHealthy = false
			break
		}
	}

	status := "healthy"
	if !overallHealthy {
		status = "unhealthy"
	}

	// Convert module health to response format
	moduleHealthResponse := make(map[string]interface{})
	for moduleName, err := range moduleHealth {
		if err != nil {
			moduleHealthResponse[moduleName] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
		} else {
			moduleHealthResponse[moduleName] = map[string]interface{}{
				"status": "healthy",
			}
		}
	}

	response := map[string]interface{}{
		"status":    status,
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
		"components": map[string]interface{}{
			"database": dbHealth,
			"eventbus": eventBusHealth,
			"modules":  moduleHealthResponse,
		},
	}

	if overallHealthy {
		return c.JSON(http.StatusOK, response)
	} else {
		return c.JSON(http.StatusServiceUnavailable, response)
	}
}

// readinessHandler indicates if the application is ready to serve traffic
func (a *App) readinessHandler(c echo.Context) error {
	ctx := c.Request().Context()

	// Check if all critical components are ready
	dbHealth := a.dbManager.GetHealthStatus(ctx)
	eventBusHealthy := a.eventBus.Health() == nil

	ready := dbHealth.Status == "healthy" && eventBusHealthy

	response := map[string]interface{}{
		"ready":     ready,
		"timestamp": time.Now().UTC(),
	}

	if ready {
		return c.JSON(http.StatusOK, response)
	} else {
		return c.JSON(http.StatusServiceUnavailable, response)
	}
}

// livenessHandler indicates if the application is alive
func (a *App) livenessHandler(c echo.Context) error {
	// Simple liveness check - if we can respond, we're alive
	response := map[string]interface{}{
		"alive":     true,
		"timestamp": time.Now().UTC(),
	}

	return c.JSON(http.StatusOK, response)
}
