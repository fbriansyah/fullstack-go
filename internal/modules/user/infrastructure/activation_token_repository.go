package infrastructure

import (
	"context"
	"fmt"
	"time"

	"go-templ-template/internal/modules/user/domain"
	"go-templ-template/internal/shared/database"

	"github.com/jmoiron/sqlx"
)

// activationTokenRepositoryImpl implements the ActivationTokenRepository interface
type activationTokenRepositoryImpl struct {
	db *database.DB
}

// NewActivationTokenRepository creates a new activation token repository instance
func NewActivationTokenRepository(db *database.DB) domain.ActivationTokenRepository {
	return &activationTokenRepositoryImpl{
		db: db,
	}
}

// Create creates a new activation token
func (r *activationTokenRepositoryImpl) Create(ctx context.Context, token *domain.ActivationToken) error {
	query := `
		INSERT INTO activation_tokens (id, user_id, token, expires_at, created_at)
		VALUES (:id, :user_id, :token, :expires_at, :created_at)`

	_, err := r.db.NamedExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to create activation token: %w", err)
	}

	return nil
}

// GetByToken retrieves an activation token by token value
func (r *activationTokenRepositoryImpl) GetByToken(ctx context.Context, token string) (*domain.ActivationToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, used_at, created_at
		FROM activation_tokens
		WHERE token = $1`

	var activationToken domain.ActivationToken
	err := r.db.GetContext(ctx, &activationToken, query, token)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, database.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get activation token: %w", err)
	}

	return &activationToken, nil
}

// GetByUserID retrieves all activation tokens for a user
func (r *activationTokenRepositoryImpl) GetByUserID(ctx context.Context, userID string) ([]*domain.ActivationToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, used_at, created_at
		FROM activation_tokens
		WHERE user_id = $1
		ORDER BY created_at DESC`

	var tokens []*domain.ActivationToken
	err := r.db.SelectContext(ctx, &tokens, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get activation tokens by user ID: %w", err)
	}

	return tokens, nil
}

// Update updates an activation token
func (r *activationTokenRepositoryImpl) Update(ctx context.Context, token *domain.ActivationToken) error {
	query := `
		UPDATE activation_tokens
		SET used_at = :used_at
		WHERE id = :id`

	result, err := r.db.NamedExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to update activation token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return database.ErrNotFound
	}

	return nil
}

// Delete deletes an activation token by ID
func (r *activationTokenRepositoryImpl) Delete(ctx context.Context, tokenID string) error {
	query := `DELETE FROM activation_tokens WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, tokenID)
	if err != nil {
		return fmt.Errorf("failed to delete activation token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return database.ErrNotFound
	}

	return nil
}

// DeleteByUserID deletes all activation tokens for a user
func (r *activationTokenRepositoryImpl) DeleteByUserID(ctx context.Context, userID string) error {
	query := `DELETE FROM activation_tokens WHERE user_id = $1`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete activation tokens by user ID: %w", err)
	}

	return nil
}

// DeleteExpired deletes all expired activation tokens
func (r *activationTokenRepositoryImpl) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM activation_tokens WHERE expires_at < $1`

	result, err := r.db.ExecContext(ctx, query, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to delete expired activation tokens: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	// Log the number of deleted tokens (in a real implementation, use proper logging)
	if rowsAffected > 0 {
		fmt.Printf("Deleted %d expired activation tokens\n", rowsAffected)
	}

	return nil
}

// CreateActivationTokensTable creates the activation_tokens table if it doesn't exist
func CreateActivationTokensTable(ctx context.Context, db *sqlx.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS activation_tokens (
			id VARCHAR(255) PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			token VARCHAR(255) NOT NULL UNIQUE,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			used_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);

		CREATE INDEX IF NOT EXISTS idx_activation_tokens_user_id ON activation_tokens(user_id);
		CREATE INDEX IF NOT EXISTS idx_activation_tokens_token ON activation_tokens(token);
		CREATE INDEX IF NOT EXISTS idx_activation_tokens_expires_at ON activation_tokens(expires_at);
		CREATE INDEX IF NOT EXISTS idx_activation_tokens_used_at ON activation_tokens(used_at);
	`

	_, err := db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create activation_tokens table: %w", err)
	}

	return nil
}
