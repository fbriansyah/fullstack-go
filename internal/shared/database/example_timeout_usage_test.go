package database

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ExampleUsage demonstrates how to use the enhanced timeout functionality
func TestExampleUsage(t *testing.T) {
	// Example 1: Skip test if database not available within 5 seconds
	SkipIfNoDatabaseWithTimeout(t, 5*time.Second)

	// Example 2: Create test database with custom timeout
	testDB := NewTestDatabaseWithTimeout(t, 10*time.Second)
	defer testDB.Close()

	// Example 3: Use test suite with custom timeout
	suite := NewTestSuiteWithTimeout(t, 15*time.Second)
	defer suite.Teardown()

	suite.Setup()

	// Example 4: Check if database is available before running expensive setup
	if !IsTestDatabaseAvailable(2 * time.Second) {
		t.Skip("Database not available for integration test")
	}

	// Example 5: Try to connect conditionally
	conditionalDB := TryConnectWithTimeout(3 * time.Second)
	if conditionalDB != nil {
		defer conditionalDB.Close()
		t.Log("Running with real database")

		// Perform database operations
		ctx := context.Background()
		err := conditionalDB.DB.HealthCheck(ctx)
		require.NoError(t, err)
	} else {
		t.Log("Running with mock database")
		// Use mock implementations instead
	}
}

// ExampleEnvironmentConfiguration demonstrates environment variable usage
func TestExampleEnvironmentConfiguration(t *testing.T) {
	// The following environment variables can be set to configure test database:
	// TEST_DB_HOST=localhost
	// TEST_DB_PORT=5432
	// TEST_DB_USER=postgres
	// TEST_DB_PASSWORD=postgres
	// TEST_DB_NAME=test_db
	// TEST_DB_TIMEOUT=30s

	cfg := DefaultTestDatabaseConfig()

	// Verify configuration includes timeout
	assert.NotZero(t, cfg.ConnectTimeout)
	assert.NotEmpty(t, cfg.Host)
	assert.NotEmpty(t, cfg.Port)

	t.Logf("Test database config: Host=%s, Port=%s, Timeout=%v",
		cfg.Host, cfg.Port, cfg.ConnectTimeout)
}
