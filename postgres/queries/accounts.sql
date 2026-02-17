-- name: CreateAccount :one
INSERT INTO accounts (owner_id, name, currency, is_system)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetAccount :one
SELECT * FROM accounts
WHERE id = $1
LIMIT 1;

-- name: GetAccountForUpdate :one
SELECT * FROM accounts
WHERE id = $1
LIMIT 1
FOR UPDATE; -- locks row for update, prevents TOCTOU races

-- name: ListAccountsByOwner :many
SELECT * FROM accounts
WHERE owner_id = $1
ORDER BY created_at DESC;

-- name: UpdateAccountBalance :exec
UPDATE accounts
SET balance = balance + $1
WHERE id = $2;

-- name: GetSettlementAccount :one
SELECT * FROM accounts
WHERE is_system = TRUE AND name = 'Settlement Account'
LIMIT 1;