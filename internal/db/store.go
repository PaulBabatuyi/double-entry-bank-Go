package db

import (
	"database/sql"

	"github.com/PaulBabatuyi/Double-Entry-Bank-Go/postgres/sqlc" // adjust import path
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
