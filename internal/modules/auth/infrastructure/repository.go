package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go-templ-template/internal/modules/auth/domain"
	"go-templ-template/internal/shared/database"
)

// SessionRepository defines the interface for session data access
type SessionRepository interface {
	// Create inserts a new session
	Create(ctx context.Context, session *domain.Session) error

	// GetByID retrieves a session by its ID
	GetByID(ctx context.Context, id string) (*domain.Session, error)

	// GetByUserID retrieves all active sessions for a user
	GetByUserID(ctx context.Context, userID string) ([]*domain.Session, error)

	// Update updates an existing session
	Update(ctx context.Context, session *domain.Session) error

	// Delete removes a session by its ID
	Delete(ctx context.Context, id string) error

	// DeleteByUserID removes all sessions for a user
	DeleteByUserID(ctx context.Context, userID string) error

	// CleanupExpired removes all expired sessions
	CleanupExpired(ctx context.Context) (int64, error)

	// ValidateAndGet retrieves a session and validates it's active and not expired
	ValidateAndGet(ctx context.Context, sessionID string) (*domain.Session, error)

	// ExtendSession extends the expiration time of a session
	ExtendSession(ctx context.Context, sessionID string, duration time.Duration) error

	// InvalidateSession marks a session as inactive
	InvalidateSession(ctx context.Context, sessionID string) error

	// CountActiveSessions returns the number of active sessions for a user
	CountActiveSessions(ctx context.Context, userID string) (int64, error)

	// GetOldestSessionsByUser retrieves the oldest sessions for a user (for cleanup)
	GetOldestSessionsByUser(ctx context.Context, userID string, limit int) ([]*domain.Session, error)
}

// sessionRepository implements SessionRepository using SQLx
type sessionRepository struct {
	*database.BaseRepository[domain.Session, string]
	db *database.DB
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *database.DB) SessionRepository {
	baseRepo := database.NewBaseRepository[domain.Session, string](db, "sessions", "id")
	return &sessionRepository{
		BaseRepository: baseRepo,
		db:             db,
	}
}

// Create inserts a new session
func (r *sessionRepository) Create(ctx context.Context, session *domain.Session) error {
	query := `
		INSERT INTO sessions (id, user_id, expires_at, created_at, ip_address, user_agent, is_active)
		VALUES (:id, :user_id, :expires_at, :created_at, :ip_address, :user_agent, :is_active)`

	return r.BaseRepository.Create(ctx, session, query)
}

// GetByID retrieves a session by its ID
func (r *sessionRepository) GetByID(ctx context.Context, id string) (*domain.Session, error) {
	query := `
		SELECT id, user_id, expires_at, created_at, ip_address, user_agent, is_active
		FROM sessions 
		WHERE id = $1`

	return r.BaseRepository.GetByID(ctx, id, query)
}

// GetByUserID retrieves all active sessions for a user
func (r *sessionRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Session, error) {
	query := `
		SELECT id, user_id, expires_at, created_at, ip_address, user_agent, is_active
		FROM sessions 
		WHERE user_id = $1 AND is_active = true AND expires_at > NOW()
		ORDER BY created_at DESC`

	var sessions []*domain.Session
	tx := database.GetTxFromContext(ctx)
	if tx != nil {
		err := tx.SelectContext(ctx, &sessions, query, userID)
		return sessions, err
	}

	err := r.db.SelectContext(ctx, &sessions, query, userID)
	return sessions, err
}

// Update updates an existing session
func (r *sessionRepository) Update(ctx context.Context, session *domain.Session) error {
	query := `
		UPDATE sessions 
		SET expires_at = :expires_at, ip_address = :ip_address, user_agent = :user_agent, is_active = :is_active
		WHERE id = :id`

	return r.BaseRepository.Update(ctx, session, query)
}

// Delete removes a session by its ID
func (r *sessionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM sessions WHERE id = $1`
	return r.BaseRepository.Delete(ctx, id, query)
}

// DeleteByUserID removes all sessions for a user
func (r *sessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	query := `DELETE FROM sessions WHERE user_id = $1`

	tx := database.GetTxFromContext(ctx)
	if tx != nil {
		_, err := tx.ExecContext(ctx, query, userID)
		return err
	}

	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// CleanupExpired removes all expired sessions
func (r *sessionRepository) CleanupExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM sessions WHERE expires_at < NOW() OR is_active = false`

	tx := database.GetTxFromContext(ctx)
	if tx != nil {
		result, err := tx.ExecContext(ctx, query)
		if err != nil {
			return 0, err
		}
		return result.RowsAffected()
	}

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// ValidateAndGet retrieves a session and validates it's active and not expired
func (r *sessionRepository) ValidateAndGet(ctx context.Context, sessionID string) (*domain.Session, error) {
	query := `
		SELECT id, user_id, expires_at, created_at, ip_address, user_agent, is_active
		FROM sessions 
		WHERE id = $1 AND is_active = true AND expires_at > NOW()`

	var session domain.Session
	tx := database.GetTxFromContext(ctx)
	if tx != nil {
		err := tx.GetContext(ctx, &session, query, sessionID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, database.ErrNotFound
			}
			return nil, err
		}
		return &session, nil
	}

	err := r.db.GetContext(ctx, &session, query, sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, database.ErrNotFound
		}
		return nil, err
	}
	return &session, nil
}

// ExtendSession extends the expiration time of a session
func (r *sessionRepository) ExtendSession(ctx context.Context, sessionID string, duration time.Duration) error {
	query := `
		UPDATE sessions 
		SET expires_at = NOW() + INTERVAL '%d seconds'
		WHERE id = $1 AND is_active = true AND expires_at > NOW()`

	formattedQuery := fmt.Sprintf(query, int(duration.Seconds()))

	tx := database.GetTxFromContext(ctx)
	if tx != nil {
		result, err := tx.ExecContext(ctx, formattedQuery, sessionID)
		if err != nil {
			return err
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rowsAffected == 0 {
			return database.ErrNotFound
		}
		return nil
	}

	result, err := r.db.ExecContext(ctx, formattedQuery, sessionID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return database.ErrNotFound
	}
	return nil
}

// InvalidateSession marks a session as inactive
func (r *sessionRepository) InvalidateSession(ctx context.Context, sessionID string) error {
	query := `UPDATE sessions SET is_active = false WHERE id = $1`

	tx := database.GetTxFromContext(ctx)
	if tx != nil {
		result, err := tx.ExecContext(ctx, query, sessionID)
		if err != nil {
			return err
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rowsAffected == 0 {
			return database.ErrNotFound
		}
		return nil
	}

	result, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return database.ErrNotFound
	}
	return nil
}

// CountActiveSessions returns the number of active sessions for a user
func (r *sessionRepository) CountActiveSessions(ctx context.Context, userID string) (int64, error) {
	query := `
		SELECT COUNT(*) 
		FROM sessions 
		WHERE user_id = $1 AND is_active = true AND expires_at > NOW()`

	var count int64
	tx := database.GetTxFromContext(ctx)
	if tx != nil {
		err := tx.GetContext(ctx, &count, query, userID)
		return count, err
	}

	err := r.db.GetContext(ctx, &count, query, userID)
	return count, err
}

// GetOldestSessionsByUser retrieves the oldest sessions for a user (for cleanup)
func (r *sessionRepository) GetOldestSessionsByUser(ctx context.Context, userID string, limit int) ([]*domain.Session, error) {
	query := `
		SELECT id, user_id, expires_at, created_at, ip_address, user_agent, is_active
		FROM sessions 
		WHERE user_id = $1 AND is_active = true
		ORDER BY created_at ASC
		LIMIT $2`

	var sessions []*domain.Session
	tx := database.GetTxFromContext(ctx)
	if tx != nil {
		err := tx.SelectContext(ctx, &sessions, query, userID, limit)
		return sessions, err
	}

	err := r.db.SelectContext(ctx, &sessions, query, userID, limit)
	return sessions, err
}
