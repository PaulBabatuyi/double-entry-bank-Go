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

The frontend is intentionally minimal — Next.js Typescript, — so the focus stays on the Go backend. See the [frontend repo](https://github.com/PaulBabatuyi/double-entry-bank) for details.

## Live Demo 

- Frontend: https://golangbank.app
- Frontend Repo: https://github.com/PaulBabatuyi/double-entry-bank
- API Docs: https://golangbank.app/swagger
- Health: https://golangbank.app/health

Dont forget to star and fork this project repo

## Article vs README

This README is intentionally concise and implementation-focused.

For the full technical narrative and tutorial, read the FreeCodeCamp article: [How to Build a Bank Ledger in Golang with PostgreSQL using Double-Entry Accounting](https://www.freecodecamp.org/news/build-a-bank-ledger-in-go-with-postgresql-using-the-double-entry-accounting-principle/)  
![freecodecamp](public/freecodecamp.png)

## Core Ledger Model

Each money movement writes balanced entries into the `entries` table:
- deposit: credit user account, debit settlement account
- withdrawal: debit user account, credit settlement account
- transfer: debit source account, credit destination account

Key constraints and behaviors implemented in code:
- single-sided entry rows (debit xor credit)
- account row locking (`FOR UPDATE`) during balance-changing operations
- serializable transactions with automatic retry on SQLSTATE `40001`
- reconciliation query computes `SUM(credit) - SUM(debit)` as source of truth
![Demo](public/frontend.png)

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
![Backend API endpoint](public/swagger.png)
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
# Set JWT_SECRET to at least 32 characters: openssl rand -base64 32

make postgres
make migrate-up
make sqlc
make server
```

Open:
- Swagger: http://localhost:8080/swagger/index.html
- Health: http://localhost:8080/health

## Testing

Recommended (requires Docker running):

```bash
make postgres
make test
```

With coverage report:

```bash
make coverage
```

Full CI-style run including migrations:

```bash
make ci-test
```

Environment used by tests:
- `TEST_DB_URL` (defaults to `postgresql://root:secret@localhost:5433/simple_ledger?sslmode=disable`)

## Make Targets

```bash
make postgres       # Start PostgreSQL container
make migrate-up     # Apply migrations
make migrate-down   # Rollback last migration
make sqlc           # Regenerate sqlc query code
make server         # Run the API server
make test           # Run tests with race detector
make coverage       # Generate coverage report
make lint           # Run golangci-lint
make ci-test        # Full test run including migrations
make docker-build   # Build Docker image locally
make docker-up      # Start full stack with Docker Compose
make docker-down    # Stop Docker Compose services
```

## Deployment

Render deployment instructions are in [DEPLOYMENT.md](DEPLOYMENT.md).

The container serves the backend API only. The frontend is deployed separately.

## Why This Project Exists

This repository is a practical fintech-backend demonstration covering:
- correctness under concurrency
- auditable money movement
- clear API boundaries
- production-minded deployment shape

If you are a recruiter or reviewer, start with this README and the Swagger UI. For the full technical narrative, read the FreeCodeCamp article linked above.
