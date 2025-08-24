package infrastructure

import (
	"context"
	"errors"
	"time"

	"go-templ-template/internal/modules/auth/domain"
	"go-templ-template/internal/shared/database"
)

var (
	// ErrSessionExpired indicates the session has expired
	ErrSessionExpired = errors.New("session has expired")

	// ErrSessionInactive indicates the session is inactive
	ErrSessionInactive = errors.New("session is inactive")

	// ErrSessionNotFound indicates the session was not found
	ErrSessionNotFound = errors.New("session not found")

	// ErrSecurityContextMismatch indicates security context validation failed
	ErrSecurityContextMismatch = errors.New("security context mismatch")
)

// SessionValidator provides session validation functionality
type SessionValidator struct {
	repo                    SessionRepository
	enforceSecurityContext  bool
	maxSessionsPerUser      int
	sessionExtensionEnabled bool
	extensionDuration       time.Duration
}

// SessionValidatorConfig holds configuration for session validation
type SessionValidatorConfig struct {
	EnforceSecurityContext  bool
	MaxSessionsPerUser      int
	SessionExtensionEnabled bool
	ExtensionDuration       time.Duration
}

// DefaultSessionValidatorConfig returns default configuration
func DefaultSessionValidatorConfig() SessionValidatorConfig {
	return SessionValidatorConfig{
		EnforceSecurityContext:  false, // Disabled by default for flexibility
		MaxSessionsPerUser:      10,    // Allow up to 10 concurrent sessions
		SessionExtensionEnabled: true,
		ExtensionDuration:       30 * time.Minute,
	}
}

// NewSessionValidator creates a new session validator
func NewSessionValidator(repo SessionRepository, config SessionValidatorConfig) *SessionValidator {
	return &SessionValidator{
		repo:                    repo,
		enforceSecurityContext:  config.EnforceSecurityContext,
		maxSessionsPerUser:      config.MaxSessionsPerUser,
		sessionExtensionEnabled: config.SessionExtensionEnabled,
		extensionDuration:       config.ExtensionDuration,
	}
}

// ValidateSessionResult holds the result of session validation
type ValidateSessionResult struct {
	Session  *domain.Session
	UserID   string
	Extended bool
	Warnings []string
}

// ValidateSession validates a session and optionally extends it
func (v *SessionValidator) ValidateSession(ctx context.Context, sessionID, ipAddress, userAgent string) (*ValidateSessionResult, error) {
	if sessionID == "" {
		return nil, ErrSessionNotFound
	}

	// Retrieve and validate session
	session, err := v.repo.ValidateAndGet(ctx, sessionID)
	if err != nil {
		if err == database.ErrNotFound {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	result := &ValidateSessionResult{
		Session:  session,
		UserID:   session.UserID,
		Extended: false,
		Warnings: make([]string, 0),
	}

	// Check if session is active
	if !session.IsActive {
		return nil, ErrSessionInactive
	}

	// Check if session is expired
	if session.IsExpired() {
		return nil, ErrSessionExpired
	}

	// Validate security context if enabled
	if v.enforceSecurityContext {
		if !session.ValidateSecurityContext(ipAddress, userAgent) {
			return nil, ErrSecurityContextMismatch
		}
	} else {
		// Add warnings for security context mismatches
		if session.IPAddress != ipAddress {
			result.Warnings = append(result.Warnings, "IP address mismatch detected")
		}
		if session.UserAgent != userAgent {
			result.Warnings = append(result.Warnings, "User agent mismatch detected")
		}
	}

	// Extend session if enabled and close to expiry
	if v.sessionExtensionEnabled && v.shouldExtendSession(session) {
		err = v.repo.ExtendSession(ctx, sessionID, v.extensionDuration)
		if err != nil {
			// Log error but don't fail validation
			result.Warnings = append(result.Warnings, "Failed to extend session")
		} else {
			result.Extended = true
		}
	}

	return result, nil
}

// CreateSession creates a new session with validation
func (v *SessionValidator) CreateSession(ctx context.Context, userID, ipAddress, userAgent string, config domain.SessionConfig) (*domain.Session, error) {
	// Check session limits
	if v.maxSessionsPerUser > 0 {
		activeCount, err := v.repo.CountActiveSessions(ctx, userID)
		if err != nil {
			return nil, err
		}

		if activeCount >= int64(v.maxSessionsPerUser) {
			// Clean up oldest sessions to make room
			err = v.cleanupOldestSessions(ctx, userID, int(activeCount)-v.maxSessionsPerUser+1)
			if err != nil {
				return nil, err
			}
		}
	}

	// Create new session
	session, err := domain.NewSession(userID, ipAddress, userAgent, config)
	if err != nil {
		return nil, err
	}

	// Save session
	err = v.repo.Create(ctx, session)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// InvalidateSession invalidates a session
func (v *SessionValidator) InvalidateSession(ctx context.Context, sessionID string) error {
	return v.repo.InvalidateSession(ctx, sessionID)
}

// InvalidateAllUserSessions invalidates all sessions for a user
func (v *SessionValidator) InvalidateAllUserSessions(ctx context.Context, userID string) error {
	return v.repo.DeleteByUserID(ctx, userID)
}

// GetUserSessions returns all active sessions for a user
func (v *SessionValidator) GetUserSessions(ctx context.Context, userID string) ([]*domain.Session, error) {
	return v.repo.GetByUserID(ctx, userID)
}

// shouldExtendSession determines if a session should be extended
func (v *SessionValidator) shouldExtendSession(session *domain.Session) bool {
	// Extend if session expires within the next 15 minutes
	threshold := time.Now().Add(15 * time.Minute)
	return session.ExpiresAt.Before(threshold)
}

// cleanupOldestSessions removes the oldest sessions for a user
func (v *SessionValidator) cleanupOldestSessions(ctx context.Context, userID string, count int) error {
	// Get the oldest sessions for the user
	sessions, err := v.repo.GetOldestSessionsByUser(ctx, userID, count)
	if err != nil {
		return err
	}

	// Delete the oldest sessions
	for _, session := range sessions {
		err = v.repo.Delete(ctx, session.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

// SessionStats provides statistics about sessions
type SessionStats struct {
	TotalActiveSessions int64
	UserSessionCount    map[string]int64
	ExpiringSoon        int64 // Sessions expiring within 1 hour
}

// GetSessionStats returns session statistics
func (v *SessionValidator) GetSessionStats(ctx context.Context) (*SessionStats, error) {
	// This would require additional repository methods for comprehensive stats
	// For now, we'll provide a basic implementation
	stats := &SessionStats{
		UserSessionCount: make(map[string]int64),
	}

	// Note: This is a simplified implementation
	// In a real application, you'd want more efficient queries

	return stats, nil
}
