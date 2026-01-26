#!/bin/bash
set -e

echo "Running Double-Entry Bank Ledger Tests"
echo "========================================="

# Check if database is running
if ! docker compose ps db | grep -q "running"; then
    echo " Database is not running. Starting database..."
    docker compose up -d db
    echo "⏳ Waiting for database to be ready..."
    sleep 5
fi

# Check database health
until docker compose exec -T db pg_isready -U root -d simple_ledger > /dev/null 2>&1; do
    echo "⏳ Waiting for database to be ready..."
    sleep 2
done

echo "✅ Database is ready"

# Set test database URL
export TEST_DB_URL="postgresql://root:secret@localhost:5433/simple_ledger?sslmode=disable"

# Run migrations if needed
echo "📦 Running migrations..."
migrate -path postgres/migrations/ -database "$TEST_DB_URL" -verbose up 2>/dev/null || echo "Migrations already applied"

echo ""
echo "🧪 Running tests..."
echo ""

# Run tests with race detection. Limit package parallelism to reduce DB contention in integration tests.
go test -p 1 -v -race -timeout 30s ./internal/service ./internal/api ./internal/db

echo ""
echo "✅ All tests completed!"

# Optionally run tests with coverage
if [ "$1" == "--coverage" ]; then
    echo ""
    echo "📊 Running tests with coverage..."
    go test -p 1 -coverprofile=coverage.out ./internal/service ./internal/api ./internal/db
    go tool cover -func=coverage.out
    echo ""
    echo "💡 To view HTML coverage report, run: go tool cover -html=coverage.out"
fi
