.PHONY: postgres createdb migrate-up migrate-down sqlc test server lint test coverage

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

server:
	go run cmd/main.go

lint:
	golangci-lint run

test:
	go test -v -race ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out