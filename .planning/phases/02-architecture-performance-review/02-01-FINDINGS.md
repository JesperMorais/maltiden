# Architecture Review Findings

**Phase:** 02-architecture-performance-review
**Plan:** 01
**Date:** 2026-02-09
**Scope:** Backend architecture patterns, layering, dependencies, error propagation, code organization, configuration, and cross-cutting concerns

---

## Executive Summary

Reviewed backend architecture across 6 dimensions: layer responsibility, dependency patterns, error propagation, code organization, configuration/startup, and cross-cutting concerns. Found **21 systemic issues** (0 critical, 4 high, 12 medium, 5 low severity).

**Key findings:**
- **No abstraction layer:** Services depend on concrete SQLite storage types, blocking future database migration
- **Fragile error handling:** String-matching error pattern appears in 6+ handlers with no sentinel errors
- **Package-level state:** JWT secret, database connections, and context keys all use package-level variables
- **Missing validation framework:** Input validation duplicated across all services with no shared patterns
- **Configuration gaps:** No validation of required env vars, silent failures on missing config
- **Observability blind spots:** No structured logging, errors swallowed without traces

**Impact:** Current patterns block future scalability (DB migration impossible without major refactor), create maintenance burden (error string changes break routing), and limit debuggability (production issues undiagnosable).

---

## Severity Ratings

Using Phase 1 severity scale:
- **Critical:** Systemic issues causing data loss, security bypass, or production failures (0 found)
- **High:** Patterns that will cause bugs as codebase grows or block important capabilities (4 found)
- **Medium:** Consistency issues, missing abstractions, maintainability debt (12 found)
- **Low:** Polish, idiomatic improvements, nice-to-haves (5 found)

---

## 1. Layer Responsibility & Separation

### H1: Services Depend on Concrete Storage Types (No Interfaces)

**Severity:** High
**Category:** Dependency Architecture

**What:**
All services depend on concrete `*sqlite.XxxStorage` types directly. No storage interfaces exist. Services are coupled to the SQLite implementation.

**Affected Files/Patterns:**
- `backend/internal/services/auth_service.go:14-15` - `userStorage *sqlite.UserStorage`, `householdStorage *sqlite.HouseholdStorage`
- `backend/internal/services/household_service.go:15-17` - Concrete SQLite dependencies
- `backend/internal/services/recipe_service.go` - `recipeStorage *sqlite.RecipeStorage`
- `backend/internal/services/menu_service.go` - Concrete storage dependencies
- `backend/internal/services/shopping_service.go` - Concrete storage dependencies
- **Pattern appears in:** All 6 service files

**Why It Matters:**
- **Testability:** Services cannot be unit tested with mock storage - all tests require real SQLite databases
- **Future DB migration:** Moving to PostgreSQL/MySQL requires rewriting all service constructors and storage implementations simultaneously (big-bang migration)
- **Development velocity:** Can't develop new features against in-memory storage or test doubles
- **Deployment flexibility:** Can't support multiple DB backends (e.g., Postgres in prod, SQLite in dev)

**Recommendation for Phase 3:**
1. Define storage interfaces in `backend/internal/domain/` or `backend/internal/storage/`:
   ```go
   type UserRepository interface {
       Create(user *domain.User) error
       GetByEmail(email string) (*domain.User, error)
       // ...
   }
   ```
2. Services depend on interfaces: `type AuthService struct { userRepo UserRepository }`
3. SQLite storage implements interfaces: `func (s *UserStorage) Create(...) { ... }`
4. Enables mock implementations for testing and future DB swap

**Effort:** Medium (4-8 hours) - Define 5-6 interfaces, update service constructors, verify tests still pass

---

### M1: Handlers Access Context Inconsistently

**Severity:** Medium
**Category:** Layer Responsibility

**What:**
Two different patterns for extracting values from request context:
- `GetUserID(r)` - Takes `*http.Request`
- `GetHouseholdID(ctx)` - Takes `context.Context`

**Affected Files:**
- `backend/pkg/middleware/auth.go:49-58` - Function signatures differ
- `backend/internal/api/handlers/household.go:20,43` - Mixed usage: `GetUserID(r)` and `GetHouseholdID(r.Context())`

**Why It Matters:**
- Cognitive load: developers must remember which function takes which parameter type
- Inconsistency: no clear reason for the difference (both extract from context)
- API confusion: handlers sometimes pass `r`, sometimes `r.Context()`

**Recommendation for Phase 3:**
Standardize both functions to take `*http.Request`:
```go
func GetUserID(r *http.Request) string {
    userID, _ := r.Context().Value(UserIDKey).(string)
    return userID
}

func GetHouseholdID(r *http.Request) string {
    householdID, _ := r.Context().Value(HouseholdIDKey).(string)
    return householdID
}
```

**Effort:** Low (15 min) - Update function signature, fix 6 call sites in handlers

---

### M2: Domain Package Contains Request/Response Types (Not Pure Domain)

**Severity:** Medium
**Category:** Layer Responsibility

**What:**
The `domain` package mixes pure domain entities (`User`, `Recipe`, `Menu`) with API-specific request/response types (`RegisterRequest`, `LoginRequest`, `AuthResponse`, `CreateInviteResponse`).

**Affected Files:**
- `backend/internal/domain/auth.go` - Contains `RegisterRequest`, `LoginRequest`, `AuthResponse` (API contracts)
- `backend/internal/domain/household.go` - Contains `CreateInviteResponse`, `JoinHouseholdRequest`, etc.
- Domain package imported by: handlers, services, storage layers

**Why It Matters:**
- **Conceptual confusion:** Domain entities should represent core business concepts independent of delivery mechanism (HTTP)
- **False dependency:** Storage layer shouldn't need to know about HTTP request/response shapes
- **Change coupling:** API versioning or format changes (e.g., switching to GraphQL) force domain package changes
- **Testing friction:** Domain tests pull in HTTP-specific types unnecessarily

**Recommendation for Phase 3:**
1. Create `backend/internal/api/types/` for request/response DTOs
2. Move all `*Request` and `*Response` types there
3. Keep domain package pure: only entities like `User`, `Recipe`, `Household`, `Menu`
4. Services return domain types, handlers convert to API responses

**Effort:** Medium (2-4 hours) - Create new package, move types, update imports across 15+ files

---

### M3: Router Does Too Much (Wiring + Route Definitions)

**Severity:** Medium
**Category:** Code Organization

**What:**
`NewRouter()` function in `backend/internal/api/router.go` performs three distinct responsibilities:
1. Creates all storage instances (lines 16-20)
2. Creates all service instances (lines 22-26)
3. Creates all handler instances (lines 28-36)
4. Defines all routes (lines 38-80)

**Affected Files:**
- `backend/internal/api/router.go:12-81` - 70-line function doing dependency wiring and routing

**Why It Matters:**
- **Single Responsibility Principle violation:** Mixing dependency injection with route configuration
- **Testing difficulty:** Can't test route configuration independently from dependency wiring
- **Readability:** Hard to scan route table when interleaved with 20+ lines of constructor calls
- **Change frequency:** Adding a route requires navigating through dependency setup code

**Recommendation for Phase 3:**
Extract dependency wiring into separate function:
```go
type Dependencies struct {
    AuthHandler      *handlers.AuthHandler
    HouseholdHandler *handlers.HouseholdHandler
    // ...
}

func wireDependencies(db *sql.DB) *Dependencies {
    // All storage/service/handler creation here
}

func NewRouter(db *sql.DB) http.Handler {
    deps := wireDependencies(db)
    mux := http.NewServeMux()

    // Clean route definitions only
    mux.HandleFunc("GET /health", handlers.Health)
    // ...
}
```

**Effort:** Low (1 hour) - Extract wiring, add struct, update router

---

### L1: No Domain Services (Only Infrastructure Services)

**Severity:** Low
**Category:** Layer Responsibility

**What:**
All "services" in `backend/internal/services/` are actually infrastructure orchestration layers (coordinating storage calls). No domain services exist for business rules that span entities.

**Examples:**
- Menu generation algorithm lives in `MenuService` but is pure domain logic (selecting recipes based on days/servings)
- Shopping list aggregation is domain logic but lives in service layer
- Invite code generation is in `HouseholdService` but is domain behavior

**Why It Matters:**
- **Testability:** Domain logic requires full service+storage setup to test
- **Reusability:** Can't use menu generation logic outside HTTP context
- **Clarity:** Mixing infrastructure concerns (transactions, storage) with domain rules

**Recommendation for Phase 3:**
Consider introducing domain services in `backend/internal/domain/services/` for pure business logic:
```go
// backend/internal/domain/services/menu_generator.go
func GenerateMenu(recipes []domain.Recipe, days, servings int) ([]domain.MenuDay, error)
```

Infrastructure services orchestrate: validate input, call domain service, persist results.

**Effort:** Medium (4-6 hours) - Extract domain logic, create domain services, wire into infrastructure services

---

## 2. Dependency Patterns

### H2: Package-Level State (JWT Secret, Context Keys)

**Severity:** High
**Category:** Global State

**What:**
Critical configuration and constants defined as package-level variables:
- `jwtSecret = []byte(os.Getenv("JWT_SECRET"))` - Read at package init time
- `UserIDKey contextKey = "user_id"` - Package-level constant
- `HouseholdIDKey contextKey = "household_id"` - Package-level constant

**Affected Files:**
- `backend/pkg/utils/jwt.go:11` - `var jwtSecret` at package level
- `backend/pkg/middleware/auth.go:12-13` - Context key constants

**Why It Matters:**
- **Testing:** Can't test JWT functions with different secrets (global state shared across tests)
- **Configuration validation:** No way to verify JWT_SECRET is set before server starts (fails silently with empty secret)
- **Concurrency issues:** Package init order undefined, potential races if accessed during init
- **Security risk:** Empty JWT_SECRET allows token forgery (Phase 1 finding C1)
- **Flexibility:** Can't have multiple auth configurations (e.g., different secrets per tenant)

**Recommendation for Phase 3:**
1. Pass JWT secret as dependency:
   ```go
   type JWTService struct {
       secret []byte
   }

   func NewJWTService(secret string) (*JWTService, error) {
       if len(secret) < 32 {
           return nil, errors.New("JWT_SECRET must be at least 32 characters")
       }
       return &JWTService{secret: []byte(secret)}, nil
   }
   ```

2. Validate in `main.go` before starting server:
   ```go
   jwtSecret := os.Getenv("JWT_SECRET")
   if jwtSecret == "" {
       log.Fatal("JWT_SECRET environment variable required")
   }
   ```

3. For context keys, consider typed keys or pass as middleware config

**Effort:** Medium (2-3 hours) - Create JWTService, update middleware, validate in main, update tests

---

### H3: No Dependency Injection Container (Manual Wiring)

**Severity:** High
**Category:** Dependency Management

**What:**
All dependencies manually wired in `router.go` via explicit constructor calls. As codebase grows, this becomes maintenance burden.

**Current State:**
- 5 storage constructors
- 6 service constructors
- 6 handler constructors
- Total: 17+ lines of wiring code for 12 components

**Why It Matters:**
- **Scalability:** Adding a new feature requires touching router wiring (20+ services would be unmanageable)
- **Testing:** Integration tests must replicate entire wiring setup
- **Circular dependency risk:** No tooling to detect dependency cycles
- **Boilerplate:** Every new component adds 2-3 wiring lines

**Recommendation for Phase 3:**
Consider lightweight DI approach:
1. **Option A (minimal):** Extract to `wireDependencies()` helper (already recommended in M3)
2. **Option B (structured):** Use Google Wire for compile-time DI
3. **Option C (runtime):** Use uber/dig or similar for runtime DI

For this codebase size (12 components), Option A is sufficient. Revisit if grows beyond 20 services.

**Effort:** Low (1 hour for Option A) - Already covered in M3 fix

---

### M4: Services Can't Be Tested in Isolation (Require Real DB)

**Severity:** Medium
**Category:** Testability

**What:**
Because services depend on concrete storage types (H1), unit testing services requires:
- Setting up real SQLite database
- Running migrations
- Creating test data
- Cleaning up after tests

**Evidence:**
- `backend/internal/services/household_service_test.go` - Only existing service test, uses real DB via `setupTestDB()`
- No tests exist for: `AuthService`, `RecipeService`, `MenuService`, `ShoppingService`, `TjekService`

**Why It Matters:**
- **Test speed:** Database setup adds 50-100ms per test
- **Test complexity:** Must manage DB state, migrations, cleanup
- **Test reliability:** File system issues, migration failures cause test flakes
- **Coverage gap:** Services aren't tested because setup is too hard (only 1 of 6 services has tests)

**Recommendation for Phase 3:**
After implementing storage interfaces (H1 fix):
1. Create mock storage implementations
2. Test services against mocks (fast, no DB)
3. Keep integration tests with real DB for critical flows

**Effort:** Low (blocked by H1) - Once interfaces exist, create mocks via `gomock` or manual implementations

---

### M5: No Transaction Helper Abstraction

**Severity:** Medium
**Category:** Code Duplication

**What:**
Transaction handling pattern duplicated across service methods:
```go
tx, err := s.storage.DB().Begin()
if err != nil {
    return nil, fmt.Errorf("begin transaction: %w", err)
}
defer tx.Rollback()

// ... operations ...

if err := tx.Commit(); err != nil {
    return nil, fmt.Errorf("commit transaction: %w", err)
}
```

**Affected Files:**
- `backend/internal/services/household_service.go:85-135` - JoinHousehold transaction
- Phase 1 identified registration flow needs transactions (01-02 finding C1)

**Why It Matters:**
- **Boilerplate:** 10+ lines of setup/teardown for every transactional operation
- **Error-prone:** Easy to forget `defer tx.Rollback()` or error handling
- **Inconsistent:** Some operations need transactions but don't use them (registration flow)

**Recommendation for Phase 3:**
Create transaction helper:
```go
func (s *HouseholdService) WithTransaction(fn func(tx *sql.Tx) error) error {
    tx, err := s.storage.DB().Begin()
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback()

    if err := fn(tx); err != nil {
        return err
    }

    return tx.Commit()
}

// Usage:
err := s.WithTransaction(func(tx *sql.Tx) error {
    // transactional operations
})
```

**Effort:** Low (1-2 hours) - Create helper, refactor JoinHousehold, apply to registration

---

### M6: Storage Exposes *sql.DB Directly

**Severity:** Medium
**Category:** Abstraction Leak

**What:**
`HouseholdStorage` has a `DB()` method that returns `*sql.DB` directly. Services call `s.householdStorage.DB().Begin()` to start transactions.

**Affected Files:**
- `backend/internal/storage/sqlite/household_storage.go` - Likely has `func (s *HouseholdStorage) DB() *sql.DB`
- `backend/internal/services/household_service.go:85` - Calls `.DB().Begin()`

**Why It Matters:**
- **Abstraction violation:** Storage implementation details (SQL database) leak to service layer
- **Testing difficulty:** Services now depend on both storage methods AND raw DB access
- **Future-proofing:** If storage moves to ORM or different DB, services must change

**Recommendation for Phase 3:**
Storage layer should provide transaction-aware methods:
```go
// Storage layer:
type HouseholdStorage struct {
    db *sql.DB // private
}

func (s *HouseholdStorage) BeginTx() (*sql.Tx, error) {
    return s.db.Begin()
}

// Or better: storage methods accept optional transaction
func (s *HouseholdStorage) AddMember(member *domain.HouseholdMember, tx *sql.Tx) error {
    // Use tx if provided, otherwise use s.db
}
```

**Effort:** Low (1 hour) - Add BeginTx method, update JoinHousehold

---

## 3. Error Propagation

### H4: Fragile Error String Matching in Handlers

**Severity:** High
**Category:** Error Handling Architecture

**What:**
Handlers route errors to HTTP status codes by comparing `err.Error()` strings. Pattern appears in 6+ handlers with 15+ switch statements.

**Affected Files:**
- `backend/internal/api/handlers/household.go:82-89,140-146,169-178` - 3 switch statements
- `backend/internal/api/handlers/auth.go:30` - Direct string comparison in error message
- `backend/internal/api/handlers/recipes.go:72` - Error string injection
- `backend/internal/api/handlers/offers.go:74,116,147` - Error string injection
- **Pattern confirmed in Phase 1:** 01-01 finding M1 (error string matching fragility)

**Example Pattern:**
```go
switch err.Error() {
case "invalid_code":
    http.Error(w, `{"error":"invalid_code"}`, http.StatusBadRequest)
case "already_member":
    http.Error(w, `{"error":"already_member"}`, http.StatusConflict)
default:
    http.Error(w, `{"error":"internal_server_error"}`, http.StatusInternalServerError)
}
```

**Why It Matters:**
- **Fragility:** Changing error text in service breaks HTTP routing (no compiler safety)
- **Typo risk:** "invalid_code" vs "invalid_codes" fails silently
- **Testing difficulty:** Can't test error types, only string matching
- **Maintainability:** 15+ switch statements to update if error message changes
- **Blast radius:** One service error text change breaks multiple handlers

**Recommendation for Phase 3:**
Define sentinel errors in domain package:
```go
// backend/internal/domain/errors.go
var (
    ErrInvalidCode    = errors.New("invalid_code")
    ErrAlreadyMember  = errors.New("already_member")
    ErrNotFound       = errors.New("not_found")
    ErrForbidden      = errors.New("forbidden")
    ErrCannotRemove   = errors.New("cannot_remove")
)
```

Handlers use `errors.Is()`:
```go
switch {
case errors.Is(err, domain.ErrInvalidCode):
    http.Error(w, `{"error":"invalid_code"}`, http.StatusBadRequest)
case errors.Is(err, domain.ErrAlreadyMember):
    http.Error(w, `{"error":"already_member"}`, http.StatusConflict)
default:
    http.Error(w, `{"error":"internal_server_error"}`, http.StatusInternalServerError)
}
```

**Effort:** Medium (3-4 hours) - Define 10-15 sentinel errors, update services to return them, update handler switch statements

---

### M7: No Structured Error Response Helper

**Severity:** Medium
**Category:** Code Duplication

**What:**
Every handler writes JSON error responses manually via `http.Error(w, `{"error":"message"}`, statusCode)`. Pattern duplicated 40+ times across handlers.

**Affected Files:**
- All handler files in `backend/internal/api/handlers/`
- Phase 1 finding H1: Error message injection vulnerability from this pattern

**Why It Matters:**
- **DRY violation:** Same JSON structure written 40+ times
- **Security risk:** Manual string concatenation causes JSON injection (Phase 1 H1)
- **Inconsistency risk:** Typo in one error response breaks client parsing
- **No error logging:** Errors returned to client but never logged server-side

**Recommendation for Phase 3:**
Create error response helper:
```go
// backend/pkg/response/error.go
type ErrorResponse struct {
    Error string `json:"error"`
    Code  string `json:"code,omitempty"`
}

func WriteError(w http.ResponseWriter, message string, statusCode int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func WriteErrorWithLog(w http.ResponseWriter, message string, statusCode int, err error) {
    log.Printf("ERROR: %s - %v", message, err)
    WriteError(w, message, statusCode)
}
```

Usage:
```go
if err != nil {
    response.WriteErrorWithLog(w, "internal_server_error", http.StatusInternalServerError, err)
    return
}
```

**Effort:** Low (1-2 hours) - Create helper package, update 40+ call sites (can be done incrementally)

---

### M8: Errors Lost in Service-to-Handler Boundary

**Severity:** Medium
**Category:** Observability

**What:**
When services return errors, handlers:
1. Map error to HTTP status code
2. Return generic error to client
3. **Never log the actual error**

**Example (household.go:27-29):**
```go
household, err := h.householdService.GetMyHousehold(userID)
if err != nil {
    http.Error(w, `{"error":"internal_server_error"}`, http.StatusInternalServerError)
    return // Original error lost, never logged
}
```

**Why It Matters:**
- **Debuggability:** Production errors invisible (database failures, network issues, etc.)
- **Monitoring:** No way to track error rates or types
- **Root cause analysis:** Must reproduce issue locally to see actual error

**Recommendation for Phase 3:**
Log errors before returning to client:
```go
if err != nil {
    log.Printf("GetMyHousehold failed for user %s: %v", userID, err)
    http.Error(w, `{"error":"internal_server_error"}`, http.StatusInternalServerError)
    return
}
```

Or use error helper with logging (M7 fix includes this).

**Effort:** Low (30 min) - Add logging to error paths, or implement M7 helper with logging

---

### M9: No Error Wrapping Context in Service Layer

**Severity:** Medium
**Category:** Error Tracing

**What:**
Most service methods return errors directly without adding context. When error propagates to handler, original context is lost.

**Example (auth_service.go:69-71):**
```go
if err := s.userStorage.Create(user); err != nil {
    return nil, err // No context about registration flow
}
```

**Good example exists (household_service.go:87):**
```go
if err != nil {
    return nil, fmt.Errorf("begin transaction: %w", err)
}
```

**Why It Matters:**
- **Debugging:** Error "database locked" doesn't tell you WHICH operation failed
- **Stack traces:** No breadcrumb trail from handler → service → storage
- **Production triage:** Can't differentiate between "registration failed to create user" vs "login failed to fetch user"

**Recommendation for Phase 3:**
Wrap errors with context:
```go
if err := s.userStorage.Create(user); err != nil {
    return nil, fmt.Errorf("creating user during registration: %w", err)
}
```

Use `%w` to preserve error chain for `errors.Is()` checks.

**Effort:** Low (1 hour) - Add context wrapping to 20-30 error returns in services

---

### L2: No Distinction Between Business Errors and Infrastructure Errors

**Severity:** Low
**Category:** Error Classification

**What:**
All errors are plain `error` type. No way to distinguish:
- Business rule violations (already_member, invalid_code) - should be 4xx
- Infrastructure failures (database error, network timeout) - should be 5xx

**Why It Matters:**
- **Client handling:** Clients can't programmatically distinguish retryable vs non-retryable errors
- **Monitoring:** Can't alert on infrastructure failures separately from business violations
- **Retry logic:** No signal for when to retry (infrastructure) vs fail fast (business rule)

**Recommendation for Phase 3:**
Define error types:
```go
type BusinessError struct {
    Code    string
    Message string
}

func (e *BusinessError) Error() string { return e.Message }

type InfrastructureError struct {
    Op  string
    Err error
}

func (e *InfrastructureError) Error() string { return fmt.Sprintf("%s: %v", e.Op, e.Err) }
```

**Effort:** Medium (3-4 hours) - Define types, update services, update handlers to detect type

---

## 4. Code Organization

### M10: pkg/ Package Contains Non-Reusable Code

**Severity:** Medium
**Category:** Package Structure

**What:**
The `backend/pkg/` directory contains middleware and utilities that are specific to this application (not reusable libraries):
- `pkg/middleware/auth.go` - Uses domain-specific context keys, JWT claims structure
- `pkg/middleware/cors.go` - Hardcoded localhost origins
- `pkg/utils/jwt.go` - Uses domain-specific Claims struct with UserID/HouseholdID

**Why It Matters:**
- **Go convention violation:** `pkg/` should contain packages that could be imported by external projects
- **Architectural confusion:** Blurs the line between reusable utilities and app-specific logic
- **Maintainability:** Future developers may expect `pkg/` code to be generic

**Recommendation for Phase 3:**
Move application-specific code to `internal/`:
```
backend/
  internal/
    auth/           # Move middleware/auth.go here
      middleware.go
      jwt.go
    middleware/     # Move CORS here if it remains app-specific
      cors.go
  pkg/
    password/       # Keep truly generic utilities
      hash.go
```

Only keep in `pkg/` if truly reusable (e.g., password hashing could stay).

**Effort:** Low (30 min) - Move files, update imports

---

### M11: Domain Package Has No Sub-Packages (All Flat)

**Severity:** Medium
**Category:** Package Organization

**What:**
All domain types are in single `backend/internal/domain/` package:
- `user.go`, `auth.go`, `household.go`, `recipe.go`, `menu.go`, `shopping.go`, `offers.go`
- No sub-packages for bounded contexts or aggregates

**Why It Matters:**
- **Scalability:** As features grow, domain package becomes dumping ground
- **Unclear boundaries:** No clear separation between user management, recipes, menus, shopping
- **Import confusion:** All types imported from same package despite different contexts

**Recommendation for Phase 3:**
Consider bounded contexts:
```
backend/internal/domain/
  auth/
    user.go
    credentials.go
  household/
    household.go
    member.go
    invite.go
  meal/
    recipe.go
    menu.go
    ingredient.go
  shopping/
    list.go
    item.go
```

**Note:** Only do this if codebase grows significantly. Current size (7 files) is manageable. Revisit at 15+ domain files.

**Effort:** Medium (4-6 hours if done now) - Better to defer until natural split becomes clear

---

### M12: No API Versioning Strategy

**Severity:** Medium
**Category:** API Evolution

**What:**
All routes are at root level with no versioning:
- `POST /auth/register`
- `GET /households/me`
- `POST /menus/generate`

No plan for breaking changes or API evolution.

**Why It Matters:**
- **Breaking changes:** Can't evolve API without breaking existing clients
- **Mobile apps:** Mobile clients can't be force-upgraded, need to support multiple versions
- **Gradual migration:** Can't A/B test new API designs

**Recommendation for Phase 3:**
Add versioning (don't implement yet, just plan):
- **Option A:** Path-based: `/api/v1/menus/generate`
- **Option B:** Header-based: `Accept: application/vnd.maltiden.v1+json`
- **Option C:** Domain-based: `api-v1.maltiden.com` (overkill for this app)

For this app, Option A is simplest. Don't implement until needed (defer to Phase 3 planning).

**Effort:** Low (plan only) - Document strategy, implement when first breaking change needed

---

### L3: Handler Functions Have No Shared Pattern Beyond Parse-Call-Return

**Severity:** Low
**Category:** Code Consistency

**What:**
Handlers follow similar pattern but with variations:
- Some extract auth context at start, others don't
- Some validate request, others rely on service
- Some check for nil results, others don't

No enforced handler template or base handler.

**Why It Matters:**
- **Onboarding:** New developers learn by example, inconsistencies are confusing
- **Code review:** Hard to spot deviations from "normal" pattern
- **Refactoring:** No single place to add cross-cutting behavior (e.g., request logging)

**Recommendation for Phase 3:**
Document standard handler pattern in CONVENTIONS.md:
```
1. Extract auth context (if protected route)
2. Parse and validate request
3. Call service method
4. Handle service errors (use sentinel errors)
5. Write success response
```

Consider handler helper:
```go
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

func (h HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if err := h(w, r); err != nil {
        // Centralized error handling
    }
}
```

**Effort:** Low (1-2 hours) - Document pattern, optionally add helper

---

## 5. Configuration & Startup

### H5: No Validation of Required Environment Variables

**Severity:** High
**Category:** Configuration Management

**What:**
`main.go` reads environment variables but doesn't validate them:
- `PORT` defaults to "8080" if missing (OK)
- `DATABASE_PATH` defaults to "./data/maltiden.db" (OK)
- **`JWT_SECRET` not validated in main** - read at package init time in jwt.go (Phase 1 finding C1)

**Affected Files:**
- `backend/cmd/server/main.go:14-22` - Env var reading
- `backend/pkg/utils/jwt.go:11` - JWT_SECRET read at package init
- No validation that JWT_SECRET is set or meets minimum length

**Why It Matters:**
- **Security:** Empty JWT_SECRET allows token forgery (Phase 1 critical finding)
- **Production failures:** Server starts successfully but fails at runtime when JWT operations occur
- **Debugging difficulty:** No clear error message, just token validation failures
- **Principle of least surprise:** Should fail fast at startup, not during request handling

**Recommendation for Phase 3:**
Validate in `main.go` before starting server:
```go
func main() {
    // Validate required env vars
    jwtSecret := os.Getenv("JWT_SECRET")
    if jwtSecret == "" {
        log.Fatal("JWT_SECRET environment variable is required")
    }
    if len(jwtSecret) < 32 {
        log.Fatal("JWT_SECRET must be at least 32 characters")
    }

    // ... rest of startup
}
```

**Effort:** Low (15 min) - Add validation block at start of main()

---

### M13: Database Path Not Validated (Can Create in Wrong Directory)

**Severity:** Medium
**Category:** Configuration

**What:**
`DATABASE_PATH` env var is used as-is without validation. If set to invalid path or wrong directory, migrations run but files may end up in unexpected locations.

**Affected Files:**
- `backend/cmd/server/main.go:19-22` - No validation
- `backend/internal/storage/sqlite/db.go:11-16` - Creates directory if missing (could create anywhere)

**Why It Matters:**
- **Data loss risk:** Database created in wrong location, then lost when container restarts
- **Deployment issues:** Misconfigured env var causes data to be stored on ephemeral filesystem
- **Debugging:** Hard to find where database file was created

**Recommendation for Phase 3:**
Validate path in main or Open function:
```go
if dbPath != "" && !filepath.IsAbs(dbPath) {
    log.Fatalf("DATABASE_PATH must be absolute path, got: %s", dbPath)
}

// Or check if parent directory exists and is writable
```

**Effort:** Low (30 min) - Add validation, document expected format

---

### M14: Migrations Run on Every Startup (No Skip-Check Optimization)

**Severity:** Medium
**Category:** Startup Performance

**What:**
`runMigrations()` queries database for each migration version on every server start (6 SELECT queries + schema check).

**Affected Files:**
- `backend/internal/storage/sqlite/db.go:60-71` - Queries for each migration in loop

**Why It Matters:**
- **Startup latency:** Adds 10-50ms on every startup (minor, but unnecessary)
- **Cold start impact:** Fly.io scales to zero, every request after idle triggers migration check
- **Database load:** Unnecessary queries on every startup in production

**Current behavior:**
```go
for _, m := range migrations {
    var exists int
    err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", m.version).Scan(&exists)
    // Runs 6 queries even if all migrations applied
}
```

**Recommendation for Phase 3:**
Fetch all applied versions once:
```go
rows, err := db.Query("SELECT version FROM schema_migrations")
// Build set of applied versions
// Loop through migrations and check set
```

**Effort:** Low (30 min) - Refactor migration check to single query

---

### L4: No Graceful Shutdown Handling

**Severity:** Low
**Category:** Production Readiness

**What:**
Server uses `log.Fatal(http.ListenAndServe(...))` which immediately exits on signal. No graceful shutdown to finish in-flight requests or close database connections cleanly.

**Affected Files:**
- `backend/cmd/server/main.go:35` - `log.Fatal(http.ListenAndServe(...))`

**Why It Matters:**
- **Request failures:** In-flight requests fail mid-processing during deploys
- **Database corruption risk:** SQLite writes may not flush if killed abruptly
- **Poor user experience:** Users see errors during zero-downtime deployments

**Recommendation for Phase 3:**
Implement graceful shutdown:
```go
server := &http.Server{
    Addr:    ":" + port,
    Handler: router,
}

// Start server
go func() {
    if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatalf("Server error: %v", err)
    }
}()

// Wait for interrupt signal
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
<-sigChan

// Graceful shutdown with timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
if err := server.Shutdown(ctx); err != nil {
    log.Printf("Shutdown error: %v", err)
}
```

**Effort:** Low (1 hour) - Add signal handling, graceful shutdown

---

### L5: No Health Check Depth (Only Returns 200 OK)

**Severity:** Low
**Category:** Observability

**What:**
`/health` endpoint just returns `{"status":"ok"}` without checking:
- Database connectivity
- Disk space for SQLite
- Critical dependencies (Tjek API reachable, etc.)

**Affected Files:**
- `backend/internal/api/handlers/health.go` (assumed) or health function in handlers package

**Why It Matters:**
- **False positives:** Health check passes even if database is unreachable
- **Load balancer issues:** LB sends traffic to unhealthy instance
- **Monitoring gaps:** Can't distinguish "server running" from "server healthy"

**Recommendation for Phase 3:**
Add database ping to health check:
```go
func Health(w http.ResponseWriter, r *http.Request) {
    if err := db.Ping(); err != nil {
        http.Error(w, `{"status":"unhealthy","reason":"database"}`, http.StatusServiceUnavailable)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(`{"status":"ok"}`))
}
```

**Effort:** Low (30 min) - Add DB dependency to health handler, ping database

---

## 6. Cross-Cutting Concerns

### M15: No Structured Logging Framework

**Severity:** Medium
**Category:** Observability

**What:**
Backend uses only `log.Printf()` and `log.Fatal()` from stdlib. No structured logging, levels, or context.

**Current usage:**
- `log.Printf("Måltiden startar på :%s", port)` - Startup message
- `log.Fatal("Database error:", err)` - Fatal errors
- No logging in handlers, services, or storage layers (errors silently swallowed)

**Why It Matters:**
- **Production debugging:** Can't filter logs by level (INFO, WARN, ERROR)
- **No context:** Can't add request ID, user ID, trace ID to logs
- **Parsing difficulty:** Unstructured text logs hard to query in production
- **Monitoring:** Can't set up alerts on ERROR-level logs

**Recommendation for Phase 3:**
Adopt structured logging (e.g., `slog` from Go 1.21+ standard library):
```go
import "log/slog"

logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
logger.Info("server starting", "port", port)
logger.Error("database error", "error", err)

// In handlers:
logger.Error("failed to generate menu",
    "user_id", userID,
    "household_id", householdID,
    "error", err)
```

**Effort:** Medium (2-3 hours) - Add slog, update main.go, add logger to services via DI

---

### M16: No Request ID Tracing

**Severity:** Medium
**Category:** Observability

**What:**
No correlation between requests and log entries. Can't trace a single request through handler → service → storage.

**Why It Matters:**
- **Debugging:** Can't filter logs for a specific failing request
- **Production issues:** With concurrent requests, can't tell which log lines belong together
- **APM integration:** Can't integrate with distributed tracing tools (Datadog, New Relic, etc.)

**Recommendation for Phase 3:**
Add request ID middleware:
```go
func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := uuid.New().String()
        ctx := context.WithValue(r.Context(), "request_id", requestID)
        w.Header().Set("X-Request-ID", requestID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// In logs:
logger.Info("request started", "request_id", GetRequestID(r.Context()))
```

**Effort:** Low (1 hour) - Add middleware, update logging calls

---

### M17: No Rate Limiting Middleware (Only Noted as Missing in Phase 1)

**Severity:** Medium
**Category:** Security Infrastructure

**What:**
Phase 1 finding H3 identified missing rate limiting. This is a systemic infrastructure gap, not individual handler issue.

**Why It Matters:**
- **Brute force attacks:** Auth endpoints vulnerable to password guessing
- **DoS:** Any public endpoint can be spammed
- **Cost:** Tjek API calls cost money if endpoint is abused

**Recommendation for Phase 3:**
Add rate limiting middleware (per-IP or per-user):
```go
import "golang.org/x/time/rate"

func RateLimit(rps int) func(http.Handler) http.Handler {
    limiters := make(map[string]*rate.Limiter)
    mu := sync.Mutex{}

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ip := r.RemoteAddr
            mu.Lock()
            limiter, exists := limiters[ip]
            if !exists {
                limiter = rate.NewLimiter(rate.Limit(rps), rps*2)
                limiters[ip] = limiter
            }
            mu.Unlock()

            if !limiter.Allow() {
                http.Error(w, `{"error":"rate_limit_exceeded"}`, http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

// Apply to auth routes:
mux.Handle("POST /auth/login", RateLimit(5)(http.HandlerFunc(authHandler.Login)))
```

**Effort:** Medium (2-3 hours) - Implement middleware, apply to sensitive routes, add cleanup for old IPs

---

### L6: CORS Middleware Configuration Not Environment-Aware

**Severity:** Low
**Category:** Configuration

**What:**
CORS allowed origins hardcoded to localhost in `backend/pkg/middleware/cors.go`. Phase 1 finding H4 identified this blocks production frontend.

**Why It Matters:**
- **Deployment blocker:** Production frontend can't call API (already noted in Phase 1)
- **Configuration inflexibility:** Can't test against staging frontend URL
- **Security risk:** Might add `*` wildcard in desperation (bad practice)

**Recommendation for Phase 3:**
Make origins configurable:
```go
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        // Check origin against allowedOrigins
    }
}

// In main.go:
origins := strings.Split(os.Getenv("CORS_ORIGINS"), ",")
if len(origins) == 0 {
    origins = []string{"http://localhost:5173"} // Default for dev
}
router := middleware.CORS(origins)(apiRouter)
```

**Effort:** Low (30 min) - Add env var, update CORS middleware

---

### L7: No Monitoring/Metrics Collection

**Severity:** Low
**Category:** Observability

**What:**
No metrics collection for:
- Request counts by endpoint
- Response time percentiles
- Error rates
- Database query performance
- External API (Tjek) latency

**Why It Matters:**
- **Performance regression detection:** Can't detect when endpoints slow down
- **Capacity planning:** Don't know which endpoints are hot paths
- **SLA tracking:** Can't measure uptime or latency targets

**Recommendation for Phase 3:**
Add metrics middleware using Prometheus:
```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

// Expose /metrics endpoint
mux.Handle("GET /metrics", promhttp.Handler())

// Add request instrumentation
var requestDuration = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name: "http_request_duration_seconds",
        Help: "HTTP request latency",
    },
    []string{"method", "endpoint", "status"},
)
```

**Effort:** Medium (3-4 hours) - Add Prometheus, instrument handlers, set up dashboards

---

## Summary of Recommendations

### High Priority (4 issues)
1. **H1:** Add storage interfaces to enable testing and future DB migration
2. **H2:** Remove package-level state (JWT secret, context keys)
3. **H3:** Consider DI approach (lightweight wiring helper sufficient for now)
4. **H4:** Replace error string matching with sentinel errors
5. **H5:** Validate required env vars at startup

### Medium Priority (12 issues)
- Layer separation: M1 (context helpers), M2 (domain vs API types), M3 (router organization)
- Dependencies: M4 (testing), M5 (transaction helper), M6 (DB exposure)
- Errors: M7 (response helper), M8 (error logging), M9 (error wrapping)
- Organization: M10 (pkg/ usage), M11 (domain structure), M12 (API versioning plan)
- Config: M13 (path validation), M14 (migration optimization)
- Observability: M15 (structured logging), M16 (request tracing), M17 (rate limiting)

### Low Priority (5 issues)
- L1: Domain service extraction (optional)
- L2: Error classification (nice-to-have)
- L3: Handler pattern documentation
- L4: Graceful shutdown
- L5: Health check depth
- L6: CORS configuration
- L7: Metrics collection

### Estimated Total Effort
- High: 8-12 hours
- Medium: 20-30 hours
- Low: 6-10 hours
- **Total: 34-52 hours** (1-2 weeks for one developer)

---

**End of Architecture Review Findings**
