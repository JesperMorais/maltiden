# Maltiden Backend

## What This Is

A Go REST API backend for a family meal planning app (Maltiden) — reviewed, hardened, and optimized. JWT auth, SQLite with WAL mode, repository interfaces for testability, in-memory Tjek API caching, structured logging, and production-ready deployment configuration.

## Core Value

A secure, performant, and well-architected backend that handles meal planning, recipe management, shopping lists, and grocery offer integration reliably.

## Requirements

### Validated

- ✓ Codebase mapped — `.planning/codebase/` with 7 analysis documents — existing
- ✓ Deep review of all backend handlers, services, and storage — v1.0
- ✓ Security audit: JWT, input validation, auth middleware, request size limits — v1.0
- ✓ Architecture review: layering, dependency patterns, error propagation — v1.0
- ✓ Performance review: N+1, caching, sorting, concurrency — v1.0
- ✓ Findings report with severity ratings (93 findings) — v1.0
- ✓ Prioritized fix plan (9 batches, dependency-ordered) — v1.0
- ✓ JWT auth hardening: secret validation, HS256 pinning, shopping endpoint auth — v1.1
- ✓ IDOR protection: menu/shopping ownership verification — v1.1
- ✓ Database integrity: foreign keys, SQL injection fixes, UUID-based IDs — v1.1
- ✓ Error infrastructure: 18 sentinel errors, WriteError/WriteJSON helpers — v1.1
- ✓ Input validation: body limits, coordinate validation, business rule bounds — v1.1
- ✓ Performance: Tjek API caching, WAL mode, N+1 elimination, connection pooling — v1.1
- ✓ Architecture: repository interfaces, injectable JWTService, extracted DI — v1.1
- ✓ Production deployment: embedded migrations, graceful shutdown, health checks — v1.1
- ✓ Observability: structured logging, request ID tracing, configurable CORS — v1.1

### Active

(None — v1.1 Fix milestone complete. Next milestone to be defined.)

### Out of Scope

- Frontend code — backend-only review and fixes
- Test writing — identified as gap, deferred to future milestone
- Infrastructure migration (e.g., PostgreSQL) — SQLite sufficient for current scale
- Rate limiting — deferred, can be added as middleware later
- Mobile app — web-first approach

## Context

- Backend is a Go 1.24 REST API serving a family meal planning app
- Layered architecture: handlers → services → storage (SQLite) with repository interfaces
- JWT auth with injectable JWTService, bcrypt passwords, prefixed UUID entity IDs
- External dependency on Tjek API for grocery offers (cached with 1-hour TTL)
- Deployed on Fly.io with Docker, embedded migrations, graceful shutdown
- **v1.0 shipped:** 93 findings across 7 areas, 9-batch fix plan
- **v1.1 shipped:** All critical/high/medium findings implemented, 53/53 verified
- 4,653 lines of Go across 32 backend files
- Tech stack: Go 1.24, SQLite (WAL mode), stdlib `net/http`, `log/slog`

## Constraints

- **Tech stack**: Go 1.24, SQLite, stdlib `net/http` — no framework dependencies
- **Deployment**: Fly.io with Docker, scales to zero
- **Testing**: Thin coverage (1 test file) — repository interfaces enable future testing

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Review backend only, exclude frontend | Focused scope, manageable context | ✓ Good |
| Report + fix plan, no immediate code changes | Clean separation of analysis and action | ✓ Good |
| Severity rating system (critical/high/medium/low) | Consistent across all 93 findings | ✓ Good |
| 9-batch fix plan with dependency ordering | Sprint-sized batches minimize context-switching | ✓ Good |
| JWT_SECRET min 32 chars, validated at startup | Prevents weak secrets in production | ✓ Good |
| Foreign keys via DSN `_foreign_keys=on` | Defense in depth for data integrity | ✓ Good |
| Sentinel errors in domain/errors.go | Type-safe error checking with `errors.Is()` | ✓ Good |
| DecodeJSON helper (size + type + decode) | Impossible to forget body limits | ✓ Good |
| sync.Mutex + map for TTL cache | Simpler than sync.Map for TTL expiry | ✓ Good |
| WAL mode with synchronous=NORMAL | Concurrent reads, safe with WAL | ✓ Good |
| Repository interfaces in domain package | Enables mock testing, future DB migration | ✓ Good |
| Injectable JWTService (no package state) | Testable, follows Go best practices | ✓ Good |
| Embedded migrations via embed.FS | Atomic deployments, no runtime dependencies | ✓ Good |
| Structured logging with log/slog | Machine-readable, enables log aggregation | ✓ Good |

---
*Last updated: 2026-02-10 after v1.1 Fix milestone*
