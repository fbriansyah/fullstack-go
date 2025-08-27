package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"go-templ-template/internal/modules/user/domain"
	"go-templ-template/internal/modules/user/infrastructure"
	"go-templ-template/internal/shared/database"
	"go-templ-template/internal/shared/events"
)

// ActivationService defines the interface for user activation operations
type ActivationService interface {
	// RequestActivation creates an activation request for a user
	RequestActivation(ctx context.Context, cmd *RequestActivationCommand) (*domain.ActivationToken, error)

	// ActivateUser activates a user using an activation token
	ActivateUser(ctx context.Context, cmd *ActivateUserCommand) (*domain.User, error)

	// DeactivateUser deactivates a user account
	DeactivateUser(ctx context.Context, cmd *DeactivateUserCommand) (*domain.User, error)

	// CleanupExpiredTokens removes expired activation tokens
	CleanupExpiredTokens(ctx context.Context) error

	// GetActivationToken retrieves an activation token by token value
	GetActivationToken(ctx context.Context, token string) (*domain.ActivationToken, error)
}

// activationServiceImpl implements the ActivationService interface
type activationServiceImpl struct {
	userRepo      infrastructure.UserRepository
	tokenRepo     domain.ActivationTokenRepository
	eventBus      events.EventBus
	db            *database.DB
	tokenDuration time.Duration
}

// NewActivationService creates a new activation service instance
func NewActivationService(
	userRepo infrastructure.UserRepository,
	tokenRepo domain.ActivationTokenRepository,
	eventBus events.EventBus,
	db *database.DB,
	tokenDuration time.Duration,
) ActivationService {
	if tokenDuration == 0 {
		tokenDuration = 24 * time.Hour // Default 24 hours
	}

	return &activationServiceImpl{
		userRepo:      userRepo,
		tokenRepo:     tokenRepo,
		eventBus:      eventBus,
		db:            db,
		tokenDuration: tokenDuration,
	}
}

// RequestActivation creates an activation request for a user
func (s *activationServiceImpl) RequestActivation(ctx context.Context, cmd *RequestActivationCommand) (*domain.ActivationToken, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	var token *domain.ActivationToken
	// Execute in transaction
	err := database.ExecuteInTransaction(ctx, s.db, func(txCtx context.Context) error {
		// Get user
		user, err := s.userRepo.GetByID(txCtx, cmd.UserID)
		if err != nil {
			if database.IsNotFoundError(err) {
				return NewUserNotFoundError(cmd.UserID)
			}
			return NewInternalError(fmt.Sprintf("failed to get user: %v", err))
		}

		// Check if user is already active
		if user.Status == domain.UserStatusActive {
			return NewValidationError("activation", "user is already active")
		}

		// Delete any existing tokens for this user
		if err := s.tokenRepo.DeleteByUserID(txCtx, cmd.UserID); err != nil {
			return NewInternalError(fmt.Sprintf("failed to delete existing tokens: %v", err))
		}

		// Generate activation token
		tokenValue, err := generateActivationToken()
		if err != nil {
			return NewInternalError(fmt.Sprintf("failed to generate token: %v", err))
		}

		// Create activation token
		token = &domain.ActivationToken{
			ID:        generateTokenID(),
			UserID:    cmd.UserID,
			Token:     tokenValue,
			ExpiresAt: time.Now().UTC().Add(s.tokenDuration),
			CreatedAt: time.Now().UTC(),
		}

		// Save token
		if err := s.tokenRepo.Create(txCtx, token); err != nil {
			return NewInternalError(fmt.Sprintf("failed to create activation token: %v", err))
		}

		// Publish activation requested event
		event := domain.NewUserActivationRequestedEvent(user, token.Token, token.ExpiresAt, cmd.RequestedBy)
		if err := s.eventBus.Publish(txCtx, event); err != nil {
			return NewInternalError(fmt.Sprintf("failed to publish activation requested event: %v", err))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return token, nil
}

// ActivateUser activates a user using an activation token
func (s *activationServiceImpl) ActivateUser(ctx context.Context, cmd *ActivateUserCommand) (*domain.User, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	var user *domain.User
	// Execute in transaction
	err := database.ExecuteInTransaction(ctx, s.db, func(txCtx context.Context) error {
		// Get activation token
		token, err := s.tokenRepo.GetByToken(txCtx, cmd.Token)
		if err != nil {
			if database.IsNotFoundError(err) {
				return NewValidationError("token", "invalid activation token")
			}
			return NewInternalError(fmt.Sprintf("failed to get activation token: %v", err))
		}

		// Validate token
		if !token.IsValid() {
			if token.IsExpired() {
				// Publish token expired event
				expiredEvent := domain.NewUserActivationTokenExpiredEvent(token.UserID, "", token.Token)
				if err := s.eventBus.Publish(txCtx, expiredEvent); err != nil {
					// Log error but don't fail the operation
				}
				return NewValidationError("token", "activation token has expired")
			}
			if token.IsUsed() {
				return NewValidationError("token", "activation token has already been used")
			}
		}

		// Get user
		user, err = s.userRepo.GetByID(txCtx, token.UserID)
		if err != nil {
			if database.IsNotFoundError(err) {
				return NewUserNotFoundError(token.UserID)
			}
			return NewInternalError(fmt.Sprintf("failed to get user: %v", err))
		}

		// Check if user is already active
		if user.Status == domain.UserStatusActive {
			return NewValidationError("activation", "user is already active")
		}

		// Activate user
		if err := user.Activate(); err != nil {
			return NewValidationError("activation", fmt.Sprintf("failed to activate user: %v", err))
		}

		// Update user
		if err := s.userRepo.Update(txCtx, user); err != nil {
			return NewInternalError(fmt.Sprintf("failed to update user: %v", err))
		}

		// Mark token as used
		now := time.Now().UTC()
		token.UsedAt = &now
		if err := s.tokenRepo.Update(txCtx, token); err != nil {
			return NewInternalError(fmt.Sprintf("failed to update token: %v", err))
		}

		// Publish user activated event
		event := domain.NewUserActivatedEvent(user, cmd.ActivatedBy, "token")
		if err := s.eventBus.Publish(txCtx, event); err != nil {
			return NewInternalError(fmt.Sprintf("failed to publish user activated event: %v", err))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}

// DeactivateUser deactivates a user account
func (s *activationServiceImpl) DeactivateUser(ctx context.Context, cmd *DeactivateUserCommand) (*domain.User, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	var user *domain.User
	// Execute in transaction
	err := database.ExecuteInTransaction(ctx, s.db, func(txCtx context.Context) error {
		// Get user
		var err error
		user, err = s.userRepo.GetByID(txCtx, cmd.UserID)
		if err != nil {
			if database.IsNotFoundError(err) {
				return NewUserNotFoundError(cmd.UserID)
			}
			return NewInternalError(fmt.Sprintf("failed to get user: %v", err))
		}

		// Check version for optimistic locking
		if user.Version != cmd.Version {
			return NewOptimisticLockError(cmd.UserID)
		}

		// Check if user is already inactive
		if user.Status == domain.UserStatusInactive {
			return NewValidationError("deactivation", "user is already inactive")
		}

		// Deactivate user
		if err := user.Deactivate(); err != nil {
			return NewValidationError("deactivation", fmt.Sprintf("failed to deactivate user: %v", err))
		}

		// Update user
		if err := s.userRepo.Update(txCtx, user); err != nil {
			if database.IsOptimisticLockError(err) {
				return NewOptimisticLockError(cmd.UserID)
			}
			return NewInternalError(fmt.Sprintf("failed to update user: %v", err))
		}

		// Delete any pending activation tokens
		if err := s.tokenRepo.DeleteByUserID(txCtx, cmd.UserID); err != nil {
			return NewInternalError(fmt.Sprintf("failed to delete activation tokens: %v", err))
		}

		// Publish user deactivated event
		event := domain.NewUserDeactivatedEvent(user, cmd.DeactivatedBy, cmd.Reason)
		if err := s.eventBus.Publish(txCtx, event); err != nil {
			return NewInternalError(fmt.Sprintf("failed to publish user deactivated event: %v", err))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}

// CleanupExpiredTokens removes expired activation tokens
func (s *activationServiceImpl) CleanupExpiredTokens(ctx context.Context) error {
	if err := s.tokenRepo.DeleteExpired(ctx); err != nil {
		return NewInternalError(fmt.Sprintf("failed to cleanup expired tokens: %v", err))
	}
	return nil
}

// GetActivationToken retrieves an activation token by token value
func (s *activationServiceImpl) GetActivationToken(ctx context.Context, token string) (*domain.ActivationToken, error) {
	activationToken, err := s.tokenRepo.GetByToken(ctx, token)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, NewValidationError("token", "activation token not found")
		}
		return nil, NewInternalError(fmt.Sprintf("failed to get activation token: %v", err))
	}

	return activationToken, nil
}

// generateActivationToken generates a secure random activation token
func generateActivationToken() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// generateTokenID generates a unique token ID
func generateTokenID() string {
	return fmt.Sprintf("token_%d", time.Now().UnixNano())
}
