package infrastructure

import (
	"context"
	"fmt"
	"testing"

	"go-templ-template/internal/modules/user/domain"
	"go-templ-template/internal/shared/database"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// UserRepositoryTestSuite provides integration tests for the user repository
type UserRepositoryTestSuite struct {
	suite.Suite
	db   *database.DB
	repo UserRepository
	ctx  context.Context
}

// SetupSuite initializes the test database and repository
func (suite *UserRepositoryTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Initialize test database
	testDB := database.NewTestDatabase(suite.T())

	suite.db = testDB.DB
	suite.repo = NewUserRepository(testDB.DB)

	// Create the users table for testing
	suite.createUsersTable()
}

// createUsersTable creates the users table for testing
func (suite *UserRepositoryTestSuite) createUsersTable() {
	createTableSQL := `
		-- Enable UUID extension
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

		-- Create users table
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			first_name VARCHAR(100) NOT NULL,
			last_name VARCHAR(100) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended')),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			version INTEGER DEFAULT 1
		);

		-- Create indexes for better performance
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
		CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
		CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

		-- Create function to update updated_at timestamp
		CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = NOW();
			NEW.version = OLD.version + 1;
			RETURN NEW;
		END;
		$$ language 'plpgsql';

		-- Create trigger for users table
		DROP TRIGGER IF EXISTS update_users_updated_at ON users;
		CREATE TRIGGER update_users_updated_at 
			BEFORE UPDATE ON users 
			FOR EACH ROW 
			EXECUTE FUNCTION update_updated_at_column();
	`

	_, err := suite.db.ExecContext(suite.ctx, createTableSQL)
	require.NoError(suite.T(), err)
}

// TearDownSuite cleans up the test database
func (suite *UserRepositoryTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
}

// SetupTest cleans the users table before each test
func (suite *UserRepositoryTestSuite) SetupTest() {
	_, err := suite.db.ExecContext(suite.ctx, "TRUNCATE TABLE users CASCADE")
	require.NoError(suite.T(), err)
}

// TestCreate tests user creation functionality
func (suite *UserRepositoryTestSuite) TestCreate() {
	// Create a valid user
	user, err := domain.NewUser(
		uuid.New().String(),
		"test@example.com",
		"Password123",
		"John",
		"Doe",
	)
	require.NoError(suite.T(), err)

	// Test successful creation
	err = suite.repo.Create(suite.ctx, user)
	assert.NoError(suite.T(), err)

	// Verify user was created
	retrieved, err := suite.repo.GetByID(suite.ctx, user.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), user.ID, retrieved.ID)
	assert.Equal(suite.T(), user.Email, retrieved.Email)
	assert.Equal(suite.T(), user.FirstName, retrieved.FirstName)
	assert.Equal(suite.T(), user.LastName, retrieved.LastName)
	assert.Equal(suite.T(), user.Status, retrieved.Status)
}

// TestCreateDuplicateEmail tests duplicate email constraint
func (suite *UserRepositoryTestSuite) TestCreateDuplicateEmail() {
	email := "duplicate@example.com"

	// Create first user
	user1, err := domain.NewUser(
		uuid.New().String(),
		email,
		"Password123",
		"John",
		"Doe",
	)
	require.NoError(suite.T(), err)

	err = suite.repo.Create(suite.ctx, user1)
	assert.NoError(suite.T(), err)

	// Try to create second user with same email
	user2, err := domain.NewUser(
		uuid.New().String(),
		email,
		"Password456",
		"Jane",
		"Smith",
	)
	require.NoError(suite.T(), err)

	err = suite.repo.Create(suite.ctx, user2)
	assert.Error(suite.T(), err)
	assert.True(suite.T(), database.IsDuplicateKeyError(err))
}

// TestGetByID tests retrieving users by ID
func (suite *UserRepositoryTestSuite) TestGetByID() {
	// Create a user
	user := suite.createTestUser("getbyid@example.com")

	// Test successful retrieval
	retrieved, err := suite.repo.GetByID(suite.ctx, user.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), user.ID, retrieved.ID)
	assert.Equal(suite.T(), user.Email, retrieved.Email)

	// Test non-existent user
	_, err = suite.repo.GetByID(suite.ctx, uuid.New().String())
	assert.Error(suite.T(), err)
	assert.True(suite.T(), database.IsNotFoundError(err))
}

// TestGetByEmail tests retrieving users by email
func (suite *UserRepositoryTestSuite) TestGetByEmail() {
	email := "getbyemail@example.com"
	user := suite.createTestUser(email)

	// Test successful retrieval
	retrieved, err := suite.repo.GetByEmail(suite.ctx, email)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), user.ID, retrieved.ID)
	assert.Equal(suite.T(), user.Email, retrieved.Email)

	// Test case insensitive retrieval
	retrieved, err = suite.repo.GetByEmail(suite.ctx, "GETBYEMAIL@EXAMPLE.COM")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), user.ID, retrieved.ID)

	// Test non-existent email
	_, err = suite.repo.GetByEmail(suite.ctx, "nonexistent@example.com")
	assert.Error(suite.T(), err)
	assert.True(suite.T(), database.IsNotFoundError(err))
}

// TestUpdate tests user update functionality
func (suite *UserRepositoryTestSuite) TestUpdate() {
	// Create a user
	user := suite.createTestUser("update@example.com")
	originalVersion := user.Version

	// Update user profile
	err := user.UpdateProfile("Jane", "Smith")
	require.NoError(suite.T(), err)

	// Test successful update
	err = suite.repo.Update(suite.ctx, user)
	assert.NoError(suite.T(), err)

	// Verify update
	retrieved, err := suite.repo.GetByID(suite.ctx, user.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Jane", retrieved.FirstName)
	assert.Equal(suite.T(), "Smith", retrieved.LastName)
	assert.Equal(suite.T(), originalVersion+1, retrieved.Version)
}

// TestUpdateOptimisticLocking tests optimistic locking during updates
func (suite *UserRepositoryTestSuite) TestUpdateOptimisticLocking() {
	// Create a user
	user := suite.createTestUser("optimistic@example.com")

	// Get two instances of the same user
	user1, err := suite.repo.GetByID(suite.ctx, user.ID)
	require.NoError(suite.T(), err)

	user2, err := suite.repo.GetByID(suite.ctx, user.ID)
	require.NoError(suite.T(), err)

	// Update first instance
	err = user1.UpdateProfile("First", "Update")
	require.NoError(suite.T(), err)
	err = suite.repo.Update(suite.ctx, user1)
	assert.NoError(suite.T(), err)

	// Try to update second instance (should fail due to version conflict)
	err = user2.UpdateProfile("Second", "Update")
	require.NoError(suite.T(), err)
	err = suite.repo.Update(suite.ctx, user2)
	assert.Error(suite.T(), err)
	assert.True(suite.T(), database.IsOptimisticLockError(err))
}

// TestUpdateNonExistentUser tests updating a non-existent user
func (suite *UserRepositoryTestSuite) TestUpdateNonExistentUser() {
	user, err := domain.NewUser(
		uuid.New().String(),
		"nonexistent@example.com",
		"Password123",
		"John",
		"Doe",
	)
	require.NoError(suite.T(), err)

	err = suite.repo.Update(suite.ctx, user)
	assert.Error(suite.T(), err)
	assert.True(suite.T(), database.IsNotFoundError(err))
}

// TestDelete tests user deletion functionality
func (suite *UserRepositoryTestSuite) TestDelete() {
	// Create a user
	user := suite.createTestUser("delete@example.com")

	// Verify user exists
	exists, err := suite.repo.Exists(suite.ctx, user.ID)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)

	// Delete user
	err = suite.repo.Delete(suite.ctx, user.ID)
	assert.NoError(suite.T(), err)

	// Verify user no longer exists
	exists, err = suite.repo.Exists(suite.ctx, user.ID)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), exists)

	// Try to delete again (should return not found error)
	err = suite.repo.Delete(suite.ctx, user.ID)
	assert.Error(suite.T(), err)
	assert.True(suite.T(), database.IsNotFoundError(err))
}

// TestList tests user listing with pagination and filtering
func (suite *UserRepositoryTestSuite) TestList() {
	// Create test users
	users := suite.createMultipleTestUsers(5)

	// Test listing all users
	filter := UserFilter{}
	retrieved, err := suite.repo.List(suite.ctx, filter, 10, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), retrieved, 5)

	// Test pagination
	retrieved, err = suite.repo.List(suite.ctx, filter, 2, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), retrieved, 2)

	retrieved, err = suite.repo.List(suite.ctx, filter, 2, 2)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), retrieved, 2)

	// Test filtering by status
	activeStatus := domain.UserStatusActive
	filter = UserFilter{Status: &activeStatus}
	retrieved, err = suite.repo.List(suite.ctx, filter, 10, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), retrieved, 5) // All test users are active

	// Suspend one user and test filtering
	err = users[0].Suspend()
	require.NoError(suite.T(), err)
	err = suite.repo.Update(suite.ctx, users[0])
	require.NoError(suite.T(), err)

	suspendedStatus := domain.UserStatusSuspended
	filter = UserFilter{Status: &suspendedStatus}
	retrieved, err = suite.repo.List(suite.ctx, filter, 10, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), retrieved, 1)
}

// TestListWithFilters tests advanced filtering options
func (suite *UserRepositoryTestSuite) TestListWithFilters() {
	// Create users with different attributes
	user1 := suite.createTestUserWithDetails("alice@example.com", "Alice", "Johnson")
	user2 := suite.createTestUserWithDetails("bob@example.com", "Bob", "Smith")
	user3 := suite.createTestUserWithDetails("charlie@example.com", "Charlie", "Johnson")

	// Test email filtering
	email := "alice"
	filter := UserFilter{Email: &email}
	retrieved, err := suite.repo.List(suite.ctx, filter, 10, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), retrieved, 1)
	assert.Equal(suite.T(), user1.ID, retrieved[0].ID)

	// Test first name filtering
	firstName := "Bob"
	filter = UserFilter{FirstName: &firstName}
	retrieved, err = suite.repo.List(suite.ctx, filter, 10, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), retrieved, 1)
	assert.Equal(suite.T(), user2.ID, retrieved[0].ID)

	// Test last name filtering
	lastName := "Johnson"
	filter = UserFilter{LastName: &lastName}
	retrieved, err = suite.repo.List(suite.ctx, filter, 10, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), retrieved, 2)

	// Verify both Alice and Charlie are returned
	ids := []string{retrieved[0].ID, retrieved[1].ID}
	assert.Contains(suite.T(), ids, user1.ID)
	assert.Contains(suite.T(), ids, user3.ID)
}

// TestCount tests user counting functionality
func (suite *UserRepositoryTestSuite) TestCount() {
	// Create test users
	suite.createMultipleTestUsers(3)

	// Test counting all users
	filter := UserFilter{}
	count, err := suite.repo.Count(suite.ctx, filter)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(3), count)

	// Test counting with status filter
	activeStatus := domain.UserStatusActive
	filter = UserFilter{Status: &activeStatus}
	count, err = suite.repo.Count(suite.ctx, filter)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(3), count)

	// Test counting with non-matching filter
	suspendedStatus := domain.UserStatusSuspended
	filter = UserFilter{Status: &suspendedStatus}
	count, err = suite.repo.Count(suite.ctx, filter)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(0), count)
}

// TestExists tests user existence checking
func (suite *UserRepositoryTestSuite) TestExists() {
	// Create a user
	user := suite.createTestUser("exists@example.com")

	// Test existing user
	exists, err := suite.repo.Exists(suite.ctx, user.ID)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)

	// Test non-existent user
	exists, err = suite.repo.Exists(suite.ctx, uuid.New().String())
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), exists)
}

// TestExistsByEmail tests user existence checking by email
func (suite *UserRepositoryTestSuite) TestExistsByEmail() {
	email := "existsbyemail@example.com"
	suite.createTestUser(email)

	// Test existing email
	exists, err := suite.repo.ExistsByEmail(suite.ctx, email)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)

	// Test case insensitive check
	exists, err = suite.repo.ExistsByEmail(suite.ctx, "EXISTSBYEMAIL@EXAMPLE.COM")
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)

	// Test non-existent email
	exists, err = suite.repo.ExistsByEmail(suite.ctx, "nonexistent@example.com")
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), exists)
}

// TestTransactionSupport tests repository operations within transactions
func (suite *UserRepositoryTestSuite) TestTransactionSupport() {
	// Test successful transaction
	err := database.ExecuteInTransaction(suite.ctx, suite.db, func(ctx context.Context) error {
		user1 := suite.createTestUserInContext(ctx, "tx1@example.com")
		user2 := suite.createTestUserInContext(ctx, "tx2@example.com")

		// Both users should be created within the transaction
		exists1, err := suite.repo.Exists(ctx, user1.ID)
		if err != nil {
			return err
		}
		assert.True(suite.T(), exists1)

		exists2, err := suite.repo.Exists(ctx, user2.ID)
		if err != nil {
			return err
		}
		assert.True(suite.T(), exists2)

		return nil
	})
	assert.NoError(suite.T(), err)

	// Test transaction rollback
	var user1ID, user2ID string
	err = database.ExecuteInTransaction(suite.ctx, suite.db, func(ctx context.Context) error {
		user1 := suite.createTestUserInContext(ctx, "rollback1@example.com")
		user2 := suite.createTestUserInContext(ctx, "rollback2@example.com")

		user1ID = user1.ID
		user2ID = user2.ID

		// Force an error to trigger rollback
		return assert.AnError
	})
	assert.Error(suite.T(), err)

	// Verify users were not created due to rollback
	exists1, err := suite.repo.Exists(suite.ctx, user1ID)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), exists1)

	exists2, err := suite.repo.Exists(suite.ctx, user2ID)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), exists2)
}

// Helper methods

// createTestUser creates a test user with default values
func (suite *UserRepositoryTestSuite) createTestUser(email string) *domain.User {
	return suite.createTestUserWithDetails(email, "Test", "User")
}

// createTestUserWithDetails creates a test user with specified details
func (suite *UserRepositoryTestSuite) createTestUserWithDetails(email, firstName, lastName string) *domain.User {
	user, err := domain.NewUser(
		uuid.New().String(),
		email,
		"Password123",
		firstName,
		lastName,
	)
	require.NoError(suite.T(), err)

	err = suite.repo.Create(suite.ctx, user)
	require.NoError(suite.T(), err)

	return user
}

// createTestUserInContext creates a test user within a specific context (for transactions)
func (suite *UserRepositoryTestSuite) createTestUserInContext(ctx context.Context, email string) *domain.User {
	user, err := domain.NewUser(
		uuid.New().String(),
		email,
		"Password123",
		"Test",
		"User",
	)
	require.NoError(suite.T(), err)

	err = suite.repo.Create(ctx, user)
	require.NoError(suite.T(), err)

	return user
}

// createMultipleTestUsers creates multiple test users
func (suite *UserRepositoryTestSuite) createMultipleTestUsers(count int) []*domain.User {
	users := make([]*domain.User, count)
	for i := 0; i < count; i++ {
		email := fmt.Sprintf("user%d@example.com", i)
		users[i] = suite.createTestUser(email)
	}
	return users
}

// TestUserRepository runs the user repository test suite
func TestUserRepository(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}
