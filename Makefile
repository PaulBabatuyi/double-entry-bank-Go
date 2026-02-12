.PHONY: postgres createdb migrate-up migrate-down sqlc test server

postgres:
	docker compose up -d

createdb:
	docker compose exec db createdb --username=root --owner=root simple_ledger || true

migrate-up:
	migrate -path postgres/migrations/ -database "postgresql://root:secret@localhost:5433/simple_ledger?sslmode=disable" -verbose up

migrate-down:
	migrate -path postgres/migrations/ -database "postgresql://root:secret@localhost:5433/simple_ledger?sslmode=disable" -verbose down

sqlc:
	sqlc generate   

test:
	go test -v ./...

server:
	go run cmd/server/main.go