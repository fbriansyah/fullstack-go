package application

import (
	"context"
	"fmt"

	"go-templ-template/internal/modules/user/domain"
	"go-templ-template/internal/modules/user/infrastructure"
	"go-templ-template/internal/shared/database"
	"go-templ-template/internal/shared/events"

	"github.com/google/uuid"
)

// UserService defines the interface for user business operations
type UserService interface {
	// CreateUser creates a new user
	CreateUser(ctx context.Context, cmd *CreateUserCommand) (*domain.User, error)

	// GetUser retrieves a user by ID
	GetUser(ctx context.Context, query *GetUserQuery) (*domain.User, error)

	// GetUserByEmail retrieves a user by email
	GetUserByEmail(ctx context.Context, query *GetUserByEmailQuery) (*domain.User, error)

	// UpdateUser updates an existing user's profile
	UpdateUser(ctx context.Context, cmd *UpdateUserCommand) (*domain.User, error)

	// UpdateUserEmail updates a user's email address
	UpdateUserEmail(ctx context.Context, cmd *UpdateUserEmailCommand) (*domain.User, error)

	// ChangeUserPassword changes a user's password
	ChangeUserPassword(ctx context.Context, cmd *ChangeUserPasswordCommand) (*domain.User, error)

	// ChangeUserStatus changes a user's status
	ChangeUserStatus(ctx context.Context, cmd *ChangeUserStatusCommand) (*domain.User, error)

	// DeleteUser deletes a user
	DeleteUser(ctx context.Context, cmd *DeleteUserCommand) error

	// ListUsers lists users with filtering and pagination
	ListUsers(ctx context.Context, query *ListUsersQuery) ([]*domain.User, int64, error)
}

// userServiceImpl implements the UserService interface
type userServiceImpl struct {
	userRepo infrastructure.UserRepository
	eventBus events.EventBus
	db       *database.DB
}

// NewUserService creates a new user service instance
func NewUserService(userRepo infrastructure.UserRepository, eventBus events.EventBus, db *database.DB) UserService {
	return &userServiceImpl{
		userRepo: userRepo,
		eventBus: eventBus,
		db:       db,
	}
}

// CreateUser creates a new user
func (s *userServiceImpl) CreateUser(ctx context.Context, cmd *CreateUserCommand) (*domain.User, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	// Check if user already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, NewInternalError(fmt.Sprintf("failed to check user existence: %v", err))
	}
	if exists {
		return nil, NewUserAlreadyExistsError(cmd.Email)
	}

	// Create new user domain object
	userID := uuid.New().String()
	user, err := domain.NewUser(userID, cmd.Email, cmd.Password, cmd.FirstName, cmd.LastName)
	if err != nil {
		return nil, NewValidationError("user", fmt.Sprintf("failed to create user: %v", err))
	}

	// Execute in transaction
	err = database.ExecuteInTransaction(ctx, s.db, func(txCtx context.Context) error {
		// Create user in repository
		if err := s.userRepo.Create(txCtx, user); err != nil {
			if database.IsDuplicateKeyError(err) {
				return NewUserAlreadyExistsError(cmd.Email)
			}
			return NewInternalError(fmt.Sprintf("failed to create user: %v", err))
		}

		// Publish user created event
		event := domain.NewUserCreatedEvent(user)
		if err := s.eventBus.Publish(txCtx, event); err != nil {
			return NewInternalError(fmt.Sprintf("failed to publish user created event: %v", err))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUser retrieves a user by ID
func (s *userServiceImpl) GetUser(ctx context.Context, query *GetUserQuery) (*domain.User, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, query.ID)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, NewUserNotFoundError(query.ID)
		}
		return nil, NewInternalError(fmt.Sprintf("failed to get user: %v", err))
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *userServiceImpl) GetUserByEmail(ctx context.Context, query *GetUserByEmailQuery) (*domain.User, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByEmail(ctx, query.Email)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, NewUserNotFoundError(query.Email)
		}
		return nil, NewInternalError(fmt.Sprintf("failed to get user by email: %v", err))
	}

	return user, nil
}

// UpdateUser updates an existing user's profile
func (s *userServiceImpl) UpdateUser(ctx context.Context, cmd *UpdateUserCommand) (*domain.User, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	var user *domain.User
	// Execute in transaction
	err := database.ExecuteInTransaction(ctx, s.db, func(txCtx context.Context) error {
		// Get existing user
		var err error
		user, err = s.userRepo.GetByID(txCtx, cmd.ID)
		if err != nil {
			if database.IsNotFoundError(err) {
				return NewUserNotFoundError(cmd.ID)
			}
			return NewInternalError(fmt.Sprintf("failed to get user: %v", err))
		}

		// Check version for optimistic locking
		if user.Version != cmd.Version {
			return NewOptimisticLockError(cmd.ID)
		}

		// Store previous values for event
		previousFirstName := user.FirstName
		previousLastName := user.LastName

		// Update user profile
		if err := user.UpdateProfile(cmd.FirstName, cmd.LastName); err != nil {
			return NewValidationError("profile", fmt.Sprintf("failed to update profile: %v", err))
		}

		// Save updated user
		if err := s.userRepo.Update(txCtx, user); err != nil {
			if database.IsOptimisticLockError(err) {
				return NewOptimisticLockError(cmd.ID)
			}
			return NewInternalError(fmt.Sprintf("failed to update user: %v", err))
		}

		// Publish user updated event
		changes := map[string]interface{}{
			"first_name": map[string]interface{}{
				"old": previousFirstName,
				"new": user.FirstName,
			},
			"last_name": map[string]interface{}{
				"old": previousLastName,
				"new": user.LastName,
			},
		}
		event := domain.NewUserUpdatedEvent(user, changes)
		if err := s.eventBus.Publish(txCtx, event); err != nil {
			return NewInternalError(fmt.Sprintf("failed to publish user updated event: %v", err))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUserEmail updates a user's email address
func (s *userServiceImpl) UpdateUserEmail(ctx context.Context, cmd *UpdateUserEmailCommand) (*domain.User, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	// Check if new email already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, NewInternalError(fmt.Sprintf("failed to check email existence: %v", err))
	}
	if exists {
		return nil, NewUserAlreadyExistsError(cmd.Email)
	}

	var user *domain.User
	// Execute in transaction
	err = database.ExecuteInTransaction(ctx, s.db, func(txCtx context.Context) error {
		// Get existing user
		var err error
		user, err = s.userRepo.GetByID(txCtx, cmd.ID)
		if err != nil {
			if database.IsNotFoundError(err) {
				return NewUserNotFoundError(cmd.ID)
			}
			return NewInternalError(fmt.Sprintf("failed to get user: %v", err))
		}

		// Check version for optimistic locking
		if user.Version != cmd.Version {
			return NewOptimisticLockError(cmd.ID)
		}

		// Store previous email for event
		previousEmail := user.Email

		// Update user email
		if err := user.UpdateEmail(cmd.Email); err != nil {
			return NewValidationError("email", fmt.Sprintf("failed to update email: %v", err))
		}

		// Save updated user
		if err := s.userRepo.Update(txCtx, user); err != nil {
			if database.IsOptimisticLockError(err) {
				return NewOptimisticLockError(cmd.ID)
			}
			if database.IsDuplicateKeyError(err) {
				return NewUserAlreadyExistsError(cmd.Email)
			}
			return NewInternalError(fmt.Sprintf("failed to update user: %v", err))
		}

		// Publish user email changed event
		event := domain.NewUserEmailChangedEvent(user, previousEmail)
		if err := s.eventBus.Publish(txCtx, event); err != nil {
			return NewInternalError(fmt.Sprintf("failed to publish user email changed event: %v", err))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}

// ChangeUserPassword changes a user's password
func (s *userServiceImpl) ChangeUserPassword(ctx context.Context, cmd *ChangeUserPasswordCommand) (*domain.User, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	var user *domain.User
	// Execute in transaction
	err := database.ExecuteInTransaction(ctx, s.db, func(txCtx context.Context) error {
		// Get existing user
		var err error
		user, err = s.userRepo.GetByID(txCtx, cmd.ID)
		if err != nil {
			if database.IsNotFoundError(err) {
				return NewUserNotFoundError(cmd.ID)
			}
			return NewInternalError(fmt.Sprintf("failed to get user: %v", err))
		}

		// Check version for optimistic locking
		if user.Version != cmd.Version {
			return NewOptimisticLockError(cmd.ID)
		}

		// Verify old password
		if !user.CheckPassword(cmd.OldPassword) {
			return NewInvalidPasswordError()
		}

		// Set new password
		if err := user.SetPassword(cmd.NewPassword); err != nil {
			return NewValidationError("password", fmt.Sprintf("failed to set password: %v", err))
		}

		// Save updated user
		if err := s.userRepo.Update(txCtx, user); err != nil {
			if database.IsOptimisticLockError(err) {
				return NewOptimisticLockError(cmd.ID)
			}
			return NewInternalError(fmt.Sprintf("failed to update user: %v", err))
		}

		// Publish user updated event (password change)
		changes := map[string]interface{}{
			"password_changed": true,
		}
		event := domain.NewUserUpdatedEvent(user, changes)
		if err := s.eventBus.Publish(txCtx, event); err != nil {
			return NewInternalError(fmt.Sprintf("failed to publish user updated event: %v", err))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}

// ChangeUserStatus changes a user's status
func (s *userServiceImpl) ChangeUserStatus(ctx context.Context, cmd *ChangeUserStatusCommand) (*domain.User, error) {
	if err := cmd.Validate(); err != nil {
		return nil, err
	}

	var user *domain.User
	// Execute in transaction
	err := database.ExecuteInTransaction(ctx, s.db, func(txCtx context.Context) error {
		// Get existing user
		var err error
		user, err = s.userRepo.GetByID(txCtx, cmd.ID)
		if err != nil {
			if database.IsNotFoundError(err) {
				return NewUserNotFoundError(cmd.ID)
			}
			return NewInternalError(fmt.Sprintf("failed to get user: %v", err))
		}

		// Check version for optimistic locking
		if user.Version != cmd.Version {
			return NewOptimisticLockError(cmd.ID)
		}

		// Store previous status for event
		previousStatus := user.Status

		// Change user status
		if err := user.ChangeStatus(cmd.Status); err != nil {
			return NewValidationError("status", fmt.Sprintf("failed to change status: %v", err))
		}

		// Save updated user
		if err := s.userRepo.Update(txCtx, user); err != nil {
			if database.IsOptimisticLockError(err) {
				return NewOptimisticLockError(cmd.ID)
			}
			return NewInternalError(fmt.Sprintf("failed to update user: %v", err))
		}

		// Publish user status changed event
		event := domain.NewUserStatusChangedEvent(user, previousStatus, cmd.ChangedBy, cmd.Reason)
		if err := s.eventBus.Publish(txCtx, event); err != nil {
			return NewInternalError(fmt.Sprintf("failed to publish user status changed event: %v", err))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser deletes a user
func (s *userServiceImpl) DeleteUser(ctx context.Context, cmd *DeleteUserCommand) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	// Execute in transaction
	err := database.ExecuteInTransaction(ctx, s.db, func(txCtx context.Context) error {
		// Get existing user for event data
		user, err := s.userRepo.GetByID(txCtx, cmd.ID)
		if err != nil {
			if database.IsNotFoundError(err) {
				return NewUserNotFoundError(cmd.ID)
			}
			return NewInternalError(fmt.Sprintf("failed to get user: %v", err))
		}

		// Delete user
		if err := s.userRepo.Delete(txCtx, cmd.ID); err != nil {
			if database.IsNotFoundError(err) {
				return NewUserNotFoundError(cmd.ID)
			}
			return NewInternalError(fmt.Sprintf("failed to delete user: %v", err))
		}

		// Publish user deleted event
		event := domain.NewUserDeletedEvent(user, cmd.DeletedBy, cmd.Reason)
		if err := s.eventBus.Publish(txCtx, event); err != nil {
			return NewInternalError(fmt.Sprintf("failed to publish user deleted event: %v", err))
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// ListUsers lists users with filtering and pagination
func (s *userServiceImpl) ListUsers(ctx context.Context, query *ListUsersQuery) ([]*domain.User, int64, error) {
	if err := query.Validate(); err != nil {
		return nil, 0, err
	}

	// Convert query to repository filter
	filter := infrastructure.UserFilter{
		Status:        query.Status,
		Email:         query.Email,
		FirstName:     query.FirstName,
		LastName:      query.LastName,
		CreatedAfter:  query.CreatedAfter,
		CreatedBefore: query.CreatedBefore,
	}

	// Get users
	users, err := s.userRepo.List(ctx, filter, query.Limit, query.Offset)
	if err != nil {
		return nil, 0, NewInternalError(fmt.Sprintf("failed to list users: %v", err))
	}

	// Get total count
	total, err := s.userRepo.Count(ctx, filter)
	if err != nil {
		return nil, 0, NewInternalError(fmt.Sprintf("failed to count users: %v", err))
	}

	return users, total, nil
}
