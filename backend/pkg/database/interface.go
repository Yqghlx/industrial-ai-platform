package database

import (
	"context"
	"database/sql"
)

// QueryExecutor defines the interface for query execution
// Both DatabaseInterface and TransactionInterface implement this
type QueryExecutor interface {
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
	Query(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) *sql.Row
}

// DatabaseInterface defines the interface for database operations
// This allows repositories to depend on an interface rather than concrete sql.DB
type DatabaseInterface interface {
	QueryExecutor
	// Transaction support
	BeginTx(ctx context.Context, opts *sql.TxOptions) (TransactionInterface, error)
	// Connection management
	Ping(ctx context.Context) error
	Stats() sql.DBStats
}

// TransactionInterface defines the interface for transaction operations
type TransactionInterface interface {
	QueryExecutor
	Commit() error
	Rollback() error
}

// DBWrapper wraps sql.DB to implement DatabaseInterface
type DBWrapper struct {
	*sql.DB
}

// NewDBWrapper creates a new DBWrapper
func NewDBWrapper(db *sql.DB) *DBWrapper {
	return &DBWrapper{DB: db}
}

// Exec implements DatabaseInterface
func (w *DBWrapper) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return w.DB.ExecContext(ctx, query, args...)
}

// Query implements DatabaseInterface
func (w *DBWrapper) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return w.DB.QueryContext(ctx, query, args...)
}

// QueryRow implements DatabaseInterface
func (w *DBWrapper) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return w.DB.QueryRowContext(ctx, query, args...)
}

// BeginTx implements DatabaseInterface
func (w *DBWrapper) BeginTx(ctx context.Context, opts *sql.TxOptions) (TransactionInterface, error) {
	tx, err := w.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &TxWrapper{Tx: tx}, nil
}

// Ping implements DatabaseInterface
func (w *DBWrapper) Ping(ctx context.Context) error {
	return w.DB.PingContext(ctx)
}

// Stats implements DatabaseInterface
func (w *DBWrapper) Stats() sql.DBStats {
	return w.DB.Stats()
}

// TxWrapper wraps sql.Tx to implement TransactionInterface
type TxWrapper struct {
	*sql.Tx
}

// Exec implements TransactionInterface
func (w *TxWrapper) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return w.Tx.ExecContext(ctx, query, args...)
}

// Query implements TransactionInterface
func (w *TxWrapper) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return w.Tx.QueryContext(ctx, query, args...)
}

// QueryRow implements TransactionInterface
func (w *TxWrapper) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return w.Tx.QueryRowContext(ctx, query, args...)
}

// Commit implements TransactionInterface
func (w *TxWrapper) Commit() error {
	return w.Tx.Commit()
}

// Rollback implements TransactionInterface
func (w *TxWrapper) Rollback() error {
	return w.Tx.Rollback()
}