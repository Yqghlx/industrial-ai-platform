package database

import (
	"context"
)

// TransactionHelper provides utility functions for transaction management
type TransactionHelper struct {
	db DatabaseInterface
}

// NewTransactionHelper creates a new transaction helper
func NewTransactionHelper(db DatabaseInterface) *TransactionHelper {
	return &TransactionHelper{db: db}
}

// WithTransaction executes a function within a transaction
// If the function returns an error, the transaction is rolled back
// If the function returns nil, the transaction is committed
func (h *TransactionHelper) WithTransaction(ctx context.Context, fn func(tx TransactionInterface) error) error {
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Ensure rollback on error
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Execute function
	if err = fn(tx); err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit()
}

// WithTransactionResult executes a function within a transaction and returns a result
func (h *TransactionHelper) WithTransactionResult(ctx context.Context, fn func(tx TransactionInterface) (interface{}, error)) (interface{}, error) {
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Ensure rollback on error
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Execute function
	result, err := fn(tx)
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return result, nil
}
