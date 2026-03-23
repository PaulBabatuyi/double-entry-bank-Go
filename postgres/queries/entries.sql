-- name: CreateEntry :one
INSERT INTO entries (account_id, debit, credit, transaction_id, operation_type, description)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListEntriesByAccount :many
SELECT * FROM entries
WHERE account_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListEntriesByTransaction :many
SELECT * FROM entries
WHERE transaction_id = $1
ORDER BY created_at;