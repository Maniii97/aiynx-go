# aiynx-go — User API with Auth, JWT & RBAC

## Overview

`aiynx-go` is a production-ready RESTful API built in Go that manages users
(name, date of birth, email, role). It features:

- **JWT authentication** (HS256, Bearer header + HttpOnly cookie)
- **Role-based access control** (`user` / `admin`)
- Dynamic age calculation — age is **never stored** in the database
- Fiber v2, PostgreSQL (Neon serverless), SQLC, pgx/v5, Zap, go-playground/validator

---

## Prerequisites

| Tool | Version | Install |
|---|---|---|
| Go | ≥ 1.23 | https://go.dev/dl |
| SQLC | latest | `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest` |
| golangci-lint | latest | https://golangci-lint.run/usage/install |
| Docker & Compose | any | https://docs.docker.com/get-docker *(optional)* |

---

## Setup

```bash
# 1. Clone the repository
git clone https://github.com/Maniii97/aiynx-go.git
cd aiynx-go

# 2. Copy the environment template and fill in real values
cp .env.example .env
# Edit .env — set DATABASE_URL and JWT_SECRET at minimum

# 3. Apply database migrations (runs all files in db/migrations/ in order)
make migrate

# 4. Generate type-safe Go code from SQL queries
make sqlc

# 5. Download Go module dependencies
go mod tidy
```

---

## Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | ✅ | — | PostgreSQL connection string |
| `JWT_SECRET` | ✅ | — | HMAC signing secret (min 32 chars) |
| `APP_PORT` | | `3000` | HTTP listen port |
| `APP_ENV` | | `development` | `development` or `production` |
| `JWT_EXPIRY_HOURS` | | `24` | Token TTL in hours |

> **Generate a strong JWT secret:** `openssl rand -hex 32`

---

## Running the Server

```bash
make run
# or:
go run ./cmd/server/main.go
```

---

## Running Tests

```bash
make test
# or:
go test -v -race ./...
```

---

## API Reference

### Public Routes (no auth)

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Health check |
| `POST` | `/auth/signup` | Register a new user |
| `POST` | `/auth/login` | Login — returns JWT + sets cookie |

### Authenticated Routes (JWT required)

| Method | Path | Description |
|---|---|---|
| `GET` | `/users/me` | Current user's profile (from token) |
| `GET` | `/users` | List users with pagination |
| `GET` | `/users/:id` | Get a user by ID |
| `POST` | `/users` | Create a user (requires email + password) |
| `PUT` | `/users/:id` | Update a user's name and DOB |

### Admin-Only Routes (JWT + `admin` role)

| Method | Path | Description |
|---|---|---|
| `DELETE` | `/admin/users/:id` | Delete a user |
| `GET` | `/admin/users` | List all users (admin view) |

---

### Authentication

**Signup — POST /auth/signup**
```json
// Request
{ "name": "Alice", "dob": "1990-05-10", "email": "alice@example.com", "password": "StrongPass1!" }

// Response 201
{ "id": 1, "name": "Alice", "email": "alice@example.com", "role": "user" }
```

Password rules: min 8 chars, at least 1 uppercase, 1 digit, 1 special character.

**Login — POST /auth/login**
```json
// Request
{ "email": "alice@example.com", "password": "StrongPass1!" }

// Response 200 — also sets auth_token HttpOnly cookie
{ "token": "<jwt>", "role": "user" }
```

**Using the token** — include it in the `Authorization` header:
```
Authorization: Bearer <jwt>
```
Or the cookie `auth_token` is sent automatically by the browser.

**GET /users/me**
```json
// Response 200
{ "id": 1, "name": "Alice", "email": "alice@example.com", "dob": "1990-05-10", "age": 35, "role": "user" }
```

---

### Standard Error Envelope

All errors return a consistent JSON shape:
```json
{
  "error": {
    "message": "human-readable description",
    "code": "MACHINE_READABLE_CODE",
    "request_id": "uuid-v4"
  }
}
```

| Scenario | Status | Code |
|---|---|---|
| Parse / bad JSON | 400 | `BAD_REQUEST` |
| Validation failure | 422 | `VALIDATION_ERROR` |
| Invalid `:id` param | 400 | `INVALID_PARAM` |
| Email already exists | 409 | `CONFLICT` |
| Wrong credentials | 401 | `INVALID_CREDENTIALS` |
| Missing / invalid token | 401 | `UNAUTHORIZED` |
| Insufficient role | 403 | `FORBIDDEN` |
| Resource not found | 404 | `NOT_FOUND` |
| Unexpected / DB error | 500 | `INTERNAL_ERROR` |

---

## Docker

```bash
docker compose up --build      # build and start
docker compose up -d --build   # background
docker compose down            # stop
```

---

## Project Structure

```
.
├── cmd/
│   ├── server/main.go          # Entry point — wires everything
│   └── migrate/main.go         # Runs all migrations in order
├── config/config.go            # Typed config from env vars
├── db/
│   ├── migrations/             # Plain SQL migration files (000001, 000002…)
│   └── sqlc/                   # SQLC-generated code (do not edit)
├── internal/
│   ├── handler/
│   │   ├── user_handler.go     # User CRUD + /users/me
│   │   └── auth_handler.go     # Signup + Login
│   ├── middleware/
│   │   ├── request_id.go       # Reuses or generates X-Request-ID
│   │   ├── logger.go           # Zap request logger
│   │   ├── auth.go             # JWT validation middleware + GetAuthUser helper
│   │   └── rbac.go             # RequireRole middleware
│   ├── models/
│   │   ├── user.go             # User DTOs + validators
│   │   ├── auth.go             # Auth DTOs (signup, login, me)
│   │   └── error.go            # Standard error envelope
│   ├── repository/             # Thin SQLC wrapper
│   ├── routes/routes.go        # Route groups (public / auth / admin)
│   ├── service/
│   │   ├── user_service.go     # User CRUD business logic
│   │   └── auth_service.go     # bcrypt, JWT, signup, login
│   └── logger/                 # Global Zap logger initialisation
├── sqlc.yaml
├── Makefile
├── Dockerfile
└── docker-compose.yml
```
