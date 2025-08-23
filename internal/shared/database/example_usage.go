package database

import (
	"context"
	"log"
	"time"

	"go-templ-template/internal/config"
)

// ExampleUsage demonstrates how to use the database package
func ExampleUsage() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create database manager
	manager, err := NewManager(&cfg.Database, "./migrations")
	if err != nil {
		log.Fatalf("Failed to create database manager: %v", err)
	}
	defer manager.Close()

	// Initialize database with migrations
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = manager.Initialize(ctx, true) // true = run migrations
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Use the database connection
	db := manager.DB

	// Example: Simple query
	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM users")
	if err != nil {
		log.Printf("Failed to count users: %v", err)
	} else {
		log.Printf("User count: %d", count)
	}

	// Example: Health check
	status := manager.GetHealthStatus(ctx)
	log.Printf("Database status: %s (latency: %v)", status.Status, status.Latency)

	// Example: Transaction
	tx, err := db.Beginx()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return
	}
	defer tx.Rollback() // Will be ignored if tx.Commit() is called

	// Do some work in transaction...
	// tx.Exec("INSERT INTO ...")

	err = tx.Commit()
	if err != nil {
		log.Printf("Failed to commit transaction: %v", err)
	}
}

// ExampleMigrationUsage demonstrates migration operations
func ExampleMigrationUsage() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create migration runner
	runner, err := NewMigrationRunner(&cfg.Database, "./migrations")
	if err != nil {
		log.Fatalf("Failed to create migration runner: %v", err)
	}
	defer runner.Close()

	// Check current version
	version, dirty, err := runner.Version()
	if err != nil {
		log.Printf("No migrations applied yet")
	} else {
		log.Printf("Current migration version: %d (dirty: %t)", version, dirty)
	}

	// Run migrations up
	err = runner.Up()
	if err != nil {
		log.Printf("Failed to run migrations: %v", err)
	} else {
		log.Println("Migrations completed successfully")
	}

	// Example: Run specific number of steps
	// err = runner.Steps(1) // Run 1 migration up
	// err = runner.Steps(-1) // Run 1 migration down
}

// ExampleHealthMonitoring demonstrates health monitoring
func ExampleHealthMonitoring() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := NewConnection(&cfg.Database, DefaultConnectionOptions())
	if err != nil {
		log.Fatalf("Failed to create database connection: %v", err)
	}
	defer db.Close()

	healthChecker := NewHealthChecker(db)

	// Continuous health monitoring
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			status := healthChecker.Check(ctx)
			cancel()

			if status.Status == "healthy" {
				log.Printf("Database healthy - latency: %v, connections: %d/%d",
					status.Latency,
					status.Connections.InUseConnections,
					status.Connections.OpenConnections)
			} else {
				log.Printf("Database unhealthy: %s", status.Message)
			}
		}
	}
}
