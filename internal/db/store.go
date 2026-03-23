// Package db provides transaction-aware access to sqlc queries.
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"

	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/postgres/sqlc"
)

// Store wraps generated queries and transaction helpers.
type Store struct {
	*sqlc.Queries
	db *sql.DB
}

// NewStore constructs a Store backed by the given database connection.
func NewStore(db *sql.DB) *Store {
	return &Store{
		Queries: sqlc.New(db),
		db:      db,
	}
}

// isSerializationError reports whether err is a PostgreSQL serialization failure.
func isSerializationError(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "40001"
}

// ExecTx runs fn inside a transaction and handles rollback on error.
// Serialization failures (SQLSTATE 40001) are automatically retried up to maxAttempts times.
func (store *Store) ExecTx(ctx context.Context, fn func(q *sqlc.Queries) error) error {
	const maxAttempts = 10
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Run one serializable transaction attempt.
		lastErr = store.execTxOnce(ctx, fn)
		if lastErr == nil {
			return nil
		}
		if !isSerializationError(lastErr) {
			// Non-retryable errors should bubble up immediately.
			return lastErr
		}
		if attempt < maxAttempts-1 {
			// Back off before retrying to reduce repeated contention.
			if waitErr := sleepWithContext(ctx, retryWait(attempt)); waitErr != nil {
				return waitErr
			}
		}
	}
	return fmt.Errorf("transaction failed after %d attempts due to serialization conflicts: %w", maxAttempts, lastErr)
}

func (store *Store) execTxOnce(ctx context.Context, fn func(q *sqlc.Queries) error) error {
	// Use serializable isolation to protect balance-changing flows from race anomalies.
	tx, err := store.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}

	// Bind sqlc queries to this transaction handle.
	q := sqlc.New(tx)
	if err := fn(q); err != nil {
		// Always rollback on business/query failure.
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx failed: %w, rollback failed: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}

// retryWait returns a capped exponential backoff duration for the given attempt (0-based).
func retryWait(attempt int) time.Duration {
	// Exponential backoff: 50ms, 100ms, 200ms ... capped at 1s.
	base := 50 * time.Millisecond
	for i := 0; i < attempt; i++ {
		base *= 2
		if base >= time.Second {
			return time.Second
		}
	}
	return base
}

// sleepWithContext waits for d or until ctx is cancelled.
func sleepWithContext(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}
