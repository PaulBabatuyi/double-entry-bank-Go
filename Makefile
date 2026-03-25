.PHONY: postgres createdb migrate-up migrate-down migrate-up-local migrate-down-local migrate-up-docker migrate-down-docker sqlc test server lint coverage ci-test docker-build docker-up docker-down clean help

DB_URL_DOCKER=postgresql://root:secret@127.0.0.1:5433/simple_ledger?sslmode=disable
DB_URL_LOCAL=postgresql://root:secret@127.0.0.1:5432/simple_ledger?sslmode=disable

postgres:
	docker compose up -d

createdb:
	docker compose exec db createdb --username=root --owner=root simple_ledger || true

migrate-up:
	migrate -path postgres/migrations/ -database "$(DB_URL_DOCKER)" -verbose up

migrate-down:
	migrate -path postgres/migrations/ -database "$(DB_URL_DOCKER)" -verbose down

migrate-up-local:
	migrate -path postgres/migrations/ -database "$(DB_URL_LOCAL)" -verbose up

migrate-down-local:
	migrate -path postgres/migrations/ -database "$(DB_URL_LOCAL)" -verbose down

migrate-up-docker:
	migrate -path postgres/migrations/ -database "$(DB_URL_DOCKER)" -verbose up

migrate-down-docker:
	migrate -path postgres/migrations/ -database "$(DB_URL_DOCKER)" -verbose down

sqlc:
	sqlc generate   

server:
	go run cmd/main.go

lint:
	golangci-lint run --timeout=5m

test:
	go test -v -race ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Run tests like CI does (with migrations)
ci-test:
	@echo "Setting up test database..."
	@docker compose up -d
	@sleep 2
	@echo "Running migrations..."
	@migrate -path postgres/migrations/ -database "$(DB_URL_DOCKER)" up
	@echo "Running tests with race detection and coverage..."
	@TEST_DB_URL="$(DB_URL_DOCKER)" go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out

# Build Docker image locally
docker-build:
	docker build -t double-entry-bank-go:local .

# Start full stack with Docker Compose
docker-up:
	docker compose up -d

# Stop Docker Compose services
docker-down:
	docker compose down

# Clean up build artifacts and test files
clean:
	rm -f coverage.out coverage.txt
	rm -f ledger ledger-*
	rm -f *.test
	docker compose down -v

# Display help
help:
	@echo "Available targets:"
	@echo "  postgres      - Start PostgreSQL container"
	@echo "  createdb      - Create database"
	@echo "  migrate-up    - Run database migrations"
	@echo "  migrate-down  - Rollback last migration"
	@echo "  migrate-up-local   - Run migrations against local PostgreSQL (5432)"
	@echo "  migrate-down-local - Rollback migration against local PostgreSQL (5432)"
	@echo "  migrate-up-docker  - Run migrations against Docker PostgreSQL (5433)"
	@echo "  migrate-down-docker - Rollback migration against Docker PostgreSQL (5433)"
	@echo "  sqlc          - Generate sqlc code"
	@echo "  server        - Run API server"
	@echo "  lint          - Run golangci-lint"
	@echo "  test          - Run tests with race detector"
	@echo "  coverage      - Generate coverage report"
	@echo "  ci-test       - Run tests like CI (with setup)"
	@echo "  docker-build  - Build Docker image locally"
	@echo "  docker-up     - Start full stack with Docker"
	@echo "  docker-down   - Stop Docker services"
	@echo "  clean         - Clean build artifacts"
	@echo "  help          - Show this help message"