# Contributing to Double-Entry Bank Ledger

Thank you for considering contributing to this project! This document outlines the development workflow, CI/CD pipeline, and contribution guidelines.

## Table of Contents

- [Development Setup](#development-setup)
- [Development Workflow](#development-workflow)
- [CI/CD Pipeline](#cicd-pipeline)
- [Code Standards](#code-standards)
- [Testing Requirements](#testing-requirements)
- [Pull Request Process](#pull-request-process)
- [Release Process](#release-process)

## Development Setup

### Prerequisites

- **Go 1.23+**
- **Docker & Docker Compose**
- **golang-migrate**: `go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest`
- **sqlc**: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`
- **golangci-lint**: Follow [installation guide](https://golangci-lint.run/usage/install/)

### Initial Setup

```bash
# Clone the repository
git clone https://github.com/PaulBabatuyi/double-entry-bank-Go.git
cd double-entry-bank-Go

# Copy environment template
cp .env.example .env

# Edit .env and set JWT_SECRET
# Generate with: openssl rand -base64 32

# Start PostgreSQL
make postgres

# Run migrations
make migrate-up

# Generate sqlc code
make sqlc

# Run tests
make test
```

## Development Workflow

### 1. Create a Feature Branch

```bash
git checkout -b feat/your-feature-name
# or
git checkout -b fix/bug-description
```

### 2. Make Your Changes

- Write code following [Code Standards](#code-standards)
- Add tests for new functionality
- Update documentation if needed
- Run linter locally: `make lint`
- Run tests locally: `make test`

### 3. Commit Your Changes

Follow [Conventional Commits](https://www.conventionalcommits.org/) format:

```bash
git commit -m "feat: add transfer reconciliation endpoint"
git commit -m "fix: correct balance calculation in concurrent deposits"
git commit -m "docs: update API documentation for transfers"
```

**Commit types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `build`: Build system changes
- `ci`: CI/CD changes
- `chore`: Other changes (dependencies, etc.)

### 4. Push and Create PR

```bash
git push origin feat/your-feature-name
```

Then create a Pull Request on GitHub with:
- Clear title following conventional commits
- Description of changes
- Related issue number (if applicable)
- Screenshots/examples if relevant

## CI/CD Pipeline

### Automated Checks on Every PR

When you create a PR, the following checks run automatically:

#### 1. **Lint Check** (~2 min)
- Runs `golangci-lint` with 30+ linters
- Enforces code quality and style
- Must pass before merge

#### 2. **Test Suite** (~3-5 min)
- Spins up PostgreSQL service
- Runs migrations
- Executes all tests with race detection
- Generates coverage report
- Coverage report commented on PR

#### 3. **Build Check** (~2 min)
- Compiles the application
- Ensures no build errors
- Creates binary artifact

#### 4. **Security Scans** (~3 min)
- **Gosec**: Static security analysis
- **CodeQL**: Advanced security scanning
- Results uploaded to GitHub Security tab

#### 5. **PR Metadata Checks** (~1 min)
- Validates PR title format
- Checks PR size (warns if too large)
- Auto-labels based on changed files
- Detects breaking changes

### Pipeline Stages

```
┌─────────────────────────────────────────────────────────────┐
│                    PR Created/Updated                        │
└─────────────────┬───────────────────────────────────────────┘
                  │
        ┌─────────┴─────────┐
        │                   │
        ▼                   ▼
   ┌────────┐         ┌──────────┐
   │  Lint  │         │   Test   │
   └────┬───┘         └─────┬────┘
        │                   │
        └─────────┬─────────┘
                  │
                  ▼
            ┌──────────┐
            │  Build   │
            └─────┬────┘
                  │
                  ▼
           ┌───────────┐
           │ Security  │
           └─────┬─────┘
                 │
                 ▼
          ┌────────────┐
          │   Ready    │
          │ to Merge   │
          └────────────┘
                 │
                 ▼ (merge to main)
          ┌────────────┐
          │   Docker   │
          │   Build    │
          └────────────┘
```

## Code Standards

### Go Code Style

1. **Follow standard Go conventions**
   - Use `gofmt` (automated)
   - Follow [Effective Go](https://golang.org/doc/effective_go)

2. **Error Handling**
   ```go
   // Good ✅
   if err != nil {
       log.Error().Err(err).Msg("failed to process transaction")
       return fmt.Errorf("process transaction: %w", err)
   }
   
   // Bad ❌
   if err != nil {
       return err  // No context
   }
   ```

3. **Logging**
   ```go
   // Use structured logging with context
   log.Info().
       Str("user_id", userID.String()).
       Str("account_id", accountID.String()).
       Str("amount", amount).
       Msg("deposit successful")
   ```

4. **Comments**
   - Add doc comments for exported functions
   - Explain "why" not "what"
   - Keep comments up to date

### Database Changes

1. **Migrations**
   - Always create both `up` and `down` migrations
   - Make migrations idempotent (use `IF EXISTS`, `IF NOT EXISTS`)
   - Test migrations locally before pushing

2. **SQL Queries**
   - Add new queries to appropriate file in `postgres/queries/`
   - Run `make sqlc` to generate Go code
   - Never write raw SQL in Go code

### Testing Standards

Every PR must include tests for new functionality:

```go
func TestDeposit_Success(t *testing.T) {
    // Arrange
    ctx := context.Background()
    ledger := setupTestLedger(t)
    accountID := createTestAccount(t, ledger)
    
    // Act
    err := ledger.Deposit(ctx, accountID, "100.00")
    
    // Assert
    require.NoError(t, err)
    balance := getAccountBalance(t, ledger, accountID)
    assert.Equal(t, "100.0000", balance)
}
```

**Test Coverage Requirements:**
- New code should have >80% coverage
- Critical paths (financial operations) need 100% coverage
- Include both success and failure cases

## Testing Requirements

### Before Submitting PR

```bash
# 1. Run linter
make lint

# 2. Run all tests
make test

# 3. Check coverage
make coverage
# Target: >80% overall coverage

# 4. Test race conditions
go test -race ./...

# 5. Manual testing (if applicable)
# Start server and test endpoints manually
```

### Writing Tests

1. **Unit Tests**: Test individual functions in isolation
2. **Integration Tests**: Test database interactions
3. **Handler Tests**: Test HTTP endpoints with authentication
4. **Concurrency Tests**: Use `-race` flag to detect race conditions

## Pull Request Process

### PR Checklist

Before creating a PR, ensure:

- [ ] Code follows style guidelines (passes `make lint`)
- [ ] All tests pass (`make test`)
- [ ] New tests added for new functionality
- [ ] Documentation updated if needed
- [ ] Commit messages follow conventional commits
- [ ] PR title follows conventional commits format
- [ ] PR description clearly explains changes
- [ ] No breaking changes (or documented if necessary)

### PR Review Process

1. **Automated Checks**: All CI checks must pass
2. **Code Review**: At least one approval required
3. **Coverage**: Coverage should not decrease
4. **Security**: No security vulnerabilities introduced

### PR Size Guidelines

- **Small** (recommended): <300 lines changed, <10 files
- **Medium**: 300-600 lines, 10-20 files  
- **Large** (split if possible): >600 lines, >20 files

Large PRs are harder to review and may be rejected with a request to split them.

## Release Process

### Creating a Release

Releases are automated via GitHub Actions:

```bash
# 1. Update version and create tag
git tag -a v1.2.3 -m "Release v1.2.3: Add transfer reconciliation"

# 2. Push tag to trigger release workflow
git push origin v1.2.3
```

This automatically:
- ✅ Builds binaries for all platforms (Linux, macOS, Windows)
- ✅ Generates changelog from commits
- ✅ Creates GitHub release
- ✅ Builds and pushes Docker image with version tag
- ✅ Updates `latest` tag for Docker image

### Version Numbering

Follow [Semantic Versioning](https://semver.org/):

- **MAJOR** (v2.0.0): Breaking changes
- **MINOR** (v1.1.0): New features (backward compatible)
- **PATCH** (v1.0.1): Bug fixes

Example tags:
```bash
v1.0.0      # Initial release
v1.1.0      # Added new feature
v1.1.1      # Bug fix
v2.0.0      # Breaking change
v2.0.0-rc.1 # Release candidate (prerelease)
```

## Getting Help

- **Issues**: Open an issue for bugs or feature requests
- **Discussions**: Use GitHub Discussions for questions
- **Security**: Report security issues privately to maintainers

## Code of Conduct

- Be respectful and inclusive
- Provide constructive feedback
- Focus on the code, not the person
- Follow the [Contributor Covenant](https://www.contributor-covenant.org/)

---

Thank you for contributing! 🚀
