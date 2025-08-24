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

type SessionRepositoryTestSuite struct {
	suite.Suite
	db   *database.DB
	repo SessionRepository
	ctx  context.Context
}

func (suite *SessionRepositoryTestSuite) SetupSuite() {
	// Skip if no database available
	database.SkipIfNoDatabase(suite.T())

	// Setup test database connection
	testDB := database.NewTestDatabase(suite.T())

	suite.db = testDB.DB
	suite.repo = NewSessionRepository(testDB.DB)
	suite.ctx = context.Background()

	// Create test tables
	suite.createTestTables()
}

func (suite *SessionRepositoryTestSuite) createTestTables() {
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

func (suite *SessionRepositoryTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
}

func (suite *SessionRepositoryTestSuite) SetupTest() {
	// Clean up sessions table before each test
	_, err := suite.db.ExecContext(suite.ctx, "DELETE FROM sessions")
	require.NoError(suite.T(), err)

	// Clean up users table before each test
	_, err = suite.db.ExecContext(suite.ctx, "DELETE FROM users")
	require.NoError(suite.T(), err)
}

func (suite *SessionRepositoryTestSuite) createTestUser() string {
	userID := "test-user-id"
	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, status)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := suite.db.ExecContext(suite.ctx, query,
		userID, "test@example.com", "hashed_password", "Test", "User", "active")
	require.NoError(suite.T(), err)

	return userID
}

func (suite *SessionRepositoryTestSuite) createTestSession(userID string) *domain.Session {
	config := domain.SessionConfig{
		DefaultDuration: 24 * time.Hour,
		MaxDuration:     7 * 24 * time.Hour,
		CleanupInterval: time.Hour,
	}

	session, err := domain.NewSession(userID, "192.168.1.1", "Mozilla/5.0", config)
	require.NoError(suite.T(), err)

	return session
}

func (suite *SessionRepositoryTestSuite) TestCreate() {
	userID := suite.createTestUser()
	session := suite.createTestSession(userID)

	// Test successful creation
	err := suite.repo.Create(suite.ctx, session)
	assert.NoError(suite.T(), err)

	// Verify session was created
	retrieved, err := suite.repo.GetByID(suite.ctx, session.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), session.ID, retrieved.ID)
	assert.Equal(suite.T(), session.UserID, retrieved.UserID)
	assert.Equal(suite.T(), session.IPAddress, retrieved.IPAddress)
	assert.Equal(suite.T(), session.UserAgent, retrieved.UserAgent)
	assert.True(suite.T(), retrieved.IsActive)
}

func (suite *SessionRepositoryTestSuite) TestGetByID() {
	userID := suite.createTestUser()
	session := suite.createTestSession(userID)

	// Create session first
	err := suite.repo.Create(suite.ctx, session)
	require.NoError(suite.T(), err)

	// Test successful retrieval
	retrieved, err := suite.repo.GetByID(suite.ctx, session.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), session.ID, retrieved.ID)
	assert.Equal(suite.T(), session.UserID, retrieved.UserID)

	// Test non-existent session
	_, err = suite.repo.GetByID(suite.ctx, "non-existent-id")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), database.ErrNotFound, err)
}

func (suite *SessionRepositoryTestSuite) TestGetByUserID() {
	userID := suite.createTestUser()

	// Create multiple sessions for the user
	session1 := suite.createTestSession(userID)
	session2 := suite.createTestSession(userID)

	err := suite.repo.Create(suite.ctx, session1)
	require.NoError(suite.T(), err)

	err = suite.repo.Create(suite.ctx, session2)
	require.NoError(suite.T(), err)

	// Create an inactive session
	session3 := suite.createTestSession(userID)
	session3.Invalidate()
	err = suite.repo.Create(suite.ctx, session3)
	require.NoError(suite.T(), err)

	// Test retrieval - should only return active sessions
	sessions, err := suite.repo.GetByUserID(suite.ctx, userID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), sessions, 2)

	// Verify all returned sessions are active
	for _, s := range sessions {
		assert.True(suite.T(), s.IsActive)
		assert.Equal(suite.T(), userID, s.UserID)
	}
}

func (suite *SessionRepositoryTestSuite) TestUpdate() {
	userID := suite.createTestUser()
	session := suite.createTestSession(userID)

	// Create session first
	err := suite.repo.Create(suite.ctx, session)
	require.NoError(suite.T(), err)

	// Update session
	session.IPAddress = "192.168.1.2"
	session.UserAgent = "Updated User Agent"
	session.Invalidate()

	err = suite.repo.Update(suite.ctx, session)
	assert.NoError(suite.T(), err)

	// Verify update
	retrieved, err := suite.repo.GetByID(suite.ctx, session.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "192.168.1.2", retrieved.IPAddress)
	assert.Equal(suite.T(), "Updated User Agent", retrieved.UserAgent)
	assert.False(suite.T(), retrieved.IsActive)
}

func (suite *SessionRepositoryTestSuite) TestDelete() {
	userID := suite.createTestUser()
	session := suite.createTestSession(userID)

	// Create session first
	err := suite.repo.Create(suite.ctx, session)
	require.NoError(suite.T(), err)

	// Delete session
	err = suite.repo.Delete(suite.ctx, session.ID)
	assert.NoError(suite.T(), err)

	// Verify deletion
	_, err = suite.repo.GetByID(suite.ctx, session.ID)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), database.ErrNotFound, err)
}

func (suite *SessionRepositoryTestSuite) TestDeleteByUserID() {
	userID := suite.createTestUser()

	// Create multiple sessions for the user
	session1 := suite.createTestSession(userID)
	session2 := suite.createTestSession(userID)

	err := suite.repo.Create(suite.ctx, session1)
	require.NoError(suite.T(), err)

	err = suite.repo.Create(suite.ctx, session2)
	require.NoError(suite.T(), err)

	// Delete all sessions for user
	err = suite.repo.DeleteByUserID(suite.ctx, userID)
	assert.NoError(suite.T(), err)

	// Verify all sessions are deleted
	sessions, err := suite.repo.GetByUserID(suite.ctx, userID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), sessions, 0)
}

func (suite *SessionRepositoryTestSuite) TestCleanupExpired() {
	userID := suite.createTestUser()

	// Create active session
	activeSession := suite.createTestSession(userID)
	err := suite.repo.Create(suite.ctx, activeSession)
	require.NoError(suite.T(), err)

	// Create expired session
	expiredSession := suite.createTestSession(userID)
	expiredSession.ExpiresAt = time.Now().Add(-time.Hour) // Expired 1 hour ago
	err = suite.repo.Create(suite.ctx, expiredSession)
	require.NoError(suite.T(), err)

	// Create inactive session
	inactiveSession := suite.createTestSession(userID)
	inactiveSession.Invalidate()
	err = suite.repo.Create(suite.ctx, inactiveSession)
	require.NoError(suite.T(), err)

	// Cleanup expired sessions
	deletedCount, err := suite.repo.CleanupExpired(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(2), deletedCount) // Should delete expired and inactive

	// Verify only active session remains
	sessions, err := suite.repo.GetByUserID(suite.ctx, userID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), sessions, 1)
	assert.Equal(suite.T(), activeSession.ID, sessions[0].ID)
}

func (suite *SessionRepositoryTestSuite) TestValidateAndGet() {
	userID := suite.createTestUser()

	// Create active session
	activeSession := suite.createTestSession(userID)
	err := suite.repo.Create(suite.ctx, activeSession)
	require.NoError(suite.T(), err)

	// Create expired session
	expiredSession := suite.createTestSession(userID)
	expiredSession.ExpiresAt = time.Now().Add(-time.Hour)
	err = suite.repo.Create(suite.ctx, expiredSession)
	require.NoError(suite.T(), err)

	// Create inactive session
	inactiveSession := suite.createTestSession(userID)
	inactiveSession.Invalidate()
	err = suite.repo.Create(suite.ctx, inactiveSession)
	require.NoError(suite.T(), err)

	// Test valid session
	retrieved, err := suite.repo.ValidateAndGet(suite.ctx, activeSession.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), activeSession.ID, retrieved.ID)

	// Test expired session
	_, err = suite.repo.ValidateAndGet(suite.ctx, expiredSession.ID)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), database.ErrNotFound, err)

	// Test inactive session
	_, err = suite.repo.ValidateAndGet(suite.ctx, inactiveSession.ID)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), database.ErrNotFound, err)
}

func (suite *SessionRepositoryTestSuite) TestExtendSession() {
	userID := suite.createTestUser()
	session := suite.createTestSession(userID)

	// Create session
	err := suite.repo.Create(suite.ctx, session)
	require.NoError(suite.T(), err)

	// Extend session
	extension := 2 * time.Hour
	err = suite.repo.ExtendSession(suite.ctx, session.ID, extension)
	assert.NoError(suite.T(), err)

	// Verify extension (approximately - allowing for small time differences)
	retrieved, err := suite.repo.GetByID(suite.ctx, session.ID)
	assert.NoError(suite.T(), err)

	expectedExpiry := time.Now().Add(extension)
	timeDiff := retrieved.ExpiresAt.Sub(expectedExpiry)
	assert.True(suite.T(), timeDiff < time.Minute && timeDiff > -time.Minute,
		"Session expiry should be approximately %v, got %v", expectedExpiry, retrieved.ExpiresAt)

	// Test extending non-existent session
	err = suite.repo.ExtendSession(suite.ctx, "non-existent", extension)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), database.ErrNotFound, err)
}

func (suite *SessionRepositoryTestSuite) TestInvalidateSession() {
	userID := suite.createTestUser()
	session := suite.createTestSession(userID)

	// Create session
	err := suite.repo.Create(suite.ctx, session)
	require.NoError(suite.T(), err)

	// Verify session is active
	retrieved, err := suite.repo.GetByID(suite.ctx, session.ID)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), retrieved.IsActive)

	// Invalidate session
	err = suite.repo.InvalidateSession(suite.ctx, session.ID)
	assert.NoError(suite.T(), err)

	// Verify session is inactive
	retrieved, err = suite.repo.GetByID(suite.ctx, session.ID)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), retrieved.IsActive)

	// Test invalidating non-existent session
	err = suite.repo.InvalidateSession(suite.ctx, "non-existent")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), database.ErrNotFound, err)
}

func (suite *SessionRepositoryTestSuite) TestCountActiveSessions() {
	userID := suite.createTestUser()

	// Initially no sessions
	count, err := suite.repo.CountActiveSessions(suite.ctx, userID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(0), count)

	// Create active sessions
	session1 := suite.createTestSession(userID)
	session2 := suite.createTestSession(userID)

	err = suite.repo.Create(suite.ctx, session1)
	require.NoError(suite.T(), err)

	err = suite.repo.Create(suite.ctx, session2)
	require.NoError(suite.T(), err)

	// Create inactive session
	inactiveSession := suite.createTestSession(userID)
	inactiveSession.Invalidate()
	err = suite.repo.Create(suite.ctx, inactiveSession)
	require.NoError(suite.T(), err)

	// Count should be 2 (only active sessions)
	count, err = suite.repo.CountActiveSessions(suite.ctx, userID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(2), count)
}

func (suite *SessionRepositoryTestSuite) TestTransactionSupport() {
	userID := suite.createTestUser()
	session := suite.createTestSession(userID)

	// Test transaction rollback
	err := database.ExecuteInTransaction(suite.ctx, suite.db, func(ctx context.Context) error {
		// Create session within transaction
		err := suite.repo.Create(ctx, session)
		if err != nil {
			return err
		}

		// Verify session exists within transaction
		_, err = suite.repo.GetByID(ctx, session.ID)
		if err != nil {
			return err
		}

		// Force rollback
		return assert.AnError
	})

	assert.Error(suite.T(), err)

	// Verify session was not created due to rollback
	_, err = suite.repo.GetByID(suite.ctx, session.ID)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), database.ErrNotFound, err)

	// Test transaction commit
	err = database.ExecuteInTransaction(suite.ctx, suite.db, func(ctx context.Context) error {
		return suite.repo.Create(ctx, session)
	})

	assert.NoError(suite.T(), err)

	// Verify session was created
	retrieved, err := suite.repo.GetByID(suite.ctx, session.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), session.ID, retrieved.ID)
}

func TestSessionRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(SessionRepositoryTestSuite))
}
