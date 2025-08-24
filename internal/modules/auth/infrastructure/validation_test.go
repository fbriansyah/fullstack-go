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

type SessionValidatorTestSuite struct {
	suite.Suite
	db        *database.DB
	repo      SessionRepository
	validator *SessionValidator
	ctx       context.Context
}

func (suite *SessionValidatorTestSuite) SetupSuite() {
	// Skip if no database available
	database.SkipIfNoDatabase(suite.T())

	// Setup test database connection
	testDB := database.NewTestDatabase(suite.T())

	suite.db = testDB.DB
	suite.repo = NewSessionRepository(testDB.DB)

	config := DefaultSessionValidatorConfig()
	suite.validator = NewSessionValidator(suite.repo, config)
	suite.ctx = context.Background()

	// Create test tables
	suite.createTestTables()
}

func (suite *SessionValidatorTestSuite) createTestTables() {
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

func (suite *SessionValidatorTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
}

func (suite *SessionValidatorTestSuite) SetupTest() {
	// Clean up sessions table before each test
	_, err := suite.db.ExecContext(suite.ctx, "DELETE FROM sessions")
	require.NoError(suite.T(), err)

	// Clean up users table before each test
	_, err = suite.db.ExecContext(suite.ctx, "DELETE FROM users")
	require.NoError(suite.T(), err)
}

func (suite *SessionValidatorTestSuite) createTestUser() string {
	userID := "test-user-id"
	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, status)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := suite.db.ExecContext(suite.ctx, query,
		userID, "test@example.com", "hashed_password", "Test", "User", "active")
	require.NoError(suite.T(), err)

	return userID
}

func (suite *SessionValidatorTestSuite) createTestSession(userID string, expired bool, inactive bool) *domain.Session {
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

func (suite *SessionValidatorTestSuite) TestValidateSession_Success() {
	userID := suite.createTestUser()
	session := suite.createTestSession(userID, false, false)

	err := suite.repo.Create(suite.ctx, session)
	require.NoError(suite.T(), err)

	// Test successful validation
	result, err := suite.validator.ValidateSession(suite.ctx, session.ID, "192.168.1.1", "Mozilla/5.0")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), session.ID, result.Session.ID)
	assert.Equal(suite.T(), userID, result.UserID)
	assert.False(suite.T(), result.Extended)
	assert.Empty(suite.T(), result.Warnings)
}

func (suite *SessionValidatorTestSuite) TestValidateSession_SecurityContextMismatch() {
	userID := suite.createTestUser()
	session := suite.createTestSession(userID, false, false)

	err := suite.repo.Create(suite.ctx, session)
	require.NoError(suite.T(), err)

	// Test with different IP and user agent (should generate warnings but not fail)
	result, err := suite.validator.ValidateSession(suite.ctx, session.ID, "192.168.1.2", "Different Agent")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Warnings, 2) // IP and user agent mismatch
}

func (suite *SessionValidatorTestSuite) TestValidateSession_EnforceSecurityContext() {
	userID := suite.createTestUser()
	session := suite.createTestSession(userID, false, false)

	err := suite.repo.Create(suite.ctx, session)
	require.NoError(suite.T(), err)

	// Create validator with security context enforcement
	config := DefaultSessionValidatorConfig()
	config.EnforceSecurityContext = true
	validator := NewSessionValidator(suite.repo, config)

	// Test with different IP (should fail)
	_, err = validator.ValidateSession(suite.ctx, session.ID, "192.168.1.2", "Mozilla/5.0")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrSecurityContextMismatch, err)
}

func (suite *SessionValidatorTestSuite) TestValidateSession_ExpiredSession() {
	userID := suite.createTestUser()
	session := suite.createTestSession(userID, true, false) // expired

	err := suite.repo.Create(suite.ctx, session)
	require.NoError(suite.T(), err)

	// Test expired session
	_, err = suite.validator.ValidateSession(suite.ctx, session.ID, "192.168.1.1", "Mozilla/5.0")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrSessionNotFound, err) // ValidateAndGet returns not found for expired
}

func (suite *SessionValidatorTestSuite) TestValidateSession_InactiveSession() {
	userID := suite.createTestUser()
	session := suite.createTestSession(userID, false, true) // inactive

	err := suite.repo.Create(suite.ctx, session)
	require.NoError(suite.T(), err)

	// Test inactive session
	_, err = suite.validator.ValidateSession(suite.ctx, session.ID, "192.168.1.1", "Mozilla/5.0")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrSessionNotFound, err) // ValidateAndGet returns not found for inactive
}

func (suite *SessionValidatorTestSuite) TestValidateSession_NotFound() {
	// Test non-existent session
	_, err := suite.validator.ValidateSession(suite.ctx, "non-existent", "192.168.1.1", "Mozilla/5.0")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrSessionNotFound, err)

	// Test empty session ID
	_, err = suite.validator.ValidateSession(suite.ctx, "", "192.168.1.1", "Mozilla/5.0")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrSessionNotFound, err)
}

func (suite *SessionValidatorTestSuite) TestCreateSession_Success() {
	userID := suite.createTestUser()

	config := domain.SessionConfig{
		DefaultDuration: 24 * time.Hour,
		MaxDuration:     7 * 24 * time.Hour,
		CleanupInterval: time.Hour,
	}

	session, err := suite.validator.CreateSession(suite.ctx, userID, "192.168.1.1", "Mozilla/5.0", config)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), session)
	assert.Equal(suite.T(), userID, session.UserID)
	assert.Equal(suite.T(), "192.168.1.1", session.IPAddress)
	assert.Equal(suite.T(), "Mozilla/5.0", session.UserAgent)
	assert.True(suite.T(), session.IsActive)
}

func (suite *SessionValidatorTestSuite) TestCreateSession_SessionLimit() {
	userID := suite.createTestUser()

	// Create validator with low session limit
	config := DefaultSessionValidatorConfig()
	config.MaxSessionsPerUser = 2
	validator := NewSessionValidator(suite.repo, config)

	sessionConfig := domain.SessionConfig{
		DefaultDuration: 24 * time.Hour,
		MaxDuration:     7 * 24 * time.Hour,
		CleanupInterval: time.Hour,
	}

	// Create sessions up to limit
	session1, err := validator.CreateSession(suite.ctx, userID, "192.168.1.1", "Mozilla/5.0", sessionConfig)
	assert.NoError(suite.T(), err)

	session2, err := validator.CreateSession(suite.ctx, userID, "192.168.1.2", "Mozilla/5.0", sessionConfig)
	assert.NoError(suite.T(), err)

	// Create one more session (should cleanup oldest)
	session3, err := validator.CreateSession(suite.ctx, userID, "192.168.1.3", "Mozilla/5.0", sessionConfig)
	assert.NoError(suite.T(), err)

	// Verify we still have only 2 sessions and the oldest was removed
	sessions, err := suite.repo.GetByUserID(suite.ctx, userID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), sessions, 2)

	// Verify session1 was removed (oldest)
	_, err = suite.repo.GetByID(suite.ctx, session1.ID)
	assert.Error(suite.T(), err)

	// Verify session2 and session3 still exist
	_, err = suite.repo.GetByID(suite.ctx, session2.ID)
	assert.NoError(suite.T(), err)

	_, err = suite.repo.GetByID(suite.ctx, session3.ID)
	assert.NoError(suite.T(), err)
}

func (suite *SessionValidatorTestSuite) TestInvalidateSession() {
	userID := suite.createTestUser()
	session := suite.createTestSession(userID, false, false)

	err := suite.repo.Create(suite.ctx, session)
	require.NoError(suite.T(), err)

	// Invalidate session
	err = suite.validator.InvalidateSession(suite.ctx, session.ID)
	assert.NoError(suite.T(), err)

	// Verify session is invalidated
	retrieved, err := suite.repo.GetByID(suite.ctx, session.ID)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), retrieved.IsActive)
}

func (suite *SessionValidatorTestSuite) TestInvalidateAllUserSessions() {
	userID := suite.createTestUser()

	// Create multiple sessions
	session1 := suite.createTestSession(userID, false, false)
	session2 := suite.createTestSession(userID, false, false)

	err := suite.repo.Create(suite.ctx, session1)
	require.NoError(suite.T(), err)

	err = suite.repo.Create(suite.ctx, session2)
	require.NoError(suite.T(), err)

	// Invalidate all user sessions
	err = suite.validator.InvalidateAllUserSessions(suite.ctx, userID)
	assert.NoError(suite.T(), err)

	// Verify all sessions are deleted
	sessions, err := suite.repo.GetByUserID(suite.ctx, userID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), sessions, 0)
}

func (suite *SessionValidatorTestSuite) TestGetUserSessions() {
	userID := suite.createTestUser()

	// Create multiple sessions
	session1 := suite.createTestSession(userID, false, false)
	session2 := suite.createTestSession(userID, false, false)

	err := suite.repo.Create(suite.ctx, session1)
	require.NoError(suite.T(), err)

	err = suite.repo.Create(suite.ctx, session2)
	require.NoError(suite.T(), err)

	// Get user sessions
	sessions, err := suite.validator.GetUserSessions(suite.ctx, userID)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), sessions, 2)
}

func TestSessionValidatorTestSuite(t *testing.T) {
	suite.Run(t, new(SessionValidatorTestSuite))
}
