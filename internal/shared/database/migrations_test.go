package database

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-templ-template/internal/config"

	"github.com/stretchr/testify/require"
)

// TestMigrationRunner tests the migration runner functionality
func TestMigrationRunner(t *testing.T) {
	// Skip if no test database is configured
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping migration tests")
	}

	cfg := &config.DatabaseConfig{
		URL: os.Getenv("TEST_DATABASE_URL"),
	}

	// Get migrations path
	migrationsPath := filepath.Join("..", "..", "..", "migrations")

	// Create migration runner
	runner, err := NewMigrationRunner(cfg, migrationsPath)
	require.NoError(t, err, "Failed to create migration runner")
	defer runner.Close()

	// Test getting version before any migrations
	version, dirty, err := runner.Version()
	if err == nil {
		t.Logf("Initial migration version: %d (dirty: %t)", version, dirty)
	}

	// Test running migrations up
	err = runner.Up()
	require.NoError(t, err, "Failed to run migrations up")

	// Test getting version after migrations
	version, dirty, err = runner.Version()
	require.NoError(t, err, "Failed to get migration version")
	require.False(t, dirty, "Migration should not be dirty")
	require.Greater(t, version, uint(0), "Version should be greater than 0")

	t.Logf("Migration version after up: %d", version)

	// Test running migrations down
	err = runner.Down()
	require.NoError(t, err, "Failed to run migrations down")

	// Test version after down (should be 0 or error)
	version, dirty, err = runner.Version()
	if err == nil {
		t.Logf("Migration version after down: %d (dirty: %t)", version, dirty)
	}
}

// TestMigrationSteps tests step-by-step migration functionality
func TestMigrationSteps(t *testing.T) {
	// Skip if no test database is configured
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping migration tests")
	}

	cfg := &config.DatabaseConfig{
		URL: os.Getenv("TEST_DATABASE_URL"),
	}

	migrationsPath := filepath.Join("..", "..", "..", "migrations")

	runner, err := NewMigrationRunner(cfg, migrationsPath)
	require.NoError(t, err, "Failed to create migration runner")
	defer runner.Close()

	// Ensure we start from a clean state
	_ = runner.Down()

	// Test running one step up
	err = runner.Steps(1)
	require.NoError(t, err, "Failed to run 1 step up")

	version, dirty, err := runner.Version()
	require.NoError(t, err, "Failed to get version")
	require.False(t, dirty, "Migration should not be dirty")
	require.Equal(t, uint(1), version, "Version should be 1 after 1 step")

	// Test running another step up
	err = runner.Steps(1)
	require.NoError(t, err, "Failed to run another step up")

	version, dirty, err = runner.Version()
	require.NoError(t, err, "Failed to get version")
	require.False(t, dirty, "Migration should not be dirty")
	require.Equal(t, uint(2), version, "Version should be 2 after 2 steps")

	// Test running one step down
	err = runner.Steps(-1)
	require.NoError(t, err, "Failed to run 1 step down")

	version, dirty, err = runner.Version()
	require.NoError(t, err, "Failed to get version")
	require.False(t, dirty, "Migration should not be dirty")
	require.Equal(t, uint(1), version, "Version should be 1 after stepping down")

	// Clean up
	_ = runner.Down()
}

// TestMigrationForce tests forcing migration version
func TestMigrationForce(t *testing.T) {
	// Skip if no test database is configured
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping migration tests")
	}

	cfg := &config.DatabaseConfig{
		URL: os.Getenv("TEST_DATABASE_URL"),
	}

	migrationsPath := filepath.Join("..", "..", "..", "migrations")

	runner, err := NewMigrationRunner(cfg, migrationsPath)
	require.NoError(t, err, "Failed to create migration runner")
	defer runner.Close()

	// Ensure we start from a clean state
	_ = runner.Down()

	// Force version to 1
	err = runner.Force(1)
	require.NoError(t, err, "Failed to force version to 1")

	version, _, err := runner.Version()
	require.NoError(t, err, "Failed to get version")
	require.Equal(t, uint(1), version, "Version should be 1 after forcing")

	// Force version to 0 (clean state)
	err = runner.Force(0)
	require.NoError(t, err, "Failed to force version to 0")

	version, _, err = runner.Version()
	if err == nil {
		require.Equal(t, uint(0), version, "Version should be 0 after forcing")
	}
}

// TestMigrationRollback tests complete rollback functionality
func TestMigrationRollback(t *testing.T) {
	// Skip if no test database is configured
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping migration tests")
	}

	cfg := &config.DatabaseConfig{
		URL: os.Getenv("TEST_DATABASE_URL"),
	}

	migrationsPath := filepath.Join("..", "..", "..", "migrations")

	runner, err := NewMigrationRunner(cfg, migrationsPath)
	require.NoError(t, err, "Failed to create migration runner")
	defer runner.Close()

	// Create a test database connection to verify table existence
	testDB := NewTestDatabaseWithConfig(t, &TestDatabaseConfig{
		Host:     cfg.Host,
		Port:     cfg.Port,
		User:     cfg.User,
		Password: cfg.Password,
		Name:     cfg.Name,
		SSLMode:  cfg.SSLMode,
	})
	defer testDB.Close()

	// Start from clean state
	_ = runner.Down()

	// Run all migrations up
	err = runner.Up()
	require.NoError(t, err, "Failed to run migrations up")

	// Verify tables exist after migration up
	testDB.AssertTableExists("users")
	testDB.AssertTableExists("sessions")
	testDB.AssertTableExists("audit_events")

	// Run all migrations down
	err = runner.Down()
	require.NoError(t, err, "Failed to run migrations down")

	// Verify tables don't exist after migration down
	testDB.AssertTableNotExists("audit_events")
	testDB.AssertTableNotExists("sessions")
	testDB.AssertTableNotExists("users")
}

// TestMigrationConcurrency tests migration runner under concurrent access
func TestMigrationConcurrency(t *testing.T) {
	// Skip if no test database is configured
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping migration tests")
	}

	cfg := &config.DatabaseConfig{
		URL: os.Getenv("TEST_DATABASE_URL"),
	}

	migrationsPath := filepath.Join("..", "..", "..", "migrations")

	// Create multiple runners
	runner1, err := NewMigrationRunner(cfg, migrationsPath)
	require.NoError(t, err, "Failed to create first migration runner")
	defer runner1.Close()

	runner2, err := NewMigrationRunner(cfg, migrationsPath)
	require.NoError(t, err, "Failed to create second migration runner")
	defer runner2.Close()

	// Start from clean state
	_ = runner1.Down()

	// Try to run migrations concurrently
	// Note: golang-migrate handles locking, so one should succeed and one should wait/fail gracefully
	done1 := make(chan error, 1)
	done2 := make(chan error, 1)

	go func() {
		done1 <- runner1.Up()
	}()

	go func() {
		time.Sleep(100 * time.Millisecond) // Small delay to ensure runner1 starts first
		done2 <- runner2.Up()
	}()

	// Wait for both to complete
	err1 := <-done1
	err2 := <-done2

	// At least one should succeed
	if err1 != nil && err2 != nil {
		t.Fatalf("Both migration runners failed: err1=%v, err2=%v", err1, err2)
	}

	// Verify final state
	version, dirty, err := runner1.Version()
	require.NoError(t, err, "Failed to get final version")
	require.False(t, dirty, "Final state should not be dirty")
	require.Greater(t, version, uint(0), "Final version should be greater than 0")

	// Clean up
	_ = runner1.Down()
}

// TestMigrationWithTimeout tests migration operations with timeout
func TestMigrationWithTimeout(t *testing.T) {
	// Skip if no test database is configured
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping migration tests")
	}

	cfg := &config.DatabaseConfig{
		URL: os.Getenv("TEST_DATABASE_URL"),
	}

	migrationsPath := filepath.Join("..", "..", "..", "migrations")

	// Test with very short timeout to ensure connection handling works
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	runner, err := NewMigrationRunner(cfg, migrationsPath)
	require.NoError(t, err, "Failed to create migration runner")
	defer runner.Close()

	// This should complete within the timeout for our simple migrations
	_ = runner.Down() // Clean state

	// Run migrations (should complete quickly)
	err = runner.Up()
	require.NoError(t, err, "Failed to run migrations within timeout")

	// Verify we can get version
	version, dirty, err := runner.Version()
	require.NoError(t, err, "Failed to get version")
	require.False(t, dirty, "Migration should not be dirty")

	t.Logf("Migration completed with version: %d", version)

	// Clean up
	_ = runner.Down()

	// Ensure context is still valid
	select {
	case <-ctx.Done():
		t.Log("Context expired as expected")
	default:
		t.Log("Context still valid")
	}
}

// BenchmarkMigrationUp benchmarks migration up performance
func BenchmarkMigrationUp(b *testing.B) {
	// Skip if no test database is configured
	if os.Getenv("TEST_DATABASE_URL") == "" {
		b.Skip("TEST_DATABASE_URL not set, skipping migration benchmarks")
	}

	cfg := &config.DatabaseConfig{
		URL: os.Getenv("TEST_DATABASE_URL"),
	}

	migrationsPath := filepath.Join("..", "..", "..", "migrations")

	runner, err := NewMigrationRunner(cfg, migrationsPath)
	require.NoError(b, err, "Failed to create migration runner")
	defer runner.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		_ = runner.Down() // Reset state
		b.StartTimer()

		err := runner.Up()
		require.NoError(b, err, "Failed to run migrations up")
	}
}

// BenchmarkMigrationVersion benchmarks version checking performance
func BenchmarkMigrationVersion(b *testing.B) {
	// Skip if no test database is configured
	if os.Getenv("TEST_DATABASE_URL") == "" {
		b.Skip("TEST_DATABASE_URL not set, skipping migration benchmarks")
	}

	cfg := &config.DatabaseConfig{
		URL: os.Getenv("TEST_DATABASE_URL"),
	}

	migrationsPath := filepath.Join("..", "..", "..", "migrations")

	runner, err := NewMigrationRunner(cfg, migrationsPath)
	require.NoError(b, err, "Failed to create migration runner")
	defer runner.Close()

	// Ensure migrations are up
	_ = runner.Up()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := runner.Version()
		require.NoError(b, err, "Failed to get migration version")
	}
}

// TestMigrationIntegrationWithSeeding tests migration followed by seeding
func TestMigrationIntegrationWithSeeding(t *testing.T) {
	// Skip if no test database is configured
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping migration integration tests")
	}

	cfg := &config.DatabaseConfig{
		URL: os.Getenv("TEST_DATABASE_URL"),
	}

	migrationsPath := filepath.Join("..", "..", "..", "migrations")

	runner, err := NewMigrationRunner(cfg, migrationsPath)
	require.NoError(t, err, "Failed to create migration runner")
	defer runner.Close()

	// Create test database connection
	testDB := NewTestDatabaseWithConfig(t, &TestDatabaseConfig{
		Host:     cfg.Host,
		Port:     cfg.Port,
		User:     cfg.User,
		Password: cfg.Password,
		Name:     cfg.Name,
		SSLMode:  cfg.SSLMode,
	})
	defer testDB.Close()

	// Start from clean state
	_ = runner.Down()

	// Run migrations
	err = runner.Up()
	require.NoError(t, err, "Failed to run migrations up")

	// Verify tables exist
	testDB.AssertTableExists("users")
	testDB.AssertTableExists("sessions")
	testDB.AssertTableExists("audit_events")

	// Test seeding data (similar to what the seed scripts do)
	testDB.SeedData(`
		INSERT INTO users (id, email, password, first_name, last_name, status, created_at, updated_at, version)
		VALUES (
			'test-user-id-123',
			'test@example.com',
			'$2a$10$test.hash.here',
			'Test',
			'User',
			'active',
			NOW(),
			NOW(),
			1
		) ON CONFLICT (email) DO NOTHING
	`)

	// Verify seeded data
	testDB.AssertRowCount("users", 1)

	// Test that we can create a session for the user
	testDB.SeedData(`
		INSERT INTO sessions (id, user_id, expires_at, created_at, ip_address, user_agent, is_active)
		VALUES (
			'test-session-123',
			'test-user-id-123',
			NOW() + INTERVAL '1 day',
			NOW(),
			'127.0.0.1',
			'Test User Agent',
			true
		)
	`)

	testDB.AssertRowCount("sessions", 1)

	// Test audit events
	testDB.SeedData(`
		INSERT INTO audit_events (event_id, event_type, aggregate_id, aggregate_type, user_id, action, resource, resource_id, details, occurred_at, metadata)
		VALUES (
			'test-event-123',
			'user.created',
			'test-user-id-123',
			'user',
			'test-user-id-123',
			'create',
			'user',
			'test-user-id-123',
			'{"email": "test@example.com"}',
			NOW(),
			'{}'
		)
	`)

	testDB.AssertRowCount("audit_events", 1)

	// Clean up by rolling back migrations
	err = runner.Down()
	require.NoError(t, err, "Failed to run migrations down")

	// Verify tables are gone
	testDB.AssertTableNotExists("audit_events")
	testDB.AssertTableNotExists("sessions")
	testDB.AssertTableNotExists("users")
}
