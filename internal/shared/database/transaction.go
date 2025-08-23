package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// TxKey is the context key for database transactions
type TxKey struct{}

// TransactionManager handles database transactions
type TransactionManager struct {
	db *DB
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(db *DB) *TransactionManager {
	return &TransactionManager{db: db}
}

// ExecuteInTransaction executes a function within a database transaction
// If the function returns an error, the transaction is rolled back
// Otherwise, the transaction is committed
func (tm *TransactionManager) ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return ExecuteInTransaction(ctx, tm.db, fn)
}

// ExecuteInTransaction is a standalone function for executing operations in a transaction
func ExecuteInTransaction(ctx context.Context, db *DB, fn func(ctx context.Context) error) error {
	// Check if we're already in a transaction
	if GetTxFromContext(ctx) != nil {
		// Already in a transaction, just execute the function
		return fn(ctx)
	}

	// Start a new transaction
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Add transaction to context
	txCtx := context.WithValue(ctx, TxKey{}, tx)

	// Set up defer for rollback in case of panic
	defer func() {
		if r := recover(); r != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				// Log rollback error but don't override the panic
				fmt.Printf("failed to rollback transaction after panic: %v\n", rollbackErr)
			}
			panic(r) // Re-panic
		}
	}()

	// Execute the function
	if err := fn(txCtx); err != nil {
		// Function returned an error, rollback the transaction
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("failed to rollback transaction: %v (original error: %w)", rollbackErr, err)
		}
		return err
	}

	// Function succeeded, commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ExecuteInTransactionWithOptions executes a function within a transaction with custom options
func ExecuteInTransactionWithOptions(ctx context.Context, db *DB, opts *sql.TxOptions, fn func(ctx context.Context) error) error {
	// Check if we're already in a transaction
	if GetTxFromContext(ctx) != nil {
		// Already in a transaction, just execute the function
		return fn(ctx)
	}

	// Start a new transaction with options
	tx, err := db.BeginTxx(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Add transaction to context
	txCtx := context.WithValue(ctx, TxKey{}, tx)

	// Set up defer for rollback in case of panic
	defer func() {
		if r := recover(); r != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				// Log rollback error but don't override the panic
				fmt.Printf("failed to rollback transaction after panic: %v\n", rollbackErr)
			}
			panic(r) // Re-panic
		}
	}()

	// Execute the function
	if err := fn(txCtx); err != nil {
		// Function returned an error, rollback the transaction
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("failed to rollback transaction: %v (original error: %w)", rollbackErr, err)
		}
		return err
	}

	// Function succeeded, commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetTxFromContext retrieves the transaction from the context
func GetTxFromContext(ctx context.Context) *sqlx.Tx {
	if tx, ok := ctx.Value(TxKey{}).(*sqlx.Tx); ok {
		return tx
	}
	return nil
}

// WithTransaction adds a transaction to the context
func WithTransaction(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, TxKey{}, tx)
}

// TransactionOptions provides common transaction option presets
type TransactionOptions struct{}

// ReadOnly returns transaction options for read-only operations
func (TransactionOptions) ReadOnly() *sql.TxOptions {
	return &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  true,
	}
}

// ReadCommitted returns transaction options for read committed isolation
func (TransactionOptions) ReadCommitted() *sql.TxOptions {
	return &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	}
}

// RepeatableRead returns transaction options for repeatable read isolation
func (TransactionOptions) RepeatableRead() *sql.TxOptions {
	return &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  false,
	}
}

// Serializable returns transaction options for serializable isolation
func (TransactionOptions) Serializable() *sql.TxOptions {
	return &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}
}

// TxOptions provides access to transaction option presets
var TxOptions TransactionOptions

// Transactional is a helper interface for repositories that support transactions
type Transactional interface {
	ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// TransactionalRepository wraps a repository with transaction support
type TransactionalRepository[T any, ID comparable] struct {
	Repository[T, ID]
	tm *TransactionManager
}

// NewTransactionalRepository creates a new transactional repository wrapper
func NewTransactionalRepository[T any, ID comparable](repo Repository[T, ID], tm *TransactionManager) *TransactionalRepository[T, ID] {
	return &TransactionalRepository[T, ID]{
		Repository: repo,
		tm:         tm,
	}
}

// ExecuteInTransaction executes a function within a transaction
func (tr *TransactionalRepository[T, ID]) ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return tr.tm.ExecuteInTransaction(ctx, fn)
}
