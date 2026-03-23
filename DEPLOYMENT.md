# Deployment Guide (Render + PostgreSQL)

This project deploys cleanly to Render with one web service and one PostgreSQL database.
The Go app serves:
- API endpoints
- Swagger docs

Frontend is deployed separately:
- Frontend URL: `https://golangbank.app`
- Frontend Repo: `https://github.com/PaulBabatuyi/double-entry-bank`

## Architecture

- Web service: Docker runtime using `Dockerfile`
- Database: Render PostgreSQL (`ledger-db`)
- Migrations: run automatically on service startup via `docker-entrypoint`

## Prerequisites

- Code pushed to GitHub
- Render account
- `render.yaml` committed to your default branch

## Deploy

1. Push your latest code:

```bash
git add .
git commit -m "chore: prepare render deployment"
git push origin main
```

2. In Render dashboard:
- Click New + -> Blueprint
- Select your repository
- Render detects `render.yaml`
- Click Apply

3. Wait for first deploy.

Render creates:
- web service: `double-entry-ledger-api`
- postgres database: `ledger-db`

## Environment variables

Configured by blueprint:
- `DB_URL` from database connection string
- `JWT_SECRET` auto-generated
- `PORT=8080`
- `CORS_ALLOWED_ORIGINS=https://golangbank.app,http://localhost:3000,http://127.0.0.1:3000,http://localhost:5173,http://127.0.0.1:5173`

If you later host frontend from another domain, set `CORS_ALLOWED_ORIGINS` to include that origin.

## Verify

After deploy succeeds:

- Health: `https://golangbank.app/health`
- Swagger: `https://golangbank.app/swagger` (proxied through frontend)
- Frontend: `https://golangbank.app`

## Notes

- Render free tier sleeps after idle time; first request can be slow.
- Migrations run on each startup and safely no-op when there are no new migrations.

## Troubleshooting

### Service fails at startup

- Check Render logs
- Confirm `DB_URL` exists and database is running

### CORS errors

- Update `CORS_ALLOWED_ORIGINS` to include the frontend origin
- Redeploy service

### Database schema missing

- Review logs for migration errors
- Ensure migration files exist in `postgres/migrations`
