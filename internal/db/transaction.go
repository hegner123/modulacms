// Package db provides database types and utilities shared across all database drivers.
package db

import (
	"context"
	"database/sql"
	"fmt"
)

// TxFunc executes within a transaction
type TxFunc func(tx *sql.Tx) error

// WithTransaction executes fn within a transaction with automatic commit/rollback.
// If fn returns an error, the transaction is rolled back.
// If fn succeeds, the transaction is committed.
func WithTransaction(ctx context.Context, db *sql.DB, fn TxFunc) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() // no-op if already committed

	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// WithTransactionResult executes fn within a transaction and returns a result.
// If fn returns an error, the transaction is rolled back and the zero value of T is returned.
// If fn succeeds, the transaction is committed and the result is returned.
func WithTransactionResult[T any](ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) (T, error)) (T, error) {
	var result T
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return result, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	result, err = fn(tx)
	if err != nil {
		return result, err
	}
	if err := tx.Commit(); err != nil {
		return result, fmt.Errorf("commit transaction: %w", err)
	}
	return result, nil
}
