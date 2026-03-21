# Double-Entry Bank Ledger in Go

[![CI](https://github.com/PaulBabatuyi/double-entry-bank-Go/actions/workflows/ci.yml/badge.svg)](https://github.com/PaulBabatuyi/double-entry-bank-Go/actions/workflows/ci.yml)
[![Docker](https://github.com/PaulBabatuyi/double-entry-bank-Go/actions/workflows/docker.yml/badge.svg)](https://github.com/PaulBabatuyi/double-entry-bank-Go/actions/workflows/docker.yml)
[![CodeQL](https://github.com/PaulBabatuyi/double-entry-bank-Go/actions/workflows/codeql.yml/badge.svg)](https://github.com/PaulBabatuyi/double-entry-bank-Go/actions/workflows/codeql.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/PaulBabatuyi/double-entry-bank-Go)](https://goreportcard.com/report/github.com/PaulBabatuyi/double-entry-bank-Go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Production-focused Go backend that models bank-style money movement using strict double-entry accounting.

It demonstrates:
- Atomic transactions with PostgreSQL
- Concurrency safety with serializable isolation + retry
- Ledger-based reconciliation
- JWT auth + account-level authorization
- API docs, health checks, and Dockerized deployment

## Live Demo

- Frontend: https://golangbank.app 
- Frontend Repo: https://github.com/PaulBabatuyi/double-entry-bank
- API Docs: https://golangbank.app/swagger
- Health: https://double-entry-bank-go.onrender.com/health

## Article vs README

This README is intentionally concise and implementation-focused.

For the full narrative/tutorial version, check FreeCodeCamp.

## Core Ledger Model

Each money movement writes balanced entries into the `entries` table:
- deposit: credit user account, debit settlement account
- withdrawal: debit user account, credit settlement account
- transfer: debit source account, credit destination account

Key constraints/behaviors implemented in code:
- single-sided entry rows (debit xor credit)
- account row locking (`FOR UPDATE`) during balance-changing operations
- serializable transactions with automatic retry on SQLSTATE `40001`
- reconciliation query computes `SUM(credit) - SUM(debit)` as source of truth

## Tech Stack

- Go 1.24+
- Router: go-chi/chi
- Database: PostgreSQL 16
- Query layer: sqlc
- Auth: JWT (go-chi/jwtauth)
- Logging: zerolog
- API docs: swaggo + http-swagger
- Testing: Go test + testify + race detector
- Runtime: Docker + docker-compose

## API Endpoints

Public:
- `POST /register`
- `POST /login`
- `GET /health`
- `GET /swagger/index.html`

Protected (Bearer token required):
- `POST /accounts`
- `GET /accounts`
- `GET /accounts/{id}`
- `POST /accounts/{id}/deposit`
- `POST /accounts/{id}/withdraw`
- `POST /transfers`
- `GET /accounts/{id}/entries`
- `GET /accounts/{id}/reconcile`
- `GET /transactions/{id}`

## Project Structure

```text
.
├── cmd/
│   └── main.go
├── internal/
│   ├── api/
│   ├── db/
│   └── service/
├── postgres/
│   ├── migrations/
│   ├── queries/
│   └── sqlc/
├── docs/
├── docker-compose.yml
├── docker-entrypoint
├── Dockerfile
├── Makefile
└── README.md
```

## Local Development

### Prerequisites

- Go 1.24+
- Docker + docker compose
- migrate CLI
- sqlc CLI
- swag CLI (only needed when regenerating docs)

Install tools:

```bash
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/swaggo/swag/cmd/swag@latest
```

### Run Locally

```bash
git clone https://github.com/PaulBabatuyi/double-entry-bank-Go.git
cd double-entry-bank-Go
cp .env.example .env
# Set JWT_SECRET to at least 32 characters

make postgres
make migrate-up
make sqlc
make server
```

Open:
- Frontend: https://golangbank.app
- Swagger: http://localhost:8080/swagger/index.html
- Health: http://localhost:8080/health

## Testing

Recommended:

```bash
chmod +x run_tests.sh
./run_tests.sh
./run_tests.sh --coverage
```

Manual:

```bash
make postgres
make test
make coverage
make ci-test
```

Environment used by tests:
- `TEST_DB_URL` (defaults to `postgresql://root:secret@localhost:5433/simple_ledger?sslmode=disable`)

## Make Targets

```bash
make postgres
make migrate-up
make migrate-down
make sqlc
make server
make test
make coverage
make lint
make ci-test
make docker-build
make docker-up
make docker-down
```

## Deployment

Render deployment instructions are in [DEPLOYMENT.md](DEPLOYMENT.md).

Helper scripts:
- `scripts/deploy-render.sh` (Linux/macOS)
- `scripts/deploy-render.bat` (Windows)

The container serves the backend API only. Frontend is deployed separately.

## Why This Project Exists

This repository is designed as a practical fintech-backend demonstration:
- correctness under concurrency
- auditable money movement
- clear API boundaries
- production-minded deployment shape

If you are a recruiter or reviewer, start with this README and Swagger; if you want the full technical narrative, use the article draft.