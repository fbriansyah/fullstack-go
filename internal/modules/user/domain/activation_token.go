package domain

import (
	"context"
	"time"
)

// ActivationToken represents an activation token
type ActivationToken struct {
	ID        string     `db:"id" json:"id"`
	UserID    string     `db:"user_id" json:"user_id"`
	Token     string     `db:"token" json:"token"`
	ExpiresAt time.Time  `db:"expires_at" json:"expires_at"`
	UsedAt    *time.Time `db:"used_at" json:"used_at,omitempty"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
}

// IsExpired checks if the activation token is expired
func (t *ActivationToken) IsExpired() bool {
	return time.Now().UTC().After(t.ExpiresAt)
}

// IsUsed checks if the activation token has been used
func (t *ActivationToken) IsUsed() bool {
	return t.UsedAt != nil
}

// IsValid checks if the activation token is valid (not expired and not used)
func (t *ActivationToken) IsValid() bool {
	return !t.IsExpired() && !t.IsUsed()
}

// ActivationTokenRepository defines the interface for activation token data access
type ActivationTokenRepository interface {
	Create(ctx context.Context, token *ActivationToken) error
	GetByToken(ctx context.Context, token string) (*ActivationToken, error)
	GetByUserID(ctx context.Context, userID string) ([]*ActivationToken, error)
	Update(ctx context.Context, token *ActivationToken) error
	Delete(ctx context.Context, tokenID string) error
	DeleteByUserID(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) error
}
