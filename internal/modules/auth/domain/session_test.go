package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSession(t *testing.T) {
	config := SessionConfig{
		DefaultDuration: 24 * time.Hour,
		MaxDuration:     7 * 24 * time.Hour,
		CleanupInterval: time.Hour,
	}

	userID := "user-123"
	ipAddress := "192.168.1.1"
	userAgent := "Mozilla/5.0"

	session, err := NewSession(userID, ipAddress, userAgent, config)

	require.NoError(t, err)
	assert.NotEmpty(t, session.ID)
	assert.Equal(t, userID, session.UserID)
	assert.Equal(t, ipAddress, session.IPAddress)
	assert.Equal(t, userAgent, session.UserAgent)
	assert.True(t, session.IsActive)
	assert.False(t, session.IsExpired())
	assert.True(t, session.IsValid())

	// Check that expiration is set correctly
	expectedExpiry := time.Now().Add(config.DefaultDuration)
	assert.WithinDuration(t, expectedExpiry, session.ExpiresAt, time.Second)
}

func TestSession_IsExpired(t *testing.T) {
	config := SessionConfig{DefaultDuration: time.Hour}
	session, err := NewSession("user-123", "192.168.1.1", "Mozilla/5.0", config)
	require.NoError(t, err)

	// Session should not be expired initially
	assert.False(t, session.IsExpired())

	// Manually set expiration to past
	session.ExpiresAt = time.Now().Add(-time.Hour)
	assert.True(t, session.IsExpired())
}

func TestSession_IsValid(t *testing.T) {
	config := SessionConfig{DefaultDuration: time.Hour}
	session, err := NewSession("user-123", "192.168.1.1", "Mozilla/5.0", config)
	require.NoError(t, err)

	// Session should be valid initially
	assert.True(t, session.IsValid())

	// Invalidate session
	session.Invalidate()
	assert.False(t, session.IsValid())

	// Reactivate but expire
	session.IsActive = true
	session.ExpiresAt = time.Now().Add(-time.Hour)
	assert.False(t, session.IsValid())
}

func TestSession_Extend(t *testing.T) {
	config := SessionConfig{DefaultDuration: time.Hour}
	session, err := NewSession("user-123", "192.168.1.1", "Mozilla/5.0", config)
	require.NoError(t, err)

	originalExpiry := session.ExpiresAt
	extension := 2 * time.Hour

	session.Extend(extension)

	// New expiry should be approximately now + extension
	expectedExpiry := time.Now().Add(extension)
	assert.WithinDuration(t, expectedExpiry, session.ExpiresAt, time.Second)
	assert.True(t, session.ExpiresAt.After(originalExpiry))
}

func TestSession_Invalidate(t *testing.T) {
	config := SessionConfig{DefaultDuration: time.Hour}
	session, err := NewSession("user-123", "192.168.1.1", "Mozilla/5.0", config)
	require.NoError(t, err)

	assert.True(t, session.IsActive)

	session.Invalidate()

	assert.False(t, session.IsActive)
	assert.False(t, session.IsValid())
}

func TestSession_ValidateSecurityContext(t *testing.T) {
	config := SessionConfig{DefaultDuration: time.Hour}
	ipAddress := "192.168.1.1"
	userAgent := "Mozilla/5.0"

	session, err := NewSession("user-123", ipAddress, userAgent, config)
	require.NoError(t, err)

	// Same context should validate
	assert.True(t, session.ValidateSecurityContext(ipAddress, userAgent))

	// Different IP should not validate
	assert.False(t, session.ValidateSecurityContext("192.168.1.2", userAgent))

	// Different user agent should not validate
	assert.False(t, session.ValidateSecurityContext(ipAddress, "Chrome/91.0"))
}

func TestGenerateSecureSessionID(t *testing.T) {
	id1, err1 := generateSecureSessionID()
	id2, err2 := generateSecureSessionID()

	require.NoError(t, err1)
	require.NoError(t, err2)

	// IDs should be different
	assert.NotEqual(t, id1, id2)

	// IDs should be 64 characters (32 bytes hex encoded)
	assert.Len(t, id1, 64)
	assert.Len(t, id2, 64)

	// IDs should be valid hex
	assert.Regexp(t, "^[a-f0-9]+$", id1)
	assert.Regexp(t, "^[a-f0-9]+$", id2)
}

func TestSessionConfig(t *testing.T) {
	config := SessionConfig{
		DefaultDuration: 24 * time.Hour,
		MaxDuration:     7 * 24 * time.Hour,
		CleanupInterval: time.Hour,
	}

	assert.Equal(t, 24*time.Hour, config.DefaultDuration)
	assert.Equal(t, 7*24*time.Hour, config.MaxDuration)
	assert.Equal(t, time.Hour, config.CleanupInterval)
}
