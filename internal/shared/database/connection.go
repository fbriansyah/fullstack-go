package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go-templ-template/internal/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// DB wraps sqlx.DB with additional functionality
type DB struct {
	*sqlx.DB
	config *config.DatabaseConfig
}

// ConnectionOptions holds database connection configuration
type ConnectionOptions struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultConnectionOptions returns sensible defaults for database connections
func DefaultConnectionOptions() ConnectionOptions {
	return ConnectionOptions{
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}
}

// NewConnection creates a new database connection with connection pooling
func NewConnection(cfg *config.DatabaseConfig, opts ConnectionOptions) (*DB, error) {
	return NewConnectionWithTimeout(cfg, opts, 30*time.Second)
}

// NewConnectionWithTimeout creates a new database connection with connection pooling and custom timeout
func NewConnectionWithTimeout(cfg *config.DatabaseConfig, opts ConnectionOptions, timeout time.Duration) (*DB, error) {
	var dsn string

	// Use DATABASE_URL if provided, otherwise build from individual components
	if cfg.URL != "" {
		dsn = cfg.URL
	} else {
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode, int(timeout.Seconds()),
		)
	}

	// Create context with timeout for connection attempt
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Open database connection with timeout
	sqlxDB, err := sqlx.ConnectContext(ctx, "postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database within %v: %w", timeout, err)
	}

	// Configure connection pool
	sqlxDB.SetMaxOpenConns(opts.MaxOpenConns)
	sqlxDB.SetMaxIdleConns(opts.MaxIdleConns)
	sqlxDB.SetConnMaxLifetime(opts.ConnMaxLifetime)
	sqlxDB.SetConnMaxIdleTime(opts.ConnMaxIdleTime)

	db := &DB{
		DB:     sqlxDB,
		config: cfg,
	}

	return db, nil
}

// HealthCheck performs a health check on the database connection
func (db *DB) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Test the connection with a simple query
	var result int
	err := db.GetContext(ctx, &result, "SELECT 1")
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// Stats returns database connection pool statistics
func (db *DB) Stats() sql.DBStats {
	return db.DB.Stats()
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// GetDSN returns the data source name for the database
func (db *DB) GetDSN() string {
	if db.config.URL != "" {
		return db.config.URL
	}

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		db.config.Host, db.config.Port, db.config.User,
		db.config.Password, db.config.Name, db.config.SSLMode,
	)
}
