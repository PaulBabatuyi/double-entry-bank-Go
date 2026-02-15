# Double-Entry Bank Ledger in Go

Production-grade demonstration of a **double-entry accounting ledger** backend in Go, simulating core banking operations with strong emphasis on financial correctness, traceability, and atomicity.

This project was built to showcase backend engineering skills relevant to fintech environments — particularly **ledger systems**, **payment processing**, **settlement**, **reliability under concurrency**, and **observability** — aligning closely with roles building scalable financial infrastructure (e.g., fiat deposits/withdrawals, internal ledgers, reconciliation, auditability).

### Features

- Strict **double-entry bookkeeping** (every transaction creates exactly two opposing entries: debit & credit)
- Atomic operations via PostgreSQL transactions (deposits, withdrawals, transfers)
- Immutable audit trail via `entries` table (source of truth)
- Settlement account for external cash flows (deposit/withdrawal simulation)
- Secure JWT-based authentication (register, login)
- Structured logging with **Zerolog**
- Health check endpoint (`/health`)
- Interactive API documentation with **Swagger/OpenAPI** (`/swagger/index.html`)
- Basic observability (request logging, health, future Prometheus metrics)
- Fully containerized with Docker + golang-migrate
- Comprehensive tests (unit, integration, race detection)

### Relevance to Fintech Roles

This project directly demonstrates patterns used in real-world financial systems (e.g., Monzo, Nubank, Revolut-like services):

- **Ledger & balance systems**: consistency, correctness, traceability of funds
- **Payment lifecycles**: authorization, settlement, reconciliation, failure handling
- **Scalable Go services**: high-throughput, low-latency, concurrent-safe operations
- **Operational excellence**: monitoring (health/logs), incident readiness, end-to-end ownership
- **Security & reliability**: JWT auth, input validation, atomic transactions, no races

Ideal for roles requiring strong production experience with Go in distributed/fintech systems.

### Tech Stack

- **Language**: Go 1.23+
- **Web framework**: go-chi/chi (lightweight & performant router)
- **Database**: PostgreSQL 16 + sqlc (type-safe queries)
- **Migrations**: golang-migrate
- **Auth**: JWT (HS256) with go-chi/jwtauth
- **Logging**: Zerolog (structured, zero-allocation)
- **Documentation**: swaggo/swag + http-swagger (OpenAPI/Swagger UI)
- **Testing**: Go built-in + testify + race detector
- **Containerization**: Docker + docker-compose

### Folder Structure
.
├── cmd/
│   └── main.go                 # Server entrypoint
├── internal/
│   ├── api/
│   │   ├── handler.go          # HTTP handlers + Swagger annotations
│   │   └── middleware.go       # JWT helpers & token generation
│   ├── db/
│   │   └── store.go            # sqlc wrapper + transaction support
│   └── service/
│       └── ledger.go           # Core business logic (double-entry ops)
├── postgres/
│   ├── migrations/             # golang-migrate SQL files
│   └── queries/                # sqlc query files (.sql)
│       ├── accounts.sql
│       ├── entries.sql
│       └── users.sql
├── docs/                       # Generated Swagger docs (do not edit)
├── .env                        # Secrets (git ignored)
├── .env.example                # Template for env vars
├── docker-compose.yml
├── Dockerfile                  # For app container
├── go.mod / go.sum
├── Makefile                    # Common commands
└── README.md
text### Quick Start (Local Development)

1. **Prerequisites**
   - Go 1.23+
   - Docker & docker-compose
   - golang-migrate CLI (`go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest`)
   - sqlc CLI (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`)
   - swag CLI (`go install github.com/swaggo/swag/cmd/swag@latest`)

2. **Clone & Setup**
   ```bash
   git clone https://github.com/PaulBabatuyi/double-entry-bank-Go.git
   cd double-entry-bank-Go
   cp .env.example .env
   # Edit .env → set JWT_SECRET (generate strong one: openssl rand -base64 32)

Start PostgreSQL
make postgres

Run Migrations
make migrate-up

Generate sqlc Code
make sqlc

Generate Swagger Docs
swag init -g cmd/main.go --parseDependency --parseInternal
Run Server
make server
# or: go run cmd/main.go

Access
API base: http://localhost:8080
Health check: http://localhost:8080/health
Swagger UI: http://localhost:8080/swagger/index.html
Register: POST /register → get JWT
Try endpoints (use Bearer token in Swagger "Authorize")

make postgres       # Start Postgres container
make migrate-up     # Apply migrations
make migrate-down   # Rollback last migration
make sqlc           # Generate sqlc code
make test           # Run tests with race detector
make lint           # Run golangci-lint
make coverage       # Generate & open coverage report
make server         # Run the API server