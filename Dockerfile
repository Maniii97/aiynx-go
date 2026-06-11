# ─────────────────────────────────────────────────────────────────────────────
# Stage 1 — Builder
# ─────────────────────────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

# Install git (required by go modules that reference VCS)
RUN apk add --no-cache git

WORKDIR /app

# Cache dependency download as a separate layer
COPY go.mod go.sum ./
RUN go mod download

# Copy the full source and build a statically-linked binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/server ./cmd/server/main.go

# ─────────────────────────────────────────────────────────────────────────────
# Stage 2 — Runner (minimal image, no toolchain)
# ─────────────────────────────────────────────────────────────────────────────
FROM alpine:latest AS runner

# Install ca-certificates so TLS connections to Neon work
RUN apk add --no-cache ca-certificates curl

WORKDIR /app

COPY --from=builder /app/bin/server ./server

EXPOSE ${APP_PORT:-3000}

# Docker health check hits our /health endpoint
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD curl -f http://localhost:${APP_PORT:-3000}/health || exit 1

ENTRYPOINT ["./server"]
