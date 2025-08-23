package database

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteInTransaction_Success(t *testing.T) {
	SkipIfNoDatabase(t)
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewTestRepository(db)
	ctx := context.Background()

	var createdEntity *TestEntity

	err := ExecuteInTransaction(ctx, db, func(txCtx context.Context) error {
		// Create entity within transaction
		entity := &TestEntity{
			Name:  "Transaction Test",
			Email: "transaction@example.com",
		}

		if err := repo.Create(txCtx, entity); err != nil {
			return err
		}

		createdEntity = entity
		return nil
	})

	assert.NoError(t, err)
	assert.NotNil(t, createdEntity)
	assert.NotZero(t, createdEntity.ID)

	// Verify entity was committed
	retrieved, err := repo.GetByID(ctx, createdEntity.ID)
	assert.NoError(t, err)
	assert.Equal(t, createdEntity.Name, retrieved.Name)
}

func TestExecuteInTransaction_Rollback(t *testing.T) {
	SkipIfNoDatabase(t)
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewTestRepository(db)
	ctx := context.Background()

	var entityID int

	err := ExecuteInTransaction(ctx, db, func(txCtx context.Context) error {
		// Create entity within transaction
		entity := &TestEntity{
			Name:  "Rollback Test",
			Email: "rollback@example.com",
		}

		if err := repo.Create(txCtx, entity); err != nil {
			return err
		}

		entityID = entity.ID

		// Force an error to trigger rollback
		return errors.New("forced error")
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forced error")

	// Verify entity was rolled back
	_, err = repo.GetByID(ctx, entityID)
	assert.Error(t, err)
	assert.True(t, IsNotFoundError(err))
}

func TestExecuteInTransaction_NestedTransaction(t *testing.T) {
	SkipIfNoDatabase(t)
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewTestRepository(db)
	ctx := context.Background()

	var entity1, entity2 *TestEntity

	err := ExecuteInTransaction(ctx, db, func(txCtx context.Context) error {
		// Create first entity
		entity1 = &TestEntity{
			Name:  "Nested Test 1",
			Email: "nested1@example.com",
		}

		if err := repo.Create(txCtx, entity1); err != nil {
			return err
		}

		// Nested transaction (should use same transaction)
		return ExecuteInTransaction(txCtx, db, func(nestedCtx context.Context) error {
			entity2 = &TestEntity{
				Name:  "Nested Test 2",
				Email: "nested2@example.com",
			}

			return repo.Create(nestedCtx, entity2)
		})
	})

	assert.NoError(t, err)
	assert.NotNil(t, entity1)
	assert.NotNil(t, entity2)

	// Verify both entities were committed
	retrieved1, err := repo.GetByID(ctx, entity1.ID)
	assert.NoError(t, err)
	assert.Equal(t, entity1.Name, retrieved1.Name)

	retrieved2, err := repo.GetByID(ctx, entity2.ID)
	assert.NoError(t, err)
	assert.Equal(t, entity2.Name, retrieved2.Name)
}

func TestExecuteInTransaction_NestedRollback(t *testing.T) {
	SkipIfNoDatabase(t)
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewTestRepository(db)
	ctx := context.Background()

	var entity1ID, entity2ID int

	err := ExecuteInTransaction(ctx, db, func(txCtx context.Context) error {
		// Create first entity
		entity1 := &TestEntity{
			Name:  "Nested Rollback 1",
			Email: "nestedrollback1@example.com",
		}

		if err := repo.Create(txCtx, entity1); err != nil {
			return err
		}
		entity1ID = entity1.ID

		// Nested transaction that fails
		return ExecuteInTransaction(txCtx, db, func(nestedCtx context.Context) error {
			entity2 := &TestEntity{
				Name:  "Nested Rollback 2",
				Email: "nestedrollback2@example.com",
			}

			if err := repo.Create(nestedCtx, entity2); err != nil {
				return err
			}
			entity2ID = entity2.ID

			// Force error to rollback entire transaction
			return errors.New("nested error")
		})
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nested error")

	// Verify both entities were rolled back
	_, err = repo.GetByID(ctx, entity1ID)
	assert.Error(t, err)
	assert.True(t, IsNotFoundError(err))

	_, err = repo.GetByID(ctx, entity2ID)
	assert.Error(t, err)
	assert.True(t, IsNotFoundError(err))
}

func TestTransactionManager(t *testing.T) {
	SkipIfNoDatabase(t)
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	tm := NewTransactionManager(db)
	repo := NewTestRepository(db)
	ctx := context.Background()

	var createdEntity *TestEntity

	err := tm.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		entity := &TestEntity{
			Name:  "Manager Test",
			Email: "manager@example.com",
		}

		if err := repo.Create(txCtx, entity); err != nil {
			return err
		}

		createdEntity = entity
		return nil
	})

	assert.NoError(t, err)
	assert.NotNil(t, createdEntity)

	// Verify entity was committed
	retrieved, err := repo.GetByID(ctx, createdEntity.ID)
	assert.NoError(t, err)
	assert.Equal(t, createdEntity.Name, retrieved.Name)
}

func TestTransactionalRepository(t *testing.T) {
	SkipIfNoDatabase(t)
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewTestRepository(db)
	tm := NewTransactionManager(db)
	txRepo := NewTransactionalRepository[TestEntity, int](repo, tm)
	ctx := context.Background()

	var entity1, entity2 *TestEntity

	err := txRepo.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		// Create first entity
		entity1 = &TestEntity{
			Name:  "TxRepo Test 1",
			Email: "txrepo1@example.com",
		}

		if err := repo.Create(txCtx, entity1); err != nil {
			return err
		}

		// Create second entity
		entity2 = &TestEntity{
			Name:  "TxRepo Test 2",
			Email: "txrepo2@example.com",
		}

		return repo.Create(txCtx, entity2)
	})

	assert.NoError(t, err)
	assert.NotNil(t, entity1)
	assert.NotNil(t, entity2)

	// Verify both entities were committed
	retrieved1, err := repo.GetByID(ctx, entity1.ID)
	assert.NoError(t, err)
	assert.Equal(t, entity1.Name, retrieved1.Name)

	retrieved2, err := repo.GetByID(ctx, entity2.ID)
	assert.NoError(t, err)
	assert.Equal(t, entity2.Name, retrieved2.Name)
}

func TestExecuteInTransactionWithOptions(t *testing.T) {
	SkipIfNoDatabase(t)
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	repo := NewTestRepository(db)
	ctx := context.Background()

	var createdEntity *TestEntity

	// Test with read-only transaction options
	err := ExecuteInTransactionWithOptions(ctx, db, TxOptions.ReadOnly(), func(txCtx context.Context) error {
		// This should work for read operations
		count, err := repo.Count(txCtx)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
		return nil
	})

	assert.NoError(t, err)

	// Test with read-write transaction
	err = ExecuteInTransactionWithOptions(ctx, db, TxOptions.ReadCommitted(), func(txCtx context.Context) error {
		entity := &TestEntity{
			Name:  "Options Test",
			Email: "options@example.com",
		}

		if err := repo.Create(txCtx, entity); err != nil {
			return err
		}

		createdEntity = entity
		return nil
	})

	assert.NoError(t, err)
	assert.NotNil(t, createdEntity)

	// Verify entity was committed
	retrieved, err := repo.GetByID(ctx, createdEntity.ID)
	assert.NoError(t, err)
	assert.Equal(t, createdEntity.Name, retrieved.Name)
}

func TestGetTxFromContext(t *testing.T) {
	SkipIfNoDatabase(t)
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	ctx := context.Background()

	// No transaction in context
	tx := GetTxFromContext(ctx)
	assert.Nil(t, tx)

	// Transaction in context
	err := ExecuteInTransaction(ctx, db, func(txCtx context.Context) error {
		tx := GetTxFromContext(txCtx)
		assert.NotNil(t, tx)
		return nil
	})

	assert.NoError(t, err)
}

func TestWithTransaction(t *testing.T) {
	SkipIfNoDatabase(t)
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	ctx := context.Background()

	// Start a transaction manually
	tx, err := db.BeginTxx(ctx, nil)
	require.NoError(t, err)
	defer tx.Rollback()

	// Add transaction to context
	txCtx := WithTransaction(ctx, tx)

	// Verify transaction is in context
	retrievedTx := GetTxFromContext(txCtx)
	assert.Equal(t, tx, retrievedTx)
}
