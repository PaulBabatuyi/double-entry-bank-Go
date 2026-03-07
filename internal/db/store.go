package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/postgres/sqlc"
	"github.com/lib/pq"
)

const (
	maxRetries    = 5
	baseRetryWait = 50 * time.Millisecond
	maxRetryWait  = 2 * time.Second
)

type Store struct {
	*sqlc.Queries
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		Queries: sqlc.New(db),
		db:      db,
	}
}

// retryWait returns the backoff duration for the given attempt, capped at maxRetryWait.
func retryWait(attempt int) time.Duration {
	d := baseRetryWait * (1 << uint(attempt))
	if d > maxRetryWait {
		d = maxRetryWait
	}
	return d
}

// isSerializationError reports whether err (or any error it wraps) is a
// PostgreSQL serialization failure (SQLSTATE 40001).
func isSerializationError(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "40001"
}

// ExecTx runs fn inside a serializable transaction and handles rollback on
// error. It automatically retries up to maxRetries times when PostgreSQL
// returns a serialization failure (error code 40001).
func (store *Store) ExecTx(ctx context.Context, fn func(q *sqlc.Queries) error) error {
	for attempt := 0; attempt < maxRetries; attempt++ {
		tx, err := store.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
		if err != nil {
			return err
		}

		q := sqlc.New(tx)
		err = fn(q)
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				return fmt.Errorf("tx failed: %w, rollback failed: %v", err, rbErr)
			}
			if isSerializationError(err) && attempt < maxRetries-1 {
				if waitErr := sleepWithContext(ctx, retryWait(attempt)); waitErr != nil {
					return waitErr
				}
				continue
			}
			return err
		}

		if err := tx.Commit(); err != nil {
			if isSerializationError(err) && attempt < maxRetries-1 {
				if waitErr := sleepWithContext(ctx, retryWait(attempt)); waitErr != nil {
					return waitErr
				}
				continue
			}
			return fmt.Errorf("commit failed: %w", err)
		}
		return nil
	}
	return fmt.Errorf("transaction failed after %d attempts", maxRetries)
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
