---
phase: 07-architecture-deployment-polish
plan: 02
subsystem: deployment-infrastructure
tags: [deployment, logging, observability, health-check, graceful-shutdown, embedded-migrations]
requires: [07-01]
provides: [production-ready-backend, embedded-migrations, structured-logging, health-monitoring]
affects: [deployment, monitoring, operations]
tech-stack:
  added: [log/slog, embed.FS, crypto/rand]
  patterns: [graceful-shutdown, structured-logging, embedded-resources, request-tracing]
key-files:
  created:
    - backend/migrations/embed.go
    - backend/pkg/middleware/requestid.go
  modified:
    - backend/internal/storage/sqlite/db.go
    - backend/cmd/server/main.go
    - fly.toml
    - backend/pkg/middleware/cors.go
    - backend/internal/api/handlers/health.go
    - backend/internal/api/router.go
decisions:
  - name: Embed migrations using embed.FS
    rationale: Eliminates runtime dependency on migration files, ensures atomic deployments
    impact: Migrations are compiled into the binary, no external files needed
  - name: Set memory_mb to 512 (removed conflicting memory='1gb')
    rationale: Fly.io requires single memory specification
    impact: Production deployment config is now valid
  - name: Graceful shutdown with 10s timeout
    rationale: Allows in-flight requests to complete before termination
    impact: Better reliability during deployments and scaling
  - name: Structured logging with log/slog JSON handler
    rationale: Machine-readable logs for production monitoring
    impact: Enables log aggregation and analysis tools
  - name: Configurable CORS origins via CORS_ORIGINS env var
    rationale: Production environments need different origins than development
    impact: Security - can restrict to production frontend domains
  - name: Health endpoint pings database with 2s timeout
    rationale: Load balancers need to know if backend can serve traffic
    impact: Enables proper health checks and auto-recovery
  - name: Request ID middleware generates 16-char hex IDs
    rationale: Enables request tracing across logs and services
    impact: Better debugging and observability
metrics:
  duration: 7m
  completed: 2026-02-10
---

# Phase 07 Plan 02: Production Deployment Configuration Summary

**One-liner:** Embedded migrations, graceful shutdown, structured JSON logging, configurable CORS, DB health checks, and request ID tracing.

## What Was Built

This plan made the backend production-ready by addressing deployment, operational, and observability gaps identified in FIX-PLAN Batch 8.

### 1. Embedded Migrations (DEPLOY-02, DEPLOY-04)

**Problem:** Migrations loaded from filesystem at runtime, creating deployment fragility.

**Solution:**
- Created `backend/migrations/embed.go` with `//go:embed *.sql` directive
- Updated `db.go` to accept `embed.FS` parameter and use `fs.ReadFile()`
- Optimized migration checking: single query fetching all applied versions into a map instead of N queries
- Removed `os.ReadFile()` dependency

**Impact:** Migrations are now compiled into the binary. No external files needed. Atomic deployments guaranteed.

### 2. Fixed fly.toml (DEPLOY-01)

**Problem:** Conflicting memory settings (`memory = '1gb'` and `memory_mb = 256`)

**Solution:** Removed `memory = '1gb'` line, kept `memory_mb = 512` (bumped from 256 for safety)

**Impact:** Deployment config is now valid for Fly.io.

### 3. Graceful Shutdown (DEPLOY-05)

**Problem:** Server terminated immediately on SIGTERM, dropping in-flight requests.

**Solution:**
- Replaced `log.Fatal(http.ListenAndServe(...))` with proper signal handling
- Created `http.Server` and ran `ListenAndServe()` in goroutine
- Added signal channel for SIGINT/SIGTERM
- Implemented `srv.Shutdown(ctx)` with 10-second timeout

**Impact:** In-flight requests complete before shutdown. Better reliability during deployments and scaling.

### 4. Structured Logging (DEPLOY-08)

**Problem:** Unstructured text logs difficult to parse and analyze.

**Solution:**
- Replaced `log` package with `log/slog` in `main.go`
- Set default handler to `slog.NewJSONHandler(os.Stdout, nil)`
- Converted all `log.Printf()` calls to `slog.Info()` with structured fields
- Converted `log.Fatal()` to `slog.Error()` + `os.Exit(1)` (slog has no Fatal)

**Impact:** Machine-readable JSON logs enable aggregation, filtering, and analysis in production monitoring tools.

### 5. Configurable CORS (AUTH-06)

**Problem:** Hardcoded localhost origins insufficient for production.

**Solution:**
- Changed `CORS()` signature to accept `allowedOrigins []string`
- Built `map[string]bool` for O(1) origin lookup
- Updated `router.go` to read `CORS_ORIGINS` env var (comma-separated)
- Default: `"http://localhost:5173,http://localhost:4173,http://127.0.0.1:5173"` for development

**Impact:** Production can restrict CORS to production frontend domains. Better security posture.

### 6. Health Check with DB Connectivity (DEPLOY-06, DEPLOY-10)

**Problem:** `/health` endpoint returned OK even if database was down.

**Solution:**
- Created `HealthHandler` struct holding `*sql.DB`
- Implemented `Check()` method that:
  - Sets `Content-Type: application/json`
  - Pings database with 2-second timeout context
  - Returns `{"status":"ok","db":"connected"}` (200) if healthy
  - Returns `{"status":"error","db":"disconnected"}` (503) if database unreachable
- Wired `HealthHandler` in `router.go` dependencies

**Impact:** Load balancers can properly health-check the backend. Enables auto-recovery and better uptime.

### 7. Request ID Middleware (DEPLOY-09)

**Problem:** No way to trace requests across logs and services.

**Solution:**
- Created `backend/pkg/middleware/requestid.go`
- Generates 16-character hex request IDs using `crypto/rand`
- Adds `X-Request-ID` response header
- Stores ID in request context for access by handlers/services
- Provides `GetRequestID(r *http.Request) string` helper
- Wrapped router: `middleware.CORS(origins)(middleware.RequestID(mux))`

**Impact:** Request tracing enables debugging across distributed logs. Better observability.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] contextKey type already declared in auth.go**

- **Found during:** Task 2 - Request ID middleware
- **Issue:** `requestid.go` redeclared `type contextKey string` which already exists in `auth.go`
- **Fix:** Removed type declaration from `requestid.go`, reused existing type from middleware package
- **Files modified:** `backend/pkg/middleware/requestid.go`
- **Commit:** ee5d12c (part of Task 2 commit)

## Testing & Verification

All verification checks passed:

- ✅ Docker build succeeds (`docker build -t maltiden-verify .`)
- ✅ fly.toml has no conflicting memory settings (single `memory_mb = 512`)
- ✅ Migrations use `embed.FS` (no `os.ReadFile` calls)
- ✅ Server shuts down gracefully on SIGTERM (`srv.Shutdown(ctx)`)
- ✅ CORS origins configurable via `CORS_ORIGINS` env var
- ✅ Health endpoint returns JSON with `Content-Type` and DB status
- ✅ Request ID middleware generates and propagates IDs

## Architecture Impact

### Production Readiness Improvements

1. **Deployment:** Embedded migrations eliminate external file dependencies
2. **Reliability:** Graceful shutdown prevents dropped requests during deployments
3. **Observability:** Structured logging enables log aggregation and analysis
4. **Monitoring:** Health checks with DB connectivity enable proper load balancing
5. **Debugging:** Request IDs enable tracing across distributed logs
6. **Security:** Configurable CORS allows production origin restrictions

### Middleware Stack (Order Matters)

```
Request Flow:
1. CORS middleware (checks origin, adds CORS headers)
2. Request ID middleware (generates ID, adds to context and headers)
3. Router (routes to handlers)
4. Auth middleware (on protected routes)
5. Handler execution
```

## Known Limitations

### Deferred Items

**Rate Limiting (AUTH-05):** Not implemented in this plan.

**Rationale:** Rate limiting requires:
- Choosing a library (e.g., `golang.org/x/time/rate`) or custom implementation
- Deciding on rate limit policies (per-IP, per-user, per-endpoint)
- Adding Redis/in-memory store for distributed rate limiting
- Testing and tuning limits

**Recommendation:** Defer to dedicated rate limiting plan or document as future improvement. Can be added as new middleware without changing existing code.

## Commits

1. **4b383de** - `feat(07-02): embed migrations, fix fly.toml, add graceful shutdown and structured logging`
   - Embedded migrations using `embed.FS`
   - Fixed fly.toml memory conflict
   - Added graceful shutdown with 10s timeout
   - Replaced `log` with `log/slog` for structured logging

2. **ee5d12c** - `feat(07-02): configurable CORS, health check with DB ping, and request ID middleware`
   - Made CORS origins configurable via env var
   - Created `HealthHandler` with DB connectivity check
   - Added request ID middleware with 16-char hex IDs
   - Wired all changes in router.go

## Next Phase Readiness

### Enables

- **Phase 7-03 (Polish):** Production-ready backend allows focus on final polish items
- **Deployment:** Backend can be deployed to production with proper health checks
- **Monitoring:** Structured logs and request IDs enable observability tooling

### Blockers

None.

### Concerns

None. All critical deployment and operational issues from Batch 8 are resolved.

## Summary Statistics

- **Tasks completed:** 2/2
- **Files created:** 2
- **Files modified:** 6
- **Commits:** 2
- **Duration:** 7 minutes
- **Findings addressed:** 10 (AUTH-05 deferred, AUTH-06, AUTH-12, DEPLOY-01, DEPLOY-02, DEPLOY-03, DEPLOY-04, DEPLOY-05, DEPLOY-06, DEPLOY-08, DEPLOY-09, DEPLOY-10)
