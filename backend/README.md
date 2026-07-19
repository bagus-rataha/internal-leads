# Fiber API Boilerplate

A REST API starter built with Go Fiber, GORM, and PostgreSQL. It includes JWT
authentication with access and refresh tokens, UUID v7 primary keys, request
validation, and a layered architecture (handler → service → repository).

## Features

- Go Fiber v2
- PostgreSQL via GORM with connection pooling
- JWT auth: separate access/refresh secrets, refresh-token rotation and revocation
- UUID v7 primary keys (time-sortable)
- Request validation (go-playground/validator)
- SQL migrations with golang-migrate (no AutoMigrate)
- Security headers, CORS, and rate limiting (production)
- Unit and integration tests (testify, golang-migrate)
- Swagger docs (development)
- Hot reload with Air

## Project structure

```
cmd/api/main.go      Entry point
internal/
  app/               Fiber app setup
  container/         Dependency injection
  routes/            Route definitions
  config/            Config loading and DB connection
  middleware/        Auth, CORS, security headers, rate limit
  models/            Database models (BaseModel = UUID v7)
  repository/        Data access
  services/          Business logic
  handlers/          HTTP handlers
  utils/             JWT, password, response, validation helpers
migrations/          SQL migrations
```

## Getting started

Prerequisites: Go 1.25+, PostgreSQL, the `migrate` CLI (golang-migrate), and
optionally `air` for hot reload.

```bash
go mod download
cp .env.example .env          # then fill in the required values
createdb fiber_api
migrate -path migrations -database "$DATABASE_URL" up
air                           # or: go run ./cmd/api
```

`$DATABASE_URL` is a connection string such as
`postgres://user:password@localhost:5432/fiber_api?sslmode=disable`.

## Configuration

Configuration is read from environment variables, loaded from `.env` if present.
See `.env.example` for the full template.

Required — the app exits on startup if any is missing:

- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- `JWT_ACCESS_SECRET`, `JWT_REFRESH_SECRET`
- `APP_ENV` (`development` or `production`)
- `ALLOWED_ORIGINS` — required only when `APP_ENV=production`

Optional (have defaults):

- `PORT`, `JWT_ACCESS_EXPIRE`, `JWT_REFRESH_EXPIRE`, `RATE_LIMIT_MAX`, `RATE_LIMIT_WINDOW`

## Bootstrap superuser account

After running migrations, a bootstrap superuser account is automatically seeded:

- **Email:** `admin@leadsales.local`
- **Password:** `ChangeMe123!`

**⚠️ CRITICAL:** This password must be changed immediately after first login in any environment. Use the admin dashboard to update it before granting access to team members or deploying to production.

## Database and migrations

The schema is managed with golang-migrate; there is no AutoMigrate. Migration
files live in `migrations/`.

```bash
migrate -path migrations -database "$DATABASE_URL" up
migrate -path migrations -database "$DATABASE_URL" down 1
```

## API endpoints

`JWTProtected` only checks the access token, so endpoints are grouped by how they
authenticate.

No credential:

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`

Refresh token (HttpOnly cookie or request body) — these do not require an access
token, because they are used when the access token has expired:

- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`

Access token (`Authorization: Bearer <token>`):

- `POST /api/v1/auth/logout-all`
- `GET /api/v1/users/me`
- `PUT /api/v1/users/me`
- `GET /api/v1/users`

Other:

- `GET /health`
- `GET /swagger/*` (development only)

## API docs

Swagger annotations live in the handler comments. Regenerate `docs/` after
changing them:

```bash
swag init -g cmd/api/main.go -o docs
```

The UI is served at `/swagger/*` in development.

## Testing

Unit tests run without a database. Integration tests run against a real
PostgreSQL instance and are gated behind the `integration` build tag, so they
never run by accident.

```bash
# Unit tests (no database)
go test ./...
go test -cover ./...

# Integration tests (real database)
createdb fiber_api_test
export TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5432/fiber_api_test?sslmode=disable"
go test -tags=integration ./internal/repository/...
```

`TEST_DATABASE_URL` is read by the test helper for both GORM and golang-migrate;
it can also go in `.env`. URL-encode special characters in the password
(e.g. `@` → `%40`). On PowerShell, set it with `$env:TEST_DATABASE_URL="..."`.

Unit tests mock one layer's dependencies: services run against mock repositories
and handlers against mock services, so each layer is tested in isolation. The
integration schema is applied with golang-migrate and tables are truncated
before each test.

## Development vs production

| Setting       | Development     | Production        |
|---------------|-----------------|-------------------|
| Swagger UI    | on              | off               |
| CORS          | allow all       | `ALLOWED_ORIGINS` |
| Rate limiting | off             | on                |
| Compression   | off             | on                |
| Prefork       | off             | on                |
| SQL logging   | on              | off               |
| Stack traces  | on              | off               |

## Adding a module

1. Create the model, repository, service, and handler under `internal/`.
2. Wire them in `internal/container/container.go`.
3. Register routes in `internal/routes/api.go`.
4. Add a migration in `migrations/` for the new table.

## Production checklist

- Set `APP_ENV=production`
- Use strong, distinct values for `JWT_ACCESS_SECRET` and `JWT_REFRESH_SECRET`
- Set `ALLOWED_ORIGINS`
- Run migrations before starting the app
- Terminate TLS at a reverse proxy

## License

MIT
