# Double-Entry Bank Ledger in Go

[![CI](https://github.com/PaulBabatuyi/double-entry-bank-Go/actions/workflows/ci.yml/badge.svg)](https://github.com/PaulBabatuyi/double-entry-bank-Go/actions/workflows/ci.yml)
[![Docker](https://github.com/PaulBabatuyi/double-entry-bank-Go/actions/workflows/docker.yml/badge.svg)](https://github.com/PaulBabatuyi/double-entry-bank-Go/actions/workflows/docker.yml)
[![CodeQL](https://github.com/PaulBabatuyi/double-entry-bank-Go/actions/workflows/codeql.yml/badge.svg)](https://github.com/PaulBabatuyi/double-entry-bank-Go/actions/workflows/codeql.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/PaulBabatuyi/double-entry-bank-Go)](https://goreportcard.com/report/github.com/PaulBabatuyi/double-entry-bank-Go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Production-grade demonstration of a **double-entry accounting ledger** backend in Go, simulating core banking operations with strong emphasis on financial correctness, traceability, and atomicity.

This project was built to showcase backend engineering skills relevant to fintech environments — particularly **ledger systems**, **payment processing**, **settlement**, **reliability under concurrency**, and **observability** — aligning closely with roles building scalable financial infrastructure (e.g., fiat deposits/withdrawals, internal ledgers, reconciliation, auditability).

##  Live Demo

- **Frontend (Render)**: https://double-entry-bank-go.onrender.com
- **API Docs (Render)**: https://your-service-name.onrender.com/swagger/index.html  
- **Health Check**: https://double-entry-bank-go.onrender.com/health

**Want to deploy your own?** See [DEPLOYMENT.md](DEPLOYMENT.md) for step-by-step instructions.

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

6. **Access**
   - **Demo Frontend**: http://localhost:8080  (Interactive web UI)
   - **API base**: http://localhost:8080
   - **Health check**: http://localhost:8080/health
   - **Swagger UI**: http://localhost:8080/swagger/index.html
   - Register: POST /register → get JWT
   - Try endpoints (use Bearer token in Swagger "Authorize")

### Demo Frontend

The project includes a modern web interface for easy demonstration:
-  User registration & login
-  Account management (create multiple accounts)
-  Deposit & Withdraw operations
-  Transfer between accounts
-  Real-time transaction history
-  Responsive design (mobile & desktop)

**Perfect for showcasing!**

Access at: http://localhost:8080 after starting the server.
- **Service layer**: Ledger operations (deposit, withdraw, transfer, reconcile)
- **API layer**: HTTP handlers with authentication and authorization
- **Database layer**: Store transactions and concurrency handling

**Option 1: Using test script** (recommended)
```bash
# Linux/Mac
chmod +x run_tests.sh
./run_tests.sh

# Windows
run_tests.bat

# With coverage report
./run_tests.sh --coverage
```

**Option 2: Manual testing**
```bash
# Ensure DB is running
make postgres

# Run tests with race detection
make test

# Run tests with coverage
make coverage

# Run specific package tests
go test -v -race ./internal/service
go test -v -race ./internal/api
go test -v -race ./internal/db
```

**Test Environment Variables:**
- `TEST_DB_URL`: PostgreSQL connection string (default: `postgresql://root:secret@localhost:5433/simple_ledger?sslmode=disable`)

**Test Coverage:**
- Service tests: Deposit, withdraw, transfer, reconcile, concurrent operations, edge cases
- Handler tests: All endpoints with auth/authz, success and failure scenarios
- Store tests: Transaction atomicity, rollback, isolation levels

### CI/CD

The project includes a comprehensive CI/CD pipeline using **GitHub Actions**.

#### Workflows

**1. CI Pipeline** (`.github/workflows/ci.yml`)
- **Triggers**: Push to `main`/`develop`, Pull Requests
- **Jobs**:
  - **Lint**: Runs `golangci-lint` with 30+ linters
  - **Test**: Runs all tests with PostgreSQL service, race detection, and coverage reporting
  - **Build**: Compiles binary and uploads as artifact
  - **Security**: Runs Gosec security scanner and uploads results to GitHub Security

**2. Docker Build** (`.github/workflows/docker.yml`)
- **Triggers**: Push to `main`, version tags (`v*.*.*`), releases
- **Actions**:
  - Builds multi-platform Docker images (amd64, arm64)
  - Pushes to GitHub Container Registry (`ghcr.io`)
  - Tags images appropriately (latest, version, sha)
  - Generates build attestations for supply chain security

**3. CodeQL Analysis** (`.github/workflows/codeql.yml`)
- **Triggers**: Push, PR, scheduled weekly
- **Actions**: Advanced security scanning using GitHub CodeQL

**4. Release** (`.github/workflows/release.yml`)
- **Triggers**: Version tags (`v*.*.*`)
- **Actions**:
  - Builds binaries for multiple platforms (Linux, macOS, Windows)
  - Generates changelog from git commits
  - Creates GitHub release with binaries attached

**5. Dependabot** (`.github/dependabot.yml`)
- Automatically updates Go modules, Docker images, and GitHub Actions
- Opens PRs weekly for dependency updates

#### CI/CD Status Badges

Add these to the top of your README for visibility:

```markdown
[![CI](https://github.com/PaulBabatuyi/double-entry-bank-Go/actions/workflows/ci.yml/badge.svg)](https://github.com/PaulBabatuyi/double-entry-bank-Go/actions/workflows/ci.yml)
[![Docker](https://github.com/PaulBabatuyi/double-entry-bank-Go/actions/workflows/docker.yml/badge.svg)](https://github.com/PaulBabatuyi/double-entry-bank-Go/actions/workflows/docker.yml)
[![CodeQL](https://github.com/PaulBabatuyi/double-entry-bank-Go/actions/workflows/codeql.yml/badge.svg)](https://github.com/PaulBabatuyi/double-entry-bank-Go/actions/workflows/codeql.yml)
```

#### Running Docker Image from GHCR

```bash
# Pull the latest image
docker pull ghcr.io/paulbabatuyi/double-entry-bank-go:latest

# Run with environment variables
docker run -p 8080:8080 \
  -e DB_URL="postgresql://user:pass@host:5432/db?sslmode=disable" \
  -e JWT_SECRET="your-secret-key-min-32-chars" \
  ghcr.io/paulbabatuyi/double-entry-bank-go:latest
```

#### Creating a Release

```bash
# Create and push a version tag
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0

# This automatically:
# 1. Triggers release workflow
# 2. Builds binaries for all platforms
# 3. Creates GitHub release with changelog
# 4. Builds and pushes Docker image with version tag
```

#### Local Linting (matches CI)

```bash
# Run the same linters as CI
make lint

# Or manually:
golangci-lint run --timeout=5m
```

### Quick Reference (Makefile)

```bash
make postgres       # Start Postgres container
make migrate-up     # Apply migrations
make migrate-down   # Rollback last migration
make sqlc           # Generate sqlc code
make test           # Run tests with race detector
make lint           # Run golangci-lint
make coverage       # Generate & open coverage report
make server         # Run the API server
```

---

## 🚀 Deployment

### Production Deployment (Render)

Deploy the backend API and frontend together to **Render** (with PostgreSQL) for a simple production setup with automatic HTTPS and CI/CD.

📖 **Complete deployment guide**: [DEPLOYMENT.md](DEPLOYMENT.md)

**Quick deployment:**

```bash
# 1. Deploy to Render
./scripts/deploy-render.sh    # Linux/Mac
# or
scripts\deploy-render.bat      # Windows
```

**What you get:**
- ✅ Backend API on Render with PostgreSQL database
- ✅ Frontend served from the same Render service
- ✅ Automatic HTTPS
- ✅ Auto-deploy on git push
- ✅ Free tier hosting (perfect for portfolio)

After deployment, update the "Live Demo" section at the top of this README with your actual URLs!
