.PHONY: run build sqlc migrate lint test promote-admin create-admin

# ── Run ──────────────────────────────────────────────────────────────────────
run:
	go run ./cmd/server/main.go

# ── Build ─────────────────────────────────────────────────────────────────────
build:
	go build -o bin/server ./cmd/server/main.go

# ── SQLC code generation ──────────────────────────────────────────────────────
sqlc:
	@if command -v sqlc >/dev/null 2>&1; then \
		sqlc generate; \
	else \
		$$(go env GOPATH)/bin/sqlc generate; \
	fi

# ── Run database migration ────────────────────────────────────────────────────
migrate:
	go run ./cmd/migrate/main.go

# ── Lint ──────────────────────────────────────────────────────────────────────
lint:
	golangci-lint run ./...

# ── Test ──────────────────────────────────────────────────────────────────────
test:
	go test -v -race ./...
