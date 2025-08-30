package database

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-templ-template/internal/config"

	"github.com/stretchr/testify/require"
)

// TestMigrationManager tests the migration manager functionality
func TestMigrationManager(t *testing.T) {
	// Skip if no test database is configured
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping migration manager tests")
	}

	cfg := &config.DatabaseConfig{
		URL: os.Getenv("TEST_DATABASE_URL"),
	}

	// Create temporary migrations directory for testing
	tempDir := t.TempDir()
	migrationsPath := filepath.Join(tempDir, "migrations")
	err := os.MkdirAll(migrationsPath, 0755)
	require.NoError(t, err)

	// Create migration manager
	manager, err := NewMigrationManager(cfg, migrationsPath)
	require.NoError(t, err, "Failed to create migration manager")
	defer manager.Close()

	// Test creating a migration
	migration, err := manager.CreateMigration("create users table")
	require.NoError(t, err, "Failed to create migration")
	require.Equal(t, uint(1), migration.Version)
	require.Equal(t, "create_users_table", migration.Name)
	require.Equal(t, "001_create_users_table.up.sql", migration.UpFile)
	require.Equal(t, "001_create_users_table.down.sql", migration.DownFile)

	// Verify files were created
	upPath := filepath.Join(migrationsPath, migration.UpFile)
	downPath := filepath.Join(migrationsPath, migration.DownFile)

	_, err = os.Stat(upPath)
	require.NoError(t, err, "Up migration file should exist")

	_, err = os.Stat(downPath)
	require.NoError(t, err, "Down migration file should exist")

	// Test creating another migration
	migration2, err := manager.CreateMigration("add user indexes")
	require.NoError(t, err, "Failed to create second migration")
	require.Equal(t, uint(2), migration2.Version)
	require.Equal(t, "add_user_indexes", migration2.Name)

	// Test listing migrations
	migrations, err := manager.ListMigrations()
	require.NoError(t, err, "Failed to list migrations")
	require.Len(t, migrations, 2)
	require.Equal(t, uint(1), migrations[0].Version)
	require.Equal(t, uint(2), migrations[1].Version)

	// Test getting status
	status, err := manager.GetStatus()
	require.NoError(t, err, "Failed to get status")
	require.Equal(t, uint(2), status.LatestVersion)
	require.Equal(t, 2, status.TotalMigrations)
	require.Equal(t, 2, status.PendingCount)
	require.Equal(t, 0, status.AppliedCount)
}

// TestMigrationManagerWithRealMigrations tests with actual migration files
func TestMigrationManagerWithRealMigrations(t *testing.T) {
	// Skip if no test database is configured
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping migration manager tests")
	}

	cfg := &config.DatabaseConfig{
		URL: os.Getenv("TEST_DATABASE_URL"),
	}

	// Use real migrations directory
	migrationsPath := filepath.Join("..", "..", "..", "migrations")

	manager, err := NewMigrationManager(cfg, migrationsPath)
	require.NoError(t, err, "Failed to create migration manager")
	defer manager.Close()

	// Test listing real migrations
	migrations, err := manager.ListMigrations()
	require.NoError(t, err, "Failed to list migrations")
	require.Greater(t, len(migrations), 0, "Should have at least one migration")

	// Test validation
	err = manager.ValidateMigrations()
	require.NoError(t, err, "Migration validation should pass")

	// Test getting status
	status, err := manager.GetStatus()
	require.NoError(t, err, "Failed to get status")
	require.Greater(t, status.TotalMigrations, 0)
	require.Equal(t, len(migrations), status.TotalMigrations)

	// Test migration operations
	currentVersion, _, err := manager.GetCurrentVersion()
	if err != nil {
		t.Logf("No current version (clean database): %v", err)
	}

	// Run migrations up
	err = manager.MigrateUp()
	require.NoError(t, err, "Failed to migrate up")

	// Check status after migration
	status, err = manager.GetStatus()
	require.NoError(t, err, "Failed to get status after migration")
	require.True(t, status.IsUpToDate(), "Should be up to date after migration")
	require.Equal(t, 0, status.PendingCount)

	// Test migrating to specific version
	if len(migrations) > 1 {
		targetVersion := migrations[0].Version
		err = manager.MigrateTo(targetVersion)
		require.NoError(t, err, "Failed to migrate to specific version")

		currentVersion, _, err = manager.GetCurrentVersion()
		require.NoError(t, err, "Failed to get current version")
		require.Equal(t, targetVersion, currentVersion)
	}

	// Clean up - migrate down
	err = manager.MigrateDown()
	require.NoError(t, err, "Failed to migrate down")
}

// TestMigrationCreation tests migration file creation
func TestMigrationCreation(t *testing.T) {
	// Skip if no test database is configured
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping migration creation tests")
	}

	cfg := &config.DatabaseConfig{
		URL: os.Getenv("TEST_DATABASE_URL"),
	}

	// Create temporary migrations directory
	tempDir := t.TempDir()
	migrationsPath := filepath.Join(tempDir, "migrations")
	err := os.MkdirAll(migrationsPath, 0755)
	require.NoError(t, err)

	manager, err := NewMigrationManager(cfg, migrationsPath)
	require.NoError(t, err)
	defer manager.Close()

	// Test creating migration with various names
	testCases := []struct {
		name     string
		expected string
	}{
		{"create users table", "create_users_table"},
		{"Add User Indexes", "add_user_indexes"},
		{"update-user-schema", "update_user_schema"},
		{"Fix User Constraints", "fix_user_constraints"},
	}

	for i, tc := range testCases {
		migration, err := manager.CreateMigration(tc.name)
		require.NoError(t, err, "Failed to create migration for: %s", tc.name)
		require.Equal(t, uint(i+1), migration.Version)
		require.Equal(t, tc.expected, migration.Name)

		// Verify file content
		upPath := filepath.Join(migrationsPath, migration.UpFile)
		content, err := os.ReadFile(upPath)
		require.NoError(t, err)
		require.Contains(t, string(content), tc.name)
		require.Contains(t, string(content), "-- Migration:")
	}

	// Test duplicate migration creation should fail
	_, err = manager.CreateMigration("create users table")
	require.Error(t, err, "Should fail to create duplicate migration")

	// Test empty name should fail
	_, err = manager.CreateMigration("")
	require.Error(t, err, "Should fail to create migration with empty name")
}

// TestMigrationValidation tests migration validation
func TestMigrationValidation(t *testing.T) {
	// Skip if no test database is configured
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping migration validation tests")
	}

	cfg := &config.DatabaseConfig{
		URL: os.Getenv("TEST_DATABASE_URL"),
	}

	// Create temporary migrations directory
	tempDir := t.TempDir()
	migrationsPath := filepath.Join(tempDir, "migrations")
	err := os.MkdirAll(migrationsPath, 0755)
	require.NoError(t, err)

	manager, err := NewMigrationManager(cfg, migrationsPath)
	require.NoError(t, err)
	defer manager.Close()

	// Create valid migrations
	upContent := `-- Migration: create users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL
);`

	downContent := `-- Migration rollback: create users table
DROP TABLE IF EXISTS users;`

	// Create migration files
	err = os.WriteFile(filepath.Join(migrationsPath, "001_create_users.up.sql"), []byte(upContent), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(migrationsPath, "001_create_users.down.sql"), []byte(downContent), 0644)
	require.NoError(t, err)

	// Validation should pass
	err = manager.ValidateMigrations()
	require.NoError(t, err, "Validation should pass for valid migrations")

	// Create migration with gap in version numbers
	err = os.WriteFile(filepath.Join(migrationsPath, "003_add_indexes.up.sql"), []byte(upContent), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(migrationsPath, "003_add_indexes.down.sql"), []byte(downContent), 0644)
	require.NoError(t, err)

	// Validation should fail due to gap
	err = manager.ValidateMigrations()
	require.Error(t, err, "Validation should fail due to version gap")
	require.Contains(t, err.Error(), "gap detected")

	// Remove the gap file
	os.Remove(filepath.Join(migrationsPath, "003_add_indexes.up.sql"))
	os.Remove(filepath.Join(migrationsPath, "003_add_indexes.down.sql"))

	// Create empty migration file
	err = os.WriteFile(filepath.Join(migrationsPath, "002_empty.up.sql"), []byte("-- Empty migration"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(migrationsPath, "002_empty.down.sql"), []byte("-- Empty rollback"), 0644)
	require.NoError(t, err)

	// Validation should fail due to empty migration
	err = manager.ValidateMigrations()
	require.Error(t, err, "Validation should fail for empty migration")
	require.Contains(t, err.Error(), "empty or contains no SQL")
}

// TestMigrationStatus tests migration status functionality
func TestMigrationStatus(t *testing.T) {
	status := &MigrationStatus{
		CurrentVersion:  2,
		LatestVersion:   5,
		IsDirty:         false,
		AppliedCount:    2,
		PendingCount:    3,
		TotalMigrations: 5,
	}

	// Test status methods
	require.False(t, status.IsUpToDate())
	require.True(t, status.NeedsMigration())

	statusString := status.String()
	require.Contains(t, statusString, "Migration needed")
	require.Contains(t, statusString, "2/5 applied")

	// Test up-to-date status
	status.CurrentVersion = 5
	status.PendingCount = 0
	status.AppliedCount = 5

	require.True(t, status.IsUpToDate())
	require.False(t, status.NeedsMigration())

	statusString = status.String()
	require.Contains(t, statusString, "All migrations applied")

	// Test dirty status
	status.IsDirty = true
	require.True(t, status.NeedsMigration())

	statusString = status.String()
	require.Contains(t, statusString, "dirty")
}

// TestMigrationConcurrentCreation tests concurrent migration creation
func TestMigrationConcurrentCreation(t *testing.T) {
	// Skip if no test database is configured
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping concurrent migration tests")
	}

	cfg := &config.DatabaseConfig{
		URL: os.Getenv("TEST_DATABASE_URL"),
	}

	// Create temporary migrations directory
	tempDir := t.TempDir()
	migrationsPath := filepath.Join(tempDir, "migrations")
	err := os.MkdirAll(migrationsPath, 0755)
	require.NoError(t, err)

	// Create multiple managers (simulating concurrent access)
	manager1, err := NewMigrationManager(cfg, migrationsPath)
	require.NoError(t, err)
	defer manager1.Close()

	manager2, err := NewMigrationManager(cfg, migrationsPath)
	require.NoError(t, err)
	defer manager2.Close()

	// Try to create migrations concurrently
	done1 := make(chan error, 1)
	done2 := make(chan error, 1)

	go func() {
		_, err := manager1.CreateMigration("migration from manager 1")
		done1 <- err
	}()

	go func() {
		time.Sleep(10 * time.Millisecond) // Small delay
		_, err := manager2.CreateMigration("migration from manager 2")
		done2 <- err
	}()

	err1 := <-done1
	err2 := <-done2

	// Both should succeed (they create different files)
	require.NoError(t, err1, "Manager 1 should succeed")
	require.NoError(t, err2, "Manager 2 should succeed")

	// Verify both migrations exist
	migrations, err := manager1.ListMigrations()
	require.NoError(t, err)
	require.Len(t, migrations, 2)
}

// BenchmarkMigrationCreation benchmarks migration creation
func BenchmarkMigrationCreation(b *testing.B) {
	// Skip if no test database is configured
	if os.Getenv("TEST_DATABASE_URL") == "" {
		b.Skip("TEST_DATABASE_URL not set, skipping migration benchmarks")
	}

	cfg := &config.DatabaseConfig{
		URL: os.Getenv("TEST_DATABASE_URL"),
	}

	tempDir := b.TempDir()
	migrationsPath := filepath.Join(tempDir, "migrations")
	err := os.MkdirAll(migrationsPath, 0755)
	require.NoError(b, err)

	manager, err := NewMigrationManager(cfg, migrationsPath)
	require.NoError(b, err)
	defer manager.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		name := fmt.Sprintf("benchmark_migration_%d", i)
		_, err := manager.CreateMigration(name)
		require.NoError(b, err)
	}
}

// BenchmarkMigrationListing benchmarks migration listing
func BenchmarkMigrationListing(b *testing.B) {
	// Skip if no test database is configured
	if os.Getenv("TEST_DATABASE_URL") == "" {
		b.Skip("TEST_DATABASE_URL not set, skipping migration benchmarks")
	}

	cfg := &config.DatabaseConfig{
		URL: os.Getenv("TEST_DATABASE_URL"),
	}

	// Use real migrations directory
	migrationsPath := filepath.Join("..", "..", "..", "migrations")

	manager, err := NewMigrationManager(cfg, migrationsPath)
	require.NoError(b, err)
	defer manager.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.ListMigrations()
		require.NoError(b, err)
	}
}
