package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/postgres/sqlc"
	"github.com/lib/pq"
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

// isSerializationError reports whether err is a PostgreSQL serialization failure.
func isSerializationError(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "40001"
}

// ExecTx runs fn inside a transaction and handles rollback on error.
// Serialization failures (SQLSTATE 40001) are automatically retried up to maxAttempts times.
func (store *Store) ExecTx(ctx context.Context, fn func(q *sqlc.Queries) error) error {
	const maxAttempts = 6
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		lastErr = store.execTxOnce(ctx, fn)
		if lastErr == nil {
			return nil
		}
		if !isSerializationError(lastErr) {
			return lastErr
		}
	}
	return fmt.Errorf("transaction failed after %d attempts due to serialization conflicts: %w", maxAttempts, lastErr)
}

func (store *Store) execTxOnce(ctx context.Context, fn func(q *sqlc.Queries) error) error {
	tx, err := store.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable}) // good default for money ops
	if err != nil {
		return err
	}

	q := sqlc.New(tx)
	if err := fn(q); err != nil {
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
