package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/postgres/sqlc"
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

// ExecTx runs fn inside a transaction and handles rollback on error
func (store *Store) ExecTx(ctx context.Context, fn func(q *sqlc.Queries) error) error {
	tx, err := store.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable}) // good default for money ops
	if err != nil {
		return err
	}

	q := sqlc.New(tx)
	err = fn(q)
	if err != nil {
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
