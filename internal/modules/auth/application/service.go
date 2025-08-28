package application

import (
	"context"
	"fmt"
	"time"

	"go-templ-template/internal/modules/auth/domain"
	"go-templ-template/internal/modules/user/application"
	userDomain "go-templ-template/internal/modules/user/domain"
	"go-templ-template/internal/shared/database"
	"go-templ-template/internal/shared/events"
)

// AuthService defines the interface for authentication business operations
type AuthService interface {
	// Login authenticates a user and creates a session
	Login(ctx context.Context, cmd *LoginCommand) (*AuthResult, error)

	// Register creates a new user account and logs them in
	Register(ctx context.Context, cmd *RegisterCommand) (*AuthResult, error)

	// Logout terminates a user session
	Logout(ctx context.Context, cmd *LogoutCommand) error

	// ValidateSession validates a session and returns user information
	ValidateSession(ctx context.Context, query *ValidateSessionQuery) (*SessionValidationResult, error)

	// RefreshSession extends a session's expiration time
	RefreshSession(ctx context.Context, cmd *RefreshSessionCommand) (*domain.Session, error)

	// ChangePassword changes a user's password
	ChangePassword(ctx context.Context, cmd *ChangePasswordCommand) error

	// CleanupExpiredSessions removes expired sessions from storage
	CleanupExpiredSessions(ctx context.Context) error

	// GetUserSessions returns all active sessions for a user (for testing/admin purposes)
	GetUserSessions(ctx context.Context, userID string) ([]*domain.Session, error)
}

// AuthResult represents the result of a successful authentication
type AuthResult struct {
	User    *userDomain.User `json:"user"`
	Session *domain.Session  `json:"session"`
}

// SessionValidationResult represents the result of session validation
type SessionValidationResult struct {
	User    *userDomain.User `json:"user"`
	Session *domain.Session  `json:"session"`
	Valid   bool             `json:"valid"`
}

// SessionRepository defines the interface for session data access
type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	GetByID(ctx context.Context, sessionID string) (*domain.Session, error)
	GetByUserID(ctx context.Context, userID string) ([]*domain.Session, error)
	Update(ctx context.Context, session *domain.Session) error
	Delete(ctx context.Context, sessionID string) error
	DeleteByUserID(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) error
	ExistsByID(ctx context.Context, sessionID string) (bool, error)
}

// authServiceImpl implements the AuthService interface
type authServiceImpl struct {
	sessionRepo   SessionRepository
	userService   application.UserService
	eventBus      events.EventBus
	db            *database.DB
	rateLimiter   RateLimiter
	sessionConfig domain.SessionConfig
	hasher        *domain.PasswordHasher
}

// NewAuthService creates a new auth service instance
func NewAuthService(
	sessionRepo SessionRepository,
	userService application.UserService,
	eventBus events.EventBus,
	db *database.DB,
	rateLimiter RateLimiter,
	sessionConfig domain.SessionConfig,
) AuthService {
	return &authServiceImpl{
		sessionRepo:   sessionRepo,
		userService:   userService,
		eventBus:      eventBus,
		db:            db,
		rateLimiter:   rateLimiter,
		sessionConfig: sessionConfig,
		hasher:        domain.NewPasswordHasher(),
	}
}

// Login authenticates a user and creates a session
func (s *authServiceImpl) Login(ctx context.Context, cmd *LoginCommand) (*AuthResult, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	// Rate limiting by email
	rateLimitKey := fmt.Sprintf("login:%s", cmd.Email)
	allowed, err := s.rateLimiter.Allow(ctx, rateLimitKey)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, err // Rate limiter already returns appropriate error
	}

	// Get user by email
	userQuery := &application.GetUserByEmailQuery{Email: cmd.Email}
	user, err := s.userService.GetUserByEmail(ctx, userQuery)
	if err != nil {
		// Don't reveal whether user exists or not
		return nil, NewInvalidCredentialsError()
	}

	// Check if user account is active
	if !user.IsActive() {
		switch user.Status {
		case userDomain.UserStatusSuspended:
			return nil, NewAccountSuspendedError()
		case userDomain.UserStatusInactive:
			return nil, NewInvalidCredentialsError()
		default:
			return nil, NewInvalidCredentialsError()
		}
	}

	// Verify password
	if !user.CheckPassword(cmd.Password) {
		return nil, NewInvalidCredentialsError()
	}

	// Reset rate limit on successful login
	if err := s.rateLimiter.Reset(ctx, rateLimitKey); err != nil {
		// Log error but don't fail the login
		// In production, you might want to log this
	}

	var result *AuthResult
	// Execute in transaction
	err = database.ExecuteInTransaction(ctx, s.db, func(txCtx context.Context) error {
		// Create new session
		session, err := domain.NewSession(user.ID, cmd.IPAddress, cmd.UserAgent, s.sessionConfig)
		if err != nil {
			return NewInternalError(fmt.Sprintf("failed to create session: %v", err))
		}

		// Save session
		if err := s.sessionRepo.Create(txCtx, session); err != nil {
			return NewInternalError(fmt.Sprintf("failed to save session: %v", err))
		}

		// Publish user logged in event
		event := domain.NewUserLoggedInEvent(user.ID, session.ID, cmd.IPAddress, cmd.UserAgent)
		if err := s.eventBus.Publish(txCtx, event); err != nil {
			return NewInternalError(fmt.Sprintf("failed to publish login event: %v", err))
		}

		result = &AuthResult{
			User:    user,
			Session: session,
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Register creates a new user account and logs them in
func (s *authServiceImpl) Register(ctx context.Context, cmd *RegisterCommand) (*AuthResult, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	// Rate limiting by IP address for registration
	rateLimitKey := fmt.Sprintf("register:%s", cmd.IPAddress)
	allowed, err := s.rateLimiter.Allow(ctx, rateLimitKey)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, err // Rate limiter already returns appropriate error
	}

	var result *AuthResult
	// Execute in transaction
	err = database.ExecuteInTransaction(ctx, s.db, func(txCtx context.Context) error {
		// Create user
		createUserCmd := &application.CreateUserCommand{
			Email:     cmd.Email,
			Password:  cmd.Password,
			FirstName: cmd.FirstName,
			LastName:  cmd.LastName,
		}

		user, err := s.userService.CreateUser(txCtx, createUserCmd)
		if err != nil {
			// Check if it's a user already exists error
			if userErr, ok := err.(*application.ApplicationError); ok && userErr.Code == "USER_ALREADY_EXISTS" {
				return NewUserAlreadyExistsError(cmd.Email)
			}
			return NewInternalError(fmt.Sprintf("failed to create user: %v", err))
		}

		// Create session for the new user
		session, err := domain.NewSession(user.ID, cmd.IPAddress, cmd.UserAgent, s.sessionConfig)
		if err != nil {
			return NewInternalError(fmt.Sprintf("failed to create session: %v", err))
		}

		// Save session
		if err := s.sessionRepo.Create(txCtx, session); err != nil {
			return NewInternalError(fmt.Sprintf("failed to save session: %v", err))
		}

		// Publish user registered event
		event := domain.NewUserRegisteredEvent(
			user.ID, user.Email, user.FirstName, user.LastName,
			cmd.IPAddress, cmd.UserAgent,
		)
		if err := s.eventBus.Publish(txCtx, event); err != nil {
			return NewInternalError(fmt.Sprintf("failed to publish registration event: %v", err))
		}

		result = &AuthResult{
			User:    user,
			Session: session,
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Reset rate limit on successful registration
	if err := s.rateLimiter.Reset(ctx, rateLimitKey); err != nil {
		// Log error but don't fail the registration
	}

	return result, nil
}

// Logout terminates a user session
func (s *authServiceImpl) Logout(ctx context.Context, cmd *LogoutCommand) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	// Execute in transaction
	err := database.ExecuteInTransaction(ctx, s.db, func(txCtx context.Context) error {
		// Get session to retrieve user ID for event
		session, err := s.sessionRepo.GetByID(txCtx, cmd.SessionID)
		if err != nil {
			if database.IsNotFoundError(err) {
				return NewSessionNotFoundError(cmd.SessionID)
			}
			return NewInternalError(fmt.Sprintf("failed to get session: %v", err))
		}

		// Delete session
		if err := s.sessionRepo.Delete(txCtx, cmd.SessionID); err != nil {
			if database.IsNotFoundError(err) {
				return NewSessionNotFoundError(cmd.SessionID)
			}
			return NewInternalError(fmt.Sprintf("failed to delete session: %v", err))
		}

		// Publish user logged out event
		event := domain.NewUserLoggedOutEvent(session.UserID, session.ID, "manual")
		if err := s.eventBus.Publish(txCtx, event); err != nil {
			return NewInternalError(fmt.Sprintf("failed to publish logout event: %v", err))
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// ValidateSession validates a session and returns user information
func (s *authServiceImpl) ValidateSession(ctx context.Context, query *ValidateSessionQuery) (*SessionValidationResult, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}

	// Get session
	session, err := s.sessionRepo.GetByID(ctx, query.SessionID)
	if err != nil {
		if database.IsNotFoundError(err) {
			return &SessionValidationResult{Valid: false}, nil
		}
		return nil, NewInternalError(fmt.Sprintf("failed to get session: %v", err))
	}

	// Check if session is valid
	if !session.IsValid() {
		// Session is expired or inactive
		if session.IsExpired() {
			// Publish session expired event
			event := domain.NewSessionExpiredEvent(session.UserID, session.ID, session.CreatedAt)
			if err := s.eventBus.Publish(ctx, event); err != nil {
				// Log error but don't fail validation
			}

			// Clean up expired session
			if err := s.sessionRepo.Delete(ctx, session.ID); err != nil {
				// Log error but don't fail validation
			}
		}

		return &SessionValidationResult{Valid: false}, nil
	}

	// Validate security context if provided
	if query.IPAddress != "" || query.UserAgent != "" {
		if !session.ValidateSecurityContext(query.IPAddress, query.UserAgent) {
			// For security, we might want to invalidate the session
			// For now, we'll just return invalid
			return &SessionValidationResult{Valid: false}, nil
		}
	}

	// Get user information
	userQuery := &application.GetUserQuery{ID: session.UserID}
	user, err := s.userService.GetUser(ctx, userQuery)
	if err != nil {
		return &SessionValidationResult{Valid: false}, nil
	}

	// Check if user is still active
	if !user.IsActive() {
		return &SessionValidationResult{Valid: false}, nil
	}

	return &SessionValidationResult{
		User:    user,
		Session: session,
		Valid:   true,
	}, nil
}

// RefreshSession extends a session's expiration time
func (s *authServiceImpl) RefreshSession(ctx context.Context, cmd *RefreshSessionCommand) (*domain.Session, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	var session *domain.Session
	// Execute in transaction
	err := database.ExecuteInTransaction(ctx, s.db, func(txCtx context.Context) error {
		// Get session
		var err error
		session, err = s.sessionRepo.GetByID(txCtx, cmd.SessionID)
		if err != nil {
			if database.IsNotFoundError(err) {
				return NewSessionNotFoundError(cmd.SessionID)
			}
			return NewInternalError(fmt.Sprintf("failed to get session: %v", err))
		}

		// Check if session is valid
		if !session.IsValid() {
			return NewSessionExpiredError()
		}

		// Validate security context if provided
		if cmd.IPAddress != "" || cmd.UserAgent != "" {
			if !session.ValidateSecurityContext(cmd.IPAddress, cmd.UserAgent) {
				return NewSessionInvalidError()
			}
		}

		// Extend session
		session.Extend(s.sessionConfig.DefaultDuration)

		// Update session
		if err := s.sessionRepo.Update(txCtx, session); err != nil {
			return NewInternalError(fmt.Sprintf("failed to update session: %v", err))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return session, nil
}

// ChangePassword changes a user's password
func (s *authServiceImpl) ChangePassword(ctx context.Context, cmd *ChangePasswordCommand) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	// Rate limiting by user ID
	rateLimitKey := fmt.Sprintf("change_password:%s", cmd.UserID)
	allowed, err := s.rateLimiter.Allow(ctx, rateLimitKey)
	if err != nil {
		return err
	}
	if !allowed {
		return err // Rate limiter already returns appropriate error
	}

	// Execute in transaction
	err = database.ExecuteInTransaction(ctx, s.db, func(txCtx context.Context) error {
		// Get user
		userQuery := &application.GetUserQuery{ID: cmd.UserID}
		user, err := s.userService.GetUser(txCtx, userQuery)
		if err != nil {
			return NewUserNotFoundError(cmd.UserID)
		}

		// Verify old password
		if !user.CheckPassword(cmd.OldPassword) {
			return NewInvalidCredentialsError()
		}

		// Change password using user service
		changePasswordCmd := &application.ChangeUserPasswordCommand{
			ID:          cmd.UserID,
			OldPassword: cmd.OldPassword,
			NewPassword: cmd.NewPassword,
			Version:     user.Version,
		}

		_, err = s.userService.ChangeUserPassword(txCtx, changePasswordCmd)
		if err != nil {
			return NewInternalError(fmt.Sprintf("failed to change password: %v", err))
		}

		// Publish password changed event
		event := domain.NewPasswordChangedEvent(cmd.UserID, cmd.IPAddress, cmd.UserAgent)
		if err := s.eventBus.Publish(txCtx, event); err != nil {
			return NewInternalError(fmt.Sprintf("failed to publish password changed event: %v", err))
		}

		// Invalidate all sessions for this user (force re-login)
		if err := s.sessionRepo.DeleteByUserID(txCtx, cmd.UserID); err != nil {
			return NewInternalError(fmt.Sprintf("failed to invalidate sessions: %v", err))
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Reset rate limit on successful password change
	if err := s.rateLimiter.Reset(ctx, rateLimitKey); err != nil {
		// Log error but don't fail the operation
	}

	return nil
}

// CleanupExpiredSessions removes expired sessions from storage
func (s *authServiceImpl) CleanupExpiredSessions(ctx context.Context) error {
	if err := s.sessionRepo.DeleteExpired(ctx); err != nil {
		return NewInternalError(fmt.Sprintf("failed to cleanup expired sessions: %v", err))
	}
	return nil
}

// GetUserSessions returns all active sessions for a user
func (s *authServiceImpl) GetUserSessions(ctx context.Context, userID string) ([]*domain.Session, error) {
	sessions, err := s.sessionRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, NewInternalError(fmt.Sprintf("failed to get user sessions: %v", err))
	}
	return sessions, nil
}

// DefaultSessionConfig returns a default session configuration
func DefaultSessionConfig() domain.SessionConfig {
	return domain.SessionConfig{
		DefaultDuration: time.Hour * 24,     // 24 hours
		MaxDuration:     time.Hour * 24 * 7, // 7 days
		CleanupInterval: time.Hour,          // cleanup every hour
	}
}
