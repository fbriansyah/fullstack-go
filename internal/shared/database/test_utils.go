package database

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"go-templ-template/internal/config"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// TestDatabaseConfig holds configuration for test database
type TestDatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// DefaultTestDatabaseConfig returns default test database configuration
func DefaultTestDatabaseConfig() *TestDatabaseConfig {
	return &TestDatabaseConfig{
		Host:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:     getEnvOrDefault("TEST_DB_PORT", "5432"),
		User:     getEnvOrDefault("TEST_DB_USER", "postgres"),
		Password: getEnvOrDefault("TEST_DB_PASSWORD", "postgres"),
		Name:     getEnvOrDefault("TEST_DB_NAME", "test_db"),
		SSLMode:  getEnvOrDefault("TEST_DB_SSLMODE", "disable"),
	}
}

// ToConfig converts TestDatabaseConfig to config.DatabaseConfig
func (cfg *TestDatabaseConfig) ToConfig() *config.DatabaseConfig {
	return &config.DatabaseConfig{
		Host:     cfg.Host,
		Port:     cfg.Port,
		User:     cfg.User,
		Password: cfg.Password,
		Name:     cfg.Name,
		SSLMode:  cfg.SSLMode,
	}
}

// TestDatabase represents a test database instance
type TestDatabase struct {
	DB     *DB
	Config *TestDatabaseConfig
	t      *testing.T
}

// NewTestDatabase creates a new test database instance
func NewTestDatabase(t *testing.T) *TestDatabase {
	cfg := DefaultTestDatabaseConfig()

	db, err := NewConnection(cfg.ToConfig(), DefaultConnectionOptions())
	require.NoError(t, err, "Failed to connect to test database")

	return &TestDatabase{
		DB:     db,
		Config: cfg,
		t:      t,
	}
}

// NewTestDatabaseWithConfig creates a new test database instance with custom config
func NewTestDatabaseWithConfig(t *testing.T, cfg *TestDatabaseConfig) *TestDatabase {
	db, err := NewConnection(cfg.ToConfig(), DefaultConnectionOptions())
	require.NoError(t, err, "Failed to connect to test database")

	return &TestDatabase{
		DB:     db,
		Config: cfg,
		t:      t,
	}
}

// CreateTable creates a table using the provided SQL
func (tdb *TestDatabase) CreateTable(sql string) {
	_, err := tdb.DB.Exec(sql)
	require.NoError(tdb.t, err, "Failed to create test table")
}

// DropTable drops a table
func (tdb *TestDatabase) DropTable(tableName string) {
	sql := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", tableName)
	_, err := tdb.DB.Exec(sql)
	require.NoError(tdb.t, err, "Failed to drop test table")
}

// TruncateTable truncates a table
func (tdb *TestDatabase) TruncateTable(tableName string) {
	sql := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", tableName)
	_, err := tdb.DB.Exec(sql)
	require.NoError(tdb.t, err, "Failed to truncate test table")
}

// ExecuteSQL executes arbitrary SQL
func (tdb *TestDatabase) ExecuteSQL(sql string, args ...interface{}) {
	_, err := tdb.DB.Exec(sql, args...)
	require.NoError(tdb.t, err, "Failed to execute SQL")
}

// SeedData inserts test data using the provided SQL
func (tdb *TestDatabase) SeedData(sql string, args ...interface{}) {
	_, err := tdb.DB.Exec(sql, args...)
	require.NoError(tdb.t, err, "Failed to seed test data")
}

// Close closes the test database connection
func (tdb *TestDatabase) Close() {
	if tdb.DB != nil {
		err := tdb.DB.Close()
		require.NoError(tdb.t, err, "Failed to close test database")
	}
}

// Cleanup performs cleanup operations (truncate tables, etc.)
func (tdb *TestDatabase) Cleanup(tableNames ...string) {
	for _, tableName := range tableNames {
		tdb.TruncateTable(tableName)
	}
}

// WithTransaction executes a function within a transaction for testing
func (tdb *TestDatabase) WithTransaction(fn func(ctx context.Context) error) error {
	ctx := context.Background()
	return ExecuteInTransaction(ctx, tdb.DB, fn)
}

// AssertTableExists asserts that a table exists
func (tdb *TestDatabase) AssertTableExists(tableName string) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)`

	err := tdb.DB.Get(&exists, query, tableName)
	require.NoError(tdb.t, err)
	require.True(tdb.t, exists, "Table %s should exist", tableName)
}

// AssertTableNotExists asserts that a table does not exist
func (tdb *TestDatabase) AssertTableNotExists(tableName string) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)`

	err := tdb.DB.Get(&exists, query, tableName)
	require.NoError(tdb.t, err)
	require.False(tdb.t, exists, "Table %s should not exist", tableName)
}

// AssertRowCount asserts the number of rows in a table
func (tdb *TestDatabase) AssertRowCount(tableName string, expectedCount int) {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)

	err := tdb.DB.Get(&count, query)
	require.NoError(tdb.t, err)
	require.Equal(tdb.t, expectedCount, count, "Row count mismatch for table %s", tableName)
}

// WaitForConnection waits for the database to be available
func (tdb *TestDatabase) WaitForConnection(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			require.Fail(tdb.t, "Timeout waiting for database connection")
		case <-ticker.C:
			if err := tdb.DB.HealthCheck(ctx); err == nil {
				return
			}
		}
	}
}

// TestSuite provides a complete test suite setup for database tests
type TestSuite struct {
	DB       *TestDatabase
	Manager  *Manager
	t        *testing.T
	tables   []string
	cleanups []func()
}

// NewTestSuite creates a new test suite
func NewTestSuite(t *testing.T) *TestSuite {
	testDB := NewTestDatabase(t)

	// Create a manager for testing
	manager, err := NewManager(testDB.Config.ToConfig(), "")
	require.NoError(t, err)

	return &TestSuite{
		DB:       testDB,
		Manager:  manager,
		t:        t,
		tables:   make([]string, 0),
		cleanups: make([]func(), 0),
	}
}

// AddTable registers a table for cleanup
func (ts *TestSuite) AddTable(tableName string) {
	ts.tables = append(ts.tables, tableName)
}

// AddCleanup registers a cleanup function
func (ts *TestSuite) AddCleanup(cleanup func()) {
	ts.cleanups = append(ts.cleanups, cleanup)
}

// Setup performs initial setup for the test suite
func (ts *TestSuite) Setup() {
	// Wait for database to be available
	ts.DB.WaitForConnection(30 * time.Second)
}

// Teardown performs cleanup for the test suite
func (ts *TestSuite) Teardown() {
	// Run custom cleanups
	for _, cleanup := range ts.cleanups {
		cleanup()
	}

	// Clean up tables
	ts.DB.Cleanup(ts.tables...)

	// Close connections
	if ts.Manager != nil {
		_ = ts.Manager.Close()
	}

	if ts.DB != nil {
		ts.DB.Close()
	}
}

// CreateTestEntitiesTable creates the test entities table for testing
func (ts *TestSuite) CreateTestEntitiesTable() {
	sql := `
		CREATE TABLE IF NOT EXISTS test_entities (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			version INTEGER NOT NULL DEFAULT 1
		)`

	ts.DB.CreateTable(sql)
	ts.AddTable("test_entities")
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SkipIfNoDatabase skips the test if no test database is available
func SkipIfNoDatabase(t *testing.T) {
	cfg := DefaultTestDatabaseConfig()

	db, err := NewConnection(cfg.ToConfig(), DefaultConnectionOptions())
	if err != nil {
		t.Skipf("Skipping database test: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.HealthCheck(ctx); err != nil {
		t.Skipf("Skipping database test: database not available: %v", err)
		return
	}

	_ = db.Close()
}

// NewTestDB creates a new test database connection (alias for NewTestDatabase)
func NewTestDB() (*TestDatabase, error) {
	cfg := DefaultTestDatabaseConfig()

	db, err := NewConnection(cfg.ToConfig(), DefaultConnectionOptions())
	if err != nil {
		return nil, err
	}

	return &TestDatabase{
		DB:     db,
		Config: cfg,
	}, nil
}

// CreateTables creates the basic tables needed for testing
func CreateTables(ctx context.Context, db *sqlx.DB) error {
	// This is a placeholder - in a real application, you would run your migrations here
	// For now, we'll create some basic tables that tests might need

	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			version INTEGER NOT NULL DEFAULT 1
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token_hash VARCHAR(255) NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
	}

	for _, query := range queries {
		if _, err := db.ExecContext(ctx, query); err != nil {
			return err
		}
	}

	return nil
}
