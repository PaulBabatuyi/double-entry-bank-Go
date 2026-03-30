# Learning Guide — Double-Entry Bank Ledger in Go

> **Full tutorial:** [How to Build a Bank Ledger in Golang with PostgreSQL using Double-Entry Accounting](https://www.freecodecamp.org/news/build-a-bank-ledger-in-go-with-postgresql-using-the-double-entry-accounting-principle/) on FreeCodeCamp

This guide maps the codebase to the article. Work through the FreeCodeCamp tutorial alongside the source code for the full picture.

## Who This Is For

- Go developers who want to see a production-shaped backend beyond CRUD
- Engineers curious about financial systems and double-entry accounting
- People preparing for backend or fintech interviews who want a real project to reference

No prior fintech experience needed. Solid Go fundamentals and some SQL knowledge are enough to get started.

## Further Reading

- [FreeCodeCamp Article](https://www.freecodecamp.org/news/build-a-bank-ledger-in-go-with-postgresql-using-the-double-entry-accounting-principle/) — the full written tutorial for this project
- [PostgreSQL Transaction Isolation](https://www.postgresql.org/docs/current/transaction-iso.html) — official docs on isolation levels and serialization failures
- [shopspring/decimal](https://github.com/shopspring/decimal) — the decimal library used for all money math
- [sqlc documentation](https://docs.sqlc.dev/) — how SQL queries are compiled to type-safe Go
- [golang-migrate](https://github.com/golang-migrate/migrate) — database migration tooling
- [go-chi/jwtauth](https://github.com/go-chi/jwtauth) — JWT middleware used in this project
- [Effective Go](https://golang.org/doc/effective_go) — Go idioms and conventions referenced in the code style guide

## Live Demo

- Frontend: https://golangbank.app
- API Docs: https://golangbank.app/swagger
- Health: https://golangbank.app/health

The live demo runs on Render's free tier, so the first request after idle may be slow while the container wakes up.
