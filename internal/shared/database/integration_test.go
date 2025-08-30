package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go-templ-template/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMigrationAndSeedingIntegration tests the complete migration and seeding workflow
func TestMigrationAndSeedingIntegration(t *testing.T) {
	// Skip if no test database is configured
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration tests")
	}

	cfg := &config.DatabaseConfig{
		URL: os.Getenv("TEST_DATABASE_URL"),
	}

	// Use real migrations directory
	migrationsPath := filepath.Join("..", "..", "..", "migrations")

	// Create migration manager
	manager, err := NewMigrationManager(cfg, migrationsPath)
	require.NoError(t, err, "Failed to create migration manager")
	defer manager.Close()

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
	err = manager.MigrateDown()
	if err != nil {
		t.Logf("Warning: could not migrate down (probably clean database): %v", err)
	}

	// Test 1: Verify clean state
	testDB.AssertTableNotExists("users")
	testDB.AssertTableNotExists("sessions")
	testDB.AssertTableNotExists("audit_events")

	// Test 2: Run migrations up
	err = manager.MigrateUp()
	require.NoError(t, err, "Failed to run migrations up")

	// Verify tables exist after migration
	testDB.AssertTableExists("users")
	testDB.AssertTableExists("sessions")
	testDB.AssertTableExists("audit_events")

	// Test 3: Verify migration status
	status, err := manager.GetStatus()
	require.NoError(t, err, "Failed to get migration status")
	require.True(t, status.IsUpToDate(), "Should be up to date after migration")
	require.False(t, status.IsDirty, "Should not be dirty")
	require.Equal(t, 0, status.PendingCount, "Should have no pending migrations")

	// Test 4: Seed data (similar to seed scripts)
	t.Run("SeedUsers", func(t *testing.T) {
		seedUsers(t, testDB)
	})

	t.Run("SeedSessions", func(t *testing.T) {
		seedSessions(t, testDB)
	})

	t.Run("SeedAuditEvents", func(t *testing.T) {
		seedAuditEvents(t, testDB)
	})

	// Test 5: Verify seeded data integrity
	t.Run("VerifyDataIntegrity", func(t *testing.T) {
		verifyDataIntegrity(t, testDB)
	})

	// Test 6: Test rollback with data
	t.Run("TestRollbackWithData", func(t *testing.T) {
		// Rollback should cascade and remove all data
		err = manager.MigrateDown()
		require.NoError(t, err, "Failed to rollback migrations")

		// Verify tables are gone
		testDB.AssertTableNotExists("audit_events")
		testDB.AssertTableNotExists("sessions")
		testDB.AssertTableNotExists("users")
	})

	// Test 7: Test re-migration and re-seeding
	t.Run("TestReMigrationAndReSeeding", func(t *testing.T) {
		// Migrate up again
		err = manager.MigrateUp()
		require.NoError(t, err, "Failed to re-migrate up")

		// Re-seed data
		seedUsers(t, testDB)
		seedSessions(t, testDB)
		seedAuditEvents(t, testDB)

		// Verify data is there again
		testDB.AssertRowCount("users", 4)
		testDB.AssertRowCount("sessions", 1)
		testDB.AssertRowCount("audit_events", 5)
	})

	// Clean up
	_ = manager.MigrateDown()
}

// TestStepByStepMigration tests step-by-step migration with data validation
func TestStepByStepMigration(t *testing.T) {
	// Skip if no test database is configured
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping step-by-step migration tests")
	}

	cfg := &config.DatabaseConfig{
		URL: os.Getenv("TEST_DATABASE_URL"),
	}

	migrationsPath := filepath.Join("..", "..", "..", "migrations")

	manager, err := NewMigrationManager(cfg, migrationsPath)
	require.NoError(t, err)
	defer manager.Close()

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
	_ = manager.MigrateDown()

	// Get list of migrations
	migrations, err := manager.ListMigrations()
	require.NoError(t, err)
	require.Greater(t, len(migrations), 0, "Should have migrations")

	// Test each migration step
	for _, migration := range migrations {
		t.Run(fmt.Sprintf("Migration_%03d_%s", migration.Version, migration.Name), func(t *testing.T) {
			// Run one step up
			err = manager.MigrateSteps(1)
			require.NoError(t, err, "Failed to run migration step %d", migration.Version)

			// Verify version
			currentVersion, dirty, err := manager.GetCurrentVersion()
			require.NoError(t, err)
			require.False(t, dirty, "Migration should not be dirty")
			require.Equal(t, migration.Version, currentVersion, "Version should match")

			// Test specific migration effects
			switch migration.Version {
			case 1:
				// After first migration, should have users and sessions tables
				testDB.AssertTableExists("users")
				testDB.AssertTableExists("sessions")

				// Test that we can insert a user
				testDB.SeedData(`
					INSERT INTO users (id, email, password, first_name, last_name, status)
					VALUES ('test-user', 'test@example.com', 'hash', 'Test', 'User', 'active')
				`)
				testDB.AssertRowCount("users", 1)

			case 2:
				// After second migration, sessions should have is_active column
				testDB.SeedData(`
					INSERT INTO sessions (id, user_id, expires_at, is_active)
					VALUES ('test-session', 'test-user', NOW() + INTERVAL '1 day', true)
				`)
				testDB.AssertRowCount("sessions", 1)

			case 3:
				// After third migration, should have audit_events table
				testDB.AssertTableExists("audit_events")

				testDB.SeedData(`
					INSERT INTO audit_events (event_id, event_type, aggregate_id, aggregate_type, user_id, action, resource, resource_id, details, occurred_at)
					VALUES ('test-event', 'test.event', 'test-user', 'user', 'test-user', 'test', 'user', 'test-user', '{}', NOW())
				`)
				testDB.AssertRowCount("audit_events", 1)
			}
		})
	}

	// Test rollback step by step
	for i := len(migrations) - 1; i >= 0; i-- {
		migration := migrations[i]
		t.Run(fmt.Sprintf("Rollback_%03d_%s", migration.Version, migration.Name), func(t *testing.T) {
			// Run one step down
			err = manager.MigrateSteps(-1)
			require.NoError(t, err, "Failed to rollback migration step %d", migration.Version)

			// Verify version
			currentVersion, dirty, err := manager.GetCurrentVersion()
			if i == 0 {
				// After rolling back all migrations, version might be 0 or error
				if err == nil {
					require.Equal(t, uint(0), currentVersion)
				}
			} else {
				require.NoError(t, err)
				require.False(t, dirty, "Migration should not be dirty")
				require.Equal(t, migrations[i-1].Version, currentVersion, "Version should match previous migration")
			}
		})
	}

	// Verify clean state
	testDB.AssertTableNotExists("audit_events")
	testDB.AssertTableNotExists("sessions")
	testDB.AssertTableNotExists("users")
}

// TestMigrationErrorRecovery tests error recovery scenarios
func TestMigrationErrorRecovery(t *testing.T) {
	// Skip if no test database is configured
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping error recovery tests")
	}

	cfg := &config.DatabaseConfig{
		URL: os.Getenv("TEST_DATABASE_URL"),
	}

	migrationsPath := filepath.Join("..", "..", "..", "migrations")

	manager, err := NewMigrationManager(cfg, migrationsPath)
	require.NoError(t, err)
	defer manager.Close()

	// Start from clean state
	_ = manager.MigrateDown()

	// Run migrations up
	err = manager.MigrateUp()
	require.NoError(t, err)

	// Test force version (simulating recovery from dirty state)
	currentVersion, _, err := manager.GetCurrentVersion()
	require.NoError(t, err)

	// Force to current version (should be safe)
	err = manager.Force(int(currentVersion))
	require.NoError(t, err, "Should be able to force to current version")

	// Verify we can still operate normally
	status, err := manager.GetStatus()
	require.NoError(t, err)
	require.False(t, status.IsDirty, "Should not be dirty after force")

	// Clean up
	_ = manager.MigrateDown()
}

// Helper functions for seeding data

func seedUsers(t *testing.T, testDB *TestDatabase) {
	// Hash for password "password123" (bcrypt cost 10)
	passwordHash := "$2a$10$rOCKx7VDum0oaEFrZQWOa.6nYgs8o8/Oa9Q8Ks8Ks8Ks8Ks8Ks8K"

	testDB.SeedData(`
		INSERT INTO users (id, email, password, first_name, last_name, status, created_at, updated_at, version)
		VALUES 
		($1, 'admin@example.com', $2, 'Admin', 'User', 'active', NOW(), NOW(), 1),
		($3, 'john.doe@example.com', $2, 'John', 'Doe', 'active', NOW(), NOW(), 1),
		($4, 'jane.smith@example.com', $2, 'Jane', 'Smith', 'active', NOW(), NOW(), 1),
		($5, 'bob.wilson@example.com', $2, 'Bob', 'Wilson', 'inactive', NOW(), NOW(), 1)
		ON CONFLICT (email) DO NOTHING
	`, "admin-user-id-1234567890", passwordHash, "user-id-1234567890", "user-id-0987654321", "user-id-1122334455")

	testDB.AssertRowCount("users", 4)
}

func seedSessions(t *testing.T, testDB *TestDatabase) {
	testDB.SeedData(`
		INSERT INTO sessions (id, user_id, expires_at, created_at, ip_address, user_agent, is_active)
		VALUES ($1, $2, NOW() + INTERVAL '1 day', NOW(), '127.0.0.1', 'Mozilla/5.0 (Development Seed)', true)
		ON CONFLICT (id) DO NOTHING
	`, "session-admin-123456", "admin-user-id-1234567890")

	testDB.AssertRowCount("sessions", 1)
}

func seedAuditEvents(t *testing.T, testDB *TestDatabase) {
	testDB.SeedData(`
		INSERT INTO audit_events (event_id, event_type, aggregate_id, aggregate_type, user_id, action, resource, resource_id, details, occurred_at, metadata)
		VALUES 
		($1, 'user.created', $2, 'user', $2, 'create', 'user', $2, $3, NOW() - INTERVAL '1 hour', $4),
		($5, 'user.created', $6, 'user', $6, 'create', 'user', $6, $7, NOW() - INTERVAL '30 minutes', $4),
		($8, 'user.login', $2, 'session', $2, 'login', 'session', $9, $10, NOW() - INTERVAL '15 minutes', $4),
		($11, 'user.created', $12, 'user', $12, 'create', 'user', $12, $13, NOW() - INTERVAL '25 minutes', $4),
		($14, 'user.created', $15, 'user', $15, 'create', 'user', $15, $16, NOW() - INTERVAL '20 minutes', $4)
		ON CONFLICT (event_id) DO NOTHING
	`,
		"audit-event-1", "admin-user-id-1234567890", `{"email": "admin@example.com", "first_name": "Admin", "last_name": "User"}`, `{"source": "test_script", "version": "1.0"}`,
		"audit-event-2", "user-id-1234567890", `{"email": "john.doe@example.com", "first_name": "John", "last_name": "Doe"}`,
		"audit-event-3", "session-admin-123456", `{"session_id": "session-admin-123456", "ip_address": "127.0.0.1", "user_agent": "Mozilla/5.0 (Development Seed)"}`,
		"audit-event-4", "user-id-0987654321", `{"email": "jane.smith@example.com", "first_name": "Jane", "last_name": "Smith"}`,
		"audit-event-5", "user-id-1122334455", `{"email": "bob.wilson@example.com", "first_name": "Bob", "last_name": "Wilson", "status": "inactive"}`)

	testDB.AssertRowCount("audit_events", 5)
}

func verifyDataIntegrity(t *testing.T, testDB *TestDatabase) {
	// Test foreign key relationships
	var sessionCount int
	err := testDB.DB.Get(&sessionCount, `
		SELECT COUNT(*) FROM sessions s 
		JOIN users u ON s.user_id = u.id 
		WHERE u.email = 'admin@example.com'
	`)
	require.NoError(t, err)
	require.Equal(t, 1, sessionCount, "Should have one session for admin user")

	// Test audit events for users
	var auditCount int
	err = testDB.DB.Get(&auditCount, `
		SELECT COUNT(*) FROM audit_events 
		WHERE event_type = 'user.created'
	`)
	require.NoError(t, err)
	require.Equal(t, 4, auditCount, "Should have 4 user creation audit events")

	// Test user status constraints
	var activeUserCount int
	err = testDB.DB.Get(&activeUserCount, `
		SELECT COUNT(*) FROM users 
		WHERE status = 'active'
	`)
	require.NoError(t, err)
	require.Equal(t, 3, activeUserCount, "Should have 3 active users")

	// Test session expiration
	var validSessionCount int
	err = testDB.DB.Get(&validSessionCount, `
		SELECT COUNT(*) FROM sessions 
		WHERE expires_at > NOW() AND is_active = true
	`)
	require.NoError(t, err)
	require.Equal(t, 1, validSessionCount, "Should have 1 valid active session")

	// Test audit event JSON data
	var adminEmail string
	err = testDB.DB.Get(&adminEmail, `
		SELECT details->>'email' FROM audit_events 
		WHERE event_type = 'user.created' AND aggregate_id = 'admin-user-id-1234567890'
	`)
	require.NoError(t, err)
	require.Equal(t, "admin@example.com", adminEmail, "Audit event should contain correct email")
}

// TestPerformanceWithLargeDataset tests migration and seeding performance with larger datasets
func TestPerformanceWithLargeDataset(t *testing.T) {
	// Skip if no test database is configured or if not running performance tests
	if os.Getenv("TEST_DATABASE_URL") == "" || os.Getenv("RUN_PERFORMANCE_TESTS") == "" {
		t.Skip("Skipping performance tests (set TEST_DATABASE_URL and RUN_PERFORMANCE_TESTS to run)")
	}

	cfg := &config.DatabaseConfig{
		URL: os.Getenv("TEST_DATABASE_URL"),
	}

	migrationsPath := filepath.Join("..", "..", "..", "migrations")

	manager, err := NewMigrationManager(cfg, migrationsPath)
	require.NoError(t, err)
	defer manager.Close()

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
	_ = manager.MigrateDown()

	// Time migration
	start := time.Now()
	err = manager.MigrateUp()
	require.NoError(t, err)
	migrationTime := time.Since(start)

	t.Logf("Migration time: %v", migrationTime)

	// Seed large dataset
	start = time.Now()
	seedLargeDataset(t, testDB, 1000) // 1000 users
	seedingTime := time.Since(start)

	t.Logf("Seeding time for 1000 users: %v", seedingTime)

	// Verify data
	testDB.AssertRowCount("users", 1000)

	// Test rollback performance
	start = time.Now()
	err = manager.MigrateDown()
	require.NoError(t, err)
	rollbackTime := time.Since(start)

	t.Logf("Rollback time: %v", rollbackTime)

	// Performance assertions (adjust based on your requirements)
	assert.Less(t, migrationTime, 10*time.Second, "Migration should complete within 10 seconds")
	assert.Less(t, seedingTime, 30*time.Second, "Seeding 1000 users should complete within 30 seconds")
	assert.Less(t, rollbackTime, 10*time.Second, "Rollback should complete within 10 seconds")
}

func seedLargeDataset(t *testing.T, testDB *TestDatabase, userCount int) {
	// Use batch inserts for better performance
	batchSize := 100
	passwordHash := "$2a$10$rOCKx7VDum0oaEFrZQWOa.6nYgs8o8/Oa9Q8Ks8Ks8Ks8Ks8Ks8K"

	for i := 0; i < userCount; i += batchSize {
		end := i + batchSize
		if end > userCount {
			end = userCount
		}

		// Build batch insert
		values := make([]string, 0, end-i)
		args := make([]interface{}, 0, (end-i)*6)

		for j := i; j < end; j++ {
			values = append(values, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)",
				len(args)+1, len(args)+2, len(args)+3, len(args)+4, len(args)+5, len(args)+6))

			args = append(args,
				fmt.Sprintf("user-id-%06d", j),
				fmt.Sprintf("user%d@example.com", j),
				passwordHash,
				fmt.Sprintf("User%d", j),
				fmt.Sprintf("Test%d", j),
				"active")
		}

		sql := fmt.Sprintf(`
			INSERT INTO users (id, email, password, first_name, last_name, status)
			VALUES %s
			ON CONFLICT (email) DO NOTHING
		`, strings.Join(values, ", "))

		testDB.SeedData(sql, args...)
	}
}
