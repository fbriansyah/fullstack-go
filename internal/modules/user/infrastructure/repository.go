package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"go-templ-template/internal/modules/user/domain"
	"go-templ-template/internal/shared/database"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	// Create inserts a new user into the database
	Create(ctx context.Context, user *domain.User) error

	// GetByID retrieves a user by their ID
	GetByID(ctx context.Context, id string) (*domain.User, error)

	// GetByEmail retrieves a user by their email address
	GetByEmail(ctx context.Context, email string) (*domain.User, error)

	// Update updates an existing user with optimistic locking
	Update(ctx context.Context, user *domain.User) error

	// Delete removes a user by their ID
	Delete(ctx context.Context, id string) error

	// List retrieves users with pagination and optional filtering
	List(ctx context.Context, filter UserFilter, limit, offset int) ([]*domain.User, error)

	// Count returns the total number of users matching the filter
	Count(ctx context.Context, filter UserFilter) (int64, error)

	// Exists checks if a user exists by ID
	Exists(ctx context.Context, id string) (bool, error)

	// ExistsByEmail checks if a user exists by email
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

// UserFilter represents filtering options for user queries
type UserFilter struct {
	Status        *domain.UserStatus
	Email         *string
	FirstName     *string
	LastName      *string
	CreatedAfter  *string
	CreatedBefore *string
}

// userRepositoryImpl implements the UserRepository interface using SQLx
type userRepositoryImpl struct {
	*database.BaseRepository[domain.User, string]
}

// NewUserRepository creates a new user repository instance
func NewUserRepository(db *database.DB) UserRepository {
	baseRepo := database.NewBaseRepository[domain.User, string](db, "users", "id")
	return &userRepositoryImpl{
		BaseRepository: baseRepo,
	}
}

// Create inserts a new user into the database
func (r *userRepositoryImpl) Create(ctx context.Context, user *domain.User) error {
	// Generate UUID if not provided
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, status, created_at, updated_at, version)
		VALUES (:id, :email, :password, :first_name, :last_name, :status, :created_at, :updated_at, :version)`

	err := r.BaseRepository.Create(ctx, user, query)
	if err != nil {
		return r.handleError("Create", err)
	}

	return nil
}

// GetByID retrieves a user by their ID
func (r *userRepositoryImpl) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash as password, first_name, last_name, status, created_at, updated_at, version
		FROM users 
		WHERE id = $1`

	user, err := r.BaseRepository.GetByID(ctx, id, query)
	if err != nil {
		return nil, r.handleError("GetByID", err)
	}

	return user, nil
}

// GetByEmail retrieves a user by their email address
func (r *userRepositoryImpl) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash as password, first_name, last_name, status, created_at, updated_at, version
		FROM users 
		WHERE email = $1`

	var user domain.User
	tx := database.GetTxFromContext(ctx)

	var err error
	if tx != nil {
		err = tx.GetContext(ctx, &user, query, strings.ToLower(email))
	} else {
		err = r.GetDB().GetContext(ctx, &user, query, strings.ToLower(email))
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, database.ErrNotFound
		}
		return nil, r.handleError("GetByEmail", err)
	}

	return &user, nil
}

// Update updates an existing user with optimistic locking
func (r *userRepositoryImpl) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users 
		SET email = :email, 
		    password_hash = :password, 
		    first_name = :first_name, 
		    last_name = :last_name, 
		    status = :status, 
		    updated_at = :updated_at, 
		    version = :version
		WHERE id = :id AND version = :version - 1`

	tx := database.GetTxFromContext(ctx)
	var result sql.Result
	var err error

	if tx != nil {
		result, err = tx.NamedExecContext(ctx, query, user)
	} else {
		result, err = r.GetDB().NamedExecContext(ctx, query, user)
	}

	if err != nil {
		return r.handleError("Update", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return r.handleError("Update", err)
	}

	if rowsAffected == 0 {
		// Check if user exists to distinguish between not found and version conflict
		exists, existsErr := r.Exists(ctx, user.ID)
		if existsErr != nil {
			return r.handleError("Update", existsErr)
		}
		if !exists {
			return database.ErrNotFound
		}
		return database.ErrOptimisticLock
	}

	return nil
}

// Delete removes a user by their ID
func (r *userRepositoryImpl) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	err := r.BaseRepository.Delete(ctx, id, query)
	if err != nil {
		return r.handleError("Delete", err)
	}

	return nil
}

// List retrieves users with pagination and optional filtering
func (r *userRepositoryImpl) List(ctx context.Context, filter UserFilter, limit, offset int) ([]*domain.User, error) {
	query, args := r.buildListQuery(filter, limit, offset)

	var users []*domain.User
	tx := database.GetTxFromContext(ctx)

	var err error
	if tx != nil {
		err = tx.SelectContext(ctx, &users, query, args...)
	} else {
		err = r.GetDB().SelectContext(ctx, &users, query, args...)
	}

	if err != nil {
		return nil, r.handleError("List", err)
	}

	return users, nil
}

// Count returns the total number of users matching the filter
func (r *userRepositoryImpl) Count(ctx context.Context, filter UserFilter) (int64, error) {
	query, args := r.buildCountQuery(filter)

	var count int64
	tx := database.GetTxFromContext(ctx)

	var err error
	if tx != nil {
		err = tx.GetContext(ctx, &count, query, args...)
	} else {
		err = r.GetDB().GetContext(ctx, &count, query, args...)
	}

	if err != nil {
		return 0, r.handleError("Count", err)
	}

	return count, nil
}

// Exists checks if a user exists by ID
func (r *userRepositoryImpl) Exists(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`

	var exists bool
	tx := database.GetTxFromContext(ctx)

	var err error
	if tx != nil {
		err = tx.GetContext(ctx, &exists, query, id)
	} else {
		err = r.GetDB().GetContext(ctx, &exists, query, id)
	}

	if err != nil {
		return false, r.handleError("Exists", err)
	}

	return exists, nil
}

// ExistsByEmail checks if a user exists by email
func (r *userRepositoryImpl) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	tx := database.GetTxFromContext(ctx)

	var err error
	if tx != nil {
		err = tx.GetContext(ctx, &exists, query, strings.ToLower(email))
	} else {
		err = r.GetDB().GetContext(ctx, &exists, query, strings.ToLower(email))
	}

	if err != nil {
		return false, r.handleError("ExistsByEmail", err)
	}

	return exists, nil
}

// buildListQuery constructs the SQL query for listing users with filters
func (r *userRepositoryImpl) buildListQuery(filter UserFilter, limit, offset int) (string, []interface{}) {
	query := `
		SELECT id, email, password_hash as password, first_name, last_name, status, created_at, updated_at, version
		FROM users`

	whereClause, args := r.buildWhereClause(filter)
	if whereClause != "" {
		query += " WHERE " + whereClause
	}

	query += " ORDER BY created_at DESC"

	// Add pagination
	argIndex := len(args) + 1
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	return query, args
}

// buildCountQuery constructs the SQL query for counting users with filters
func (r *userRepositoryImpl) buildCountQuery(filter UserFilter) (string, []interface{}) {
	query := "SELECT COUNT(*) FROM users"

	whereClause, args := r.buildWhereClause(filter)
	if whereClause != "" {
		query += " WHERE " + whereClause
	}

	return query, args
}

// buildWhereClause constructs the WHERE clause for filtering
func (r *userRepositoryImpl) buildWhereClause(filter UserFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, string(*filter.Status))
		argIndex++
	}

	if filter.Email != nil {
		conditions = append(conditions, fmt.Sprintf("email ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Email+"%")
		argIndex++
	}

	if filter.FirstName != nil {
		conditions = append(conditions, fmt.Sprintf("first_name ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.FirstName+"%")
		argIndex++
	}

	if filter.LastName != nil {
		conditions = append(conditions, fmt.Sprintf("last_name ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.LastName+"%")
		argIndex++
	}

	if filter.CreatedAfter != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, *filter.CreatedAfter)
		argIndex++
	}

	if filter.CreatedBefore != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, *filter.CreatedBefore)
		argIndex++
	}

	return strings.Join(conditions, " AND "), args
}

// handleError converts database errors to appropriate domain errors
func (r *userRepositoryImpl) handleError(operation string, err error) error {
	if err == nil {
		return nil
	}

	// Handle PostgreSQL specific errors
	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Code {
		case "23505": // unique_violation
			if strings.Contains(pqErr.Detail, "email") {
				return database.NewDatabaseErrorWithCode(operation, "users", "23505", "email already exists", database.ErrDuplicateKey)
			}
			return database.NewDatabaseErrorWithCode(operation, "users", "23505", "duplicate key violation", database.ErrDuplicateKey)
		case "23503": // foreign_key_violation
			return database.NewDatabaseErrorWithCode(operation, "users", "23503", "foreign key constraint violation", database.ErrForeignKeyViolation)
		case "23514": // check_violation
			return database.NewDatabaseErrorWithCode(operation, "users", "23514", "check constraint violation", database.ErrInvalidInput)
		}
	}

	// Handle common database errors
	if err == sql.ErrNoRows {
		return database.ErrNotFound
	}

	// Wrap other errors
	return database.NewDatabaseError(operation, "users", err)
}
