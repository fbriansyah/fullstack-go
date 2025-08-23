package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// HealthStatus represents the health status of the database
type HealthStatus struct {
	Status      string        `json:"status"`
	Message     string        `json:"message,omitempty"`
	Latency     time.Duration `json:"latency"`
	Connections DBStats       `json:"connections"`
	Timestamp   time.Time     `json:"timestamp"`
}

// DBStats represents database connection statistics
type DBStats struct {
	OpenConnections   int `json:"open_connections"`
	InUseConnections  int `json:"in_use_connections"`
	IdleConnections   int `json:"idle_connections"`
	WaitCount         int `json:"wait_count"`
	WaitDuration      int `json:"wait_duration_ms"`
	MaxIdleClosed     int `json:"max_idle_closed"`
	MaxIdleTimeClosed int `json:"max_idle_time_closed"`
	MaxLifetimeClosed int `json:"max_lifetime_closed"`
}

// HealthChecker provides database health checking functionality
type HealthChecker struct {
	db *DB
}

// NewHealthChecker creates a new health checker for the database
func NewHealthChecker(db *DB) *HealthChecker {
	return &HealthChecker{db: db}
}

// Check performs a comprehensive health check on the database
func (hc *HealthChecker) Check(ctx context.Context) *HealthStatus {
	start := time.Now()
	status := &HealthStatus{
		Timestamp: start,
	}

	// Perform health check with timeout
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := hc.db.HealthCheck(checkCtx)
	latency := time.Since(start)
	status.Latency = latency

	if err != nil {
		status.Status = "unhealthy"
		status.Message = err.Error()
	} else {
		status.Status = "healthy"
		status.Message = "Database is responding normally"
	}

	// Get connection statistics
	stats := hc.db.Stats()
	status.Connections = DBStats{
		OpenConnections:   stats.OpenConnections,
		InUseConnections:  stats.InUse,
		IdleConnections:   stats.Idle,
		WaitCount:         int(stats.WaitCount),
		WaitDuration:      int(stats.WaitDuration.Milliseconds()),
		MaxIdleClosed:     int(stats.MaxIdleClosed),
		MaxIdleTimeClosed: int(stats.MaxIdleTimeClosed),
		MaxLifetimeClosed: int(stats.MaxLifetimeClosed),
	}

	return status
}

// QuickCheck performs a simple ping to check if database is reachable
func (hc *HealthChecker) QuickCheck(ctx context.Context) error {
	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return hc.db.HealthCheck(checkCtx)
}

// ValidateConnection validates the database connection and configuration
func (hc *HealthChecker) ValidateConnection(ctx context.Context) error {
	// Check basic connectivity
	if err := hc.QuickCheck(ctx); err != nil {
		return fmt.Errorf("basic connectivity check failed: %w", err)
	}

	// Check if we can perform basic operations
	checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Test transaction capability
	tx, err := hc.db.BeginTxx(checkCtx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Test query execution within transaction
	var result int
	err = tx.GetContext(checkCtx, &result, "SELECT 1")
	if err != nil {
		return fmt.Errorf("failed to execute query in transaction: %w", err)
	}

	// Test rollback
	err = tx.Rollback()
	if err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	return nil
}

// WaitForConnection waits for the database to become available
func (hc *HealthChecker) WaitForConnection(ctx context.Context, maxRetries int, retryInterval time.Duration) error {
	for i := 0; i < maxRetries; i++ {
		if err := hc.QuickCheck(ctx); err == nil {
			return nil
		}

		if i < maxRetries-1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryInterval):
				continue
			}
		}
	}

	return fmt.Errorf("database did not become available after %d retries", maxRetries)
}
