package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-templ-template/internal/config"
)

// Manager handles database connections, migrations, and health checks
type Manager struct {
	DB              *DB
	HealthChecker   *HealthChecker
	MigrationRunner *MigrationRunner
	config          *config.DatabaseConfig
}

// NewManager creates a new database manager with all components
func NewManager(cfg *config.DatabaseConfig, migrationsPath string) (*Manager, error) {
	// Create database connection
	db, err := NewConnection(cfg, DefaultConnectionOptions())
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	// Create health checker
	healthChecker := NewHealthChecker(db)

	// Validate connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := healthChecker.ValidateConnection(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("database connection validation failed: %w", err)
	}

	// Create migration runner
	migrationRunner, err := NewMigrationRunner(cfg, migrationsPath)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create migration runner: %w", err)
	}

	return &Manager{
		DB:              db,
		HealthChecker:   healthChecker,
		MigrationRunner: migrationRunner,
		config:          cfg,
	}, nil
}

// Initialize sets up the database with migrations and validation
func (m *Manager) Initialize(ctx context.Context, runMigrations bool) error {
	log.Println("Initializing database...")

	// Wait for database to be available
	if err := m.HealthChecker.WaitForConnection(ctx, 30, 2*time.Second); err != nil {
		return fmt.Errorf("database not available: %w", err)
	}

	// Run migrations if requested
	if runMigrations {
		log.Println("Running database migrations...")
		if err := m.MigrationRunner.Up(); err != nil {
			return fmt.Errorf("failed to run migrations: %w", err)
		}

		// Log current migration version
		version, dirty, err := m.MigrationRunner.Version()
		if err != nil {
			log.Printf("Warning: could not get migration version: %v", err)
		} else {
			log.Printf("Database migration version: %d (dirty: %t)", version, dirty)
		}
	}

	// Final health check
	status := m.HealthChecker.Check(ctx)
	if status.Status != "healthy" {
		return fmt.Errorf("database health check failed: %s", status.Message)
	}

	log.Printf("Database initialized successfully (latency: %v)", status.Latency)
	return nil
}

// Close gracefully closes all database connections
func (m *Manager) Close() error {
	log.Println("Closing database connections...")

	var lastErr error

	if m.MigrationRunner != nil {
		if err := m.MigrationRunner.Close(); err != nil {
			log.Printf("Error closing migration runner: %v", err)
			lastErr = err
		}
	}

	if m.DB != nil {
		if err := m.DB.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
			lastErr = err
		}
	}

	return lastErr
}

// GetHealthStatus returns the current health status of the database
func (m *Manager) GetHealthStatus(ctx context.Context) *HealthStatus {
	return m.HealthChecker.Check(ctx)
}
