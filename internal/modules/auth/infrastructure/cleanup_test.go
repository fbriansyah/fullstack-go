package infrastructure

import (
	"context"
	"testing"
	"time"

	"go-templ-template/internal/modules/auth/domain"
	"go-templ-template/internal/shared/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type SessionCleanupTestSuite struct {
	suite.Suite
	db      *database.DB
	repo    SessionRepository
	cleanup *SessionCleanupService
	ctx     context.Context
}

func (suite *SessionCleanupTestSuite) SetupSuite() {
	// Skip if no database available
	database.SkipIfNoDatabase(suite.T())

	// Setup test database connection
	testDB := database.NewTestDatabase(suite.T())

	suite.db = testDB.DB
	suite.repo = NewSessionRepository(testDB.DB)
	suite.cleanup = NewSessionCleanupService(suite.repo, 100*time.Millisecond)
	suite.ctx = context.Background()

	// Create test tables
	suite.createTestTables()
}

func (suite *SessionCleanupTestSuite) createTestTables() {
	// Drop existing tables to ensure clean state
	_, _ = suite.db.ExecContext(suite.ctx, "DROP TABLE IF EXISTS sessions")
	_, _ = suite.db.ExecContext(suite.ctx, "DROP TABLE IF EXISTS users")

	// Create users table
	usersSQL := `
		CREATE TABLE users (
			id VARCHAR(255) PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			first_name VARCHAR(100) NOT NULL,
			last_name VARCHAR(100) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended')),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			version INTEGER DEFAULT 1
		)`

	_, err := suite.db.ExecContext(suite.ctx, usersSQL)
	require.NoError(suite.T(), err)

	// Create sessions table
	sessionsSQL := `
		CREATE TABLE sessions (
			id VARCHAR(255) PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			ip_address INET,
			user_agent TEXT,
			is_active BOOLEAN NOT NULL DEFAULT true
		)`

	_, err = suite.db.ExecContext(suite.ctx, sessionsSQL)
	require.NoError(suite.T(), err)
}

func (suite *SessionCleanupTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
}

func (suite *SessionCleanupTestSuite) SetupTest() {
	// Clean up sessions table before each test
	_, err := suite.db.ExecContext(suite.ctx, "DELETE FROM sessions")
	require.NoError(suite.T(), err)

	// Clean up users table before each test
	_, err = suite.db.ExecContext(suite.ctx, "DELETE FROM users")
	require.NoError(suite.T(), err)
}

func (suite *SessionCleanupTestSuite) createTestUser() string {
	userID := "test-user-id"
	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, status)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := suite.db.ExecContext(suite.ctx, query,
		userID, "test@example.com", "hashed_password", "Test", "User", "active")
	require.NoError(suite.T(), err)

	return userID
}

func (suite *SessionCleanupTestSuite) createTestSession(userID string, expired bool, inactive bool) *domain.Session {
	config := domain.SessionConfig{
		DefaultDuration: 24 * time.Hour,
		MaxDuration:     7 * 24 * time.Hour,
		CleanupInterval: time.Hour,
	}

	session, err := domain.NewSession(userID, "192.168.1.1", "Mozilla/5.0", config)
	require.NoError(suite.T(), err)

	if expired {
		session.ExpiresAt = time.Now().Add(-time.Hour)
	}

	if inactive {
		session.Invalidate()
	}

	return session
}

func (suite *SessionCleanupTestSuite) TestCleanupNow() {
	userID := suite.createTestUser()

	// Create active session
	activeSession := suite.createTestSession(userID, false, false)
	err := suite.repo.Create(suite.ctx, activeSession)
	require.NoError(suite.T(), err)

	// Create expired session
	expiredSession := suite.createTestSession(userID, true, false)
	err = suite.repo.Create(suite.ctx, expiredSession)
	require.NoError(suite.T(), err)

	// Create inactive session
	inactiveSession := suite.createTestSession(userID, false, true)
	err = suite.repo.Create(suite.ctx, inactiveSession)
	require.NoError(suite.T(), err)

	// Perform cleanup
	deletedCount, err := suite.cleanup.CleanupNow(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(2), deletedCount) // Should delete expired and inactive

	// Verify only active session remains
	sessions, err := suite.repo.GetByUserID(suite.ctx, userID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), sessions, 1)
	assert.Equal(suite.T(), activeSession.ID, sessions[0].ID)
}

func (suite *SessionCleanupTestSuite) TestPeriodicCleanup() {
	userID := suite.createTestUser()

	// Create expired session
	expiredSession := suite.createTestSession(userID, true, false)
	err := suite.repo.Create(suite.ctx, expiredSession)
	require.NoError(suite.T(), err)

	// Start cleanup service
	ctx, cancel := context.WithTimeout(suite.ctx, 500*time.Millisecond)
	defer cancel()

	suite.cleanup.Start(ctx)

	// Wait for cleanup to run
	time.Sleep(200 * time.Millisecond)

	// Verify session was cleaned up
	sessions, err := suite.repo.GetByUserID(suite.ctx, userID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), sessions, 0)
}

func (suite *SessionCleanupTestSuite) TestStopCleanup() {
	// Start cleanup service
	ctx := context.Background()
	suite.cleanup.Start(ctx)

	// Stop the service
	done := make(chan struct{})
	go func() {
		suite.cleanup.Stop()
		close(done)
	}()

	// Should stop within reasonable time
	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		suite.T().Fatal("Cleanup service did not stop within timeout")
	}
}

func (suite *SessionCleanupTestSuite) TestContextCancellation() {
	userID := suite.createTestUser()

	// Create expired session
	expiredSession := suite.createTestSession(userID, true, false)
	err := suite.repo.Create(suite.ctx, expiredSession)
	require.NoError(suite.T(), err)

	// Start cleanup service with cancellable context
	ctx, cancel := context.WithCancel(suite.ctx)
	suite.cleanup.Start(ctx)

	// Cancel context after short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	// Wait for service to stop
	time.Sleep(200 * time.Millisecond)

	// Service should have stopped gracefully
	// We can't easily test this without exposing internal state,
	// but the test ensures no panic occurs
}

func TestSessionCleanupTestSuite(t *testing.T) {
	suite.Run(t, new(SessionCleanupTestSuite))
}
