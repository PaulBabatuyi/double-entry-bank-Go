@echo off
echo Running Double-Entry Bank Ledger Tests
echo =========================================

REM Check if database is running
docker compose ps db | findstr "running" >nul
if %errorlevel% neq 0 (
    echo Database is not running. Starting database...
    docker compose up -d db
    echo Waiting for database to be ready...
    timeout /t 10 /nobreak >nul
)

echo Database is ready

REM Set test database URL
set TEST_DB_URL=postgresql://root:secret@localhost:5433/simple_ledger?sslmode=disable

REM Run migrations
echo Running migrations...
migrate -path postgres/migrations/ -database "%TEST_DB_URL%" -verbose up 2>nul

echo.
echo Running tests...
echo.

REM Run tests with race detection
go test -v -race -timeout 30s ./internal/service ./internal/api ./internal/db

echo.
echo All tests completed!

if "%1"=="--coverage" (
    echo.
    echo Running tests with coverage...
    go test -coverprofile=coverage.out ./internal/service ./internal/api ./internal/db
    go tool cover -func=coverage.out
    echo.
    echo To view HTML coverage report, run: go tool cover -html=coverage.out
)
