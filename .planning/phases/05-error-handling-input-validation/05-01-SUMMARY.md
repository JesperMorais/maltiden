---
phase: 05-error-handling-input-validation
plan: 01
subsystem: backend-error-handling
tags: [error-handling, sentinel-errors, response-helpers, injection-prevention]
depends_on:
  requires: [04-01, 04-02, 04-03]
  provides: [sentinel-errors, response-helpers, error-logging]
  affects: [05-02, 06-xx, 07-xx]
tech-stack:
  added: []
  patterns: [sentinel-error-pattern, errors-is-matching, structured-json-responses]
key-files:
  created:
    - backend/internal/domain/errors.go
    - backend/internal/api/handlers/response.go
  modified:
    - backend/internal/services/auth_service.go
    - backend/internal/services/household_service.go
    - backend/internal/services/recipe_service.go
    - backend/internal/api/handlers/auth.go
    - backend/internal/api/handlers/household.go
    - backend/internal/api/handlers/recipes.go
    - backend/internal/api/handlers/menus.go
    - backend/internal/api/handlers/shopping.go
    - backend/internal/api/handlers/offers.go
    - backend/pkg/middleware/auth.go
    - backend/internal/services/household_service_test.go
decisions:
  - Sentinel errors as package-level vars in domain/errors.go (14 errors)
  - WriteError/WriteJSON helpers in handlers package
  - Middleware uses local writeError to avoid import cycle with handlers
  - Offers handlers return generic "service_unavailable" (no upstream error leakage)
metrics:
  duration: 6min
  completed: 2026-02-09
---

# Phase 5 Plan 1: Error Handling Infrastructure Summary

Sentinel errors, response helpers, and type-safe error checking across all handlers and services.

## One-liner

Sentinel errors in domain/errors.go with WriteError/WriteJSON helpers eliminating 60 inline JSON responses, 3 string-matching switches, and 5 error injection sites.

## What Was Done

### Task 1: Create sentinel errors and migrate services (e6cbcb8)

Created `backend/internal/domain/errors.go` with 14 sentinel errors covering all business error codes:
- Auth: ErrUnauthorized, ErrForbidden, ErrInvalidCredentials, ErrWeakPassword
- User: ErrDuplicateEmail
- Household: ErrInvalidCode, ErrAlreadyMember, ErrCannotRemove, ErrCodeRequired
- Resource: ErrNotFound
- Recipe validation: ErrNameRequired, ErrInvalidServings, ErrIngredientsRequired, ErrInstructionsRequired

Migrated all services to return sentinel errors instead of `errors.New("string")`:
- `auth_service.go`: 4 replacements (weak_password, email_already_exists, invalid_credentials x2)
- `household_service.go`: 8 replacements (invalid_code x3, already_member, not_found x2, forbidden, cannot_remove x2, code_required)
- `recipe_service.go`: 4 replacements (name_required, invalid_servings, ingredients_required, instructions_required)

### Task 2: Create response helpers and migrate all handlers (2691d96)

Created `backend/internal/api/handlers/response.go` with:
- `WriteJSON(w, status, data)` -- structured JSON success responses
- `WriteError(w, status, code)` -- JSON error responses via json.Encoder (prevents injection)

Migrated all handler files (auth, household, recipes, menus, shopping, offers) and auth middleware:
- Replaced all 60 `http.Error(w, '{"error":"..."}', status)` calls with `WriteError()`
- Replaced 3 `switch err.Error()` blocks with `errors.Is()` chains
- Fixed 5 `err.Error()` injection sites (auth Register, recipes Create, offers x3)
- Replaced all manual `w.Header().Set + json.NewEncoder` patterns with `WriteJSON()`
- Added `log.Printf("ERROR [HandlerName] %v", err)` at every internal error boundary

## Findings Addressed

| Finding | Severity | Fix |
|---------|----------|-----|
| ERROR-01: String matching for error dispatch | High | All `switch err.Error()` replaced with `errors.Is()` |
| ERROR-02: No structured error response helper | Medium | WriteError/WriteJSON helpers in response.go |
| ERROR-03: Errors lost at service-to-handler boundary | Medium | Error logging at all default/500 branches |
| VALID-01: Error message injection in responses | High | WriteError uses json.Encoder, no string concat |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Go not installed on build machine**
- **Found during:** Task 1 verification
- **Issue:** Go binary not available in PATH or any standard location
- **Fix:** Downloaded and installed Go 1.24.1 (matching go.mod) to /home/david/go
- **Impact:** Required for build verification; no code changes

**2. [Rule 1 - Bug] Test file had wrong NewAuthService call signature**
- **Found during:** Task 2 verification (go vet)
- **Issue:** `household_service_test.go` called `NewAuthService(userStorage, householdStorage)` missing the `db` parameter added in Phase 4-02
- **Fix:** Updated all 12 test function calls to `NewAuthService(db, userStorage, householdStorage)`
- **Files modified:** `backend/internal/services/household_service_test.go`
- **Commit:** 2691d96

## Decisions Made

| Decision | Rationale |
|----------|-----------|
| 14 sentinel errors as package-level vars | Direct returns enable `errors.Is()` without unwrapping; wrapping deferred to Phase 7 (ERROR-04) |
| WriteError/WriteJSON in handlers package | Natural location; all handlers already in this package |
| Local writeError in middleware | Avoids import cycle (middleware -> handlers -> middleware) |
| Generic "service_unavailable" for offers errors | Prevents leaking upstream Tjek API error details to clients |
| Error logging with handler name prefix | Enables grep-based log analysis: `ERROR [HandlerName]` pattern |

## Verification Results

- [x] `go build ./...` passes
- [x] `go vet ./...` passes
- [x] Zero `http.Error(` in handlers
- [x] Zero `http.Error(` in middleware
- [x] Zero `err.Error()` in handler code (only in response.go comment)
- [x] Zero `switch err.Error()` in backend
- [x] Zero `errors.New(` with business error codes in services

## Next Phase Readiness

Phase 5 Plan 2 (input validation) can proceed. All error infrastructure is in place:
- Sentinel errors available for any new validation errors
- WriteError/WriteJSON ready for new handler code
- Error logging pattern established for consistency
