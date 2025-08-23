package database

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"go-templ-template/internal/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// MigrationRunner handles database migrations
type MigrationRunner struct {
	migrate *migrate.Migrate
	db      *sql.DB
}

// NewMigrationRunner creates a new migration runner
func NewMigrationRunner(cfg *config.DatabaseConfig, migrationsPath string) (*MigrationRunner, error) {
	var dsn string

	// Use DATABASE_URL if provided, otherwise build from individual components
	if cfg.URL != "" {
		dsn = cfg.URL
	} else {
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
		)
	}

	// Open database connection for migrations
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database for migrations: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create postgres driver instance
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Get absolute path for migrations
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to get absolute path for migrations: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", absPath),
		"postgres",
		driver,
	)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return &MigrationRunner{
		migrate: m,
		db:      db,
	}, nil
}

// Up runs all available migrations
func (mr *MigrationRunner) Up() error {
	err := mr.migrate.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations up: %w", err)
	}
	return nil
}

// Down rolls back all migrations
func (mr *MigrationRunner) Down() error {
	err := mr.migrate.Down()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations down: %w", err)
	}
	return nil
}

// Steps runs n migration steps (positive for up, negative for down)
func (mr *MigrationRunner) Steps(n int) error {
	err := mr.migrate.Steps(n)
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migration steps: %w", err)
	}
	return nil
}

// Version returns the current migration version
func (mr *MigrationRunner) Version() (uint, bool, error) {
	version, dirty, err := mr.migrate.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}
	return version, dirty, nil
}

// Force sets the migration version without running migrations
func (mr *MigrationRunner) Force(version int) error {
	err := mr.migrate.Force(version)
	if err != nil {
		return fmt.Errorf("failed to force migration version: %w", err)
	}
	return nil
}

// Close closes the migration runner and database connection
func (mr *MigrationRunner) Close() error {
	sourceErr, dbErr := mr.migrate.Close()
	if sourceErr != nil {
		return fmt.Errorf("failed to close migration source: %w", sourceErr)
	}
	if dbErr != nil {
		return fmt.Errorf("failed to close migration database: %w", dbErr)
	}
	return mr.db.Close()
}

// CreateMigration creates a new migration file pair (up and down)
func CreateMigration(migrationsPath, name string) error {
	// This is a utility function that could be used in CLI tools
	// For now, we'll just document the expected file naming convention
	// Migration files should be named: {version}_{name}.up.sql and {version}_{name}.down.sql
	// Example: 001_create_users_table.up.sql, 001_create_users_table.down.sql
	return fmt.Errorf("migration creation should be done manually or via migrate CLI tool")
}
