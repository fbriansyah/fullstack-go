package infrastructure

import (
	"context"

	"go-templ-template/internal/modules/auth/application"
	"go-templ-template/internal/modules/auth/domain"
	"go-templ-template/internal/shared/database"
)

// sessionRepositoryAdapter adapts the infrastructure SessionRepository to the application SessionRepository interface
type sessionRepositoryAdapter struct {
	repo SessionRepository
}

// NewSessionRepositoryAdapter creates a new adapter that implements the application SessionRepository interface
func NewSessionRepositoryAdapter(db *database.DB) application.SessionRepository {
	infraRepo := NewSessionRepository(db)
	return &sessionRepositoryAdapter{
		repo: infraRepo,
	}
}

// Create inserts a new session
func (a *sessionRepositoryAdapter) Create(ctx context.Context, session *domain.Session) error {
	return a.repo.Create(ctx, session)
}

// GetByID retrieves a session by its ID
func (a *sessionRepositoryAdapter) GetByID(ctx context.Context, sessionID string) (*domain.Session, error) {
	return a.repo.GetByID(ctx, sessionID)
}

// GetByUserID retrieves all active sessions for a user
func (a *sessionRepositoryAdapter) GetByUserID(ctx context.Context, userID string) ([]*domain.Session, error) {
	return a.repo.GetByUserID(ctx, userID)
}

// Update updates an existing session
func (a *sessionRepositoryAdapter) Update(ctx context.Context, session *domain.Session) error {
	return a.repo.Update(ctx, session)
}

// Delete removes a session by its ID
func (a *sessionRepositoryAdapter) Delete(ctx context.Context, sessionID string) error {
	return a.repo.Delete(ctx, sessionID)
}

// DeleteByUserID removes all sessions for a user
func (a *sessionRepositoryAdapter) DeleteByUserID(ctx context.Context, userID string) error {
	return a.repo.DeleteByUserID(ctx, userID)
}

// DeleteExpired removes all expired sessions (adapts CleanupExpired)
func (a *sessionRepositoryAdapter) DeleteExpired(ctx context.Context) error {
	_, err := a.repo.CleanupExpired(ctx)
	return err
}

// ExistsByID checks if a session exists by its ID
func (a *sessionRepositoryAdapter) ExistsByID(ctx context.Context, sessionID string) (bool, error) {
	session, err := a.repo.GetByID(ctx, sessionID)
	if err != nil {
		if database.IsNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return session != nil, nil
}
