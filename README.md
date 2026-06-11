# aiynx-go — User API with Dynamic Age Calculation

## Overview

`aiynx-go` is a production-ready RESTful API built in Go that manages users
(name + date of birth). When fetching users, the API dynamically calculates
and returns the user's age using Go's standard `time` package — age is **never
stored** in the database. The project uses Fiber v2, PostgreSQL (Neon serverless),
SQLC for type-safe query codegen, pgx/v5 for connection pooling, Zap for
structured logging, and go-playground/validator for input validation.

---

## Prerequisites

| Tool | Version | Install |
|---|---|---|
| Go | ≥ 1.23 | https://go.dev/dl |
| SQLC | latest | `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest` |
| golangci-lint | latest | https://golangci-lint.run/usage/install |
| psql | any | bundled with PostgreSQL client tools |
| Docker & Compose | any | https://docs.docker.com/get-docker *(optional)* |

---

## Setup

```bash
# 1. Clone the repository
git clone https://github.com/Maniii97/aiynx-go.git
cd aiynx-go

# 2. Copy the environment template and fill in your Neon DATABASE_URL
cp .env.example .env
# Edit .env and set DATABASE_URL, APP_PORT, APP_ENV

# 3. Apply the database migration
make migrate

# 4. Generate type-safe Go code from SQL queries
make sqlc           # runs: sqlc generate

# 5. Download Go module dependencies
go mod tidy
```

---

## Running the Server

```bash
make run
# or equivalently:
go run ./cmd/server/main.go
```

The server starts on the port defined in `APP_PORT` (default `3000`).

---

## Running Tests

```bash
make test
# or equivalently:
go test -v -race ./...
```

---

## API Reference

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Health check — returns `{"status":"ok"}` |
| `POST` | `/users` | Create a user (`name`, `dob` required) |
| `GET` | `/users` | List users with pagination (`?page=1&limit=10`) |
| `GET` | `/users/:id` | Get a user by ID (includes calculated `age`) |
| `PUT` | `/users/:id` | Update a user's name and DOB |
| `DELETE` | `/users/:id` | Delete a user (returns 204 No Content) |

### Request / Response examples

**POST /users**
```json
// Request
{ "name": "Alice", "dob": "1990-05-10" }

// Response 201
{ "id": 1, "name": "Alice", "dob": "1990-05-10" }
```

**GET /users/1**
```json
// Response 200
{ "id": 1, "name": "Alice", "dob": "1990-05-10", "age": 35 }
```

**GET /users?page=1&limit=10**
```json
// Response 200
{
  "data": [{ "id": 1, "name": "Alice", "dob": "1990-05-10", "age": 35 }],
  "meta": { "page": 1, "limit": 10, "total": 1 }
}
```

### Error envelope

All errors follow a consistent JSON shape:
```json
{ "error": "human-readable message" }
```

| Scenario | Status |
|---|---|
| Body parse failure | 400 |
| Validation failure | 422 |
| Invalid `:id` param | 400 |
| User not found | 404 |
| Unexpected error | 500 |

---

## Docker

```bash
# Build and start the API container
docker compose up --build

# Run in background
docker compose up -d --build

# Stop
docker compose down
```

The container exposes the port defined in `APP_PORT` and includes a
`HEALTHCHECK` that polls `GET /health` every 30 seconds.

---

## Project Structure

```
.
├── cmd/server/main.go          # Entry point — wires everything
├── config/config.go            # Typed config from env vars
├── db/
│   ├── migrations/             # Plain SQL migration files
│   └── sqlc/                   # SQLC-generated code (do not edit)
├── internal/
│   ├── handler/                # HTTP handlers (parse → service → respond)
│   ├── middleware/             # RequestID + request logger
│   ├── models/                 # Request/response DTOs + custom validators
│   ├── repository/             # Thin SQLC wrapper
│   ├── routes/                 # Route registration
│   ├── service/                # Business logic (age calculation lives here)
│   └── logger/                 # Global Zap logger initialisation
├── sqlc.yaml                   # SQLC configuration
├── Makefile
├── Dockerfile
└── docker-compose.yml
```
