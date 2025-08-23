package database

import (
	"context"
	"testing"
	"time"

	"go-templ-template/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEntity represents a test entity for repository testing
type TestEntity struct {
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Email     string    `db:"email" json:"email"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	Version   int       `db:"version" json:"version"`
}

// TestRepository implements Repository interface for TestEntity
type TestRepository struct {
	*BaseRepository[TestEntity, int]
}

// NewTestRepository creates a new test repository
func NewTestRepository(db *DB) *TestRepository {
	return &TestRepository{
		BaseRepository: NewBaseRepository[TestEntity, int](db, "test_entities", "id"),
	}
}

// Create inserts a new test entity
func (r *TestRepository) Create(ctx context.Context, entity *TestEntity) error {
	query := `
		INSERT INTO test_entities (name, email, created_at, updated_at, version)
		VALUES (:name, :email, NOW(), NOW(), 1)
		RETURNING id, created_at, updated_at`

	return r.BaseRepository.Create(ctx, entity, query)
}

// GetByID retrieves a test entity by ID
func (r *TestRepository) GetByID(ctx context.Context, id int) (*TestEntity, error) {
	query := `SELECT id, name, email, created_at, updated_at, version FROM test_entities WHERE id = $1`
	return r.BaseRepository.GetByID(ctx, id, query)
}

// Update updates a test entity with optimistic locking
func (r *TestRepository) Update(ctx context.Context, entity *TestEntity) error {
	query := `
		UPDATE test_entities 
		SET name = :name, email = :email, updated_at = NOW(), version = version + 1
		WHERE id = :id AND version = :version`

	return r.BaseRepository.Update(ctx, entity, query)
}

// Delete removes a test entity by ID
func (r *TestRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM test_entities WHERE id = $1`
	return r.BaseRepository.Delete(ctx, id, query)
}

// List retrieves test entities with pagination
func (r *TestRepository) List(ctx context.Context, limit, offset int) ([]*TestEntity, error) {
	query := `
		SELECT id, name, email, created_at, updated_at, version 
		FROM test_entities 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2`

	return r.BaseRepository.List(ctx, query, limit, offset)
}

// Count returns the total number of test entities
func (r *TestRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM test_entities`
	return r.BaseRepository.Count(ctx, query)
}

// Exists checks if a test entity exists by ID
func (r *TestRepository) Exists(ctx context.Context, id int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM test_entities WHERE id = $1)`
	return r.BaseRepository.Exists(ctx, id, query)
}

// GetByEmail retrieves a test entity by email
func (r *TestRepository) GetByEmail(ctx context.Context, email string) (*TestEntity, error) {
	var entity TestEntity
	query := `SELECT id, name, email, created_at, updated_at, version FROM test_entities WHERE email = $1`

	tx := GetTxFromContext(ctx)
	if tx != nil {
		err := tx.GetContext(ctx, &entity, query, email)
		if err != nil {
			return nil, err
		}
		return &entity, nil
	}

	err := r.GetDB().GetContext(ctx, &entity, query, email)
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// setupTestDatabase creates a test database and returns a cleanup function
func setupTestDatabase(t *testing.T) (*DB, func()) {
	// Use test database configuration
	cfg := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "postgres",
		Name:     "test_db",
		SSLMode:  "disable",
	}

	// Create database connection
	db, err := NewConnection(cfg, DefaultConnectionOptions())
	require.NoError(t, err)

	// Create test table
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS test_entities (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			version INTEGER NOT NULL DEFAULT 1
		)`

	_, err = db.Exec(createTableSQL)
	require.NoError(t, err)

	// Return cleanup function
	cleanup := func() {
		_, _ = db.Exec("DROP TABLE IF EXISTS test_entities")
		_ = db.Close()
	}

	return db, cleanup
}

func TestBaseRepository_Create(t *testing.T) {
	SkipIfNoDatabase(t)
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewTestRepository(db)
	ctx := context.Background()

	entity := &TestEntity{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	err := repo.Create(ctx, entity)
	assert.NoError(t, err)
	assert.NotZero(t, entity.ID)
	assert.NotZero(t, entity.CreatedAt)
	assert.NotZero(t, entity.UpdatedAt)
}

func TestBaseRepository_GetByID(t *testing.T) {
	SkipIfNoDatabase(t)
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewTestRepository(db)
	ctx := context.Background()

	// Create test entity
	entity := &TestEntity{
		Name:  "Jane Doe",
		Email: "jane@example.com",
	}
	err := repo.Create(ctx, entity)
	require.NoError(t, err)

	// Retrieve entity
	retrieved, err := repo.GetByID(ctx, entity.ID)
	assert.NoError(t, err)
	assert.Equal(t, entity.Name, retrieved.Name)
	assert.Equal(t, entity.Email, retrieved.Email)
}

func TestBaseRepository_GetByID_NotFound(t *testing.T) {
	SkipIfNoDatabase(t)
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewTestRepository(db)
	ctx := context.Background()

	// Try to retrieve non-existent entity
	_, err := repo.GetByID(ctx, 999)
	assert.Error(t, err)
	assert.True(t, IsNotFoundError(err))
}

func TestBaseRepository_Update(t *testing.T) {
	SkipIfNoDatabase(t)
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewTestRepository(db)
	ctx := context.Background()

	// Create test entity
	entity := &TestEntity{
		Name:  "John Doe",
		Email: "john@example.com",
	}
	err := repo.Create(ctx, entity)
	require.NoError(t, err)

	// Update entity
	entity.Name = "John Smith"
	err = repo.Update(ctx, entity)
	assert.NoError(t, err)

	// Verify update
	updated, err := repo.GetByID(ctx, entity.ID)
	require.NoError(t, err)
	assert.Equal(t, "John Smith", updated.Name)
	assert.Equal(t, entity.Version+1, updated.Version)
}

func TestBaseRepository_Delete(t *testing.T) {
	SkipIfNoDatabase(t)
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewTestRepository(db)
	ctx := context.Background()

	// Create test entity
	entity := &TestEntity{
		Name:  "John Doe",
		Email: "john@example.com",
	}
	err := repo.Create(ctx, entity)
	require.NoError(t, err)

	// Delete entity
	err = repo.Delete(ctx, entity.ID)
	assert.NoError(t, err)

	// Verify deletion
	_, err = repo.GetByID(ctx, entity.ID)
	assert.Error(t, err)
	assert.True(t, IsNotFoundError(err))
}

func TestBaseRepository_List(t *testing.T) {
	SkipIfNoDatabase(t)
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewTestRepository(db)
	ctx := context.Background()

	// Create test entities
	entities := []*TestEntity{
		{Name: "John Doe", Email: "john@example.com"},
		{Name: "Jane Doe", Email: "jane@example.com"},
		{Name: "Bob Smith", Email: "bob@example.com"},
	}

	for _, entity := range entities {
		err := repo.Create(ctx, entity)
		require.NoError(t, err)
	}

	// Test pagination
	results, err := repo.List(ctx, 2, 0)
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	results, err = repo.List(ctx, 2, 1)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestBaseRepository_Count(t *testing.T) {
	SkipIfNoDatabase(t)
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewTestRepository(db)
	ctx := context.Background()

	// Initial count should be 0
	count, err := repo.Count(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Create test entities
	entities := []*TestEntity{
		{Name: "John Doe", Email: "john@example.com"},
		{Name: "Jane Doe", Email: "jane@example.com"},
	}

	for _, entity := range entities {
		err := repo.Create(ctx, entity)
		require.NoError(t, err)
	}

	// Count should be 2
	count, err = repo.Count(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestBaseRepository_Exists(t *testing.T) {
	SkipIfNoDatabase(t)
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewTestRepository(db)
	ctx := context.Background()

	// Create test entity
	entity := &TestEntity{
		Name:  "John Doe",
		Email: "john@example.com",
	}
	err := repo.Create(ctx, entity)
	require.NoError(t, err)

	// Check existence
	exists, err := repo.Exists(ctx, entity.ID)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Check non-existence
	exists, err = repo.Exists(ctx, 999)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestQueryBuilder(t *testing.T) {
	qb := NewQueryBuilder()
	query, args := qb.
		Select("id, name, email").
		From("users").
		Where("status = ?", "active").
		And("created_at > ?", time.Now().AddDate(0, -1, 0)).
		OrderBy("created_at", "DESC").
		Limit(10).
		Offset(0).
		Build()

	assert.Contains(t, query, "SELECT id, name, email")
	assert.Contains(t, query, "FROM users")
	assert.Contains(t, query, "WHERE status = ?")
	assert.Contains(t, query, "AND created_at > ?")
	assert.Contains(t, query, "ORDER BY created_at DESC")
	assert.Contains(t, query, "LIMIT")
	assert.Contains(t, query, "OFFSET")
	assert.Len(t, args, 4) // status, created_at, limit, offset
}
