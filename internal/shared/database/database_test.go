package database

import (
	"context"
	"os"
	"testing"
	"time"

	"go-templ-template/internal/config"
)

// TestDatabaseConnection tests the database connection functionality
func TestDatabaseConnection(t *testing.T) {
	// Skip if no test database is configured
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping database tests")
	}

	cfg := &config.DatabaseConfig{
		URL: os.Getenv("TEST_DATABASE_URL"),
	}

	// Test connection creation
	db, err := NewConnection(cfg, DefaultConnectionOptions())
	if err != nil {
		t.Fatalf("Failed to create database connection: %v", err)
	}
	defer db.Close()

	// Test health check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.HealthCheck(ctx)
	if err != nil {
		t.Fatalf("Database health check failed: %v", err)
	}

	// Test connection stats
	stats := db.Stats()
	if stats.OpenConnections == 0 {
		t.Error("Expected at least one open connection")
	}
}

// TestHealthChecker tests the health checking functionality
func TestHealthChecker(t *testing.T) {
	// Skip if no test database is configured
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping database tests")
	}

	cfg := &config.DatabaseConfig{
		URL: os.Getenv("TEST_DATABASE_URL"),
	}

	db, err := NewConnection(cfg, DefaultConnectionOptions())
	if err != nil {
		t.Fatalf("Failed to create database connection: %v", err)
	}
	defer db.Close()

	healthChecker := NewHealthChecker(db)

	// Test quick check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = healthChecker.QuickCheck(ctx)
	if err != nil {
		t.Fatalf("Quick health check failed: %v", err)
	}

	// Test comprehensive check
	status := healthChecker.Check(ctx)
	if status.Status != "healthy" {
		t.Fatalf("Expected healthy status, got: %s - %s", status.Status, status.Message)
	}

	if status.Latency == 0 {
		t.Error("Expected non-zero latency")
	}

	// Test connection validation
	err = healthChecker.ValidateConnection(ctx)
	if err != nil {
		t.Fatalf("Connection validation failed: %v", err)
	}
}

// TestConnectionOptions tests different connection configurations
func TestConnectionOptions(t *testing.T) {
	opts := DefaultConnectionOptions()

	if opts.MaxOpenConns <= 0 {
		t.Error("MaxOpenConns should be positive")
	}

	if opts.MaxIdleConns <= 0 {
		t.Error("MaxIdleConns should be positive")
	}

	if opts.ConnMaxLifetime <= 0 {
		t.Error("ConnMaxLifetime should be positive")
	}

	if opts.ConnMaxIdleTime <= 0 {
		t.Error("ConnMaxIdleTime should be positive")
	}
}

// BenchmarkHealthCheck benchmarks the health check performance
func BenchmarkHealthCheck(b *testing.B) {
	// Skip if no test database is configured
	if os.Getenv("TEST_DATABASE_URL") == "" {
		b.Skip("TEST_DATABASE_URL not set, skipping database benchmarks")
	}

	cfg := &config.DatabaseConfig{
		URL: os.Getenv("TEST_DATABASE_URL"),
	}

	db, err := NewConnection(cfg, DefaultConnectionOptions())
	if err != nil {
		b.Fatalf("Failed to create database connection: %v", err)
	}
	defer db.Close()

	healthChecker := NewHealthChecker(db)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := healthChecker.QuickCheck(ctx)
		if err != nil {
			b.Fatalf("Health check failed: %v", err)
		}
	}
}
