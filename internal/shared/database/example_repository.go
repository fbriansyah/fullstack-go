package database

import (
	"context"
	"fmt"
	"time"
)

// ExampleUser represents a user entity for demonstration
type ExampleUser struct {
	ID        string    `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	FirstName string    `db:"first_name" json:"first_name"`
	LastName  string    `db:"last_name" json:"last_name"`
	Status    string    `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	Version   int       `db:"version" json:"version"`
}

// ExampleUserRepository demonstrates how to implement a concrete repository
type ExampleUserRepository struct {
	*BaseRepository[ExampleUser, string]
}

// NewExampleUserRepository creates a new user repository
func NewExampleUserRepository(db *DB) *ExampleUserRepository {
	return &ExampleUserRepository{
		BaseRepository: NewBaseRepository[ExampleUser, string](db, "users", "id"),
	}
}

// Create inserts a new user
func (r *ExampleUserRepository) Create(ctx context.Context, user *ExampleUser) error {
	query := `
		INSERT INTO users (id, email, first_name, last_name, status, created_at, updated_at, version)
		VALUES (:id, :email, :first_name, :last_name, :status, NOW(), NOW(), 1)
		RETURNING created_at, updated_at`

	return r.BaseRepository.Create(ctx, user, query)
}

// GetByID retrieves a user by ID
func (r *ExampleUserRepository) GetByID(ctx context.Context, id string) (*ExampleUser, error) {
	query := `
		SELECT id, email, first_name, last_name, status, created_at, updated_at, version 
		FROM users 
		WHERE id = $1`

	return r.BaseRepository.GetByID(ctx, id, query)
}

// GetByEmail retrieves a user by email
func (r *ExampleUserRepository) GetByEmail(ctx context.Context, email string) (*ExampleUser, error) {
	var user ExampleUser
	query := `
		SELECT id, email, first_name, last_name, status, created_at, updated_at, version 
		FROM users 
		WHERE email = $1`

	tx := GetTxFromContext(ctx)
	if tx != nil {
		err := tx.GetContext(ctx, &user, query, email)
		if err != nil {
			return nil, err
		}
		return &user, nil
	}

	err := r.GetDB().GetContext(ctx, &user, query, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates a user with optimistic locking
func (r *ExampleUserRepository) Update(ctx context.Context, user *ExampleUser) error {
	query := `
		UPDATE users 
		SET email = :email, first_name = :first_name, last_name = :last_name, 
		    status = :status, updated_at = NOW(), version = version + 1
		WHERE id = :id AND version = :version`

	return r.BaseRepository.Update(ctx, user, query)
}

// Delete removes a user by ID
func (r *ExampleUserRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	return r.BaseRepository.Delete(ctx, id, query)
}

// List retrieves users with pagination and optional filtering
func (r *ExampleUserRepository) List(ctx context.Context, limit, offset int) ([]*ExampleUser, error) {
	query := `
		SELECT id, email, first_name, last_name, status, created_at, updated_at, version 
		FROM users 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2`

	return r.BaseRepository.List(ctx, query, limit, offset)
}

// ListByStatus retrieves users by status with pagination
func (r *ExampleUserRepository) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*ExampleUser, error) {
	var users []*ExampleUser
	query := `
		SELECT id, email, first_name, last_name, status, created_at, updated_at, version 
		FROM users 
		WHERE status = $1
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`

	tx := GetTxFromContext(ctx)
	if tx != nil {
		err := tx.SelectContext(ctx, &users, query, status, limit, offset)
		return users, err
	}

	err := r.GetDB().SelectContext(ctx, &users, query, status, limit, offset)
	return users, err
}

// Count returns the total number of users
func (r *ExampleUserRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM users`
	return r.BaseRepository.Count(ctx, query)
}

// CountByStatus returns the number of users with a specific status
func (r *ExampleUserRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM users WHERE status = $1`

	tx := GetTxFromContext(ctx)
	if tx != nil {
		err := tx.GetContext(ctx, &count, query, status)
		return count, err
	}

	err := r.GetDB().GetContext(ctx, &count, query, status)
	return count, err
}

// Exists checks if a user exists by ID
func (r *ExampleUserRepository) Exists(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
	return r.BaseRepository.Exists(ctx, id, query)
}

// ExistsWithEmail checks if a user exists with the given email
func (r *ExampleUserRepository) ExistsWithEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	tx := GetTxFromContext(ctx)
	if tx != nil {
		err := tx.GetContext(ctx, &exists, query, email)
		return exists, err
	}

	err := r.GetDB().GetContext(ctx, &exists, query, email)
	return exists, err
}

// UpdateStatus updates only the status of a user
func (r *ExampleUserRepository) UpdateStatus(ctx context.Context, id, status string) error {
	query := `
		UPDATE users 
		SET status = $2, updated_at = NOW(), version = version + 1
		WHERE id = $1`

	tx := GetTxFromContext(ctx)
	if tx != nil {
		result, err := tx.ExecContext(ctx, query, id, status)
		if err != nil {
			return err
		}
		return r.checkRowsAffected(result)
	}

	result, err := r.GetDB().ExecContext(ctx, query, id, status)
	if err != nil {
		return err
	}
	return r.checkRowsAffected(result)
}

// SearchUsers searches users by name or email using ILIKE
func (r *ExampleUserRepository) SearchUsers(ctx context.Context, searchTerm string, limit, offset int) ([]*ExampleUser, error) {
	var users []*ExampleUser
	query := `
		SELECT id, email, first_name, last_name, status, created_at, updated_at, version 
		FROM users 
		WHERE first_name ILIKE $1 OR last_name ILIKE $1 OR email ILIKE $1
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`

	searchPattern := fmt.Sprintf("%%%s%%", searchTerm)

	tx := GetTxFromContext(ctx)
	if tx != nil {
		err := tx.SelectContext(ctx, &users, query, searchPattern, limit, offset)
		return users, err
	}

	err := r.GetDB().SelectContext(ctx, &users, query, searchPattern, limit, offset)
	return users, err
}

// ExampleUserService demonstrates how to use the repository with transactions
type ExampleUserService struct {
	userRepo *ExampleUserRepository
	tm       *TransactionManager
}

// NewExampleUserService creates a new user service
func NewExampleUserService(userRepo *ExampleUserRepository, tm *TransactionManager) *ExampleUserService {
	return &ExampleUserService{
		userRepo: userRepo,
		tm:       tm,
	}
}

// CreateUser creates a new user with validation
func (s *ExampleUserService) CreateUser(ctx context.Context, user *ExampleUser) error {
	return s.tm.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		// Check if user with email already exists
		exists, err := s.userRepo.ExistsWithEmail(txCtx, user.Email)
		if err != nil {
			return fmt.Errorf("failed to check email existence: %w", err)
		}

		if exists {
			return fmt.Errorf("user with email %s already exists", user.Email)
		}

		// Set default values
		user.Status = "active"
		user.Version = 1

		// Create the user
		if err := s.userRepo.Create(txCtx, user); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		return nil
	})
}

// UpdateUserProfile updates user profile information
func (s *ExampleUserService) UpdateUserProfile(ctx context.Context, userID string, firstName, lastName string) error {
	return s.tm.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		// Get current user
		user, err := s.userRepo.GetByID(txCtx, userID)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		// Update fields
		user.FirstName = firstName
		user.LastName = lastName

		// Save changes
		if err := s.userRepo.Update(txCtx, user); err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}

		return nil
	})
}

// DeactivateUser deactivates a user account
func (s *ExampleUserService) DeactivateUser(ctx context.Context, userID string) error {
	return s.tm.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		// Update user status
		if err := s.userRepo.UpdateStatus(txCtx, userID, "inactive"); err != nil {
			return fmt.Errorf("failed to deactivate user: %w", err)
		}

		// Additional cleanup operations could go here
		// (e.g., invalidate sessions, send notifications, etc.)

		return nil
	})
}

// GetUserStats returns user statistics
func (s *ExampleUserService) GetUserStats(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Get total count
	total, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get total user count: %w", err)
	}
	stats["total"] = total

	// Get active count
	active, err := s.userRepo.CountByStatus(ctx, "active")
	if err != nil {
		return nil, fmt.Errorf("failed to get active user count: %w", err)
	}
	stats["active"] = active

	// Get inactive count
	inactive, err := s.userRepo.CountByStatus(ctx, "inactive")
	if err != nil {
		return nil, fmt.Errorf("failed to get inactive user count: %w", err)
	}
	stats["inactive"] = inactive

	return stats, nil
}
