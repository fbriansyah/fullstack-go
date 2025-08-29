package database

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultTestDatabaseConfig_Timeout(t *testing.T) {
	// Test default timeout
	cfg := DefaultTestDatabaseConfig()
	assert.Equal(t, 30*time.Second, cfg.ConnectTimeout)

	// Test custom timeout via environment variable
	os.Setenv("TEST_DB_TIMEOUT", "10s")
	defer os.Unsetenv("TEST_DB_TIMEOUT")

	cfg = DefaultTestDatabaseConfig()
	assert.Equal(t, 10*time.Second, cfg.ConnectTimeout)

	// Test invalid timeout falls back to default
	os.Setenv("TEST_DB_TIMEOUT", "invalid")
	cfg = DefaultTestDatabaseConfig()
	assert.Equal(t, 30*time.Second, cfg.ConnectTimeout)
}

func TestSkipIfNoDatabaseWithTimeout(t *testing.T) {
	// This test will skip if no database is available
	// We use a very short timeout to test the timeout functionality
	SkipIfNoDatabaseWithTimeout(t, 1*time.Millisecond)

	// If we reach here, database is available
	t.Log("Database is available for testing")
}

func TestNewTestDatabaseWithTimeout(t *testing.T) {
	// Skip if no database available
	SkipIfNoDatabase(t)

	// Test creating database with custom timeout
	testDB := NewTestDatabaseWithTimeout(t, 5*time.Second)
	defer testDB.Close()

	assert.NotNil(t, testDB)
	assert.Equal(t, 5*time.Second, testDB.Config.ConnectTimeout)

	// Verify connection works
	ctx := context.Background()
	err := testDB.DB.HealthCheck(ctx)
	require.NoError(t, err)
}

func TestTryConnectWithTimeout(t *testing.T) {
	// Test with very short timeout - should fail if database takes time to connect
	testDB := TryConnectWithTimeout(1 * time.Millisecond)
	if testDB != nil {
		testDB.Close()
		t.Log("Database connected within 1ms")
	} else {
		t.Log("Database connection timed out within 1ms (expected)")
	}

	// Test with reasonable timeout
	testDB = TryConnectWithTimeout(30 * time.Second)
	if testDB != nil {
		defer testDB.Close()
		t.Log("Database connected within 30s")
	} else {
		t.Log("Database not available within 30s")
	}
}

func TestIsTestDatabaseAvailable(t *testing.T) {
	// Test with very short timeout
	available := IsTestDatabaseAvailable(1 * time.Millisecond)
	t.Logf("Database available within 1ms: %v", available)

	// Test with reasonable timeout
	available = IsTestDatabaseAvailable(30 * time.Second)
	t.Logf("Database available within 30s: %v", available)
}

func TestNewTestSuiteWithTimeout(t *testing.T) {
	// Skip if no database available
	SkipIfNoDatabase(t)

	// Test creating test suite with custom timeout
	suite := NewTestSuiteWithTimeout(t, 10*time.Second)
	defer suite.Teardown()

	assert.NotNil(t, suite)
	assert.Equal(t, 10*time.Second, suite.DB.Config.ConnectTimeout)

	// Setup should work with the configured timeout
	suite.Setup()
}
