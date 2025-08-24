package domain

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// Session represents a user authentication session
type Session struct {
	ID        string    `db:"id" json:"id"`
	UserID    string    `db:"user_id" json:"user_id"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	IPAddress string    `db:"ip_address" json:"ip_address"`
	UserAgent string    `db:"user_agent" json:"user_agent"`
	IsActive  bool      `db:"is_active" json:"is_active"`
}

// SessionConfig holds configuration for session management
type SessionConfig struct {
	DefaultDuration time.Duration
	MaxDuration     time.Duration
	CleanupInterval time.Duration
}

// NewSession creates a new session with security features
func NewSession(userID, ipAddress, userAgent string, config SessionConfig) (*Session, error) {
	sessionID, err := generateSecureSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	now := time.Now()
	expiresAt := now.Add(config.DefaultDuration)

	return &Session{
		ID:        sessionID,
		UserID:    userID,
		ExpiresAt: expiresAt,
		CreatedAt: now,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		IsActive:  true,
	}, nil
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValid checks if the session is valid (active and not expired)
func (s *Session) IsValid() bool {
	return s.IsActive && !s.IsExpired()
}

// Extend extends the session expiration time
func (s *Session) Extend(duration time.Duration) {
	s.ExpiresAt = time.Now().Add(duration)
}

// Invalidate marks the session as inactive
func (s *Session) Invalidate() {
	s.IsActive = false
}

// ValidateSecurityContext checks if the session matches the current security context
func (s *Session) ValidateSecurityContext(ipAddress, userAgent string) bool {
	// For enhanced security, we could enforce IP and user agent matching
	// For now, we'll just log mismatches but allow the session
	return s.IPAddress == ipAddress && s.UserAgent == userAgent
}

// generateSecureSessionID generates a cryptographically secure session ID
func generateSecureSessionID() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
