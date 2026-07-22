# Frontend build stage
FROM node:20-bookworm AS frontend-builder

WORKDIR /build

# Vite bakes VITE_* env vars into the bundle at build time.
# Pass via `fly deploy --build-arg VITE_SENTRY_DSN_FE=<dsn>` or set in fly.toml [build.args].
ARG VITE_SENTRY_DSN_FE=""
ENV VITE_SENTRY_DSN_FE=$VITE_SENTRY_DSN_FE

COPY frontend/package*.json ./
RUN npm ci

COPY frontend/ .
RUN npm run build

# Backend build stage
FROM golang:1.26-bookworm AS builder

WORKDIR /build

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .

RUN CGO_ENABLED=1 go build -o server ./cmd/server

# Runtime stage
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /build/server .
COPY --from=builder /build/migrations/ ./migrations/
COPY --from=frontend-builder /build/dist ./static/

EXPOSE 8080

CMD ["./server"]
