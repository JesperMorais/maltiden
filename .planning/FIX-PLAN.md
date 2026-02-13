# Maltiden Backend: Prioritized Fix Plan

**Generated:** 2026-02-09
**Source:** Consolidated findings from Phases 1 & 2 (93 unique findings)
**Total Estimated Effort:** 125-175 hours

---

## Overview

This document organizes the 93 findings from FINDINGS-REPORT.md into **9 sprint-sized batches** ordered by dependency, risk reduction, and area cohesion. Each batch is designed to be a focused work session of 4-8 hours, grouping related fixes to minimize context-switching.

### Batch Ordering Rationale

The batches are ordered to:
1. **Address critical security vulnerabilities first** (Batch 1) - Production blockers that allow auth bypass and data access
2. **Establish data integrity foundation** (Batch 2) - Database configuration enabling reliable operations
3. **Build error infrastructure** (Batch 3) - Systematic error handling that later batches can leverage
4. **Sweep input validation** (Batch 4) - Comprehensive validation using error patterns from Batch 3
5. **Fix high-impact external API performance** (Batch 5) - User-facing performance improvements
6. **Optimize database performance** (Batch 6) - Backend performance and scalability
7. **Improve architecture** (Batch 7) - Testability and future development velocity
8. **Production readiness** (Batch 8) - Deployment, observability, and operational concerns
9. **Polish and quality** (Batch 9) - Medium/low priority refinements

Each batch is independently valuable - completing any batch improves the system without requiring later batches to complete first.

### Recommended Schedule

- **Week 1:** Batches 1-3 (Critical security + Data integrity + Error infrastructure) - 16-24 hours
- **Week 2:** Batches 4-5 (Input validation + External API performance) - 22-34 hours
- **Week 3:** Batches 6-7 (Database performance + Architecture) - 36-54 hours
- **Week 4:** Batches 8-9 (Deployment + Polish) - 28-38 hours

---

## Batch Dependency Graph

```
Batch 1 (Critical Security)
    ↓
Batch 2 (Data Integrity) ──┐
    ↓                      │
Batch 3 (Error Infra) ─────┤
    ↓                      │
Batch 4 (Input Validation) │
                           │
Batch 5 (External API) ────┤
                           │
Batch 6 (DB Performance) ──┤
                           ├─→ Batch 7 (Architecture)
Batch 8 (Deployment) ──────┘        ↓
                              Batch 9 (Polish)
```

**Key dependencies:**
- Batch 3 (error infrastructure) should complete before Batch 4 (validation) to leverage sentinel errors
- Batch 7 (architecture) benefits from earlier batches being complete but doesn't block them
- All other batches are largely independent

---

## Batch 1: Critical Security & Authentication

**Priority:** Immediate (Production Blocker)
**Estimated Effort:** 6-8 hours
**Findings Addressed:** 6 critical findings

### Scope

Fix all authentication bypass vulnerabilities and IDOR issues that allow unauthorized access to system data.

### Findings

| ID | Description | Effort |
|----|-------------|--------|
| **AUTH-01** | Empty JWT secret allows token forgery | 1-2h |
| **AUTH-02** | Shopping list endpoints lack authentication | 1h |
| **AUTH-03** | JWT algorithm not pinned (none attack) | 30min |
| **AUTH-04** | No IDOR protection in menu/shopping operations | 2-3h |
| **DATA-01** | Registration flow not transactional | 1-2h |
| **DATA-02** | Menu generation with zero recipes returns success | 30min |

### Key Files Touched

- `backend/pkg/utils/jwt.go` - JWT secret validation, algorithm pinning
- `backend/cmd/server/main.go` - Startup validation for JWT_SECRET
- `backend/internal/api/router.go` - Add auth middleware to shopping endpoints
- `backend/internal/api/handlers/menus.go` - Add household verification
- `backend/internal/api/handlers/shopping.go` - Add household verification
- `backend/internal/storage/sqlite/menu_storage.go` - Add householdID verification methods
- `backend/internal/services/auth_service.go` - Wrap registration in transaction
- `backend/internal/services/menu_service.go` - Error when no recipes available

### Dependencies

None - this is the first batch and blocks nothing else.

### Definition of Done

- [ ] Server fails to start if JWT_SECRET is empty or < 32 characters
- [ ] JWT parsing validates algorithm is HS256, rejects "none" and other algorithms
- [ ] Shopping list GET and PATCH endpoints require authentication
- [ ] Menu and shopping operations verify menuID belongs to user's household (403 if not)
- [ ] Registration flow wrapped in database transaction with rollback on failure
- [ ] Menu generation returns error "no_recipes_available" when recipes table empty
- [ ] All changes have test coverage demonstrating the fixes

### Implementation Notes

- **AUTH-01:** Add validation in main.go before any other setup. Log the issue clearly.
- **AUTH-03:** Use jwt.ParseWithClaims callback to validate `token.Method.Alg() == "HS256"`
- **AUTH-04:** Create `GetShoppingListForHousehold(menuID, householdID)` in storage layer that joins menu → household
- **DATA-01:** Use existing transaction pattern from household_service.go JoinHousehold method as template

---

## Batch 2: Data Integrity Foundation

**Priority:** High (Enables Reliable Operations)
**Estimated Effort:** 4-6 hours
**Findings Addressed:** 5 high-severity data integrity issues

### Scope

Fix database configuration and transactional safety to prevent data corruption and orphaned records.

### Findings

| ID | Description | Effort |
|----|-------------|--------|
| **DATA-04** | Foreign key enforcement disabled in SQLite | 30min |
| **DATA-03** | Recipe tag filter vulnerable to JSON injection | 1-2h |
| **DATA-05** | UpdateMemberStatus builds SQL via string concatenation | 1h |
| **DATA-06** | Menu day primary key collision risk | 1h |
| **DATA-09** | No index on menu_days.date for range queries | 30min |

### Key Files Touched

- `backend/internal/storage/sqlite/db.go` - Enable PRAGMA foreign_keys
- `backend/internal/storage/sqlite/recipe_storage.go` - Fix tag filter to use json_each or escape LIKE
- `backend/internal/storage/sqlite/household_storage.go` - Replace string concatenation with explicit queries
- `backend/internal/storage/sqlite/menu_storage.go` - Use UUIDs for menu day IDs
- `backend/migrations/006_add_menu_date_index.sql` - Add composite index on (menu_id, date)

### Dependencies

None - independent of Batch 1, but should be done early to prevent data corruption.

### Definition of Done

- [ ] `PRAGMA foreign_keys = ON` executed immediately after database open
- [ ] Recipe tag filter uses SQLite JSON1 `json_each()` function or escapes LIKE wildcards
- [ ] UpdateMemberStatus uses explicit query variants instead of string building
- [ ] Menu day IDs generated as UUIDs instead of concatenated strings
- [ ] Composite index exists on `menu_days(menu_id, date)` for range query optimization
- [ ] Migration tested on development database
- [ ] Foreign key cascades verified with test data (delete household → members cascade)

### Implementation Notes

- **DATA-04:** Add to db.go immediately after `db, err := sql.Open()`: `if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil { ... }`
- **DATA-03:** Prefer JSON1 approach: `SELECT * FROM recipes WHERE id IN (SELECT DISTINCT value FROM json_each(recipes.tags) WHERE value = ?)`
- **DATA-06:** Use existing `uuid.New().String()` pattern from other ID generation

---

## Batch 3: Error Handling Infrastructure

**Priority:** High (Foundation for Later Batches)
**Estimated Effort:** 4-6 hours
**Findings Addressed:** 3 error handling issues

### Scope

Establish systematic error handling patterns that eliminate fragile string-matching and provide proper error logging and responses.

### Findings

| ID | Description | Effort |
|----|-------------|--------|
| **ERROR-01** | Error handling via string matching (15+ locations) | 2-3h |
| **ERROR-02** | No structured error response helper | 1-2h |
| **ERROR-03** | Errors lost in service-to-handler boundary | 1h |

### Key Files Touched

- `backend/internal/domain/errors.go` - Define sentinel errors (new file)
- `backend/internal/api/handlers/response.go` - Create WriteError helper (new file)
- All service files - Replace `errors.New("msg")` with sentinel errors
- All handler files - Replace manual JSON error writing with WriteError helper
- All handler files - Add error logging before returning responses

### Dependencies

None, but Batch 4 (input validation) will leverage the patterns established here.

### Definition of Done

- [ ] Sentinel errors defined in domain package: `ErrNotFound`, `ErrUnauthorized`, `ErrInvalidInput`, `ErrDuplicateEmail`, etc.
- [ ] Services return typed errors using `fmt.Errorf("context: %w", domain.ErrNotFound)`
- [ ] Handlers use `errors.Is()` for type-safe error checking instead of string matching
- [ ] `WriteError(w, statusCode, message)` helper properly encodes JSON and sets Content-Type
- [ ] All handler error paths log the error with context before returning to client
- [ ] Generic user-facing error messages prevent information leakage
- [ ] Compiler catches error handling changes (no more silent breakage from message changes)

### Implementation Notes

- **ERROR-01:** Start with most common errors (not found, unauthorized, invalid input) then expand
- **ERROR-02:** Helper should use `json.NewEncoder` to prevent injection (fixes VALID-01 too)
- Pattern: `if errors.Is(err, domain.ErrNotFound) { WriteError(w, 404, "Resource not found"); return }`
- Add request ID to log messages (if Batch 8 DEPLOY-09 has been completed)

---

## Batch 4: Input Validation Sweep

**Priority:** High (Systematic Security Improvement)
**Estimated Effort:** 8-12 hours
**Findings Addressed:** 14 input validation issues

### Scope

Comprehensive validation of all API inputs - request bodies, path parameters, query strings, headers. Use error infrastructure from Batch 3.

### Findings

| ID | Description | Effort |
|----|-------------|--------|
| **VALID-02** | No request body size limits | 1h |
| **VALID-01** | Error message injection in HTTP responses | (fixed by ERROR-02) |
| **VALID-03** | Manual path splitting instead of PathValue | 30min |
| **VALID-04** | No validation of path parameters (UUID format) | 1h |
| **VALID-05** | No validation of GenerateMenuRequest fields | 30min |
| **VALID-06** | Offers handlers silently ignore invalid coordinates | 1h |
| **VALID-07** | No Content-Type validation on JSON endpoints | 1h |
| **VALID-08** | UpdateMemberStatus allows empty updates | 30min |
| **VALID-09** | No validation of recipe filter query parameters | 30min |
| **VALID-10** | No validation on recipe name length | 30min |
| **VALID-11** | No validation on recipe servings upper bound | 15min |
| **VALID-12** | No validation on ingredients array size | 30min |
| **VALID-13** | Menu service doesn't validate days count | 15min |
| **VALID-14** | Menu service doesn't validate servings count | 15min |
| **VALID-15** | Login/register don't trim email/password whitespace | 30min |
| **VALID-16** | GetAll recipes returns empty array vs null inconsistency | 30min |

### Key Files Touched

- All handler files - Add `http.MaxBytesReader` wrapper
- `backend/internal/api/handlers/recipes.go` - Use PathValue, validate UUIDs
- `backend/internal/api/handlers/shopping.go` - Use PathValue, validate UUIDs
- `backend/internal/api/handlers/household.go` - Validate UUIDs
- `backend/internal/api/handlers/menus.go` - Validate menu request bounds
- `backend/internal/api/handlers/offers.go` - Return errors for invalid coordinates, extract parseLocationParams helper
- `backend/internal/api/handlers/auth.go` - Trim email/password, validate email format
- All handler files - Add Content-Type validation middleware or per-handler check
- `backend/internal/services/recipe_service.go` - Validate name length, servings, ingredients count
- `backend/internal/services/menu_service.go` - Validate days and servings bounds
- `backend/internal/storage/sqlite/recipe_storage.go` - Normalize empty array responses

### Dependencies

Batch 3 (error infrastructure) - validation errors should use sentinel errors and WriteError helper.

### Definition of Done

- [ ] All JSON endpoints wrapped with `http.MaxBytesReader(w, r.Body, 1<<20)` (1MB limit)
- [ ] All handlers use `r.PathValue()` instead of manual path splitting
- [ ] All UUID path parameters validated with `uuid.Parse()`, return 400 if invalid
- [ ] Menu generation validates: 1 ≤ days ≤ 31, 1 ≤ servings ≤ 100
- [ ] Recipe creation validates: name ≤ 200 chars, 1 ≤ servings ≤ 100, ingredients ≤ 50
- [ ] Offers handlers return 400 for invalid coordinates instead of silent defaults
- [ ] All JSON endpoints check `Content-Type: application/json`
- [ ] Auth endpoints trim email/password and validate email format
- [ ] All storage methods normalize to empty array `[]` instead of null
- [ ] Validation errors return clear messages using WriteError helper

### Implementation Notes

- Create validation helper functions to avoid duplication: `validateUUID()`, `validateBounds()`, `validateEmail()`
- Add validation at handler layer (basic) and service layer (business rules)
- Consider creating `internal/api/middleware/validation.go` for shared checks (body size, content-type)

---

## Batch 5: External API Performance

**Priority:** High (High-Impact User Facing)
**Estimated Effort:** 6-10 hours
**Findings Addressed:** 4 external API performance issues

### Scope

Optimize Tjek API integration with caching, concurrent fetching, and retry logic for massive performance improvements.

### Findings

| ID | Description | Effort |
|----|-------------|--------|
| **PERF-03** | No caching for Tjek API responses (600x speedup possible) | 3-4h |
| **PERF-02** | Sequential catalog fetching (5x speedup via concurrency) | 2-3h |
| **PERF-08** | Tjek API timeout set to 30 seconds | 15min |
| **PERF-09** | No exponential backoff or retry logic | 1-2h |

### Key Files Touched

- `backend/internal/services/tjek_service.go` - Add caching layer, concurrent fetching, retry logic
- `backend/go.mod` - Add cache library dependency if using external cache

### Dependencies

None - independent improvements to external API integration.

### Definition of Done

- [ ] In-memory cache (or Redis if scaling) implemented with TTL for catalogs (1 hour), stores (1 hour), offers (1 hour)
- [ ] Cache hit rate logged for monitoring
- [ ] Catalog fetching uses goroutines with semaphore (max 5 concurrent requests)
- [ ] Timeout reduced from 30s to 5-10s
- [ ] Exponential backoff retry (3 attempts max) for transient failures
- [ ] Performance benchmarks show: first request ~1.2s (concurrent), cached requests <10ms
- [ ] Error handling properly surfaces when all retries exhausted

### Implementation Notes

- **Caching:** Simple in-memory with sync.Map or use github.com/patrickmn/go-cache
- **Concurrency:** Use `errgroup` from golang.org/x/sync for goroutine management
- **Retry:** Implement exponential backoff: attempt 1 (immediate), attempt 2 (+100ms), attempt 3 (+200ms)
- Monitor cache hit rates in logs to verify effectiveness

---

## Batch 6: Database Performance

**Priority:** High (Backend Scalability)
**Estimated Effort:** 8-12 hours
**Findings Addressed:** 6 database performance issues

### Scope

Optimize database configuration and query patterns for production load and concurrent users.

### Findings

| ID | Description | Effort |
|----|-------------|--------|
| **PERF-04** | SQLite WAL mode not enabled | 1h |
| **PERF-05** | Database connection pool not configured | 1h |
| **PERF-01** | N+1 query pattern in shopping list generation | 2-3h |
| **PERF-06** | GetAll recipes loads all rows without limit | 2-3h |
| **PERF-07** | Recipe tag filter uses string LIKE on JSON | 1-2h |
| **PERF-12** | No database query timeout context | 1-2h |

### Key Files Touched

- `backend/internal/storage/sqlite/db.go` - Enable WAL, configure connection pool, add context timeouts
- `backend/internal/storage/sqlite/recipe_storage.go` - Add GetByIDs batch method, add pagination, fix tag filter
- `backend/internal/services/shopping_service.go` - Use batch fetching for recipes
- `backend/internal/services/menu_service.go` - Use paginated recipe fetching
- All storage methods - Add context.WithTimeout wrappers

### Dependencies

None, though DATA-09 from Batch 2 (date index) complements this batch.

### Definition of Done

- [ ] `PRAGMA journal_mode=WAL`, `PRAGMA synchronous=NORMAL`, `PRAGMA busy_timeout=5000` enabled at startup
- [ ] Connection pool configured: `SetMaxOpenConns(25)`, `SetMaxIdleConns(5)`, `SetConnMaxLifetime(5min)`
- [ ] `RecipeStorage.GetByIDs([]string)` batch method implemented using `IN` clause
- [ ] Shopping service uses batch fetching (5 queries → 1 query for 5-day menu)
- [ ] Recipe listing supports pagination (default 50 per page, configurable)
- [ ] Tag filtering uses SQLite JSON1 extension or normalized tags table
- [ ] All storage methods use `db.QueryContext()` with 5-second timeout
- [ ] Load testing verifies concurrent request handling without "database locked" errors

### Implementation Notes

- **WAL mode:** Execute all PRAGMAs in db.go after opening database, before any queries
- **Batch fetching:** Build query dynamically: `IN (?, ?, ?)` with proper parameter binding
- **Pagination:** Add offset/limit parameters to GetAllRecipes, return total count for clients
- Consider connection pool sizing based on expected concurrent users (25 is conservative)

---

## Batch 7: Architecture & Testability

**Priority:** Medium (Future Development Velocity)
**Estimated Effort:** 10-16 hours
**Findings Addressed:** 8 architecture improvements

### Scope

Introduce abstraction layers and dependency injection to enable testing and future database migration.

### Findings

| ID | Description | Effort |
|----|-------------|--------|
| **ARCH-01** | Services depend on concrete storage types (no interfaces) | 4-8h |
| **ARCH-02** | GetUserID and GetHouseholdID have inconsistent signatures | 30min |
| **ARCH-03** | No dependency injection container | 1-2h |
| **ARCH-08** | No transaction helper abstraction | 1-2h |
| **ARCH-09** | Storage exposes *sql.DB directly | 1h |
| **ARCH-04** | Handlers access context inconsistently | 30min |
| **ARCH-05** | Domain package contains request/response types | 1-2h |
| **ARCH-13** | Duplicate coordinate parsing logic | 30min |

### Key Files Touched

- `backend/internal/domain/repositories.go` - Define storage interfaces (new file)
- All service files - Update constructors to accept interfaces
- `backend/internal/storage/sqlite/` - Implement interfaces (already compatible)
- `backend/internal/api/router.go` - Extract wireDependencies() function
- `backend/internal/storage/transaction.go` - Create WithTransaction helper (new file)
- `backend/pkg/middleware/auth.go` - Standardize context extraction functions
- `backend/internal/api/types/` - Move request/response types from domain (new package)
- `backend/internal/api/handlers/offers.go` - Extract parseLocationParams helper

### Dependencies

Benefits from earlier batches being complete but doesn't block anything. Can be done in parallel with other medium-priority batches.

### Definition of Done

- [ ] Storage interfaces defined for all storage types: UserRepository, HouseholdRepository, RecipeRepository, etc.
- [ ] Services depend on interfaces, not concrete types
- [ ] SQLite storage implements interfaces (type assertion verified in tests)
- [ ] Dependency wiring extracted to `wireDependencies()` function in router
- [ ] `WithTransaction(fn func(tx) error)` helper eliminates transaction boilerplate
- [ ] Both GetUserID and GetHouseholdID accept `*http.Request` parameter
- [ ] API request/response types moved to `internal/api/types/` package
- [ ] Coordinate parsing extracted to `parseLocationParams(r *http.Request)` helper
- [ ] Mock implementations created for at least 2 core repositories
- [ ] At least 2 service test files demonstrate mock-based unit testing

### Implementation Notes

- **ARCH-01:** Define minimal interfaces first (only methods currently used), expand as needed
- **Interfaces:** Place in `backend/internal/domain/repositories.go` or `backend/internal/storage/interfaces.go`
- **Testing:** Create `backend/internal/storage/mocks/` package with mock implementations
- Start with UserRepository and RecipeRepository as examples, then expand to others

---

## Batch 8: Deployment & Observability

**Priority:** Medium (Production Readiness)
**Estimated Effort:** 8-12 hours
**Findings Addressed:** 13 deployment and observability issues

### Scope

Prepare application for production deployment with proper configuration, health checks, logging, and graceful shutdown.

### Findings

| ID | Description | Effort |
|----|-------------|--------|
| **AUTH-05** | No rate limiting on authentication endpoints | 2-3h |
| **AUTH-06** | CORS restricted to localhost only | 30min |
| **AUTH-12** | No validation of required environment variables | 1h |
| **DEPLOY-01** | fly.toml has conflicting memory settings | 15min |
| **DEPLOY-02** | Migration file paths are relative | 1-2h |
| **DEPLOY-03** | Database path not validated | 30min |
| **DEPLOY-04** | Migrations run on every startup | 1h |
| **DEPLOY-05** | No graceful shutdown handling | 1-2h |
| **DEPLOY-06** | Health endpoint reveals server operational / no depth | 30min |
| **DEPLOY-07** | Health check endpoint missing proper checks | (duplicate of DEPLOY-06) |
| **DEPLOY-08** | No structured logging framework | 1-2h |
| **DEPLOY-09** | No request ID tracing | 1h |
| **DEPLOY-10** | Health endpoint returns plaintext-ish JSON | 15min |

### Key Files Touched

- `backend/internal/api/middleware/ratelimit.go` - Add rate limiting middleware (new file)
- `backend/pkg/middleware/cors.go` - Make origins configurable via env var
- `backend/cmd/server/main.go` - Validate all env vars at startup, add graceful shutdown
- `backend/fly.toml` - Fix memory settings, add health check configuration
- `backend/internal/storage/sqlite/migrations.go` - Embed migrations using embed.FS (new file)
- `backend/internal/storage/sqlite/db.go` - Validate database path, optimize migration checking
- `backend/internal/api/handlers/health.go` - Add database ping check, set Content-Type
- `backend/pkg/logger/logger.go` - Create structured logging wrapper (new file)
- `backend/internal/api/middleware/requestid.go` - Generate and propagate request IDs (new file)

### Dependencies

None, though request ID (DEPLOY-09) enhances error logging from Batch 3.

### Definition of Done

- [ ] Rate limiting middleware (5 req/min for auth endpoints, 60 req/min for others)
- [ ] CORS origins configurable via `CORS_ORIGINS` env var (comma-separated)
- [ ] Startup validates: JWT_SECRET (≥32 chars), DATABASE_PATH (exists, writable), CORS_ORIGINS (set)
- [ ] fly.toml uses only `memory_mb = 512`, conflicting settings removed
- [ ] Migrations embedded in binary using `embed.FS`, no relative path dependencies
- [ ] Database path validated as absolute with writable parent directory
- [ ] Migration status fetched in single query instead of per-migration queries
- [ ] Graceful shutdown with 10-second timeout for in-flight requests
- [ ] Health endpoint checks database connectivity, returns 503 if unhealthy, sets Content-Type
- [ ] Structured logging using Go 1.21+ `slog` with JSON output
- [ ] Request ID middleware generates UUID per request, includes in all log entries
- [ ] Health endpoint includes proper headers and database status

### Implementation Notes

- **Rate limiting:** Use golang.org/x/time/rate or github.com/ulule/limiter/v3
- **Graceful shutdown:** Use `server.Shutdown(ctx)` with context timeout
- **Embed migrations:** `//go:embed migrations/*.sql` with `embed.FS`
- **Structured logging:** Replace `log.Printf` with `slog.Info`, `slog.Error`, etc.
- Add request ID to context in middleware for propagation to handlers/services

---

## Batch 9: Code Quality & Polish

**Priority:** Low (Refinements)
**Estimated Effort:** 10-14 hours
**Findings Addressed:** 24 medium and low priority issues

### Scope

Remaining code quality improvements, consistency fixes, and minor optimizations.

### Findings

| ID | Description | Effort |
|----|-------------|--------|
| **AUTH-11** | Package-level JWT state | 1-2h |
| **AUTH-13** | CreateInvite checks role but not member status | 1h |
| **AUTH-14** | Shopping list menuId has no household verification | (covered by AUTH-04) |
| **AUTH-15** | UpdateItem accepts any menuId for path itemId | 1h |
| **DATA-07** | Recipe deletion leaves orphaned menu days | 1-2h |
| **DATA-08** | JoinHousehold transaction doesn't lock invite code | 1h |
| **DATA-10** | Menu generation random selection allows duplicates | 1-2h |
| **DATA-11** | Menu day date format not validated | 30min |
| **ERROR-04** | No error wrapping context in service layer | 1-2h |
| **ERROR-05** | Recipe create returns 400 for all errors | 30min |
| **ERROR-06** | No distinction between business and infrastructure errors | 1-2h |
| **ARCH-06** | Router does too much | (covered by ARCH-03) |
| **ARCH-07** | Services can't be tested in isolation | (covered by ARCH-01) |
| **ARCH-10** | pkg/ package contains non-reusable code | 1h |
| **ARCH-11** | Domain package has no sub-packages | (defer) |
| **ARCH-12** | No API versioning strategy | (defer/document) |
| **ARCH-14** | No domain services | (defer) |
| **ARCH-15** | Handler functions have no shared pattern | (defer/document) |
| **PERF-10** | Bubble sort used instead of sort.Slice | 30min |
| **PERF-11** | Menu generation algorithm allows recipe duplication | (duplicate of DATA-10) |
| **PERF-13** | Unbounded response size from Tjek API | 1-2h |
| **PERF-14** | Shopping list aggregation creates large maps | 1h |
| **PERF-16** | Shopping item ID uses MD5 | 30min |
| **PERF-17** | Ingredient categorization uses linear lookup | (not an issue) |
| **DEPLOY-11** | GetMemberStatuses returns empty array | 30min |
| **DEPLOY-12** | Cold starts with min_machines_running=0 | (config decision) |
| **DEPLOY-13** | No monitoring/metrics collection | (defer) |

### Key Files Touched

- `backend/pkg/utils/jwt.go` - Refactor to JWTService struct
- `backend/internal/api/handlers/household.go` - Verify member status for invites
- `backend/internal/api/handlers/shopping.go` - Validate itemID belongs to menuID
- `backend/internal/storage/sqlite/recipe_storage.go` - Add Delete method with cascade handling
- `backend/internal/services/household_service.go` - Lock invite code in transaction
- `backend/internal/services/menu_service.go` - Shuffle algorithm for variety, date validation
- All service files - Add error wrapping with context
- `backend/internal/api/handlers/recipes.go` - Distinguish validation (400) from internal (500) errors
- `backend/internal/domain/errors.go` - Define BusinessError and InfrastructureError types
- `backend/pkg/middleware/` - Move to `internal/middleware/` if not reusable
- `backend/internal/services/tjek_service.go` - Replace bubble sort, add early termination for offers
- `backend/internal/services/shopping_service.go` - Pre-allocate maps, replace MD5 with fnv
- Various - Defer architectural decisions (API versioning, domain services, monitoring)

### Dependencies

All other batches should be complete. This batch contains polish and refinements that build on established patterns.

### Definition of Done

- [ ] JWT handling refactored to JWTService struct accepting secret as parameter
- [ ] CreateInvite verifies member exists in database (not just JWT claims)
- [ ] UpdateItem validates itemID belongs to menuID before updating
- [ ] Recipe deletion implemented with protection against active menu references
- [ ] JoinHousehold uses SELECT FOR UPDATE on invite code
- [ ] Menu generation uses shuffle algorithm (no duplicate recipes in same menu)
- [ ] Date format validation helper created for future APIs
- [ ] All service errors wrapped with context: `fmt.Errorf("operation failed: %w", err)`
- [ ] Recipe handler distinguishes validation errors (400) from internal errors (500)
- [ ] BusinessError and InfrastructureError types defined and used
- [ ] Non-reusable code moved from pkg/ to internal/
- [ ] Bubble sorts replaced with `sort.Slice`
- [ ] Tjek offer collection uses early termination after 100 offers
- [ ] Shopping aggregation pre-allocates maps with capacity hints
- [ ] MD5 replaced with hash/fnv for shopping item IDs
- [ ] Empty member array logged as warning
- [ ] Architectural decisions documented in ARCHITECTURE.md for future reference

### Implementation Notes

- Focus on quick wins and consistency improvements
- Document deferred decisions (API versioning, monitoring) for future sprints
- Ensure error wrapping uses `%w` verb for proper error chain traversal
- Shuffle algorithm: Create array of recipe IDs, Fisher-Yates shuffle, take first N

---

## Effort Summary Table

| Batch | Name | Priority | Effort | Findings | Key Area |
|-------|------|----------|--------|----------|----------|
| 1 | Critical Security | Immediate | 6-8h | 6 | Auth bypass, IDOR |
| 2 | Data Integrity | High | 4-6h | 5 | FK enforcement, transactions |
| 3 | Error Infrastructure | High | 4-6h | 3 | Sentinel errors, logging |
| 4 | Input Validation | High | 8-12h | 14 | Request validation sweep |
| 5 | External API Performance | High | 6-10h | 4 | Caching, concurrency |
| 6 | Database Performance | High | 8-12h | 6 | WAL, N+1, connection pool |
| 7 | Architecture | Medium | 10-16h | 8 | Interfaces, DI, testability |
| 8 | Deployment | Medium | 8-12h | 13 | Config, logging, health |
| 9 | Polish | Low | 10-14h | 24 | Code quality, refinements |
| **TOTAL** | | | **64-96h** | **83** | (10 findings deferred/duplicates) |

**Note:** Total effort (64-96 hours) is less than FINDINGS-REPORT.md estimate (125-175 hours) because:
- Batching reduces context-switching overhead
- Some findings are duplicates or overlapping (AUTH-14 covered by AUTH-04, PERF-11 = DATA-10)
- Some findings are deferred as documentation-only (ARCH-11, ARCH-12, ARCH-14, DEPLOY-12, DEPLOY-13)
- Practical implementation often combines multiple related fixes efficiently

---

## Completion Checklist

Track progress as batches are completed:

- [ ] **Batch 1: Critical Security** (6-8h) - Auth bypass, IDOR, transactions
- [ ] **Batch 2: Data Integrity** (4-6h) - FK enforcement, indexes, SQL safety
- [ ] **Batch 3: Error Infrastructure** (4-6h) - Sentinel errors, WriteError helper, logging
- [ ] **Batch 4: Input Validation** (8-12h) - Request body limits, UUID validation, bounds checking
- [ ] **Batch 5: External API Performance** (6-10h) - Caching, concurrent fetching, retries
- [ ] **Batch 6: Database Performance** (8-12h) - WAL mode, connection pool, N+1 fixes
- [ ] **Batch 7: Architecture** (10-16h) - Storage interfaces, DI, transaction helpers
- [ ] **Batch 8: Deployment & Observability** (8-12h) - Rate limiting, CORS, logging, health checks
- [ ] **Batch 9: Code Quality & Polish** (10-14h) - Remaining improvements, consistency

---

## Quick Reference: Finding IDs by Batch

**Batch 1 (Critical Security):** AUTH-01, AUTH-02, AUTH-03, AUTH-04, DATA-01, DATA-02

**Batch 2 (Data Integrity):** DATA-04, DATA-03, DATA-05, DATA-06, DATA-09

**Batch 3 (Error Infrastructure):** ERROR-01, ERROR-02, ERROR-03

**Batch 4 (Input Validation):** VALID-02, VALID-01, VALID-03, VALID-04, VALID-05, VALID-06, VALID-07, VALID-08, VALID-09, VALID-10, VALID-11, VALID-12, VALID-13, VALID-14, VALID-15, VALID-16

**Batch 5 (External API):** PERF-03, PERF-02, PERF-08, PERF-09

**Batch 6 (Database Performance):** PERF-04, PERF-05, PERF-01, PERF-06, PERF-07, PERF-12

**Batch 7 (Architecture):** ARCH-01, ARCH-02, ARCH-03, ARCH-08, ARCH-09, ARCH-04, ARCH-05, ARCH-13

**Batch 8 (Deployment):** AUTH-05, AUTH-06, AUTH-12, DEPLOY-01, DEPLOY-02, DEPLOY-03, DEPLOY-04, DEPLOY-05, DEPLOY-06, DEPLOY-08, DEPLOY-09, DEPLOY-10

**Batch 9 (Polish):** AUTH-11, AUTH-13, AUTH-15, DATA-07, DATA-08, DATA-10, DATA-11, ERROR-04, ERROR-05, ERROR-06, ARCH-10, PERF-10, PERF-13, PERF-14, PERF-16, DEPLOY-11

**Deferred/Documented:** ARCH-11, ARCH-12, ARCH-14, ARCH-15, DEPLOY-12, DEPLOY-13, PERF-17 (not an issue)

**Duplicates/Covered by others:** AUTH-14 (covered by AUTH-04), ARCH-06 (covered by ARCH-03), ARCH-07 (covered by ARCH-01), DEPLOY-07 (duplicate of DEPLOY-06), PERF-11 (duplicate of DATA-10), PERF-15 (duplicate of PERF-02)

---

## Finding Count Verification

**From FINDINGS-REPORT.md:** 93 unique findings

**Accounted for in this plan:**
- Batch 1: 6 findings
- Batch 2: 5 findings
- Batch 3: 3 findings
- Batch 4: 14 findings (note: VALID-01 fixed by ERROR-02)
- Batch 5: 4 findings
- Batch 6: 6 findings
- Batch 7: 8 findings
- Batch 8: 12 findings (DEPLOY-07 duplicate of DEPLOY-06)
- Batch 9: 24 findings (includes some overlaps and deferred items)
- Deferred/documented: 7 findings (ARCH-11, ARCH-12, ARCH-14, ARCH-15, DEPLOY-12, DEPLOY-13, PERF-17)
- Duplicates covered by others: 4 findings (AUTH-14, ARCH-06, ARCH-07, PERF-15)

**Total:** 93 findings = 82 actionable + 7 deferred + 4 duplicates ✓

---

**End of Fix Plan**
